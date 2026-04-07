package dto

import (
	"encoding/json"
	"strings"
)

// CustomMenuItem represents a user-configured custom menu entry.
type CustomMenuItem struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	IconSVG    string `json:"icon_svg"`
	URL        string `json:"url"`
	Visibility string `json:"visibility"` // "user" or "admin"
	SortOrder  int    `json:"sort_order"`
}

// SystemSettings represents the admin settings API response payload.
type SystemSettings struct {
	RegistrationEnabled              bool     `json:"registration_enabled"`
	EmailVerifyEnabled               bool     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool     `json:"password_reset_enabled"`
	InvitationCodeEnabled            bool     `json:"invitation_code_enabled"`
	TotpEnabled                      bool     `json:"totp_enabled"`                   // TOTP 双因素认证
	TotpEncryptionKeyConfigured      bool     `json:"totp_encryption_key_configured"` // TOTP 加密密钥是否已配置

	SMTPHost               string `json:"smtp_host"`
	SMTPPort               int    `json:"smtp_port"`
	SMTPUsername           string `json:"smtp_username"`
	SMTPPasswordConfigured bool   `json:"smtp_password_configured"`
	SMTPFrom               string `json:"smtp_from_email"`
	SMTPFromName           string `json:"smtp_from_name"`
	SMTPUseTLS             bool   `json:"smtp_use_tls"`

	TurnstileEnabled             bool   `json:"turnstile_enabled"`
	TurnstileSiteKey             string `json:"turnstile_site_key"`
	TurnstileSecretKeyConfigured bool   `json:"turnstile_secret_key_configured"`

	LinuxDoConnectEnabled                bool   `json:"linuxdo_connect_enabled"`
	LinuxDoConnectClientID               string `json:"linuxdo_connect_client_id"`
	LinuxDoConnectClientSecretConfigured bool   `json:"linuxdo_connect_client_secret_configured"`
	LinuxDoConnectRedirectURL            string `json:"linuxdo_connect_redirect_url"`

	SiteName                    string           `json:"site_name"`
	SiteLogo                    string           `json:"site_logo"`
	SiteSubtitle                string           `json:"site_subtitle"`
	APIBaseURL                  string           `json:"api_base_url"`
	ContactInfo                 string           `json:"contact_info"`
	DocURL                      string           `json:"doc_url"`
	HomeContent                 string           `json:"home_content"`
	HideCcsImportButton         bool             `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled bool             `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL     string           `json:"purchase_subscription_url"`
	CustomMenuItems             []CustomMenuItem `json:"custom_menu_items"`

	DefaultConcurrency   int                          `json:"default_concurrency"`
	DefaultBalance       float64                      `json:"default_balance"`
	DefaultSubscriptions []DefaultSubscriptionSetting `json:"default_subscriptions"`

	// Model fallback configuration
	EnableModelFallback      bool   `json:"enable_model_fallback"`
	FallbackModelAnthropic   string `json:"fallback_model_anthropic"`
	FallbackModelOpenAI      string `json:"fallback_model_openai"`
	FallbackModelGemini      string `json:"fallback_model_gemini"`
	FallbackModelAntigravity string `json:"fallback_model_antigravity"`

	// Identity patch configuration (Claude -> Gemini)
	EnableIdentityPatch bool   `json:"enable_identity_patch"`
	IdentityPatchPrompt string `json:"identity_patch_prompt"`

	// Ops monitoring (vNext)
	OpsMonitoringEnabled         bool   `json:"ops_monitoring_enabled"`
	OpsRealtimeMonitoringEnabled bool   `json:"ops_realtime_monitoring_enabled"`
	OpsQueryModeDefault          string `json:"ops_query_mode_default"`
	OpsMetricsIntervalSeconds    int    `json:"ops_metrics_interval_seconds"`

	MinClaudeCodeVersion string `json:"min_claude_code_version"`

	// 分组隔离
	AllowUngroupedKeyScheduling bool `json:"allow_ungrouped_key_scheduling"`

	// 登录 IP 提醒
	LoginIPAlertEnabled bool `json:"login_ip_alert_enabled"`
}

type DefaultSubscriptionSetting struct {
	GroupID      int64 `json:"group_id"`
	ValidityDays int   `json:"validity_days"`
}

type PublicSettings struct {
	RegistrationEnabled              bool             `json:"registration_enabled"`
	EmailVerifyEnabled               bool             `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string         `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool             `json:"promo_code_enabled"`
	PasswordResetEnabled             bool             `json:"password_reset_enabled"`
	InvitationCodeEnabled            bool             `json:"invitation_code_enabled"`
	TotpEnabled                      bool             `json:"totp_enabled"` // TOTP 双因素认证
	TurnstileEnabled                 bool             `json:"turnstile_enabled"`
	TurnstileSiteKey                 string           `json:"turnstile_site_key"`
	SiteName                         string           `json:"site_name"`
	SiteLogo                         string           `json:"site_logo"`
	SiteSubtitle                     string           `json:"site_subtitle"`
	APIBaseURL                       string           `json:"api_base_url"`
	ContactInfo                      string           `json:"contact_info"`
	DocURL                           string           `json:"doc_url"`
	HomeContent                      string           `json:"home_content"`
	HideCcsImportButton              bool             `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled      bool             `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL          string           `json:"purchase_subscription_url"`
	CustomMenuItems                  []CustomMenuItem `json:"custom_menu_items"`
	LinuxDoOAuthEnabled              bool             `json:"linuxdo_oauth_enabled"`
	ReferralEnabled                  bool             `json:"referral_enabled"`
	LoginIPAlertEnabled              bool             `json:"login_ip_alert_enabled"`
	Version                          string           `json:"version"`
}

