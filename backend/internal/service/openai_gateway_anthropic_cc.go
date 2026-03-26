package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// anthropicChatCompletionsAPIURL is the default Anthropic Messages API URL
// used when the account does not specify a custom base_url.
const anthropicChatCompletionsAPIURL = "https://api.anthropic.com/v1/messages"

// ForwardChatCompletionsViaAnthropic accepts a Chat Completions request body,
// converts it to Anthropic Messages API format, forwards to the Anthropic
// upstream, and converts the response back to Chat Completions format.
func (s *OpenAIGatewayService) ForwardChatCompletionsViaAnthropic(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	promptCacheKey string,
	defaultMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	// 1. Parse Chat Completions request
	var chatReq apicompat.ChatCompletionsRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		return nil, fmt.Errorf("parse chat completions request: %w", err)
	}
	originalModel := chatReq.Model
	clientStream := chatReq.Stream
	includeUsage := chatReq.StreamOptions != nil && chatReq.StreamOptions.IncludeUsage

	// 2. Convert to Anthropic Messages format
	anthropicReq, err := apicompat.ChatCompletionsToAnthropic(&chatReq)
	if err != nil {
		return nil, fmt.Errorf("convert chat completions to anthropic: %w", err)
	}

	// 3. Model mapping
	mappedModel := resolveOpenAIForwardModel(account, originalModel, defaultMappedModel)
	anthropicReq.Model = mappedModel

	logger.L().Debug("openai chat_completions via anthropic: model mapping applied",
		zap.Int64("account_id", account.ID),
		zap.String("original_model", originalModel),
		zap.String("mapped_model", mappedModel),
		zap.Bool("stream", clientStream),
	)

	// 4. Marshal Anthropic request body
	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	// 5. Get API key
	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic account missing api_key credential")
	}

	// 6. Build upstream request
	targetURL := anthropicChatCompletionsAPIURL
	baseURL := account.GetBaseURL()
	if baseURL != "" {
		validated, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base_url: %w", err)
		}
		targetURL = validated + "/v1/messages"
	}

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(anthropicBody))
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}

	// 7. Set headers
	upstreamReq.Header.Set("x-api-key", apiKey)
	upstreamReq.Header.Set("content-type", "application/json")
	upstreamReq.Header.Set("anthropic-version", "2023-06-01")

	// 8. Send request
	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			Kind:               "request_error",
			Message:            safeErr,
		})
		writeChatCompletionsError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(startTime).Milliseconds())

	// 9. Handle error response with failover
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			upstreamDetail := ""
			if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
				maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
				if maxBytes <= 0 {
					maxBytes = 2048
				}
				upstreamDetail = truncateString(string(respBody), maxBytes)
			}
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("request-id"),
				Kind:               "failover",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})
			if s.rateLimitService != nil {
				s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody)
			}
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: account.IsPoolMode() && isPoolModeRetryableStatus(resp.StatusCode),
			}
		}
		return s.handleChatCompletionsErrorResponse(resp, c, account)
	}

	// 10. Handle normal response
	var result *OpenAIForwardResult
	var handleErr error
	if clientStream {
		result, handleErr = s.handleAnthropicToChatStreamingResponse(resp, c, originalModel, mappedModel, includeUsage, startTime)
	} else {
		result, handleErr = s.handleAnthropicToChatBufferedResponse(resp, c, originalModel, mappedModel, startTime)
	}

	return result, handleErr
}

