package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	financerepo "openbows-rms/internal/repository/finance"
	paymentrepo "openbows-rms/internal/repository/payment"
)

var locAsiaShanghai = func() *time.Location {
	l, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return l
}()

type Service struct {
	finance  *financerepo.Repository
	payments *paymentrepo.Repository
}

func New(finance *financerepo.Repository, payments *paymentrepo.Repository) *Service {
	return &Service{finance: finance, payments: payments}
}

type Operator struct {
	UserID int64
}

type DailySettleInput struct {
	BusinessDay string `json:"business_day"` // YYYY-MM-DD，默认当日
	Remark      string `json:"remark"`
}

type ChannelEntryInput struct {
	OutTradeNo string `json:"out_trade_no"`
	Amount     string `json:"amount" binding:"required"`
}

type CreateReconInput struct {
	BusinessDay   string              `json:"business_day" binding:"required"`
	Channel       string              `json:"channel" binding:"required"`
	ChannelAmount string              `json:"channel_amount" binding:"required"`
	Remark        string              `json:"remark"`
	Entries       []ChannelEntryInput `json:"entries"` // 可选逐笔；为空则仅总额对账
}

type SettlementDetail struct {
	Settlement *model.Settlement      `json:"settlement"`
	Items      []model.SettlementItem `json:"items"`
}

type ReconDetail struct {
	Reconciliation *model.Reconciliation      `json:"reconciliation"`
	Items          []model.ReconciliationItem `json:"items"`
}

func (s *Service) DailySettle(ctx context.Context, storeID int64, op Operator, in DailySettleInput) (*SettlementDetail, error) {
	day, err := parseBusinessDay(in.BusinessDay)
	if err != nil {
		return nil, err
	}
	if _, err := s.finance.GetSettlementByDay(ctx, storeID, day); err == nil {
		return nil, fmt.Errorf("该营业日已日结，禁止重复日结")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	start, end := dayRange(day)
	payments, err := s.payments.ListSuccessByStoreDay(ctx, storeID, start, end, "")
	if err != nil {
		return nil, err
	}

	var detail SettlementDetail
	err = s.finance.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		st := &model.Settlement{
			StoreID: storeID, BusinessDay: day, Status: model.SettlementStatusClosed,
			Remark: in.Remark, CreatedBy: op.UserID,
			TotalAmount: decimal.Zero, CashAmount: decimal.Zero, WechatAmount: decimal.Zero,
			AlipayAmount: decimal.Zero, OtherAmount: decimal.Zero,
		}
		orderSet := map[int64]struct{}{}
		for _, p := range payments {
			st.PaymentCount++
			orderSet[p.OrderID] = struct{}{}
			st.TotalAmount = st.TotalAmount.Add(p.Amount)
			switch p.Channel {
			case model.PaymentChannelCash:
				st.CashAmount = st.CashAmount.Add(p.Amount)
			case model.PaymentChannelWechat:
				st.WechatAmount = st.WechatAmount.Add(p.Amount)
			case model.PaymentChannelAlipay:
				st.AlipayAmount = st.AlipayAmount.Add(p.Amount)
			default:
				st.OtherAmount = st.OtherAmount.Add(p.Amount)
			}
		}
		st.OrderCount = len(orderSet)
		st.TotalAmount = st.TotalAmount.Round(2)
		st.CashAmount = st.CashAmount.Round(2)
		st.WechatAmount = st.WechatAmount.Round(2)
		st.AlipayAmount = st.AlipayAmount.Round(2)
		st.OtherAmount = st.OtherAmount.Round(2)

		if err := s.finance.CreateSettlementTx(ctx, tx, st); err != nil {
			return fmt.Errorf("创建日结失败（可能已日结）: %w", err)
		}
		items := make([]model.SettlementItem, 0, len(payments))
		for _, p := range payments {
			it := model.SettlementItem{
				StoreID: storeID, SettlementID: st.ID, PaymentID: p.ID, OrderID: p.OrderID,
				Channel: p.Channel, Amount: p.Amount, PaidAt: p.PaidAt,
			}
			if err := s.finance.CreateSettlementItemTx(ctx, tx, &it); err != nil {
				return err
			}
			items = append(items, it)
		}
		detail.Settlement = st
		detail.Items = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *Service) ListSettlements(ctx context.Context, storeID int64, businessDay string, page, pageSize int) ([]model.Settlement, int64, error) {
	f := financerepo.SettlementListFilter{StoreID: storeID, Page: page, PageSize: pageSize}
	if businessDay != "" {
		d, err := parseBusinessDay(businessDay)
		if err != nil {
			return nil, 0, err
		}
		f.BusinessDay = &d
	}
	return s.finance.ListSettlements(ctx, f)
}

func (s *Service) GetSettlement(ctx context.Context, storeID, id int64) (*SettlementDetail, error) {
	st, err := s.finance.GetSettlementByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("日结不存在")
		}
		return nil, err
	}
	if st.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店日结")
	}
	items, err := s.finance.ListSettlementItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SettlementDetail{Settlement: st, Items: items}, nil
}

