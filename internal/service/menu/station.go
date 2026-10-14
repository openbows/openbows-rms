package menu

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	menurepo "openbows-rms/internal/repository/menu"
)

type StationService struct {
	repo *menurepo.StationRepo
}

func NewStationService(repo *menurepo.StationRepo) *StationService {
	return &StationService{repo: repo}
}

type StationInput struct {
	Code   string `json:"code" binding:"required,max=64"`
	Name   string `json:"name" binding:"required,max=128"`
	SortNo int    `json:"sort_no"`
	Status *bool  `json:"status"`
}

func (s *StationService) Create(ctx context.Context, storeID int64, in StationInput) (*model.Station, error) {
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Station{
		StoreID: storeID,
		Code:    in.Code,
		Name:    in.Name,
		SortNo:  in.SortNo,
		Status:  status,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建档口失败: %w", err)
	}
	return m, nil
}

func (s *StationService) Update(ctx context.Context, storeID, id int64, in StationInput) (*model.Station, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	m.Code = in.Code
	m.Name = in.Name
	m.SortNo = in.SortNo
	if in.Status != nil {
		m.Status = *in.Status
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新档口失败: %w", err)
	}
	return m, nil
}

func (s *StationService) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *StationService) Get(ctx context.Context, storeID, id int64) (*model.Station, error) {
	return s.getOwned(ctx, storeID, id)
}

func (s *StationService) List(ctx context.Context, storeID int64) ([]model.Station, error) {
	return s.repo.ListByStore(ctx, storeID)
}

func (s *StationService) getOwned(ctx context.Context, storeID, id int64) (*model.Station, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("档口不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店档口")
	}
	return m, nil
}
