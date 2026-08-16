package model

import (
	"errors"
	"fmt"
	"strings"
)

// LLMModel 标识一个受支持的 LLM 供应商。
type LLMModel string

// 受支持的 LLM 供应商。值为官方/社区通行的供应商标识。
const (
	ByteDance  LLMModel = "doubao"
	DeepSeek   LLMModel = "deepseek"
	Kimi       LLMModel = "kimi"
	OpenRouter LLMModel = "openrouter"
	Anthropic  LLMModel = "anthropic"
	Openai     LLMModel = "openai"
)

// DeepSeekModel 标识 DeepSeek 提供的具体模型。
// 历史常量保留以兼容既有调用；官方 API 当前使用 deepseek-v4-flash 与 deepseek-v4-pro。
type DeepSeekModel string

const (
	DEEPSEEK_FLASH_V4 DeepSeekModel = "deepseek-flash-v4"
	DEEPSEEK_PRO_V4   DeepSeekModel = "deepseek-pro-v4"
)

// ErrUnknownProvider 表示给定的字符串无法映射到受支持的 LLM 供应商。
var ErrUnknownProvider = errors.New("未知的 LLM 供应商")

// ErrMissingCredential 表示创建客户端时缺少供应商凭据或模型配置。
type ErrMissingCredential struct {
	Provider LLMModel
	Variable string
}

// Error 返回带供应商与缺失环境变量名的错误信息，方便用户直接定位配置项。
func (e *ErrMissingCredential) Error() string {
	return fmt.Sprintf("LLM 供应商 %s 缺少配置：请设置环境变量 %s", e.Provider, e.Variable)
}

// allProviders 固定声明顺序，保证随机选择与列表展示的可预测性。
var allProviders = []LLMModel{Openai, DeepSeek, Kimi, ByteDance, OpenRouter, Anthropic}

// AllProviders 返回全部受支持的 LLM 供应商。
// 返回副本，调用方修改不会影响内部声明。
func AllProviders() []LLMModel {
	return append([]LLMModel(nil), allProviders...)
}

// ParseLLMModel 把字符串解析为 LLM 供应商，无法识别时返回 ErrUnknownProvider。
// 解析忽略大小写与首尾空白，兼容 "OpenAI" 这类写法。
func ParseLLMModel(value string) (LLMModel, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case string(Openai):
		return Openai, nil
	case string(DeepSeek):
		return DeepSeek, nil
	case string(Kimi):
		return Kimi, nil
	case string(ByteDance):
		return ByteDance, nil
	case string(OpenRouter):
		return OpenRouter, nil
	case string(Anthropic):
		return Anthropic, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnknownProvider, value)
	}
}

// String 返回供应商的规范名称。
func (m LLMModel) String() string {
	return string(m)
}

// IsKnown 判断供应商是否受支持。
func (m LLMModel) IsKnown() bool {
	for _, supported := range allProviders {
		if m == supported {
			return true
		}
	}
	return false
}

// CredentialVariable 返回该供应商 API 凭据对应的环境变量名。
// 未知供应商返回空串，调用方应先用 IsKnown 校验。
func (m LLMModel) CredentialVariable() string {
	switch m {
	case Openai:
		return "OPENAI_API_KEY"
	case DeepSeek:
		return "DEEPSEEK_API_KEY"
	case Kimi:
		return "MOONSHOT_API_KEY"
	case ByteDance:
		return "ARK_API_KEY"
	case OpenRouter:
		return "OPENROUTER_API_KEY"
	case Anthropic:
		return "ANTHROPIC_API_KEY"
	default:
		return ""
	}
}

// ModelVariable 返回该供应商模型名对应的环境变量名。
// 未知供应商返回空串，调用方应先用 IsKnown 校验。
func (m LLMModel) ModelVariable() string {
	switch m {
	case Openai:
		return "OPENAI_MODEL"
	case DeepSeek:
		return "DEEPSEEK_MODEL"
	case Kimi:
		return "MOONSHOT_MODEL"
	case ByteDance:
		return "ARK_MODEL"
	case OpenRouter:
		return "OPENROUTER_MODEL"
	case Anthropic:
		return "ANTHROPIC_MODEL"
	default:
		return ""
	}
}
