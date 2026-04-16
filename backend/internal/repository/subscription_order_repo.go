package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionorder"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// Ensure subscriptionOrderRepository implements service.SubscriptionOrderRepository.
var _ service.SubscriptionOrderRepository = (*subscriptionOrderRepository)(nil)

type subscriptionOrderRepository struct {
	client *dbent.Client
}

func NewSubscriptionOrderRepository(client *dbent.Client) service.SubscriptionOrderRepository {
	return &subscriptionOrderRepository{client: client}
}

func (r *subscriptionOrderRepository) Create(ctx context.Context, order *service.SubscriptionOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.SubscriptionOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetGroupID(order.GroupID).
		SetAmount(order.Amount).
		SetOriginalPrice(order.OriginalPrice).
		SetDiscountAmount(order.DiscountAmount).
		SetPaymentMethod(order.PaymentMethod).
		SetStatus(order.Status)

	if !order.CreatedAt.IsZero() {
		builder.SetCreatedAt(order.CreatedAt)
	}

	if order.TradeNo != nil {
		builder.SetTradeNo(*order.TradeNo)
	}
	if order.NotifyData != nil {
		builder.SetNotifyData(*order.NotifyData)
	}
	if order.SubscriptionID != nil {
		builder.SetSubscriptionID(*order.SubscriptionID)
	}
	if order.PaidAt != nil {
		builder.SetPaidAt(*order.PaidAt)
	}
	if order.ExpiredAt != nil {
		builder.SetExpiredAt(*order.ExpiredAt)
	}
	if order.ActivatedAt != nil {
		builder.SetActivatedAt(*order.ActivatedAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}

	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	return nil
}

func (r *subscriptionOrderRepository) GetByID(ctx context.Context, id int64) (*service.SubscriptionOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.IDEQ(id)).
		WithGroup().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, nil, nil)
	}
	return subscriptionOrderEntityToService(m), nil
}

func (r *subscriptionOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.SubscriptionOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.OrderNoEQ(orderNo)).
		WithGroup().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, nil, nil)
	}
	return subscriptionOrderEntityToService(m), nil
}

func (r *subscriptionOrderRepository) GetByOrderNoForUpdate(ctx context.Context, orderNo string) (*service.SubscriptionOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.OrderNoEQ(orderNo)).
		WithGroup().
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, nil, nil)
	}
	return subscriptionOrderEntityToService(m), nil
}

func (r *subscriptionOrderRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	client := clientFromContext(ctx, r.client)
	return client.SubscriptionOrder.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}

func (r *subscriptionOrderRepository) UpdatePaid(ctx context.Context, id int64, tradeNo string, notifyData string, subscriptionID int64) error {
	client := clientFromContext(ctx, r.client)
	now := time.Now()
	builder := client.SubscriptionOrder.UpdateOneID(id).
		SetStatus("paid").
		SetPaidAt(now).
		SetActivatedAt(now)
	if tradeNo != "" {
		builder.SetTradeNo(tradeNo)
	}
	if notifyData != "" {
		builder.SetNotifyData(notifyData)
	}
	if subscriptionID > 0 {
		builder.SetSubscriptionID(subscriptionID)
	}
	return builder.Exec(ctx)
}

func (r *subscriptionOrderRepository) UpdateExpired(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	now := time.Now()
	return client.SubscriptionOrder.UpdateOneID(id).
		SetStatus("expired").
		SetExpiredAt(now).
		Exec(ctx)
}

func (r *subscriptionOrderRepository) ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]service.SubscriptionOrder, int64, error) {
	client := clientFromContext(ctx, r.client)

	query := client.SubscriptionOrder.Query().
		Where(subscriptionorder.UserIDEQ(userID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	items, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.UserIDEQ(userID)).
		WithGroup().
		Order(dbent.Desc(subscriptionorder.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]service.SubscriptionOrder, len(items))
	for i, m := range items {
		result[i] = *subscriptionOrderEntityToService(m)
	}
	return result, int64(total), nil
}

func (r *subscriptionOrderRepository) ExpirePendingOrders(ctx context.Context, olderThan time.Duration) (int, error) {
	client := clientFromContext(ctx, r.client)
	cutoff := time.Now().Add(-olderThan)
	n, err := client.SubscriptionOrder.Update().
		Where(
			subscriptionorder.StatusEQ("pending"),
			subscriptionorder.CreatedAtLT(cutoff),
		).
		SetStatus("expired").
		SetExpiredAt(time.Now()).
		Save(ctx)
	return n, err
}

func subscriptionOrderEntityToService(m *dbent.SubscriptionOrder) *service.SubscriptionOrder {
	if m == nil {
		return nil
	}
	order := &service.SubscriptionOrder{
		ID:             m.ID,
		OrderNo:        m.OrderNo,
		TradeNo:        m.TradeNo,
		UserID:         m.UserID,
		GroupID:        m.GroupID,
		Amount:         m.Amount,
		OriginalPrice:  m.OriginalPrice,
		DiscountAmount: m.DiscountAmount,
		PaymentMethod:  m.PaymentMethod,
		Status:         m.Status,
		NotifyData:     m.NotifyData,
		CreatedAt:      m.CreatedAt,
		PaidAt:         m.PaidAt,
		ExpiredAt:      m.ExpiredAt,
		ActivatedAt:    m.ActivatedAt,
		SubscriptionID: m.SubscriptionID,
	}

	if g, err := m.Edges.GroupOrErr(); err == nil {
		order.Group = groupEntityToService(g)
	}

	return order
}
