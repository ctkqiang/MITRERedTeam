package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"mitre_red_team/internal/model"
	"mitre_red_team/internal/utilities"
)

// anthropicAPIVersion 是 Anthropic Messages API 要求的版本头。
const anthropicAPIVersion = "2023-06-01"

// anthropicMaxTokens 单次回复的最大生成 token 数。
const anthropicMaxTokens = 4096

// anthropicMessage 是 Messages API 中的单条消息。
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest 是 Anthropic Messages API 的请求体。
// system 提示词单独成字段，不占用消息序列。
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

// anthropicContentBlock 是响应内容块，文本内容存放在 Text 字段。
type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicResponse 是 Anthropic Messages API 的响应体。
type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// anthropicClient 实现 Anthropic Messages API 客户端。
// 鉴权使用 x-api-key 头与 anthropic-version 版本头，与 OpenAI 兼容系列不同。
type anthropicClient struct {
	provider   model.LLMModel
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *utilities.Logger
}

// Provider 返回当前客户端对应的供应商。
func (c *anthropicClient) Provider() model.LLMModel {
	return c.provider
}

// Complete 发送对话并返回模型回复文本。
// 响应 content 为文本块数组，按顺序拼接后返回。
func (c *anthropicClient) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	payload := anthropicRequest{
		Model:     c.model,
		MaxTokens: anthropicMaxTokens,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化请求体: %w", err)
	}

	endpoint := strings.TrimSuffix(c.baseURL, "/") + "/v1/messages"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("构造请求: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", anthropicAPIVersion)

	c.logger.Debug("llm.request", nil, fmt.Sprintf("provider=%s model=%s endpoint=%s", c.provider, c.model, endpoint))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("调用 %s: %w", c.provider, err)
	}
	defer response.Body.Close()

	var completion anthropicResponse
	if err := json.NewDecoder(response.Body).Decode(&completion); err != nil {
		return "", fmt.Errorf("解析 %s 响应: %w", c.provider, err)
	}
	if response.StatusCode != http.StatusOK {
		message := ""
		if completion.Error != nil {
			message = completion.Error.Message
		}
		return "", &ErrHTTPStatus{StatusCode: response.StatusCode, Body: message}
	}

	var output strings.Builder
	for _, block := range completion.Content {
		if block.Type == "text" {
			output.WriteString(block.Text)
		}
	}
	if strings.TrimSpace(output.String()) == "" {
		return "", ErrEmptyResponse
	}
	return output.String(), nil
}
