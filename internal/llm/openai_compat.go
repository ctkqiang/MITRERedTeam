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

// chatMessage 是对话补全中的单条消息。
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionRequest 是 OpenAI 兼容的对话补全请求体。
type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatCompletionResponse 是 OpenAI 兼容的对话补全响应体。
// Choices 为空表示没有生成内容；Error 携带服务端错误描述。
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// openAICompatClient 实现 OpenAI Chat Completions 协议的客户端。
// OpenAI、DeepSeek、Kimi、豆包、OpenRouter 均使用同一协议，仅 baseURL 与模型不同。
type openAICompatClient struct {
	provider   model.LLMModel
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *utilities.Logger
}

// Provider 返回当前客户端对应的供应商。
func (c *openAICompatClient) Provider() model.LLMModel {
	return c.provider
}

// Complete 发送对话并返回模型回复文本。
// 参数以切片与结构体传递，不做任何 shell 解释；凭据只出现在 Authorization 头中。
func (c *openAICompatClient) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	payload := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化请求体: %w", err)
	}

	endpoint := strings.TrimSuffix(c.baseURL, "/") + "/chat/completions"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("构造请求: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	c.logger.Debug("llm.request", nil, fmt.Sprintf("provider=%s model=%s endpoint=%s", c.provider, c.model, endpoint))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("调用 %s: %w", c.provider, err)
	}
	defer response.Body.Close()

	var completion chatCompletionResponse
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
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		return "", ErrEmptyResponse
	}
	return completion.Choices[0].Message.Content, nil
}
