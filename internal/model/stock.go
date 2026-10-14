package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	StockDocTypeInbound  = "inbound"
	StockDocTypeOutbound = "outbound"
	StockDocTypeAdjust   = "adjust"

	StockDocStatusDraft  = "draft"
	StockDocStatusPosted = "posted"
	StockDocStatusVoided = "voided"
)

type Material struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Code      string         `gorm:"column:code;type:varchar(64);not null;comment:物料编码" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:物料名称" json:"name"`
	Unit      string         `gorm:"column:unit;type:varchar(32);not null;default:个;comment:计量单位" json:"unit"`
	Category  string         `gorm:"column:category;type:varchar(64);not null;default:'';comment:分类" json:"category"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	Remark    string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Material) TableName() string { return "materials" }

type StockBalance struct {
	ID         int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	MaterialID int64           `gorm:"column:material_id;index;not null;comment:物料ID" json:"material_id"`
	Qty        decimal.Decimal `gorm:"column:qty;type:numeric(12,3);not null;default:0;comment:结存数量" json:"qty"`
	CreatedAt  time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (StockBalance) TableName() string { return "stock_balances" }

type StockDoc struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	DocNo      string         `gorm:"column:doc_no;type:varchar(64);not null;comment:单据号" json:"doc_no"`
	DocType    string         `gorm:"column:doc_type;type:varchar(32);not null;comment:单据类型" json:"doc_type"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:draft;comment:状态" json:"status"`
	Remark     string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	VoidReason string         `gorm:"column:void_reason;type:varchar(255);not null;default:'';comment:作废原因" json:"void_reason"`
	CreatedBy  int64          `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	PostedBy   int64          `gorm:"column:posted_by;not null;default:0;comment:审核人" json:"posted_by"`
	VoidedBy   int64          `gorm:"column:voided_by;not null;default:0;comment:作废人" json:"voided_by"`
	PostedAt   *time.Time     `gorm:"column:posted_at;comment:审核时间" json:"posted_at"`
	VoidedAt   *time.Time     `gorm:"column:voided_at;comment:作废时间" json:"voided_at"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (StockDoc) TableName() string { return "stock_docs" }

type StockDocItem struct {
	ID         int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	DocID      int64           `gorm:"column:doc_id;index;not null;comment:单据ID" json:"doc_id"`
	MaterialID int64           `gorm:"column:material_id;index;not null;comment:物料ID" json:"material_id"`
	Qty        decimal.Decimal `gorm:"column:qty;type:numeric(12,3);not null;comment:数量" json:"qty"`
	UnitCost   decimal.Decimal `gorm:"column:unit_cost;type:numeric(10,2);not null;default:0;comment:单位成本" json:"unit_cost"`
	Remark     string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt  time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (StockDocItem) TableName() string { return "stock_doc_items" }
