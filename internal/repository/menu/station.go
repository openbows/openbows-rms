package menu

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type StationRepo struct {
	db *gorm.DB
}

func NewStationRepo(db *gorm.DB) *StationRepo {
	return &StationRepo{db: db}
}

func (r *StationRepo) Create(ctx context.Context, m *model.Station) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *StationRepo) Update(ctx context.Context, m *model.Station) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *StationRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Station{}, id).Error
}

func (r *StationRepo) GetByID(ctx context.Context, id int64) (*model.Station, error) {
	var m model.Station
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *StationRepo) ListByStore(ctx context.Context, storeID int64) ([]model.Station, error) {
	var list []model.Station
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).
		Order("sort_no ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *StationRepo) ExistsInStore(ctx context.Context, storeID, id int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Station{}).
		Where("store_id = ? AND id = ?", storeID, id).Count(&n).Error
	return n > 0, err
}
