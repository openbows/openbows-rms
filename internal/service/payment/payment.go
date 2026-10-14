package payment

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
	paymentrepo "openbows-rms/internal/repository/payment"
	tablerepo "openbows-rms/internal/repository/table"
	membersvc "openbows-rms/internal/service/member"
	ordersvc "openbows-rms/internal/service/order"
	stocksvc "openbows-rms/internal/service/stock"
)

type Service struct {
	payments *paymentrepo.Repository
	orders   *orderrepo.Repository
	tables   *tablerepo.TableRepo
	dishes   *menurepo.DishRepo
	stock    *stocksvc.Service
	members  *membersvc.Service
}

func New(
	payments *paymentrepo.Repository,
	orders *orderrepo.Repository,
	tables *tablerepo.TableRepo,
	dishes *menurepo.DishRepo,
	stock *stocksvc.Service,
	members *membersvc.Service,
) *Service {
	return &Service{
		payments: payments, orders: orders, tables: tables,
		dishes: dishes, stock: stock, members: members,
	}
}

type PayInput struct {
	Channel    string `json:"channel" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	MemberID   *int64 `json:"member_id"`
	OutTradeNo string `json:"out_trade_no"`
	Remark     string `json:"remark"`
}

type PayResult struct {
	Payment *model.Payment `json:"payment"`
	Order   *model.Order   `json:"order"`
}

func (s *Service) Pay(ctx context.Context, storeID, orderID int64, op ordersvc.Operator, in PayInput) (*PayResult, error) {
	if !validChannel(in.Channel) {
		return nil, fmt.Errorf("无效支付渠道，可选：cash/wechat/alipay/other/member")
	}
	if in.Channel == model.PaymentChannelMember && (in.MemberID == nil || *in.MemberID <= 0) {
		return nil, fmt.Errorf("会员支付须指定 member_id")
	}
	payAmt, err := decimal.NewFromString(in.Amount)
	if err != nil || !payAmt.IsPositive() {
		return nil, fmt.Errorf("支付金额无效")
	}
	payAmt = payAmt.Round(2)

	var paymentID int64
	err = s.payments.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		if o.Status != model.OrderStatusPendingPay {
			return fmt.Errorf("仅待结账订单可支付，当前状态：%s", o.Status)
		}

		n, err := s.payments.CountSuccessByOrderTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("订单已支付，禁止重复支付")
		}
		if !payAmt.Equal(o.PayableAmount) {
			return fmt.Errorf("支付金额须等于应付金额 %s", o.PayableAmount.StringFixed(2))
		}

		paymentNo := s.payments.CreatePaymentNo(storeID, orderID)

		// 关键：会员余额扣款与支付记录同事务
		if in.Channel == model.PaymentChannelMember {
			if s.members == nil {
				return fmt.Errorf("会员服务未配置")
			}
			oid := orderID
			if err := s.members.ConsumeTx(ctx, tx, storeID, *in.MemberID, membersvc.Operator{UserID: op.UserID}, membersvc.AmountInput{
				Amount:  payAmt.StringFixed(2),
				BizNo:   paymentNo,
				OrderID: &oid,
				Remark:  "订单支付",
			}); err != nil {
				return err
			}
			o.MemberID = in.MemberID
		}

		p := &model.Payment{
			StoreID:    storeID,
			OrderID:    orderID,
			PaymentNo:  paymentNo,
			Channel:    in.Channel,
			Amount:     payAmt,
			Status:     model.PaymentStatusSuccess,
			PaidAt:     time.Now(),
			OutTradeNo: in.OutTradeNo,
			OperatorID: op.UserID,
			Remark:     in.Remark,
		}
		if err := s.payments.CreateTx(ctx, tx, p); err != nil {
			return fmt.Errorf("创建支付记录失败: %w", err)
		}
		paymentID = p.ID

		// 关键：按配方出库，结存不足则整单回滚
		if err := s.deductStockByOrderTx(ctx, tx, storeID, orderID, op.UserID); err != nil {
			return err
		}

		from := o.Status
		o.Status = model.OrderStatusCompleted
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
			StoreID:      storeID,
			OrderID:      orderID,
			FromStatus:   from,
			ToStatus:     model.OrderStatusCompleted,
			OperatorID:   op.UserID,
			OperatorName: op.Username,
			Remark:       "支付成功：" + in.Channel,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	list, err := s.payments.ListByOrder(ctx, storeID, orderID)
	if err != nil {
		return nil, err
	}
	var paid *model.Payment
	for i := range list {
		if list[i].ID == paymentID {
			paid = &list[i]
			break
		}
	}
	o, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &PayResult{Payment: paid, Order: o}, nil
}

func (s *Service) deductStockByOrderTx(ctx context.Context, tx *gorm.DB, storeID, orderID, userID int64) error {
	if s.stock == nil || s.dishes == nil {
		return nil
	}
	items, err := s.orders.ListItemsTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	dishIDs := make([]int64, 0)
	qtyByDish := map[int64]int{}
	for _, it := range items {
		if it.Status != model.OrderItemStatusNormal {
			continue
		}
		if _, ok := qtyByDish[it.DishID]; !ok {
			dishIDs = append(dishIDs, it.DishID)
		}
		qtyByDish[it.DishID] += it.Qty
	}
	if len(dishIDs) == 0 {
		return nil
	}
	boms, err := s.dishes.ListMaterialsByDishIDs(ctx, dishIDs)
	if err != nil {
		return err
	}
	need := map[int64]decimal.Decimal{}
	for _, b := range boms {
		dishQty := qtyByDish[b.DishID]
		if dishQty <= 0 {
			continue
		}
		add := b.Qty.Mul(decimal.NewFromInt(int64(dishQty)))
		need[b.MaterialID] = need[b.MaterialID].Add(add)
	}
	if len(need) == 0 {
		return nil
	}
	docItems := make([]stocksvc.DocItemInput, 0, len(need))
	for mid, qty := range need {
		docItems = append(docItems, stocksvc.DocItemInput{
			MaterialID: mid,
			Qty:        qty.Round(3).String(),
		})
	}
	_, err = s.stock.CreateAndPostOutboundTx(ctx, tx, storeID, stocksvc.Operator{UserID: userID},
		fmt.Sprintf("订单结账出库#%d", orderID), docItems)
	return err
}

func (s *Service) ListByOrder(ctx context.Context, storeID, orderID int64) ([]model.Payment, error) {
	o, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, err
	}
	if o.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店订单")
	}
	return s.payments.ListByOrder(ctx, storeID, orderID)
}

func validChannel(ch string) bool {
	switch ch {
	case model.PaymentChannelCash, model.PaymentChannelWechat, model.PaymentChannelAlipay,
		model.PaymentChannelOther, model.PaymentChannelMember:
		return true
	default:
		return false
	}
}
