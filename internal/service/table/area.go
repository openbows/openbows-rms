package table

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	tablerepo "openbows-rms/internal/repository/table"
)

type AreaService struct {
	repo *tablerepo.AreaRepo
}

func NewAreaService(repo *tablerepo.AreaRepo) *AreaService {
	return &AreaService{repo: repo}
}

type AreaInput struct {
	Name   string `json:"name" binding:"required,max=128"`
	SortNo int    `json:"sort_no"`
	Status *bool  `json:"status"`
}

func (s *AreaService) Create(ctx context.Context, storeID int64, in AreaInput) (*model.Area, error) {
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Area{StoreID: storeID, Name: in.Name, SortNo: in.SortNo, Status: status}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建区域失败: %w", err)
	}
	return m, nil
}

func (s *AreaService) Update(ctx context.Context, storeID, id int64, in AreaInput) (*model.Area, error) {
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
		return nil, fmt.Errorf("更新区域失败: %w", err)
	}
	return m, nil
}

func (s *AreaService) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *AreaService) Get(ctx context.Context, storeID, id int64) (*model.Area, error) {
	return s.getOwned(ctx, storeID, id)
}

func (s *AreaService) List(ctx context.Context, storeID int64) ([]model.Area, error) {
	return s.repo.ListByStore(ctx, storeID)
}

func (s *AreaService) getOwned(ctx context.Context, storeID, id int64) (*model.Area, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("区域不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店区域")
	}
	return m, nil
}
