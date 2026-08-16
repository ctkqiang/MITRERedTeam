package tests

import (
	"errors"
	"strings"
	"testing"

	"mitre_red_team/internal/model"
)

// 验证六家供应商都能从规范名称解析，且大小写与首尾空白不敏感。
func TestParseLLMModelKnown(t *testing.T) {
	cases := map[string]model.LLMModel{
		"openai":      model.Openai,
		"OpenAI":      model.Openai,
		"  deepseek ": model.DeepSeek,
		"kimi":        model.Kimi,
		"doubao":      model.ByteDance,
		"openrouter":  model.OpenRouter,
		"anthropic":   model.Anthropic,
	}
	for input, want := range cases {
		got, err := model.ParseLLMModel(input)
		if err != nil {
			t.Errorf("解析 %q 不应报错: %v", input, err)
		}
		if got != want {
			t.Errorf("解析 %q 期望 %s，实际 %s", input, want, got)
		}
	}
}

// 验证无法识别的字符串返回 ErrUnknownProvider。
func TestParseLLMModelUnknown(t *testing.T) {
	if _, err := model.ParseLLMModel("gemini"); !errors.Is(err, model.ErrUnknownProvider) {
		t.Errorf("期望 ErrUnknownProvider，实际 %v", err)
	}
	if _, err := model.ParseLLMModel(""); !errors.Is(err, model.ErrUnknownProvider) {
		t.Errorf("空字符串应报未知供应商，实际 %v", err)
	}
}

// 验证 AllProviders 返回全部六家且无重复。
func TestAllProviders(t *testing.T) {
	providers := model.AllProviders()
	if len(providers) != 6 {
		t.Fatalf("期望 6 家供应商，实际 %d", len(providers))
	}
	seen := make(map[model.LLMModel]bool)
	for _, provider := range providers {
		if seen[provider] {
			t.Errorf("供应商 %s 重复出现", provider)
		}
		seen[provider] = true
	}
	// 修改返回副本不应影响后续调用。
	model.AllProviders()[0] = model.Anthropic
	if model.AllProviders()[0] == model.Anthropic {
		t.Error("AllProviders 应返回副本，修改不应影响内部声明")
	}
}

// 验证每个已知供应商 IsKnown 为 true，未知供应商为 false。
func TestLLMModelIsKnown(t *testing.T) {
	for _, provider := range model.AllProviders() {
		if !provider.IsKnown() {
			t.Errorf("供应商 %s 应被识别为已知", provider)
		}
	}
	if model.LLMModel("gemini").IsKnown() {
		t.Error("未知供应商不应被识别为已知")
	}
}

// 验证供应商的 String 返回规范名称。
func TestLLMModelString(t *testing.T) {
	if model.DeepSeek.String() != "deepseek" {
		t.Errorf("DeepSeek.String() 期望 deepseek，实际 %s", model.DeepSeek.String())
	}
}

// 验证六家供应商的凭据环境变量名映射正确。
func TestCredentialVariable(t *testing.T) {
	cases := map[model.LLMModel]string{
		model.Openai:     "OPENAI_API_KEY",
		model.DeepSeek:   "DEEPSEEK_API_KEY",
		model.Kimi:       "MOONSHOT_API_KEY",
		model.ByteDance:  "ARK_API_KEY",
		model.OpenRouter: "OPENROUTER_API_KEY",
		model.Anthropic:  "ANTHROPIC_API_KEY",
	}
	for provider, want := range cases {
		if got := provider.CredentialVariable(); got != want {
			t.Errorf("%s 凭据变量期望 %s，实际 %s", provider, want, got)
		}
	}
	if got := model.LLMModel("unknown").CredentialVariable(); got != "" {
		t.Errorf("未知供应商凭据变量应为空，实际 %s", got)
	}
}

// 验证六家供应商的模型环境变量名映射正确。
func TestModelVariable(t *testing.T) {
	cases := map[model.LLMModel]string{
		model.Openai:     "OPENAI_MODEL",
		model.DeepSeek:   "DEEPSEEK_MODEL",
		model.Kimi:       "MOONSHOT_MODEL",
		model.ByteDance:  "ARK_MODEL",
		model.OpenRouter: "OPENROUTER_MODEL",
		model.Anthropic:  "ANTHROPIC_MODEL",
	}
	for provider, want := range cases {
		if got := provider.ModelVariable(); got != want {
			t.Errorf("%s 模型变量期望 %s，实际 %s", provider, want, got)
		}
	}
	if got := model.LLMModel("unknown").ModelVariable(); got != "" {
		t.Errorf("未知供应商模型变量应为空，实际 %s", got)
	}
}

// 验证 ErrMissingCredential 的错误信息包含供应商与环境变量名。
func TestErrMissingCredentialError(t *testing.T) {
	err := &model.ErrMissingCredential{Provider: model.DeepSeek, Variable: "DEEPSEEK_API_KEY"}
	message := err.Error()
	if !strings.Contains(message, "deepseek") || !strings.Contains(message, "DEEPSEEK_API_KEY") {
		t.Errorf("错误信息应包含供应商与环境变量名，实际: %s", message)
	}
}

// 验证 DeepSeekModel 历史常量保持不变（向后兼容）。
func TestDeepSeekModelConstants(t *testing.T) {
	if model.DEEPSEEK_FLASH_V4 != "deepseek-flash-v4" {
		t.Errorf("DEEPSEEK_FLASH_V4 期望 deepseek-flash-v4，实际 %s", model.DEEPSEEK_FLASH_V4)
	}
	if model.DEEPSEEK_PRO_V4 != "deepseek-pro-v4" {
		t.Errorf("DEEPSEEK_PRO_V4 期望 deepseek-pro-v4，实际 %s", model.DEEPSEEK_PRO_V4)
	}
}
