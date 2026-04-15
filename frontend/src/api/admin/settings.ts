/**
 * Admin Settings API endpoints
 * Handles system settings management for administrators
 */

import { apiClient } from '../client'
import type { CustomMenuItem } from '@/types'

export interface DefaultSubscriptionSetting {
  group_id: number
  validity_days: number
}

/**
 * System settings interface
 */
export interface SystemSettings {
  // Registration settings
  registration_enabled: boolean
  email_verify_enabled: boolean
  registration_email_suffix_whitelist: string[]
  promo_code_enabled: boolean
  password_reset_enabled: boolean
  invitation_code_enabled: boolean
  totp_enabled: boolean // TOTP 双因素认证
  totp_encryption_key_configured: boolean // TOTP 加密密钥是否已配置
  // Default settings
  default_balance: number
  default_concurrency: number
  default_subscriptions: DefaultSubscriptionSetting[]
  // OEM settings
  site_name: string
  site_logo: string
  site_subtitle: string
  api_base_url: string
  contact_info: string
  doc_url: string
  home_content: string
  hide_ccs_import_button: boolean
  model_square_enabled: boolean
  availability_check_enabled: boolean
  purchase_subscription_enabled: boolean
  purchase_subscription_url: string
  custom_menu_items: CustomMenuItem[]
  vip_enabled: boolean
  vip_rules: string
  // SMTP settings
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password_configured: boolean
  smtp_from_email: string
  smtp_from_name: string
  smtp_use_tls: boolean
  // Cloudflare Turnstile settings
  turnstile_enabled: boolean
  turnstile_site_key: string
  turnstile_secret_key_configured: boolean

  // LinuxDo Connect OAuth settings
  linuxdo_connect_enabled: boolean
  linuxdo_connect_client_id: string
  linuxdo_connect_client_secret_configured: boolean
  linuxdo_connect_redirect_url: string

  // Model fallback configuration
  enable_model_fallback: boolean
  fallback_model_anthropic: string
  fallback_model_openai: string
  fallback_model_gemini: string
  fallback_model_antigravity: string

  // Identity patch configuration (Claude -> Gemini)
  enable_identity_patch: boolean
  identity_patch_prompt: string

  // Ops Monitoring (vNext)
  ops_monitoring_enabled: boolean
  ops_realtime_monitoring_enabled: boolean
  ops_query_mode_default: 'auto' | 'raw' | 'preagg' | string
  ops_metrics_interval_seconds: number

  // Claude Code version check
  min_claude_code_version: string

  // 分组隔离
  allow_ungrouped_key_scheduling: boolean

  // 登录 IP 提醒
  login_ip_alert_enabled: boolean

  // Payment settings
  payment_enabled: boolean
  payment_currency: string
  payment_exchange_rate: number
  payment_preset_amounts: string
  payment_min_amount: number
  payment_max_amount: number
  payment_alipay_enabled: boolean
  payment_alipay_app_id: string
  payment_alipay_private_key_configured: boolean
  payment_alipay_public_key_configured: boolean
  payment_alipay_f2f_enabled: boolean
  payment_wechat_enabled: boolean
  payment_wechat_app_id: string
  payment_wechat_mch_id: string
  payment_wechat_api_key_configured: boolean
  payment_epay_enabled: boolean
  payment_epay_type: string
  payment_epay_api_url: string
  payment_epay_pid: string
  payment_epay_key_configured: boolean

  // Referral settings
  referral_enabled: boolean
  referral_commission_rate: number
}

export interface UpdateSettingsRequest {
  registration_enabled?: boolean
  email_verify_enabled?: boolean
  registration_email_suffix_whitelist?: string[]
  promo_code_enabled?: boolean
  password_reset_enabled?: boolean
  invitation_code_enabled?: boolean
  totp_enabled?: boolean // TOTP 双因素认证
  default_balance?: number
  default_concurrency?: number
  default_subscriptions?: DefaultSubscriptionSetting[]
  site_name?: string
  site_logo?: string
  site_subtitle?: string
  api_base_url?: string
  contact_info?: string
  doc_url?: string
  home_content?: string
  hide_ccs_import_button?: boolean
  model_square_enabled?: boolean
  availability_check_enabled?: boolean
  purchase_subscription_enabled?: boolean
  purchase_subscription_url?: string
  custom_menu_items?: CustomMenuItem[]
  vip_enabled?: boolean
  vip_rules?: string
  smtp_host?: string
  smtp_port?: number
  smtp_username?: string
  smtp_password?: string
  smtp_from_email?: string
  smtp_from_name?: string
  smtp_use_tls?: boolean
  turnstile_enabled?: boolean
  turnstile_site_key?: string
  turnstile_secret_key?: string
  linuxdo_connect_enabled?: boolean
  linuxdo_connect_client_id?: string
  linuxdo_connect_client_secret?: string
  linuxdo_connect_redirect_url?: string
  enable_model_fallback?: boolean
  fallback_model_anthropic?: string
  fallback_model_openai?: string
  fallback_model_gemini?: string
  fallback_model_antigravity?: string
  enable_identity_patch?: boolean
  identity_patch_prompt?: string
  ops_monitoring_enabled?: boolean
  ops_realtime_monitoring_enabled?: boolean
  ops_query_mode_default?: 'auto' | 'raw' | 'preagg' | string
  ops_metrics_interval_seconds?: number
  min_claude_code_version?: string
  allow_ungrouped_key_scheduling?: boolean
  login_ip_alert_enabled?: boolean

