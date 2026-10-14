package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	ReservationStatusPendingConfirm = "pending_confirm"
	ReservationStatusConfirmed      = "confirmed"
	ReservationStatusSeated         = "seated"
	ReservationStatusCancelled      = "cancelled"
	ReservationStatusNoShow         = "no_show"

	ReservationChannelPhone    = "phone"
	ReservationChannelWalkIn   = "walk_in"
	ReservationChannelWechat   = "wechat"
	ReservationChannelMeituan  = "meituan"
	ReservationChannelDianping = "dianping"
	ReservationChannelOnline   = "online"
	ReservationChannelOther    = "other"
)

type Reservation struct {
	ID            int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID       int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	TableID       *int64          `gorm:"column:table_id;index;comment:预订桌台ID" json:"table_id"`
	CustomerName  string          `gorm:"column:customer_name;type:varchar(128);not null;comment:顾客姓名" json:"customer_name"`
	Phone         string          `gorm:"column:phone;type:varchar(32);not null;default:'';comment:联系电话" json:"phone"`
	GuestCount    int             `gorm:"column:guest_count;type:int;not null;default:2;comment:就餐人数" json:"guest_count"`
	ReserveAt     time.Time       `gorm:"column:reserve_at;index;not null;comment:预订到店时间" json:"reserve_at"`
	Channel       string          `gorm:"column:channel;type:varchar(32);not null;default:phone;comment:渠道" json:"channel"`
	Status        string          `gorm:"column:status;type:varchar(32);not null;default:pending_confirm;comment:状态" json:"status"`
	Remark        string          `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	DepositAmount decimal.Decimal `gorm:"column:deposit_amount;type:numeric(10,2);not null;default:0;comment:订金金额" json:"deposit_amount"`
	DepositPaid   bool            `gorm:"column:deposit_paid;type:boolean;not null;default:false;comment:订金是否已付" json:"deposit_paid"`
	OrderID       *int64          `gorm:"column:order_id;comment:入座后关联订单ID" json:"order_id"`
	CancelReason  string          `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:取消/违约原因" json:"cancel_reason"`
	ConfirmedAt   *time.Time      `gorm:"column:confirmed_at;comment:确认时间" json:"confirmed_at"`
	SeatedAt      *time.Time      `gorm:"column:seated_at;comment:入座时间" json:"seated_at"`
	CancelledAt   *time.Time      `gorm:"column:cancelled_at;comment:取消时间" json:"cancelled_at"`
	NoShowAt      *time.Time      `gorm:"column:no_show_at;comment:违约时间" json:"no_show_at"`
	CreatedBy     int64           `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	UpdatedBy     int64           `gorm:"column:updated_by;not null;default:0;comment:更新人" json:"updated_by"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Reservation) TableName() string { return "reservations" }
