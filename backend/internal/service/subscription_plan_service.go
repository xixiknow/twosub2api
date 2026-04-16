package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service/payment"
)

var (
	ErrSubscriptionPurchaseDisabled = infraerrors.Forbidden("SUBSCRIPTION_PURCHASE_DISABLED", "subscription purchase is currently disabled")
	ErrPlanNotFound                 = infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found or not visible")
	ErrPlanPriceInvalid             = infraerrors.BadRequest("PLAN_PRICE_INVALID", "subscription plan price is not configured or is zero")
	ErrSubscriptionOrderNotFound    = infraerrors.NotFound("SUBSCRIPTION_ORDER_NOT_FOUND", "subscription order not found")
	ErrSubscriptionAmountMismatch   = infraerrors.BadRequest("AMOUNT_MISMATCH", "callback amount does not match order amount")
)

// SubscriptionPlan 套餐展示信息
type SubscriptionPlan struct {
	GroupID         int64    `json:"group_id"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	Price           float64  `json:"price"`
	DiscountedPrice *float64 `json:"discounted_price,omitempty"`
	ValidityDays    int      `json:"validity_days"`
	DailyLimitUSD   *float64 `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUSD  *float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd,omitempty"`
	Features        []string `json:"features"`
	RateMultiplier  float64  `json:"rate_multiplier"`
	SupportedModels []string `json:"supported_model_scopes"`
	IsSubscribed    bool     `json:"is_subscribed"`
	CurrentExpiry   *string  `json:"current_expiry,omitempty"`
	SortOrder       int      `json:"-"` // internal use for sorting
}

