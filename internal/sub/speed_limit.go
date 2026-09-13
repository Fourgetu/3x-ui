package sub

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func (s *SubService) loadClientSpeedPorts() {
	type route struct {
		InboundID    int    `gorm:"column:inbound_id"`
		Email        string `gorm:"column:email"`
		ExternalPort int    `gorm:"column:external_port"`
	}
	var rows []route
	_ = database.GetDB().Table("client_speed_limits AS sl").
		Select("sl.inbound_id, clients.email, sl.external_port").
		Joins("JOIN clients ON clients.id = sl.client_id").
		Where("sl.enabled = ? AND sl.external_port IS NOT NULL", true).Scan(&rows).Error
	s.clientSpeedPorts = make(map[int]map[string]int)
	for _, row := range rows {
		if s.clientSpeedPorts[row.InboundID] == nil {
			s.clientSpeedPorts[row.InboundID] = make(map[string]int)
		}
		s.clientSpeedPorts[row.InboundID][strings.ToLower(row.Email)] = row.ExternalPort
	}
}

func (s *SubService) withClientSpeedPort(inbound *model.Inbound, email string) (*model.Inbound, int) {
	if inbound == nil {
		return inbound, 0
	}
	port := s.clientSpeedPorts[inbound.Id][strings.ToLower(email)]
	if port == 0 {
		return inbound, 0
	}
	clone := *inbound
	clone.Port = port
	var stream map[string]any
	if json.Unmarshal([]byte(clone.StreamSettings), &stream) == nil {
		if endpoints, ok := stream["externalProxy"].([]any); ok {
			for _, value := range endpoints {
				if endpoint, ok := value.(map[string]any); ok {
					endpoint["port"] = float64(port)
				}
			}
		}
		if encoded, err := json.Marshal(stream); err == nil {
			clone.StreamSettings = string(encoded)
		}
	}
	return &clone, port
}

func forceEndpointPort(endpoints []map[string]any, port int) []map[string]any {
	if port == 0 {
		return endpoints
	}
	result := make([]map[string]any, len(endpoints))
	for i, endpoint := range endpoints {
		result[i] = maps.Clone(endpoint)
		result[i]["port"] = float64(port)
	}
	return result
}
