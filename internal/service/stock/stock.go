package stock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	stockrepo "openbows-rms/internal/repository/stock"
)

/*
作废 / 冲销规则：
1. draft：可作废，不改结存；也可软删除物料无关的草稿单（本接口用 void）。
2. posted：作废=冲销，按明细反向变更结存后置 voided；冲销后结存不得为负。
3. voided：不可再审核/再作废。
4. 禁止任何接口直接改 stock_balances；仅 post/void(posted) 事务内更新。

结账扣料：payment 调用 CreateAndPostOutboundTx。
*/

type Service struct {
	repo *stockrepo.Repository
}

func New(repo *stockrepo.Repository) *Service {
	return &Service{repo: repo}
}

type Operator struct {
	UserID int64
}

type MaterialInput struct {
	Code     string `json:"code" binding:"required,max=64"`
	Name     string `json:"name" binding:"required,max=128"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
	Status   *bool  `json:"status"`
	Remark   string `json:"remark"`
}

type DocItemInput struct {
	MaterialID int64  `json:"material_id" binding:"required"`
	Qty        string `json:"qty" binding:"required"`
	UnitCost   string `json:"unit_cost"`
	Remark     string `json:"remark"`
}

type CreateDocInput struct {
	DocType string         `json:"doc_type" binding:"required"`
	Remark  string         `json:"remark"`
	Items   []DocItemInput `json:"items" binding:"required,min=1,dive"`
}

type VoidInput struct {
	Reason string `json:"reason"`
}

type DocDetail struct {
	Doc   *model.StockDoc       `json:"doc"`
	Items []model.StockDocItem  `json:"items"`
}

func (s *Service) CreateMaterial(ctx context.Context, storeID int64, in MaterialInput) (*model.Material, error) {
	unit := in.Unit
	if unit == "" {
		unit = "个"
	}
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Material{
		StoreID: storeID, Code: in.Code, Name: in.Name, Unit: unit,
		Category: in.Category, Status: status, Remark: in.Remark,
	}
	if err := s.repo.CreateMaterial(ctx, m); err != nil {
		return nil, fmt.Errorf("创建物料失败: %w", err)
	}
	return m, nil
}

func (s *Service) UpdateMaterial(ctx context.Context, storeID, id int64, in MaterialInput) (*model.Material, error) {
	m, err := s.getOwnedMaterial(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	unit := in.Unit
	if unit == "" {
		unit = m.Unit
	}
	m.Code = in.Code
	m.Name = in.Name
	m.Unit = unit
	m.Category = in.Category
	m.Remark = in.Remark
	if in.Status != nil {
		m.Status = *in.Status
	}
	if err := s.repo.SaveMaterial(ctx, m); err != nil {
		return nil, fmt.Errorf("更新物料失败: %w", err)
	}
	return m, nil
}

func (s *Service) DeleteMaterial(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwnedMaterial(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDeleteMaterial(ctx, id)
}

func (s *Service) GetMaterial(ctx context.Context, storeID, id int64) (*model.Material, error) {
	return s.getOwnedMaterial(ctx, storeID, id)
}

func (s *Service) ListMaterials(ctx context.Context, storeID int64, keyword string, status *bool, page, pageSize int) ([]model.Material, int64, error) {
	return s.repo.ListMaterials(ctx, stockrepo.MaterialListFilter{
		StoreID: storeID, Keyword: keyword, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) ListBalances(ctx context.Context, storeID, materialID int64, page, pageSize int) ([]stockrepo.BalanceView, int64, error) {
	return s.repo.ListBalances(ctx, stockrepo.BalanceListFilter{
		StoreID: storeID, MaterialID: materialID, Page: page, PageSize: pageSize,
	})
}

func (s *Service) CreateDoc(ctx context.Context, storeID int64, op Operator, in CreateDocInput) (*DocDetail, error) {
	if !validDocType(in.DocType) {
		return nil, fmt.Errorf("无效单据类型，可选：inbound/outbound/adjust")
	}
	built, err := s.buildItems(ctx, storeID, in.DocType, in.Items)
	if err != nil {
		return nil, err
	}
	var detail DocDetail
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		d := &model.StockDoc{
			StoreID: storeID, DocNo: s.repo.CreateDocNo(storeID), DocType: in.DocType,
			Status: model.StockDocStatusDraft, Remark: in.Remark, CreatedBy: op.UserID,
		}
		if err := s.repo.CreateDocTx(ctx, tx, d); err != nil {
			return fmt.Errorf("创建单据失败: %w", err)
		}
		items := make([]model.StockDocItem, 0, len(built))
		for _, it := range built {
			it.StoreID = storeID
			it.DocID = d.ID
			if err := s.repo.CreateDocItemTx(ctx, tx, &it); err != nil {
				return fmt.Errorf("创建明细失败: %w", err)
			}
			items = append(items, it)
		}
		detail.Doc = d
		detail.Items = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *Service) GetDoc(ctx context.Context, storeID, id int64) (*DocDetail, error) {
	d, err := s.getOwnedDoc(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListDocItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DocDetail{Doc: d, Items: items}, nil
}

func (s *Service) ListDocs(ctx context.Context, storeID int64, docType, status string, page, pageSize int) ([]model.StockDoc, int64, error) {
	return s.repo.ListDocs(ctx, stockrepo.DocListFilter{
		StoreID: storeID, DocType: docType, Status: status, Page: page, PageSize: pageSize,
	})
}

// CreateAndPostInboundTx 在调用方事务内创建入库单并审核生效（采购收货复用）
func (s *Service) CreateAndPostInboundTx(ctx context.Context, tx *gorm.DB, storeID int64, op Operator, remark string, items []DocItemInput) (*model.StockDoc, error) {
	return s.createAndPostDocTx(ctx, tx, storeID, op, model.StockDocTypeInbound, remark, items)
}

// CreateAndPostOutboundTx 在调用方事务内创建出库单并审核生效（结账扣料复用）
func (s *Service) CreateAndPostOutboundTx(ctx context.Context, tx *gorm.DB, storeID int64, op Operator, remark string, items []DocItemInput) (*model.StockDoc, error) {
	return s.createAndPostDocTx(ctx, tx, storeID, op, model.StockDocTypeOutbound, remark, items)
}

func (s *Service) createAndPostDocTx(ctx context.Context, tx *gorm.DB, storeID int64, op Operator, docType, remark string, items []DocItemInput) (*model.StockDoc, error) {
	built, err := s.buildItems(ctx, storeID, docType, items)
	if err != nil {
		return nil, err
	}
	d := &model.StockDoc{
		StoreID: storeID, DocNo: s.repo.CreateDocNo(storeID), DocType: docType,
		Status: model.StockDocStatusDraft, Remark: remark, CreatedBy: op.UserID,
	}
	label := "入库"
	if docType == model.StockDocTypeOutbound {
		label = "出库"
	}
	if err := s.repo.CreateDocTx(ctx, tx, d); err != nil {
		return nil, fmt.Errorf("创建库存%s单失败: %w", label, err)
	}
	for i := range built {
		built[i].StoreID = storeID
		built[i].DocID = d.ID
		if err := s.repo.CreateDocItemTx(ctx, tx, &built[i]); err != nil {
			return nil, fmt.Errorf("创建库存%s明细失败: %w", label, err)
		}
		delta, err := signedDelta(docType, built[i].Qty, false)
		if err != nil {
			return nil, err
		}
		if err := s.applyDeltaTx(ctx, tx, storeID, built[i].MaterialID, delta); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	d.Status = model.StockDocStatusPosted
	d.PostedBy = op.UserID
	d.PostedAt = &now
	if err := s.repo.SaveDocTx(ctx, tx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Post 审核生效：事务更新结存
func (s *Service) Post(ctx context.Context, storeID, id int64, op Operator) (*DocDetail, error) {
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		d, err := s.repo.GetDocByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("单据不存在")
			}
			return err
		}
		if d.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店单据")
		}
		if d.Status != model.StockDocStatusDraft {
			return fmt.Errorf("仅草稿单可审核，当前：%s", d.Status)
		}
		items, err := s.repo.ListDocItemsTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("单据明细不能为空")
		}
		for _, it := range items {
			delta, err := signedDelta(d.DocType, it.Qty, false)
			if err != nil {
				return err
			}
			if err := s.applyDeltaTx(ctx, tx, storeID, it.MaterialID, delta); err != nil {
				return err
			}
		}
		now := time.Now()
		d.Status = model.StockDocStatusPosted
		d.PostedBy = op.UserID
		d.PostedAt = &now
		return s.repo.SaveDocTx(ctx, tx, d)
	})
	if err != nil {
		return nil, err
	}
	return s.GetDoc(ctx, storeID, id)
}

// Void 作废/冲销
func (s *Service) Void(ctx context.Context, storeID, id int64, op Operator, in VoidInput) (*DocDetail, error) {
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		d, err := s.repo.GetDocByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("单据不存在")
			}
			return err
		}
		if d.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店单据")
		}
		switch d.Status {
		case model.StockDocStatusDraft:
			// 草稿作废：不改结存
		case model.StockDocStatusPosted:
			// 已审核作废：冲销结存
			items, err := s.repo.ListDocItemsTx(ctx, tx, id)
			if err != nil {
				return err
			}
			for _, it := range items {
				delta, err := signedDelta(d.DocType, it.Qty, true) // reverse
				if err != nil {
					return err
				}
				if err := s.applyDeltaTx(ctx, tx, storeID, it.MaterialID, delta); err != nil {
					return fmt.Errorf("冲销失败: %w", err)
				}
			}
		default:
			return fmt.Errorf("当前状态不可作废：%s", d.Status)
		}
		now := time.Now()
		d.Status = model.StockDocStatusVoided
		d.VoidedBy = op.UserID
		d.VoidedAt = &now
		d.VoidReason = in.Reason
		return s.repo.SaveDocTx(ctx, tx, d)
	})
	if err != nil {
		return nil, err
	}
	return s.GetDoc(ctx, storeID, id)
}

func (s *Service) applyDeltaTx(ctx context.Context, tx *gorm.DB, storeID, materialID int64, delta decimal.Decimal) error {
	mat, err := s.repo.GetMaterialByIDTx(ctx, tx, materialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("物料不存在: %d", materialID)
		}
		return err
	}
	if mat.StoreID != storeID {
		return fmt.Errorf("物料不属于本门店")
	}
	b, err := s.repo.GetBalanceTx(ctx, tx, storeID, materialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if delta.IsNegative() {
				return fmt.Errorf("物料[%s]结存不足", mat.Name)
			}
			nb := &model.StockBalance{
				StoreID: storeID, MaterialID: materialID, Qty: delta.Round(3),
			}
			return s.repo.CreateBalanceTx(ctx, tx, nb)
		}
		return err
	}
	after := b.Qty.Add(delta)
	if after.IsNegative() {
		return fmt.Errorf("物料[%s]结存不足，当前：%s", mat.Name, b.Qty.String())
	}
	b.Qty = after.Round(3)
	return s.repo.SaveBalanceTx(ctx, tx, b)
}

func (s *Service) buildItems(ctx context.Context, storeID int64, docType string, items []DocItemInput) ([]model.StockDocItem, error) {
	out := make([]model.StockDocItem, 0, len(items))
	for _, in := range items {
		mat, err := s.repo.GetMaterialByID(ctx, in.MaterialID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("物料不存在: %d", in.MaterialID)
			}
			return nil, err
		}
		if mat.StoreID != storeID {
			return nil, fmt.Errorf("物料不属于本门店")
		}
		if !mat.Status {
			return nil, fmt.Errorf("物料已停用: %s", mat.Code)
		}
		qty, err := decimal.NewFromString(in.Qty)
		if err != nil || qty.IsZero() {
			return nil, fmt.Errorf("明细数量无效")
		}
		qty = qty.Round(3)
		if docType == model.StockDocTypeInbound || docType == model.StockDocTypeOutbound {
			if !qty.IsPositive() {
				return nil, fmt.Errorf("入库/出库数量须大于0")
			}
		}
		cost := decimal.Zero
		if in.UnitCost != "" {
			c, err := decimal.NewFromString(in.UnitCost)
			if err != nil || c.IsNegative() {
				return nil, fmt.Errorf("单位成本无效")
			}
			cost = c.Round(2)
		}
		out = append(out, model.StockDocItem{
			MaterialID: in.MaterialID, Qty: qty, UnitCost: cost, Remark: in.Remark,
		})
	}
	return out, nil
}

func signedDelta(docType string, qty decimal.Decimal, reverse bool) (decimal.Decimal, error) {
	var delta decimal.Decimal
	switch docType {
	case model.StockDocTypeInbound:
		delta = qty
	case model.StockDocTypeOutbound:
		delta = qty.Neg()
	case model.StockDocTypeAdjust:
		delta = qty // 可为正负
	default:
		return decimal.Zero, fmt.Errorf("无效单据类型")
	}
	if reverse {
		delta = delta.Neg()
	}
	return delta, nil
}

func (s *Service) getOwnedMaterial(ctx context.Context, storeID, id int64) (*model.Material, error) {
	m, err := s.repo.GetMaterialByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("物料不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店物料")
	}
	return m, nil
}

func (s *Service) getOwnedDoc(ctx context.Context, storeID, id int64) (*model.StockDoc, error) {
	d, err := s.repo.GetDocByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("单据不存在")
		}
		return nil, err
	}
	if d.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店单据")
	}
	return d, nil
}

func validDocType(t string) bool {
	switch t {
	case model.StockDocTypeInbound, model.StockDocTypeOutbound, model.StockDocTypeAdjust:
		return true
	default:
		return false
	}
}
