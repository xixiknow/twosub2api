package apicompat

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Non-streaming: AnthropicResponse → ChatCompletionsResponse
// ---------------------------------------------------------------------------

// AnthropicToChatCompletions converts a non-streaming Anthropic Messages API
// response into an OpenAI Chat Completions response.
func AnthropicToChatCompletions(resp *AnthropicResponse, model string) *ChatCompletionsResponse {
	id := resp.ID
	if id == "" {
		id = generateChatCmplID()
	}

	out := &ChatCompletionsResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
	}

	var contentText string
	var reasoningText string
	var toolCalls []ChatToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			contentText += block.Text
		case "thinking":
			reasoningText += block.Thinking
		case "tool_use":
			inputStr := "{}"
			if len(block.Input) > 0 {
				inputStr = string(block.Input)
			}
			toolCalls = append(toolCalls, ChatToolCall{
				ID:   block.ID,
				Type: "function",
				Function: ChatFunctionCall{
					Name:      block.Name,
					Arguments: inputStr,
				},
			})
		}
	}

	msg := ChatMessage{Role: "assistant"}
	if contentText != "" {
		raw, _ := json.Marshal(contentText)
		msg.Content = raw
	}
	if reasoningText != "" {
		msg.ReasoningContent = reasoningText
	}
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	finishReason := anthropicStopReasonToChatFinishReason(resp.StopReason, toolCalls)

	out.Choices = []ChatChoice{{
		Index:        0,
		Message:      msg,
		FinishReason: finishReason,
	}}

	out.Usage = anthropicUsageToChatUsage(&resp.Usage)

	return out
}

func anthropicStopReasonToChatFinishReason(stopReason string, toolCalls []ChatToolCall) string {
	switch stopReason {
	case "end_turn":
		if len(toolCalls) > 0 {
			return "tool_calls"
		}
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}

func anthropicUsageToChatUsage(u *AnthropicUsage) *ChatUsage {
	if u == nil {
		return nil
	}
	usage := &ChatUsage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      u.InputTokens + u.OutputTokens,
	}
	if u.CacheReadInputTokens > 0 || u.CacheCreationInputTokens > 0 {
		usage.PromptTokensDetails = &ChatTokenDetails{
			CachedTokens:             u.CacheReadInputTokens,
			CacheCreationInputTokens: u.CacheCreationInputTokens,
		}
	}
	return usage
}

// ---------------------------------------------------------------------------
// Streaming: AnthropicStreamEvent → []ChatCompletionsChunk (stateful converter)
// ---------------------------------------------------------------------------

// AnthropicEventToChatState tracks state for converting a sequence of
// Anthropic SSE events into Chat Completions SSE chunks.
type AnthropicEventToChatState struct {
	ID      string
	Model   string
	Created int64

	SentRole    bool
	SawToolCall bool
	SawText     bool
	Finalized   bool

	// Tracks content block index → tool call sequential index mapping
	NextToolCallIndex     int
	BlockIndexToToolIndex map[int]int
	CurrentBlockIndex     int
	CurrentBlockType      string

	IncludeUsage bool
	Usage        *ChatUsage
}

// NewAnthropicEventToChatState returns an initialised stream state.
func NewAnthropicEventToChatState() *AnthropicEventToChatState {
	return &AnthropicEventToChatState{
		ID:                    generateChatCmplID(),
		Created:               time.Now().Unix(),
		BlockIndexToToolIndex: make(map[int]int),
	}
}

// AnthropicEventToChatChunks converts a single Anthropic SSE event into zero
// or more Chat Completions chunks, updating state as it goes.
func AnthropicEventToChatChunks(evt *AnthropicStreamEvent, state *AnthropicEventToChatState) []ChatCompletionsChunk {
	switch evt.Type {
	case "message_start":
		return anthropicHandleMessageStart(evt, state)
	case "content_block_start":
		return anthropicHandleContentBlockStart(evt, state)
	case "content_block_delta":
		return anthropicHandleContentBlockDelta(evt, state)
	case "content_block_stop":
		return nil // no output needed
	case "message_delta":
		return anthropicHandleMessageDelta(evt, state)
	case "message_stop":
		return nil // finalization handled by message_delta
	default:
		return nil
	}
}

// FinalizeAnthropicChatStream emits a final chunk with finish_reason if the
// stream ended without a proper message_delta event. It is idempotent.
func FinalizeAnthropicChatStream(state *AnthropicEventToChatState) []ChatCompletionsChunk {
	if state.Finalized {
		return nil
	}
	state.Finalized = true

	finishReason := "stop"
	if state.SawToolCall {
		finishReason = "tool_calls"
	}

	chunks := []ChatCompletionsChunk{makeAnthropicChatFinishChunk(state, finishReason)}

	if state.IncludeUsage && state.Usage != nil {
		chunks = append(chunks, ChatCompletionsChunk{
			ID:      state.ID,
			Object:  "chat.completion.chunk",
			Created: state.Created,
			Model:   state.Model,
			Choices: []ChatChunkChoice{},
			Usage:   state.Usage,
		})
	}

	return chunks
}

// --- internal handlers ---

