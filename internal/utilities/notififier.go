package utilities

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Notifier 是通知发送的统一接口，屏蔽底层平台差异。
type Notifier interface {
	// Send 发送消息。message 为要送达的文本内容。
	Send(ctx context.Context, message string) error
}

// NoopNotifier 是不执行任何发送的空通知器，用于平台未配置的场景。
type NoopNotifier struct{}

// Send 直接返回 nil，不产生任何副作用。
func (NoopNotifier) Send(context.Context, string) error { return nil }

// NewNotifier 按平台配置创建对应的通知器。
// platform 取值：telegram / wechat / none（或空串）。
// 敏感凭据从环境变量读取，不进入配置仓库，也不出现在任何日志中。
func NewNotifier(platform string, logger *Logger) (Notifier, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	switch platform {
	case "telegram":
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID := os.Getenv("TELEGRAM_CHAT_ID")
		if botToken == "" || chatID == "" {
			return nil, fmt.Errorf("通知平台 telegram 缺少凭据，请设置 TELEGRAM_BOT_TOKEN 与 TELEGRAM_CHAT_ID")
		}
		return NewTelegramNotifier(botToken, chatID, telegramAPIBaseURL, httpClient, logger), nil
	case "wechat":
		gatewayURL := os.Getenv("OPENCLAW_GATEWAY_URL")
		if gatewayURL == "" {
			gatewayURL = "http://localhost:18789"
		}
		apiToken := os.Getenv("OPENCLAW_GATEWAY_TOKEN")
		channel := os.Getenv("OPENCLAW_WECHAT_CHANNEL")
		if channel == "" {
			channel = "openclaw-weixin"
		}
		recipient := os.Getenv("OPENCLAW_WECHAT_TO")
		if recipient == "" {
			return nil, fmt.Errorf("通知平台 wechat 缺少接收方，请设置 OPENCLAW_WECHAT_TO")
		}
		return NewOpenClawNotifier(gatewayURL, apiToken, channel, recipient, httpClient, logger), nil
	case "", "none":
		return NoopNotifier{}, nil
	default:
		return nil, fmt.Errorf("未知的通知平台: %s", platform)
	}
}