  // Payment settings
  payment_enabled?: boolean
  payment_currency?: string
  payment_exchange_rate?: number
  payment_preset_amounts?: string
  payment_min_amount?: number
  payment_max_amount?: number
  payment_alipay_enabled?: boolean
  payment_alipay_app_id?: string
  payment_alipay_private_key?: string
  payment_alipay_public_key?: string
  payment_alipay_f2f_enabled?: boolean
  payment_wechat_enabled?: boolean
  payment_wechat_app_id?: string
  payment_wechat_mch_id?: string
  payment_wechat_api_key?: string
  payment_epay_enabled?: boolean
  payment_epay_type?: string
  payment_epay_api_url?: string
  payment_epay_pid?: string
  payment_epay_key?: string

  // Referral settings
  referral_enabled?: boolean
  referral_commission_rate?: number
}

/**
 * Get all system settings
 * @returns System settings
 */
export async function getSettings(): Promise<SystemSettings> {
  const { data } = await apiClient.get<SystemSettings>('/admin/settings')
  return data
}

/**
 * Update system settings
 * @param settings - Partial settings to update
 * @returns Updated settings
 */
export async function updateSettings(settings: UpdateSettingsRequest): Promise<SystemSettings> {
  const { data } = await apiClient.put<SystemSettings>('/admin/settings', settings)
  return data
}

/**
 * Test SMTP connection request
 */
export interface TestSmtpRequest {
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_use_tls: boolean
}

/**
 * Test SMTP connection with provided config
 * @param config - SMTP configuration to test
 * @returns Test result message
 */
export async function testSmtpConnection(config: TestSmtpRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/settings/test-smtp', config)
  return data
}

/**
 * Send test email request
 */
export interface SendTestEmailRequest {
  email: string
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_from_email: string
  smtp_from_name: string
  smtp_use_tls: boolean
}

/**
 * Send test email with provided SMTP config
 * @param request - Email address and SMTP config
 * @returns Test result message
 */
export async function sendTestEmail(request: SendTestEmailRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    '/admin/settings/send-test-email',
    request
  )
  return data
}

/**
 * Admin API Key status response
 */
export interface AdminApiKeyStatus {
  exists: boolean
  masked_key: string
}

/**
 * Get admin API key status
 * @returns Status indicating if key exists and masked version
 */
export async function getAdminApiKey(): Promise<AdminApiKeyStatus> {
  const { data } = await apiClient.get<AdminApiKeyStatus>('/admin/settings/admin-api-key')
  return data
}

/**
 * Regenerate admin API key
 * @returns The new full API key (only shown once)
 */
export async function regenerateAdminApiKey(): Promise<{ key: string }> {
  const { data } = await apiClient.post<{ key: string }>('/admin/settings/admin-api-key/regenerate')
  return data
}

/**
 * Delete admin API key
 * @returns Success message
 */
export async function deleteAdminApiKey(): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>('/admin/settings/admin-api-key')
  return data
}

/**
 * Stream timeout settings interface
 */
export interface StreamTimeoutSettings {
  enabled: boolean
  action: 'temp_unsched' | 'error' | 'none'
  temp_unsched_minutes: number
  threshold_count: number
  threshold_window_minutes: number
}

/**
 * Get stream timeout settings
 * @returns Stream timeout settings
 */
export async function getStreamTimeoutSettings(): Promise<StreamTimeoutSettings> {
  const { data } = await apiClient.get<StreamTimeoutSettings>('/admin/settings/stream-timeout')
  return data
}

/**
 * Update stream timeout settings
 * @param settings - Stream timeout settings to update
 * @returns Updated settings
 */
