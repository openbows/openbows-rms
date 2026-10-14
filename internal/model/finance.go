package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	SettlementStatusClosed = "closed"

	ReconStatusMatched    = "matched"
	ReconStatusMismatched = "mismatched"

	ReconDiffMatch            = "match"
	ReconDiffAmountMismatch   = "amount_mismatch"
	ReconDiffMissingInChannel = "missing_in_channel"
	ReconDiffMissingInSystem  = "missing_in_system"
	ReconDiffTotalDiff        = "total_diff"
)

type Settlement struct {
	ID            int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID       int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	BusinessDay   time.Time       `gorm:"column:business_day;type:date;index;not null;comment:营业日" json:"business_day"`
	OrderCount    int             `gorm:"column:order_count;type:int;not null;default:0;comment:关联订单数" json:"order_count"`
	PaymentCount  int             `gorm:"column:payment_count;type:int;not null;default:0;comment:支付笔数" json:"payment_count"`
	TotalAmount   decimal.Decimal `gorm:"column:total_amount;type:numeric(10,2);not null;default:0;comment:支付合计" json:"total_amount"`
	CashAmount    decimal.Decimal `gorm:"column:cash_amount;type:numeric(10,2);not null;default:0;comment:现金合计" json:"cash_amount"`
	WechatAmount  decimal.Decimal `gorm:"column:wechat_amount;type:numeric(10,2);not null;default:0;comment:微信合计" json:"wechat_amount"`
	AlipayAmount  decimal.Decimal `gorm:"column:alipay_amount;type:numeric(10,2);not null;default:0;comment:支付宝合计" json:"alipay_amount"`
	OtherAmount   decimal.Decimal `gorm:"column:other_amount;type:numeric(10,2);not null;default:0;comment:其他渠道合计" json:"other_amount"`
	Status        string          `gorm:"column:status;type:varchar(32);not null;default:closed;comment:状态" json:"status"`
	Remark        string          `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy     int64           `gorm:"column:created_by;not null;default:0;comment:日结人" json:"created_by"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Settlement) TableName() string { return "settlements" }

type SettlementItem struct {
	ID           int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	SettlementID int64           `gorm:"column:settlement_id;index;not null;comment:日结ID" json:"settlement_id"`
	PaymentID    int64           `gorm:"column:payment_id;index;not null;comment:支付ID" json:"payment_id"`
	OrderID      int64           `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	Channel      string          `gorm:"column:channel;type:varchar(32);not null;comment:支付渠道" json:"channel"`
	Amount       decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;comment:支付金额" json:"amount"`
	PaidAt       time.Time       `gorm:"column:paid_at;not null;comment:支付时间" json:"paid_at"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (SettlementItem) TableName() string { return "settlement_items" }

type Reconciliation struct {
	ID            int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID       int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	BusinessDay   time.Time       `gorm:"column:business_day;type:date;index;not null;comment:营业日" json:"business_day"`
	Channel       string          `gorm:"column:channel;type:varchar(32);not null;comment:支付渠道" json:"channel"`
	SystemAmount  decimal.Decimal `gorm:"column:system_amount;type:numeric(10,2);not null;default:0;comment:系统支付合计" json:"system_amount"`
	ChannelAmount decimal.Decimal `gorm:"column:channel_amount;type:numeric(10,2);not null;default:0;comment:渠道账单合计" json:"channel_amount"`
	DiffAmount    decimal.Decimal `gorm:"column:diff_amount;type:numeric(10,2);not null;default:0;comment:差异金额" json:"diff_amount"`
	SystemCount   int             `gorm:"column:system_count;type:int;not null;default:0;comment:系统笔数" json:"system_count"`
	ChannelCount  int             `gorm:"column:channel_count;type:int;not null;default:0;comment:渠道笔数" json:"channel_count"`
	Status        string          `gorm:"column:status;type:varchar(32);not null;default:matched;comment:状态" json:"status"`
	Remark        string          `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy     int64           `gorm:"column:created_by;not null;default:0;comment:对账人" json:"created_by"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Reconciliation) TableName() string { return "reconciliations" }

type ReconciliationItem struct {
	ID               int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID          int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	ReconciliationID int64           `gorm:"column:reconciliation_id;index;not null;comment:对账ID" json:"reconciliation_id"`
	PaymentID        *int64          `gorm:"column:payment_id;index;comment:系统支付ID" json:"payment_id"`
	OrderID          *int64          `gorm:"column:order_id;index;comment:订单ID" json:"order_id"`
	OutTradeNo       string          `gorm:"column:out_trade_no;type:varchar(128);not null;default:'';comment:第三方交易号" json:"out_trade_no"`
	SystemAmount     decimal.Decimal `gorm:"column:system_amount;type:numeric(10,2);not null;default:0;comment:系统金额" json:"system_amount"`
	ChannelAmount    decimal.Decimal `gorm:"column:channel_amount;type:numeric(10,2);not null;default:0;comment:渠道金额" json:"channel_amount"`
	DiffAmount       decimal.Decimal `gorm:"column:diff_amount;type:numeric(10,2);not null;default:0;comment:差异金额" json:"diff_amount"`
	DiffType         string          `gorm:"column:diff_type;type:varchar(32);not null;comment:差异类型" json:"diff_type"`
	Remark           string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (ReconciliationItem) TableName() string { return "reconciliation_items" }
