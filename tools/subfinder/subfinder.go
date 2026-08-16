package subfinder

import (
	"context"
	"time"

	"mitre_red_team/tools"
)

// Enumerator 封装 subfinder 子域名枚举。
type Enumerator struct {
	runner *tools.Runner
}

// New 创建 subfinder 枚举器。executablePath 为 subfinder 可执行文件路径，timeout 为单次执行上限。
func New(executablePath string, timeout time.Duration) *Enumerator {
	return &Enumerator{runner: tools.NewRunner(executablePath, timeout)}
}

// Enumerate 枚举 domain 下的子域名。
func (e *Enumerator) Enumerate(ctx context.Context, domain string) (*tools.Result, error) {
	return e.runner.Run(ctx, []string{"-silent", "-d", domain})
}
