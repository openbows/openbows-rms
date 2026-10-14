package queue

import (
	"context"
	"errors"
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

func (r *Repository) CreateTx(ctx context.Context, tx *gorm.DB, m *model.QueueTicket) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.QueueTicket, error) {
	var m model.QueueTicket
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.QueueTicket, error) {
	var m model.QueueTicket
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) SaveTx(ctx context.Context, tx *gorm.DB, m *model.QueueTicket) error {
	return tx.WithContext(ctx).Save(m).Error
}

// NextTicketNoTx 取「门店+营业日」下一号（锁当日最大号行，防并发撞号）
func (r *Repository) NextTicketNoTx(ctx context.Context, tx *gorm.DB, storeID int64, day time.Time) (int, error) {
	day = truncateDate(day)
	var last model.QueueTicket
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("store_id = ? AND business_day = ?", storeID, day).
		Order("ticket_no DESC").
		Limit(1).
		Take(&last).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 1, nil
		}
		return 0, err
	}
	return last.TicketNo + 1, nil
}

type ListFilter struct {
	StoreID     int64
	BusinessDay *time.Time
	Status      string
	Page        int
	PageSize    int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]model.QueueTicket, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.QueueTicket{}).Where("store_id = ?", f.StoreID)
	if f.BusinessDay != nil {
		q = q.Where("business_day = ?", truncateDate(*f.BusinessDay))
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
		f.PageSize = 20
	}
	var list []model.QueueTicket
	err := q.Order("business_day DESC, ticket_no ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
