package reservation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	orderrepo "openbows-rms/internal/repository/order"
	resrepo "openbows-rms/internal/repository/reservation"
	tablerepo "openbows-rms/internal/repository/table"
)

type Service struct {
	res    *resrepo.Repository
	tables *tablerepo.TableRepo
	orders *orderrepo.Repository
}

func New(res *resrepo.Repository, tables *tablerepo.TableRepo, orders *orderrepo.Repository) *Service {
	return &Service{res: res, tables: tables, orders: orders}
}

type Operator struct {
	UserID   int64
	Username string
}

type CreateInput struct {
	TableID       *int64 `json:"table_id"`
	CustomerName  string `json:"customer_name" binding:"required,max=128"`
	Phone         string `json:"phone" binding:"max=32"`
	GuestCount    int    `json:"guest_count"`
	ReserveAt     string `json:"reserve_at" binding:"required"` // RFC3339
	Channel       string `json:"channel"`
	Remark        string `json:"remark"`
	DepositAmount string `json:"deposit_amount"`
}

type UpdateInput struct {
	TableID       *int64 `json:"table_id"`
	CustomerName  string `json:"customer_name" binding:"required,max=128"`
	Phone         string `json:"phone" binding:"max=32"`
	GuestCount    int    `json:"guest_count"`
	ReserveAt     string `json:"reserve_at" binding:"required"`
	Channel       string `json:"channel"`
	Remark        string `json:"remark"`
	DepositAmount string `json:"deposit_amount"`
}

type SeatInput struct {
	TableID     *int64 `json:"table_id"`
	CreateOrder bool   `json:"create_order"`
	Remark      string `json:"remark"`
}

type CancelInput struct {
	Reason string `json:"reason"`
	Remark string `json:"remark"`
}

type NoShowInput struct {
	Reason string `json:"reason"`
	Remark string `json:"remark"`
}

type SeatResult struct {
	Reservation *model.Reservation `json:"reservation"`
	Order       *model.Order       `json:"order,omitempty"`
	Table       *model.DiningTable `json:"table,omitempty"`
}

func (s *Service) Create(ctx context.Context, storeID int64, op Operator, in CreateInput) (*model.Reservation, error) {
	reserveAt, err := parseTime(in.ReserveAt)
	if err != nil {
		return nil, err
	}
	channel := in.Channel
	if channel == "" {
		channel = model.ReservationChannelPhone
	}
	if !validChannel(channel) {
		return nil, fmt.Errorf("无效渠道")
	}
	guest := in.GuestCount
	if guest <= 0 {
		guest = 2
	}
	deposit, err := parseDeposit(in.DepositAmount)
	if err != nil {
		return nil, err
	}
	if in.TableID != nil {
		if err := s.ensureTable(ctx, storeID, *in.TableID); err != nil {
			return nil, err
		}
	}
	m := &model.Reservation{
		StoreID:       storeID,
		TableID:       in.TableID,
		CustomerName:  in.CustomerName,
		Phone:         in.Phone,
		GuestCount:    guest,
		ReserveAt:     reserveAt,
		Channel:       channel,
		Status:        model.ReservationStatusPendingConfirm,
		Remark:        in.Remark,
		DepositAmount: deposit,
		CreatedBy:     op.UserID,
		UpdatedBy:     op.UserID,
	}
	if err := s.res.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建预订失败: %w", err)
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, storeID, id int64, op Operator, in UpdateInput) (*model.Reservation, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if m.Status != model.ReservationStatusPendingConfirm && m.Status != model.ReservationStatusConfirmed {
		return nil, fmt.Errorf("当前状态不可修改：%s", m.Status)
	}
	reserveAt, err := parseTime(in.ReserveAt)
	if err != nil {
		return nil, err
	}
	channel := in.Channel
	if channel == "" {
		channel = m.Channel
	}
	if !validChannel(channel) {
		return nil, fmt.Errorf("无效渠道")
	}
	guest := in.GuestCount
	if guest <= 0 {
		guest = m.GuestCount
	}
	deposit := m.DepositAmount
	if in.DepositAmount != "" {
		d, err := parseDeposit(in.DepositAmount)
		if err != nil {
			return nil, err
		}
		deposit = d
	}
	if in.TableID != nil {
		if err := s.ensureTable(ctx, storeID, *in.TableID); err != nil {
			return nil, err
		}
	}
	applyFields := func() {
		m.TableID = in.TableID
		m.CustomerName = in.CustomerName
		m.Phone = in.Phone
		m.GuestCount = guest
		m.ReserveAt = reserveAt
		m.Channel = channel
		m.Remark = in.Remark
		m.DepositAmount = deposit
		m.UpdatedBy = op.UserID
	}
	// 已确认且换桌：释放旧桌预留、检查新桌
	if m.Status == model.ReservationStatusConfirmed {
		oldTableID := m.TableID
		newTableID := in.TableID
		err = s.res.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if oldTableID != nil && (newTableID == nil || *oldTableID != *newTableID) {
				if err := s.releaseTableTx(ctx, tx, storeID, *oldTableID); err != nil {
					return err
				}
			}
			if newTableID != nil && (oldTableID == nil || *oldTableID != *newTableID) {
				if err := s.reserveTableTx(ctx, tx, storeID, *newTableID); err != nil {
					return err
				}
			}
			applyFields()
			return s.res.SaveTx(ctx, tx, m)
		})
		if err != nil {
			return nil, err
		}
		return m, nil
	}

	applyFields()
	if err := s.res.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("更新预订失败: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, storeID, id int64) error {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return err
	}
	if m.Status != model.ReservationStatusPendingConfirm {
		return fmt.Errorf("仅待确认预订可删除")
	}
	return s.res.SoftDelete(ctx, id)
}

