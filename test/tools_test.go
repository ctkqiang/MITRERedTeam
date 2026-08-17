package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"mitre_red_team/tools"
	"mitre_red_team/tools/ffuf"
	"mitre_red_team/tools/httpx"
	"mitre_red_team/tools/nmap"
	"mitre_red_team/tools/sqlmap"
	"mitre_red_team/tools/subfinder"
)

// 验证成功执行时 stdout 内容与退出码。
func TestRunnerSuccessfulExecution(t *testing.T) {
	runner := tools.NewRunner("/bin/echo", 5*time.Second)
	result, err := runner.Run(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("期望 stdout=hello\\n，实际 %q", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("期望退出码 0，实际 %d", result.ExitCode)
	}
}

// 验证参数以切片原样传递，不做 shell 解释。
func TestRunnerPreservesArguments(t *testing.T) {
	runner := tools.NewRunner("/bin/echo", 5*time.Second)
	result, err := runner.Run(context.Background(), []string{"-p", "80", "example.com"})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !strings.Contains(result.Stdout, "-p 80 example.com") {
		t.Errorf("参数未被原样传递，stdout=%q", result.Stdout)
	}
}

// 验证非零退出码与 stderr 捕获。
func TestRunnerNonZeroExit(t *testing.T) {
	runner := tools.NewRunner("/bin/ls", 5*time.Second)
	result, err := runner.Run(context.Background(), []string{"/nonexistent-path-xyz"})
	if err != nil {
		t.Fatalf("非零退出不应视为系统错误: %v", err)
	}
	if result.ExitCode == 0 {
		t.Error("期望非零退出码")
	}
	if result.Stderr == "" {
		t.Error("stderr 不应为空")
	}
}

// 验证超时返回错误。
func TestRunnerTimeout(t *testing.T) {
	runner := tools.NewRunner("/usr/bin/sleep", 100*time.Millisecond)
	if _, err := runner.Run(context.Background(), []string{"5"}); err == nil {
		t.Fatal("期望超时返回错误，实际无错误")
	}
}

// 验证 nmap 适配器的参数构造。
func TestNmapAdapterArguments(t *testing.T) {
	scanner := nmap.New("/bin/echo", 5*time.Second)
	result, err := scanner.Scan(context.Background(), "example.com", "80,443")
	if err != nil {
		t.Fatalf("扫描失败: %v", err)
	}
	if !strings.Contains(result.Stdout, "-p 80,443 example.com") {
		t.Errorf("nmap 参数构造不符，stdout=%q", result.Stdout)
	}
}

// 验证 httpx 适配器的参数构造。
func TestHttpxAdapterArguments(t *testing.T) {
	prober := httpx.New("/bin/echo", 5*time.Second)
	result, err := prober.Probe(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("探测失败: %v", err)
	}
	if !strings.Contains(result.Stdout, "-silent -timeout 5 https://example.com") {
		t.Errorf("httpx 参数构造不符，stdout=%q", result.Stdout)
	}
}

// 验证 subfinder 适配器的参数构造。
func TestSubfinderAdapterArguments(t *testing.T) {
	enumerator := subfinder.New("/bin/echo", 5*time.Second)
	result, err := enumerator.Enumerate(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("枚举失败: %v", err)
	}
	if !strings.Contains(result.Stdout, "-silent -d example.com") {
		t.Errorf("subfinder 参数构造不符，stdout=%q", result.Stdout)
	}
}

// 验证 sqlmap 适配器的参数构造。
func TestSqlmapAdapterArguments(t *testing.T) {
	injector := sqlmap.New("/bin/echo", 5*time.Second)
	result, err := injector.Inject(context.Background(), "https://example.com?id=1")
	if err != nil {
		t.Fatalf("注入检测失败: %v", err)
	}
	if !strings.Contains(result.Stdout, "-u https://example.com?id=1 --batch") {
		t.Errorf("sqlmap 参数构造不符，stdout=%q", result.Stdout)
	}
}

// 验证 ffuf 适配器按默认选项构造完整参数。
func TestFfufAdapterArguments(t *testing.T) {
	fuzzer := ffuf.New("/bin/echo", 5*time.Second)
	result, err := fuzzer.Fuzz(context.Background(), "https://example.com/FUZZ", "/tmp/words.txt", ffuf.DefaultFuzzOptions())
	if err != nil {
		t.Fatalf("枚举失败: %v", err)
	}
	fragments := []string{
		"-u https://example.com/FUZZ",
		"-w /tmp/words.txt",
		"-mc 200,301,302,403",
		"-t 20",
		"-timeout 10",
		"-rate 100",
		"-maxtime 50",
		"-s",
	}
	for _, fragment := range fragments {
		if !strings.Contains(result.Stdout, fragment) {
			t.Errorf("ffuf 参数缺少 %q，stdout=%q", fragment, result.Stdout)
		}
	}
}

// 验证零值 FuzzOptions 不产生多余参数，仅保留必填的 -u 与 -w。
func TestFfufAdapterZeroOptions(t *testing.T) {
	fuzzer := ffuf.New("/bin/echo", 5*time.Second)
	result, err := fuzzer.Fuzz(context.Background(), "https://example.com/FUZZ", "/tmp/words.txt", ffuf.FuzzOptions{})
	if err != nil {
		t.Fatalf("枚举失败: %v", err)
	}
	if result.Stdout != "-u https://example.com/FUZZ -w /tmp/words.txt\n" {
		t.Errorf("零值选项应只保留必填参数，stdout=%q", result.Stdout)
	}
}

// 验证 Succeeded 按退出码判断执行成败。
func TestResultSucceeded(t *testing.T) {
	if got := (&tools.Result{ExitCode: 0}).Succeeded(); !got {
		t.Error("退出码 0 应视为成功")
	}
	if got := (&tools.Result{ExitCode: 1}).Succeeded(); got {
		t.Error("非零退出码不应视为成功")
	}
}
