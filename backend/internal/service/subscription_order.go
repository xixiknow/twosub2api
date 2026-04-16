package service

import (
	"context"
	"time"
)

// SubscriptionOrder 订阅购买订单
type SubscriptionOrder struct {
	ID             int64
	OrderNo        string
	TradeNo        *string
	UserID         int64
	GroupID        int64
	Amount         float64
	OriginalPrice  float64
	DiscountAmount float64
	PaymentMethod  string
	Status         string
	NotifyData     *string
	CreatedAt      time.Time
	PaidAt         *time.Time
	ExpiredAt      *time.Time
	ActivatedAt    *time.Time
	SubscriptionID *int64

	// Group 关联的分组信息（eager-loaded）
	Group *Group
}

// SubscriptionOrderRepository 订阅订单仓储接口
type SubscriptionOrderRepository interface {
	Create(ctx context.Context, order *SubscriptionOrder) error
	GetByID(ctx context.Context, id int64) (*SubscriptionOrder, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*SubscriptionOrder, error)
	GetByOrderNoForUpdate(ctx context.Context, orderNo string) (*SubscriptionOrder, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdatePaid(ctx context.Context, id int64, tradeNo string, notifyData string, subscriptionID int64) error
	UpdateExpired(ctx context.Context, id int64) error
	ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]SubscriptionOrder, int64, error)
	ExpirePendingOrders(ctx context.Context, olderThan time.Duration) (int, error)
}
