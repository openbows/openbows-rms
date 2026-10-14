package online

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	memberrepo "openbows-rms/internal/repository/member"
	onlinerepo "openbows-rms/internal/repository/online"
	ordersvc "openbows-rms/internal/service/order"
)

/*
鉴权策略（顾客端扫码 / H5）：
1. 不使用员工 JWT。
2. 下单：请求体携带 store_id + table_id（由桌台二维码解析），服务端校验桌台归属该门店。
3. 查询：必须携带下单返回的 access_token（Query access_token 或 Header X-Order-Access-Token），
   与 online_order_extras 校验通过后才返回进度。
4. 制作 / 结账仍走管理端既有 /apis/v1/orders 接口（员工 JWT）。
*/

type Service struct {
	orders  *ordersvc.Service
	online  *onlinerepo.Repository
	members *memberrepo.Repository
}

func New(orders *ordersvc.Service, online *onlinerepo.Repository, members *memberrepo.Repository) *Service {
	return &Service{orders: orders, online: online, members: members}
}

type CreateInput struct {
	StoreID      int64               `json:"store_id" binding:"required"`
	TableID      int64               `json:"table_id" binding:"required"`
	GuestCount   int                 `json:"guest_count"`
	Remark       string              `json:"remark"`
	Source       string              `json:"source"` // h5 | scan
	ContactName  string              `json:"contact_name"`
	ContactPhone string              `json:"contact_phone"`
	MemberPhone  string              `json:"member_phone"` // 可选：关联会员
	Items        []ordersvc.ItemInput `json:"items" binding:"required,min=1,dive"`
}

type CreateResult struct {
	Order       *model.Order              `json:"order"`
	Items       []ordersvc.ItemView       `json:"items"`
	Fees        []model.OrderFeeItem      `json:"fees"`
	StatusLogs  []model.OrderStatusLog    `json:"status_logs"`
	AccessToken string                    `json:"access_token"`
	Extra       *model.OnlineOrderExtra   `json:"extra"`
}

type ProgressResult struct {
	Order      *model.Order           `json:"order"`
	Items      []ordersvc.ItemView    `json:"items"`
	Fees       []model.OrderFeeItem   `json:"fees"`
	StatusLogs []model.OrderStatusLog `json:"status_logs"`
	Extra      *model.OnlineOrderExtra `json:"extra"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*CreateResult, error) {
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = "scan"
	}
	if source != "h5" && source != "scan" {
		return nil, fmt.Errorf("无效来源，可选：h5/scan")
	}
	channel := model.OrderChannelOnlineScan
	if source == "h5" {
		channel = model.OrderChannelOnlineH5
	}

	var memberID *int64
	if phone := strings.TrimSpace(in.MemberPhone); phone != "" {
		m, err := s.members.GetMemberByPhone(ctx, in.StoreID, phone)
		if err != nil {
			if !memberrepo.IsNotFound(err) {
				return nil, err
			}
			// 未找到会员则忽略，不阻断下单
		} else if m.Status {
			id := m.ID
			memberID = &id
		}
	}

	accessToken := strings.ReplaceAll(uuid.NewString(), "-", "")
	var extra *model.OnlineOrderExtra

	detail, err := s.orders.Create(ctx, in.StoreID, ordersvc.Operator{
		UserID: 0, Username: "online",
	}, ordersvc.CreateInput{
		TableID:    in.TableID,
		GuestCount: in.GuestCount,
		Remark:     in.Remark,
		Channel:    channel,
		MemberID:   memberID,
		Items:      in.Items,
		AfterCreateTx: func(ctx context.Context, tx *gorm.DB, order *model.Order) error {
			extra = &model.OnlineOrderExtra{
				StoreID:      in.StoreID,
				OrderID:      order.ID,
				ContactName:  in.ContactName,
				ContactPhone: in.ContactPhone,
				AccessToken:  accessToken,
				Source:       source,
			}
			return s.online.CreateExtrasTx(ctx, tx, extra)
		},
	})
	if err != nil {
		return nil, err
	}
	// 状态日志备注对顾客端可见即可；管理端仍用同一订单
	return &CreateResult{
		Order: detail.Order, Items: detail.Items, Fees: detail.Fees, StatusLogs: detail.Logs,
		AccessToken: accessToken, Extra: extra,
	}, nil
}

func (s *Service) GetProgress(ctx context.Context, orderID int64, accessToken string) (*ProgressResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("缺少 access_token")
	}
	extra, err := s.online.GetByAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("无效的查询凭证")
		}
		return nil, err
	}
	if extra.OrderID != orderID {
		return nil, fmt.Errorf("订单与凭证不匹配")
	}
	detail, err := s.orders.GetDetail(ctx, extra.StoreID, orderID)
	if err != nil {
		return nil, err
	}
	return &ProgressResult{
		Order: detail.Order, Items: detail.Items, Fees: detail.Fees, StatusLogs: detail.Logs, Extra: extra,
	}, nil
}