func (s *Service) CreateReconciliation(ctx context.Context, storeID int64, op Operator, in CreateReconInput) (*ReconDetail, error) {
	day, err := parseBusinessDay(in.BusinessDay)
	if err != nil {
		return nil, err
	}
	if !validPayChannel(in.Channel) {
		return nil, fmt.Errorf("无效支付渠道")
	}
	channelAmt, err := parseAmount(in.ChannelAmount)
	if err != nil {
		return nil, fmt.Errorf("渠道金额无效")
	}

	start, end := dayRange(day)
	payments, err := s.payments.ListSuccessByStoreDay(ctx, storeID, start, end, in.Channel)
	if err != nil {
		return nil, err
	}
	systemAmt := decimal.Zero
	for _, p := range payments {
		systemAmt = systemAmt.Add(p.Amount)
	}
	systemAmt = systemAmt.Round(2)
	diff := channelAmt.Sub(systemAmt).Round(2)

	var detail ReconDetail
	err = s.finance.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		recon := &model.Reconciliation{
			StoreID: storeID, BusinessDay: day, Channel: in.Channel,
			SystemAmount: systemAmt, ChannelAmount: channelAmt, DiffAmount: diff,
			SystemCount: len(payments), ChannelCount: len(in.Entries),
			Remark: in.Remark, CreatedBy: op.UserID,
		}
		if !diff.IsZero() || len(in.Entries) > 0 {
			// 状态在明细比对后最终确定
		}
		items := make([]model.ReconciliationItem, 0)

		if len(in.Entries) == 0 {
			// 仅总额对账
			if !diff.IsZero() {
				recon.Status = model.ReconStatusMismatched
				items = append(items, model.ReconciliationItem{
					StoreID: storeID, SystemAmount: systemAmt, ChannelAmount: channelAmt,
					DiffAmount: diff, DiffType: model.ReconDiffTotalDiff,
					Remark: "总额差异",
				})
			} else {
				recon.Status = model.ReconStatusMatched
			}
		} else {
			// 按 out_trade_no 逐笔对
			payByTrade := map[string]model.Payment{}
			payNoTrade := make([]model.Payment, 0)
			for _, p := range payments {
				if p.OutTradeNo != "" {
					payByTrade[p.OutTradeNo] = p
				} else {
					payNoTrade = append(payNoTrade, p)
				}
			}
			usedPay := map[int64]struct{}{}
			channelSumCheck := decimal.Zero
			hasDiff := false
			for _, e := range in.Entries {
				amt, err := parseAmount(e.Amount)
				if err != nil {
					return fmt.Errorf("渠道明细金额无效")
				}
				channelSumCheck = channelSumCheck.Add(amt)
				p, ok := payByTrade[e.OutTradeNo]
				if !ok || e.OutTradeNo == "" {
					hasDiff = true
					items = append(items, model.ReconciliationItem{
						StoreID: storeID, OutTradeNo: e.OutTradeNo,
						ChannelAmount: amt, DiffAmount: amt, DiffType: model.ReconDiffMissingInSystem,
						Remark: "渠道有、系统无",
					})
					continue
				}
				usedPay[p.ID] = struct{}{}
				oid := p.OrderID
				pid := p.ID
				d := amt.Sub(p.Amount).Round(2)
				diffType := model.ReconDiffMatch
				if !d.IsZero() {
					diffType = model.ReconDiffAmountMismatch
					hasDiff = true
				}
				items = append(items, model.ReconciliationItem{
					StoreID: storeID, PaymentID: &pid, OrderID: &oid, OutTradeNo: e.OutTradeNo,
					SystemAmount: p.Amount, ChannelAmount: amt, DiffAmount: d, DiffType: diffType,
				})
			}
			for _, p := range payments {
				if _, ok := usedPay[p.ID]; ok {
					continue
				}
				// 系统有、渠道无（含无 out_trade_no 的现金等）
				hasDiff = true
				pid := p.ID
				oid := p.OrderID
				items = append(items, model.ReconciliationItem{
					StoreID: storeID, PaymentID: &pid, OrderID: &oid, OutTradeNo: p.OutTradeNo,
					SystemAmount: p.Amount, DiffAmount: p.Amount.Neg(), DiffType: model.ReconDiffMissingInChannel,
					Remark: "系统有、渠道无",
				})
			}
			_ = payNoTrade
			if !channelSumCheck.Round(2).Equal(channelAmt) {
				// 明细合计与传入总额不一致时，补一条总额差异提示（不改主金额）
				hasDiff = true
			}
			if hasDiff || !diff.IsZero() {
				recon.Status = model.ReconStatusMismatched
			} else {
				recon.Status = model.ReconStatusMatched
			}
			recon.ChannelCount = len(in.Entries)
		}

		if err := s.finance.CreateReconciliationTx(ctx, tx, recon); err != nil {
			return fmt.Errorf("创建对账失败: %w", err)
		}
		saved := make([]model.ReconciliationItem, 0, len(items))
		for i := range items {
			items[i].ReconciliationID = recon.ID
			if err := s.finance.CreateReconciliationItemTx(ctx, tx, &items[i]); err != nil {
				return err
			}
			saved = append(saved, items[i])
		}
		detail.Reconciliation = recon
		detail.Items = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *Service) ListReconciliations(ctx context.Context, storeID int64, businessDay, channel, status string, page, pageSize int) ([]model.Reconciliation, int64, error) {
	f := financerepo.ReconListFilter{StoreID: storeID, Channel: channel, Status: status, Page: page, PageSize: pageSize}
	if businessDay != "" {
		d, err := parseBusinessDay(businessDay)
		if err != nil {
			return nil, 0, err
		}
		f.BusinessDay = &d
	}
	return s.finance.ListReconciliations(ctx, f)
}