func anthropicHandleMessageStart(evt *AnthropicStreamEvent, state *AnthropicEventToChatState) []ChatCompletionsChunk {
	if evt.Message != nil {
		if evt.Message.ID != "" {
			state.ID = evt.Message.ID
		}
		if state.Model == "" && evt.Message.Model != "" {
			state.Model = evt.Message.Model
		}
		// Extract usage from message_start (input tokens)
		if state.Usage == nil {
			state.Usage = &ChatUsage{}
		}
		state.Usage.PromptTokens = evt.Message.Usage.InputTokens
		if evt.Message.Usage.CacheReadInputTokens > 0 || evt.Message.Usage.CacheCreationInputTokens > 0 {
			state.Usage.PromptTokensDetails = &ChatTokenDetails{
				CachedTokens:             evt.Message.Usage.CacheReadInputTokens,
				CacheCreationInputTokens: evt.Message.Usage.CacheCreationInputTokens,
			}
		}
	}

	if state.SentRole {
		return nil
	}
	state.SentRole = true

	return []ChatCompletionsChunk{makeAnthropicChatDeltaChunk(state, ChatDelta{Role: "assistant"})}
}

func anthropicHandleContentBlockStart(evt *AnthropicStreamEvent, state *AnthropicEventToChatState) []ChatCompletionsChunk {
	if evt.Index != nil {
		state.CurrentBlockIndex = *evt.Index
	}

	if evt.ContentBlock == nil {
		return nil
	}

	state.CurrentBlockType = evt.ContentBlock.Type

	switch evt.ContentBlock.Type {
	case "tool_use":
		state.SawToolCall = true
		idx := state.NextToolCallIndex
		state.BlockIndexToToolIndex[state.CurrentBlockIndex] = idx
		state.NextToolCallIndex++

		return []ChatCompletionsChunk{makeAnthropicChatDeltaChunk(state, ChatDelta{
			ToolCalls: []ChatToolCall{{
				Index: &idx,
				ID:    evt.ContentBlock.ID,
				Type:  "function",
				Function: ChatFunctionCall{
					Name: evt.ContentBlock.Name,
				},
			}},
		})}
	case "text":
		// No output on text block start
		return nil
	case "thinking":
		// No output on thinking block start
		return nil
	default:
		return nil
	}
}

func anthropicHandleContentBlockDelta(evt *AnthropicStreamEvent, state *AnthropicEventToChatState) []ChatCompletionsChunk {
	if evt.Delta == nil {
		return nil
	}

	switch evt.Delta.Type {
	case "text_delta":
		if evt.Delta.Text == "" {
			return nil
		}
		state.SawText = true
		content := evt.Delta.Text
		return []ChatCompletionsChunk{makeAnthropicChatDeltaChunk(state, ChatDelta{Content: &content})}

	case "thinking_delta":
		if evt.Delta.Thinking == "" {
			return nil
		}
		reasoning := evt.Delta.Thinking
		return []ChatCompletionsChunk{makeAnthropicChatDeltaChunk(state, ChatDelta{ReasoningContent: &reasoning})}

	case "input_json_delta":
		if evt.Delta.PartialJSON == "" {
			return nil
		}
		blockIdx := state.CurrentBlockIndex
		if evt.Index != nil {
			blockIdx = *evt.Index
		}
		idx, ok := state.BlockIndexToToolIndex[blockIdx]
		if !ok {
			return nil
		}
		return []ChatCompletionsChunk{makeAnthropicChatDeltaChunk(state, ChatDelta{
			ToolCalls: []ChatToolCall{{
				Index: &idx,
				Function: ChatFunctionCall{
					Arguments: evt.Delta.PartialJSON,
				},
			}},
		})}

	case "signature_delta":
		// Signatures are not exposed in Chat Completions format
		return nil

	default:
		return nil
	}
}

func anthropicHandleMessageDelta(evt *AnthropicStreamEvent, state *AnthropicEventToChatState) []ChatCompletionsChunk {
	state.Finalized = true

	finishReason := "stop"
	if evt.Delta != nil {
		switch evt.Delta.StopReason {
		case "end_turn":
			if state.SawToolCall {
				finishReason = "tool_calls"
			}
		case "max_tokens":
			finishReason = "length"
		case "tool_use":
			finishReason = "tool_calls"
		case "stop_sequence":
			finishReason = "stop"
		}
	}

	// Update usage from message_delta
	if evt.Usage != nil {
		if state.Usage == nil {
			state.Usage = &ChatUsage{}
		}
		state.Usage.CompletionTokens = evt.Usage.OutputTokens
		state.Usage.TotalTokens = state.Usage.PromptTokens + state.Usage.CompletionTokens
	}

	var chunks []ChatCompletionsChunk
	chunks = append(chunks, makeAnthropicChatFinishChunk(state, finishReason))

	if state.IncludeUsage && state.Usage != nil {
		chunks = append(chunks, ChatCompletionsChunk{
			ID:      state.ID,
			Object:  "chat.completion.chunk",
			Created: state.Created,
			Model:   state.Model,
			Choices: []ChatChunkChoice{},
			Usage:   state.Usage,
		})
	}

	return chunks
}

// --- chunk builders ---

func makeAnthropicChatDeltaChunk(state *AnthropicEventToChatState, delta ChatDelta) ChatCompletionsChunk {
	return ChatCompletionsChunk{
		ID:      state.ID,
		Object:  "chat.completion.chunk",
		Created: state.Created,
		Model:   state.Model,
		Choices: []ChatChunkChoice{{
			Index:        0,
			Delta:        delta,
			FinishReason: nil,
		}},
	}
}

func makeAnthropicChatFinishChunk(state *AnthropicEventToChatState, finishReason string) ChatCompletionsChunk {
	empty := ""
	return ChatCompletionsChunk{
		ID:      state.ID,
		Object:  "chat.completion.chunk",
		Created: state.Created,
		Model:   state.Model,
		Choices: []ChatChunkChoice{{
			Index:        0,
			Delta:        ChatDelta{Content: &empty},
			FinishReason: &finishReason,
		}},
	}
}
