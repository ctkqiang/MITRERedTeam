package utilities

import (
	"fmt"
	"strconv"
)

// BuildDarwinNotificationScript 生成 osascript 的 AppleScript 文本。
// 文本经 strconv.Quote 转义，作为单个参数传给 osascript，不做 shell 解释。
func BuildDarwinNotificationScript(notification LocalNotification) string {
	script := fmt.Sprintf("display notification %s with title %s",
		strconv.Quote(notification.Body), strconv.Quote(notification.Title))
	if notification.Sound != "" {
		script += fmt.Sprintf(" sound name %s", strconv.Quote(notification.Sound))
	}
	return script
}

// BuildLinuxNotificationArgs 构造 notify-send 的参数列表。
// 优先级映射：1->low、2->normal、3->critical；声音经 sound-name 提示传入。
func BuildLinuxNotificationArgs(notification LocalNotification) []string {
	var arguments []string
	if notification.Icon != "" {
		arguments = append(arguments, "-i", notification.Icon)
	}
	switch notification.Priority {
	case 1:
		arguments = append(arguments, "-u", "low")
	case 2:
		arguments = append(arguments, "-u", "normal")
	case 3:
		arguments = append(arguments, "-u", "critical")
	}
	if notification.Sound != "" {
		arguments = append(arguments, "-h", "string:sound-name:"+notification.Sound)
	}
	return append(arguments, notification.Title, notification.Body)
}

// BuildWindowsNotificationScript 生成 PowerShell 弹窗脚本。
// 使用 WScript.Shell 的 Popup，图标类型按优先级映射：默认->信息、低->警告、紧急->错误。
func BuildWindowsNotificationScript(notification LocalNotification) string {
	iconType := 64
	switch notification.Priority {
	case 1:
		iconType = 48
	case 3:
		iconType = 16
	}
	title := strconv.Quote(notification.Title)
	body := strconv.Quote(notification.Body)
	return fmt.Sprintf("(New-Object -ComObject Wscript.Shell).Popup(%s, 0, %s, %d)", body, title, iconType)
}
