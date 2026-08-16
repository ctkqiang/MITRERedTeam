---
alwaysApply: true
scene: security_structure
---

# MITRERedTeam 安全与内存规范

本文件约束 `cmd/`、`internal/`、`tools/` 下所有代码的安全性与内存使用，含 cgo 场景。与 `dir_structure.md`（架构与目录）、`code_structure.md`（命名与注释）配套使用：本文档深化安全与内存维度，不复述另外两份文档的条款。

## 1. 适用范围与定位

- 适用于所有 Go 代码，以及通过 `import "C"` 引入的 C 代码。
- 底线原则：
  - 应用自身不得引入可利用缺陷；测试工具的安全缺陷会直接放大攻击风险。
  - 执行外部工具必须受控：限定授权范围、设置超时、不静默扩大目标。
  - 安全关键逻辑必须在注释中说明"为什么"（中文注释遵循 `code_structure.md`）。

## 2. 内存管理规范

### 硬性约束

运行期常驻内存（RSS）峰值不得超过 **500MB**。这是验收指标，任何实现不得以"运行时不检查"规避。

### 运行时配置

- 设置 `GOMEMLIMIT=480MiB` 作为运行时软限制，为 500MB 目标留出 20MiB 头寸；`GOGC` 按场景调整，默认保持 runtime 内置策略。
- 软限制可经 `configs/redteam.json` 配置，但上限不得高于 500MB。

### 编码要求

**大输入必须流式处理。** catalog 数据、工具 stdout 等可能变大的输入，使用 `json.Decoder` 或 `io.Reader` 逐块消费，禁止 `os.ReadFile` 整文件读入后常驻内存。

```go
// 反例：整文件读入，catalog 变大时内存随文件线性增长
rawCatalog, err := os.ReadFile(catalogPath)

// 正例：流式解码，内存占用与单条记录成正比
catalogFile, err := os.Open(catalogPath)
if err != nil {
	return nil, err
}
defer catalogFile.Close()
decoder := json.NewDecoder(catalogFile)
```

**并发执行外部工具有上限。** 用信号量模式限制同时在跑的进程数，防止批量技术检查把本地资源耗尽。

```go
// 最多同时运行 4 个外部工具进程
semaphore := make(chan struct{}, 4)
semaphore <- struct{}{}
defer func() { <-semaphore }()
```

**复用缓冲区。** 高频分配的临时 buffer 走 `sync.Pool`，避免反复触发 GC。

**资源释放。** 文件、网络连接、工具进程必须 `defer` 关闭或 `context` 取消；goroutine 必须能随 context 退出，不得静默常驻。

**禁止无界增长。** 字符串与切片拼接前预估容量；外部输入（URL、header、文件内容）先做长度校验再进入逻辑，超限即报错。

### 强制监控机制

内存约束不止靠约定，代码必须实现运行时监控：

- 启动阶段：Unix 平台通过 `syscall.Setrlimit` 应用 `RLIMIT_AS` 硬上限；容器环境下读取 cgroup 内存限制作为软阈值。
- 运行阶段：监控协程定期采样 `runtime.ReadMemStats`。连续 N 次超过阈值先输出告警，仍不回落则主动退出，返回明确退出码与提示。
- 验证：内存压力测试 + `runtime/pprof` 堆分析作为验收手段；`go test -race` 覆盖并发路径。

```go
// 监控协程示意：每 5 秒采样一次，连续 3 次超限则退出
func watchMemory(ctx context.Context, limitBytes uint64) {
	var exceedCount int
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			if stats.Sys > limitBytes {
				exceedCount++
				if exceedCount >= 3 {
					fmt.Fprintln(os.Stderr, "内存占用持续超过限制，进程退出")
					os.Exit(1)
				}
				continue
			}
			exceedCount = 0
		}
	}
}
```

## 3. 漏洞预防规范（Go 侧）

### 缓冲区溢出

Go 对数组与切片有运行时边界检查，但不得把它当成唯一防线。索引访问前先校验长度；`copy` 与 `append` 前确认目标容量；来自外部的大小值不直接作为索引。

```go
// 反例：外部传入的 offset 未校验就索引
responseHeader := toolOutput[offset : offset+length]

// 正例：先校验边界再切片
if offset > len(toolOutput) || length > len(toolOutput)-offset {
	return nil, fmt.Errorf("越界访问: offset=%d length=%d 数据长度=%d", offset, length, len(toolOutput))
}
responseHeader := toolOutput[offset : offset+length]
```

### 整数溢出

涉及长度、偏移、大小的算术运算，先检查是否超过 `math.MaxInt`；禁止用 `int` 承接外部大小值后直接参与运算。

### 输入校验

