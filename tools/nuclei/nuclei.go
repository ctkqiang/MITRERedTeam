package nuclei

import (
	"context"
	"time"

	"mitre_red_team/tools"
)

// Scanner 封装 nuclei 漏洞检测。
type Scanner struct {
	*tools.Adapter
}

// New 创建 nuclei 扫描器。executablePath 为 nuclei 可执行文件路径，timeout 为单次扫描上限。
func New(executablePath string, timeout time.Duration) *Scanner {
	return &Scanner{Adapter: tools.NewAdapter(executablePath, timeout)}
}

// Scan 对 url 执行漏洞检测。
func (s *Scanner) Scan(ctx context.Context, url string) (*tools.Result, error) {
	return s.Adapter.Run(ctx, []string{"-u", url})
}