// handleAnthropicToChatStreamingResponse reads Anthropic SSE events from upstream,
// converts each to Chat Completions SSE chunks, and writes them to the client.
func (s *OpenAIGatewayService) handleAnthropicToChatStreamingResponse(
	resp *http.Response,
	c *gin.Context,
	originalModel string,
	mappedModel string,
	includeUsage bool,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	requestID := resp.Header.Get("request-id")

	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	state := apicompat.NewAnthropicEventToChatState()
	state.Model = originalModel
	state.IncludeUsage = includeUsage

	var usage OpenAIUsage
	var firstTokenMs *int
	firstChunk := true

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	resultWithUsage := func() *OpenAIForwardResult {
		return &OpenAIForwardResult{
			RequestID:    requestID,
			Usage:        usage,
			Model:        originalModel,
			BillingModel: mappedModel,
			Stream:       true,
			Duration:     time.Since(startTime),
			FirstTokenMs: firstTokenMs,
		}
	}

	// Anthropic SSE format: "event: <type>\ndata: <json>\n\n"
	// We need to track event type lines and pair them with data lines.
	var currentEventType string

	for scanner.Scan() {
		line := scanner.Text()

		// Track event type
		if strings.HasPrefix(line, "event: ") {
			currentEventType = strings.TrimSpace(line[7:])
			continue
		}

		// Skip non-data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := line[6:]
		if payload == "[DONE]" {
			continue
		}

		if firstChunk {
			firstChunk = false
			ms := int(time.Since(startTime).Milliseconds())
			firstTokenMs = &ms
		}

		// Parse the SSE event
		var evt apicompat.AnthropicStreamEvent
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			logger.L().Warn("openai chat_completions via anthropic stream: failed to parse event",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
			continue
		}

		// Use event type from "event:" line if present, otherwise from JSON
		if currentEventType != "" && evt.Type == "" {
			evt.Type = currentEventType
		}
		currentEventType = ""

		// Extract usage from message_start and message_delta events
		if evt.Type == "message_start" && evt.Message != nil {
			usage.InputTokens = evt.Message.Usage.InputTokens
			usage.CacheReadInputTokens = evt.Message.Usage.CacheReadInputTokens
			usage.CacheCreationInputTokens = evt.Message.Usage.CacheCreationInputTokens
		}
		if evt.Type == "message_delta" && evt.Usage != nil {
			usage.OutputTokens = evt.Usage.OutputTokens
		}

		// Convert to Chat Completions chunks
		chunks := apicompat.AnthropicEventToChatChunks(&evt, state)
		for _, chunk := range chunks {
			sse, err := apicompat.ChatChunkToSSE(chunk)
			if err != nil {
				logger.L().Warn("openai chat_completions via anthropic stream: failed to marshal chunk",
					zap.Error(err),
					zap.String("request_id", requestID),
				)
				continue
			}
			if _, err := fmt.Fprint(c.Writer, sse); err != nil {
				logger.L().Info("openai chat_completions via anthropic stream: client disconnected",
					zap.String("request_id", requestID),
				)
				return resultWithUsage(), nil
			}
		}
		if len(chunks) > 0 {
			c.Writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			logger.L().Warn("openai chat_completions via anthropic stream: read error",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
		}
	}

	// Finalize stream
	if finalChunks := apicompat.FinalizeAnthropicChatStream(state); len(finalChunks) > 0 {
		for _, chunk := range finalChunks {
			sse, err := apicompat.ChatChunkToSSE(chunk)
			if err != nil {
				continue
			}
			fmt.Fprint(c.Writer, sse) //nolint:errcheck
		}
	}
	fmt.Fprint(c.Writer, "data: [DONE]\n\n") //nolint:errcheck
	c.Writer.Flush()

	return resultWithUsage(), nil
}

// handleAnthropicToChatBufferedResponse reads all Anthropic SSE events from the
// upstream, assembles a complete response, converts to Chat Completions JSON,
// and writes it to the client.
func (s *OpenAIGatewayService) handleAnthropicToChatBufferedResponse(
	resp *http.Response,
	c *gin.Context,
	originalModel string,
	mappedModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	requestID := resp.Header.Get("request-id")

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	var finalResponse *apicompat.AnthropicResponse
	var usage OpenAIUsage

	var currentEventType string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEventType = strings.TrimSpace(line[7:])
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := line[6:]
		if payload == "[DONE]" {
			continue
		}

		var evt apicompat.AnthropicStreamEvent
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			logger.L().Warn("openai chat_completions via anthropic buffered: failed to parse event",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
			continue
		}

		if currentEventType != "" && evt.Type == "" {
			evt.Type = currentEventType
		}
		currentEventType = ""

		// Build the complete response from streamed events
		switch evt.Type {
		case "message_start":
			if evt.Message != nil {
				finalResponse = evt.Message
				usage.InputTokens = evt.Message.Usage.InputTokens
				usage.CacheReadInputTokens = evt.Message.Usage.CacheReadInputTokens
				usage.CacheCreationInputTokens = evt.Message.Usage.CacheCreationInputTokens
			}
		case "content_block_start":
			if finalResponse != nil && evt.ContentBlock != nil {
				finalResponse.Content = append(finalResponse.Content, *evt.ContentBlock)
			}
		case "content_block_delta":
			if finalResponse != nil && evt.Delta != nil && evt.Index != nil {
				idx := *evt.Index
				if idx < len(finalResponse.Content) {
					switch evt.Delta.Type {
					case "text_delta":
						finalResponse.Content[idx].Text += evt.Delta.Text
					case "thinking_delta":
						finalResponse.Content[idx].Thinking += evt.Delta.Thinking
					case "input_json_delta":
						existing := string(finalResponse.Content[idx].Input)
						if existing == "" || existing == "null" {
							existing = ""
						}
						finalResponse.Content[idx].Input = json.RawMessage(existing + evt.Delta.PartialJSON)
					}
				}
			}
		case "message_delta":
			if finalResponse != nil && evt.Delta != nil {
				finalResponse.StopReason = evt.Delta.StopReason
				finalResponse.StopSequence = evt.Delta.StopSequence
			}
			if evt.Usage != nil {
				usage.OutputTokens = evt.Usage.OutputTokens
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			logger.L().Warn("openai chat_completions via anthropic buffered: read error",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
		}
	}

	if finalResponse == nil {
		writeChatCompletionsError(c, http.StatusBadGateway, "api_error", "Upstream stream ended without a terminal response event")
		return nil, fmt.Errorf("upstream stream ended without terminal event")
	}

	// Convert to Chat Completions response
	chatResp := apicompat.AnthropicToChatCompletions(finalResponse, originalModel)

	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	}
	c.JSON(http.StatusOK, chatResp)

	return &OpenAIForwardResult{
		RequestID:    requestID,
		Usage:        usage,
		Model:        originalModel,
		BillingModel: mappedModel,
		Stream:       false,
		Duration:     time.Since(startTime),
	}, nil
}
