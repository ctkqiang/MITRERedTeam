package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mitre_red_team/internal/llm"
	"mitre_red_team/internal/model"
)

// 验证缺少 API Key 时返回 ErrMissingCredential，并指明缺失的环境变量。
func TestNewClientMissingAPIKey(t *testing.T) {
	client, err := llm.NewClient(llm.Config{Provider: model.DeepSeek, Model: "deepseek-v4-flash"})
	if err == nil {
		t.Fatal("期望缺少凭据报错，实际无错误")
	}
	if client != nil {
		t.Error("失败时不应返回客户端")
	}
	var missing *model.ErrMissingCredential
	if !errors.As(err, &missing) {
		t.Fatalf("期望 ErrMissingCredential，实际 %T", err)
	}
	if missing.Variable != "DEEPSEEK_API_KEY" {
		t.Errorf("期望提示 DEEPSEEK_API_KEY，实际 %s", missing.Variable)
	}
}

// 验证无默认模型且未提供模型名时返回错误（OpenRouter 需显式配置模型）。
func TestNewClientMissingModel(t *testing.T) {
	_, err := llm.NewClient(llm.Config{Provider: model.OpenRouter, APIKey: "sk-test"})
	if err == nil {
		t.Fatal("期望缺少模型配置报错，实际无错误")
	}
	var missing *model.ErrMissingCredential
	if !errors.As(err, &missing) {
		t.Fatalf("期望 ErrMissingCredential，实际 %T", err)
	}
	if missing.Variable != "OPENROUTER_MODEL" {
		t.Errorf("期望提示 OPENROUTER_MODEL，实际 %s", missing.Variable)
	}
}

// 验证未知供应商返回 ErrUnknownProvider。
func TestNewClientUnknownProvider(t *testing.T) {
	_, err := llm.NewClient(llm.Config{Provider: model.LLMModel("gemini"), APIKey: "key"})
	if !errors.Is(err, model.ErrUnknownProvider) {
		t.Errorf("期望 ErrUnknownProvider，实际 %v", err)
	}
}