func (s *Service) Get(ctx context.Context, storeID, id int64) (*model.Reservation, error) {
	return s.getOwned(ctx, storeID, id)
}

func (s *Service) List(ctx context.Context, storeID int64, tableID int64, status, channel, keyword, from, to string, page, pageSize int) ([]model.Reservation, int64, error) {
	f := resrepo.ListFilter{
		StoreID: storeID, TableID: tableID, Status: status, Channel: channel, Keyword: keyword,
		Page: page, PageSize: pageSize,
	}
	if from != "" {
		t, err := parseTime(from)
		if err != nil {
			return nil, 0, fmt.Errorf("reserve_from 无效")
		}
		f.ReserveFrom = &t
	}
	if to != "" {
		t, err := parseTime(to)
		if err != nil {
			return nil, 0, fmt.Errorf("reserve_to 无效")
		}
		f.ReserveTo = &t
	}
	return s.res.List(ctx, f)
}

func (s *Service) Confirm(ctx context.Context, storeID, id int64, op Operator) (*model.Reservation, error) {
	var out *model.Reservation
	err := s.res.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.res.GetByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("预订不存在")
			}
			return err
		}
		if m.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店预订")
		}
		if m.Status != model.ReservationStatusPendingConfirm {
			return fmt.Errorf("仅待确认可确认，当前：%s", m.Status)
		}
		if m.TableID != nil {
			if err := s.reserveTableTx(ctx, tx, storeID, *m.TableID); err != nil {
				return err
			}
		}
		now := time.Now()
		m.Status = model.ReservationStatusConfirmed
		m.ConfirmedAt = &now
		m.UpdatedBy = op.UserID
		if err := s.res.SaveTx(ctx, tx, m); err != nil {
			return err
		}
		out = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) Seat(ctx context.Context, storeID, id int64, op Operator, in SeatInput) (*SeatResult, error) {
	var result SeatResult
	err := s.res.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.res.GetByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("预订不存在")
			}
			return err
		}
		if m.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店预订")
		}
		if m.Status != model.ReservationStatusConfirmed && m.Status != model.ReservationStatusPendingConfirm {
			return fmt.Errorf("当前状态不可入座：%s", m.Status)
		}
		tableID := m.TableID
		if in.TableID != nil {
			tableID = in.TableID
		}
		if tableID == nil {
			return fmt.Errorf("入座需指定桌台")
		}
		tbl, err := s.tables.GetByIDTx(ctx, tx, *tableID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("桌台不存在")
			}
			return err
		}
		if tbl.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店桌台")
		}
		// 关键：入座与桌台状态一致 → occupied
		if tbl.Status != model.TableStatusIdle && tbl.Status != model.TableStatusReserved {
			return fmt.Errorf("桌台非空闲/预留，无法入座，当前：%s", tbl.Status)
		}
		// 换桌且原桌是预留则释放
		if m.TableID != nil && *m.TableID != *tableID {
			if err := s.releaseTableTx(ctx, tx, storeID, *m.TableID); err != nil {
				return err
			}
		}

		var order *model.Order
		if in.CreateOrder {
			now := time.Now()
			order = &model.Order{
				StoreID:        storeID,
				TableID:        *tableID,
				OrderNo:        s.orders.CreateOrderNo(storeID),
				Status:         model.OrderStatusPendingConfirm,
				GuestCount:     m.GuestCount,
				SubtotalAmount: decimal.Zero,
				FeeAmount:      decimal.Zero,
				DiscountAmount: decimal.Zero,
				PayableAmount:  decimal.Zero,
				Remark:         "预订入座开台",
				CreatedBy:      op.UserID,
				UpdatedBy:      op.UserID,
				OpenedAt:       now,
			}
			if err := tx.WithContext(ctx).Create(order).Error; err != nil {
				return fmt.Errorf("创建空订单失败: %w", err)
			}
			if err := tx.WithContext(ctx).Create(&model.OrderStatusLog{
				StoreID: storeID, OrderID: order.ID, FromStatus: "", ToStatus: model.OrderStatusPendingConfirm,
				OperatorID: op.UserID, OperatorName: op.Username, Remark: "预订入座开台",
			}).Error; err != nil {
				return err
			}
			tbl.CurrentOrderID = &order.ID
			m.OrderID = &order.ID
		}

		tbl.Status = model.TableStatusOccupied
		if err := s.tables.SaveTx(ctx, tx, tbl); err != nil {
			return err
		}

		now := time.Now()
		m.TableID = tableID
		m.Status = model.ReservationStatusSeated
		m.SeatedAt = &now
		if m.ConfirmedAt == nil {
			m.ConfirmedAt = &now
		}
		m.UpdatedBy = op.UserID
		if in.Remark != "" {
			m.Remark = in.Remark
		}
		if err := s.res.SaveTx(ctx, tx, m); err != nil {
			return err
		}
		result.Reservation = m
		result.Order = order
		result.Table = tbl
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) Cancel(ctx context.Context, storeID, id int64, op Operator, in CancelInput) (*model.Reservation, error) {
	return s.closeAs(ctx, storeID, id, op, model.ReservationStatusCancelled, in.Reason, in.Remark)
}

