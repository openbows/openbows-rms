package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	OrderStatusPendingConfirm = "pending_confirm" // 待确认
	OrderStatusCooking        = "cooking"         // 制作中
	OrderStatusPendingPay     = "pending_pay"     // 待结账
	OrderStatusCompleted      = "completed"       // 已完成
	OrderStatusCancelled      = "cancelled"       // 已取消

	OrderItemStatusNormal   = "normal"
	OrderItemStatusReturned = "returned"

	FeeTypeDiscount = "discount"
	FeeTypeService  = "service"
	FeeTypeTax      = "tax"
	FeeTypeOther    = "other"

	OrderChannelDineIn     = "dine_in"
	OrderChannelOnlineH5   = "online_h5"
	OrderChannelOnlineScan = "online_scan"
	OrderChannelTakeout    = "takeout"
)

type Order struct {
	ID             int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID        int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	TableID        int64           `gorm:"column:table_id;index;not null;comment:桌台ID" json:"table_id"`
	OrderNo        string          `gorm:"column:order_no;type:varchar(64);not null;comment:订单号" json:"order_no"`
	Channel        string          `gorm:"column:channel;type:varchar(32);not null;default:dine_in;comment:下单渠道" json:"channel"`
	MemberID       *int64          `gorm:"column:member_id;index;comment:关联会员ID" json:"member_id"`
	Status         string          `gorm:"column:status;type:varchar(32);not null;default:pending_confirm;comment:状态" json:"status"`
	GuestCount     int             `gorm:"column:guest_count;type:int;not null;default:1;comment:就餐人数" json:"guest_count"`
	SubtotalAmount decimal.Decimal `gorm:"column:subtotal_amount;type:numeric(10,2);not null;default:0;comment:菜品小计" json:"subtotal_amount"`
	FeeAmount      decimal.Decimal `gorm:"column:fee_amount;type:numeric(10,2);not null;default:0;comment:费用合计" json:"fee_amount"`
	DiscountAmount decimal.Decimal `gorm:"column:discount_amount;type:numeric(10,2);not null;default:0;comment:优惠合计" json:"discount_amount"`
	PayableAmount  decimal.Decimal `gorm:"column:payable_amount;type:numeric(10,2);not null;default:0;comment:应付金额" json:"payable_amount"`
	Remark         string          `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy      int64           `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	UpdatedBy      int64           `gorm:"column:updated_by;not null;default:0;comment:更新人" json:"updated_by"`
	OpenedAt       time.Time       `gorm:"column:opened_at;comment:开台时间" json:"opened_at"`
	ClosedAt       *time.Time      `gorm:"column:closed_at;comment:关台时间" json:"closed_at"`
	CreatedAt      time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

// OnlineOrderExtra 线上订单扩展（顾客端 access_token 与联系人）
type OnlineOrderExtra struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID      int64          `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	ContactName  string         `gorm:"column:contact_name;type:varchar(128);not null;default:'';comment:联系人姓名" json:"contact_name"`
	ContactPhone string         `gorm:"column:contact_phone;type:varchar(32);not null;default:'';comment:联系人电话" json:"contact_phone"`
	AccessToken  string         `gorm:"column:access_token;type:varchar(64);not null;comment:顾客端查询凭证" json:"-"`
	Source       string         `gorm:"column:source;type:varchar(32);not null;default:h5;comment:来源" json:"source"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (OnlineOrderExtra) TableName() string { return "online_order_extras" }

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID         int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID    int64           `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	DishID     int64           `gorm:"column:dish_id;not null;comment:菜品ID" json:"dish_id"`
	DishName   string          `gorm:"column:dish_name;type:varchar(128);not null;comment:菜品名称快照" json:"dish_name"`
	DishSpecID *int64          `gorm:"column:dish_spec_id;comment:规格ID" json:"dish_spec_id"`
	SpecName   string          `gorm:"column:spec_name;type:varchar(128);not null;default:'';comment:规格名称快照" json:"spec_name"`
	UnitPrice  decimal.Decimal `gorm:"column:unit_price;type:numeric(10,2);not null;default:0;comment:单价" json:"unit_price"`
	Qty        int             `gorm:"column:qty;type:int;not null;default:1;comment:数量" json:"qty"`
	Amount        decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;default:0;comment:行金额" json:"amount"`
	Status        string          `gorm:"column:status;type:varchar(32);not null;default:normal;comment:状态" json:"status"`
	KitchenStatus string          `gorm:"column:kitchen_status;type:varchar(32);not null;default:pending;comment:出品状态" json:"kitchen_status"`
	Remark        string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	SortNo        int             `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	CreatedAt  time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (OrderItem) TableName() string { return "order_items" }

type OrderItemModifier struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderItemID int64           `gorm:"column:order_item_id;index;not null;comment:订单明细ID" json:"order_item_id"`
	OptionID    *int64          `gorm:"column:option_id;comment:加料选项ID" json:"option_id"`
	OptionName  string          `gorm:"column:option_name;type:varchar(128);not null;comment:加料名称快照" json:"option_name"`
	Price       decimal.Decimal `gorm:"column:price;type:numeric(10,2);not null;default:0;comment:加料单价" json:"price"`
	Qty         int             `gorm:"column:qty;type:int;not null;default:1;comment:数量" json:"qty"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (OrderItemModifier) TableName() string { return "order_item_modifiers" }

type OrderStatusLog struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID      int64          `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	FromStatus   string         `gorm:"column:from_status;type:varchar(32);not null;default:'';comment:原状态" json:"from_status"`
	ToStatus     string         `gorm:"column:to_status;type:varchar(32);not null;comment:新状态" json:"to_status"`
	OperatorID   int64          `gorm:"column:operator_id;not null;default:0;comment:操作人ID" json:"operator_id"`
	OperatorName string         `gorm:"column:operator_name;type:varchar(64);not null;default:'';comment:操作人名称" json:"operator_name"`
	Remark       string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (OrderStatusLog) TableName() string { return "order_status_logs" }

type OrderFeeItem struct {
	ID        int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID   int64           `gorm:"column:order_id;index;not null;comment:订单ID" json:"order_id"`
	FeeType   string          `gorm:"column:fee_type;type:varchar(32);not null;comment:费用类型" json:"fee_type"`
	Name      string          `gorm:"column:name;type:varchar(128);not null;comment:分项名称" json:"name"`
	Amount    decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;default:0;comment:金额" json:"amount"`
	Remark    string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (OrderFeeItem) TableName() string { return "order_fee_items" }