// 验证 OpenAI 兼容客户端请求构造与响应解析。
func TestOpenAICompatComplete(t *testing.T) {
	var receivedMethod, receivedPath, receivedAuth string
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedMethod = request.Method
		receivedPath = request.URL.Path
		receivedAuth = request.Header.Get("Authorization")
		if err := json.NewDecoder(request.Body).Decode(&receivedBody); err != nil {
			t.Errorf("解析请求体失败: %v", err)
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"mock 回复"}}]}`))
	}))
	defer server.Close()

	client, err := llm.NewClient(llm.Config{
		Provider: model.Openai,
		APIKey:   "sk-test",
		Model:    "gpt-4o-mini",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	content, err := client.Complete(context.Background(), "system 提示", "user 问题")
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	if content != "mock 回复" {
		t.Errorf("期望 mock 回复，实际 %q", content)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("期望 POST，实际 %s", receivedMethod)
	}
	if receivedPath != "/chat/completions" {
		t.Errorf("期望路径 /chat/completions，实际 %s", receivedPath)
	}
	if receivedAuth != "Bearer sk-test" {
		t.Errorf("期望 Bearer 鉴权，实际 %q", receivedAuth)
	}
	if receivedBody["model"] != "gpt-4o-mini" {
		t.Errorf("请求体 model 不符，实际 %v", receivedBody["model"])
	}
	messages, ok := receivedBody["messages"].([]interface{})
	if !ok || len(messages) != 2 {
		t.Fatalf("期望 2 条消息，实际 %v", receivedBody["messages"])
	}
}

// 验证非成功状态码时返回 ErrHTTPStatus，并携带服务端错误描述。
func TestOpenAICompatHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte(`{"error":{"message":"invalid api key"}}`))
	}))
	defer server.Close()

	client, err := llm.NewClient(llm.Config{
		Provider: model.Kimi,
		APIKey:   "bad-key",
		Model:    "kimi-k3",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	_, err = client.Complete(context.Background(), "system", "user")
	var statusErr *llm.ErrHTTPStatus
	if !errors.As(err, &statusErr) {
		t.Fatalf("期望 ErrHTTPStatus，实际 %T", err)
	}
	if statusErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("期望状态 401，实际 %d", statusErr.StatusCode)
	}
	if !strings.Contains(statusErr.Body, "invalid api key") {
		t.Errorf("期望携带服务端错误描述，实际 %q", statusErr.Body)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("错误文本应包含状态码，实际: %v", err)
	}
}

// 验证 FromEnv 从环境变量读取凭据与模型名创建客户端。
func TestFromEnv(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "sk-env-key")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4-flash")
	client, err := llm.FromEnv(model.DeepSeek, nil, nil)
	if err != nil {
		t.Fatalf("从环境变量创建客户端失败: %v", err)
	}
	if client.Provider() != model.DeepSeek {
		t.Errorf("期望 Provider=deepseek，实际 %s", client.Provider())
	}
}

// 验证 FromEnv 在凭据缺失时返回 ErrMissingCredential。
func TestFromEnvMissingCredential(t *testing.T) {
	t.Setenv("MOONSHOT_API_KEY", "")
	if _, err := llm.FromEnv(model.Kimi, nil, nil); err == nil {
		t.Fatal("期望凭据缺失报错，实际无错误")
	}
}

// 验证空 choices 返回 ErrEmptyResponse。
func TestOpenAICompatEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client, err := llm.NewClient(llm.Config{
		Provider: model.ByteDance,
		APIKey:   "ark-key",
		Model:    "doubao-test",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	if _, err := client.Complete(context.Background(), "system", "user"); !errors.Is(err, llm.ErrEmptyResponse) {
		t.Errorf("期望 ErrEmptyResponse，实际 %v", err)
	}
}

// 验证 Anthropic 客户端使用 x-api-key 鉴权并解析 content 文本块。
func TestAnthropicComplete(t *testing.T) {
	var receivedAuth, receivedVersion string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedAuth = request.Header.Get("x-api-key")
		receivedVersion = request.Header.Get("anthropic-version")
		_, _ = writer.Write([]byte(`{"content":[{"type":"text","text":"claude 回复"}]}`))
	}))
	defer server.Close()

	client, err := llm.NewClient(llm.Config{
		Provider: model.Anthropic,
		APIKey:   "anthropic-key",
		Model:    "claude-test",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	content, err := client.Complete(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	if content != "claude 回复" {
		t.Errorf("期望 claude 回复，实际 %q", content)
	}
	if receivedAuth != "anthropic-key" {
		t.Errorf("期望 x-api-key 鉴权，实际 %q", receivedAuth)
	}
	if receivedVersion == "" {
		t.Error("期望携带 anthropic-version 头")
	}
}

// 验证 AvailableProviders 只返回凭据已配置的供应商。
func TestAvailableProviders(t *testing.T) {
	lookup := func(name string) string {
		if name == "DEEPSEEK_API_KEY" || name == "ARK_API_KEY" {
			return "configured"
		}
		return ""
	}
	available := llm.AvailableProviders(lookup)
	if len(available) != 2 {
		t.Fatalf("期望 2 家已配置供应商，实际 %v", available)
	}
	for _, provider := range available {
		if provider != model.DeepSeek && provider != model.ByteDance {
			t.Errorf("不应包含未配置的供应商 %s", provider)
		}
	}
}

// 验证 Provider 方法返回客户端对应的供应商。
func TestClientProvider(t *testing.T) {
	client, err := llm.NewClient(llm.Config{
		Provider: model.DeepSeek,
		APIKey:   "sk-test",
		Model:    "deepseek-v4-flash",
		BaseURL:  httptest.NewServer(http.NotFoundHandler()).URL,
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	if client.Provider() != model.DeepSeek {
		t.Errorf("期望 Provider=deepseek，实际 %s", client.Provider())
	}
}
