package user

import (
	"context"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) Create(ctx context.Context, m *model.User) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) Update(ctx context.Context, m *model.User) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var m model.User
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var m model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListByStore(ctx context.Context, storeID int64, page, pageSize int, keyword string) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("store_id = ?", storeID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username ILIKE ? OR display_name ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.User
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListRoleIDs(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	return ids, err
}

// ReplaceRoles 事务内替换用户角色（先软删再插入）
func (r *Repository) ReplaceRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			ur := model.UserRole{UserID: userID, RoleID: rid}
			if err := tx.Create(&ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListRolesByUserID(ctx context.Context, userID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id AND ur.deleted_at IS NULL").
		Where("ur.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

func (r *Repository) ListPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id AND rp.deleted_at IS NULL").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id AND ur.deleted_at IS NULL").
		Where("ur.user_id = ? AND permissions.deleted_at IS NULL", userID).
		Pluck("permissions.code", &codes).Error
	return codes, err
}
