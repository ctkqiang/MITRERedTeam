//go:build !darwin && !windows && !linux

package utilities

import (
	"context"
	"fmt"
	"runtime"
)

// newPlatformNotifier 返回报告平台不支持的错误实现。
func newPlatformNotifier(logger *Logger) LocalNotifier {
	return &unsupportedNotifier{goos: runtime.GOOS, logger: logger}
}

// unsupportedNotifier 用于当前系统不支持本地通知的场景。
type unsupportedNotifier struct {
	goos   string
	logger *Logger
}

// Show 返回平台不支持的错误。
func (n *unsupportedNotifier) Show(context.Context, LocalNotification) error {
	return fmt.Errorf("当前系统 %s 不支持本地通知", n.goos)
}
