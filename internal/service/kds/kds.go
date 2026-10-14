package kds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	kdsrepo "openbows-rms/internal/repository/kds"
	menurepo "openbows-rms/internal/repository/menu"
)

type Service struct {
	kds    *kdsrepo.Repository
	dishes *menurepo.DishRepo
}

func New(kds *kdsrepo.Repository, dishes *menurepo.DishRepo) *Service {
	return &Service{kds: kds, dishes: dishes}
}

type TicketView struct {
	Ticket model.KdsTicket       `json:"ticket"`
	Items  []model.KdsTicketItem `json:"items"`
}

// CreateTicketsForOrder 按档口拆分生成厨单（供订单事务内调用）
func (s *Service) CreateTicketsForOrder(ctx context.Context, tx *gorm.DB, storeID int64, order *model.Order, items []model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	groups := map[int64][]model.OrderItem{}
	for _, it := range items {
		if it.Status == model.OrderItemStatusReturned {
			continue
		}
		dish, err := s.dishes.GetByID(ctx, it.DishID)
		if err != nil {
			return err
		}
		if dish.StationID == nil {
			continue
		}
		sid := *dish.StationID
		groups[sid] = append(groups[sid], it)
		// 回写订单项出品状态为 pending
		if err := tx.WithContext(ctx).Model(&model.OrderItem{}).
			Where("id = ?", it.ID).
			Update("kitchen_status", model.KitchenStatusPending).Error; err != nil {
			return err
		}
	}
	for stationID, group := range groups {
		t := &model.KdsTicket{
			StoreID:   storeID,
			StationID: stationID,
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			TableID:   order.TableID,
			Status:    model.KitchenStatusPending,
		}
		if err := s.kds.CreateTicketTx(ctx, tx, t); err != nil {
			return fmt.Errorf("创建厨单失败: %w", err)
		}
		for _, it := range group {
			ti := &model.KdsTicketItem{
				StoreID:     storeID,
				TicketID:    t.ID,
				OrderItemID: it.ID,
				DishName:    it.DishName,
				SpecName:    it.SpecName,
				Qty:         it.Qty,
				Status:      model.KitchenStatusPending,
			}
			if err := s.kds.CreateTicketItemTx(ctx, tx, ti); err != nil {
				return fmt.Errorf("创建厨单明细失败: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, storeID, stationID int64, status string) ([]TicketView, error) {
	tickets, err := s.kds.ListTickets(ctx, kdsrepo.ListFilter{
		StoreID: storeID, StationID: stationID, Status: status,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(tickets))
	for _, t := range tickets {
		ids = append(ids, t.ID)
	}
	items, err := s.kds.ListItemsByTicketIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	m := map[int64][]model.KdsTicketItem{}
	for _, it := range items {
		m[it.TicketID] = append(m[it.TicketID], it)
	}
	out := make([]TicketView, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, TicketView{Ticket: t, Items: m[t.ID]})
	}
	return out, nil
}

func (s *Service) Start(ctx context.Context, storeID, ticketID int64) (*TicketView, error) {
	err := s.kds.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, err := s.kds.GetByIDTx(ctx, tx, ticketID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("厨单不存在")
			}
			return err
		}
		if t.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店厨单")
		}
		if t.Status != model.KitchenStatusPending {
			return fmt.Errorf("仅待制作厨单可开始，当前：%s", t.Status)
		}
		now := time.Now()
		t.Status = model.KitchenStatusCooking
		t.StartedAt = &now
		if err := s.kds.SaveTx(ctx, tx, t); err != nil {
			return err
		}
		if err := s.kds.UpdateItemsStatusTx(ctx, tx, ticketID, model.KitchenStatusCooking); err != nil {
			return err
		}
		items, err := s.kds.ListItemsTx(ctx, tx, ticketID)
		if err != nil {
			return err
		}
		for _, it := range items {
			if err := tx.Model(&model.OrderItem{}).Where("id = ?", it.OrderItemID).
				Update("kitchen_status", model.KitchenStatusCooking).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.getView(ctx, storeID, ticketID)
}

func (s *Service) Done(ctx context.Context, storeID, ticketID int64) (*TicketView, error) {
	err := s.kds.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, err := s.kds.GetByIDTx(ctx, tx, ticketID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("厨单不存在")
			}
			return err
		}
		if t.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店厨单")
		}
		if t.Status != model.KitchenStatusCooking && t.Status != model.KitchenStatusPending {
			return fmt.Errorf("当前状态不可出餐完成：%s", t.Status)
		}
		now := time.Now()
		t.Status = model.KitchenStatusDone
		t.DoneAt = &now
		if t.StartedAt == nil {
			t.StartedAt = &now
		}
		if err := s.kds.SaveTx(ctx, tx, t); err != nil {
			return err
		}
		if err := s.kds.UpdateItemsStatusTx(ctx, tx, ticketID, model.KitchenStatusDone); err != nil {
			return err
		}
		// 关键：回写订单明细出品状态，供收银侧可见
		items, err := s.kds.ListItemsTx(ctx, tx, ticketID)
		if err != nil {
			return err
		}
		for _, it := range items {
			if err := tx.Model(&model.OrderItem{}).Where("id = ?", it.OrderItemID).
				Update("kitchen_status", model.KitchenStatusDone).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.getView(ctx, storeID, ticketID)
}

func (s *Service) getView(ctx context.Context, storeID, ticketID int64) (*TicketView, error) {
	t, err := s.kds.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t.StoreID != storeID {
		return nil, fmt.Errorf("无权查看其他门店厨单")
	}
	items, err := s.kds.ListItems(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	return &TicketView{Ticket: *t, Items: items}, nil
}
