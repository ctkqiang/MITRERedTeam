package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mitre_red_team/internal/model"
	"mitre_red_team/internal/utilities"
)

// Client 是 LLM 对话补全的统一接口，屏蔽底层供应商差异。
type Client interface {
	// Complete 发送 systemPrompt 与 userPrompt，返回模型生成的回复文本。
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
	// Provider 返回当前客户端对应的供应商。
	Provider() model.LLMModel
}

// ErrEmptyResponse 表示供应商接口返回了空内容，无法推进执行。
var ErrEmptyResponse = errors.New("LLM 返回内容为空")

// ErrHTTPStatus 表示供应商接口返回非成功状态码，Body 携带服务端错误描述。
type ErrHTTPStatus struct {
	StatusCode int
	Body       string
}

// Error 返回带状态码与服务端描述的错误信息。
func (e *ErrHTTPStatus) Error() string {
	return fmt.Sprintf("LLM 接口返回状态 %d: %s", e.StatusCode, e.Body)
}

// Config 描述创建 LLM 客户端所需的参数。
// APIKey 与 Model 由调用方从环境变量解析后传入，凭据不落盘、不写入日志。
// BaseURL 用于覆盖供应商默认端点（测试时指向本地 mock 服务）；生产环境留空。
type Config struct {
	Provider   model.LLMModel
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *utilities.Logger
}

// providerSpec 描述单个供应商的接入参数。
type providerSpec struct {
	baseURL      string
	defaultModel string
	anthropic    bool
}

// providerSpecs 汇总六家供应商的接入信息。
// 模型名优先取环境变量，未设置时回退到 defaultModel；两者皆空则要求显式配置。
var providerSpecs = map[model.LLMModel]providerSpec{
	model.Openai: {
		baseURL:      "https://api.openai.com/v1",
		defaultModel: "gpt-4o-mini",
	},
	model.DeepSeek: {
		baseURL:      "https://api.deepseek.com",
		defaultModel: "deepseek-v4-flash",
	},
	model.Kimi: {
		baseURL:      "https://api.moonshot.cn/v1",
		defaultModel: "kimi-k3",
	},
	model.ByteDance: {
		baseURL:      "https://ark.cn-beijing.volces.com/api/v3",
		defaultModel: "doubao-seed-2-1-pro-260628",
	},
	model.OpenRouter: {
		baseURL:      "https://openrouter.ai/api/v1",
		defaultModel: "",
	},
	model.Anthropic: {
		baseURL:      "https://api.anthropic.com",
		defaultModel: "",
		anthropic:    true,
	},
}

// NewClient 按供应商创建客户端，并校验凭据与模型配置是否齐全。
// APIKey 为空时返回 ErrMissingCredential，提示应设置的环境变量名。
func NewClient(cfg Config) (Client, error) {
	if !cfg.Provider.IsKnown() {
		return nil, fmt.Errorf("%w: %s", model.ErrUnknownProvider, cfg.Provider)
	}
	spec, registered := providerSpecs[cfg.Provider]
	if !registered {
		return nil, fmt.Errorf("%w: %s", model.ErrUnknownProvider, cfg.Provider)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, &model.ErrMissingCredential{Provider: cfg.Provider, Variable: cfg.Provider.CredentialVariable()}
	}
	chatModel := strings.TrimSpace(cfg.Model)
	if chatModel == "" {
		chatModel = spec.defaultModel
	}
	if chatModel == "" {
		return nil, &model.ErrMissingCredential{Provider: cfg.Provider, Variable: cfg.Provider.ModelVariable()}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	logger := cfg.Logger
	if logger == nil {
		logger = utilities.Default
	}
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = spec.baseURL
	}

	if spec.anthropic {
		return &anthropicClient{
			provider:   cfg.Provider,
			baseURL:    baseURL,
			apiKey:     cfg.APIKey,
			model:      chatModel,
			httpClient: httpClient,
			logger:     logger,
		}, nil
	}
	return &openAICompatClient{
		provider:   cfg.Provider,
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		model:      chatModel,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// FromEnv 从进程环境变量读取凭据与模型名创建客户端。
// 环境变量名由供应商常量约定，见 model.LLMModel.CredentialVariable 与 ModelVariable。
func FromEnv(provider model.LLMModel, logger *utilities.Logger, httpClient *http.Client) (Client, error) {
	return NewClient(Config{
		Provider:   provider,
		APIKey:     os.Getenv(provider.CredentialVariable()),
		Model:      os.Getenv(provider.ModelVariable()),
		HTTPClient: httpClient,
		Logger:     logger,
	})
}

// AvailableProviders 返回环境变量中已配置凭据的供应商列表。
// 只有凭据齐全的供应商才参与随机选择，未配置的供应商被忽略。
func AvailableProviders(lookup func(string) string) []model.LLMModel {
	var available []model.LLMModel
	for _, provider := range model.AllProviders() {
		if lookup(provider.CredentialVariable()) != "" {
			available = append(available, provider)
		}
	}
	return available
}
