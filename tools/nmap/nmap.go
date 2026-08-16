package nmap

import (
	"context"
	"time"

	"mitre_red_team/tools"
)

// Scanner 封装 nmap 端口扫描。
type Scanner struct {
	runner *tools.Runner
}

// New 创建 nmap 扫描器。executablePath 为 nmap 可执行文件路径，timeout 为单次扫描上限。
func New(executablePath string, timeout time.Duration) *Scanner {
	return &Scanner{runner: tools.NewRunner(executablePath, timeout)}
}

// Scan 对 target 执行端口扫描，ports 指定端口或范围，如 "80,443" 或 "1-1000"。
func (s *Scanner) Scan(ctx context.Context, target string, ports string) (*tools.Result, error) {
	return s.runner.Run(ctx, []string{"-p", ports, target})
}
