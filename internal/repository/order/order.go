package order

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) CreateOrderNo(storeID int64) string {
	return fmt.Sprintf("O%d%s%04d", storeID, time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var m model.Order
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.Order, error) {
	var m model.Order
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type ListFilter struct {
	StoreID  int64
	TableID  int64
	Status   string
	Page     int
	PageSize int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]model.Order, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("store_id = ?", f.StoreID)
	if f.TableID > 0 {
		q = q.Where("table_id = ?", f.TableID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Order
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListItems(ctx context.Context, orderID int64) ([]model.OrderItem, error) {
	var list []model.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListItemsTx(ctx context.Context, tx *gorm.DB, orderID int64) ([]model.OrderItem, error) {
	var list []model.OrderItem
	err := tx.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListModifiersByItemIDs(ctx context.Context, itemIDs []int64) ([]model.OrderItemModifier, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}
	var list []model.OrderItemModifier
	err := r.db.WithContext(ctx).Where("order_item_id IN ?", itemIDs).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListFeeItems(ctx context.Context, orderID int64) ([]model.OrderFeeItem, error) {
	var list []model.OrderFeeItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListStatusLogs(ctx context.Context, orderID int64) ([]model.OrderStatusLog, error) {
	var list []model.OrderStatusLog
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) GetItemByID(ctx context.Context, id int64) (*model.OrderItem, error) {
	var m model.OrderItem
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
