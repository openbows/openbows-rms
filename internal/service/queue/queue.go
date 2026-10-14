package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	orderrepo "openbows-rms/internal/repository/order"
	queuerepo "openbows-rms/internal/repository/queue"
	tablerepo "openbows-rms/internal/repository/table"
)

var locAsiaShanghai = func() *time.Location {
	l, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return l
}()

type Service struct {
	repo   *queuerepo.Repository
	tables *tablerepo.TableRepo
	orders *orderrepo.Repository
}

func New(repo *queuerepo.Repository, tables *tablerepo.TableRepo, orders *orderrepo.Repository) *Service {
	return &Service{repo: repo, tables: tables, orders: orders}
}

type Operator struct {
	UserID   int64
	Username string
}

type TakeInput struct {
	GuestCount  int    `json:"guest_count"`
	Phone       string `json:"phone"`
	Remark      string `json:"remark"`
	BusinessDay string `json:"business_day"`
	OrderID     *int64 `json:"order_id"`
}

type SeatInput struct {
	TableID     *int64 `json:"table_id" binding:"required"`
	CreateOrder bool   `json:"create_order"`
	Remark      string `json:"remark"`
}

type CancelInput struct {
	Remark string `json:"remark"`
}

func (s *Service) Take(ctx context.Context, storeID int64, in TakeInput) (*model.QueueTicket, error) {
	if in.GuestCount <= 0 {
		in.GuestCount = 1
	}
	day, err := parseBusinessDay(in.BusinessDay)
	if err != nil {
		return nil, err
	}
	var out *model.QueueTicket
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		no, err := s.repo.NextTicketNoTx(ctx, tx, storeID, day)
		if err != nil {
			return err
		}
		m := &model.QueueTicket{
			StoreID: storeID, BusinessDay: day, TicketNo: no,
			GuestCount: in.GuestCount, Phone: in.Phone, Status: model.QueueStatusWaiting,
			OrderID: in.OrderID, Remark: in.Remark,
		}
		if err := s.repo.CreateTx(ctx, tx, m); err != nil {
			return fmt.Errorf("取号失败: %w", err)
		}
		out = m
		return nil
	})
	return out, err
}

func (s *Service) List(ctx context.Context, storeID int64, businessDay, status string, page, pageSize int) ([]model.QueueTicket, int64, error) {
	var day *time.Time
	if businessDay != "" {
		t, err := parseBusinessDay(businessDay)
		if err != nil {
			return nil, 0, err
		}
		day = &t
	} else {
		t := todayBusinessDay()
		day = &t
	}
	return s.repo.List(ctx, queuerepo.ListFilter{
		StoreID: storeID, BusinessDay: day, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) Call(ctx context.Context, storeID, id int64) (*model.QueueTicket, error) {
	return s.transition(ctx, storeID, id, func(m *model.QueueTicket, _ *gorm.DB) error {
		if m.Status != model.QueueStatusWaiting {
			return fmt.Errorf("仅等待中可叫号，当前：%s", m.Status)
		}
		now := time.Now()
		m.Status = model.QueueStatusCalled
		m.CalledAt = &now
		return nil
	})
}

func (s *Service) Seat(ctx context.Context, storeID, id int64, op Operator, in SeatInput) (*model.QueueTicket, error) {
	if in.TableID == nil || *in.TableID <= 0 {
		return nil, fmt.Errorf("入座需指定桌台")
	}
	return s.transition(ctx, storeID, id, func(m *model.QueueTicket, tx *gorm.DB) error {
		if m.Status != model.QueueStatusWaiting && m.Status != model.QueueStatusCalled {
			return fmt.Errorf("当前状态不可入座：%s", m.Status)
		}
		tbl, err := s.tables.GetByIDTx(ctx, tx, *in.TableID)
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
			return fmt.Errorf("桌台非空闲，无法入座，当前：%s", tbl.Status)
		}

		if in.CreateOrder {
			now := time.Now()
			order := &model.Order{
				StoreID: storeID, TableID: *in.TableID, OrderNo: s.orders.CreateOrderNo(storeID),
				Status: model.OrderStatusPendingConfirm, GuestCount: m.GuestCount,
				SubtotalAmount: decimal.Zero, FeeAmount: decimal.Zero, DiscountAmount: decimal.Zero,
				PayableAmount: decimal.Zero, Remark: "排队入座开台",
				CreatedBy: op.UserID, UpdatedBy: op.UserID, OpenedAt: now,
			}
			if err := tx.WithContext(ctx).Create(order).Error; err != nil {
				return fmt.Errorf("创建空订单失败: %w", err)
			}
			if err := tx.WithContext(ctx).Create(&model.OrderStatusLog{
				StoreID: storeID, OrderID: order.ID, FromStatus: "", ToStatus: model.OrderStatusPendingConfirm,
				OperatorID: op.UserID, OperatorName: op.Username, Remark: "排队入座开台",
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
		m.Status = model.QueueStatusSeated
		m.SeatedAt = &now
		if m.CalledAt == nil {
			m.CalledAt = &now
		}
		m.TableID = in.TableID
		if in.Remark != "" {
			m.Remark = in.Remark
		}
		return nil
	})
}

func (s *Service) Cancel(ctx context.Context, storeID, id int64, in CancelInput) (*model.QueueTicket, error) {
	return s.transition(ctx, storeID, id, func(m *model.QueueTicket, _ *gorm.DB) error {
		if m.Status != model.QueueStatusWaiting && m.Status != model.QueueStatusCalled {
			return fmt.Errorf("当前状态不可取消：%s", m.Status)
		}
		now := time.Now()
		m.Status = model.QueueStatusCancelled
		m.CancelledAt = &now
		if in.Remark != "" {
			m.Remark = in.Remark
		}
		return nil
	})
}

func (s *Service) transition(ctx context.Context, storeID, id int64, apply func(*model.QueueTicket, *gorm.DB) error) (*model.QueueTicket, error) {
	var out *model.QueueTicket
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.repo.GetByIDTx(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("排队票不存在")
			}
			return err
		}
		if m.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店排队票")
		}
		if err := apply(m, tx); err != nil {
			return err
		}
		if err := s.repo.SaveTx(ctx, tx, m); err != nil {
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

func todayBusinessDay() time.Time {
	now := time.Now().In(locAsiaShanghai)
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, locAsiaShanghai)
}

func parseBusinessDay(s string) (time.Time, error) {
	if s == "" {
		return todayBusinessDay(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, locAsiaShanghai)
	if err != nil {
		return time.Time{}, fmt.Errorf("营业日格式无效，应为 YYYY-MM-DD")
	}
	return t, nil
}
