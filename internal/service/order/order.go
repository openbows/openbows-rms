package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	menurepo "openbows-rms/internal/repository/menu"
	orderrepo "openbows-rms/internal/repository/order"
	tablerepo "openbows-rms/internal/repository/table"
)

// KitchenSyncer 下单/加菜后同步厨单（由 KDS 服务实现，避免循环依赖）
type KitchenSyncer interface {
	CreateTicketsForOrder(ctx context.Context, tx *gorm.DB, storeID int64, order *model.Order, items []model.OrderItem) error
}

type Service struct {
	orders  *orderrepo.Repository
	tables  *tablerepo.TableRepo
	dishes  *menurepo.DishRepo
	kitchen KitchenSyncer
}

func New(orders *orderrepo.Repository, tables *tablerepo.TableRepo, dishes *menurepo.DishRepo) *Service {
	return &Service{orders: orders, tables: tables, dishes: dishes}
}

func (s *Service) SetKitchenSyncer(k KitchenSyncer) { s.kitchen = k }

type Operator struct {
	UserID   int64
	Username string
}

type ModifierInput struct {
	OptionID int64 `json:"option_id"`
	Qty      int   `json:"qty"`
}

type ItemInput struct {
	DishID     int64           `json:"dish_id" binding:"required"`
	DishSpecID *int64          `json:"dish_spec_id"`
	Qty        int             `json:"qty" binding:"required,min=1"`
	Remark     string          `json:"remark"`
	Modifiers  []ModifierInput `json:"modifiers"`
}