func (s *Service) GetReconciliation(ctx context.Context, storeID, id int64) (*ReconDetail, error) {
	r, err := s.finance.GetReconciliationByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("对账记录不存在")
		}
		return nil, err
	}
	if r.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店对账")
	}
	items, err := s.finance.ListReconciliationItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ReconDetail{Reconciliation: r, Items: items}, nil
}

func (s *Service) ListDiffs(ctx context.Context, storeID int64, businessDay, channel string, page, pageSize int) ([]model.ReconciliationItem, int64, error) {
	f := financerepo.DiffListFilter{StoreID: storeID, Channel: channel, Page: page, PageSize: pageSize}
	if businessDay != "" {
		d, err := parseBusinessDay(businessDay)
		if err != nil {
			return nil, 0, err
		}
		f.BusinessDay = &d
	}
	return s.finance.ListDiffItems(ctx, f)
}

func parseBusinessDay(s string) (time.Time, error) {
	if s == "" {
		now := time.Now().In(locAsiaShanghai)
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, locAsiaShanghai), nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, locAsiaShanghai)
	if err != nil {
		return time.Time{}, fmt.Errorf("营业日格式无效，应为 YYYY-MM-DD")
	}
	return t, nil
}

func dayRange(day time.Time) (time.Time, time.Time) {
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, locAsiaShanghai)
	return day, day.Add(24 * time.Hour)
}

func parseAmount(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil || d.IsNegative() {
		return decimal.Zero, fmt.Errorf("金额无效")
	}
	return d.Round(2), nil
}

func validPayChannel(c string) bool {
	switch c {
	case model.PaymentChannelCash, model.PaymentChannelWechat, model.PaymentChannelAlipay, model.PaymentChannelOther:
		return true
	default:
		return false
	}
}
