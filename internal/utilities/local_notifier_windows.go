//go:build windows

package utilities

import (
	"context"
	"fmt"
	"time"

	"mitre_red_team/tools"
)

// windowsNotifier 通过 PowerShell 弹窗显示 Windows 通知。
type windowsNotifier struct {
	runner *tools.Runner
	logger *Logger
}

// newPlatformNotifier 创建 Windows 通知器。
func newPlatformNotifier(logger *Logger) LocalNotifier {
	return &windowsNotifier{
		runner: tools.NewRunner("powershell.exe", 10*time.Second),
		logger: logger,
	}
}

// Show 显示一条 Windows 通知。
func (n *windowsNotifier) Show(ctx context.Context, notification LocalNotification) error {
	script := BuildWindowsNotificationScript(notification)
	result, err := n.runner.Run(ctx, []string{"-NoProfile", "-Command", script})
	if err != nil {
		n.logger.Error("LocalNotification", nil, "Windows 通知失败: "+err.Error())
		return fmt.Errorf("Windows 通知失败: %w", err)
	}
	if result.ExitCode != 0 {
		n.logger.Error("LocalNotification", nil, fmt.Sprintf("Windows 通知失败，退出码 %d", result.ExitCode))
		return fmt.Errorf("Windows 通知失败，PowerShell 退出码 %d: %s", result.ExitCode, result.Stderr)
	}
	n.logger.Info("LocalNotification", nil, "Windows 通知已显示")
	return nil
}
