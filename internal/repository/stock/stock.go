package stock

import (
	"context"
	"fmt"
	"time"

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

func (r *Repository) CreateDocNo(storeID int64) string {
	return fmt.Sprintf("ST%d%s%04d", storeID, time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

// ---- materials ----

func (r *Repository) CreateMaterial(ctx context.Context, m *model.Material) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveMaterial(ctx context.Context, m *model.Material) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDeleteMaterial(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Material{}, id).Error
}

func (r *Repository) GetMaterialByID(ctx context.Context, id int64) (*model.Material, error) {
	var m model.Material
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetMaterialByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.Material, error) {
	var m model.Material
	err := tx.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type MaterialListFilter struct {
	StoreID  int64
	Keyword  string
	Status   *bool
	Page     int
	PageSize int
}

func (r *Repository) ListMaterials(ctx context.Context, f MaterialListFilter) ([]model.Material, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Material{}).Where("store_id = ?", f.StoreID)
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
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
	var list []model.Material
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// ---- balances ----

func (r *Repository) GetBalanceTx(ctx context.Context, tx *gorm.DB, storeID, materialID int64) (*model.StockBalance, error) {
	var b model.StockBalance
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("store_id = ? AND material_id = ?", storeID, materialID).
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) CreateBalanceTx(ctx context.Context, tx *gorm.DB, b *model.StockBalance) error {
	return tx.WithContext(ctx).Create(b).Error
}

func (r *Repository) SaveBalanceTx(ctx context.Context, tx *gorm.DB, b *model.StockBalance) error {
	return tx.WithContext(ctx).Save(b).Error
}

type BalanceListFilter struct {
	StoreID    int64
	MaterialID int64
	Page       int
	PageSize   int
}

type BalanceView struct {
	Balance  model.StockBalance `json:"balance"`
	Material *model.Material    `json:"material,omitempty"`
}

func (r *Repository) ListBalances(ctx context.Context, f BalanceListFilter) ([]BalanceView, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.StockBalance{}).Where("store_id = ?", f.StoreID)
	if f.MaterialID > 0 {
		q = q.Where("material_id = ?", f.MaterialID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	var balances []model.StockBalance
	if err := q.Order("id ASC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&balances).Error; err != nil {
		return nil, 0, err
	}
	ids := make([]int64, 0, len(balances))
	for _, b := range balances {
		ids = append(ids, b.MaterialID)
	}
	matMap := map[int64]model.Material{}
	if len(ids) > 0 {
		var mats []model.Material
		_ = r.db.WithContext(ctx).Where("id IN ?", ids).Find(&mats).Error
		for _, m := range mats {
			matMap[m.ID] = m
		}
	}
	out := make([]BalanceView, 0, len(balances))
	for _, b := range balances {
		v := BalanceView{Balance: b}
		if m, ok := matMap[b.MaterialID]; ok {
			mm := m
			v.Material = &mm
		}
		out = append(out, v)
	}
	return out, total, nil
}

// ---- docs ----

func (r *Repository) CreateDocTx(ctx context.Context, tx *gorm.DB, d *model.StockDoc) error {
	return tx.WithContext(ctx).Create(d).Error
}

func (r *Repository) CreateDocItemTx(ctx context.Context, tx *gorm.DB, it *model.StockDocItem) error {
	return tx.WithContext(ctx).Create(it).Error
}

func (r *Repository) SaveDocTx(ctx context.Context, tx *gorm.DB, d *model.StockDoc) error {
	return tx.WithContext(ctx).Save(d).Error
}

func (r *Repository) GetDocByID(ctx context.Context, id int64) (*model.StockDoc, error) {
	var d model.StockDoc
	err := r.db.WithContext(ctx).First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) GetDocByIDTx(ctx context.Context, tx *gorm.DB, id int64) (*model.StockDoc, error) {
	var d model.StockDoc
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) ListDocItems(ctx context.Context, docID int64) ([]model.StockDocItem, error) {
	var list []model.StockDocItem
	err := r.db.WithContext(ctx).Where("doc_id = ?", docID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) ListDocItemsTx(ctx context.Context, tx *gorm.DB, docID int64) ([]model.StockDocItem, error) {
	var list []model.StockDocItem
	err := tx.WithContext(ctx).Where("doc_id = ?", docID).Order("id ASC").Find(&list).Error
	return list, err
}

type DocListFilter struct {
	StoreID  int64
	DocType  string
	Status   string
	Page     int
	PageSize int
}

func (r *Repository) ListDocs(ctx context.Context, f DocListFilter) ([]model.StockDoc, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.StockDoc{}).Where("store_id = ?", f.StoreID)
	if f.DocType != "" {
		q = q.Where("doc_type = ?", f.DocType)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
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
	var list []model.StockDoc
	err := q.Order("id DESC").Offset((f.Page-1)*f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}
