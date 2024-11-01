package db

import "time"

type GormConfig struct {
	MaxLifetime       int  `mapstructure:"max_life_time"`
	MaxOpenConns      int  `mapstructure:"max_open_conns"`
	MaxIdleConns      int  `mapstructure:"max_idle_conns"`
	EnableAutoMigrate bool `mapstructure:"enable_auto_migrate"`
}

type DbModel interface {
	TableName() string
}

// TableSlice gorm自动建表模型列表
var TableSlice = make([]interface{}, 0)

type CommonField struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
