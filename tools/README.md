# tools/ 工具适配层

`tools/` 是外部安全工具的统一适配层。技术实现（`internal/technique/`）不得直接执行 `exec.Command`，一律经此层调用外部工具。本层遵循 `dir_structure.md`、`code_structure.md`、`security_structure.md` 的约束：`exec.CommandContext` + 超时、参数切片传递、工具路径可配置、stderr 不吞。

## 统一执行器

[interface.go](interface.go) 提供所有适配器复用的执行器：

- `tools.NewRunner(executablePath string, timeout time.Duration) *Runner`
- `(*Runner).Run(ctx, arguments []string) (*Result, error)`
- `Result{Stdout, Stderr string; ExitCode int}`

所有工具路径来自 `configs/redteam.json` 的 `tools` 段，禁止硬编码。

## 适配器清单

| 包 | 类型 | 方法 | 参数 | 对应 catalog 技术 |
|---|---|---|---|---|
| `tools/nmap` | `Scanner` | `Scan(ctx, target, ports)` | `-p <ports> <target>` | BB02.005 端口发现 |
| `tools/ffuf` | `Fuzzer` | `Fuzz(ctx, url, wordlist)` | `-u <url> -w <wordlist>` | BB05.001 目录枚举 |
| `tools/nuclei` | `Scanner` | `Scan(ctx, url)` | `-u <url>` | BB22.001 已知 CVE 检测 |
| `tools/httpx` | `Prober` | `Probe(ctx, url)` | `-silent -timeout 5 <url>` | BB02.001 存活主机发现、BB05.003 端点发现 |
| `tools/subfinder` | `Enumerator` | `Enumerate(ctx, domain)` | `-silent -d <domain>` | BB01.003 子域名枚举 |
| `tools/sqlmap` | `Injector` | `Inject(ctx, url)` | `-u <url> --batch` | BB10.001 SQL 注入 |

## 用法示例

```go
import "mitre_red_team/tools/httpx"

// 路径来自 configs/redteam.json，单次执行超时 30 秒
prober := httpx.New(configuration.Tools["httpx"], 30*time.Second)
result, err := prober.Probe(ctx, "https://example.com")
```

## 集成点

- **配置**：`configs/redteam.json` 的 `tools` 段声明各工具可执行路径。
- **技术层**：`internal/technique/` 的领域实现按 catalog 的 `tools` 字段选用对应适配器（executor 注册模式）。
- **执行器**：所有适配器复用 `tools.Runner`，统一处理超时、stdout/stderr 与退出码。
- **测试**：`test/tools_test.go` 用 `/bin/echo` 作假命令验证各适配器的参数构造，不依赖真实工具二进制。
