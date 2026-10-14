package member

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"openbows-rms/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) CreateMemberTx(ctx context.Context, tx *gorm.DB, m *model.Member) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) CreateWalletTx(ctx context.Context, tx *gorm.DB, w *model.MemberWallet) error {
	return tx.WithContext(ctx).Create(w).Error
}

func (r *Repository) SaveMember(ctx context.Context, m *model.Member) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDeleteMember(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Member{}, id).Error
}

func (r *Repository) SoftDeleteWalletByMember(ctx context.Context, memberID int64) error {
	return r.db.WithContext(ctx).Where("member_id = ?", memberID).Delete(&model.MemberWallet{}).Error
}

func (r *Repository) GetMemberByID(ctx context.Context, id int64) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetMemberByPhone(ctx context.Context, storeID int64, phone string) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).Where("store_id = ? AND phone = ?", storeID, phone).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetWalletByMemberID(ctx context.Context, memberID int64) (*model.MemberWallet, error) {
	var w model.MemberWallet
	err := r.db.WithContext(ctx).Where("member_id = ?", memberID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) GetWalletByMemberIDTx(ctx context.Context, tx *gorm.DB, memberID int64) (*model.MemberWallet, error) {
	var w model.MemberWallet
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("member_id = ?", memberID).
		First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) SaveWalletTx(ctx context.Context, tx *gorm.DB, w *model.MemberWallet) error {
	return tx.WithContext(ctx).Save(w).Error
}

func (r *Repository) CreateWalletLogTx(ctx context.Context, tx *gorm.DB, log *model.MemberWalletLog) error {
	return tx.WithContext(ctx).Create(log).Error
}

func (r *Repository) CreatePointLogTx(ctx context.Context, tx *gorm.DB, log *model.MemberPointLog) error {
	return tx.WithContext(ctx).Create(log).Error
}

type ListFilter struct {
	StoreID  int64
	Keyword  string
	Status   *bool
	Page     int
	PageSize int
}

func (r *Repository) ListMembers(ctx context.Context, f ListFilter) ([]model.Member, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Member{}).Where("store_id = ?", f.StoreID)
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("phone ILIKE ? OR name ILIKE ?", like, like)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.Member
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListWalletLogs(ctx context.Context, storeID, memberID int64, page, pageSize int) ([]model.MemberWalletLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MemberWalletLog{}).
		Where("store_id = ? AND member_id = ?", storeID, memberID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var list []model.MemberWalletLog
	err := q.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListPointLogs(ctx context.Context, storeID, memberID int64, page, pageSize int) ([]model.MemberPointLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MemberPointLog{}).
		Where("store_id = ? AND member_id = ?", storeID, memberID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var list []model.MemberPointLog
	err := q.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
