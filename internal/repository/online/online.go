package online

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

func (r *Repository) CreateExtrasTx(ctx context.Context, tx *gorm.DB, m *model.OnlineOrderExtra) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetByAccessToken(ctx context.Context, token string) (*model.OnlineOrderExtra, error) {
	var m model.OnlineOrderExtra
	err := r.db.WithContext(ctx).Where("access_token = ?", token).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID int64) (*model.OnlineOrderExtra, error) {
	var m model.OnlineOrderExtra
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
