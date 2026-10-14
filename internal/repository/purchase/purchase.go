package purchase

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	return fmt.Sprintf("PO%d%s%04d", storeID, time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

func (r *Repository) CreateReceiptNo(storeID int64) string {
	return fmt.Sprintf("PR%d%s%04d", storeID, time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

// ---- suppliers ----

func (r *Repository) CreateSupplier(ctx context.Context, m *model.Supplier) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveSupplier(ctx context.Context, m *model.Supplier) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDeleteSupplier(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Supplier{}, id).Error
}

func (r *Repository) GetSupplierByID(ctx context.Context, id int64) (*model.Supplier, error) {
	var m model.Supplier
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type SupplierListFilter struct {
	StoreID  int64
	Keyword  string
	Status   *bool
	Page     int
	PageSize int
}

func (r *Repository) ListSuppliers(ctx context.Context, f SupplierListFilter) ([]model.Supplier, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Supplier{}).Where("store_id = ?", f.StoreID)
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.Supplier
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// ---- purchase orders ----

func (r *Repository) CreateOrderTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseOrder) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateOrderItemTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseOrderItem) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveOrderTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseOrder) error {
	return tx.WithContext(ctx).Save(m).Error
}

func (r *Repository) SaveOrderItemTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseOrderItem) error {
	return tx.WithContext(ctx).Save(m).Error
}

func (r *Repository) GetOrderByID(ctx context.Context, id int64) (*model.PurchaseOrder, error) {
	var m model.PurchaseOrder
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetOrderByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.PurchaseOrder, error) {
	var m model.PurchaseOrder
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListOrderItems(ctx context.Context, orderID int64) ([]model.PurchaseOrderItem, error) {
	var list []model.PurchaseOrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListOrderItemsTx(ctx context.Context, tx *gorm.DB, orderID int64) ([]model.PurchaseOrderItem, error) {
	var list []model.PurchaseOrderItem
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) GetOrderItemByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.PurchaseOrderItem, error) {
	var m model.PurchaseOrderItem
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type OrderListFilter struct {
	StoreID    int64
	SupplierID int64
	Status     string
	Page       int
	PageSize   int
}

func (r *Repository) ListOrders(ctx context.Context, f OrderListFilter) ([]model.PurchaseOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.PurchaseOrder{}).Where("store_id = ?", f.StoreID)
	if f.SupplierID > 0 {
		q = q.Where("supplier_id = ?", f.SupplierID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.PurchaseOrder
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// ---- receipts ----

func (r *Repository) CreateReceiptTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseReceipt) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateReceiptItemTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseReceiptItem) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveReceiptTx(ctx context.Context, tx *gorm.DB, m *model.PurchaseReceipt) error {
	return tx.WithContext(ctx).Save(m).Error
}

func (r *Repository) GetReceiptByID(ctx context.Context, id int64) (*model.PurchaseReceipt, error) {
	var m model.PurchaseReceipt
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetReceiptByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.PurchaseReceipt, error) {
	var m model.PurchaseReceipt
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListReceiptItems(ctx context.Context, receiptID int64) ([]model.PurchaseReceiptItem, error) {
	var list []model.PurchaseReceiptItem
	err := r.db.WithContext(ctx).Where("receipt_id = ?", receiptID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListReceiptItemsTx(ctx context.Context, tx *gorm.DB, receiptID int64) ([]model.PurchaseReceiptItem, error) {
	var list []model.PurchaseReceiptItem
	err := tx.WithContext(ctx).Where("receipt_id = ?", receiptID).Order("id ASC").Find(&list).Error
	return list, err
}

type ReceiptListFilter struct {
	StoreID         int64
	PurchaseOrderID int64
	Status          string
	Page            int
	PageSize        int
}

func (r *Repository) ListReceipts(ctx context.Context, f ReceiptListFilter) ([]model.PurchaseReceipt, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.PurchaseReceipt{}).Where("store_id = ?", f.StoreID)
	if f.PurchaseOrderID > 0 {
		q = q.Where("purchase_order_id = ?", f.PurchaseOrderID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.PurchaseReceipt
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}
