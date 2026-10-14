package model

import (
	"time"

	"gorm.io/gorm"
)

type Store struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	Code      string         `gorm:"column:code;type:varchar(64);not null;comment:门店编码" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:门店名称" json:"name"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Store) TableName() string { return "stores" }
