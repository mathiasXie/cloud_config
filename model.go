package cloud_config

import (
	"time"
)

func (CloudConfig) TableName() string {
	return "cloud_configs"
}

type CloudConfig struct {
	Id          int32     `gorm:"column:id;primary_key" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;uniqueIndex:uni_config_key" json:"config_key"`
	ConfigName  string    `gorm:"column:config_name" json:"config_name"`
	ConfigValue string    `gorm:"column:config_value" json:"config_value"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	DeleteAt    time.Time `gorm:"column:delete_at" json:"delete_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func NewCloudConfig() *CloudConfig {
	return &CloudConfig{}
}
