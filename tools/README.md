# tools/ 工具适配层

`tools/` 是外部安全工具的统一适配层。技术实现（`internal/technique/`）不允许直接 `exec.Command`，一律经这层调外部工具。本层守 `dir_structure.md`、`code_structure.md`、`security_structure.md` 的规矩：`exec.CommandContext` + 超时、参数切片传递、工具路径可配置、stderr 不吞。

## 目录结构

```
tools/
├── README.md          # 本说明
├── interface.go       # Result / Runner 统一执行器
├── adapter.go         # Adapter 公共基座、默认超时
├── ffuf/              # ffuf 目录与参数枚举（BB05.001 目录枚举）
├── nmap/              # nmap 端口扫描（BB02.005 端口发现）
├── nuclei/            # nuclei 已知 CVE 检测（BB22.001）
├── httpx/             # httpx HTTP 存活探测（BB02.001 / BB05.003）
├── subfinder/         # subfinder 子域名枚举（BB01.003）
└── sqlmap/            # sqlmap SQL 注入检测（BB10.001）
```

## 统一执行器

[interface.go](interface.go) 提供所有适配器复用的执行器：

- `tools.NewRunner(executablePath string, timeout time.Duration) *Runner`
- `(*Runner).Run(ctx, arguments []string) (*Result, error)`
- `Result{Stdout, Stderr string; ExitCode int}`
- `(*Result).Succeeded() bool`：退出码为 0 时返回 true，供调用方快速判断成败

[adapter.go](adapter.go) 提供适配器公共基座，消除各工具包重复的构造样板：

- `tools.NewAdapter(executablePath string, timeout time.Duration) *Adapter`
- `(*Adapter).Run(ctx, arguments []string) (*Result, error)`
- `tools.DefaultToolTimeout = 60s`：`NewAdapter` 收到小于等于零的超时时自动回退，防止误传 0 导致立即超时

具体适配器以组合方式嵌入 `*tools.Adapter`，只需声明业务方法并组装参数。

## 适配器清单

| 包 | 类型 | 方法 | 参数 | 对应 catalog 技术 |
|---|---|---|---|---|
| `tools/nmap` | `Scanner` | `Scan(ctx, target, ports)` | `-p <ports> <target>` | BB02.005 端口发现 |
| `tools/ffuf` | `Fuzzer` | `Fuzz(ctx, url, wordlist, opts)` | `-u <url> -w <wordlist>` + options | BB05.001 目录枚举 |
| `tools/nuclei` | `Scanner` | `Scan(ctx, url)` | `-u <url>` | BB22.001 已知 CVE 检测 |
| `tools/httpx` | `Prober` | `Probe(ctx, url)` | `-silent -timeout 5 <url>` | BB02.001 存活主机发现、BB05.003 端点发现 |
| `tools/subfinder` | `Enumerator` | `Enumerate(ctx, domain)` | `-silent -d <domain>` | BB01.003 子域名枚举 |
| `tools/sqlmap` | `Injector` | `Inject(ctx, url)` | `-u <url> --batch` | BB10.001 SQL 注入 |

所有工具在 `configs/redteam.json` 的 `tools` 段声明。值为命令名时经 PATH 解析（推荐，跨机器可用）；值为绝对路径时直接使用该二进制。禁止硬编码工具位置。

## ffuf 选项

ffuf 参数较多的场景使用 `FuzzOptions` 集中构造，零值字段不进入命令行：

```go
// DefaultFuzzOptions 返回目录枚举常用参数：命中 200/301/302/403、20 线程、
// 每请求超时 10 秒、限速 100、最长 50 秒、静默输出。
options := ffuf.DefaultFuzzOptions()

// 按需覆盖个别字段，未设置的字段保持默认
options.Threads = 50

result, err := fuzzer.Fuzz(ctx, "https://example.com/FUZZ", "/tmp/words.txt", options)
```

| 字段 | 对应参数 | 说明 |
|---|---|---|
| `MatchCodes` | `-mc` | 命中的 HTTP 状态码列表 |
| `Threads` | `-t` | 并发线程数 |
| `TimeoutSec` | `-timeout` | 单请求超时秒数 |
| `RateLimit` | `-rate` | 每秒最大请求数 |
| `MaxTimeSec` | `-maxtime` | 最大执行秒数 |
| `Silent` | `-s` | 静默模式，仅输出命中条目 |

## 用法示例

```go
import "mitre_red_team/tools/httpx"

// 路径来自 configs/redteam.json，单次执行超时 30 秒
prober := httpx.New(configuration.Tools["httpx"], 30*time.Second)
result, err := prober.Probe(ctx, "https://example.com")
if err != nil {
	// 执行器层错误（如超时、可执行文件缺失）
}
if !result.Succeeded() {
	// 工具非零退出，stderr 携带原因
}
```

## 新增一个工具适配器

按以下五步接入新工具，各适配器保持结构一致：

1. 新建目录 `tools/<name>/`，包内文件命名 `<name>.go`（包名与目录同名）。
2. 声明业务类型并嵌入公共基座，构造器保持 `New(executablePath, timeout)` 签名：

   ```go
   type Scanner struct {
   	*tools.Adapter
   }

   func New(executablePath string, timeout time.Duration) *Scanner {
   	return &Scanner{Adapter: tools.NewAdapter(executablePath, timeout)}
   }
   ```

3. 实现业务方法：校验入参后调用 `s.Adapter.Run(ctx, arguments)` 组装参数切片，禁止拼接 shell 字符串。参数复杂时仿照 `ffuf.FuzzOptions` 提供选项结构。
4. 在 `configs/redteam.json` 的 `tools` 段声明工具路径。
5. 在 `test/tools_test.go` 新增参数构造测试，用 `fakeToolPath` 作假命令断言命令行内容；同时更新本 README 的适配器清单表。

## 集成点

- **配置**：`configs/redteam.json` 的 `tools` 段声明各工具可执行路径。
- **技术层**：`internal/technique/` 的领域实现按 catalog 的 `tools` 字段选用对应适配器（executor 注册模式）。目录枚举（BB05.001）经 `tools/ffuf` 适配器执行，参数构造不散落在技术层。
- **执行器**：所有适配器复用 `tools.Runner`，统一处理超时、stdout/stderr 与退出码。
- **测试**：`test/tools_test.go` 用测试二进制自身（`fakeToolPath`）作假命令验证各适配器的参数构造，跨平台可用，不依赖真实工具。
