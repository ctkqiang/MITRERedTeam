package tools

import (
	"context"
	"time"
)

// DefaultToolTimeout 是工具适配器未显式指定超时时的单次执行上限。
const DefaultToolTimeout = 60 * time.Second

// Adapter 是所有工具适配器的公共基座，持有统一执行器并暴露委托方法。
// 具体适配器嵌入本类型后，只需声明自己的业务方法并组装工具参数，
// 无需重复维护 runner 字段与执行逻辑。
type Adapter struct {
	runner *Runner
}

// NewAdapter 创建公共执行基座。executablePath 为可执行文件路径；
// timeout 小于等于零时回退到 DefaultToolTimeout，避免误传 0 导致立即超时。
func NewAdapter(executablePath string, timeout time.Duration) *Adapter {
	if timeout <= 0 {
		timeout = DefaultToolTimeout
	}
	return &Adapter{runner: NewRunner(executablePath, timeout)}
}

// Run 委托统一执行器执行参数切片，不做 shell 解释。
func (a *Adapter) Run(ctx context.Context, arguments []string) (*Result, error) {
	return a.runner.Run(ctx, arguments)
}
