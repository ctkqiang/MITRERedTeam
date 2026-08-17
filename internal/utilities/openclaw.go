package utilities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// openclawMessageEndpoint 是 OpenClaw gateway 发送消息的默认路径。
const openclawMessageEndpoint = "/api/message"

// OpenClawNotifier 通过本机 OpenClaw gateway 的本地 HTTP API 发送消息。
// 微信账号需先执行 openclaw channels login --channel openclaw-weixin 扫码绑定，
// 消息经 gateway 的长连接管道投递，无需外部 API 凭据。
type OpenClawNotifier struct {
	gatewayURL string
	apiToken   string
	channel    string
	recipient  string
	httpClient *http.Client
	logger     *Logger
}

// NewOpenClawNotifier 创建 OpenClaw 通知器。
// gatewayURL 为 gateway 地址（默认 http://localhost:18789）；apiToken 为空表示本机回环免认证；
// channel 为目标渠道名（如 openclaw-weixin）；recipient 为接收方标识（由 gateway 侧定义）。
func NewOpenClawNotifier(
	gatewayURL string,
	apiToken string,
	channel string,
	recipient string,
	httpClient *http.Client,
	logger *Logger,
) *OpenClawNotifier {
	return &OpenClawNotifier{
		gatewayURL: gatewayURL,
		apiToken:   apiToken,
		channel:    channel,
		recipient:  recipient,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Send 把 message 经 OpenClaw gateway 投递到目标渠道。
// 请求体为 {channel, to, message}；配置了 apiToken 时以 Bearer 头携带。
func (n *OpenClawNotifier) Send(ctx context.Context, message string) error {
	payload := map[string]string{
		"channel": n.channel,
		"to":      n.recipient,
		"message": message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	endpoint := n.gatewayURL + openclawMessageEndpoint
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if n.apiToken != "" {
		request.Header.Set("Authorization", "Bearer "+n.apiToken)
	}

	response, err := n.httpClient.Do(request)
	if err != nil {
		n.logger.Error("OpenClawNotification", nil, "消息发送失败: "+err.Error())
		return fmt.Errorf("发送 OpenClaw 消息失败: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		n.logger.Error("OpenClawNotification", nil, fmt.Sprintf("发送失败，HTTP %d", response.StatusCode))
		return fmt.Errorf("OpenClaw gateway 返回异常状态码 %d: %s", response.StatusCode, string(responseBody))
	}
	n.logger.Info("OpenClawNotification", nil, "消息发送成功")
	return nil
}
