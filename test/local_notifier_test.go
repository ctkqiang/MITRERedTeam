package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"mitre_red_team/internal/utilities"
)

// mockLocalNotifier 是 LocalNotifier 的测试替身。
type mockLocalNotifier func(ctx context.Context, notification utilities.LocalNotification) error

// Show 转发到闭包实现。
func (fn mockLocalNotifier) Show(ctx context.Context, notification utilities.LocalNotification) error {
	return fn(ctx, notification)
}

// 验证 macOS 通知脚本内容与转义。
func TestBuildDarwinNotificationScript(t *testing.T) {
	script := utilities.BuildDarwinNotificationScript(utilities.LocalNotification{
		Title: "扫描完成",
		Body:  "发现 3 个漏洞",
		Sound: "Glass",
	})
	for _, fragment := range []string{"display notification", "扫描完成", "发现 3 个漏洞", "Glass"} {
		if !strings.Contains(script, fragment) {
			t.Errorf("脚本缺少 %q，实际 %s", fragment, script)
		}
	}
}

// 验证 Linux notify-send 参数构造：图标、优先级、声音与标题正文顺序。
func TestBuildLinuxNotificationArgs(t *testing.T) {
	arguments := utilities.BuildLinuxNotificationArgs(utilities.LocalNotification{
		Title:    "扫描完成",
		Body:     "发现 3 个漏洞",
		Icon:     "/tmp/icon.png",
		Priority: 3,
		Sound:    "message-new-instant",
	})
	joined := strings.Join(arguments, " ")
	fragments := []string{
		"-i /tmp/icon.png",
		"-u critical",
		"-h string:sound-name:message-new-instant",
		"扫描完成 发现 3 个漏洞",
	}
	for _, fragment := range fragments {
		if !strings.Contains(joined, fragment) {
			t.Errorf("参数缺少 %q，实际 %v", fragment, arguments)
		}
	}
}

// 验证 Windows PowerShell 弹窗脚本：紧急优先级映射到错误图标。
func TestBuildWindowsNotificationScript(t *testing.T) {
	script := utilities.BuildWindowsNotificationScript(utilities.LocalNotification{
		Title:    "扫描完成",
		Body:     "发现漏洞",
		Priority: 3,
	})
	if !strings.Contains(script, "Popup") || !strings.Contains(script, "16") {
		t.Errorf("脚本内容不符，实际 %s", script)
	}
}

// 验证调度在延迟后显示通知。
func TestScheduleShowsAfterDelay(t *testing.T) {
	var shown bool
	notifier := mockLocalNotifier(func(ctx context.Context, notification utilities.LocalNotification) error {
		shown = true
		return nil
	})

	err := utilities.Schedule(
		context.Background(),
		notifier,
		utilities.LocalNotification{},
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("调度执行失败: %v", err)
	}
	if !shown {
		t.Error("期望延迟后已显示通知")
	}
}

// 验证调度可被 context 取消。
func TestScheduleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	notifier := mockLocalNotifier(func(ctx context.Context, notification utilities.LocalNotification) error {
		t.Error("取消后不应显示通知")
		return nil
	})
	if err := utilities.Schedule(ctx, notifier, utilities.LocalNotification{}, time.Hour); err == nil {
		t.Fatal("期望取消返回错误，实际无错误")
	}
}

// 验证工厂在当前系统返回可用的本地通知器。
func TestNewLocalNotifier(t *testing.T) {
	logger, _ := newTestLogger()
	notifier := utilities.NewLocalNotifier(logger)
	if notifier == nil {
		t.Fatal("工厂不应返回 nil")
	}
}
