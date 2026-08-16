//go:build darwin

package utilities

import (
	"context"
	"fmt"
	"time"

	"mitre_red_team/tools"
)

// darwinNotifier 通过 osascript 触发 macOS 系统通知。
type darwinNotifier struct {
	runner *tools.Runner
	logger *Logger
}

// newPlatformNotifier 创建 macOS 通知器。
func newPlatformNotifier(logger *Logger) LocalNotifier {
	return &darwinNotifier{
		runner: tools.NewRunner("/usr/bin/osascript", 10*time.Second),
		logger: logger,
	}
}

// Show 显示一条 macOS 通知。
func (n *darwinNotifier) Show(ctx context.Context, notification LocalNotification) error {
	script := BuildDarwinNotificationScript(notification)
	result, err := n.runner.Run(ctx, []string{"-e", script})
	if err != nil {
		n.logger.Error("LocalNotification", nil, "macOS 通知失败: "+err.Error())
		return fmt.Errorf("macOS 通知失败: %w", err)
	}
	if result.ExitCode != 0 {
		n.logger.Error("LocalNotification", nil, fmt.Sprintf("macOS 通知失败，退出码 %d", result.ExitCode))
		return fmt.Errorf("macOS 通知失败，osascript 退出码 %d: %s", result.ExitCode, result.Stderr)
	}
	n.logger.Info("LocalNotification", nil, "macOS 通知已显示")
	return nil
}
