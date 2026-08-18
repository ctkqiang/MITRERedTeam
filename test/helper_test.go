package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// 伪命令环境变量常量：测试二进制被当作外部工具执行时，
// 通过这两个环境变量指定行为模式，避免依赖 Unix 特有的 /bin/echo 等路径。
const (
	helperEnabledEnv  = "GO_MRT_HELPER_PROCESS"
	helperBehaviorEnv = "GO_MRT_HELPER_BEHAVIOR"

	behaviorEcho  = "echo"
	behaviorTrue  = "true"
	behaviorFail  = "fail"
	behaviorSleep = "sleep"
)

// TestMain 拦截伪命令模式：当测试二进制以"外部工具"身份被 exec 拉起时，
// 按环境变量指定的行为运行后立即退出，不进入常规测试调度，避免递归执行全部测试。
func TestMain(m *testing.M) {
	if os.Getenv(helperEnabledEnv) == "1" {
		runHelperProcess()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runHelperProcess 按行为模式执行伪命令并退出。
// echo 原样输出参数（与 /bin/echo 一致）；true 零退出；fail 输出 stderr 后以 1 退出；
// sleep 长时间阻塞，供超时测试在指定期限后由调用方终止。
func runHelperProcess() {
	switch os.Getenv(helperBehaviorEnv) {
	case behaviorEcho:
		fmt.Fprintln(os.Stdout, strings.Join(os.Args[1:], " "))
		os.Exit(0)
	case behaviorTrue:
		os.Exit(0)
	case behaviorFail:
		fmt.Fprintln(os.Stderr, "模拟工具执行失败")
		os.Exit(1)
	case behaviorSleep:
		time.Sleep(time.Hour)
		os.Exit(0)
	default:
		fmt.Fprintln(os.Stderr, "未知的伪命令行为模式")
		os.Exit(1)
	}
}

// fakeToolPath 返回一个可跨平台执行的"伪命令"路径（即当前测试二进制），
// 并配置好伪命令模式所需的环境变量。调用方把返回值当作外部工具的可执行文件路径使用。
func fakeToolPath(t *testing.T, behavior string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("获取测试二进制路径失败: %v", err)
	}
	t.Setenv(helperEnabledEnv, "1")
	t.Setenv(helperBehaviorEnv, behavior)
	return executable
}
