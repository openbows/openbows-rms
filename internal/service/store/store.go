package store

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	storerepo "openbows-rms/internal/repository/store"
)

type Service struct {
	repo *storerepo.Repository
}

func New(repo *storerepo.Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Code   string `json:"code" binding:"required,max=64"`
	Name   string `json:"name" binding:"required,max=128"`
	Status *bool  `json:"status"`
}

type UpdateInput struct {
	Code   string `json:"code" binding:"required,max=64"`
	Name   string `json:"name" binding:"required,max=128"`
	Status *bool  `json:"status"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Store, error) {
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Store{Code: in.Code, Name: in.Name, Status: status}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建门店失败: %w", err)
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, id int64, in UpdateInput) (*model.Store, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("门店不存在")
		}
		return nil, err
	}
	m.Code = in.Code
	m.Name = in.Name
	if in.Status != nil {
		m.Status = *in.Status
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新门店失败: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("门店不存在")
		}
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id int64) (*model.Store, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("门店不存在")
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Store, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.repo.List(ctx, page, pageSize, keyword)
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
