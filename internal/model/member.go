package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	MemberGenderUnknown = "unknown"
	MemberGenderMale    = "male"
	MemberGenderFemale  = "female"

	WalletChangeRecharge = "recharge"
	WalletChangeConsume  = "consume"
	WalletChangeAdjust   = "adjust"

	PointChangeEarn   = "earn"
	PointChangeSpend  = "spend"
	PointChangeAdjust = "adjust"
)

type Member struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Phone     string         `gorm:"column:phone;type:varchar(32);not null;comment:手机号" json:"phone"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;default:'';comment:姓名" json:"name"`
	Gender    string         `gorm:"column:gender;type:varchar(16);not null;default:unknown;comment:性别" json:"gender"`
	Birthday  *time.Time     `gorm:"column:birthday;type:date;comment:生日" json:"birthday"`
	Level     string         `gorm:"column:level;type:varchar(32);not null;default:normal;comment:会员等级" json:"level"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	Remark    string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy int64          `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	UpdatedBy int64          `gorm:"column:updated_by;not null;default:0;comment:更新人" json:"updated_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Member) TableName() string { return "members" }

type MemberWallet struct {
	ID        int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	MemberID  int64           `gorm:"column:member_id;index;not null;comment:会员ID" json:"member_id"`
	Balance   decimal.Decimal `gorm:"column:balance;type:numeric(10,2);not null;default:0;comment:余额" json:"balance"`
	Points    int             `gorm:"column:points;type:int;not null;default:0;comment:积分" json:"points"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (MemberWallet) TableName() string { return "member_wallets" }

type MemberWalletLog struct {
	ID            int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID       int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	MemberID      int64           `gorm:"column:member_id;index;not null;comment:会员ID" json:"member_id"`
	WalletID      int64           `gorm:"column:wallet_id;index;not null;comment:钱包ID" json:"wallet_id"`
	ChangeType    string          `gorm:"column:change_type;type:varchar(32);not null;comment:变动类型" json:"change_type"`
	Amount        decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;comment:变动金额" json:"amount"`
	BalanceBefore decimal.Decimal `gorm:"column:balance_before;type:numeric(10,2);not null;comment:变动前余额" json:"balance_before"`
	BalanceAfter  decimal.Decimal `gorm:"column:balance_after;type:numeric(10,2);not null;comment:变动后余额" json:"balance_after"`
	BizNo         string          `gorm:"column:biz_no;type:varchar(64);not null;default:'';comment:业务单号" json:"biz_no"`
	OrderID       *int64          `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	Remark        string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	OperatorID    int64           `gorm:"column:operator_id;not null;default:0;comment:操作人ID" json:"operator_id"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (MemberWalletLog) TableName() string { return "member_wallet_logs" }

type MemberPointLog struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	MemberID     int64          `gorm:"column:member_id;index;not null;comment:会员ID" json:"member_id"`
	WalletID     int64          `gorm:"column:wallet_id;index;not null;comment:钱包ID" json:"wallet_id"`
	ChangeType   string         `gorm:"column:change_type;type:varchar(32);not null;comment:变动类型" json:"change_type"`
	Points       int            `gorm:"column:points;type:int;not null;comment:变动积分" json:"points"`
	PointsBefore int            `gorm:"column:points_before;type:int;not null;comment:变动前积分" json:"points_before"`
	PointsAfter  int            `gorm:"column:points_after;type:int;not null;comment:变动后积分" json:"points_after"`
	BizNo        string         `gorm:"column:biz_no;type:varchar(64);not null;default:'';comment:业务单号" json:"biz_no"`
	OrderID      *int64         `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	Remark       string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	OperatorID   int64          `gorm:"column:operator_id;not null;default:0;comment:操作人ID" json:"operator_id"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (MemberPointLog) TableName() string { return "member_point_logs" }
