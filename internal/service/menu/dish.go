package menu

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	menurepo "openbows-rms/internal/repository/menu"
)

type DishService struct {
	dishes     *menurepo.DishRepo
	categories *menurepo.CategoryRepo
	stations   *menurepo.StationRepo
	materials  MaterialChecker
}

// MaterialChecker 校验物料归属（由 stock repo 实现，避免反向依赖）
type MaterialChecker interface {
	GetMaterialByID(ctx context.Context, id int64) (*model.Material, error)
}

func NewDishService(dishes *menurepo.DishRepo, categories *menurepo.CategoryRepo, stations *menurepo.StationRepo, materials MaterialChecker) *DishService {
	return &DishService{dishes: dishes, categories: categories, stations: stations, materials: materials}
}

type DishInput struct {
	CategoryID  int64   `json:"category_id" binding:"required"`
	StationID   *int64  `json:"station_id"`
	Name        string  `json:"name" binding:"required,max=128"`
	Description string  `json:"description" binding:"max=512"`
	BasePrice   string  `json:"base_price" binding:"required"`
	IsOnSale    *bool   `json:"is_on_sale"`
	SortNo      int     `json:"sort_no"`
	ImageURL    string  `json:"image_url" binding:"max=512"`
}

type SpecItemInput struct {
	Name   string `json:"name" binding:"required,max=128"`
	Price  string `json:"price" binding:"required"`
	SortNo int    `json:"sort_no"`
}

type OptionItemInput struct {
	Name   string `json:"name" binding:"required,max=128"`
	Price  string `json:"price" binding:"required"`
	SortNo int    `json:"sort_no"`
}

type ReplaceSpecsInput struct {
	Specs []SpecItemInput `json:"specs" binding:"required,dive"`
}

type ReplaceOptionsInput struct {
	Options []OptionItemInput `json:"options" binding:"required,dive"`
}

type MaterialItemInput struct {
	MaterialID int64  `json:"material_id" binding:"required"`
	Qty        string `json:"qty" binding:"required"`
	Remark     string `json:"remark"`
}

type ReplaceMaterialsInput struct {
	Materials []MaterialItemInput `json:"materials" binding:"required,dive"`
}

type DishDetail struct {
	Dish      *model.Dish           `json:"dish"`
	Specs     []model.DishSpec      `json:"specs"`
	Options   []model.DishOption    `json:"options"`
	Materials []model.DishMaterial  `json:"materials"`
}

func (s *DishService) Create(ctx context.Context, storeID int64, in DishInput) (*model.Dish, error) {
	if err := s.validateRefs(ctx, storeID, in.CategoryID, in.StationID); err != nil {
		return nil, err
	}
	price, err := parseMoney(in.BasePrice)
	if err != nil {
		return nil, err
	}
	onSale := true
	if in.IsOnSale != nil {
		onSale = *in.IsOnSale
	}
	m := &model.Dish{
		StoreID:     storeID,
		CategoryID:  in.CategoryID,
		StationID:   in.StationID,
		Name:        in.Name,
		Description: in.Description,
		BasePrice:   price,
		IsOnSale:    onSale,
		SortNo:      in.SortNo,
		ImageURL:    in.ImageURL,
	}
	if err := s.dishes.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建菜品失败: %w", err)
	}
	return m, nil
}

func (s *DishService) Update(ctx context.Context, storeID, id int64, in DishInput) (*model.Dish, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if err := s.validateRefs(ctx, storeID, in.CategoryID, in.StationID); err != nil {
		return nil, err
	}
	price, err := parseMoney(in.BasePrice)
	if err != nil {
		return nil, err
	}
	m.CategoryID = in.CategoryID
	m.StationID = in.StationID
	m.Name = in.Name
	m.Description = in.Description
	m.BasePrice = price
	if in.IsOnSale != nil {
		m.IsOnSale = *in.IsOnSale
	}
	m.SortNo = in.SortNo
	m.ImageURL = in.ImageURL
	if err := s.dishes.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新菜品失败: %w", err)
	}
	return m, nil
}

func (s *DishService) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.dishes.SoftDelete(ctx, id)
}

