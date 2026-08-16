package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mitre_red_team/internal/utilities"
)

// 验证 Telegram 通知器成功发送：请求路径与负载正确。
func TestTelegramNotifierSendsMessage(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		rawBody, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(rawBody, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := newTestLogger()
	notifier := utilities.NewTelegramNotifier("test-token", "12345", server.URL, server.Client(), logger)
	if err := notifier.Send(context.Background(), "扫描完成"); err != nil {
		t.Fatalf("发送失败: %v", err)
	}
	if capturedPath != "/bottest-token/sendMessage" {
		t.Errorf("请求路径不符，实际 %s", capturedPath)
	}
	if capturedBody["chat_id"] != "12345" || capturedBody["text"] != "扫描完成" {
		t.Errorf("请求负载不符，实际 %v", capturedBody)
	}
}

// 验证 Telegram 通知器在接口返回非 2xx 时返回错误。
func TestTelegramNotifierDeliveryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	logger, _ := newTestLogger()
	notifier := utilities.NewTelegramNotifier("test-token", "12345", server.URL, server.Client(), logger)
	if err := notifier.Send(context.Background(), "扫描完成"); err == nil {
		t.Fatal("期望发送失败返回错误，实际无错误")
	}
}

// 验证 OpenClaw 通知器成功发送：请求路径、Authorization 头与负载正确。
func TestOpenClawNotifierSendsMessage(t *testing.T) {
	var capturedPath string
	var capturedAuth string
	var capturedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")
		rawBody, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(rawBody, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := newTestLogger()
	notifier := utilities.NewOpenClawNotifier(server.URL, "gateway-token", "openclaw-weixin", "me", server.Client(), logger)
	if err := notifier.Send(context.Background(), "扫描完成"); err != nil {
		t.Fatalf("发送失败: %v", err)
	}
	if capturedPath != "/api/message" {
		t.Errorf("请求路径不符，实际 %s", capturedPath)
	}
	if capturedAuth != "Bearer gateway-token" {
		t.Errorf("Authorization 头不符，实际 %q", capturedAuth)
	}
	if capturedBody["channel"] != "openclaw-weixin" || capturedBody["to"] != "me" || capturedBody["message"] != "扫描完成" {
		t.Errorf("请求负载不符，实际 %v", capturedBody)
	}
}

// 验证 OpenClaw 通知器在 gateway 返回非 2xx 时返回错误。
func TestOpenClawNotifierDeliveryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gateway error", http.StatusBadGateway)
	}))
	defer server.Close()

	logger, _ := newTestLogger()
	notifier := utilities.NewOpenClawNotifier(server.URL, "", "openclaw-weixin", "me", server.Client(), logger)
	if err := notifier.Send(context.Background(), "扫描完成"); err == nil {
		t.Fatal("期望发送失败返回错误，实际无错误")
	}
}

// 验证平台工厂的分支：none 返回空通知器、未知平台报错、telegram 缺凭据报错。
func TestNewNotifierPlatforms(t *testing.T) {
	logger, _ := newTestLogger()

	notifier, err := utilities.NewNotifier("none", logger)
	if err != nil {
		t.Fatalf("none 平台不应报错: %v", err)
	}
	if _, ok := notifier.(utilities.NoopNotifier); !ok {
		t.Errorf("期望 NoopNotifier，实际 %T", notifier)
	}

	if _, err := utilities.NewNotifier("slack", logger); err == nil {
		t.Error("未知平台应报错")
	}

	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	if _, err := utilities.NewNotifier("telegram", logger); err == nil {
		t.Error("telegram 缺凭据应报错")
	}

	t.Setenv("OPENCLAW_WECHAT_TO", "")
	if _, err := utilities.NewNotifier("wechat", logger); err == nil {
		t.Error("wechat 缺接收方应报错")
	}
}

// 验证 .env 文件解析：注释忽略、引号剥离、空值处理。
func TestLoadDotenv(t *testing.T) {
	directory := t.TempDir()
	envPath := filepath.Join(directory, ".env")
	content := "# 注释行\nFOO=bar\nEMPTY=\nQUOTED=\"hello world\"\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入环境文件失败: %v", err)
	}

	// 先清除目标变量，确保加载逻辑执行写入
	_ = os.Unsetenv("FOO")
	_ = os.Unsetenv("QUOTED")
	if err := utilities.LoadDotenv(envPath); err != nil {
		t.Fatalf("加载环境文件失败: %v", err)
	}
	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO 期望 bar，实际 %q", got)
	}
	if got := os.Getenv("QUOTED"); got != "hello world" {
		t.Errorf("QUOTED 期望 hello world，实际 %q", got)
	}
	_ = os.Unsetenv("FOO")
	_ = os.Unsetenv("QUOTED")
}

// 验证已存在的环境变量不会被 .env 文件覆盖。
func TestLoadDotenvKeepsExistingValues(t *testing.T) {
	directory := t.TempDir()
	envPath := filepath.Join(directory, ".env")
	if err := os.WriteFile(envPath, []byte("EXISTING_KEY=new-value\n"), 0600); err != nil {
		t.Fatalf("写入环境文件失败: %v", err)
	}

	t.Setenv("EXISTING_KEY", "original-value")
	if err := utilities.LoadDotenv(envPath); err != nil {
		t.Fatalf("加载环境文件失败: %v", err)
	}
	if got := os.Getenv("EXISTING_KEY"); got != "original-value" {
		t.Errorf("已存在的变量不应被覆盖，实际 %q", got)
	}
}

// 验证 .env 文件缺失时视为无配置，不报错。
func TestLoadDotenvMissingFile(t *testing.T) {
	if err := utilities.LoadDotenv(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Errorf("缺失文件应视为无配置，实际报错: %v", err)
	}
}