// PurchaseRequest 购买请求
type PurchaseRequest struct {
	GroupID       int64  `json:"group_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// PurchaseResult 购买结果
type PurchaseResult struct {
	OrderID    int64  `json:"order_id"`
	OrderNo    string `json:"order_no"`
	Status     string `json:"status"`
	PaymentURL string `json:"payment_url,omitempty"`
	QRCodeURL  string `json:"qr_code_url,omitempty"`
	FormHTML   string `json:"form_html,omitempty"`
}

// PriceCalculation 价格计算结果
type PriceCalculation struct {
	OriginalPrice   float64 `json:"original_price"`
	DiscountedPrice float64 `json:"discounted_price"`
	DiscountAmount  float64 `json:"discount_amount"`
	OnlineAmount    float64 `json:"online_amount"` // 在线支付金额，当前与套餐价 1:1
}

// SubscriptionPlanService 订阅套餐服务
type SubscriptionPlanService struct {
	db                  *sql.DB
	entClient           *dbent.Client
	groupRepo           GroupRepository
	subscriptionService *SubscriptionService
	paymentService      *PaymentService
	vipService          *VIPService
	settingRepo         SettingRepository
	userRepo            UserRepository
	billingCacheService *BillingCacheService
	subOrderRepo        SubscriptionOrderRepository
}

// NewSubscriptionPlanService 创建订阅套餐服务
func NewSubscriptionPlanService(
	db *sql.DB,
	entClient *dbent.Client,
	groupRepo GroupRepository,
	subscriptionService *SubscriptionService,
	paymentService *PaymentService,
	vipService *VIPService,
	settingRepo SettingRepository,
	userRepo UserRepository,
	billingCacheService *BillingCacheService,
	subOrderRepo SubscriptionOrderRepository,
) *SubscriptionPlanService {
	return &SubscriptionPlanService{
		db:                  db,
		entClient:           entClient,
		groupRepo:           groupRepo,
		subscriptionService: subscriptionService,
		paymentService:      paymentService,
		vipService:          vipService,
		settingRepo:         settingRepo,
		userRepo:            userRepo,
		billingCacheService: billingCacheService,
		subOrderRepo:        subOrderRepo,
	}
}

// ListPlans 查询所有可见套餐，附带用户订阅状态和 VIP 折扣价
func (s *SubscriptionPlanService) ListPlans(ctx context.Context, userID int64) ([]SubscriptionPlan, error) {
	// 查询所有活跃分组
	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}

	// 获取用户 VIP 信息
	var baseMultiplier float64 = 1
	if s.vipService != nil {
		vipSummary, _, _, vipErr := s.vipService.ResolveUserVIP(ctx, userID)
		if vipErr == nil && vipSummary != nil && vipSummary.Enabled {
			baseMultiplier = vipSummary.BaseMultiplier
		}
	}

	var plans []SubscriptionPlan
	for i := range groups {
		g := &groups[i]

		// 仅包含 subscription_visible=true 且 subscription_type 为订阅类型的分组
		if !g.SubscriptionVisible || !g.IsSubscriptionType() {
			continue
		}

		plan := SubscriptionPlan{
			GroupID:         g.ID,
			DisplayName:     g.SubscriptionDisplayName,
			Description:     g.Description,
			Price:           0,
			ValidityDays:    g.DefaultValidityDays,
			DailyLimitUSD:   g.DailyLimitUSD,
			WeeklyLimitUSD:  g.WeeklyLimitUSD,
			MonthlyLimitUSD: g.MonthlyLimitUSD,
			Features:        g.SubscriptionFeatures,
			RateMultiplier:  g.RateMultiplier,
			SupportedModels: g.SupportedModelScopes,
			SortOrder:       g.SortOrder,
		}

		if plan.DisplayName == "" {
			plan.DisplayName = g.Name
		}
		if plan.Features == nil {
			plan.Features = []string{}
		}
		if plan.SupportedModels == nil {
			plan.SupportedModels = []string{}
		}

		// 设置价格
		if g.SubscriptionPrice != nil {
			plan.Price = *g.SubscriptionPrice
		}

		// 计算 VIP 折扣价
		if baseMultiplier < 1 && plan.Price > 0 {
			discounted := roundTo2(plan.Price * baseMultiplier)
			plan.DiscountedPrice = &discounted
		}

		// 检查用户是否有活跃订阅
		if userID > 0 {
			sub, subErr := s.subscriptionService.GetActiveSubscription(ctx, userID, g.ID)
			if subErr == nil && sub != nil && sub.IsActive() {
				plan.IsSubscribed = true
				expiry := sub.ExpiresAt.Format(time.RFC3339)
				plan.CurrentExpiry = &expiry
			}
		}

		plans = append(plans, plan)
	}

	// 按 sort_order 升序排列
	sort.SliceStable(plans, func(i, j int) bool {
		return plans[i].SortOrder < plans[j].SortOrder
	})

	if plans == nil {
		plans = []SubscriptionPlan{}
	}

	return plans, nil
}

// CalculatePrice 计算最终价格（含 VIP 折扣）
func (s *SubscriptionPlanService) CalculatePrice(ctx context.Context, userID int64, groupID int64) (*PriceCalculation, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, ErrPlanNotFound
	}

	if group.SubscriptionPrice == nil || *group.SubscriptionPrice <= 0 {
		return nil, ErrPlanPriceInvalid
	}

	originalPrice := *group.SubscriptionPrice

	// 获取 VIP 折扣倍率
	var baseMultiplier float64 = 1
	if s.vipService != nil {
		vipSummary, _, _, vipErr := s.vipService.ResolveUserVIP(ctx, userID)
		if vipErr == nil && vipSummary != nil && vipSummary.Enabled {
			baseMultiplier = vipSummary.BaseMultiplier
		}
	}

	// 最终价格 = round(price × multiplier, 2)
	discountedPrice := roundTo2(originalPrice * baseMultiplier)
	discountAmount := roundTo2(originalPrice - discountedPrice)

	return &PriceCalculation{
		OriginalPrice:   originalPrice,
		DiscountedPrice: discountedPrice,
		DiscountAmount:  discountAmount,
		OnlineAmount:    discountedPrice,
	}, nil
}

// Purchase 购买入口：验证 → 计算价格 → 余额扣款或创建在线支付订单
func (s *SubscriptionPlanService) Purchase(ctx context.Context, userID int64, req *PurchaseRequest) (*PurchaseResult, error) {
	// 1. 检查功能开关
	if !s.isPurchaseEnabled(ctx) {
		return nil, ErrSubscriptionPurchaseDisabled
	}

	// 2. 验证套餐
	group, err := s.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, ErrPlanNotFound
	}
	if !group.SubscriptionVisible || !group.IsSubscriptionType() {
		return nil, ErrPlanNotFound
	}
	if group.SubscriptionPrice == nil || *group.SubscriptionPrice <= 0 {
		return nil, ErrPlanPriceInvalid
	}

	// 3. 验证支付方式
	if !s.isPaymentMethodAvailable(ctx, req.PaymentMethod) {
		return nil, ErrPaymentMethodNotAvailable
	}

	// 4. 计算价格
	priceCalc, err := s.CalculatePrice(ctx, userID, req.GroupID)
	if err != nil {
		return nil, err
	}

	// 5. 根据支付方式分流
	if req.PaymentMethod == "balance" {
		return s.purchaseWithBalance(ctx, userID, group, priceCalc)
	}
	return s.purchaseWithOnlinePayment(ctx, userID, group, priceCalc, req.PaymentMethod)
}

// purchaseWithBalance 余额支付路径
func (s *SubscriptionPlanService) purchaseWithBalance(ctx context.Context, userID int64, group *Group, priceCalc *PriceCalculation) (*PurchaseResult, error) {
	orderNo := generateSubOrderNo()
	finalPrice := priceCalc.DiscountedPrice
	validityDays := group.DefaultValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// 1. 检查并扣减余额（原子操作）
	result, err := tx.Client().ExecContext(txCtx,
		`UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		finalPrice, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("deduct balance: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrInsufficientBalance
	}

	// 2. 同一事务内开通/续期订阅
	sub, _, assignErr := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
		UserID:       userID,
		GroupID:      group.ID,
		ValidityDays: validityDays,
		Notes:        fmt.Sprintf("自助购买 订单号:%s 金额:%.2f", orderNo, finalPrice),
	})
	if assignErr != nil {
		return nil, fmt.Errorf("assign subscription: %w", assignErr)
	}

	// 3. 创建已支付订单
	now := time.Now()
	order := &SubscriptionOrder{
		OrderNo:        orderNo,
		UserID:         userID,
		GroupID:        group.ID,
		Amount:         finalPrice,
		OriginalPrice:  priceCalc.OriginalPrice,
		DiscountAmount: priceCalc.DiscountAmount,
		PaymentMethod:  "balance",
		Status:         string(payment.OrderStatusPaid),
		CreatedAt:      now,
		PaidAt:         &now,
		ActivatedAt:    &now,
	}
	if sub != nil {
		order.SubscriptionID = &sub.ID
	}
	if err := s.subOrderRepo.Create(txCtx, order); err != nil {
		return nil, fmt.Errorf("create subscription order: %w", err)
	}

	// 4. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// 5. 清除用户余额缓存
	if s.billingCacheService != nil {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if cacheErr := s.billingCacheService.InvalidateUserBalance(cacheCtx, userID); cacheErr != nil {
			log.Printf("subscription purchase: invalidate balance cache failed for user %d: %v", userID, cacheErr)
		}
	}

	return &PurchaseResult{
		OrderID: order.ID,
		OrderNo: orderNo,
		Status:  string(payment.OrderStatusPaid),
	}, nil
}

