package ffuf

import (
	"context"
	"mitre_red_team/tools"
	"time"
)

// Fuzzer 封装 ffuf 目录与参数枚举。
type Fuzzer struct {
	runner *tools.Runner
}

// New 创建 ffuf 模糊器。executablePath 为 ffuf 可执行文件路径，timeout 为单次执行上限。
func New(executablePath string, timeout time.Duration) *Fuzzer {
	return &Fuzzer{runner: tools.NewRunner(executablePath, timeout)}
}

// Fuzz 对 url 执行枚举，wordlist 指定字典文件路径。
func (f *Fuzzer) Fuzz(ctx context.Context, url string, wordlist string) (*tools.Result, error) {
	return f.runner.Run(ctx, []string{"-u", url, "-w", wordlist})
}
