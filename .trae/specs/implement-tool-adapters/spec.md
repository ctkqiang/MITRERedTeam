# 工具适配层实现 Spec

## Why

技术实现必须统一经工具适配层调用外部工具，当前 `tools/` 仅有空壳（`nmap`、`ffuf`、`nuclei`），适配层尚未实现。这是执行链路的下一块基石，`internal-phase-completion-plan.md` 的 Phase 2。

## What Changes

- 完成 `tools/` 下 ffuf 目录与文件的重命名（`dir_structure.md` §7 记录的问题已解决）。
- `tools/` 根新增公共执行器 `Runner` 与结果类型 `Result`：`exec.CommandContext` + 超时、参数切片传递、stdout/stderr 捕获、退出码记录。
- 实现 `nmap`（端口扫描）、`ffuf`（目录枚举）、`nuclei`（漏洞检测）三个适配器，可执行路径经构造器注入（来自 `configs/redteam.json`）。
- 新增 `test/tools_test.go`，用假命令（`/bin/echo`、`/bin/ls`、`/usr/bin/sleep`）mock，不依赖真实工具二进制。

## Impact

- Affected specs: `dir_structure.md`（tools/ 层职责）、`security_structure.md`（命令执行与超时约束）。
- Affected code: `tools/*`、`test/tools_test.go`。
- 依赖：`internal/config`（工具路径来源，已完成）。

## ADDED Requirements

### Requirement: 工具统一执行器

系统 SHALL 提供统一工具执行器 `Runner`：

- 使用 `exec.CommandContext`，参数以切片传递，不做 shell 解释，禁止 `sh -c` 拼接。
- 每次执行设置超时（构造时注入），超时返回错误。
- 捕获 stdout 与 stderr，记录退出码；stderr 不得被吞掉。

#### Scenario: 成功执行

- **WHEN** 调用 `Run(ctx, []string{"hello"})` 且可执行文件为 `/bin/echo`
- **THEN** 返回 `Result{Stdout: "hello\n", ExitCode: 0}`，无错误

#### Scenario: 命令超时

- **WHEN** 调用 `Run` 执行耗时超过 timeout 的命令
- **THEN** 返回错误，不静默继续

#### Scenario: 非零退出码

- **WHEN** 命令以非零码退出（如 `ls` 访问不存在的路径）
- **THEN** 返回 `Result` 且 `ExitCode` 正确记录，stderr 含错误输出

### Requirement: 工具适配器

系统 SHALL 为 `nmap`、`ffuf`、`nuclei` 提供适配器：

- 每个适配器经构造器接收可执行文件路径与超时，路径不得硬编码。
- 适配器负责构造该工具的特定参数（如 `nmap -p <ports> <target>`）。

#### Scenario: 端口扫描

- **WHEN** 调用 `nmap.New(path, timeout).Scan(ctx, target, ports)`
- **THEN** 以切片参数 `["-p", ports, target]` 调用执行器

#### Scenario: 目录枚举

- **WHEN** 调用 `ffuf.New(path, timeout).Fuzz(ctx, url, wordlist)`
- **THEN** 以切片参数 `["-u", url, "-w", wordlist]` 调用执行器

#### Scenario: 漏洞检测

- **WHEN** 调用 `nuclei.New(path, timeout).Scan(ctx, url)`
- **THEN** 以切片参数 `["-u", url]` 调用执行器

## MODIFIED Requirements

无。

## REMOVED Requirements

无。
