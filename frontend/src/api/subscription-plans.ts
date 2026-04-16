/**
 * Subscription Plans API endpoints
 * Handles subscription plan browsing, purchasing, and order management
 */

import { apiClient } from './client'

// ==================== Types ====================

export interface SubscriptionPlan {
  group_id: number
  display_name: string
  description: string
  price: number
  discounted_price?: number
  validity_days: number
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  features: string[]
  rate_multiplier: number
  supported_model_scopes: string[]
  is_subscribed: boolean
  current_expiry?: string
}

export interface PurchaseRequest {
  group_id: number
  payment_method: string
}

export interface PurchaseResult {
  order_id: number
  order_no: string
  status: string
  payment_url?: string
  qr_code_url?: string
  form_html?: string
}

export interface SubscriptionOrderItem {
  id: number
  order_no: string
  trade_no?: string
  user_id: number
  group_id: number
  amount: number
  original_price: number
  discount_amount: number
  payment_method: string
  status: string
  created_at: string
  paid_at?: string
  group?: { name: string; description?: string }
}


// ==================== API Functions ====================

/**
 * Get all available subscription plans
 */
export async function getPlans(): Promise<SubscriptionPlan[]> {
  const { data } = await apiClient.get<SubscriptionPlan[]>('/subscription-plans')
  return data
}

/**
 * Purchase a subscription plan
 */
export async function purchasePlan(req: PurchaseRequest): Promise<PurchaseResult> {
  const { data } = await apiClient.post<PurchaseResult>('/subscription-plans/purchase', req)
  return data
}

/**
 * Get user's subscription order history
 */
export async function getOrders(params?: {
  page?: number
  page_size?: number
}): Promise<{ orders: SubscriptionOrderItem[]; total: number; page: number }> {
  const { data } = await apiClient.get<{ orders: SubscriptionOrderItem[]; total: number; page: number }>(
    '/subscription-plans/orders',
    { params }
  )
  return data
}

/**
 * Get order status by order ID (for polling)
 */
export async function getOrderStatus(orderId: number): Promise<{ status: string }> {
  const { data } = await apiClient.get<{ status: string }>(`/subscription-plans/orders/${orderId}/status`)
  return data
}

export const subscriptionPlansAPI = {
  getPlans,
  purchasePlan,
  getOrders,
  getOrderStatus
}

export default subscriptionPlansAPI