// purchaseWithOnlinePayment 在线支付路径
func (s *SubscriptionPlanService) purchaseWithOnlinePayment(ctx context.Context, userID int64, group *Group, priceCalc *PriceCalculation, method string) (*PurchaseResult, error) {
	orderNo := generateSubOrderNo()
	payAmount := priceCalc.OnlineAmount

	// 1. 创建订阅订单（status=pending）
	order := &SubscriptionOrder{
		OrderNo:        orderNo,
		UserID:         userID,
		GroupID:        group.ID,
		Amount:         payAmount,
		OriginalPrice:  priceCalc.OriginalPrice,
		DiscountAmount: priceCalc.DiscountAmount,
		PaymentMethod:  method,
		Status:         string(payment.OrderStatusPending),
		CreatedAt:      time.Now(),
	}
	if err := s.subOrderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create subscription order: %w", err)
	}

	// 2. 获取支付网关
	gateway, err := s.paymentService.GetGatewayWithCallbacks(
		ctx,
		payment.PaymentMethod(method),
		subscriptionNotifyPath(payment.PaymentMethod(method)),
		"/api/v1/subscription-plans/return",
	)
	if err != nil {
		// 标记订单失败
		_ = s.subOrderRepo.UpdateStatus(ctx, order.ID, string(payment.OrderStatusFailed))
		return nil, fmt.Errorf("get payment gateway: %w", err)
	}

	// 3. 创建支付
	displayName := group.SubscriptionDisplayName
	if displayName == "" {
		displayName = group.Name
	}
	payResult, err := gateway.CreatePay(&payment.CreatePayRequest{
		OrderNo: orderNo,
		Amount:  payAmount,
		Subject: fmt.Sprintf("订阅套餐 - %s", displayName),
	})
	if err != nil {
		_ = s.subOrderRepo.UpdateStatus(ctx, order.ID, string(payment.OrderStatusFailed))
		return nil, fmt.Errorf("create payment: %w", err)
	}

	return &PurchaseResult{
		OrderID:    order.ID,
		OrderNo:    orderNo,
		Status:     string(payment.OrderStatusPending),
		PaymentURL: payResult.PaymentURL,
		QRCodeURL:  payResult.QRCodeURL,
		FormHTML:   payResult.FormHTML,
	}, nil
}

// isPurchaseEnabled 检查自助购买功能是否启用
func (s *SubscriptionPlanService) isPurchaseEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeySubscriptionPurchaseEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