type FeeInput struct {
	FeeType string `json:"fee_type" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Amount  string `json:"amount" binding:"required"`
	Remark  string `json:"remark"`
}

type CreateInput struct {
	TableID    int64       `json:"table_id" binding:"required"`
	GuestCount int         `json:"guest_count"`
	Remark     string      `json:"remark"`
	Channel    string      `json:"channel"`  // 默认 dine_in
	MemberID   *int64      `json:"member_id"`
	Items      []ItemInput `json:"items" binding:"required,min=1,dive"`
	Fees       []FeeInput  `json:"fees"`
	// AfterCreateTx 在创建订单同一事务内回调（如写入 online_order_extras）
	AfterCreateTx func(ctx context.Context, tx *gorm.DB, order *model.Order) error `json:"-"`
}

type AddItemsInput struct {
	Items []ItemInput `json:"items" binding:"required,min=1,dive"`
}

type ReturnItemInput struct {
	Remark string `json:"remark"`
}

type AdvanceInput struct {
	Remark string `json:"remark"`
}

type CancelInput struct {
	Remark string `json:"remark"`
}

type ItemView struct {
	Item      model.OrderItem           `json:"item"`
	Modifiers []model.OrderItemModifier `json:"modifiers"`
}

type Detail struct {
	Order  *model.Order           `json:"order"`
	Items  []ItemView             `json:"items"`
	Fees   []model.OrderFeeItem   `json:"fees"`
	Logs   []model.OrderStatusLog `json:"status_logs"`
}

func (s *Service) Create(ctx context.Context, storeID int64, op Operator, in CreateInput) (*Detail, error) {
	if in.GuestCount <= 0 {
		in.GuestCount = 1
	}
	builtItems, builtFees, subtotal, feeSum, discountSum, err := s.buildAmounts(ctx, storeID, in.Items, in.Fees)
	if err != nil {
		return nil, err
	}
	payable := subtotal.Add(feeSum).Sub(discountSum)
	if payable.IsNegative() {
		payable = decimal.Zero
	}

	var orderID int64
	err = s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tbl, err := s.tables.GetByIDTx(ctx, tx, in.TableID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("桌台不存在")
			}
			return err
		}
		if tbl.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店桌台")
		}
		if tbl.Status != model.TableStatusIdle {
			return fmt.Errorf("桌台非空闲，无法开台")
		}

		channel := in.Channel
		if channel == "" {
			channel = model.OrderChannelDineIn
		}
		if !validOrderChannel(channel) {
			return fmt.Errorf("无效下单渠道")
		}
		now := time.Now()
		o := &model.Order{
			StoreID:        storeID,
			TableID:        in.TableID,
			OrderNo:        s.orders.CreateOrderNo(storeID),
			Channel:        channel,
			MemberID:       in.MemberID,
			Status:         model.OrderStatusPendingConfirm,
			GuestCount:     in.GuestCount,
			SubtotalAmount: subtotal,
			FeeAmount:      feeSum,
			DiscountAmount: discountSum,
			PayableAmount:  payable,
			Remark:         in.Remark,
			CreatedBy:      op.UserID,
			UpdatedBy:      op.UserID,
			OpenedAt:       now,
		}
		if err := tx.Create(o).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}
		orderID = o.ID

		createdItems := make([]model.OrderItem, 0, len(builtItems))
		for i := range builtItems {
			builtItems[i].StoreID = storeID
			builtItems[i].OrderID = o.ID
			mods := builtItems[i].modifiers
			item := builtItems[i].OrderItem
			item.KitchenStatus = model.KitchenStatusPending
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("创建明细失败: %w", err)
			}
			createdItems = append(createdItems, item)
			for j := range mods {
				mods[j].StoreID = storeID
				mods[j].OrderItemID = item.ID
				if err := tx.Create(&mods[j]).Error; err != nil {
					return fmt.Errorf("创建加料失败: %w", err)
				}
			}
		}
		for i := range builtFees {
			builtFees[i].StoreID = storeID
			builtFees[i].OrderID = o.ID
			if err := tx.Create(&builtFees[i]).Error; err != nil {
				return fmt.Errorf("创建费用分项失败: %w", err)
			}
		}

		// 关键：开台占用桌台并挂单
		tbl.Status = model.TableStatusOccupied
		tbl.CurrentOrderID = &o.ID
		if err := s.tables.SaveTx(ctx, tx, tbl); err != nil {
			return err
		}
		logRemark := "开台下单"
		if channel == model.OrderChannelOnlineH5 || channel == model.OrderChannelOnlineScan {
			logRemark = "线上点餐下单"
		}
		if err := tx.Create(&model.OrderStatusLog{
			StoreID:      storeID,
			OrderID:      o.ID,
			FromStatus:   "",
			ToStatus:     model.OrderStatusPendingConfirm,
			OperatorID:   op.UserID,
			OperatorName: op.Username,
			Remark:       logRemark,
		}).Error; err != nil {
			return err
		}
		// 关键：按档口同步厨单
		if s.kitchen != nil {
			if err := s.kitchen.CreateTicketsForOrder(ctx, tx, storeID, o, createdItems); err != nil {
				return err
			}
		}
		if in.AfterCreateTx != nil {
			if err := in.AfterCreateTx(ctx, tx, o); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, storeID, orderID)
}

func (s *Service) List(ctx context.Context, storeID, tableID int64, status string, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.orders.List(ctx, orderrepo.ListFilter{
		StoreID: storeID, TableID: tableID, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) GetDetail(ctx context.Context, storeID, id int64) (*Detail, error) {
	o, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	items, err := s.orders.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	mods, err := s.orders.ListModifiersByItemIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	modMap := map[int64][]model.OrderItemModifier{}
	for _, m := range mods {
		modMap[m.OrderItemID] = append(modMap[m.OrderItemID], m)
	}
	views := make([]ItemView, 0, len(items))
	for _, it := range items {
		views = append(views, ItemView{Item: it, Modifiers: modMap[it.ID]})
	}
	fees, err := s.orders.ListFeeItems(ctx, id)
	if err != nil {
		return nil, err
	}
	logs, err := s.orders.ListStatusLogs(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Detail{Order: o, Items: views, Fees: fees, Logs: logs}, nil
}

func (s *Service) AddItems(ctx context.Context, storeID, orderID int64, op Operator, in AddItemsInput) (*Detail, error) {
	err := s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.orders.GetByIDTx(ctx, tx, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("订单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店订单")
		}
		if !editableStatus(o.Status) {
			return fmt.Errorf("当前状态不允许改单")
		}
		builtItems, _, addSub, _, _, err := s.buildAmounts(ctx, storeID, in.Items, nil)
		if err != nil {
			return err
		}
		createdItems := make([]model.OrderItem, 0, len(builtItems))
		for i := range builtItems {
			item := builtItems[i].OrderItem
			item.StoreID = storeID
			item.OrderID = orderID
			item.KitchenStatus = model.KitchenStatusPending
			mods := builtItems[i].modifiers
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			createdItems = append(createdItems, item)
			for j := range mods {
				mods[j].StoreID = storeID
				mods[j].OrderItemID = item.ID
				if err := tx.Create(&mods[j]).Error; err != nil {
					return err
				}
			}
		}
		o.SubtotalAmount = o.SubtotalAmount.Add(addSub)
		o.PayableAmount = o.SubtotalAmount.Add(o.FeeAmount).Sub(o.DiscountAmount)
		if o.PayableAmount.IsNegative() {
			o.PayableAmount = decimal.Zero
		}
		o.UpdatedBy = op.UserID
		if err := tx.Save(o).Error; err != nil {
			return err
		}
		// 关键：加菜同步厨单
		if s.kitchen != nil {
			if err := s.kitchen.CreateTicketsForOrder(ctx, tx, storeID, o, createdItems); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, storeID, orderID)
}

func (s *Service) ReturnItem(ctx context.Context, storeID, orderID, itemID int64, op Operator, in ReturnItemInput) (*Detail, error) {
	err := s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.orders.GetByIDTx(ctx, tx, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("订单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店订单")
		}
		if !editableStatus(o.Status) {
			return fmt.Errorf("当前状态不允许退菜")
		}
		var item model.OrderItem
		if err := tx.First(&item, itemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("明细不存在")
			}
			return err
		}
		if item.OrderID != orderID || item.StoreID != storeID {
			return fmt.Errorf("明细不属于该订单")
		}
		if item.Status == model.OrderItemStatusReturned {
			return fmt.Errorf("明细已退菜")
		}
		item.Status = model.OrderItemStatusReturned
		if in.Remark != "" {
			item.Remark = in.Remark
		}
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		o.SubtotalAmount = o.SubtotalAmount.Sub(item.Amount)
		if o.SubtotalAmount.IsNegative() {
			o.SubtotalAmount = decimal.Zero
		}
		o.PayableAmount = o.SubtotalAmount.Add(o.FeeAmount).Sub(o.DiscountAmount)
		if o.PayableAmount.IsNegative() {
			o.PayableAmount = decimal.Zero
		}
		o.UpdatedBy = op.UserID
		return tx.Save(o).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, storeID, orderID)
}

func (s *Service) Advance(ctx context.Context, storeID, orderID int64, op Operator, in AdvanceInput) (*Detail, error) {
	err := s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.orders.GetByIDTx(ctx, tx, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("订单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店订单")
		}
		next, ok := nextStatus(o.Status)
		if !ok {
			return fmt.Errorf("当前状态禁止推进：%s", o.Status)
		}
		from := o.Status
		o.Status = next
		o.UpdatedBy = op.UserID
		if next == model.OrderStatusCompleted {
			now := time.Now()
			o.ClosedAt = &now
			tbl, err := s.tables.GetByIDTx(ctx, tx, o.TableID)
			if err == nil && tbl.StoreID == storeID {
				tbl.Status = model.TableStatusIdle
				tbl.CurrentOrderID = nil
				_ = s.tables.SaveTx(ctx, tx, tbl)
			}
		}
		if err := tx.Save(o).Error; err != nil {
			return err
		}
		return tx.Create(&model.OrderStatusLog{
			StoreID: storeID, OrderID: orderID, FromStatus: from, ToStatus: next,
			OperatorID: op.UserID, OperatorName: op.Username, Remark: in.Remark,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, storeID, orderID)
}

func (s *Service) Cancel(ctx context.Context, storeID, orderID int64, op Operator, in CancelInput) (*Detail, error) {
	err := s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.orders.GetByIDTx(ctx, tx, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("订单不存在")
			}
			return err
		}
		if o.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店订单")
		}
		if o.Status == model.OrderStatusCompleted || o.Status == model.OrderStatusCancelled {
			return fmt.Errorf("当前状态不可取消")
		}
		from := o.Status
		o.Status = model.OrderStatusCancelled
		o.UpdatedBy = op.UserID
		now := time.Now()
		o.ClosedAt = &now
		if err := tx.Save(o).Error; err != nil {
			return err
		}
		tbl, err := s.tables.GetByIDTx(ctx, tx, o.TableID)
		if err == nil && tbl.StoreID == storeID {
			tbl.Status = model.TableStatusIdle
			tbl.CurrentOrderID = nil
			if err := s.tables.SaveTx(ctx, tx, tbl); err != nil {
				return err
			}
		}
		return tx.Create(&model.OrderStatusLog{
			StoreID: storeID, OrderID: orderID, FromStatus: from, ToStatus: model.OrderStatusCancelled,
			OperatorID: op.UserID, OperatorName: op.Username, Remark: in.Remark,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, storeID, orderID)
}

type builtItem struct {
	model.OrderItem
	modifiers []model.OrderItemModifier
}

func (s *Service) buildAmounts(ctx context.Context, storeID int64, items []ItemInput, fees []FeeInput) ([]builtItem, []model.OrderFeeItem, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	built := make([]builtItem, 0, len(items))
	subtotal := decimal.Zero
	for idx, it := range items {
		dish, err := s.dishes.GetByID(ctx, it.DishID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("菜品不存在: %d", it.DishID)
			}
			return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		if dish.StoreID != storeID {
			return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("菜品不属于本门店")
		}
		unit := dish.BasePrice
		specName := ""
		var specID *int64
		if it.DishSpecID != nil {
			specs, err := s.dishes.ListSpecs(ctx, dish.ID)
			if err != nil {
				return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, err
			}
			found := false
			for _, sp := range specs {
				if sp.ID == *it.DishSpecID {
					unit = sp.Price
					specName = sp.Name
					sid := sp.ID
					specID = &sid
					found = true
					break
				}
			}
			if !found {
				return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("规格不存在")
			}
		}
		mods := make([]model.OrderItemModifier, 0, len(it.Modifiers))
		modExtra := decimal.Zero
		if len(it.Modifiers) > 0 {
			options, err := s.dishes.ListOptions(ctx, dish.ID)
			if err != nil {
				return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, err
			}
			optMap := map[int64]model.DishOption{}
			for _, op := range options {
				optMap[op.ID] = op
			}
			for _, mi := range it.Modifiers {
				op, ok := optMap[mi.OptionID]
				if !ok {
					return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("加料不存在: %d", mi.OptionID)
				}
				q := mi.Qty
				if q <= 0 {
					q = 1
				}
				modExtra = modExtra.Add(op.Price.Mul(decimal.NewFromInt(int64(q))))
				oid := op.ID
				mods = append(mods, model.OrderItemModifier{
					OptionID: &oid, OptionName: op.Name, Price: op.Price, Qty: q,
				})
			}
		}
		lineUnit := unit.Add(modExtra)
		amount := lineUnit.Mul(decimal.NewFromInt(int64(it.Qty))).Round(2)
		subtotal = subtotal.Add(amount)
		built = append(built, builtItem{
			OrderItem: model.OrderItem{
				DishID: dish.ID, DishName: dish.Name, DishSpecID: specID, SpecName: specName,
				UnitPrice: lineUnit.Round(2), Qty: it.Qty, Amount: amount,
				Status: model.OrderItemStatusNormal, Remark: it.Remark, SortNo: idx,
			},
			modifiers: mods,
		})
	}

	builtFees := make([]model.OrderFeeItem, 0, len(fees))
	feeSum := decimal.Zero
	discountSum := decimal.Zero
	for _, f := range fees {
		switch f.FeeType {
		case model.FeeTypeDiscount, model.FeeTypeService, model.FeeTypeTax, model.FeeTypeOther:
		default:
			return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("无效费用类型")
		}
		amt, err := decimal.NewFromString(f.Amount)
		if err != nil || amt.IsNegative() {
			return nil, nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("费用金额无效")
		}
		amt = amt.Round(2)
		if f.FeeType == model.FeeTypeDiscount {
			discountSum = discountSum.Add(amt)
		} else {
			feeSum = feeSum.Add(amt)
		}
		builtFees = append(builtFees, model.OrderFeeItem{
			FeeType: f.FeeType, Name: f.Name, Amount: amt, Remark: f.Remark,
		})
	}
	return built, builtFees, subtotal.Round(2), feeSum.Round(2), discountSum.Round(2), nil
}

func (s *Service) getOwned(ctx context.Context, storeID, id int64) (*model.Order, error) {
	o, err := s.orders.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, err
	}
	if o.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店订单")
	}
	return o, nil
}

func editableStatus(status string) bool {
	return status == model.OrderStatusPendingConfirm || status == model.OrderStatusCooking
}

func validOrderChannel(c string) bool {
	switch c {
	case model.OrderChannelDineIn, model.OrderChannelOnlineH5, model.OrderChannelOnlineScan, model.OrderChannelTakeout:
		return true
	default:
		return false
	}
}

func nextStatus(cur string) (string, bool) {
	switch cur {
	case model.OrderStatusPendingConfirm:
		return model.OrderStatusCooking, true
	case model.OrderStatusCooking:
		return model.OrderStatusPendingPay, true
	case model.OrderStatusPendingPay:
		// 完结请走支付接口 /orders/:id/pay
		return "", false
	default:
		return "", false
	}
}
