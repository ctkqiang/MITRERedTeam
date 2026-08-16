//go:build linux

package utilities

import (
	"context"
	"fmt"
	"time"

	"mitre_red_team/tools"
)

// linuxNotifier 通过 notify-send 触发桌面通知。
type linuxNotifier struct {
	runner *tools.Runner
	logger *Logger
}

// newPlatformNotifier 创建 Linux 通知器。
func newPlatformNotifier(logger *Logger) LocalNotifier {
	return &linuxNotifier{
		runner: tools.NewRunner("/usr/bin/notify-send", 10*time.Second),
		logger: logger,
	}
}

// Show 显示一条 Linux 桌面通知。
func (n *linuxNotifier) Show(ctx context.Context, notification LocalNotification) error {
	arguments := BuildLinuxNotificationArgs(notification)
	result, err := n.runner.Run(ctx, arguments)
	if err != nil {
		n.logger.Error("LocalNotification", nil, "Linux 通知失败: "+err.Error())
		return fmt.Errorf("Linux 通知失败: %w", err)
	}
	if result.ExitCode != 0 {
		n.logger.Error("LocalNotification", nil, fmt.Sprintf("Linux 通知失败，退出码 %d", result.ExitCode))
		return fmt.Errorf("Linux 通知失败，notify-send 退出码 %d: %s", result.ExitCode, result.Stderr)
	}
	n.logger.Info("LocalNotification", nil, "Linux 通知已显示")
	return nil
}
