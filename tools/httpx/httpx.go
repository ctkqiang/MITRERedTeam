package httpx

import (
	"context"
	"time"

	"mitre_red_team/tools"
)

// Prober 封装 httpx HTTP 存活探测。
type Prober struct {
	runner *tools.Runner
}

// New 创建 httpx 探测器。executablePath 为 httpx 可执行文件路径，timeout 为单次执行上限。
func New(executablePath string, timeout time.Duration) *Prober {
	return &Prober{runner: tools.NewRunner(executablePath, timeout)}
}

// Probe 探测 url 的存活状态，输出匹配的响应摘要。
func (p *Prober) Probe(ctx context.Context, url string) (*tools.Result, error) {
	return p.runner.Run(ctx, []string{"-silent", "-timeout", "5", url})
}
