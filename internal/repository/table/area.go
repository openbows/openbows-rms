package table

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type AreaRepo struct {
	db *gorm.DB
}

func NewAreaRepo(db *gorm.DB) *AreaRepo {
	return &AreaRepo{db: db}
}

func (r *AreaRepo) Create(ctx context.Context, m *model.Area) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *AreaRepo) Update(ctx context.Context, m *model.Area) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *AreaRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Area{}, id).Error
}

func (r *AreaRepo) GetByID(ctx context.Context, id int64) (*model.Area, error) {
	var m model.Area
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *AreaRepo) ListByStore(ctx context.Context, storeID int64) ([]model.Area, error) {
	var list []model.Area
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).
		Order("sort_no ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *AreaRepo) ExistsInStore(ctx context.Context, storeID, id int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Area{}).
		Where("store_id = ? AND id = ?", storeID, id).Count(&n).Error
	return n > 0, err
}
