package kds

import (
	"context"

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

func (r *Repository) CreateTicketTx(ctx context.Context, tx *gorm.DB, t *model.KdsTicket) error {
	return tx.WithContext(ctx).Create(t).Error
}

func (r *Repository) CreateTicketItemTx(ctx context.Context, tx *gorm.DB, it *model.KdsTicketItem) error {
	return tx.WithContext(ctx).Create(it).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.KdsTicket, error) {
	var m model.KdsTicket
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.KdsTicket, error) {
	var m model.KdsTicket
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) SaveTx(ctx context.Context, tx *gorm.DB, t *model.KdsTicket) error {
	return tx.WithContext(ctx).Save(t).Error
}

func (r *Repository) ListItems(ctx context.Context, ticketID int64) ([]model.KdsTicketItem, error) {
	var list []model.KdsTicketItem
	err := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListItemsTx(ctx context.Context, tx *gorm.DB, ticketID int64) ([]model.KdsTicketItem, error) {
	var list []model.KdsTicketItem
	err := tx.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) UpdateItemsStatusTx(ctx context.Context, tx *gorm.DB, ticketID int64, status string) error {
	return tx.WithContext(ctx).Model(&model.KdsTicketItem{}).
		Where("ticket_id = ?", ticketID).
		Update("status", status).Error
}

type ListFilter struct {
	StoreID   int64
	StationID int64
	Status    string
}

func (r *Repository) ListTickets(ctx context.Context, f ListFilter) ([]model.KdsTicket, error) {
	q := r.db.WithContext(ctx).Model(&model.KdsTicket{}).Where("store_id = ?", f.StoreID)
	if f.StationID > 0 {
		q = q.Where("station_id = ?", f.StationID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	} else {
		// 默认不返回已完成，便于厨显
		q = q.Where("status IN ?", []string{model.KitchenStatusPending, model.KitchenStatusCooking})
	}
	var list []model.KdsTicket
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListItemsByTicketIDs(ctx context.Context, ticketIDs []int64) ([]model.KdsTicketItem, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}
	var list []model.KdsTicketItem
	err := r.db.WithContext(ctx).Where("ticket_id IN ?", ticketIDs).Order("id ASC").Find(&list).Error
	return list, err
}
