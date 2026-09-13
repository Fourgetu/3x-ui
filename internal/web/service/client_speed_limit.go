package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/userspeed"
	"github.com/mhsanaei/3x-ui/v3/internal/util/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxClientSpeedMbps = 100000

var speedLimitWriteMu sync.Mutex

type ClientSpeedLimitView struct {
	InboundID         int    `json:"inboundId"`
	InboundRemark     string `json:"inboundRemark"`
	Protocol          string `json:"protocol"`
	Enabled           bool   `json:"enabled"`
	UploadMbps        int    `json:"uploadMbps"`
	DownloadMbps      int    `json:"downloadMbps"`
	ExternalPort      *int   `json:"externalPort"`
	InternalPort      *int   `json:"internalPort"`
	Supported         bool   `json:"supported"`
	UnsupportedReason string `json:"unsupportedReason,omitempty"`
	RuntimeStatus     string `json:"runtimeStatus"`
	LastError         string `json:"lastError,omitempty"`
}

type ClientSpeedLimitUpdate struct {
	InboundID    int  `json:"inboundId" validate:"required,gt=0"`
	Enabled      bool `json:"enabled"`
	UploadMbps   int  `json:"uploadMbps" validate:"gte=0,lte=100000"`
	DownloadMbps int  `json:"downloadMbps" validate:"gte=0,lte=100000"`
}

type userSpeedLimitRow struct {
	model.ClientSpeedLimit
	ClientEmail    string         `gorm:"column:client_email"`
	ClientEnabled  bool           `gorm:"column:client_enabled"`
	InboundEnabled bool           `gorm:"column:inbound_enabled"`
	Protocol       model.Protocol `gorm:"column:protocol"`
	StreamSettings string         `gorm:"column:stream_settings"`
	NodeID         *int           `gorm:"column:node_id"`
}

type UserSpeedLimitService struct{}

func (s *UserSpeedLimitService) ListByEmail(email string) ([]ClientSpeedLimitView, error) {
	rec := &model.ClientRecord{}
	db := database.GetDB()
	if err := db.Where("email = ?", email).First(rec).Error; err != nil {
		return nil, err
	}
	var inbounds []model.Inbound
	if err := db.Table("inbounds").
		Joins("JOIN client_inbounds ON client_inbounds.inbound_id = inbounds.id").
		Where("client_inbounds.client_id = ?", rec.Id).
		Order("inbounds.id ASC").Find(&inbounds).Error; err != nil {
		return nil, err
	}
	var rows []model.ClientSpeedLimit
	if err := db.Where("client_id = ?", rec.Id).Find(&rows).Error; err != nil {
		return nil, err
	}
	byInbound := make(map[int]model.ClientSpeedLimit, len(rows))
	for _, row := range rows {
		byInbound[row.InboundID] = row
	}
	status := userspeed.GetManager().Status()
	result := make([]ClientSpeedLimitView, 0, len(inbounds))
	for _, inbound := range inbounds {
		row := byInbound[inbound.Id]
		networks, reason := speedLimitNetworks(&inbound)
		supported := len(networks) > 0
		runtimeStatus := "disabled"
		if row.Enabled {
			runtimeStatus = "error"
			if status.Running && row.LastError == "" {
				runtimeStatus = "running"
			}
		}
		result = append(result, ClientSpeedLimitView{
			InboundID: inbound.Id, InboundRemark: inbound.Remark, Protocol: string(inbound.Protocol),
			Enabled: row.Enabled, UploadMbps: row.UploadMbps, DownloadMbps: row.DownloadMbps,
			ExternalPort: row.ExternalPort, InternalPort: row.InternalPort,
			Supported: supported, UnsupportedReason: reason, RuntimeStatus: runtimeStatus, LastError: row.LastError,
		})
	}
	return result, nil
}

