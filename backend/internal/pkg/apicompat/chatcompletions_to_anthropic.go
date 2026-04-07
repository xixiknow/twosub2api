package apicompat

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatCompletionsToAnthropic converts a Chat Completions request into an
// Anthropic Messages API request. The converted request always uses stream=true
// so that the caller can handle both streaming and buffered responses uniformly.
func ChatCompletionsToAnthropic(req *ChatCompletionsRequest) (*AnthropicRequest, error) {
	out := &AnthropicRequest{
		Model:       req.Model,
		Stream:      true,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// max_tokens (Claude requires this field; default to 8192)
	maxTokens := 8192
	if req.MaxCompletionTokens != nil && *req.MaxCompletionTokens > 0 {
		maxTokens = *req.MaxCompletionTokens
	} else if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}
	out.MaxTokens = maxTokens

	// stop → stop_sequences
	if len(req.Stop) > 0 {
		out.StopSeqs = parseChatStopSequences(req.Stop)
	}

	// messages → system + messages
	system, messages, err := convertChatMessagesToAnthropic(req.Messages)
	if err != nil {
		return nil, err
	}
	if system != nil {
		out.System = system
	}
	out.Messages = messages

	// tools
	if len(req.Tools) > 0 {
		out.Tools = convertChatToolsToAnthropic(req.Tools)
	}

	// tool_choice
	if len(req.ToolChoice) > 0 {
		tc, err := convertChatToolChoiceToAnthropic(req.ToolChoice)
		if err != nil {
			return nil, fmt.Errorf("convert tool_choice: %w", err)
		}
		out.ToolChoice = tc
	}

	return out, nil
}

// parseChatStopSequences parses the stop field which can be a string or []string.
func parseChatStopSequences(raw json.RawMessage) []string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s != "" {
			return []string{s}
		}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	return nil
}

// convertChatMessagesToAnthropic splits Chat Completions messages into an
// optional system prompt and a sequence of Anthropic messages.
func convertChatMessagesToAnthropic(msgs []ChatMessage) (json.RawMessage, []AnthropicMessage, error) {
	var systemParts []string
	var anthropicMsgs []AnthropicMessage

	for _, m := range msgs {
		switch m.Role {
		case "system":
			text, err := parseChatContentFlexible(m.Content)
			if err != nil {
				return nil, nil, fmt.Errorf("parse system content: %w", err)
			}
			if text != "" {
				systemParts = append(systemParts, text)
			}

		case "user":
			blocks, err := chatUserContentToAnthropicBlocks(m.Content)
			if err != nil {
				return nil, nil, fmt.Errorf("parse user content: %w", err)
			}
			content, err := json.Marshal(blocks)
			if err != nil {
				return nil, nil, err
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    "user",
				Content: content,
			})

		case "assistant":
			blocks, err := chatAssistantContentToAnthropicBlocks(m)
			if err != nil {
				return nil, nil, fmt.Errorf("parse assistant content: %w", err)
			}
			content, err := json.Marshal(blocks)
			if err != nil {
				return nil, nil, err
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    "assistant",
				Content: content,
			})

		case "tool":
			block := chatToolResultToAnthropicBlock(m)
			content, err := json.Marshal([]AnthropicContentBlock{block})
			if err != nil {
				return nil, nil, err
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    "user",
				Content: content,
			})

		default:
			// Treat unknown roles as user messages
			text, _ := parseChatContentFlexible(m.Content)
			if text != "" {
				block := AnthropicContentBlock{Type: "text", Text: text}
				content, _ := json.Marshal([]AnthropicContentBlock{block})
				anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
					Role:    "user",
					Content: content,
				})
			}
		}
	}

	// Merge consecutive messages with the same role (Anthropic requires alternating)
	anthropicMsgs = mergeConsecutiveSameRoleMessages(anthropicMsgs)

	var systemJSON json.RawMessage
	if len(systemParts) > 0 {
		combined := strings.Join(systemParts, "\n\n")
		systemJSON, _ = json.Marshal(combined)
	}

	return systemJSON, anthropicMsgs, nil
}

// mergeConsecutiveSameRoleMessages merges consecutive messages with the same
// role into a single message by concatenating their content blocks. Anthropic
// requires strictly alternating user/assistant messages.
func mergeConsecutiveSameRoleMessages(msgs []AnthropicMessage) []AnthropicMessage {
	if len(msgs) <= 1 {
		return msgs
	}
	var merged []AnthropicMessage
	for _, msg := range msgs {
		if len(merged) > 0 && merged[len(merged)-1].Role == msg.Role {
			// Merge content arrays
			var existingBlocks []AnthropicContentBlock
			_ = json.Unmarshal(merged[len(merged)-1].Content, &existingBlocks)
			var newBlocks []AnthropicContentBlock
			_ = json.Unmarshal(msg.Content, &newBlocks)
			existingBlocks = append(existingBlocks, newBlocks...)
			merged[len(merged)-1].Content, _ = json.Marshal(existingBlocks)
		} else {
			merged = append(merged, msg)
		}
	}
	return merged
}

// parseChatContentFlexible parses content that can be a string or array of content parts.
func parseChatContentFlexible(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	// Try string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	// Try array of content parts
	var parts []ChatContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", fmt.Errorf("parse content: %w", err)
	}
	var texts []string
	for _, p := range parts {
		if p.Type == "text" && p.Text != "" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, ""), nil
}