func (s *Service) NoShow(ctx context.Context, storeID, id int64, op Operator, in NoShowInput) (*model.Reservation, error) {
	return s.closeAs(ctx, storeID, id, op, model.ReservationStatusNoShow, in.Reason, in.Remark)
}

func (s *Service) closeAs(ctx context.Context, storeID, id int64, op Operator, toStatus, reason, remark string) (*model.Reservation, error) {
	var out *model.Reservation
	err := s.res.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.res.GetByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("预订不存在")
			}
			return err
		}
		if m.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店预订")
		}
		if m.Status != model.ReservationStatusPendingConfirm && m.Status != model.ReservationStatusConfirmed {
			return fmt.Errorf("当前状态不可操作：%s", m.Status)
		}
		if m.TableID != nil && m.Status == model.ReservationStatusConfirmed {
			if err := s.releaseTableTx(ctx, tx, storeID, *m.TableID); err != nil {
				return err
			}
		}
		now := time.Now()
		m.Status = toStatus
		m.CancelReason = reason
		m.UpdatedBy = op.UserID
		if remark != "" {
			m.Remark = remark
		}
		if toStatus == model.ReservationStatusCancelled {
			m.CancelledAt = &now
		} else {
			m.NoShowAt = &now
		}
		if err := s.res.SaveTx(ctx, tx, m); err != nil {
			return err
		}
		out = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) reserveTableTx(ctx context.Context, tx *gorm.DB, storeID, tableID int64) error {
	tbl, err := s.tables.GetByIDTx(ctx, tx, tableID)
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
		return fmt.Errorf("桌台非空闲，无法预留，当前：%s", tbl.Status)
	}
	tbl.Status = model.TableStatusReserved
	return s.tables.SaveTx(ctx, tx, tbl)
}

func (s *Service) releaseTableTx(ctx context.Context, tx *gorm.DB, storeID, tableID int64) error {
	tbl, err := s.tables.GetByIDTx(ctx, tx, tableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if tbl.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店桌台")
	}
	// 仅释放预留态，避免误清占用
	if tbl.Status == model.TableStatusReserved && tbl.CurrentOrderID == nil {
		tbl.Status = model.TableStatusIdle
		return s.tables.SaveTx(ctx, tx, tbl)
	}
	return nil
}

func (s *Service) ensureTable(ctx context.Context, storeID, tableID int64) error {
	tbl, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("桌台不存在")
		}
		return err
	}
	if tbl.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店桌台")
	}
	return nil
}

func (s *Service) getOwned(ctx context.Context, storeID, id int64) (*model.Reservation, error) {
	m, err := s.res.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("预订不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店预订")
	}
	return m, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("时间格式无效，请用 RFC3339")
}

func parseDeposit(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil || d.IsNegative() {
		return decimal.Zero, fmt.Errorf("订金金额无效")
	}
	return d.Round(2), nil
}

func validChannel(c string) bool {
	switch c {
	case model.ReservationChannelPhone, model.ReservationChannelWalkIn, model.ReservationChannelWechat,
		model.ReservationChannelMeituan, model.ReservationChannelDianping, model.ReservationChannelOnline,
		model.ReservationChannelOther:
		return true
	default:
		return false
	}
}