func (s *DishService) GetDetail(ctx context.Context, storeID, id int64) (*DishDetail, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	specs, err := s.dishes.ListSpecs(ctx, id)
	if err != nil {
		return nil, err
	}
	options, err := s.dishes.ListOptions(ctx, id)
	if err != nil {
		return nil, err
	}
	materials, err := s.dishes.ListMaterials(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DishDetail{Dish: m, Specs: specs, Options: options, Materials: materials}, nil
}

func (s *DishService) List(ctx context.Context, storeID int64, categoryID, stationID int64, isOnSale *bool, keyword string, page, pageSize int) ([]model.Dish, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.dishes.List(ctx, menurepo.DishListFilter{
		StoreID:    storeID,
		CategoryID: categoryID,
		StationID:  stationID,
		IsOnSale:   isOnSale,
		Keyword:    keyword,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (s *DishService) ReplaceSpecs(ctx context.Context, storeID, dishID int64, in ReplaceSpecsInput) ([]model.DishSpec, error) {
	if _, err := s.getOwned(ctx, storeID, dishID); err != nil {
		return nil, err
	}
	specs := make([]model.DishSpec, 0, len(in.Specs))
	for _, item := range in.Specs {
		price, err := parseMoney(item.Price)
		if err != nil {
			return nil, err
		}
		specs = append(specs, model.DishSpec{Name: item.Name, Price: price, SortNo: item.SortNo})
	}
	// 关键：批量替换规格
	if err := s.dishes.ReplaceSpecs(ctx, storeID, dishID, specs); err != nil {
		return nil, fmt.Errorf("保存规格失败: %w", err)
	}
	return s.dishes.ListSpecs(ctx, dishID)
}

func (s *DishService) ReplaceOptions(ctx context.Context, storeID, dishID int64, in ReplaceOptionsInput) ([]model.DishOption, error) {
	if _, err := s.getOwned(ctx, storeID, dishID); err != nil {
		return nil, err
	}
	options := make([]model.DishOption, 0, len(in.Options))
	for _, item := range in.Options {
		price, err := parseMoney(item.Price)
		if err != nil {
			return nil, err
		}
		options = append(options, model.DishOption{Name: item.Name, Price: price, SortNo: item.SortNo})
	}
	// 关键：批量替换加料
	if err := s.dishes.ReplaceOptions(ctx, storeID, dishID, options); err != nil {
		return nil, fmt.Errorf("保存加料失败: %w", err)
	}
	return s.dishes.ListOptions(ctx, dishID)
}

func (s *DishService) ReplaceMaterials(ctx context.Context, storeID, dishID int64, in ReplaceMaterialsInput) ([]model.DishMaterial, error) {
	if _, err := s.getOwned(ctx, storeID, dishID); err != nil {
		return nil, err
	}
	if s.materials == nil {
		return nil, fmt.Errorf("物料服务未配置")
	}
	items := make([]model.DishMaterial, 0, len(in.Materials))
	seen := map[int64]struct{}{}
	for _, item := range in.Materials {
		if _, ok := seen[item.MaterialID]; ok {
			return nil, fmt.Errorf("配方物料重复: %d", item.MaterialID)
		}
		seen[item.MaterialID] = struct{}{}
		mat, err := s.materials.GetMaterialByID(ctx, item.MaterialID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("物料不存在: %d", item.MaterialID)
			}
			return nil, err
		}
		if mat.StoreID != storeID {
			return nil, fmt.Errorf("物料不属于本门店")
		}
		qty, err := decimal.NewFromString(item.Qty)
		if err != nil || !qty.IsPositive() {
			return nil, fmt.Errorf("配方用量须大于0")
		}
		items = append(items, model.DishMaterial{
			MaterialID: item.MaterialID, Qty: qty.Round(3), Remark: item.Remark,
		})
	}
	if err := s.dishes.ReplaceMaterials(ctx, storeID, dishID, items); err != nil {
		return nil, fmt.Errorf("保存配方失败: %w", err)
	}
	return s.dishes.ListMaterials(ctx, dishID)
}

func (s *DishService) validateRefs(ctx context.Context, storeID, categoryID int64, stationID *int64) error {
	ok, err := s.categories.ExistsInStore(ctx, storeID, categoryID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("分类不存在或不属于本门店")
	}
	if stationID != nil {
		ok, err := s.stations.ExistsInStore(ctx, storeID, *stationID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("档口不存在或不属于本门店")
		}
	}
	return nil
}

func (s *DishService) getOwned(ctx context.Context, storeID, id int64) (*model.Dish, error) {
	m, err := s.dishes.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("菜品不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店菜品")
	}
	return m, nil
}

func parseMoney(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("价格格式错误")
	}
	if d.IsNegative() {
		return decimal.Zero, fmt.Errorf("价格不能为负")
	}
	return d.Round(2), nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
