package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	KitchenStatusPending = "pending"
	KitchenStatusCooking = "cooking"
	KitchenStatusDone    = "done"
)

type KdsTicket struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	StationID int64          `gorm:"column:station_id;index;not null;comment:档口ID" json:"station_id"`
	OrderID   int64          `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	OrderNo   string         `gorm:"column:order_no;type:varchar(64);not null;comment:订单号快照" json:"order_no"`
	TableID   int64          `gorm:"column:table_id;not null;comment:桌台ID" json:"table_id"`
	Status    string         `gorm:"column:status;type:varchar(32);not null;default:pending;comment:厨单状态" json:"status"`
	StartedAt *time.Time     `gorm:"column:started_at;comment:开始制作时间" json:"started_at"`
	DoneAt    *time.Time     `gorm:"column:done_at;comment:出餐完成时间" json:"done_at"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (KdsTicket) TableName() string { return "kds_tickets" }

type KdsTicketItem struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	TicketID    int64          `gorm:"column:ticket_id;index;not null;comment:厨单ID" json:"ticket_id"`
	OrderItemID int64          `gorm:"column:order_item_id;index;not null;comment:订单明细ID" json:"order_item_id"`
	DishName    string         `gorm:"column:dish_name;type:varchar(128);not null;comment:菜品名称快照" json:"dish_name"`
	SpecName    string         `gorm:"column:spec_name;type:varchar(128);not null;default:'';comment:规格名称快照" json:"spec_name"`
	Qty         int            `gorm:"column:qty;type:int;not null;default:1;comment:数量" json:"qty"`
	Status      string         `gorm:"column:status;type:varchar(32);not null;default:pending;comment:出品状态" json:"status"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (KdsTicketItem) TableName() string { return "kds_ticket_items" }
