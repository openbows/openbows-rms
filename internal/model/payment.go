package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	PaymentChannelCash   = "cash"
	PaymentChannelWechat = "wechat"
	PaymentChannelAlipay = "alipay"
	PaymentChannelOther  = "other"
	PaymentChannelMember = "member"

	PaymentStatusSuccess = "success"
	PaymentStatusFailed  = "failed"
)

type Payment struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID     int64           `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	PaymentNo   string          `gorm:"column:payment_no;type:varchar(64);not null;comment:支付单号" json:"payment_no"`
	Channel     string          `gorm:"column:channel;type:varchar(32);not null;comment:支付渠道" json:"channel"`
	Amount      decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;comment:支付金额" json:"amount"`
	Status      string          `gorm:"column:status;type:varchar(32);not null;default:success;comment:状态" json:"status"`
	PaidAt      time.Time       `gorm:"column:paid_at;comment:支付时间" json:"paid_at"`
	OutTradeNo  string          `gorm:"column:out_trade_no;type:varchar(128);not null;default:'';comment:第三方交易号" json:"out_trade_no"`
	CallbackRaw *string         `gorm:"column:callback_raw;type:jsonb;comment:第三方回调原文" json:"callback_raw,omitempty"`
	OperatorID  int64           `gorm:"column:operator_id;not null;default:0;comment:操作人ID" json:"operator_id"`
	Remark      string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Payment) TableName() string { return "payments" }
