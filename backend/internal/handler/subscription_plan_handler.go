package handler

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanHandler 订阅套餐处理器
type SubscriptionPlanHandler struct {
	planService *service.SubscriptionPlanService
}

// NewSubscriptionPlanHandler 创建订阅套餐处理器
func NewSubscriptionPlanHandler(planService *service.SubscriptionPlanService) *SubscriptionPlanHandler {
	return &SubscriptionPlanHandler{planService: planService}
}

// ListPlans 获取可购买的订阅套餐列表
// GET /api/v1/subscription-plans
func (h *SubscriptionPlanHandler) ListPlans(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	plans, err := h.planService.ListPlans(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, plans)
}

// Purchase 购买订阅套餐
// POST /api/v1/subscription-plans/purchase
func (h *SubscriptionPlanHandler) Purchase(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req service.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.planService.Purchase(c.Request.Context(), subject.UserID, &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

// ListOrders 获取用户订阅订单列表
// GET /api/v1/subscription-plans/orders
func (h *SubscriptionPlanHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.planService.ListUserOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
	})
}

// GetOrderStatus 查询订阅订单状态
// GET /api/v1/subscription-plans/orders/:id/status
func (h *SubscriptionPlanHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	status, err := h.planService.GetOrderStatus(c.Request.Context(), orderID, subject.UserID)
	if err != nil {
		log.Printf("get subscription order status error: orderID=%d userID=%d err=%v", orderID, subject.UserID, err)
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"status": status})
}

// NotifyAlipay 支付宝异步回调（订阅套餐）
// POST /api/v1/subscription-plans/notify/alipay
func (h *SubscriptionPlanHandler) NotifyAlipay(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 通过 out_trade_no 查询订单的实际支付方式
	channel := "alipay"
	if outTradeNo := params["out_trade_no"]; outTradeNo != "" {
		if method, err := h.planService.GetOrderPaymentMethod(c.Request.Context(), outTradeNo); err == nil && method != "" {
			channel = method
		}
	}
	log.Printf("subscription alipay notify received: trade_no=%s out_trade_no=%s trade_status=%s channel=%s",
		params["trade_no"], params["out_trade_no"], params["trade_status"], channel)

	if err := h.planService.HandleSubscriptionNotify(c.Request.Context(), channel, params); err != nil {
		log.Printf("subscription alipay notify handle error: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// NotifyWechat 微信支付异步回调（订阅套餐）
// POST /api/v1/subscription-plans/notify/wechat
func (h *SubscriptionPlanHandler) NotifyWechat(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(http.StatusOK, gin.H{"return_code": "FAIL", "return_msg": "read body error"})
		return
	}

	params, err := parseSubscriptionWechatXML(body)
	if err != nil {
		c.XML(http.StatusOK, gin.H{"return_code": "FAIL", "return_msg": "parse xml error"})
		return
	}

	if err := h.planService.HandleSubscriptionNotify(c.Request.Context(), "wechat", params); err != nil {
		c.XML(http.StatusOK, gin.H{"return_code": "FAIL", "return_msg": err.Error()})
		return
	}

	c.XML(http.StatusOK, gin.H{"return_code": "SUCCESS", "return_msg": "OK"})
}

// NotifyEpay 易支付异步回调（订阅套餐）
// GET /api/v1/subscription-plans/notify/epay
func (h *SubscriptionPlanHandler) NotifyEpay(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if err := h.planService.HandleSubscriptionNotify(c.Request.Context(), "epay", params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// PaymentReturn 支付完成跳转页面（订阅套餐）
// GET /api/v1/subscription-plans/return
func (h *SubscriptionPlanHandler) PaymentReturn(c *gin.Context) {
	c.Redirect(http.StatusFound, "/purchase")
}

// parseSubscriptionWechatXML 解析微信 XML 通知（订阅套餐专用）
func parseSubscriptionWechatXML(data []byte) (map[string]string, error) {
	type xmlEntry struct {
		XMLName xml.Name
		Value   string `xml:",chardata"`
	}
	type xmlMap struct {
		XMLName xml.Name   `xml:"xml"`
		Entries []xmlEntry `xml:",any"`
	}

	var m xmlMap
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	result := make(map[string]string, len(m.Entries))
	for _, e := range m.Entries {
		result[e.XMLName.Local] = e.Value
	}
	return result, nil
}
