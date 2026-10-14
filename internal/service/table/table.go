package table

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
	tablerepo "openbows-rms/internal/repository/table"
)

type TableService struct {
	tables *tablerepo.TableRepo
	areas  *tablerepo.AreaRepo
}

func NewTableService(tables *tablerepo.TableRepo, areas *tablerepo.AreaRepo) *TableService {
	return &TableService{tables: tables, areas: areas}
}

type TableInput struct {
	AreaID   int64  `json:"area_id" binding:"required"`
	Code     string `json:"code" binding:"required,max=64"`
	Name     string `json:"name" binding:"max=128"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
	SortNo   int    `json:"sort_no"`
}

type ChangeStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func (s *TableService) Create(ctx context.Context, storeID int64, in TableInput) (*model.DiningTable, error) {
	if err := s.ensureArea(ctx, storeID, in.AreaID); err != nil {
		return nil, err
	}
	status := model.TableStatusIdle
	if in.Status != "" {
		if !validStatus(in.Status) {
			return nil, fmt.Errorf("无效桌台状态，可选：idle/occupied/reserved")
		}
		status = in.Status
	}
	cap := in.Capacity
	if cap <= 0 {
		cap = 4
	}
	m := &model.DiningTable{
		StoreID:  storeID,
		AreaID:   in.AreaID,
		Code:     in.Code,
		Name:     in.Name,
		Capacity: cap,
		Status:   status,
		SortNo:   in.SortNo,
	}
	if err := s.tables.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建桌台失败: %w", err)
	}
	return m, nil
}

func (s *TableService) Update(ctx context.Context, storeID, id int64, in TableInput) (*model.DiningTable, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureArea(ctx, storeID, in.AreaID); err != nil {
		return nil, err
	}
	if in.Status != "" {
		if !validStatus(in.Status) {
			return nil, fmt.Errorf("无效桌台状态，可选：idle/occupied/reserved")
		}
		m.Status = in.Status
	}
	cap := in.Capacity
	if cap <= 0 {
		cap = m.Capacity
	}
	m.AreaID = in.AreaID
	m.Code = in.Code
	m.Name = in.Name
	m.Capacity = cap
	m.SortNo = in.SortNo
	if err := s.tables.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新桌台失败: %w", err)
	}
	return m, nil
}

func (s *TableService) ChangeStatus(ctx context.Context, storeID, id int64, in ChangeStatusInput) (*model.DiningTable, error) {
	if !validStatus(in.Status) {
		return nil, fmt.Errorf("无效桌台状态，可选：idle/occupied/reserved")
	}
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	// 本阶段仅改状态，不联动订单/预订
	m.Status = in.Status
	if err := s.tables.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("变更桌台状态失败: %w", err)
	}
	return m, nil
}

func (s *TableService) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.tables.SoftDelete(ctx, id)
}

func (s *TableService) Get(ctx context.Context, storeID, id int64) (*model.DiningTable, error) {
	return s.getOwned(ctx, storeID, id)
}

func (s *TableService) List(ctx context.Context, storeID, areaID int64, status, keyword string, page, pageSize int) ([]model.DiningTable, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if status != "" && !validStatus(status) {
		return nil, 0, fmt.Errorf("无效桌台状态，可选：idle/occupied/reserved")
	}
	return s.tables.List(ctx, tablerepo.ListFilter{
		StoreID:  storeID,
		AreaID:   areaID,
		Status:   status,
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *TableService) ensureArea(ctx context.Context, storeID, areaID int64) error {
	ok, err := s.areas.ExistsInStore(ctx, storeID, areaID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("区域不存在或不属于本门店")
	}
	return nil
}

func (s *TableService) getOwned(ctx context.Context, storeID, id int64) (*model.DiningTable, error) {
	m, err := s.tables.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("桌台不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店桌台")
	}
	return m, nil
}

func validStatus(s string) bool {
	switch s {
	case model.TableStatusIdle, model.TableStatusOccupied, model.TableStatusReserved:
		return true
	default:
		return false
	}
}
