package table

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type TableRepo struct {
	db *gorm.DB
}

func NewTableRepo(db *gorm.DB) *TableRepo {
	return &TableRepo{db: db}
}

func (r *TableRepo) Create(ctx context.Context, m *model.DiningTable) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *TableRepo) Update(ctx context.Context, m *model.DiningTable) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *TableRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.DiningTable{}, id).Error
}

func (r *TableRepo) GetByID(ctx context.Context, id int64) (*model.DiningTable, error) {
	var m model.DiningTable
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *TableRepo) GetByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.DiningTable, error) {
	var m model.DiningTable
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *TableRepo) SaveTx(ctx context.Context, tx *gorm.DB, m *model.DiningTable) error {
	return tx.WithContext(ctx).Save(m).Error
}

type ListFilter struct {
	StoreID  int64
	AreaID   int64
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

func (r *TableRepo) List(ctx context.Context, f ListFilter) ([]model.DiningTable, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.DiningTable{}).Where("store_id = ?", f.StoreID)
	if f.AreaID > 0 {
		q = q.Where("area_id = ?", f.AreaID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.DiningTable
	err := q.Order("sort_no ASC, id DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}
