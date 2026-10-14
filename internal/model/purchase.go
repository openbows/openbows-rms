package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	PurchaseOrderStatusDraft           = "draft"
	PurchaseOrderStatusSubmitted       = "submitted"
	PurchaseOrderStatusPartialReceived = "partial_received"
	PurchaseOrderStatusReceived        = "received"
	PurchaseOrderStatusCancelled       = "cancelled"

	PurchaseReceiptStatusDraft  = "draft"
	PurchaseReceiptStatusPosted = "posted"
)

type Supplier struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Code        string         `gorm:"column:code;type:varchar(64);not null;comment:供应商编码" json:"code"`
	Name        string         `gorm:"column:name;type:varchar(128);not null;comment:供应商名称" json:"name"`
	ContactName string         `gorm:"column:contact_name;type:varchar(64);not null;default:'';comment:联系人" json:"contact_name"`
	Phone       string         `gorm:"column:phone;type:varchar(32);not null;default:'';comment:电话" json:"phone"`
	Address     string         `gorm:"column:address;type:varchar(255);not null;default:'';comment:地址" json:"address"`
	Status      bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	Remark      string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Supplier) TableName() string { return "suppliers" }

type PurchaseOrder struct {
	ID           int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID      int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	SupplierID   int64           `gorm:"column:supplier_id;index;not null;comment:供应商ID" json:"supplier_id"`
	OrderNo      string          `gorm:"column:order_no;type:varchar(64);not null;comment:采购单号" json:"order_no"`
	Status       string          `gorm:"column:status;type:varchar(32);not null;default:draft;comment:状态" json:"status"`
	TotalAmount  decimal.Decimal `gorm:"column:total_amount;type:numeric(10,2);not null;default:0;comment:订单金额合计" json:"total_amount"`
	Remark       string          `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy    int64           `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	SubmittedBy  int64           `gorm:"column:submitted_by;not null;default:0;comment:提交人" json:"submitted_by"`
	CancelledBy  int64           `gorm:"column:cancelled_by;not null;default:0;comment:取消人" json:"cancelled_by"`
	SubmittedAt  *time.Time      `gorm:"column:submitted_at;comment:提交时间" json:"submitted_at"`
	CancelledAt  *time.Time      `gorm:"column:cancelled_at;comment:取消时间" json:"cancelled_at"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

type PurchaseOrderItem struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	OrderID     int64           `gorm:"column:order_id;index;not null;comment:采购订单ID" json:"order_id"`
	MaterialID  int64           `gorm:"column:material_id;index;not null;comment:物料ID" json:"material_id"`
	Qty         decimal.Decimal `gorm:"column:qty;type:numeric(12,3);not null;comment:采购数量" json:"qty"`
	UnitPrice   decimal.Decimal `gorm:"column:unit_price;type:numeric(10,2);not null;default:0;comment:采购单价" json:"unit_price"`
	Amount      decimal.Decimal `gorm:"column:amount;type:numeric(10,2);not null;default:0;comment:行金额" json:"amount"`
	ReceivedQty decimal.Decimal `gorm:"column:received_qty;type:numeric(12,3);not null;default:0;comment:已收货数量" json:"received_qty"`
	Remark      string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (PurchaseOrderItem) TableName() string { return "purchase_order_items" }

type PurchaseReceipt struct {
	ID              int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID         int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	PurchaseOrderID int64          `gorm:"column:purchase_order_id;index;not null;comment:采购订单ID" json:"purchase_order_id"`
	SupplierID      int64          `gorm:"column:supplier_id;index;not null;comment:供应商ID" json:"supplier_id"`
	ReceiptNo       string         `gorm:"column:receipt_no;type:varchar(64);not null;comment:收货单号" json:"receipt_no"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:draft;comment:状态" json:"status"`
	StockDocID      *int64         `gorm:"column:stock_doc_id;index;comment:关联库存入库单ID" json:"stock_doc_id"`
	Remark          string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy       int64          `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	PostedBy        int64          `gorm:"column:posted_by;not null;default:0;comment:入库确认人" json:"posted_by"`
	PostedAt        *time.Time     `gorm:"column:posted_at;comment:入库确认时间" json:"posted_at"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (PurchaseReceipt) TableName() string { return "purchase_receipts" }

type PurchaseReceiptItem struct {
	ID                   int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID              int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	ReceiptID            int64           `gorm:"column:receipt_id;index;not null;comment:收货单ID" json:"receipt_id"`
	PurchaseOrderItemID  int64           `gorm:"column:purchase_order_item_id;index;not null;comment:采购订单明细ID" json:"purchase_order_item_id"`
	MaterialID           int64           `gorm:"column:material_id;index;not null;comment:物料ID" json:"material_id"`
	Qty                  decimal.Decimal `gorm:"column:qty;type:numeric(12,3);not null;comment:本次收货数量" json:"qty"`
	UnitPrice            decimal.Decimal `gorm:"column:unit_price;type:numeric(10,2);not null;default:0;comment:单价快照" json:"unit_price"`
	Remark               string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt            time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (PurchaseReceiptItem) TableName() string { return "purchase_receipt_items" }
