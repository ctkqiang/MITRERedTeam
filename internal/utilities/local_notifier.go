package utilities

import (
	"context"
	"time"
)

// LocalNotification 描述一条系统原生通知。
// Priority 取值：0 默认、1 低、2 高、3 紧急；Icon 与 Sound 仅在平台支持时生效。
type LocalNotification struct {
	Title    string
	Body     string
	Icon     string
	Priority int
	Sound    string
}

// LocalNotifier 显示系统原生通知的统一接口。
type LocalNotifier interface {
	// Show 立即显示一条通知。
	Show(ctx context.Context, notification LocalNotification) error
}

// NewLocalNotifier 按当前操作系统创建本地通知器。
// 平台实现由带 build tag 的文件按编译目标提供；不支持的平台返回可报告错误的实现。
func NewLocalNotifier(logger *Logger) LocalNotifier {
	return newPlatformNotifier(logger)
}

// Schedule 在指定延迟后显示通知，可被 ctx 取消。
func Schedule(ctx context.Context, notifier LocalNotifier, notification LocalNotification, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return notifier.Show(ctx, notification)
	}
}
