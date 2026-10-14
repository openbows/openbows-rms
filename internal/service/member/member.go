package member

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"openbows-rms/internal/model"
	memberrepo "openbows-rms/internal/repository/member"
)

type Service struct {
	repo *memberrepo.Repository
}

func New(repo *memberrepo.Repository) *Service {
	return &Service{repo: repo}
}

type Operator struct {
	UserID int64
}

type CreateInput struct {
	Phone    string `json:"phone" binding:"required,max=32"`
	Name     string `json:"name" binding:"max=128"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday"` // YYYY-MM-DD
	Level    string `json:"level"`
	Status   *bool  `json:"status"`
	Remark   string `json:"remark"`
}

type UpdateInput struct {
	Phone    string `json:"phone" binding:"required,max=32"`
	Name     string `json:"name" binding:"max=128"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday"`
	Level    string `json:"level"`
	Status   *bool  `json:"status"`
	Remark   string `json:"remark"`
}

type AmountInput struct {
	Amount  string `json:"amount" binding:"required"`
	BizNo   string `json:"biz_no"`
	OrderID *int64 `json:"order_id"`
	Remark  string `json:"remark"`
}

type PointsInput struct {
	Points  int    `json:"points" binding:"required,min=1"`
	BizNo   string `json:"biz_no"`
	OrderID *int64 `json:"order_id"`
	Remark  string `json:"remark"`
}

type Detail struct {
	Member *model.Member       `json:"member"`
	Wallet *model.MemberWallet `json:"wallet"`
}

func (s *Service) Create(ctx context.Context, storeID int64, op Operator, in CreateInput) (*Detail, error) {
	gender := in.Gender
	if gender == "" {
		gender = model.MemberGenderUnknown
	}
	if !validGender(gender) {
		return nil, fmt.Errorf("无效性别")
	}
	level := in.Level
	if level == "" {
		level = "normal"
	}
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	birthday, err := parseBirthday(in.Birthday)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetMemberByPhone(ctx, storeID, in.Phone); err == nil {
		return nil, fmt.Errorf("手机号已存在")
	} else if !memberrepo.IsNotFound(err) {
		return nil, err
	}

	var detail Detail
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := &model.Member{
			StoreID: storeID, Phone: in.Phone, Name: in.Name, Gender: gender,
			Birthday: birthday, Level: level, Status: status, Remark: in.Remark,
			CreatedBy: op.UserID, UpdatedBy: op.UserID,
		}
		if err := s.repo.CreateMemberTx(ctx, tx, m); err != nil {
			return fmt.Errorf("创建会员失败: %w", err)
		}
		w := &model.MemberWallet{
			StoreID: storeID, MemberID: m.ID,
			Balance: decimal.Zero, Points: 0,
		}
		if err := s.repo.CreateWalletTx(ctx, tx, w); err != nil {
			return fmt.Errorf("创建钱包失败: %w", err)
		}
		detail.Member = m
		detail.Wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *Service) Update(ctx context.Context, storeID, id int64, op Operator, in UpdateInput) (*model.Member, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	gender := in.Gender
	if gender == "" {
		gender = m.Gender
	}
	if !validGender(gender) {
		return nil, fmt.Errorf("无效性别")
	}
	if in.Phone != m.Phone {
		if other, err := s.repo.GetMemberByPhone(ctx, storeID, in.Phone); err == nil && other.ID != id {
			return nil, fmt.Errorf("手机号已存在")
		} else if err != nil && !memberrepo.IsNotFound(err) {
			return nil, err
		}
	}
	birthday, err := parseBirthday(in.Birthday)
	if err != nil {
		return nil, err
	}
	m.Phone = in.Phone
	m.Name = in.Name
	m.Gender = gender
	if in.Birthday != "" {
		m.Birthday = birthday
	}
	if in.Level != "" {
		m.Level = in.Level
	}
	if in.Status != nil {
		m.Status = *in.Status
	}
	m.Remark = in.Remark
	m.UpdatedBy = op.UserID
	if err := s.repo.SaveMember(ctx, m); err != nil {
		return nil, fmt.Errorf("更新会员失败: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwned(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("member_id = ?", id).Delete(&model.MemberWallet{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Member{}, id).Error
	})
}

