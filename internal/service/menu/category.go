package menu

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	menurepo "openbows-rms/internal/repository/menu"
)

type CategoryService struct {
	repo *menurepo.CategoryRepo
}

func NewCategoryService(repo *menurepo.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

type CategoryInput struct {
	Name   string `json:"name" binding:"required,max=128"`
	SortNo int    `json:"sort_no"`
	Status *bool  `json:"status"`
}

func (s *CategoryService) Create(ctx context.Context, storeID int64, in CategoryInput) (*model.MenuCategory, error) {
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.MenuCategory{
		StoreID: storeID,
		Name:    in.Name,
		SortNo:  in.SortNo,
		Status:  status,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}
	return m, nil
}

func (s *CategoryService) Update(ctx context.Context, storeID, id int64, in CategoryInput) (*model.MenuCategory, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	m.Name = in.Name
	m.SortNo = in.SortNo
	if in.Status != nil {
		m.Status = *in.Status
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新分类失败: %w", err)
	}
	return m, nil
}

func (s *CategoryService) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *CategoryService) Get(ctx context.Context, storeID, id int64) (*model.MenuCategory, error) {
	return s.getOwned(ctx, storeID, id)
}

func (s *CategoryService) List(ctx context.Context, storeID int64) ([]model.MenuCategory, error) {
	return s.repo.ListByStore(ctx, storeID)
}

func (s *CategoryService) getOwned(ctx context.Context, storeID, id int64) (*model.MenuCategory, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分类不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店分类")
	}
	return m, nil
}
