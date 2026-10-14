package finance

import (
	"context"
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

func (r *Repository) GetSettlementByDay(ctx context.Context, storeID int64, day time.Time) (*model.Settlement, error) {
	var m model.Settlement
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND business_day = ?", storeID, truncateDate(day)).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) CreateSettlementTx(ctx context.Context, tx *gorm.DB, m *model.Settlement) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateSettlementItemTx(ctx context.Context, tx *gorm.DB, m *model.SettlementItem) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetSettlementByID(ctx context.Context, id int64) (*model.Settlement, error) {
	var m model.Settlement
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListSettlementItems(ctx context.Context, settlementID int64) ([]model.SettlementItem, error) {
	var list []model.SettlementItem
	err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Order("id ASC").Find(&list).Error
	return list, err
}

type SettlementListFilter struct {
	StoreID     int64
	BusinessDay *time.Time
	Page        int
	PageSize    int
}

func (r *Repository) ListSettlements(ctx context.Context, f SettlementListFilter) ([]model.Settlement, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Settlement{}).Where("store_id = ?", f.StoreID)
	if f.BusinessDay != nil {
		q = q.Where("business_day = ?", truncateDate(*f.BusinessDay))
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
	var list []model.Settlement
	err := q.Order("business_day DESC, id DESC").
		Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) CreateReconciliationTx(ctx context.Context, tx *gorm.DB, m *model.Reconciliation) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateReconciliationItemTx(ctx context.Context, tx *gorm.DB, m *model.ReconciliationItem) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetReconciliationByID(ctx context.Context, id int64) (*model.Reconciliation, error) {
	var m model.Reconciliation
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListReconciliationItems(ctx context.Context, reconID int64) ([]model.ReconciliationItem, error) {
	var list []model.ReconciliationItem
	err := r.db.WithContext(ctx).Where("reconciliation_id = ?", reconID).Order("id ASC").Find(&list).Error
	return list, err
}

type ReconListFilter struct {
	StoreID     int64
	BusinessDay *time.Time
	Channel     string
	Status      string
	Page        int
	PageSize    int
}

func (r *Repository) ListReconciliations(ctx context.Context, f ReconListFilter) ([]model.Reconciliation, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Reconciliation{}).Where("store_id = ?", f.StoreID)
	if f.BusinessDay != nil {
		q = q.Where("business_day = ?", truncateDate(*f.BusinessDay))
	}
	if f.Channel != "" {
		q = q.Where("channel = ?", f.Channel)
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
	var list []model.Reconciliation
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

type DiffListFilter struct {
	StoreID     int64
	BusinessDay *time.Time
	Channel     string
	Page        int
	PageSize    int
}

func (r *Repository) ListDiffItems(ctx context.Context, f DiffListFilter) ([]model.ReconciliationItem, int64, error) {
	q := r.db.WithContext(ctx).Table("reconciliation_items AS i").
		Where("i.deleted_at IS NULL AND i.store_id = ? AND i.diff_type <> ?", f.StoreID, model.ReconDiffMatch)
	if f.Channel != "" || f.BusinessDay != nil {
		q = q.Joins("JOIN reconciliations r ON r.id = i.reconciliation_id AND r.deleted_at IS NULL")
		if f.Channel != "" {
			q = q.Where("r.channel = ?", f.Channel)
		}
		if f.BusinessDay != nil {
			q = q.Where("r.business_day = ?", truncateDate(*f.BusinessDay))
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	var list []model.ReconciliationItem
	err := q.Select("i.*").Order("i.id DESC").
		Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Scan(&list).Error
	return list, total, err
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
