package utilities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// telegramAPIBaseURL 是 Telegram Bot API 的固定服务地址。
const telegramAPIBaseURL = "https://api.telegram.org"

// TelegramNotifier 通过 Telegram Bot API 发送消息到指定会话。
type TelegramNotifier struct {
	baseURL    string
	botToken   string
	chatID     string
	httpClient *http.Client
	logger     *Logger
}

// NewTelegramNotifier 创建 Telegram 通知器。
// baseURL 生产环境传 telegramAPIBaseURL，测试时可指向本地 mock 服务。
// httpClient 允许注入自定义客户端以便测试与超时控制。
func NewTelegramNotifier(
	botToken string,
	chatID string,
	baseURL string,
	httpClient *http.Client,
	logger *Logger,
) *TelegramNotifier {
	return &TelegramNotifier{
		baseURL:    baseURL,
		botToken:   botToken,
		chatID:     chatID,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Send 把 message 发送到配置的会话。
// 请求体为 {chat_id, text}，经 JSON 编码后 POST 到 /bot<token>/sendMessage。
func (n *TelegramNotifier) Send(ctx context.Context, message string) error {
	payload := map[string]string{
		"chat_id": n.chatID,
		"text":    message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", n.baseURL, n.botToken)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := n.httpClient.Do(request)
	if err != nil {
		n.logger.Error("TelegramNotification", nil, "消息发送失败: "+err.Error())
		return fmt.Errorf("发送 Telegram 消息失败: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		n.logger.Error("TelegramNotification", nil, fmt.Sprintf("发送失败，HTTP %d", response.StatusCode))
		return fmt.Errorf("Telegram API 返回异常状态码 %d: %s", response.StatusCode, string(responseBody))
	}
	n.logger.Info("TelegramNotification", nil, "消息发送成功")
	return nil
}