func (s *UserSpeedLimitService) Update(email string, request ClientSpeedLimitUpdate, xrayService *XrayService) (*ClientSpeedLimitView, error) {
	if request.UploadMbps < 0 || request.DownloadMbps < 0 || request.UploadMbps > maxClientSpeedMbps || request.DownloadMbps > maxClientSpeedMbps {
		return nil, common.NewError("speed limit must be between 0 and 100000 Mbps")
	}
	speedLimitWriteMu.Lock()
	defer speedLimitWriteMu.Unlock()

	db := database.GetDB()
	rec := &model.ClientRecord{}
	if err := db.Where("email = ?", email).First(rec).Error; err != nil {
		return nil, err
	}
	inbound := &model.Inbound{}
	if err := db.Table("inbounds").
		Joins("JOIN client_inbounds ON client_inbounds.inbound_id = inbounds.id").
		Where("client_inbounds.client_id = ? AND inbounds.id = ?", rec.Id, request.InboundID).
		First(inbound).Error; err != nil {
		return nil, common.NewError("client is not attached to inbound")
	}
	if request.Enabled {
		if networks, reason := speedLimitNetworks(inbound); len(networks) == 0 {
			return nil, common.NewError(reason)
		}
	}

	previous := &model.ClientSpeedLimit{}
	findErr := db.Where("client_id = ? AND inbound_id = ?", rec.Id, inbound.Id).First(previous).Error
	hadPrevious := findErr == nil
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, findErr
	}
	next := model.ClientSpeedLimit{
		ClientID: rec.Id, InboundID: inbound.Id, Enabled: request.Enabled,
		UploadMbps: request.UploadMbps, DownloadMbps: request.DownloadMbps,
	}
	if hadPrevious {
		next.CreatedAt = previous.CreatedAt
		next.ExternalPort = previous.ExternalPort
		next.InternalPort = previous.InternalPort
	}
	if request.Enabled {
		if next.ExternalPort == nil || next.InternalPort == nil {
			external, internal, err := allocateSpeedPorts(db, rec.Id, inbound.Id)
			if err != nil {
				return nil, err
			}
			next.ExternalPort = &external
			next.InternalPort = &internal
		}
	} else {
		next.ExternalPort = nil
		next.InternalPort = nil
	}
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&next).Error; err != nil {
		return nil, err
	}

	if err := s.Apply(xrayService); err != nil {
		applyError := err.Error()
		if hadPrevious {
			previous.LastError = applyError
			_ = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(previous).Error
		} else {
			failed := model.ClientSpeedLimit{
				ClientID: rec.Id, InboundID: inbound.Id, UploadMbps: request.UploadMbps,
				DownloadMbps: request.DownloadMbps, LastError: applyError,
			}
			_ = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&failed).Error
		}
		_ = s.Apply(xrayService)
		return nil, fmt.Errorf("speed-limit runtime apply failed: %w", err)
	}
	if err := db.Model(&model.ClientSpeedLimit{}).
		Where("client_id = ? AND inbound_id = ?", rec.Id, inbound.Id).
		Update("last_error", "").Error; err != nil {
		return nil, err
	}
	views, err := s.ListByEmail(email)
	if err != nil {
		return nil, err
	}
	for i := range views {
		if views[i].InboundID == inbound.Id {
			return &views[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *UserSpeedLimitService) Apply(xrayService *XrayService) error {
	if xrayService != nil {
		if err := xrayService.RestartXray(false); err != nil {
			return fmt.Errorf("apply Xray identity routes: %w", err)
		}
	}
	routes, err := desiredUserSpeedRoutes()
	if err != nil {
		return err
	}
	return userspeed.GetManager().Reconcile(routes)
}

func desiredUserSpeedRoutes() ([]userspeed.DesiredRoute, error) {
	var rows []userSpeedLimitRow
	err := database.GetDB().Table("client_speed_limits AS sl").
		Select("sl.*, clients.email AS client_email, clients.enable AS client_enabled, inbounds.enable AS inbound_enabled, inbounds.protocol, inbounds.stream_settings, inbounds.node_id").
		Joins("JOIN clients ON clients.id = sl.client_id").
		Joins("JOIN inbounds ON inbounds.id = sl.inbound_id").
		Where("sl.enabled = ? AND sl.external_port IS NOT NULL AND sl.internal_port IS NOT NULL", true).
		Order("sl.inbound_id ASC, sl.client_id ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	routes := make([]userspeed.DesiredRoute, 0, len(rows))
	for _, row := range rows {
		if !row.ClientEnabled || !row.InboundEnabled || row.ExternalPort == nil || row.InternalPort == nil {
			continue
		}
		inbound := &model.Inbound{Protocol: row.Protocol, StreamSettings: row.StreamSettings, NodeID: row.NodeID}
		networks, _ := speedLimitNetworks(inbound)
		if len(networks) == 0 {
			continue
		}
		routes = append(routes, userspeed.DesiredRoute{
			ID: fmt.Sprintf("%d-%d", row.InboundID, row.ClientID), InboundID: row.InboundID, ClientID: row.ClientID,
			ClientLabel: maskedClientLabel(row.ClientEmail), ExternalPort: *row.ExternalPort, InternalPort: *row.InternalPort,
			UploadMbps: row.UploadMbps, DownloadMbps: row.DownloadMbps, Networks: networks,
		})
	}
	return routes, nil
}

func activeUserSpeedLimits() (map[int]map[string]model.ClientSpeedLimit, error) {
	var rows []userSpeedLimitRow
	err := database.GetDB().Table("client_speed_limits AS sl").
		Select("sl.*, clients.email AS client_email, clients.enable AS client_enabled, inbounds.enable AS inbound_enabled, inbounds.protocol, inbounds.stream_settings, inbounds.node_id").
		Joins("JOIN clients ON clients.id = sl.client_id").
		Joins("JOIN inbounds ON inbounds.id = sl.inbound_id").
		Where("sl.enabled = ? AND sl.internal_port IS NOT NULL AND sl.external_port IS NOT NULL", true).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int]map[string]model.ClientSpeedLimit)
	for _, row := range rows {
		inbound := &model.Inbound{Protocol: row.Protocol, StreamSettings: row.StreamSettings, NodeID: row.NodeID}
		if !row.ClientEnabled || !row.InboundEnabled {
			continue
		}
		if networks, _ := speedLimitNetworks(inbound); len(networks) == 0 {
			continue
		}
		if result[row.InboundID] == nil {
			result[row.InboundID] = make(map[string]model.ClientSpeedLimit)
		}
		result[row.InboundID][strings.ToLower(row.ClientEmail)] = row.ClientSpeedLimit
	}
	return result, nil
}

func speedLimitNetworks(inbound *model.Inbound) ([]string, string) {
	if inbound == nil {
		return nil, "inbound not found"
	}
	if inbound.NodeID != nil {
		return nil, "configure the speed limit on the panel that hosts this inbound"
	}
	var stream map[string]any
	_ = json.Unmarshal([]byte(inbound.StreamSettings), &stream)
	network, _ := stream["network"].(string)
	if network == "" {
		network = "tcp"
	}
	switch inbound.Protocol {
	case model.VLESS, model.VMESS, model.Trojan:
		switch network {
		case "tcp", "raw", "ws", "grpc", "httpupgrade", "xhttp", "splithttp":
			return []string{"tcp"}, ""
		default:
			return nil, fmt.Sprintf("%s transport is not supported by the identity-bound limiter", network)
		}
	case model.Shadowsocks:
		if network == "tcp" || network == "raw" {
			return []string{"tcp", "udp"}, ""
		}
		return nil, fmt.Sprintf("Shadowsocks %s transport is not supported by the identity-bound limiter", network)
	default:
		return nil, fmt.Sprintf("%s is not supported by the identity-bound limiter", inbound.Protocol)
	}
}

func allocateSpeedPorts(db *gorm.DB, clientID, inboundID int) (int, int, error) {
	used := map[int]struct{}{}
	var inboundPorts []int
	if err := db.Model(&model.Inbound{}).Pluck("port", &inboundPorts).Error; err != nil {
		return 0, 0, err
	}
	for _, port := range inboundPorts {
		used[port] = struct{}{}
	}
	var rows []model.ClientSpeedLimit
	if err := db.Where("NOT (client_id = ? AND inbound_id = ?)", clientID, inboundID).Find(&rows).Error; err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		if row.ExternalPort != nil {
			used[*row.ExternalPort] = struct{}{}
		}
		if row.InternalPort != nil {
			used[*row.InternalPort] = struct{}{}
		}
	}
	external, err := firstFreePort(used, envPort("XUI_USER_SPEED_EXTERNAL_PORT_START", 32000), envPort("XUI_USER_SPEED_EXTERNAL_PORT_END", 32999))
	if err != nil {
		return 0, 0, err
	}
	used[external] = struct{}{}
	internal, err := firstFreePort(used, envPort("XUI_USER_SPEED_INTERNAL_PORT_START", 42000), envPort("XUI_USER_SPEED_INTERNAL_PORT_END", 42999))
	if err != nil {
		return 0, 0, err
	}
	return external, internal, nil
}

func firstFreePort(used map[int]struct{}, start, end int) (int, error) {
	if start < 1 || end > 65535 || start > end {
		return 0, errors.New("invalid user speed-limit port range")
	}
	for port := start; port <= end; port++ {
		if _, exists := used[port]; !exists {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port in range %d-%d", start, end)
}

func envPort(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value == 0 {
		return fallback
	}
	return value
}

func maskedClientLabel(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:6])
}