所有外部输入——CLI 参数、配置文件、工具 stdout/stderr、网络数据——必须先校验格式与长度，再进入业务逻辑；非法输入返回明确错误，不静默跳过。

### 命令注入

外部命令一律 `exec.CommandContext(ctx, executable, arguments...)`，参数以切片传递，不做 shell 解释。禁止 `sh -c` 与字符串拼接（承接 `dir_structure.md` 安全条款）。

```go
// 反例：拼接进 shell，工具名或参数含特殊字符时被解释执行
command := "nmap -p " + targetPort + " " + targetHost
output, err := exec.Command("sh", "-c", command).Output()

// 正例：参数切片传递，不经 shell
command := exec.CommandContext(ctx, "nmap", "-p", targetPort, targetHost)
output, err := command.CombinedOutput()
```

### 路径处理

使用 `path/filepath` 的 `Clean`/`Join` 处理路径，拒绝 `..` 逃逸出授权目录。写入文件前确认最终路径仍在预期目录内。

### 拒绝服务防护

外部调用全部带超时与上限：请求体大小、输出行数、运行时长。防止目标或中间链路异常拖垮本地资源。

### 凭据处理

密钥、令牌、会话凭据只存内存；不回显、不落盘、不写日志。日志与输出必须脱敏，扫描结果中的敏感值用占位符替代。

## 4. cgo 使用规范

项目以 Go 为主。`import "C"` 仅允许在 Go 标准库无法满足性能或安全要求时使用，例如调用经过审计的原生加密库、零拷贝协议处理。每个使用点必须在代码注释中说明引入理由；能用 Go 实现的一律用 Go，禁止把 cgo 当便捷手段。

### C 侧内存安全

C 代码是缓冲区溢出的主要风险来源，必须遵守：

- C 分配的内存必须由 C 释放：`C.malloc` 与 `C.free` 严格配对，错误路径也要释放。封装辅助函数并用 `defer` 保证释放。

```go
// 辅助封装：拷贝 Go 字节串给 C 使用，并保证调用方完成后释放
func copyToC(data []byte) *C.char {
	buffer := C.CBytes(data)
	return (*C.char)(buffer)
}

// 使用处必须配对释放，避免内存泄漏：
// defer C.free(unsafe.Pointer(buffer))
```

- 字符串与缓冲区操作禁用 `strcpy`、`strcat`、`sprintf` 等无界函数，改用有界形式并显式传入长度。
- 所有写入 C 缓冲区的数据先校验长度，防止越界写。
- 不得在无检查的情况下解引用外部传入的指针。

### 与 Go 运行时的交互

- C 代码不得跨调用持有 Go 指针：`cgo` 调用期间 C 侧保存的 Go 指针在返回后无效，违反 runtime 规则。需要跨调用使用的数据用 `C.CBytes` 拷贝，使用后释放。
- C 调用可能长时间占用 OS 线程：长耗时 C 调用需评估并发模型与线程上限，避免线程饥饿。
- 构建限制：`CGO_ENABLED=0` 时 cgo 代码无法编译，涉及 cgo 的包需在文档中说明跨平台构建策略。

### 审查要求

每处 cgo 必须经过代码审查，PR 中说明安全理由与替代方案评估。无审查记录的 cgo 不得合入。

## 5. 安全编码清单（提交前自查）

- 所有输入均经校验；外部命令经 `exec.CommandContext` 并带超时。
- 无 shell 拼接、无路径逃逸、无凭据落盘或写入日志。
- 无整文件读入大输入、无无界增长；文件/连接/进程均释放；并发外部进程数量有限。
- 内存监控机制已接入（RLIMIT_AS / 采样 + 超限处理）。
- 关键安全逻辑有中文注释说明"为什么"。
- 用户界面文本提供中英双语：CLI 输出按 `--lang` 切换或中英并列，不得只输出单语。

## 6. 验证与门禁

| 类型 | 手段 | 覆盖 |
|---|---|---|
| 静态 | `go vet`、`golangci-lint`（启用 `gosec`、`govet`、`staticcheck`、`unused`） | 常见缺陷模式 |
| 动态 | `go test -race ./...` | 竞态与数据竞争 |
| 模糊 | `go test -fuzz`（解析类逻辑） | 异常输入下的崩溃与越界 |
| 内存 | `runtime/pprof` 堆分析 + 压力测试 | RSS < 500MB 验收 |
| 依赖 | `govulncheck` | 已知漏洞依赖 |

`.githooks/pre-commit` 已覆盖构建、测试与 lint；安全专项检查（race、fuzz、govulncheck、内存验收）纳入 CI 阶段执行，不随提交门禁阻塞日常开发。
