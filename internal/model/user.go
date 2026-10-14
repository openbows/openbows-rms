package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Username     string         `gorm:"column:username;type:varchar(64);not null;comment:登录用户名" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希" json:"-"`
	Phone        *string        `gorm:"column:phone;type:varchar(32);comment:手机号" json:"phone"`
	DisplayName  string         `gorm:"column:display_name;type:varchar(64);not null;default:'';comment:显示名称" json:"display_name"`
	Status       bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (User) TableName() string { return "users" }
