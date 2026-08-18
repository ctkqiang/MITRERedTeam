package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mitre_red_team/internal/config"
)

// 验证加载真实配置文件，工具路径条目完整。
func TestLoadRealConfig(t *testing.T) {
	configuration, err := config.Load(filepath.Join("..", "configs", "redteam.json"))
	if err != nil {
		t.Fatalf("加载真实配置失败: %v", err)
	}
	if configuration.Tools["nmap"] == "" {
		t.Error("工具 nmap 的路径不应为空")
	}
	if len(configuration.Tools) < 6 {
		t.Errorf("期望至少 6 个工具条目，实际 %d", len(configuration.Tools))
	}
}

// 验证配置文件缺失时报错。
func TestLoadMissingFile(t *testing.T) {
	if _, err := config.Load(filepath.Join("..", "configs", "missing.json")); err == nil {
		t.Fatal("期望加载缺失文件报错，实际无错误")
	}
}

// 验证非法 JSON 报错，且错误信息带路径上下文。
func TestLoadInvalidJSON(t *testing.T) {
	tempDirectory := t.TempDir()
	invalidPath := filepath.Join(tempDirectory, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("{ not json"), 0600); err != nil {
		t.Fatalf("写入非法配置失败: %v", err)
	}

	_, err := config.Load(invalidPath)
	if err == nil {
		t.Fatal("期望解析非法 JSON 报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "解析配置文件") {
		t.Errorf("错误信息应包含上下文，实际: %v", err)
	}
}

// 验证真实配置中本地通知默认开启。
func TestLocalNotificationDefaultEnabled(t *testing.T) {
	configuration, err := config.Load(filepath.Join("..", "configs", "redteam.json"))
	if err != nil {
		t.Fatalf("加载真实配置失败: %v", err)
	}
	if !configuration.Notifications.LocalEnabled() {
		t.Error("本地通知默认应开启")
	}
}

// 验证配置缺省（无 local 字段）时本地通知同样视为开启。
func TestLocalNotificationDefaultsWhenAbsent(t *testing.T) {
	configuration := &config.Config{}
	if !configuration.Notifications.LocalEnabled() {
		t.Error("未声明 local 字段时默认应开启")
	}
}

// 验证显式关闭后本地通知返回 false。
func TestLocalNotificationExplicitlyDisabled(t *testing.T) {
	disabled := false
	configuration := &config.Config{
		Notifications: config.NotificationConfig{Local: &disabled},
	}
	if configuration.Notifications.LocalEnabled() {
		t.Error("显式关闭后应返回 false")
	}
}

// 验证全部工具路径可用时无缺失。
func TestCheckToolsAvailableAllPresent(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("获取测试二进制路径失败: %v", err)
	}
	configuration := &config.Config{
		Tools: map[string]string{
			"echo": executable,
			"ls":   executable,
		},
	}
	if missing := configuration.CheckToolsAvailable(); len(missing) != 0 {
		t.Errorf("期望无缺失工具，实际 %v", missing)
	}
}

// 验证存在无效路径时报告对应工具名。
func TestCheckToolsAvailableMissing(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("获取测试二进制路径失败: %v", err)
	}
	ghostPath := filepath.Join(t.TempDir(), "ghost")
	configuration := &config.Config{
		Tools: map[string]string{
			"echo":  executable,
			"ghost": ghostPath,
			"empty": "",
		},
	}
	missing := configuration.CheckToolsAvailable()
	if len(missing) != 2 || missing[0] != "empty" || missing[1] != "ghost" {
		t.Errorf("期望缺失 empty 与 ghost（排序），实际 %v", missing)
	}
}
