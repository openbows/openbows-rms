package store

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

func (r *Repository) Create(ctx context.Context, m *model.Store) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) Update(ctx context.Context, m *model.Store) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Store{}, id).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Store, error) {
	var m model.Store
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Store, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Store{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Store
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
