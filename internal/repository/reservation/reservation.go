package reservation

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

func (r *Repository) Create(ctx context.Context, m *model.Reservation) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateTx(ctx context.Context, tx *gorm.DB, m *model.Reservation) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) Save(ctx context.Context, m *model.Reservation) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SaveTx(ctx context.Context, tx *gorm.DB, m *model.Reservation) error {
	return tx.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Reservation{}, id).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Reservation, error) {
	var m model.Reservation
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.Reservation, error) {
	var m model.Reservation
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type ListFilter struct {
	StoreID   int64
	TableID   int64
	Status    string
	Channel   string
	Keyword   string
	ReserveFrom *time.Time
	ReserveTo   *time.Time
	Page      int
	PageSize  int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]model.Reservation, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Reservation{}).Where("store_id = ?", f.StoreID)
	if f.TableID > 0 {
		q = q.Where("table_id = ?", f.TableID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Channel != "" {
		q = q.Where("channel = ?", f.Channel)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("customer_name ILIKE ? OR phone ILIKE ?", like, like)
	}
	if f.ReserveFrom != nil {
		q = q.Where("reserve_at >= ?", *f.ReserveFrom)
	}
	if f.ReserveTo != nil {
		q = q.Where("reserve_at < ?", *f.ReserveTo)
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
	var list []model.Reservation
	err := q.Order("reserve_at ASC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}
