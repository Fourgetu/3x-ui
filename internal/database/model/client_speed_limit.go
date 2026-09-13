package model

// ClientSpeedLimit owns one identity-bound public route for one inbound.
// Nil ports mean the route is disabled and the allocation is available again.
type ClientSpeedLimit struct {
	ClientID     int    `json:"-" gorm:"primaryKey;column:client_id;index"`
	InboundID    int    `json:"inboundId" gorm:"primaryKey;column:inbound_id;index"`
	Enabled      bool   `json:"enabled" gorm:"default:false"`
	UploadMbps   int    `json:"uploadMbps" gorm:"column:upload_mbps;default:0"`
	DownloadMbps int    `json:"downloadMbps" gorm:"column:download_mbps;default:0"`
	ExternalPort *int   `json:"externalPort" gorm:"column:external_port;uniqueIndex:idx_client_speed_external_port"`
	InternalPort *int   `json:"internalPort" gorm:"column:internal_port;uniqueIndex:idx_client_speed_internal_port"`
	LastError    string `json:"lastError,omitempty" gorm:"column:last_error"`
	CreatedAt    int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt    int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (ClientSpeedLimit) TableName() string { return "client_speed_limits" }