// isPaymentMethodAvailable 检查支付方式是否可用
func (s *SubscriptionPlanService) isPaymentMethodAvailable(ctx context.Context, method string) bool {
	if method == "balance" {
		return true
	}

	// 在线支付方式需要检查支付配置
	config, err := s.paymentService.GetPaymentConfig(ctx)
	if err != nil || !config.Enabled {
		return false
	}

	switch payment.PaymentMethod(method) {
	case payment.MethodAlipay:
		return config.Methods.Alipay
	case payment.MethodAlipayF2F:
		return config.Methods.AlipayF2F
	case payment.MethodWechat:
		return config.Methods.Wechat
	case payment.MethodEpay:
		return config.Methods.Epay
	case payment.MethodEpayAlipay:
		return config.Methods.EpayAlipay
	case payment.MethodEpayWechat:
		return config.Methods.EpayWechat
	default:
		return false
	}
}

// generateSubOrderNo 生成订阅订单号（SUB前缀）
func generateSubOrderNo() string {
	now := time.Now()
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("SUB%s%s", now.Format("20060102150405"), hex.EncodeToString(b))
}

// roundTo2 四舍五入到小数点后两位
func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}

// HandleSubscriptionNotify 处理订阅支付回调通知
func (s *SubscriptionPlanService) HandleSubscriptionNotify(ctx context.Context, channel string, params map[string]string) error {
	gateway, err := s.paymentService.GetGatewayWithCallbacks(
		ctx,
		payment.PaymentMethod(channel),
		subscriptionNotifyPath(payment.PaymentMethod(channel)),
		"/api/v1/subscription-plans/return",
	)
	if err != nil {
		return fmt.Errorf("get gateway for subscription notify: %w", err)
	}

	result, err := gateway.VerifyNotify(params)
	if err != nil {
		return fmt.Errorf("verify subscription notify: %w", err)
	}

	if !result.Success {
		log.Printf("subscription notify: order %s not success", result.OrderNo)
		return nil
	}

	// 使用 Ent 事务处理
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// 查询订单（悲观锁）
	order, err := s.subOrderRepo.GetByOrderNoForUpdate(txCtx, result.OrderNo)
	if err != nil {
		return fmt.Errorf("subscription order not found: %s, err: %w", result.OrderNo, err)
	}

	// 幂等处理：已支付订单直接返回成功
	if order.Status != string(payment.OrderStatusPending) {
		log.Printf("subscription notify: order %s already processed (status=%s)", result.OrderNo, order.Status)
		return nil
	}

	// 金额校验
	if result.Amount > 0 && fmt.Sprintf("%.2f", result.Amount) != fmt.Sprintf("%.2f", order.Amount) {
		return fmt.Errorf("subscription amount mismatch: expected %.2f, got %.2f", order.Amount, result.Amount)
	}

	// 获取分组信息以获取 DefaultValidityDays
	group, err := s.groupRepo.GetByID(txCtx, order.GroupID)
	if err != nil {
		return fmt.Errorf("get group for subscription: %w", err)
	}

	validityDays := group.DefaultValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}

	// 调用 AssignOrExtendSubscription
	sub, _, assignErr := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      order.GroupID,
		ValidityDays: validityDays,
		Notes:        fmt.Sprintf("自助购买(在线支付) 订单号:%s 金额:%.2f", order.OrderNo, order.Amount),
	})
	if assignErr != nil {
		return fmt.Errorf("assign subscription: %w", assignErr)
	}

	// 更新订单为已支付状态
	notifyJSON, _ := json.Marshal(params)
	subscriptionID := int64(0)
	if sub != nil {
		subscriptionID = sub.ID
	}
	if err := s.subOrderRepo.UpdatePaid(txCtx, order.ID, result.TradeNo, string(notifyJSON), subscriptionID); err != nil {
		return fmt.Errorf("update subscription order paid: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("subscription payment success: order=%s user=%d group=%d amount=%.2f", order.OrderNo, order.UserID, order.GroupID, order.Amount)
	return nil
}

// GetOrderStatus 查询订阅订单状态（支持主动查询上游支付网关）
func (s *SubscriptionPlanService) GetOrderStatus(ctx context.Context, orderID int64, userID int64) (string, error) {
	order, err := s.subOrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", ErrSubscriptionOrderNotFound
	}

	// 验证订单属于该用户
	if order.UserID != userID {
		return "", ErrSubscriptionOrderNotFound
	}

	// 如果订单还是 pending，尝试主动查询支付网关
	if order.Status == string(payment.OrderStatusPending) {
		method := payment.PaymentMethod(order.PaymentMethod)
		// 只支持支付宝和当面付的主动查询
		switch method {
		case payment.MethodAlipay, payment.MethodAlipayF2F:
			if updatedStatus, err := s.activeQuerySubscriptionPayment(ctx, order); err != nil {
				log.Printf("subscription active query error: orderID=%d err=%v", orderID, err)
			} else if updatedStatus != "" {
				return updatedStatus, nil
			}
		}
	}

	return order.Status, nil
}

