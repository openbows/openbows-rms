package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	"openbows-rms/internal/pkg/jwtutil"
	userrepo "openbows-rms/internal/repository/user"
)

type Service struct {
	users *userrepo.Repository
	jwt   *jwtutil.Manager
}

func New(users *userrepo.Repository, jwt *jwtutil.Manager) *Service {
	return &Service{users: users, jwt: jwt}
}

type LoginResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

type MeResult struct {
	User            *model.User  `json:"user"`
	Roles           []model.Role `json:"roles"`
	PermissionCodes []string     `json:"permission_codes"`
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, err
	}
	if !u.Status {
		return nil, fmt.Errorf("账号已禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	codes, err := s.users.ListPermissionCodesByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	token, err := s.jwt.Sign(u.ID, u.StoreID, u.Username, codes)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}
	return &LoginResult{Token: token, User: u}, nil
}

func (s *Service) Me(ctx context.Context, userID int64) (*MeResult, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles, err := s.users.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	codes, err := s.users.ListPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &MeResult{User: u, Roles: roles, PermissionCodes: codes}, nil
}
