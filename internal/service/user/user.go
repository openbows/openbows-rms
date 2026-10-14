package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	rolerepo "openbows-rms/internal/repository/role"
	userrepo "openbows-rms/internal/repository/user"
)

type Service struct {
	users *userrepo.Repository
	roles *rolerepo.Repository
}

func New(users *userrepo.Repository, roles *rolerepo.Repository) *Service {
	return &Service{users: users, roles: roles}
}

type CreateInput struct {
	StoreID     int64   `json:"store_id"`
	Username    string  `json:"username" binding:"required,max=64"`
	Password    string  `json:"password" binding:"required,min=6,max=64"`
	Phone       *string `json:"phone"`
	DisplayName string  `json:"display_name" binding:"max=64"`
	Status      *bool   `json:"status"`
}

type UpdateInput struct {
	Phone       *string `json:"phone"`
	DisplayName string  `json:"display_name" binding:"max=64"`
	Password    string  `json:"password" binding:"omitempty,min=6,max=64"`
	Status      *bool   `json:"status"`
}

type AssignRolesInput struct {
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %w", err)
	}
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.User{
		StoreID:      in.StoreID,
		Username:     in.Username,
		PasswordHash: string(hash),
		Phone:        in.Phone,
		DisplayName:  in.DisplayName,
		Status:       status,
	}
	if err := s.users.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, storeID, id int64, in UpdateInput) (*model.User, error) {
	m, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店用户")
	}
	m.Phone = in.Phone
	m.DisplayName = in.DisplayName
	if in.Status != nil {
		m.Status = *in.Status
	}
	if in.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("加密密码失败: %w", err)
		}
		m.PasswordHash = string(hash)
	}
	if err := s.users.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, storeID, id int64) error {
	m, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("用户不存在")
		}
		return err
	}
	if m.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店用户")
	}
	return s.users.SoftDelete(ctx, id)
}

func (s *Service) Get(ctx context.Context, storeID, id int64) (*model.User, error) {
	m, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店用户")
	}
	return m, nil
}

func (s *Service) List(ctx context.Context, storeID int64, page, pageSize int, keyword string) ([]model.User, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.users.ListByStore(ctx, storeID, page, pageSize, keyword)
}

func (s *Service) AssignRoles(ctx context.Context, storeID, userID int64, roleIDs []int64) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("用户不存在")
		}
		return err
	}
	if u.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店用户")
	}
	n, err := s.roles.CountRolesByIDsInStore(ctx, storeID, roleIDs)
	if err != nil {
		return err
	}
	if int(n) != len(roleIDs) {
		return fmt.Errorf("存在无效或不属于本门店的角色")
	}
	// 关键：替换用户角色绑定
	return s.users.ReplaceRoles(ctx, userID, roleIDs)
}

func (s *Service) ListRoleIDs(ctx context.Context, storeID, userID int64) ([]int64, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	if u.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店用户")
	}
	return s.users.ListRoleIDs(ctx, userID)
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