// activeQuerySubscriptionPayment 主动查询订阅支付网关交易状态
func (s *SubscriptionPlanService) activeQuerySubscriptionPayment(ctx context.Context, order *SubscriptionOrder) (string, error) {
	gateway, err := s.paymentService.GetGatewayWithCallbacks(
		ctx,
		payment.PaymentMethod(order.PaymentMethod),
		subscriptionNotifyPath(payment.PaymentMethod(order.PaymentMethod)),
		"/api/v1/subscription-plans/return",
	)
	if err != nil {
		return "", fmt.Errorf("get gateway: %w", err)
	}

	querier, ok := gateway.(payment.TradeQuerier)
	if !ok {
		return "", nil
	}

	result, err := querier.QueryTrade(order.OrderNo)
	if err != nil {
		return "", fmt.Errorf("query trade: %w", err)
	}

	if !result.Success {
		return "", nil
	}

	// 交易成功，在 Ent 事务中处理
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// 再次检查订单状态（悲观锁），防止并发
	currentOrder, err := s.subOrderRepo.GetByOrderNoForUpdate(txCtx, order.OrderNo)
	if err != nil {
		return "", fmt.Errorf("recheck order: %w", err)
	}
	if currentOrder.Status != string(payment.OrderStatusPending) {
		return currentOrder.Status, nil
	}

	// 金额校验
	if result.Amount > 0 && fmt.Sprintf("%.2f", result.Amount) != fmt.Sprintf("%.2f", order.Amount) {
		return "", fmt.Errorf("amount mismatch: expected %.2f, got %.2f", order.Amount, result.Amount)
	}

	// 获取分组信息
	group, err := s.groupRepo.GetByID(txCtx, order.GroupID)
	if err != nil {
		return "", fmt.Errorf("get group: %w", err)
	}

	validityDays := group.DefaultValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}

	// 激活订阅
	sub, _, assignErr := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      order.GroupID,
		ValidityDays: validityDays,
		Notes:        fmt.Sprintf("自助购买(主动查询) 订单号:%s 金额:%.2f", order.OrderNo, order.Amount),
	})
	if assignErr != nil {
		return "", fmt.Errorf("assign subscription: %w", assignErr)
	}

	// 更新订单状态
	notifyJSON, _ := json.Marshal(map[string]string{
		"source":       "active_query",
		"trade_no":     result.TradeNo,
		"trade_status": result.TradeStatus,
	})
	subscriptionID := int64(0)
	if sub != nil {
		subscriptionID = sub.ID
	}
	if err := s.subOrderRepo.UpdatePaid(txCtx, order.ID, result.TradeNo, string(notifyJSON), subscriptionID); err != nil {
		return "", fmt.Errorf("update order: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("subscription payment success (active query): order=%s user=%d amount=%.2f", order.OrderNo, order.UserID, order.Amount)
	return string(payment.OrderStatusPaid), nil
}

// GetOrderPaymentMethod 根据订单号获取支付方式
func (s *SubscriptionPlanService) GetOrderPaymentMethod(ctx context.Context, orderNo string) (string, error) {
	order, err := s.subOrderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return "", err
	}
	return order.PaymentMethod, nil
}

// ListUserOrders 查询用户的订阅订单列表
func (s *SubscriptionPlanService) ListUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]SubscriptionOrder, int64, error) {
	return s.subOrderRepo.ListByUserID(ctx, userID, page, pageSize)
}

// ExpireOrders 过期超时的 pending 订阅订单（>30分钟）
func (s *SubscriptionPlanService) ExpireOrders(ctx context.Context) (int, error) {
	return s.subOrderRepo.ExpirePendingOrders(ctx, 30*time.Minute)
}

func subscriptionNotifyPath(method payment.PaymentMethod) string {
	switch method {
	case payment.MethodAlipay, payment.MethodAlipayF2F:
		return "/api/v1/subscription-plans/notify/alipay"
	case payment.MethodWechat:
		return "/api/v1/subscription-plans/notify/wechat"
	case payment.MethodEpay, payment.MethodEpayAlipay, payment.MethodEpayWechat:
		return "/api/v1/subscription-plans/notify/epay"
	default:
		return ""
	}
}
