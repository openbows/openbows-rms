package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Station struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Code      string         `gorm:"column:code;type:varchar(64);not null;comment:档口编码" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:档口名称" json:"name"`
	SortNo    int            `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Station) TableName() string { return "stations" }

type MenuCategory struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:分类名称" json:"name"`
	SortNo    int            `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (MenuCategory) TableName() string { return "menu_categories" }

type Dish struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID     int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	CategoryID  int64           `gorm:"column:category_id;index;not null;comment:分类ID" json:"category_id"`
	StationID   *int64          `gorm:"column:station_id;index;comment:档口ID" json:"station_id"`
	Name        string          `gorm:"column:name;type:varchar(128);not null;comment:菜品名称" json:"name"`
	Description string          `gorm:"column:description;type:varchar(512);not null;default:'';comment:菜品描述" json:"description"`
	BasePrice   decimal.Decimal `gorm:"column:base_price;type:numeric(10,2);not null;default:0;comment:基础价" json:"base_price"`
	IsOnSale    bool            `gorm:"column:is_on_sale;type:boolean;not null;default:true;comment:是否上架" json:"is_on_sale"`
	SortNo      int             `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	ImageURL    string          `gorm:"column:image_url;type:varchar(512);not null;default:'';comment:图片地址" json:"image_url"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Dish) TableName() string { return "dishes" }

type DishSpec struct {
	ID        int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	DishID    int64           `gorm:"column:dish_id;index;not null;comment:菜品ID" json:"dish_id"`
	Name      string          `gorm:"column:name;type:varchar(128);not null;comment:规格名称" json:"name"`
	Price     decimal.Decimal `gorm:"column:price;type:numeric(10,2);not null;default:0;comment:规格价格" json:"price"`
	SortNo    int             `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (DishSpec) TableName() string { return "dish_specs" }

type DishOption struct {
	ID        int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	DishID    int64           `gorm:"column:dish_id;index;not null;comment:菜品ID" json:"dish_id"`
	Name      string          `gorm:"column:name;type:varchar(128);not null;comment:加料名称" json:"name"`
	Price     decimal.Decimal `gorm:"column:price;type:numeric(10,2);not null;default:0;comment:加料价格" json:"price"`
	SortNo    int             `gorm:"column:sort_no;type:int;not null;default:0;comment:排序号" json:"sort_no"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (DishOption) TableName() string { return "dish_options" }

// DishMaterial 菜品配方（BOM）
type DishMaterial struct {
	ID         int64           `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64           `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	DishID     int64           `gorm:"column:dish_id;index;not null;comment:菜品ID" json:"dish_id"`
	MaterialID int64           `gorm:"column:material_id;index;not null;comment:物料ID" json:"material_id"`
	Qty        decimal.Decimal `gorm:"column:qty;type:numeric(12,3);not null;comment:单份用量" json:"qty"`
	Remark     string          `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedAt  time.Time       `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (DishMaterial) TableName() string { return "dish_materials" }