// chatUserContentToAnthropicBlocks converts user message content (string or
// multimodal array) to Anthropic content blocks.
func chatUserContentToAnthropicBlocks(raw json.RawMessage) ([]AnthropicContentBlock, error) {
	if len(raw) == 0 {
		return []AnthropicContentBlock{{Type: "text", Text: ""}}, nil
	}

	// Try plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return []AnthropicContentBlock{{Type: "text", Text: s}}, nil
	}

	// Multi-modal array
	var parts []ChatContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return nil, fmt.Errorf("parse user content parts: %w", err)
	}

	var blocks []AnthropicContentBlock
	for _, p := range parts {
		switch p.Type {
		case "text":
			if p.Text != "" {
				blocks = append(blocks, AnthropicContentBlock{Type: "text", Text: p.Text})
			}
		case "image_url":
			if p.ImageURL != nil && p.ImageURL.URL != "" {
				block, err := imageURLToAnthropicBlock(p.ImageURL.URL)
				if err != nil {
					continue // skip invalid images
				}
				blocks = append(blocks, block)
			}
		}
	}
	if len(blocks) == 0 {
		blocks = append(blocks, AnthropicContentBlock{Type: "text", Text: ""})
	}
	return blocks, nil
}

// imageURLToAnthropicBlock converts a data URI or URL to an Anthropic image block.
func imageURLToAnthropicBlock(url string) (AnthropicContentBlock, error) {
	// Handle data URIs: data:image/png;base64,<data>
	if strings.HasPrefix(url, "data:") {
		parts := strings.SplitN(url, ",", 2)
		if len(parts) != 2 {
			return AnthropicContentBlock{}, fmt.Errorf("invalid data URI")
		}
		// Parse media type from "data:image/png;base64"
		header := parts[0] // "data:image/png;base64"
		header = strings.TrimPrefix(header, "data:")
		mediaType := strings.SplitN(header, ";", 2)[0]

		return AnthropicContentBlock{
			Type: "image",
			Source: &AnthropicImageSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      parts[1],
			},
		}, nil
	}

	// For regular URLs, Anthropic doesn't support URL-type images in the
	// same AnthropicImageSource struct. Fall back to including it as text.
	return AnthropicContentBlock{
		Type: "text",
		Text: fmt.Sprintf("[Image: %s]", url),
	}, nil
}

// chatAssistantContentToAnthropicBlocks converts an assistant message to
// Anthropic content blocks, including text and tool_use blocks.
func chatAssistantContentToAnthropicBlocks(m ChatMessage) ([]AnthropicContentBlock, error) {
	var blocks []AnthropicContentBlock

	// Parse text content
	if len(m.Content) > 0 {
		text, err := parseChatContentFlexible(m.Content)
		if err != nil {
			return nil, err
		}
		if text != "" {
			blocks = append(blocks, AnthropicContentBlock{Type: "text", Text: text})
		}
	}

	// Convert tool_calls to tool_use blocks
	for _, tc := range m.ToolCalls {
		var input json.RawMessage
		if tc.Function.Arguments != "" {
			input = json.RawMessage(tc.Function.Arguments)
		} else {
			input = json.RawMessage("{}")
		}
		blocks = append(blocks, AnthropicContentBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: input,
		})
	}

	if len(blocks) == 0 {
		blocks = append(blocks, AnthropicContentBlock{Type: "text", Text: ""})
	}
	return blocks, nil
}

// chatToolResultToAnthropicBlock converts a tool result message to an
// Anthropic tool_result content block.
func chatToolResultToAnthropicBlock(m ChatMessage) AnthropicContentBlock {
	output, _ := parseChatContentFlexible(m.Content)
	var contentRaw json.RawMessage
	if output != "" {
		contentRaw, _ = json.Marshal(output)
	}
	return AnthropicContentBlock{
		Type:      "tool_result",
		ToolUseID: m.ToolCallID,
		Content:   contentRaw,
	}
}

// convertChatToolsToAnthropic converts Chat Completions tool definitions to
// Anthropic tool definitions.
func convertChatToolsToAnthropic(tools []ChatTool) []AnthropicTool {
	var out []AnthropicTool
	for _, t := range tools {
		if t.Type != "function" || t.Function == nil {
			continue
		}
		at := AnthropicTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
		out = append(out, at)
	}
	return out
}

// convertChatToolChoiceToAnthropic maps Chat Completions tool_choice values to
// Anthropic tool_choice format.
//
//	"auto" → {"type":"auto"}
//	"none" → {"type":"none"}  (as {"type":"any"} with disable_parallel_tool_use)
//	"required" → {"type":"any"}
//	{"type":"function","function":{"name":"X"}} → {"type":"tool","name":"X"}
func convertChatToolChoiceToAnthropic(raw json.RawMessage) (json.RawMessage, error) {
	// Try string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		switch s {
		case "auto":
			return json.Marshal(map[string]string{"type": "auto"})
		case "none":
			// Anthropic doesn't have "none" - omit tool_choice (caller should remove tools)
			return nil, nil
		case "required":
			return json.Marshal(map[string]string{"type": "any"})
		default:
			return json.Marshal(map[string]string{"type": "auto"})
		}
	}

	// Object form: {"type":"function","function":{"name":"X"}}
	var obj struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if obj.Function.Name != "" {
		return json.Marshal(map[string]string{
			"type": "tool",
			"name": obj.Function.Name,
		})
	}

	return json.Marshal(map[string]string{"type": "auto"})
}
