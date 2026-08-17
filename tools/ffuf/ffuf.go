package ffuf

import (
	"context"
	"strconv"
	"time"

	"mitre_red_team/tools"
)

// Fuzzer 封装 ffuf 目录与参数枚举。
type Fuzzer struct {
	*tools.Adapter
}

// New 创建 ffuf 模糊器。executablePath 为 ffuf 可执行文件路径，timeout 为单次执行上限。
func New(executablePath string, timeout time.Duration) *Fuzzer {
	return &Fuzzer{Adapter: tools.NewAdapter(executablePath, timeout)}
}

// FuzzOptions 描述 ffuf 枚举的可选参数，零值字段在构造命令行时被省略。
type FuzzOptions struct {
	MatchCodes string // 命中的 HTTP 状态码列表，如 "200,301,302,403"
	Threads    int    // 并发线程数
	TimeoutSec int    // 单请求超时秒数
	RateLimit  int    // 每秒最大请求数
	MaxTimeSec int    // 最大执行秒数
	Silent     bool   // 静默模式：仅输出命中条目
}

// DefaultFuzzOptions 返回目录枚举常用的推荐参数。
func DefaultFuzzOptions() FuzzOptions {
	return FuzzOptions{
		MatchCodes: "200,301,302,403",
		Threads:    20,
		TimeoutSec: 10,
		RateLimit:  100,
		MaxTimeSec: 50,
		Silent:     true,
	}
}

// Fuzz 对 url 执行枚举，wordlist 指定字典文件路径。
// url 需包含占位符（如 https://example.com/FUZZ），opts 控制命中的状态码、并发等参数。
func (f *Fuzzer) Fuzz(ctx context.Context, url string, wordlist string, opts FuzzOptions) (*tools.Result, error) {
	return f.Adapter.Run(ctx, buildArguments(url, wordlist, opts))
}

// buildArguments 按 opts 组装 ffuf 命令行参数，零值字段不进入命令行。
func buildArguments(url string, wordlist string, opts FuzzOptions) []string {
	arguments := []string{"-u", url, "-w", wordlist}
	if opts.MatchCodes != "" {
		arguments = append(arguments, "-mc", opts.MatchCodes)
	}
	if opts.Threads > 0 {
		arguments = append(arguments, "-t", strconv.Itoa(opts.Threads))
	}
	if opts.TimeoutSec > 0 {
		arguments = append(arguments, "-timeout", strconv.Itoa(opts.TimeoutSec))
	}
	if opts.RateLimit > 0 {
		arguments = append(arguments, "-rate", strconv.Itoa(opts.RateLimit))
	}
	if opts.MaxTimeSec > 0 {
		arguments = append(arguments, "-maxtime", strconv.Itoa(opts.MaxTimeSec))
	}
	if opts.Silent {
		arguments = append(arguments, "-s")
	}
	return arguments
}
