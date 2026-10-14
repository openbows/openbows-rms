package role

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	rolerepo "openbows-rms/internal/repository/role"
)

type Service struct {
	repo *rolerepo.Repository
}

func New(repo *rolerepo.Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	StoreID int64  `json:"store_id"`
	Code    string `json:"code" binding:"required,max=64"`
	Name    string `json:"name" binding:"required,max=128"`
	Remark  string `json:"remark" binding:"max=255"`
}

type UpdateInput struct {
	Code   string `json:"code" binding:"required,max=64"`
	Name   string `json:"name" binding:"required,max=128"`
	Remark string `json:"remark" binding:"max=255"`
}

type AssignPermissionsInput struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Role, error) {
	m := &model.Role{
		StoreID: in.StoreID,
		Code:    in.Code,
		Name:    in.Name,
		Remark:  in.Remark,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, storeID, id int64, in UpdateInput) (*model.Role, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店角色")
	}
	m.Code = in.Code
	m.Name = in.Name
	m.Remark = in.Remark
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, storeID, id int64) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return err
	}
	if m.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店角色")
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Get(ctx context.Context, storeID, id int64) (*model.Role, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店角色")
	}
	return m, nil
}

func (s *Service) List(ctx context.Context, storeID int64, page, pageSize int, keyword string) ([]model.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.repo.ListByStore(ctx, storeID, page, pageSize, keyword)
}

func (s *Service) AssignPermissions(ctx context.Context, storeID, roleID int64, permissionIDs []int64) error {
	m, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return err
	}
	if m.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店角色")
	}
	n, err := s.repo.CountPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		return err
	}
	if int(n) != len(permissionIDs) {
		return fmt.Errorf("存在无效权限ID")
	}
	// 关键：替换角色权限绑定
	return s.repo.ReplacePermissions(ctx, roleID, permissionIDs)
}

func (s *Service) ListPermissionIDs(ctx context.Context, storeID, roleID int64) ([]int64, error) {
	m, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店角色")
	}
	return s.repo.ListPermissionIDs(ctx, roleID)
}

func (s *Service) ListAllPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.repo.ListPermissions(ctx)
}