export async function updateStreamTimeoutSettings(
  settings: StreamTimeoutSettings
): Promise<StreamTimeoutSettings> {
  const { data } = await apiClient.put<StreamTimeoutSettings>(
    '/admin/settings/stream-timeout',
    settings
  )
  return data
}

// ==================== Rectifier Settings ====================

/**
 * Rectifier settings interface
 */
export interface RectifierSettings {
  enabled: boolean
  thinking_signature_enabled: boolean
  thinking_budget_enabled: boolean
}

/**
 * Get rectifier settings
 * @returns Rectifier settings
 */
export async function getRectifierSettings(): Promise<RectifierSettings> {
  const { data } = await apiClient.get<RectifierSettings>('/admin/settings/rectifier')
  return data
}

/**
 * Update rectifier settings
 * @param settings - Rectifier settings to update
 * @returns Updated settings
 */
export async function updateRectifierSettings(
  settings: RectifierSettings
): Promise<RectifierSettings> {
  const { data } = await apiClient.put<RectifierSettings>(
    '/admin/settings/rectifier',
    settings
  )
  return data
}

// ==================== Beta Policy Settings ====================

/**
 * Beta policy rule interface
 */
export interface BetaPolicyRule {
  beta_token: string
  action: 'pass' | 'filter' | 'block'
  scope: 'all' | 'oauth' | 'apikey' | 'bedrock'
  error_message?: string
}

/**
 * Beta policy settings interface
 */
export interface BetaPolicySettings {
  rules: BetaPolicyRule[]
}

/**
 * Get beta policy settings
 * @returns Beta policy settings
 */
export async function getBetaPolicySettings(): Promise<BetaPolicySettings> {
  const { data } = await apiClient.get<BetaPolicySettings>('/admin/settings/beta-policy')
  return data
}

/**
 * Update beta policy settings
 * @param settings - Beta policy settings to update
 * @returns Updated settings
 */
export async function updateBetaPolicySettings(
  settings: BetaPolicySettings
): Promise<BetaPolicySettings> {
  const { data } = await apiClient.put<BetaPolicySettings>(
    '/admin/settings/beta-policy',
    settings
  )
  return data
}

// ==================== Payment Settings ====================

export interface PaymentSettingsResponse {
  payment_enabled: boolean
  payment_currency: string
  payment_exchange_rate: number
  payment_preset_amounts: string
  payment_min_amount: number
  payment_max_amount: number
  payment_alipay_enabled: boolean
  payment_alipay_app_id: string
  payment_alipay_private_key_configured: boolean
  payment_alipay_public_key_configured: boolean
  payment_alipay_f2f_enabled: boolean
  payment_wechat_enabled: boolean
  payment_wechat_app_id: string
  payment_wechat_mch_id: string
  payment_wechat_api_key_configured: boolean
  payment_epay_enabled: boolean
  payment_epay_api_url: string
  payment_epay_pid: string
  payment_epay_key_configured: boolean
  payment_epay_type: string
  referral_enabled: boolean
  referral_commission_rate: number
}

export interface UpdatePaymentSettingsRequest {
  payment_enabled: boolean
  payment_currency: string
  payment_exchange_rate: number
  payment_preset_amounts: string
  payment_min_amount: number
  payment_max_amount: number
  payment_alipay_enabled: boolean
  payment_alipay_app_id: string
  payment_alipay_private_key?: string
  payment_alipay_public_key?: string
  payment_alipay_f2f_enabled: boolean
  payment_wechat_enabled: boolean
  payment_wechat_app_id: string
  payment_wechat_mch_id: string
  payment_wechat_api_key?: string
  payment_epay_enabled: boolean
  payment_epay_type: string
  payment_epay_api_url: string
  payment_epay_pid: string
  payment_epay_key?: string
  referral_enabled: boolean
  referral_commission_rate: number
}

export async function getPaymentSettings(): Promise<PaymentSettingsResponse> {
  const { data } = await apiClient.get<PaymentSettingsResponse>('/admin/settings/payment')
  return data
}

export async function updatePaymentSettings(settings: UpdatePaymentSettingsRequest): Promise<PaymentSettingsResponse> {
  const { data } = await apiClient.put<PaymentSettingsResponse>('/admin/settings/payment', settings)
  return data
}

export const settingsAPI = {
  getSettings,
  updateSettings,
  testSmtpConnection,
  sendTestEmail,
  getAdminApiKey,
  regenerateAdminApiKey,
  deleteAdminApiKey,
  getStreamTimeoutSettings,
  updateStreamTimeoutSettings,
  getRectifierSettings,
  updateRectifierSettings,
  getBetaPolicySettings,
  updateBetaPolicySettings,
  getPaymentSettings,
  updatePaymentSettings
}

export default settingsAPI
