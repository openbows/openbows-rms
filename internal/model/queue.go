package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	QueueStatusWaiting   = "waiting"
	QueueStatusCalled    = "called"
	QueueStatusSeated    = "seated"
	QueueStatusCancelled = "cancelled"
)

type QueueTicket struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	BusinessDay  time.Time      `gorm:"column:business_day;type:date;index;not null;comment:营业日" json:"business_day"`
	TicketNo     int            `gorm:"column:ticket_no;type:int;not null;comment:排队号" json:"ticket_no"`
	GuestCount   int            `gorm:"column:guest_count;type:int;not null;default:1;comment:就餐人数" json:"guest_count"`
	Phone        string         `gorm:"column:phone;type:varchar(32);not null;default:'';comment:联系电话" json:"phone"`
	Status       string         `gorm:"column:status;type:varchar(32);not null;default:waiting;comment:状态" json:"status"`
	OrderID      *int64         `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	TableID      *int64         `gorm:"column:table_id;comment:入座桌台ID" json:"table_id"`
	Remark       string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CalledAt     *time.Time     `gorm:"column:called_at;comment:叫号时间" json:"called_at"`
	SeatedAt     *time.Time     `gorm:"column:seated_at;comment:入座时间" json:"seated_at"`
	CancelledAt  *time.Time     `gorm:"column:cancelled_at;comment:取消时间" json:"cancelled_at"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (QueueTicket) TableName() string { return "queue_tickets" }