// StreamTimeoutSettings 流超时处理配置 DTO
type StreamTimeoutSettings struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}

// RectifierSettings 请求整流器配置 DTO
type RectifierSettings struct {
	Enabled                  bool `json:"enabled"`
	ThinkingSignatureEnabled bool `json:"thinking_signature_enabled"`
	ThinkingBudgetEnabled    bool `json:"thinking_budget_enabled"`
}

// BetaPolicyRule Beta 策略规则 DTO
type BetaPolicyRule struct {
	BetaToken    string `json:"beta_token"`
	Action       string `json:"action"`
	Scope        string `json:"scope"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// BetaPolicySettings Beta 策略配置 DTO
type BetaPolicySettings struct {
	Rules []BetaPolicyRule `json:"rules"`
}

// ParseCustomMenuItems parses a JSON string into a slice of CustomMenuItem.
// Returns empty slice on empty/invalid input.
func ParseCustomMenuItems(raw string) []CustomMenuItem {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []CustomMenuItem{}
	}
	var items []CustomMenuItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []CustomMenuItem{}
	}
	return items
}

// ParseUserVisibleMenuItems parses custom menu items and filters out admin-only entries.
func ParseUserVisibleMenuItems(raw string) []CustomMenuItem {
	items := ParseCustomMenuItems(raw)
	filtered := make([]CustomMenuItem, 0, len(items))
	for _, item := range items {
		if item.Visibility != "admin" {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// PaymentSettings 支付设置 DTO（响应用）
type PaymentSettings struct {
	PaymentEnabled       bool    `json:"payment_enabled"`
	PaymentCurrency      string  `json:"payment_currency"`
	PaymentExchangeRate  float64 `json:"payment_exchange_rate"`
	PaymentPresetAmounts string  `json:"payment_preset_amounts"`
	PaymentMinAmount     float64 `json:"payment_min_amount"`
	PaymentMaxAmount     float64 `json:"payment_max_amount"`

	AlipayEnabled              bool   `json:"payment_alipay_enabled"`
	AlipayAppID                string `json:"payment_alipay_app_id"`
	AlipayPrivateKeyConfigured bool   `json:"payment_alipay_private_key_configured"`
	AlipayPublicKeyConfigured  bool   `json:"payment_alipay_public_key_configured"`
	AlipayF2FEnabled           bool   `json:"payment_alipay_f2f_enabled"`

	WechatEnabled          bool   `json:"payment_wechat_enabled"`
	WechatAppID            string `json:"payment_wechat_app_id"`
	WechatMchID            string `json:"payment_wechat_mch_id"`
	WechatAPIKeyConfigured bool   `json:"payment_wechat_api_key_configured"`

	EpayEnabled       bool   `json:"payment_epay_enabled"`
	EpayAPIURL        string `json:"payment_epay_api_url"`
	EpayPID           string `json:"payment_epay_pid"`
	EpayKeyConfigured bool   `json:"payment_epay_key_configured"`
	EpayType          string `json:"payment_epay_type"`

	ReferralEnabled        bool    `json:"referral_enabled"`
	ReferralCommissionRate float64 `json:"referral_commission_rate"`
}

// UpdatePaymentSettingsRequest 更新支付设置请求
type UpdatePaymentSettingsRequest struct {
	PaymentEnabled       bool    `json:"payment_enabled"`
	PaymentCurrency      string  `json:"payment_currency"`
	PaymentExchangeRate  float64 `json:"payment_exchange_rate"`
	PaymentPresetAmounts string  `json:"payment_preset_amounts"`
	PaymentMinAmount     float64 `json:"payment_min_amount"`
	PaymentMaxAmount     float64 `json:"payment_max_amount"`

	AlipayEnabled    bool   `json:"payment_alipay_enabled"`
	AlipayAppID      string `json:"payment_alipay_app_id"`
	AlipayPrivateKey string `json:"payment_alipay_private_key,omitempty"`
	AlipayPublicKey  string `json:"payment_alipay_public_key,omitempty"`
	AlipayF2FEnabled bool   `json:"payment_alipay_f2f_enabled"`

	WechatEnabled bool   `json:"payment_wechat_enabled"`
	WechatAppID   string `json:"payment_wechat_app_id"`
	WechatMchID   string `json:"payment_wechat_mch_id"`
	WechatAPIKey  string `json:"payment_wechat_api_key,omitempty"`

	EpayEnabled bool   `json:"payment_epay_enabled"`
	EpayAPIURL  string `json:"payment_epay_api_url"`
	EpayPID     string `json:"payment_epay_pid"`
	EpayKey     string `json:"payment_epay_key,omitempty"`
	EpayType    string `json:"payment_epay_type"`

	ReferralEnabled        bool    `json:"referral_enabled"`
	ReferralCommissionRate float64 `json:"referral_commission_rate"`
}
