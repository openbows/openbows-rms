package menu

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type DishRepo struct {
	db *gorm.DB
}

func NewDishRepo(db *gorm.DB) *DishRepo {
	return &DishRepo{db: db}
}

func (r *DishRepo) DB() *gorm.DB { return r.db }

func (r *DishRepo) Create(ctx context.Context, m *model.Dish) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *DishRepo) Update(ctx context.Context, m *model.Dish) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *DishRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Dish{}, id).Error
}

func (r *DishRepo) GetByID(ctx context.Context, id int64) (*model.Dish, error) {
	var m model.Dish
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type DishListFilter struct {
	StoreID    int64
	CategoryID int64
	StationID  int64
	IsOnSale   *bool
	Keyword    string
	Page       int
	PageSize   int
}

func (r *DishRepo) List(ctx context.Context, f DishListFilter) ([]model.Dish, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Dish{}).Where("store_id = ?", f.StoreID)
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.StationID > 0 {
		q = q.Where("station_id = ?", f.StationID)
	}
	if f.IsOnSale != nil {
		q = q.Where("is_on_sale = ?", *f.IsOnSale)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("name ILIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Dish
	err := q.Order("sort_no ASC, id DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *DishRepo) ListSpecs(ctx context.Context, dishID int64) ([]model.DishSpec, error) {
	var list []model.DishSpec
	err := r.db.WithContext(ctx).Where("dish_id = ?", dishID).
		Order("sort_no ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *DishRepo) ListOptions(ctx context.Context, dishID int64) ([]model.DishOption, error) {
	var list []model.DishOption
	err := r.db.WithContext(ctx).Where("dish_id = ?", dishID).
		Order("sort_no ASC, id ASC").Find(&list).Error
	return list, err
}

// ReplaceSpecs 事务内整表替换规格
func (r *DishRepo) ReplaceSpecs(ctx context.Context, storeID, dishID int64, specs []model.DishSpec) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dish_id = ?", dishID).Delete(&model.DishSpec{}).Error; err != nil {
			return err
		}
		for i := range specs {
			specs[i].ID = 0
			specs[i].StoreID = storeID
			specs[i].DishID = dishID
			if err := tx.Create(&specs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceOptions 事务内整表替换加料
func (r *DishRepo) ReplaceOptions(ctx context.Context, storeID, dishID int64, options []model.DishOption) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dish_id = ?", dishID).Delete(&model.DishOption{}).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].ID = 0
			options[i].StoreID = storeID
			options[i].DishID = dishID
			if err := tx.Create(&options[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DishRepo) ListMaterials(ctx context.Context, dishID int64) ([]model.DishMaterial, error) {
	var list []model.DishMaterial
	err := r.db.WithContext(ctx).Where("dish_id = ?", dishID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *DishRepo) ListMaterialsByDishIDs(ctx context.Context, dishIDs []int64) ([]model.DishMaterial, error) {
	if len(dishIDs) == 0 {
		return nil, nil
	}
	var list []model.DishMaterial
	err := r.db.WithContext(ctx).Where("dish_id IN ?", dishIDs).Find(&list).Error
	return list, err
}

// ReplaceMaterials 事务内整表替换配方
func (r *DishRepo) ReplaceMaterials(ctx context.Context, storeID, dishID int64, items []model.DishMaterial) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dish_id = ?", dishID).Delete(&model.DishMaterial{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].StoreID = storeID
			items[i].DishID = dishID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
