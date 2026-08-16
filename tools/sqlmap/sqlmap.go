package sqlmap

import (
	"context"
	"time"

	"mitre_red_team/tools"
)

// Injector 封装 sqlmap 注入检测。
type Injector struct {
	runner *tools.Runner
}

// New 创建 sqlmap 注入器。executablePath 为 sqlmap 可执行文件路径，timeout 为单次执行上限。
func New(executablePath string, timeout time.Duration) *Injector {
	return &Injector{runner: tools.NewRunner(executablePath, timeout)}
}

// Inject 对 url 执行自动注入检测，--batch 使 sqlmap 按默认选择运行，不等待交互输入。
func (i *Injector) Inject(ctx context.Context, url string) (*tools.Result, error) {
	return i.runner.Run(ctx, []string{"-u", url, "--batch"})
}
