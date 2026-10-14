package role

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

func (r *Repository) Create(ctx context.Context, m *model.Role) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) Update(ctx context.Context, m *model.Role) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	var m model.Role
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListByStore(ctx context.Context, storeID int64, page, pageSize int, keyword string) ([]model.Role, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Role{}).Where("store_id = ?", storeID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Role
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListPermissionIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var list []model.Permission
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) CountPermissionsByIDs(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("id IN ?", ids).Count(&n).Error
	return n, err
}

func (r *Repository) CountRolesByIDsInStore(ctx context.Context, storeID int64, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Role{}).
		Where("store_id = ? AND id IN ?", storeID, ids).
		Count(&n).Error
	return n, err
}

// ReplacePermissions 事务内替换角色权限
func (r *Repository) ReplacePermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			rp := model.RolePermission{RoleID: roleID, PermissionID: pid}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
