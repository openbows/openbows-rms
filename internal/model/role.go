package model

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Code      string         `gorm:"column:code;type:varchar(64);not null;comment:角色编码" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:角色名称" json:"name"`
	Remark    string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	Code      string         `gorm:"column:code;type:varchar(128);not null;comment:权限编码" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:权限名称" json:"name"`
	Module    string         `gorm:"column:module;type:varchar(64);not null;default:'';comment:所属模块" json:"module"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	RoleID       int64          `gorm:"column:role_id;index;not null;comment:角色ID" json:"role_id"`
	PermissionID int64          `gorm:"column:permission_id;index;not null;comment:权限ID" json:"permission_id"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (RolePermission) TableName() string { return "role_permissions" }

type UserRole struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	UserID    int64          `gorm:"column:user_id;index;not null;comment:用户ID" json:"user_id"`
	RoleID    int64          `gorm:"column:role_id;index;not null;comment:角色ID" json:"role_id"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (UserRole) TableName() string { return "user_roles" }
