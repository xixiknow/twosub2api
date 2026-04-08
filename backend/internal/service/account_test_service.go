package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// sseDataPrefix matches SSE data lines with optional whitespace after colon.
// Some upstream APIs return non-standard "data:" without space (should be "data: ").
var sseDataPrefix = regexp.MustCompile(`^data:\s*`)

const (
	testClaudeAPIURL   = "https://api.anthropic.com/v1/messages"
	chatgptCodexAPIURL = "https://chatgpt.com/backend-api/codex/responses"
)

// TestEvent represents a SSE event for account testing
type TestEvent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Model    string `json:"model,omitempty"`
	Status   string `json:"status,omitempty"`
	Code     string `json:"code,omitempty"`
	Data     any    `json:"data,omitempty"`
	Success  bool   `json:"success,omitempty"`
	Error    string `json:"error,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

// endpointDef describes an API endpoint format for auto-detection
type endpointDef struct {
	Name       string                                           // Human-readable name, e.g. "OpenAI Chat"
	Path       string                                           // URL path suffix, e.g. "/v1/chat/completions"
	BuildProbe func(modelID string) ([]byte, map[string]string) // Returns (body, extra headers)
}

// autoDetectEndpoints is the ordered list of endpoints to probe during auto-detection.
var autoDetectEndpoints = []endpointDef{
	{
		Name: "OpenAI Chat (v1/chat/completions)",
		Path: "/v1/chat/completions",
		BuildProbe: func(modelID string) ([]byte, map[string]string) {
			prompt := buildModelAwareTestPrompt("openai_compat", modelID)
			payload := map[string]any{
				"model": modelID,
				"messages": []map[string]string{
					{"role": "user", "content": prompt},
				},
				"stream":     true,
				"max_tokens": 100,
			}
			b, _ := json.Marshal(payload)
			return b, map[string]string{"Content-Type": "application/json"}
		},
	},
	{
		Name: "OpenAI Responses (v1/responses)",
		Path: "/v1/responses",
		BuildProbe: func(modelID string) ([]byte, map[string]string) {
			prompt := buildModelAwareTestPrompt("openai_responses", modelID)
			payload := map[string]any{
				"model": modelID,
				"input": []map[string]any{
					{
						"role": "user",
						"content": []map[string]any{
							{"type": "input_text", "text": prompt},
						},
					},
				},
				"stream": true,
			}
			b, _ := json.Marshal(payload)
			return b, map[string]string{"Content-Type": "application/json"}
		},
	},
	{
		Name: "Anthropic (v1/messages)",
		Path: "/v1/messages",
		BuildProbe: func(modelID string) ([]byte, map[string]string) {
			prompt := buildModelAwareTestPrompt("anthropic", modelID)
			payload := map[string]any{
				"model": modelID,
				"messages": []map[string]any{
					{
						"role": "user",
						"content": []map[string]any{
							{"type": "text", "text": prompt},
						},
					},
				},
				"max_tokens": 100,
				"stream":     true,
			}
			b, _ := json.Marshal(payload)
			return b, map[string]string{
				"Content-Type":      "application/json",
				"anthropic-version": "2023-06-01",
			}
		},
	},
	{
		Name: "Gemini (v1beta/models/generateContent)",
		Path: "/v1beta/models/{model}:generateContent",
		BuildProbe: func(modelID string) ([]byte, map[string]string) {
			m := modelID
			if m == "" {
				m = "gemini-2.0-flash"
			}
			prompt := buildModelAwareTestPrompt("gemini", m)
			payload := map[string]any{
				"contents": []map[string]any{
					{
						"role": "user",
						"parts": []map[string]any{
							{"text": prompt},
						},
					},
				},
			}
			b, _ := json.Marshal(payload)
			return b, map[string]string{
				"Content-Type":  "application/json",
				"_gemini_model": m, // internal marker for path substitution
			}
		},
	},
	{
		Name: "Jina Rerank (v1/rerank)",
		Path: "/v1/rerank",
		BuildProbe: func(modelID string) ([]byte, map[string]string) {
			payload := map[string]any{
				"model":     "jina-reranker-v2-base-multilingual",
				"query":     "test",
				"top_n":     1,
				"documents": []string{"test document"},
			}
			b, _ := json.Marshal(payload)
			return b, map[string]string{"Content-Type": "application/json"}
		},
	},
}

// AccountTestService handles account testing operations
type AccountTestService struct {
	accountRepo               AccountRepository
	geminiTokenProvider       *GeminiTokenProvider
	antigravityGatewayService *AntigravityGatewayService
	httpUpstream              HTTPUpstream
	cfg                       *config.Config
}

// NewAccountTestService creates a new AccountTestService
func NewAccountTestService(
	accountRepo AccountRepository,
	geminiTokenProvider *GeminiTokenProvider,
	antigravityGatewayService *AntigravityGatewayService,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
) *AccountTestService {
	return &AccountTestService{
		accountRepo:               accountRepo,
		geminiTokenProvider:       geminiTokenProvider,
		antigravityGatewayService: antigravityGatewayService,
		httpUpstream:              httpUpstream,
		cfg:                       cfg,
	}
}

func (s *AccountTestService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg == nil {
		return "", errors.New("config is not available")
	}
	if !s.cfg.Security.URLAllowlist.Enabled {
		// 白名单未启用时，始终允许 http（中转场景常见 http 地址）
		return urlvalidator.ValidateURLFormat(raw, true)
	}
	// Use ValidateHTTPURL instead of ValidateHTTPSURL so that AllowInsecureHTTP is respected
	// even when the URL allowlist is enabled.
	normalized, err := urlvalidator.ValidateHTTPURL(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", err
	}
	return normalized, nil
}

// generateSessionString generates a Claude Code style session string
func generateSessionString() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	hex64 := hex.EncodeToString(bytes)
	sessionUUID := uuid.New().String()
	return fmt.Sprintf("user_%s_account__session_%s", hex64, sessionUUID), nil
}

// createTestPayload creates a Claude Code style test request payload
func createTestPayload(modelID string) (map[string]any, error) {
	sessionID, err := generateSessionString()
	if err != nil {
		return nil, err
	}
	prompt := buildModelAwareTestPrompt("anthropic", modelID)

	return map[string]any{
		"model": modelID,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": prompt,
						"cache_control": map[string]string{
							"type": "ephemeral",
						},
					},
				},
			},
		},
		"system": []map[string]any{
			{
				"type": "text",
				"text": claudeCodeSystemPrompt,
				"cache_control": map[string]string{
					"type": "ephemeral",
				},
			},
		},
		"metadata": map[string]string{
			"user_id": sessionID,
		},
		"max_tokens":  1024,
		"temperature": 1,
		"stream":      true,
	}, nil
}

func buildModelAwareTestPrompt(platform, modelID string) string {
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = "unknown"
	}
	return fmt.Sprintf(
		"[%s connectivity test] Reply briefly in Chinese and include marker [MODEL:%s].",
		platform,
		model,
	)
}

// RunTestBackground runs a connectivity test without an HTTP/SSE context and returns a ScheduledTestResult.
// 根据账号平台路由到对应的端点，避免 anthropic 账号走 OpenAI Codex 等不兼容端点。
func (s *AccountTestService) RunTestBackground(ctx context.Context, accountID int64, modelID string) (*ScheduledTestResult, error) {
	startedAt := time.Now()

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return &ScheduledTestResult{
			Status:       "error",
			ErrorMessage: "Account not found: " + err.Error(),
			StartedAt:    startedAt,
			FinishedAt:   time.Now(),
			LatencyMs:    time.Since(startedAt).Milliseconds(),
		}, nil
	}

	testModel := modelID
	if testModel == "" {
		testModel = s.getBackgroundDefaultModel(account)
	}

	// 应用模型映射（如果账号配置了 model_mapping）
	if mapping := account.GetModelMapping(); len(mapping) > 0 {
		if mappedModel, exists := mapping[testModel]; exists {
			testModel = mappedModel
		}
	}

	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// 根据平台路由到对应的后台测试逻辑
	switch account.Platform {
	case PlatformOpenAI:
		return s.runBackgroundOpenAI(ctx, account, testModel, proxyURL, startedAt)
	case PlatformGemini:
		return s.runBackgroundGemini(ctx, account, testModel, proxyURL, startedAt)
	default:
		// Anthropic, Antigravity 以及其他 OpenAI 兼容平台走原有逻辑
		return s.runBackgroundAnthropic(ctx, account, testModel, proxyURL, startedAt)
	}
}

// runBackgroundAnthropic 处理 Anthropic/Antigravity 及 OpenAI 兼容平台的后台测试
func (s *AccountTestService) runBackgroundAnthropic(ctx context.Context, account *Account, testModel, proxyURL string, startedAt time.Time) (*ScheduledTestResult, error) {
	result := &ScheduledTestResult{
		Status:    "success",
		StartedAt: startedAt,
	}

	baseURL := account.GetCredential("base_url")
	if baseURL == "" {
		switch account.Platform {
		case PlatformAntigravity:
			baseURL = "https://api.anthropic.com"
		case PlatformQwen:
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode"
		case PlatformDeepSeek:
			baseURL = "https://api.deepseek.com"
		case PlatformGLM:
			baseURL = "https://open.bigmodel.cn/api/paas"
		case PlatformKimi:
			baseURL = "https://api.moonshot.cn"
		case PlatformIFlow:
			baseURL = "https://api.iflow.cn"
		default:
			baseURL = "https://api.anthropic.com"
		}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		apiKey = account.GetCredential("session_key")
	}

	// 对于 OpenAI 兼容平台（Qwen/DeepSeek/GLM/Kimi/iFlow），使用 OpenAI Chat 端点
	if account.IsOpenAICompatible() {
		return s.runBackgroundOpenAICompat(ctx, account, baseURL, apiKey, proxyURL, testModel, startedAt)
	}

	// 对于非默认 base URL 的 Anthropic/Antigravity 账号，使用平台感知的自动探测
	defaultBaseURL := "https://api.anthropic.com"
	if account.Platform == PlatformAntigravity {
		// Antigravity 带 /antigravity 后缀
		if strings.HasSuffix(baseURL, "/antigravity") {
			baseURL = strings.TrimSuffix(baseURL, "/antigravity")
		}
	}
	if baseURL != defaultBaseURL {
		return s.runBackgroundWithAutoDetection(ctx, account, baseURL, apiKey, proxyURL, testModel, startedAt)
	}

	// 默认 Anthropic 端点
	testPrompt := buildModelAwareTestPrompt(string(account.Platform), testModel)
	reqBody := fmt.Sprintf(`{"model":"%s","max_tokens":64,"messages":[{"role":"user","content":"%s"}]}`, testModel, testPrompt)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", strings.NewReader(reqBody))
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Failed to create request: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Request failed: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result.FinishedAt = time.Now()
	result.LatencyMs = time.Since(startedAt).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("Upstream returned status %d: %s", resp.StatusCode, string(body))
	} else {
		result.ResponseText = string(body)
		if len(result.ResponseText) > 2000 {
			result.ResponseText = result.ResponseText[:2000]
		}
	}

	return result, nil
}

// runBackgroundOpenAI 处理 OpenAI 平台的后台测试
func (s *AccountTestService) runBackgroundOpenAI(ctx context.Context, account *Account, testModel, proxyURL string, startedAt time.Time) (*ScheduledTestResult, error) {
	result := &ScheduledTestResult{
		Status:    "success",
		StartedAt: startedAt,
	}

	var apiURL string
	var authToken string

	if account.IsOAuth() {
		// OAuth 账号不走 Codex 端点做后台测试，使用 /v1/chat/completions 兼容端点
		authToken = account.GetOpenAIAccessToken()
		if authToken == "" {
			authToken = account.GetCredential("access_token")
		}
		apiURL = "https://api.openai.com/v1/chat/completions"
	} else if account.Type == AccountTypeAPIKey {
		authToken = account.GetOpenAIApiKey()
		if authToken == "" {
			authToken = account.GetCredential("api_key")
		}
		baseURL := account.GetOpenAIBaseURL()
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		baseURL = strings.TrimSuffix(baseURL, "/")

		// 非默认 base URL 使用平台感知自动探测
		if baseURL != "https://api.openai.com" {
			return s.runBackgroundWithAutoDetection(ctx, account, baseURL, authToken, proxyURL, testModel, startedAt)
		}
		apiURL = baseURL + "/v1/chat/completions"
	} else {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("Unsupported OpenAI account type: %s", account.Type)
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}

	testPrompt := buildModelAwareTestPrompt("openai", testModel)
	payload := map[string]any{
		"model": testModel,
		"messages": []map[string]string{
			{"role": "user", "content": testPrompt},
		},
		"max_tokens": 64,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Failed to create request: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Request failed: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result.FinishedAt = time.Now()
	result.LatencyMs = time.Since(startedAt).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("Upstream returned status %d: %s", resp.StatusCode, string(body))
	} else {
		result.ResponseText = string(body)
		if len(result.ResponseText) > 2000 {
			result.ResponseText = result.ResponseText[:2000]
		}
	}

	return result, nil
}

// runBackgroundGemini 处理 Gemini 平台的后台测试
func (s *AccountTestService) runBackgroundGemini(ctx context.Context, account *Account, testModel, proxyURL string, startedAt time.Time) (*ScheduledTestResult, error) {
	result := &ScheduledTestResult{
		Status:    "success",
		StartedAt: startedAt,
	}

	baseURL := account.GetCredential("base_url")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	apiKey := account.GetCredential("api_key")

	if testModel == "" {
		testModel = geminicli.DefaultTestModel
	}

	testPrompt := buildModelAwareTestPrompt("gemini", testModel)
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": testPrompt},
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	apiURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", baseURL, testModel)
	if apiKey != "" {
		apiURL += "?key=" + apiKey
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Failed to create request: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Request failed: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result.FinishedAt = time.Now()
	result.LatencyMs = time.Since(startedAt).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("Upstream returned status %d: %s", resp.StatusCode, string(body))
	} else {
		result.ResponseText = string(body)
		if len(result.ResponseText) > 2000 {
			result.ResponseText = result.ResponseText[:2000]
		}
	}

	return result, nil
}

// runBackgroundOpenAICompat 处理 OpenAI 兼容平台（Qwen/DeepSeek/GLM/Kimi/iFlow）的后台测试
func (s *AccountTestService) runBackgroundOpenAICompat(ctx context.Context, account *Account, baseURL, apiKey, proxyURL, testModel string, startedAt time.Time) (*ScheduledTestResult, error) {
	result := &ScheduledTestResult{
		Status:    "success",
		StartedAt: startedAt,
	}

	testPrompt := buildModelAwareTestPrompt(string(account.Platform), testModel)
	payload := map[string]any{
		"model": testModel,
		"messages": []map[string]string{
			{"role": "user", "content": testPrompt},
		},
		"max_tokens": 64,
	}
	payloadBytes, _ := json.Marshal(payload)

	apiURL := baseURL + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Failed to create request: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = "Request failed: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result.FinishedAt = time.Now()
	result.LatencyMs = time.Since(startedAt).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("Upstream returned status %d: %s", resp.StatusCode, string(body))
	} else {
		result.ResponseText = string(body)
		if len(result.ResponseText) > 2000 {
			result.ResponseText = result.ResponseText[:2000]
		}
	}

	return result, nil
}

// runBackgroundWithAutoDetection probes autoDetectEndpoints to find a working
// endpoint on a custom base URL, then performs the actual test request.
// 使用 reorderEndpointsForPlatform 按平台优先排序端点，避免 anthropic 账号被误路由到 Codex 端点。
func (s *AccountTestService) runBackgroundWithAutoDetection(
	ctx context.Context, account *Account,
	baseURL, apiKey, proxyURL, testModel string,
	startedAt time.Time,
) (*ScheduledTestResult, error) {
	result := &ScheduledTestResult{Status: "error", StartedAt: startedAt}

	// 按平台重排端点探测顺序，确保优先探测本平台对应的端点
	endpoints := reorderEndpointsForPlatform(account.Platform)

	// Probe each endpoint to find one that doesn't 404
	var matchedIdx = -1
	for i := range endpoints {
		ep := &endpoints[i]
		probeBody, extraHeaders := ep.BuildProbe(testModel)

		fullURL := baseURL + ep.Path
		if strings.Contains(ep.Path, "{model}") {
			m := testModel
			if gm, ok := extraHeaders["_gemini_model"]; ok && gm != "" {
				m = gm
			}
			if m == "" {
				m = "gemini-2.0-flash"
			}
			fullURL = baseURL + strings.Replace(ep.Path, "{model}", m, 1)
		}
		delete(extraHeaders, "_gemini_model")

		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		probeReq, err := http.NewRequestWithContext(probeCtx, "POST", fullURL, bytes.NewReader(probeBody))
		if err != nil {
			cancel()
			continue
		}
		for k, v := range extraHeaders {
			probeReq.Header.Set(k, v)
		}
		if strings.Contains(ep.Path, "/v1/messages") {
			probeReq.Header.Set("x-api-key", apiKey)
		} else {
			probeReq.Header.Set("Authorization", "Bearer "+apiKey)
		}

		resp, err := s.httpUpstream.Do(probeReq, proxyURL, account.ID, account.Concurrency)
		cancel()
		if err != nil {
			continue
		}
		sc := resp.StatusCode
		_ = resp.Body.Close()

		if sc == http.StatusNotFound || sc == http.StatusMethodNotAllowed {
			continue
		}
		matchedIdx = i
		break
	}

	if matchedIdx == -1 {
		result.ErrorMessage = "No working API endpoint detected on " + baseURL
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}

	// Send the real test request using the matched endpoint
	ep := &endpoints[matchedIdx]
	reqBody, extraHeaders := ep.BuildProbe(testModel)
	fullURL := baseURL + ep.Path
	if strings.Contains(ep.Path, "{model}") {
		m := testModel
		if gm, ok := extraHeaders["_gemini_model"]; ok && gm != "" {
			m = gm
		}
		if m == "" {
			m = "gemini-2.0-flash"
		}
		fullURL = baseURL + strings.Replace(ep.Path, "{model}", m, 1)
	}
	delete(extraHeaders, "_gemini_model")

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(reqBody))
	if err != nil {
		result.ErrorMessage = "Failed to create request: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	if strings.Contains(ep.Path, "/v1/messages") {
		req.Header.Set("x-api-key", apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		result.ErrorMessage = "Request failed: " + err.Error()
		result.FinishedAt = time.Now()
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result.FinishedAt = time.Now()
	result.LatencyMs = time.Since(startedAt).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = "success"
		result.ResponseText = string(body)
		if len(result.ResponseText) > 2000 {
			result.ResponseText = result.ResponseText[:2000]
		}
	} else {
		result.ErrorMessage = fmt.Sprintf("Upstream returned status %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// TestAccountConnection tests an account's connection by sending a test request
// All account types use full Claude Code client characteristics, only auth header differs
// modelID is optional - if empty, defaults to claude.DefaultTestModel
func (s *AccountTestService) TestAccountConnection(c *gin.Context, accountID int64, modelID string, prompt ...string) error {
	ctx := c.Request.Context()

	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return s.sendErrorAndEnd(c, "Account not found")
	}

	// Route to platform-specific test method
	if account.IsOpenAI() {
		return s.testOpenAIAccountConnection(c, account, modelID)
	}

	if account.IsGemini() {
		return s.testGeminiAccountConnection(c, account, modelID)
	}

	if account.Platform == PlatformAntigravity {
		return s.testAntigravityAccountConnection(c, account, modelID)
	}

	// Handle OpenAI-compatible platforms (Qwen, DeepSeek, GLM, Kimi, iFlow)
	if account.IsOpenAICompatible() {
		return s.testOpenAICompatAccountConnection(c, account, modelID)
	}

	return s.testClaudeAccountConnection(c, account, modelID)
}

// testClaudeAccountConnection tests an Anthropic Claude account's connection
func (s *AccountTestService) testClaudeAccountConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()

	// Determine the model to use
	testModelID := modelID
	if testModelID == "" {
		testModelID = claude.DefaultTestModel
	}

	// For API Key accounts with model mapping, map the model
	if account.Type == "apikey" {
		mapping := account.GetModelMapping()
		if len(mapping) > 0 {
			if mappedModel, exists := mapping[testModelID]; exists {
				testModelID = mappedModel
			}
		}
	}

	// Determine authentication method and API URL
	var authToken string
	var useBearer bool
	var apiURL string

	if account.IsOAuth() {
		// OAuth or Setup Token - use Bearer token
		useBearer = true
		apiURL = testClaudeAPIURL
		authToken = account.GetCredential("access_token")
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No access token available")
		}
	} else if account.Type == "apikey" {
		// API Key - use x-api-key header
		useBearer = false
		authToken = account.GetCredential("api_key")
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No API key available")
		}

		baseURL := account.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://api.anthropic.com"
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
		}

		// For non-default base URLs, use auto endpoint detection
		if baseURL != "https://api.anthropic.com" {
			return s.testAPIKeyWithAutoDetection(c, account, normalizedBaseURL, authToken, testModelID)
		}

		apiURL = strings.TrimSuffix(normalizedBaseURL, "/") + "/v1/messages?beta=true"
	} else {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Create Claude Code style payload (same for all account types)
	payload, err := createTestPayload(testModelID)
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create test payload")
	}
	payloadBytes, _ := json.Marshal(payload)

	// Send test_start event
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	// Apply Claude Code client headers
	for key, value := range claude.DefaultHeaders {
		req.Header.Set(key, value)
	}

	// Set authentication header
	if useBearer {
		req.Header.Set("anthropic-beta", claude.DefaultBetaHeader)
		req.Header.Set("Authorization", "Bearer "+authToken)
	} else {
		req.Header.Set("anthropic-beta", claude.APIKeyBetaHeader)
		req.Header.Set("x-api-key", authToken)
	}

	// Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body))
		// 401/403 标记账号为 error 状态
		if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && s.accountRepo != nil {
			_ = s.accountRepo.SetError(c.Request.Context(), account.ID, errMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}

	// Process SSE stream
	return s.processClaudeStream(c, resp.Body)
}

// testOpenAIAccountConnection tests an OpenAI account's connection
func (s *AccountTestService) testOpenAIAccountConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()

	// Default to openai.DefaultTestModel for OpenAI testing
	testModelID := modelID
	if testModelID == "" {
		testModelID = openai.DefaultTestModel
	}

	// For API Key accounts with model mapping, map the model
	if account.Type == "apikey" {
		mapping := account.GetModelMapping()
		if len(mapping) > 0 {
			if mappedModel, exists := mapping[testModelID]; exists {
				testModelID = mappedModel
			}
		}
	}

	// Determine authentication method and API URL
	var authToken string
	var apiURL string
	var isOAuth bool
	var chatgptAccountID string

	if account.IsOAuth() {
		isOAuth = true
		// OAuth - use Bearer token with ChatGPT internal API
		authToken = account.GetOpenAIAccessToken()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No access token available")
		}

		// OAuth uses ChatGPT internal API
		apiURL = chatgptCodexAPIURL
		chatgptAccountID = account.GetChatGPTAccountID()
	} else if account.Type == "apikey" {
		// API Key - use Platform API
		authToken = account.GetOpenAIApiKey()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No API key available")
		}

		baseURL := account.GetOpenAIBaseURL()
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
		}

		// For non-default base URLs, use auto endpoint detection
		if baseURL != "https://api.openai.com" {
			return s.testAPIKeyWithAutoDetection(c, account, normalizedBaseURL, authToken, testModelID)
		}

		apiURL = strings.TrimSuffix(normalizedBaseURL, "/") + "/responses"
	} else {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Create OpenAI Responses API payload
	payload := createOpenAITestPayload(testModelID, isOAuth)
	payloadBytes, _ := json.Marshal(payload)

	// Send test_start event
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Set OAuth-specific headers for ChatGPT internal API
	if isOAuth {
		req.Host = "chatgpt.com"
		req.Header.Set("accept", "text/event-stream")
		if chatgptAccountID != "" {
			req.Header.Set("chatgpt-account-id", chatgptAccountID)
		}
	}

	// Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body))
		if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && s.accountRepo != nil {
			_ = s.accountRepo.SetError(c.Request.Context(), account.ID, errMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}

	// Process SSE stream
	return s.processOpenAIStream(c, resp.Body)
}

// testGeminiAccountConnection tests a Gemini account's connection
func (s *AccountTestService) testGeminiAccountConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()

	// Determine the model to use
	testModelID := modelID
	if testModelID == "" {
		testModelID = geminicli.DefaultTestModel
	}

	// For API Key accounts with model mapping, map the model
	if account.Type == AccountTypeAPIKey {
		mapping := account.GetModelMapping()
		if len(mapping) > 0 {
			if mappedModel, exists := mapping[testModelID]; exists {
				testModelID = mappedModel
			}
		}
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Create test payload (Gemini format)
	payload := createGeminiTestPayload(testModelID)

	// Build request based on account type
	var req *http.Request
	var err error

	switch account.Type {
	case AccountTypeAPIKey:
		req, err = s.buildGeminiAPIKeyRequest(ctx, account, testModelID, payload)
	case AccountTypeOAuth:
		req, err = s.buildGeminiOAuthRequest(ctx, account, testModelID, payload)
	default:
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
	}

	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to build request: %s", err.Error()))
	}

	// Send test_start event
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	// Get proxy and execute request
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return s.sendErrorAndEnd(c, fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
	}

	// Process SSE stream
	return s.processGeminiStream(c, resp.Body)
}

// testAntigravityAccountConnection tests an Antigravity account's connection
// 支持 Claude 和 Gemini 两种协议，使用非流式请求
func (s *AccountTestService) testAntigravityAccountConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()

	// 默认模型：从 antigravity.DefaultModels 中选取第一个可用模型
	testModelID := modelID
	if testModelID == "" {
		models := antigravity.DefaultModels()
		if len(models) > 0 {
			testModelID = models[0].ID
		} else {
			testModelID = "claude-sonnet-4-5"
		}
	}

	// For apikey type, use auto-endpoint detection (same as Claude apikey with custom base URL)
	if account.Type == "apikey" {
		authToken := account.GetCredential("api_key")
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No API key available")
		}

		baseURL := account.GetBaseURL()
		if baseURL == "" {
			return s.sendErrorAndEnd(c, "No base URL configured for API key account")
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
		}

		// Apply model mapping if configured
		mapping := account.GetModelMapping()
		if len(mapping) > 0 {
			if mappedModel, exists := mapping[testModelID]; exists {
				testModelID = mappedModel
			}
		}

		// 根据账号平台选择优先端点，避免 anthropic 平台请求走到 OpenAI 端点
		return s.testAPIKeyWithAutoDetectionPlatformAware(c, account, normalizedBaseURL, authToken, testModelID)
	}

	if s.antigravityGatewayService == nil {
		return s.sendErrorAndEnd(c, "Antigravity gateway service not configured")
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Send test_start event
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	// 调用 AntigravityGatewayService.TestConnection（复用协议转换逻辑）
	result, err := s.antigravityGatewayService.TestConnection(ctx, account, testModelID)
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}

	// 发送响应内容
	if result.Text != "" {
		s.sendEvent(c, TestEvent{Type: "content", Text: result.Text})
	}
	if strings.TrimSpace(result.MappedModel) != "" {
		s.sendEvent(c, TestEvent{
			Type: "content",
			Text: fmt.Sprintf("[MODEL:%s]", result.MappedModel),
		})
	}

	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

// buildGeminiAPIKeyRequest builds request for Gemini API Key accounts
func (s *AccountTestService) buildGeminiAPIKeyRequest(ctx context.Context, account *Account, modelID string, payload []byte) (*http.Request, error) {
	apiKey := account.GetCredential("api_key")
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("no API key available")
	}

	baseURL := account.GetCredential("base_url")
	if baseURL == "" {
		baseURL = geminicli.AIStudioBaseURL
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	// Use streamGenerateContent for real-time feedback
	fullURL := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse",
		strings.TrimRight(normalizedBaseURL, "/"), modelID)

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	return req, nil
}

// buildGeminiOAuthRequest builds request for Gemini OAuth accounts
func (s *AccountTestService) buildGeminiOAuthRequest(ctx context.Context, account *Account, modelID string, payload []byte) (*http.Request, error) {
	if s.geminiTokenProvider == nil {
		return nil, fmt.Errorf("gemini token provider not configured")
	}

	// Get access token (auto-refreshes if needed)
	accessToken, err := s.geminiTokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	projectID := strings.TrimSpace(account.GetCredential("project_id"))
	if projectID == "" {
		// AI Studio OAuth mode (no project_id): call generativelanguage API directly with Bearer token.
		baseURL := account.GetCredential("base_url")
		if strings.TrimSpace(baseURL) == "" {
			baseURL = geminicli.AIStudioBaseURL
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		fullURL := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", strings.TrimRight(normalizedBaseURL, "/"), modelID)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return req, nil
	}

	// Code Assist mode (with project_id)
	return s.buildCodeAssistRequest(ctx, accessToken, projectID, modelID, payload)
}

// buildCodeAssistRequest builds request for Google Code Assist API (used by Gemini CLI and Antigravity)
func (s *AccountTestService) buildCodeAssistRequest(ctx context.Context, accessToken, projectID, modelID string, payload []byte) (*http.Request, error) {
	var inner map[string]any
	if err := json.Unmarshal(payload, &inner); err != nil {
		return nil, err
	}

	wrapped := map[string]any{
		"model":   modelID,
		"project": projectID,
		"request": inner,
	}
	wrappedBytes, _ := json.Marshal(wrapped)

	normalizedBaseURL, err := s.validateUpstreamBaseURL(geminicli.GeminiCliBaseURL)
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/v1internal:streamGenerateContent?alt=sse", normalizedBaseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(wrappedBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", geminicli.GeminiCLIUserAgent)

	return req, nil
}

// createGeminiTestPayload creates a minimal test payload for Gemini API
// 可选第二个参数用于自定义 prompt（图片模型测试等场景）
func createGeminiTestPayload(modelID string, customPrompt ...string) []byte {
	prompt := buildModelAwareTestPrompt("gemini", modelID)
	if len(customPrompt) > 0 && customPrompt[0] != "" {
		prompt = customPrompt[0]
	}
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []map[string]any{
				{"text": "You are a helpful AI assistant."},
			},
		},
	}

	// 图片生成模型需要设置 responseModalities 和 imageConfig
	if isImageGenerationModel(modelID) {
		payload["generationConfig"] = map[string]any{
			"responseModalities": []string{"TEXT", "IMAGE"},
			"imageConfig": map[string]any{
				"aspectRatio": "1:1",
			},
		}
	}

	bytes, _ := json.Marshal(payload)
	return bytes
}

// processGeminiStream processes SSE stream from Gemini API
func (s *AccountTestService) processGeminiStream(c *gin.Context, body io.Reader) error {
	reader := bufio.NewReader(body)
	modelReported := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
				return nil
			}
			return s.sendErrorAndEnd(c, fmt.Sprintf("Stream read error: %s", err.Error()))
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimPrefix(line, "data: ")
		if jsonStr == "[DONE]" {
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		// Support two Gemini response formats:
		// - AI Studio: {"candidates": [...]}
		// - Gemini CLI: {"response": {"candidates": [...]}}
		if resp, ok := data["response"].(map[string]any); ok && resp != nil {
			data = resp
		}
		if !modelReported {
			if mv, ok := data["modelVersion"].(string); ok && strings.TrimSpace(mv) != "" {
				s.sendEvent(c, TestEvent{
					Type: "content",
					Text: fmt.Sprintf("[MODEL:%s]", strings.TrimSpace(mv)),
				})
				modelReported = true
			}
		}
		if candidates, ok := data["candidates"].([]any); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]any); ok {
				// Extract content first (before checking completion)
				if content, ok := candidate["content"].(map[string]any); ok {
					if parts, ok := content["parts"].([]any); ok {
						for _, part := range parts {
							if partMap, ok := part.(map[string]any); ok {
								if text, ok := partMap["text"].(string); ok && text != "" {
									s.sendEvent(c, TestEvent{Type: "content", Text: text})
								}
							}
						}
					}
				}

				// Check for completion after extracting content
				if finishReason, ok := candidate["finishReason"].(string); ok && finishReason != "" {
					s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
					return nil
				}
			}
		}

		// Handle errors
		if errData, ok := data["error"].(map[string]any); ok {
			errorMsg := "Unknown error"
			if msg, ok := errData["message"].(string); ok {
				errorMsg = msg
			}
			return s.sendErrorAndEnd(c, errorMsg)
		}
	}
}

// createOpenAITestPayload creates a test payload for OpenAI Responses API
func createOpenAITestPayload(modelID string, isOAuth bool) map[string]any {
	prompt := buildModelAwareTestPrompt("openai", modelID)
	payload := map[string]any{
		"model": modelID,
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": prompt,
					},
				},
			},
		},
		"stream": true,
	}

	// OAuth accounts using ChatGPT internal API require store: false
	if isOAuth {
		payload["store"] = false
	}

	// All accounts require instructions for Responses API
	payload["instructions"] = openai.DefaultInstructions

	return payload
}

// processClaudeStream processes the SSE stream from Claude API
func (s *AccountTestService) processClaudeStream(c *gin.Context, body io.Reader) error {
	reader := bufio.NewReader(body)
	modelReported := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
				return nil
			}
			return s.sendErrorAndEnd(c, fmt.Sprintf("Stream read error: %s", err.Error()))
		}

		line = strings.TrimSpace(line)
		if line == "" || !sseDataPrefix.MatchString(line) {
			continue
		}

		jsonStr := sseDataPrefix.ReplaceAllString(line, "")
		if jsonStr == "[DONE]" {
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		eventType, _ := data["type"].(string)
		switch eventType {
		case "message_start":
			if !modelReported {
				if msg, ok := data["message"].(map[string]any); ok {
					if m, ok := msg["model"].(string); ok && strings.TrimSpace(m) != "" {
						s.sendEvent(c, TestEvent{
							Type: "content",
							Text: fmt.Sprintf("[MODEL:%s]", strings.TrimSpace(m)),
						})
						modelReported = true
					}
				}
			}
		case "content_block_delta":
			if delta, ok := data["delta"].(map[string]any); ok {
				if text, ok := delta["text"].(string); ok {
					s.sendEvent(c, TestEvent{Type: "content", Text: text})
				}
			}
		case "message_stop":
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		case "error":
			errorMsg := "Unknown error"
			if errData, ok := data["error"].(map[string]any); ok {
				if msg, ok := errData["message"].(string); ok {
					errorMsg = msg
				}
			}
			return s.sendErrorAndEnd(c, errorMsg)
		}
	}
}

// processOpenAIStream processes the SSE stream from OpenAI Responses API
func (s *AccountTestService) processOpenAIStream(c *gin.Context, body io.Reader) error {
	reader := bufio.NewReader(body)
	modelReported := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
				return nil
			}
			return s.sendErrorAndEnd(c, fmt.Sprintf("Stream read error: %s", err.Error()))
		}

		line = strings.TrimSpace(line)
		if line == "" || !sseDataPrefix.MatchString(line) {
			continue
		}

		jsonStr := sseDataPrefix.ReplaceAllString(line, "")
		if jsonStr == "[DONE]" {
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		eventType, _ := data["type"].(string)
		switch eventType {
		case "response.created":
			if !modelReported {
				if respData, ok := data["response"].(map[string]any); ok {
					if m, ok := respData["model"].(string); ok && strings.TrimSpace(m) != "" {
						s.sendEvent(c, TestEvent{
							Type: "content",
							Text: fmt.Sprintf("[MODEL:%s]", strings.TrimSpace(m)),
						})
						modelReported = true
					}
				}
			}
		case "response.output_text.delta":
			// OpenAI Responses API uses "delta" field for text content
			if delta, ok := data["delta"].(string); ok && delta != "" {
				s.sendEvent(c, TestEvent{Type: "content", Text: delta})
			}
		case "response.completed":
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		case "error":
			errorMsg := "Unknown error"
			if errData, ok := data["error"].(map[string]any); ok {
				if msg, ok := errData["message"].(string); ok {
					errorMsg = msg
				}
			}
			return s.sendErrorAndEnd(c, errorMsg)
		}
	}
}

// sendEvent sends a SSE event to the client
func (s *AccountTestService) sendEvent(c *gin.Context, event TestEvent) {
	eventJSON, _ := json.Marshal(event)
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", eventJSON); err != nil {
		log.Printf("failed to write SSE event: %v", err)
		return
	}
	c.Writer.Flush()
}

// sendErrorAndEnd sends an error event and ends the stream
func (s *AccountTestService) sendErrorAndEnd(c *gin.Context, errorMsg string) error {
	log.Printf("Account test error: %s", errorMsg)
	s.sendEvent(c, TestEvent{Type: "error", Error: errorMsg})
	return fmt.Errorf("%s", errorMsg)
}

// testOpenAICompatAccountConnection tests an OpenAI-compatible platform account (Qwen, DeepSeek, GLM, Kimi, iFlow)
func (s *AccountTestService) testOpenAICompatAccountConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()

	// Determine the model to use - use the first default model for the platform
	testModelID := modelID
	if testModelID == "" {
		testModelID = s.getDefaultTestModelForPlatform(account)
	}

	// Determine authentication token
	var authToken string
	if account.IsOAuth() {
		authToken = account.GetCredential("access_token")
		if authToken == "" {
			authToken = account.GetCredential("api_key")
		}
	} else {
		authToken = account.GetCredential("api_key")
	}
	if authToken == "" {
		return s.sendErrorAndEnd(c, "No authentication token available")
	}

	// Get base URL
	baseURL := account.GetBaseURL()
	if baseURL == "" {
		return s.sendErrorAndEnd(c, "No base URL configured for this account")
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
	}
	apiURL := strings.TrimSuffix(normalizedBaseURL, "/") + "/chat/completions"

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Create standard chat completion payload
	testPrompt := buildModelAwareTestPrompt(account.Platform, testModelID)
	payload := map[string]any{
		"model": testModelID,
		"messages": []map[string]string{
			{"role": "user", "content": testPrompt},
		},
		"stream":     true,
		"max_tokens": 100,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Send test_start event
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return s.sendErrorAndEnd(c, fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
	}

	// Process standard chat completion SSE stream
	return s.processChatCompletionStream(c, resp.Body)
}

// getDefaultTestModelForPlatform returns a sensible default test model for the given platform
func (s *AccountTestService) getDefaultTestModelForPlatform(account *Account) string {
	switch account.Platform {
	case PlatformQwen:
		if account.IsOAuth() {
			return "qwen3-coder-plus"
		}
		return "qwen-plus"
	case PlatformDeepSeek:
		return "deepseek-chat"
	case PlatformGLM:
		return "glm-4-flash"
	case PlatformKimi:
		return "kimi-k2"
	case PlatformIFlow:
		return "qwen3-max"
	default:
		return "gpt-4o-mini"
	}
}

// getBackgroundDefaultModel returns the appropriate default test model for RunTestBackground.
// Unlike getDefaultTestModelForPlatform (which is for upstream/multi-platform accounts),
// this function covers all platform types including anthropic, openai, gemini, antigravity.
func (s *AccountTestService) getBackgroundDefaultModel(account *Account) string {
	switch account.Platform {
	case PlatformAnthropic:
		return claude.DefaultTestModel
	case PlatformOpenAI:
		return openai.DefaultTestModel
	case PlatformGemini:
		return geminicli.DefaultTestModel
	case PlatformAntigravity:
		return claude.DefaultTestModel
	case PlatformQwen:
		if account.IsOAuth() {
			return "qwen3-coder-plus"
		}
		return "qwen-plus"
	case PlatformDeepSeek:
		return "deepseek-chat"
	case PlatformGLM:
		return "glm-4-flash"
	case PlatformKimi:
		return "kimi-k2"
	case PlatformIFlow:
		return "qwen3-max"
	default:
		return claude.DefaultTestModel
	}
}

// processChatCompletionStream processes the SSE stream from standard OpenAI Chat Completions API
func (s *AccountTestService) processChatCompletionStream(c *gin.Context, body io.Reader) error {
	reader := bufio.NewReader(body)
	modelReported := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
				return nil
			}
			return s.sendErrorAndEnd(c, fmt.Sprintf("Stream read error: %s", err.Error()))
		}

		line = strings.TrimSpace(line)
		if line == "" || !sseDataPrefix.MatchString(line) {
			continue
		}

		jsonStr := sseDataPrefix.ReplaceAllString(line, "")
		if jsonStr == "[DONE]" {
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}
		if !modelReported {
			if m, ok := data["model"].(string); ok && strings.TrimSpace(m) != "" {
				s.sendEvent(c, TestEvent{
					Type: "content",
					Text: fmt.Sprintf("[MODEL:%s]", strings.TrimSpace(m)),
				})
				modelReported = true
			}
		}

		// Standard chat completion format: choices[0].delta.content
		choices, ok := data["choices"].([]any)
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]any)
		if !ok {
			continue
		}
		delta, ok := choice["delta"].(map[string]any)
		if !ok {
			continue
		}
		if content, ok := delta["content"].(string); ok && content != "" {
			s.sendEvent(c, TestEvent{Type: "content", Text: content})
		}
	}
}

// testAPIKeyWithAutoDetectionPlatformAware 根据账号平台优先探测对应的端点格式。
// 对于 anthropic 平台优先 /v1/messages，gemini 平台优先 Gemini 端点，
// 避免 anthropic 账号被误路由到 OpenAI Codex 号池。
func (s *AccountTestService) testAPIKeyWithAutoDetectionPlatformAware(c *gin.Context, account *Account, normalizedBaseURL, authToken, testModelID string) error {
	// 根据平台重排端点探测顺序
	reordered := reorderEndpointsForPlatform(account.Platform)
	return s.testAPIKeyWithAutoDetectionCustomEndpoints(c, account, normalizedBaseURL, authToken, testModelID, reordered)
}

// reorderEndpointsForPlatform 将指定平台对应的端点排到最前面
func reorderEndpointsForPlatform(platform string) []endpointDef {
	// 确定平台对应的端点路径前缀
	var preferredPath string
	switch platform {
	case PlatformAnthropic:
		preferredPath = "/v1/messages"
	case PlatformGemini:
		preferredPath = "/v1beta/models/"
	default:
		// openai, deepseek, qwen 等默认用原始顺序（OpenAI 端点优先）
		return autoDetectEndpoints
	}

	reordered := make([]endpointDef, 0, len(autoDetectEndpoints))
	var rest []endpointDef
	for _, ep := range autoDetectEndpoints {
		if strings.Contains(ep.Path, preferredPath) {
			reordered = append(reordered, ep)
		} else {
			rest = append(rest, ep)
		}
	}
	reordered = append(reordered, rest...)
	return reordered
}

// testAPIKeyWithAutoDetectionCustomEndpoints 与 testAPIKeyWithAutoDetection 相同，
// 但使用自定义的端点列表（允许按平台重排顺序）
func (s *AccountTestService) testAPIKeyWithAutoDetectionCustomEndpoints(c *gin.Context, account *Account, normalizedBaseURL, authToken, testModelID string, endpoints []endpointDef) error {
	ctx := c.Request.Context()

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	// Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	base := strings.TrimSuffix(normalizedBaseURL, "/")

	s.sendEvent(c, TestEvent{Type: "content", Text: "Detecting API endpoint format ...\n"})

	var detectedIdx = -1
	var successIdx = -1

	for i, ep := range endpoints {
		epPath := ep.Path
		probeBody, extraHeaders := ep.BuildProbe(testModelID)

		fullURL := base + epPath

		if strings.Contains(epPath, "{model}") {
			geminiModel := testModelID
			if m, ok := extraHeaders["_gemini_model"]; ok && m != "" {
				geminiModel = m
				delete(extraHeaders, "_gemini_model")
			}
			if geminiModel == "" {
				geminiModel = "gemini-2.0-flash"
			}
			fullURL = base + strings.Replace(epPath, "{model}", geminiModel, 1)
		}
		delete(extraHeaders, "_gemini_model")

		probeReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(probeBody))
		if err != nil {
			continue
		}

		for k, v := range extraHeaders {
			probeReq.Header.Set(k, v)
		}
		if strings.Contains(epPath, "/v1/messages") {
			probeReq.Header.Set("x-api-key", authToken)
		} else {
			probeReq.Header.Set("Authorization", "Bearer "+authToken)
		}

		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		probeReq = probeReq.WithContext(probeCtx)

		resp, err := s.httpUpstream.Do(probeReq, proxyURL, account.ID, account.Concurrency)
		cancel()
		if err != nil {
			continue
		}

		statusCode := resp.StatusCode
		_ = resp.Body.Close()

		if statusCode == http.StatusNotFound || statusCode == http.StatusMethodNotAllowed {
			continue
		}

		s.sendEvent(c, TestEvent{
			Type:     "endpoint_check",
			Endpoint: ep.Name,
			Status:   "success",
			Text:     fmt.Sprintf("HTTP %d - endpoint detected", statusCode),
		})
		if detectedIdx == -1 {
			detectedIdx = i
		}
		if successIdx == -1 && statusCode >= 200 && statusCode < 300 {
			successIdx = i
			break
		}
	}

	finalIdx := successIdx
	if finalIdx == -1 {
		finalIdx = detectedIdx
	}

	if finalIdx == -1 {
		return s.sendErrorAndEnd(c, "No supported API endpoint detected on this base URL")
	}

	detected := endpoints[finalIdx]

	s.sendEvent(c, TestEvent{
		Type: "content",
		Text: "\nRunning streaming test ...\n",
	})

	return s.runStreamingEndpointTest(c, ctx, account, base, authToken, testModelID, detected, proxyURL)
}

// testAPIKeyWithAutoDetection probes multiple endpoint formats on a custom base URL to detect
// which API format is supported, then runs a full streaming test on the first detected endpoint.
// 检测结果会记忆到账号 Extra 中，下次直接跳过探测。
func (s *AccountTestService) testAPIKeyWithAutoDetection(c *gin.Context, account *Account, normalizedBaseURL, authToken, testModelID string) error {
	ctx := c.Request.Context()

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	// Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	base := strings.TrimSuffix(normalizedBaseURL, "/")

	// Always do real-time endpoint detection (no cache)
	s.sendEvent(c, TestEvent{Type: "content", Text: "Detecting API endpoint format ...\n"})

	var detectedIdx = -1 // first non-404/405 endpoint (fallback)
	var successIdx = -1  // first endpoint that returned HTTP 2xx (preferred)

	for i, ep := range autoDetectEndpoints {
		// Build the full URL
		epPath := ep.Path
		probeBody, extraHeaders := ep.BuildProbe(testModelID)

		fullURL := base + epPath

		// Handle Gemini path substitution: replace {model} placeholder
		if strings.Contains(epPath, "{model}") {
			geminiModel := testModelID
			if m, ok := extraHeaders["_gemini_model"]; ok && m != "" {
				geminiModel = m
				delete(extraHeaders, "_gemini_model")
			}
			if geminiModel == "" {
				geminiModel = "gemini-2.0-flash"
			}
			fullURL = base + strings.Replace(epPath, "{model}", geminiModel, 1)
		}
		// Clean up internal markers
		delete(extraHeaders, "_gemini_model")

		probeReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(probeBody))
		if err != nil {
			continue
		}

		for k, v := range extraHeaders {
			probeReq.Header.Set(k, v)
		}
		// Use x-api-key for Anthropic-style, Bearer for everything else
		if strings.Contains(epPath, "/v1/messages") {
			probeReq.Header.Set("x-api-key", authToken)
		} else {
			probeReq.Header.Set("Authorization", "Bearer "+authToken)
		}

		// Quick probe with timeout
		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		probeReq = probeReq.WithContext(probeCtx)

		resp, err := s.httpUpstream.Do(probeReq, proxyURL, account.ID, account.Concurrency)
		cancel()
		if err != nil {
			continue
		}

		statusCode := resp.StatusCode
		_ = resp.Body.Close()

		// A 404 means the endpoint does not exist; anything else means it's live.
		if statusCode == http.StatusNotFound || statusCode == http.StatusMethodNotAllowed {
			continue
		}

		// Endpoint exists — only show successful detections
		s.sendEvent(c, TestEvent{
			Type:     "endpoint_check",
			Endpoint: ep.Name,
			Status:   "success",
			Text:     fmt.Sprintf("HTTP %d - endpoint detected", statusCode),
		})
		if detectedIdx == -1 {
			detectedIdx = i
		}
		// Prefer the first endpoint that returned a successful (2xx) status
		if successIdx == -1 && statusCode >= 200 && statusCode < 300 {
			successIdx = i
			break // Found a working endpoint, no need to probe further
		}
	}

	// Prefer the endpoint that returned 2xx; fall back to first detected
	finalIdx := successIdx
	if finalIdx == -1 {
		finalIdx = detectedIdx
	}

	if finalIdx == -1 {
		return s.sendErrorAndEnd(c, "No supported API endpoint detected on this base URL")
	}

	detected := autoDetectEndpoints[finalIdx]

	s.sendEvent(c, TestEvent{
		Type: "content",
		Text: "\nRunning streaming test ...\n",
	})

	return s.runStreamingEndpointTest(c, ctx, account, base, authToken, testModelID, detected, proxyURL)
}

// runStreamingEndpointTest sends a full request to the detected endpoint and streams back tokens.
func (s *AccountTestService) runStreamingEndpointTest(
	c *gin.Context,
	ctx context.Context,
	account *Account,
	baseURL, authToken, testModelID string,
	ep endpointDef,
	proxyURL string,
) error {
	epPath := ep.Path
	body, extraHeaders := ep.BuildProbe(testModelID)

	fullURL := baseURL + epPath

	// Handle Gemini path substitution
	if strings.Contains(epPath, "{model}") {
		geminiModel := testModelID
		if m, ok := extraHeaders["_gemini_model"]; ok && m != "" {
			geminiModel = m
			delete(extraHeaders, "_gemini_model")
		}
		if geminiModel == "" {
			geminiModel = "gemini-2.0-flash"
		}
		fullURL = baseURL + strings.Replace(epPath, "{model}", geminiModel, 1)
	}
	delete(extraHeaders, "_gemini_model")

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(body))
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to create request: %s", err.Error()))
	}

	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Set auth header
	if strings.Contains(epPath, "/v1/messages") {
		req.Header.Set("x-api-key", authToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return s.sendErrorAndEnd(c, fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(respBody)))
	}

	// Route to the appropriate stream processor based on endpoint type
	switch {
	case strings.Contains(ep.Path, "/chat/completions"):
		return s.processChatCompletionStream(c, resp.Body)
	case strings.Contains(ep.Path, "/responses"):
		return s.processOpenAIStream(c, resp.Body)
	case strings.Contains(ep.Path, "/v1/messages"):
		return s.processClaudeStream(c, resp.Body)
	case strings.Contains(ep.Path, "generateContent"):
		return s.processGeminiStream(c, resp.Body)
	case strings.Contains(ep.Path, "/rerank"):
		return s.processRerankResponse(c, resp.Body)
	default:
		return s.processGenericStream(c, resp.Body)
	}
}

// processRerankResponse processes a non-streaming JSON response from a rerank API.
func (s *AccountTestService) processRerankResponse(c *gin.Context, body io.Reader) error {
	respBody, err := io.ReadAll(io.LimitReader(body, 64*1024))
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		s.sendEvent(c, TestEvent{Type: "content", Text: string(respBody)})
		s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
		return nil
	}

	// Display rerank results
	if results, ok := result["results"].([]any); ok {
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Rerank returned %d results", len(results))})
		for i, r := range results {
			if rMap, ok := r.(map[string]any); ok {
				score, _ := rMap["relevance_score"].(float64)
				s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("  Result %d: score=%.4f", i+1, score)})
			}
		}
	} else {
		s.sendEvent(c, TestEvent{Type: "content", Text: string(respBody)})
	}

	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

// processGenericStream attempts to process a generic SSE stream, forwarding any text content.
func (s *AccountTestService) processGenericStream(c *gin.Context, body io.Reader) error {
	reader := bufio.NewReader(body)
	gotContent := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if !gotContent {
					s.sendEvent(c, TestEvent{Type: "content", Text: "(empty response)"})
				}
				s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
				return nil
			}
			return s.sendErrorAndEnd(c, fmt.Sprintf("Stream read error: %s", err.Error()))
		}

		line = strings.TrimSpace(line)
		if line == "" || !sseDataPrefix.MatchString(line) {
			continue
		}

		jsonStr := sseDataPrefix.ReplaceAllString(line, "")
		if jsonStr == "[DONE]" {
			s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
			return nil
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			// If it's not JSON, send it as raw text
			s.sendEvent(c, TestEvent{Type: "content", Text: jsonStr})
			gotContent = true
			continue
		}

		// Try to extract text from various possible formats
		extracted := extractTextFromGenericResponse(data)
		if extracted != "" {
			s.sendEvent(c, TestEvent{Type: "content", Text: extracted})
			gotContent = true
		}
	}
}

// extractTextFromGenericResponse attempts to extract text content from various API response formats.
func extractTextFromGenericResponse(data map[string]any) string {
	// OpenAI chat completion: choices[0].delta.content
	if choices, ok := data["choices"].([]any); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]any); ok {
			if delta, ok := choice["delta"].(map[string]any); ok {
				if content, ok := delta["content"].(string); ok {
					return content
				}
			}
		}
	}

	// OpenAI responses: delta field
	if delta, ok := data["delta"].(string); ok && delta != "" {
		return delta
	}

	// Claude: delta.text
	if delta, ok := data["delta"].(map[string]any); ok {
		if text, ok := delta["text"].(string); ok {
			return text
		}
	}

	// Gemini: candidates[0].content.parts[0].text
	if candidates, ok := data["candidates"].([]any); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]any); ok {
			if content, ok := candidate["content"].(map[string]any); ok {
				if parts, ok := content["parts"].([]any); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]any); ok {
						if text, ok := part["text"].(string); ok {
							return text
						}
					}
				}
			}
		}
	}

	// Direct text field
	if text, ok := data["text"].(string); ok {
		return text
	}
	// Direct content field
	if content, ok := data["content"].(string); ok {
		return content
	}

	return ""
}

// UpstreamModel represents a model from upstream API
type UpstreamModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
}

// FetchUpstreamModels fetches available models from upstream API for the given account
func (s *AccountTestService) FetchUpstreamModels(ctx context.Context, accountID int64) ([]UpstreamModel, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if account.Type != "apikey" && account.Type != "upstream" {
		return nil, errors.New("only apikey and upstream accounts support fetching upstream models")
	}

	creds := account.Credentials
	if creds == nil {
		return nil, errors.New("account has no credentials")
	}

	// Create context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	switch account.Platform {
	case "openai", "deepseek", "qwen", "glm", "kimi", "iflow":
		return s.fetchOpenAICompatibleModels(fetchCtx, creds)
	case "gemini":
		return s.fetchGeminiModels(fetchCtx, account, creds)
	case "anthropic":
		return s.fetchAnthropicModels(fetchCtx, creds)
	case "antigravity":
		return s.fetchAntigravityModels(fetchCtx, account)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", account.Platform)
	}
}

func (s *AccountTestService) fetchOpenAICompatibleModels(ctx context.Context, creds map[string]any) ([]UpstreamModel, error) {
	baseURL, _ := creds["base_url"].(string)
	apiKey, _ := creds["api_key"].(string)

	if apiKey == "" {
		return nil, errors.New("api_key is required")
	}

	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	modelsURL := baseURL + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpUpstream.Do(req, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		return nil, fmt.Errorf("upstream returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]UpstreamModel, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, UpstreamModel{ID: m.ID})
	}

	return models, nil
}

func (s *AccountTestService) fetchGeminiModels(ctx context.Context, account *Account, creds map[string]any) ([]UpstreamModel, error) {
	baseURL, _ := creds["base_url"].(string)
	apiKey, _ := creds["api_key"].(string)

	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	modelsURL := baseURL + "/v1beta/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Check if using OAuth
	if apiKey == "" && s.geminiTokenProvider != nil {
		token, err := s.geminiTokenProvider.GetAccessToken(ctx, account)
		if err == nil && token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	} else if apiKey != "" {
		req.Header.Set("x-goog-api-key", apiKey)
	}

	resp, err := s.httpUpstream.Do(req, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		return nil, fmt.Errorf("upstream returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]UpstreamModel, 0, len(result.Models))
	for _, m := range result.Models {
		// Remove "models/" prefix if present
		id := strings.TrimPrefix(m.Name, "models/")
		models = append(models, UpstreamModel{
			ID:          id,
			DisplayName: m.DisplayName,
		})
	}

	return models, nil
}

func (s *AccountTestService) fetchAnthropicModels(ctx context.Context, creds map[string]any) ([]UpstreamModel, error) {
	baseURL, _ := creds["base_url"].(string)
	apiKey, _ := creds["api_key"].(string)

	if apiKey == "" {
		return nil, errors.New("api_key is required")
	}

	// If no base_url, return default Claude models
	if baseURL == "" {
		models := make([]UpstreamModel, 0, len(claude.DefaultModels))
		for _, m := range claude.DefaultModels {
			models = append(models, UpstreamModel{ID: m.ID})
		}
		return models, nil
	}

	// Try to fetch from custom base_url
	baseURL = strings.TrimSuffix(baseURL, "/")
	modelsURL := baseURL + "/v1/models"

	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpUpstream.Do(req, "", 0, 0)
	if err != nil {
		// Fallback to default models on error
		models := make([]UpstreamModel, 0, len(claude.DefaultModels))
		for _, m := range claude.DefaultModels {
			models = append(models, UpstreamModel{ID: m.ID})
		}
		return models, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to default models
		models := make([]UpstreamModel, 0, len(claude.DefaultModels))
		for _, m := range claude.DefaultModels {
			models = append(models, UpstreamModel{ID: m.ID})
		}
		return models, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]UpstreamModel, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, UpstreamModel{ID: m.ID})
	}

	return models, nil
}

func (s *AccountTestService) fetchAntigravityModels(ctx context.Context, account *Account) ([]UpstreamModel, error) {
	accessToken := account.GetCredential("access_token")
	if accessToken == "" {
		return nil, errors.New("access_token is required for Antigravity")
	}

	projectID := account.GetCredential("project_id")

	client, err := antigravity.NewClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to create antigravity client: %w", err)
	}
	modelsResp, _, err := client.FetchAvailableModels(ctx, accessToken, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch antigravity models: %w", err)
	}

	if modelsResp == nil || modelsResp.Models == nil {
		return []UpstreamModel{}, nil
	}

	models := make([]UpstreamModel, 0, len(modelsResp.Models))
	for modelName := range modelsResp.Models {
		if modelName != "" {
			models = append(models, UpstreamModel{
				ID: modelName,
			})
		}
	}

	return models, nil
}
