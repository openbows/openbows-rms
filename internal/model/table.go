package model

import (
	"time"

	"gorm.io/gorm"
)

// 桌台状态枚举（与后续挂单/预订约定一致）
const (
	TableStatusIdle     = "idle"     // 空闲
	TableStatusOccupied = "occupied" // 占用
	TableStatusReserved = "reserved" // 预留
)

type Area struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:区域名称" json:"name"`
	SortNo    int            `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Area) TableName() string { return "areas" }

// DiningTable 桌台（Go 类型名避开关键字 Table）
type DiningTable struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID        int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	AreaID         int64          `gorm:"column:area_id;index;not null;comment:区域ID" json:"area_id"`
	Code           string         `gorm:"column:code;type:varchar(64);not null;comment:桌台编号" json:"code"`
	Name           string         `gorm:"column:name;type:varchar(128);not null;default:'';comment:桌台名称" json:"name"`
	Capacity       int            `gorm:"column:capacity;type:int;not null;default:4;comment:可容纳人数" json:"capacity"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:idle;comment:状态" json:"status"`
	SortNo         int            `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	CurrentOrderID *int64         `gorm:"column:current_order_id;comment:当前挂单订单ID" json:"current_order_id"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (DiningTable) TableName() string { return "tables" }
