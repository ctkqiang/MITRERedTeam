package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Result 描述一次外部工具执行的结果。
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Succeeded 返回工具是否以零退出码完成，供调用方快速判断成功与否。
func (r *Result) Succeeded() bool {
	return r.ExitCode == 0
}

// Runner 封装外部工具的统一执行入口。
type Runner struct {
	executablePath string
	timeout        time.Duration
}

// NewRunner 创建执行器。executablePath 为可执行文件路径，timeout 为单次执行上限。
func NewRunner(executablePath string, timeout time.Duration) *Runner {
	return &Runner{executablePath: executablePath, timeout: timeout}
}

// Run 执行工具，参数以切片传递，不做 shell 解释。
// 每次执行受 timeout 限制，超时返回错误；非零退出码记录在 Result 中，stderr 不吞。
func (r *Runner) Run(ctx context.Context, arguments []string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	command := exec.CommandContext(ctx, r.executablePath, arguments...)
	var stdoutBuffer, stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer

	err := command.Run()
	var exitCode int
	if err != nil {
		// 超时或取消时进程会被直接杀掉，Wait 返回的 *exec.ExitError 只是"进程被信号终止"，
		// 不能当作工具的正常非零退出处理，否则超时错误会被吞掉、调用方误判为执行成功。
		if ctx.Err() != nil {
			return nil, fmt.Errorf("执行工具 %s 超时或取消: %w", r.executablePath, ctx.Err())
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("执行工具 %s: %w", r.executablePath, err)
		}
	}
	return &Result{Stdout: stdoutBuffer.String(), Stderr: stderrBuffer.String(), ExitCode: exitCode}, nil
}
