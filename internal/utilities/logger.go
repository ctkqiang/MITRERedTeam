package utilities

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel 定义日志级别。
type LogLevel string

const (
	Info     LogLevel = "信息"
	Debug    LogLevel = "调试"
	Warn     LogLevel = "警告"
	Fatal    LogLevel = "致命"
	Security LogLevel = "安全"
	Trace    LogLevel = "追踪"
	Error    LogLevel = "错误"
)

// MemoryStat 记录单个变量的内存占用估算。
// Go 运行时无法直接获取任意变量的实际占用，由调用方基于 len/cap 或 unsafe.Sizeof 估算后传入。
type MemoryStat struct {
	Name  string // 变量名
	Bytes uint64 // 估算字节数
}

// Logger 输出单行结构化日志。
// 每行依次包含时间戳、级别、操作名、变量内存统计与执行描述，方便用户快速定位事件。
type Logger struct {
	writer io.Writer
	mu     sync.Mutex
}

// New 创建日志器。writer 为空时回退到 os.Stderr。
func New(writer io.Writer) *Logger {
	if writer == nil {
		writer = os.Stderr
	}
	return &Logger{writer: writer}
}

// Default 是包级默认日志器，输出到 stderr。
var Default = New(os.Stderr)

// Log 输出一行日志。
// operation 指明被调用的操作或函数；memory 是待统计的变量内存占用列表；
// description 说明执行期间发生了什么。输出示例：
// 2026-08-16T10:15:30Z [INFO] op=DirectoryEnumeration mem[x]=1.0KiB mem[y]=4.0KiB desc=完成目录扫描
func (l *Logger) Log(level LogLevel, operation string, memory []MemoryStat, description string) {
	var line strings.Builder
	line.WriteString(time.Now().UTC().Format(time.RFC3339))
	line.WriteString(" [")
	line.WriteString(string(level))
	line.WriteString("] op=")
	line.WriteString(operation)

	for _, stat := range memory {
		line.WriteString(" mem[")
		line.WriteString(stat.Name)
		line.WriteString("]=")
		line.WriteString(formatBytes(stat.Bytes))
	}

	line.WriteString(" desc=")
	line.WriteString(sanitize(description))
	line.WriteString("\n")

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.writer, line.String())
}

// Info 记录信息级日志。
func (l *Logger) Info(operation string, memory []MemoryStat, description string) {
	l.Log(Info, operation, memory, description)
}

// Debug 记录调试级日志。
func (l *Logger) Debug(operation string, memory []MemoryStat, description string) {
	l.Log(Debug, operation, memory, description)
}

// Warn 记录警告级日志。
func (l *Logger) Warn(operation string, memory []MemoryStat, description string) {
	l.Log(Warn, operation, memory, description)
}

// Error 记录错误级日志。
func (l *Logger) Error(operation string, memory []MemoryStat, description string) {
	l.Log(Error, operation, memory, description)
}

// Security 记录安全相关事件日志。
func (l *Logger) Security(operation string, memory []MemoryStat, description string) {
	l.Log(Security, operation, memory, description)
}

// Trace 记录追踪级日志。
func (l *Logger) Trace(operation string, memory []MemoryStat, description string) {
	l.Log(Trace, operation, memory, description)
}

// Fatal 记录致命错误并退出进程。
func (l *Logger) Fatal(operation string, memory []MemoryStat, description string) {
	l.Log(Fatal, operation, memory, description)
	os.Exit(1)
}

// formatBytes 把字节数格式化为人类可读单位，如 1536 -> 1.5KiB。
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	divisor, unitName := uint64(unit), "KiB"

	switch {
	case bytes >= unit*unit*unit:
		divisor, unitName = unit*unit*unit, "GiB"
	case bytes >= unit*unit:
		divisor, unitName = unit*unit, "MiB"
	}

	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(divisor), unitName)
}

// sanitize 把描述中的换行与回车替换为空格，保证日志始终为单行。
func sanitize(description string) string {
	description = strings.ReplaceAll(description, "\n", " ")
	return strings.ReplaceAll(description, "\r", " ")
}
