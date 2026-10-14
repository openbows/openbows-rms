package menu

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, m *model.MenuCategory) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *CategoryRepo) Update(ctx context.Context, m *model.MenuCategory) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.MenuCategory{}, id).Error
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*model.MenuCategory, error) {
	var m model.MenuCategory
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *CategoryRepo) ListByStore(ctx context.Context, storeID int64) ([]model.MenuCategory, error) {
	var list []model.MenuCategory
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).
		Order("sort_no ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepo) ExistsInStore(ctx context.Context, storeID, id int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.MenuCategory{}).
		Where("store_id = ? AND id = ?", storeID, id).Count(&n).Error
	return n > 0, err
}