func (s *Service) Get(ctx context.Context, storeID, id int64) (*Detail, error) {
	m, err := s.getOwned(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	w, err := s.repo.GetWalletByMemberID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Detail{Member: m, Wallet: w}, nil
}

func (s *Service) List(ctx context.Context, storeID int64, keyword string, status *bool, page, pageSize int) ([]model.Member, int64, error) {
	return s.repo.ListMembers(ctx, memberrepo.ListFilter{
		StoreID: storeID, Keyword: keyword, Status: status, Page: page, PageSize: pageSize,
	})
}

func (s *Service) Recharge(ctx context.Context, storeID, memberID int64, op Operator, in AmountInput) (*Detail, error) {
	return s.changeBalance(ctx, storeID, memberID, op, in, model.WalletChangeRecharge)
}

func (s *Service) Consume(ctx context.Context, storeID, memberID int64, op Operator, in AmountInput) (*Detail, error) {
	return s.changeBalance(ctx, storeID, memberID, op, in, model.WalletChangeConsume)
}

// ConsumeTx 在调用方事务内扣减会员余额并写流水（支付复用）
func (s *Service) ConsumeTx(ctx context.Context, tx *gorm.DB, storeID, memberID int64, op Operator, in AmountInput) error {
	m, err := s.getOwned(ctx, storeID, memberID)
	if err != nil {
		return err
	}
	if !m.Status {
		return fmt.Errorf("会员已停用")
	}
	amount, err := parseAmount(in.Amount)
	if err != nil {
		return err
	}
	w, err := s.repo.GetWalletByMemberIDTx(ctx, tx, memberID)
	if err != nil {
		return err
	}
	if w.StoreID != storeID {
		return fmt.Errorf("无权操作其他门店钱包")
	}
	before := w.Balance
	after := before.Sub(amount)
	if after.IsNegative() {
		return fmt.Errorf("余额不足")
	}
	w.Balance = after.Round(2)
	if err := s.repo.SaveWalletTx(ctx, tx, w); err != nil {
		return err
	}
	log := &model.MemberWalletLog{
		StoreID: storeID, MemberID: memberID, WalletID: w.ID,
		ChangeType: model.WalletChangeConsume, Amount: amount.Round(2),
		BalanceBefore: before.Round(2), BalanceAfter: w.Balance,
		BizNo: in.BizNo, OrderID: in.OrderID, Remark: in.Remark, OperatorID: op.UserID,
	}
	if err := s.repo.CreateWalletLogTx(ctx, tx, log); err != nil {
		return fmt.Errorf("写余额流水失败: %w", err)
	}
	return nil
}

func (s *Service) EarnPoints(ctx context.Context, storeID, memberID int64, op Operator, in PointsInput) (*Detail, error) {
	return s.changePoints(ctx, storeID, memberID, op, in, model.PointChangeEarn)
}

func (s *Service) SpendPoints(ctx context.Context, storeID, memberID int64, op Operator, in PointsInput) (*Detail, error) {
	return s.changePoints(ctx, storeID, memberID, op, in, model.PointChangeSpend)
}

func (s *Service) ListWalletLogs(ctx context.Context, storeID, memberID int64, page, pageSize int) ([]model.MemberWalletLog, int64, error) {
	if _, err := s.getOwned(ctx, storeID, memberID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListWalletLogs(ctx, storeID, memberID, page, pageSize)
}

func (s *Service) ListPointLogs(ctx context.Context, storeID, memberID int64, page, pageSize int) ([]model.MemberPointLog, int64, error) {
	if _, err := s.getOwned(ctx, storeID, memberID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListPointLogs(ctx, storeID, memberID, page, pageSize)
}

// 关键：改余额与写流水必须同事务
func (s *Service) changeBalance(ctx context.Context, storeID, memberID int64, op Operator, in AmountInput, changeType string) (*Detail, error) {
	m, err := s.getOwned(ctx, storeID, memberID)
	if err != nil {
		return nil, err
	}
	if !m.Status {
		return nil, fmt.Errorf("会员已停用")
	}
	amount, err := parseAmount(in.Amount)
	if err != nil {
		return nil, err
	}
	var wallet *model.MemberWallet
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := s.repo.GetWalletByMemberIDTx(ctx, tx, memberID)
		if err != nil {
			return err
		}
		if w.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店钱包")
		}
		before := w.Balance
		var after decimal.Decimal
		switch changeType {
		case model.WalletChangeRecharge:
			after = before.Add(amount)
		case model.WalletChangeConsume:
			after = before.Sub(amount)
			if after.IsNegative() {
				return fmt.Errorf("余额不足")
			}
		default:
			return fmt.Errorf("无效余额变动类型")
		}
		w.Balance = after.Round(2)
		if err := s.repo.SaveWalletTx(ctx, tx, w); err != nil {
			return err
		}
		log := &model.MemberWalletLog{
			StoreID: storeID, MemberID: memberID, WalletID: w.ID,
			ChangeType: changeType, Amount: amount.Round(2),
			BalanceBefore: before.Round(2), BalanceAfter: w.Balance,
			BizNo: in.BizNo, OrderID: in.OrderID, Remark: in.Remark, OperatorID: op.UserID,
		}
		if err := s.repo.CreateWalletLogTx(ctx, tx, log); err != nil {
			return fmt.Errorf("写余额流水失败: %w", err)
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Detail{Member: m, Wallet: wallet}, nil
}

// 关键：改积分与写流水必须同事务
func (s *Service) changePoints(ctx context.Context, storeID, memberID int64, op Operator, in PointsInput, changeType string) (*Detail, error) {
	m, err := s.getOwned(ctx, storeID, memberID)
	if err != nil {
		return nil, err
	}
	if !m.Status {
		return nil, fmt.Errorf("会员已停用")
	}
	if in.Points <= 0 {
		return nil, fmt.Errorf("积分必须大于0")
	}
	var wallet *model.MemberWallet
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := s.repo.GetWalletByMemberIDTx(ctx, tx, memberID)
		if err != nil {
			return err
		}
		if w.StoreID != storeID {
			return fmt.Errorf("无权操作其他门店钱包")
		}
		before := w.Points
		after := before
		switch changeType {
		case model.PointChangeEarn:
			after = before + in.Points
		case model.PointChangeSpend:
			after = before - in.Points
			if after < 0 {
				return fmt.Errorf("积分不足")
			}
		default:
			return fmt.Errorf("无效积分变动类型")
		}
		w.Points = after
		if err := s.repo.SaveWalletTx(ctx, tx, w); err != nil {
			return err
		}
		log := &model.MemberPointLog{
			StoreID: storeID, MemberID: memberID, WalletID: w.ID,
			ChangeType: changeType, Points: in.Points,
			PointsBefore: before, PointsAfter: after,
			BizNo: in.BizNo, OrderID: in.OrderID, Remark: in.Remark, OperatorID: op.UserID,
		}
		if err := s.repo.CreatePointLogTx(ctx, tx, log); err != nil {
			return fmt.Errorf("写积分流水失败: %w", err)
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Detail{Member: m, Wallet: wallet}, nil
}

func (s *Service) getOwned(ctx context.Context, storeID, id int64) (*model.Member, error) {
	m, err := s.repo.GetMemberByID(ctx, id)
	if err != nil {
		if memberrepo.IsNotFound(err) {
			return nil, fmt.Errorf("会员不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店会员")
	}
	return m, nil
}

func parseAmount(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil || !d.IsPositive() {
		return decimal.Zero, fmt.Errorf("金额无效，须大于0")
	}
	return d.Round(2), nil
}

func parseBirthday(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, fmt.Errorf("生日格式无效，应为 YYYY-MM-DD")
	}
	return &t, nil
}

func validGender(g string) bool {
	switch g {
	case model.MemberGenderUnknown, model.MemberGenderMale, model.MemberGenderFemale:
		return true
	default:
		return false
	}
}
