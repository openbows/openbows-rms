package purchase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	purchaserepo "openbows-rms/internal/repository/purchase"
	stockrepo "openbows-rms/internal/repository/stock"
	stocksvc "openbows-rms/internal/service/stock"
)

type Service struct {
	repo  *purchaserepo.Repository
	stock *stocksvc.Service
	mats  *stockrepo.Repository
}

func New(repo *purchaserepo.Repository, stock *stocksvc.Service, mats *stockrepo.Repository) *Service {
	return &Service{repo: repo, stock: stock, mats: mats}
}

type Operator struct {
	UserID int64
}

type SupplierInput struct {
	Code        string `json:"code" binding:"required,max=64"`
	Name        string `json:"name" binding:"required,max=128"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Status      *bool  `json:"status"`
	Remark      string `json:"remark"`
}

type OrderItemInput struct {
	MaterialID int64  `json:"material_id" binding:"required"`
	Qty        string `json:"qty" binding:"required"`
	UnitPrice  string `json:"unit_price"`
	Remark     string `json:"remark"`
}

type CreateOrderInput struct {
	SupplierID int64            `json:"supplier_id" binding:"required"`
	Remark     string           `json:"remark"`
	Items      []OrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type ReceiptItemInput struct {
	PurchaseOrderItemID int64  `json:"purchase_order_item_id" binding:"required"`
	Qty                 string `json:"qty" binding:"required"`
	Remark              string `json:"remark"`
}

type CreateReceiptInput struct {
	PurchaseOrderID int64              `json:"purchase_order_id" binding:"required"`
	Remark          string             `json:"remark"`
	Items           []ReceiptItemInput `json:"items" binding:"required,min=1,dive"`
}

type OrderDetail struct {
	Order *model.PurchaseOrder      `json:"order"`
	Items []model.PurchaseOrderItem `json:"items"`
}

type ReceiptDetail struct {
	Receipt  *model.PurchaseReceipt      `json:"receipt"`
	Items    []model.PurchaseReceiptItem `json:"items"`
	StockDoc *model.StockDoc             `json:"stock_doc,omitempty"`
}

func (s *Service) CreateSupplier(ctx context.Context, storeID int64, in SupplierInput) (*model.Supplier, error) {
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Supplier{
		StoreID: storeID, Code: in.Code, Name: in.Name, ContactName: in.ContactName,
		Phone: in.Phone, Address: in.Address, Status: status, Remark: in.Remark,
	}
	if err := s.repo.CreateSupplier(ctx, m); err != nil {
		return nil, fmt.Errorf("创建供应商失败: %w", err)
	}
	return m, nil
}

func (s *Service) UpdateSupplier(ctx context.Context, storeID, id int64, in SupplierInput) (*model.Supplier, error) {
	m, err := s.getOwnedSupplier(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	m.Code = in.Code
	m.Name = in.Name
	m.ContactName = in.ContactName
	m.Phone = in.Phone
	m.Address = in.Address
	m.Remark = in.Remark
	if in.Status != nil {
		m.Status = *in.Status
	}
	if err := s.repo.SaveSupplier(ctx, m); err != nil {
		return nil, fmt.Errorf("更新供应商失败: %w", err)
	}
	return m, nil
}

func (s *Service) DeleteSupplier(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwnedSupplier(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDeleteSupplier(ctx, id)
}

func (s *Service) GetSupplier(ctx context.Context, storeID, id int64) (*model.Supplier, error) {
	return s.getOwnedSupplier(ctx, storeID, id)
}

func (s *Service) ListSuppliers(ctx context.Context, storeID int64, keyword string, status *bool, page, pageSize int) ([]model.Supplier, int64, error) {
	return s.repo.ListSuppliers(ctx, purchaserepo.SupplierListFilter{
		StoreID: storeID, Keyword: keyword, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) CreateOrder(ctx context.Context, storeID int64, op Operator, in CreateOrderInput) (*OrderDetail, error) {
	sup, err := s.getOwnedSupplier(ctx, storeID, in.SupplierID)
	if err != nil {
		return nil, err
	}
	if !sup.Status {
		return nil, fmt.Errorf("供应商已停用")
	}
	built, total, err := s.buildOrderItems(ctx, storeID, in.Items)
	if err != nil {
		return nil, err
	}
	var detail OrderDetail
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o := &model.PurchaseOrder{
			StoreID: storeID, SupplierID: in.SupplierID, OrderNo: s.repo.CreateOrderNo(storeID),
			Status: model.PurchaseOrderStatusDraft, TotalAmount: total, Remark: in.Remark, CreatedBy: op.UserID,
		}
		if err := s.repo.CreateOrderTx(ctx, tx, o); err != nil {
			return fmt.Errorf("创建采购单失败: %w", err)
		}
		items := make([]model.PurchaseOrderItem, 0, len(built))
		for _, it := range built {
			it.StoreID = storeID
			it.OrderID = o.ID
			if err := s.repo.CreateOrderItemTx(ctx, tx, &it); err != nil {
				return fmt.Errorf("创建采购明细失败: %w", err)
			}
			items = append(items, it)
		}
		detail.Order = o
		detail.Items = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *Service) GetOrder(ctx context.Context, storeID, id int64) (*OrderDetail, error) {
	o, err := s.getOwnedOrder(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: o, Items: items}, nil
}

func (s *Service) ListOrders(ctx context.Context, storeID, supplierID int64, status string, page, pageSize int) ([]model.PurchaseOrder, int64, error) {
	return s.repo.ListOrders(ctx, purchaserepo.OrderListFilter{
		StoreID: storeID, SupplierID: supplierID, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) SubmitOrder(ctx context.Context, storeID, id int64, op Operator) (*OrderDetail, error) {
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.repo.GetOrderByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("采购单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店采购单")
		}
		if o.Status != model.PurchaseOrderStatusDraft {
			return fmt.Errorf("仅草稿可提交，当前：%s", o.Status)
		}
		items, err := s.repo.ListOrderItemsTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("采购明细不能为空")
		}
		now := time.Now()
		o.Status = model.PurchaseOrderStatusSubmitted
		o.SubmittedBy = op.UserID
		o.SubmittedAt = &now
		return s.repo.SaveOrderTx(ctx, tx, o)
	})
	if err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, storeID, id)
}

func (s *Service) CancelOrder(ctx context.Context, storeID, id int64, op Operator) (*OrderDetail, error) {
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.repo.GetOrderByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("采购单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店采购单")
		}
		if o.Status != model.PurchaseOrderStatusDraft && o.Status != model.PurchaseOrderStatusSubmitted {
			return fmt.Errorf("当前状态不可取消：%s", o.Status)
		}
		items, err := s.repo.ListOrderItemsTx(ctx, tx, id)
		if err != nil {
			return err
		}
		for _, it := range items {
			if it.ReceivedQty.IsPositive() {
				return fmt.Errorf("已有收货，不可取消")
			}
		}
		now := time.Now()
		o.Status = model.PurchaseOrderStatusCancelled
		o.CancelledBy = op.UserID
		o.CancelledAt = &now
		return s.repo.SaveOrderTx(ctx, tx, o)
	})
	if err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, storeID, id)
}

func (s *Service) CreateReceipt(ctx context.Context, storeID int64, op Operator, in CreateReceiptInput) (*ReceiptDetail, error) {
	var detail ReceiptDetail
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.repo.GetOrderByIDTx(ctx, tx, in.PurchaseOrderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("采购单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店采购单")
		}
		if o.Status != model.PurchaseOrderStatusSubmitted && o.Status != model.PurchaseOrderStatusPartialReceived {
			return fmt.Errorf("仅已提交/部分收货的采购单可收货，当前：%s", o.Status)
		}
		receipt := &model.PurchaseReceipt{
			StoreID: storeID, PurchaseOrderID: o.ID, SupplierID: o.SupplierID,
			ReceiptNo: s.repo.CreateReceiptNo(storeID), Status: model.PurchaseReceiptStatusDraft,
			Remark: in.Remark, CreatedBy: op.UserID,
		}
		if err := s.repo.CreateReceiptTx(ctx, tx, receipt); err != nil {
			return fmt.Errorf("创建收货单失败: %w", err)
		}
		items := make([]model.PurchaseReceiptItem, 0, len(in.Items))
		for _, ri := range in.Items {
			poi, err := s.repo.GetOrderItemByIDTx(ctx, tx, ri.PurchaseOrderItemID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("采购明细不存在: %d", ri.PurchaseOrderItemID)
				}
				return err
			}
			if poi.OrderID != o.ID || poi.StoreID != storeID {
				return fmt.Errorf("采购明细不属于该订单")
			}
			qty, err := decimal.NewFromString(ri.Qty)
			if err != nil || !qty.IsPositive() {
				return fmt.Errorf("收货数量无效")
			}
			qty = qty.Round(3)
			remain := poi.Qty.Sub(poi.ReceivedQty)
			if qty.GreaterThan(remain) {
				return fmt.Errorf("物料收货超量，剩余可收：%s", remain.String())
			}
			it := model.PurchaseReceiptItem{
				StoreID: storeID, ReceiptID: receipt.ID, PurchaseOrderItemID: poi.ID,
				MaterialID: poi.MaterialID, Qty: qty, UnitPrice: poi.UnitPrice, Remark: ri.Remark,
			}
			if err := s.repo.CreateReceiptItemTx(ctx, tx, &it); err != nil {
				return fmt.Errorf("创建收货明细失败: %w", err)
			}
			items = append(items, it)
		}
		detail.Receipt = receipt
		detail.Items = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// PostReceipt 确认收货：同事务生成库存入库单并增加结存，回写采购已收数量
func (s *Service) PostReceipt(ctx context.Context, storeID, id int64, op Operator) (*ReceiptDetail, error) {
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		receipt, err := s.repo.GetReceiptByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("收货单不存在")
			}
			return err
		}
		if receipt.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店收货单")
		}
		if receipt.Status != model.PurchaseReceiptStatusDraft {
			return fmt.Errorf("仅草稿收货单可确认入库，当前：%s", receipt.Status)
		}
		items, err := s.repo.ListReceiptItemsTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("收货明细不能为空")
		}

		stockItems := make([]stocksvc.DocItemInput, 0, len(items))
		for _, it := range items {
			poi, err := s.repo.GetOrderItemByIDTx(ctx, tx, it.PurchaseOrderItemID)
			if err != nil {
				return err
			}
			remain := poi.Qty.Sub(poi.ReceivedQty)
			if it.Qty.GreaterThan(remain) {
				return fmt.Errorf("收货超量，订单明细剩余：%s", remain.String())
			}
			poi.ReceivedQty = poi.ReceivedQty.Add(it.Qty).Round(3)
			if err := s.repo.SaveOrderItemTx(ctx, tx, poi); err != nil {
				return err
			}
			stockItems = append(stockItems, stocksvc.DocItemInput{
				MaterialID: it.MaterialID,
				Qty:        it.Qty.String(),
				UnitCost:   it.UnitPrice.String(),
				Remark:     fmt.Sprintf("采购收货 %s", receipt.ReceiptNo),
			})
		}

		// 关键：收货成功必须产生库存入库单据并增加结存
		stockDoc, err := s.stock.CreateAndPostInboundTx(ctx, tx, storeID, stocksvc.Operator{UserID: op.UserID},
			fmt.Sprintf("采购收货单 %s", receipt.ReceiptNo), stockItems)
		if err != nil {
			return err
		}

		o, err := s.repo.GetOrderByIDTx(ctx, tx, receipt.PurchaseOrderID)
		if err != nil {
			return err
		}
		orderItems, err := s.repo.ListOrderItemsTx(ctx, tx, o.ID)
		if err != nil {
			return err
		}
		allReceived := true
		anyReceived := false
		for _, oi := range orderItems {
			if oi.ReceivedQty.IsPositive() {
				anyReceived = true
			}
			if oi.ReceivedQty.LessThan(oi.Qty) {
				allReceived = false
			}
		}
		if allReceived {
			o.Status = model.PurchaseOrderStatusReceived
		} else if anyReceived {
			o.Status = model.PurchaseOrderStatusPartialReceived
		}
		if err := s.repo.SaveOrderTx(ctx, tx, o); err != nil {
			return err
		}

		now := time.Now()
		receipt.Status = model.PurchaseReceiptStatusPosted
		receipt.StockDocID = &stockDoc.ID
		receipt.PostedBy = op.UserID
		receipt.PostedAt = &now
		return s.repo.SaveReceiptTx(ctx, tx, receipt)
	})
	if err != nil {
		return nil, err
	}
	return s.GetReceipt(ctx, storeID, id)
}

func (s *Service) GetReceipt(ctx context.Context, storeID, id int64) (*ReceiptDetail, error) {
	r, err := s.repo.GetReceiptByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("收货单不存在")
		}
		return nil, err
	}
	if r.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店收货单")
	}
	items, err := s.repo.ListReceiptItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ReceiptDetail{Receipt: r, Items: items}, nil
}

func (s *Service) ListReceipts(ctx context.Context, storeID, purchaseOrderID int64, status string, page, pageSize int) ([]model.PurchaseReceipt, int64, error) {
	return s.repo.ListReceipts(ctx, purchaserepo.ReceiptListFilter{
		StoreID: storeID, PurchaseOrderID: purchaseOrderID, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) buildOrderItems(ctx context.Context, storeID int64, items []OrderItemInput) ([]model.PurchaseOrderItem, decimal.Decimal, error) {
	out := make([]model.PurchaseOrderItem, 0, len(items))
	total := decimal.Zero
	for _, in := range items {
		mat, err := s.mats.GetMaterialByID(ctx, in.MaterialID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, decimal.Zero, fmt.Errorf("物料不存在: %d", in.MaterialID)
			}
			return nil, decimal.Zero, err
		}
		if mat.StoreID != storeID {
			return nil, decimal.Zero, fmt.Errorf("物料不属于本门店")
		}
		if !mat.Status {
			return nil, decimal.Zero, fmt.Errorf("物料已停用: %s", mat.Code)
		}
		qty, err := decimal.NewFromString(in.Qty)
		if err != nil || !qty.IsPositive() {
			return nil, decimal.Zero, fmt.Errorf("采购数量无效")
		}
		qty = qty.Round(3)
		price := decimal.Zero
		if in.UnitPrice != "" {
			p, err := decimal.NewFromString(in.UnitPrice)
			if err != nil || p.IsNegative() {
				return nil, decimal.Zero, fmt.Errorf("采购单价无效")
			}
			price = p.Round(2)
		}
		amount := qty.Mul(price).Round(2)
		total = total.Add(amount)
		out = append(out, model.PurchaseOrderItem{
			MaterialID: in.MaterialID, Qty: qty, UnitPrice: price, Amount: amount,
			ReceivedQty: decimal.Zero, Remark: in.Remark,
		})
	}
	return out, total.Round(2), nil
}

func (s *Service) getOwnedSupplier(ctx context.Context, storeID, id int64) (*model.Supplier, error) {
	m, err := s.repo.GetSupplierByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("供应商不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店供应商")
	}
	return m, nil
}

func (s *Service) getOwnedOrder(ctx context.Context, storeID, id int64) (*model.PurchaseOrder, error) {
	m, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("采购单不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店采购单")
	}
	return m, nil
}
