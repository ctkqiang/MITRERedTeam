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
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("执行工具 %s: %w", r.executablePath, err)
		}
	}
	return &Result{Stdout: stdoutBuffer.String(), Stderr: stderrBuffer.String(), ExitCode: exitCode}, nil
}
