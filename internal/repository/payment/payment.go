package payment

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) CreatePaymentNo(storeID, orderID int64) string {
	return fmt.Sprintf("P%d-%d-%s", storeID, orderID, time.Now().Format("150405.000"))
}

func (r *Repository) CreateTx(ctx context.Context, tx *gorm.DB, m *model.Payment) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *Repository) ListByOrder(ctx context.Context, storeID, orderID int64) ([]model.Payment, error) {
	var list []model.Payment
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND order_id = ?", storeID, orderID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *Repository) CountSuccessByOrderTx(ctx context.Context, tx *gorm.DB, orderID int64) (int64, error) {
	var n int64
	err := tx.WithContext(ctx).Model(&model.Payment{}).
		Where("order_id = ? AND status = ?", orderID, model.PaymentStatusSuccess).
		Count(&n).Error
	return n, err
}

// ListSuccessByStoreDay 营业日成功支付（paid_at 落在 [dayStart, dayEnd)）
func (r *Repository) ListSuccessByStoreDay(ctx context.Context, storeID int64, dayStart, dayEnd time.Time, channel string) ([]model.Payment, error) {
	q := r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("store_id = ? AND status = ? AND paid_at >= ? AND paid_at < ?",
			storeID, model.PaymentStatusSuccess, dayStart, dayEnd)
	if channel != "" {
		q = q.Where("channel = ?", channel)
	}
	var list []model.Payment
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}
