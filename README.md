# MITRERedTeam

**基于 MITRE ATT&CK 的授权红队评估 CLI 工具**

> **作者**：钟智强 · **许可证**：[GPLv3](https://www.gnu.org/licenses/gpl-3.0.html)

## 目录

- [项目概述](#项目概述)
- [系统架构](#系统架构)
- [执行时序](#执行时序)
- [支持的战术](#支持的战术)
- [支持的技术](#支持的技术)
- [工具集成](#工具集成)
- [执行模式](#执行模式)
- [AI 辅助执行](#ai-辅助执行)
- [通知系统](#通知系统)
- [日志系统](#日志系统)
- [项目结构](#项目结构)
- [编译与运行](#编译与运行)
  - [前置条件](#前置条件)
  - [安装依赖](#安装依赖)
  - [方式一：直接运行](#方式一直接运行)
  - [方式二：Make 构建](#方式二make-构建)
- [CLI 用法](#cli-用法)
- [快速上手](#快速上手)
- [运行输出](#运行输出)
- [配置说明](#配置说明)
- [扩展指南](#扩展指南)
- [安全最佳实践](#安全最佳实践)
- [技术实现细节](#技术实现细节)
- [依赖项](#依赖项)
- [许可证](#许可证)

## 项目概述

MITRERedTeam 是一个用 **Go** 编写的授权红队 / 漏洞赏金 CLI 安全工具。它以 MITRE ATT&CK 战术与技术模型为骨架，把完整的评估流程拆分为**战术（Tactic）**与**技术（Technique）**两级，用户显式选择单条技术或整个战术执行，应用绝不自动编排所有内容，也不扩大用户声明的目标范围。

战术与技术清单由 `catalog/*.json` 外部数据驱动，技术主键采用自研稳定编号（`BBxx` 战术、`BBxx.xxx` 技术），MITRE ATT&CK ID 仅作为行业标准对齐的元数据。当前目录覆盖 **25 个战术、27 条技术**，并已落地目录枚举执行器与六类外部工具适配层。

**核心特性：**

| 特性 | 描述 |
|---|---|
| 战术驱动 | 25 个战术（BB01–BB25）覆盖侦察、攻击面发现、Web 枚举、注入、业务逻辑、云安全、证据报告等完整评估流程 |
| 数据驱动 | 战术与技术清单定义在 `catalog/*.json`，新增内容无需改动 Go 代码 |
| MITRE 对齐 | 每条技术映射 MITRE ATT&CK ID（如 T1046、T1190、T1083）作为行业标准对齐元数据 |
| 显式执行 | 用户通过 `--technique` / `--tactic` / `--mitre` 显式触发，严格限定在声明目标内 |
| 工具适配层 | ffuf、nmap、nuclei、httpx、subfinder、sqlmap 统一封装，参数构造与超时收敛在适配层 |
| 结构化日志 | 全链路日志携带 RFC3339 时间戳、级别、操作名与内存统计，便于诊断 |
| AI 辅助模式 | 对接六家 LLM 供应商，分析 TTP 执行输出并自动推进建议的下一步技术 |
| 通知系统 | Telegram、OpenClaw 微信、本地系统通知三通道 |
| 零第三方依赖 | 仅使用 Go 标准库，Go 1.26 实现 |

## 系统架构

MITRERedTeam 采用**目录驱动的注册式插件架构**（Catalog-Driven Registry Plugin Architecture，CDRPA），组合五种模式：目录驱动、注册模式、插件架构、分层架构、依赖倒置。核心系统只依赖抽象，不依赖具体安全工具与技术实现；新增技术、工具集成时尽量不改动核心执行引擎。

![系统架构与数据流](docs/assets/flow.png)

📁 图表源码：[flow.puml](docs/assets/flow.puml) | 🖼️ PNG：[flow.png](docs/assets/flow.png)

| 层 | 职责 |
|---|---|
| CLI | 解析用户请求（`--tactic` / `--technique` / `--mitre` / `--url`），仅执行用户选定项 |
| Planner | 把用户请求解析为可执行的执行计划 |
| Engine | 编排执行：目录查询、计划调度、结果汇总 |
| Technique Registry | 维护 `executor → 实现` 的注册表，按名称解析 |
| 领域技术 | 具体技术实现，按领域分目录（recon、enumeration、injection 等） |
| Tool 适配层 | 封装外部工具：路径解析、参数构造、超时、输出处理 |
| Result / Evidence / Finding | 执行结果、证据与最终发现的递进产出 |

完整链路：用户通过 CLI 提交请求 → Planner 生成执行计划 → Engine 校验目标并调度 → Technique Registry 按 `executor` 解析出领域技术实现 → 技术实现经 Tool 适配层调用外部工具（ffuf、nmap、nuclei 等）→ 产出结构化 Result。

## 执行时序

下图展示一次目录枚举技术（BB05.001）从 CLI 触发到外部工具执行、结果回传的完整时序：

![技术执行时序](docs/assets/sequence.png)

📁 图表源码：[sequence.puml](docs/assets/sequence.puml) | 🖼️ PNG：[sequence.png](docs/assets/sequence.png)

## 支持的战术

| ID | 战术 | 描述 |
|---|---|---|
| BB01 | 侦察 | 发现目标及其组织的公开可观察信息 |
| BB02 | 攻击面发现 | 识别可访问的主机、服务、应用程序及暴露的接口 |
| BB03 | DNS 与域名分析 | 分析 DNS 基础设施、域名关系及 DNS 安全配置 |
| BB04 | 网络与服务枚举 | 识别暴露的网络服务并分析其配置特征 |
| BB05 | Web 应用枚举 | 发现 Web 应用的路由、资源、参数和接口 |
| BB06 | API 枚举 | 发现并分析 REST、GraphQL 及其他 API 接口 |
| BB07 | 认证 | 评估认证、身份验证与账户恢复机制 |
| BB08 | 授权与访问控制 | 评估授权边界、对象所有权与权限分离 |
| BB09 | 会话管理 | 评估会话生命周期、Cookie 与令牌处理 |
| BB10 | 注入 | 评估应用输入处理中的注入漏洞 |
| BB11 | 客户端安全 | 评估浏览器端应用的安全控制及客户端攻击面 |
| BB12 | 服务端请求处理 | 评估服务端对 URL、回调及外部资源的处理 |
| BB13 | 文件与资源处理 | 评估文件上传、下载、路径处理及资源处理 |
| BB14 | 业务逻辑 | 评估应用工作流、状态转换、事务与业务规则 |
| BB15 | API 安全 | 评估 API 授权、对象处理、速率限制及协议相关安全 |
| BB16 | 配置与部署 | 识别与安全相关的配置及部署弱点 |
| BB17 | 信息泄露 | 识别应用或基础设施敏感信息的非预期泄露 |
| BB18 | 云与存储 | 评估云资源、对象存储、无服务器服务及云暴露面 |
| BB19 | 依赖与组件安全 | 识别存在漏洞、过时或暴露的软件组件 |
| BB20 | 密码与传输安全 | 评估 TLS、证书、密码学及敏感数据传输控制 |
| BB21 | 基础设施安全 | 评估暴露的基础设施服务与管理接口 |
| BB22 | 漏洞检测 | 识别已知漏洞及安全相关条件 |
| BB23 | 验证与影响评估 | 验证发现、评估影响并减少误报 |
| BB24 | 攻击路径与漏洞利用链 | 关联各项发现并识别多步安全影响 |
| BB25 | 证据与报告 | 收集、关联并导出可复现的安全发现 |

## 支持的技术

`catalog/techniques.json` 当前收录 27 条技术，每条声明执行器、执行模式、依赖工具与 MITRE 映射。已实现状态指对应的 Go 执行器是否已在 `internal/technique/` 注册：

| ID | 名称 | 战术 | 模式 | 工具 | MITRE | 已实现 |
|---|---|---|---|---|---|---|
| BB01.001 | 组织信息发现 | BB01 | passive | - | T1589 | 否 |
| BB01.002 | 域名发现 | BB01 | passive | whois | T1589.001 | 否 |
| BB01.003 | 子域名枚举 | BB01 | passive | subfinder | T1595 | 否 |
| BB01.004 | 证书透明度枚举 | BB01 | passive | crtsh | T1596.003 | 否 |
| BB01.005 | ASN 枚举 | BB01 | passive | - | T1590.005 | 否 |
| BB01.006 | IP 段发现 | BB01 | passive | - | T1590.005 | 否 |
| BB01.007 | WHOIS 枚举 | BB01 | passive | whois | T1596.002 | 否 |
| BB01.008 | DNS 枚举 | BB01 | passive | dig | T1590.002 | 否 |
| BB01.009 | 反向 DNS 枚举 | BB01 | passive | dig | T1590.002 | 否 |
| BB01.010 | 历史 DNS 发现 | BB01 | passive | - | T1596.002 | 否 |
| BB02.001 | 存活主机发现 | BB02 | active | httpx | T1595 | 否 |
| BB02.005 | 端口发现 | BB02 | active | nmap | T1046 | 否 |
| BB05.001 | 目录枚举 | BB05 | active | ffuf | T1083 | **是** |
| BB05.003 | 端点发现 | BB05 | active | ffuf, httpx | T1046 | 否 |
| BB07.001 | 登录流程分析 | BB07 | manual | - | - | 否 |
| BB08.001 | IDOR 测试 | BB08 | manual | - | - | 否 |
| BB10.001 | SQL 注入 | BB10 | active | sqlmap | T1190 | 否 |
| BB11.001 | 反射型 XSS | BB11 | active | nuclei | T1189 | 否 |
| BB12.001 | SSRF 发现 | BB12 | active | - | T1090 | 否 |
| BB13.001 | 路径遍历 | BB13 | active | nuclei | T1083 | 否 |
| BB14.001 | 工作流绕过 | BB14 | manual | - | - | 否 |
| BB14.003 | 竞态条件测试 | BB14 | active | - | - | 否 |
| BB15.005 | 批量赋值 | BB15 | manual | - | - | 否 |
| BB16.011 | Git 仓库泄露 | BB16 | active | nuclei | T1213 | 否 |
| BB17.002 | 敏感信息泄露 | BB17 | passive | - | T1552 | 否 |
| BB18.002 | 公共存储桶检测 | BB18 | active | - | T1530 | 否 |
| BB22.001 | 已知 CVE 检测 | BB22 | active | nuclei | T1203 | 否 |

执行规划中的技术时，引擎会跳过未注册的执行器并明确提示，不静默成功。

## 工具集成

所有外部工具经 `tools/` 适配层调用，统一处理路径解析、参数构造、context 超时、stdout/stderr 与退出码。技术实现不得直接执行 `exec.Command`。工具路径在 `configs/redteam.json` 的 `tools` 段声明，值为命令名时经 PATH 解析，值为绝对路径时直接使用。

| 工具 | 适配器 | 方法 | 典型参数 | 对应技术 |
|---|---|---|---|---|
| ffuf | `Fuzzer` | `Fuzz(ctx, url, wordlist, opts)` | `-u <url> -w <wordlist> -mc ...` | BB05.001 目录枚举、BB05.003 端点发现 |
| nmap | `Scanner` | `Scan(ctx, target, ports)` | `-p <ports> <target>` | BB02.005 端口发现 |
| nuclei | `Scanner` | `Scan(ctx, url)` | `-u <url>` | BB22.001 已知 CVE 检测 |
| httpx | `Prober` | `Probe(ctx, url)` | `-silent -timeout 5 <url>` | BB02.001 存活主机发现、BB05.003 端点发现 |
| subfinder | `Enumerator` | `Enumerate(ctx, domain)` | `-silent -d <domain>` | BB01.003 子域名枚举 |
| sqlmap | `Injector` | `Inject(ctx, url)` | `-u <url> --batch` | BB10.001 SQL 注入 |

## 执行模式

| 模式 | 含义 |
|---|---|
| passive | 被动评估，仅基于公开数据，不向目标发送主动请求 |
| active | 主动评估，向目标发送探测或测试请求 |
| manual | 需人工介入验证，工具只输出上下文与指引 |

## AI 辅助执行

通过 `--ai` 启用 AI 辅助模式：程序执行用户请求的初始技术，把执行输出交给 LLM 分析，并由 LLM 从目录中选择并自动推进建议的下一步技术（最多 3 轮）。至少配置一家供应商的 API Key 后才能启用；未配置任何供应商时 `--ai` 会明确报错，不会静默失败。

| 供应商 | 环境变量 | 默认模型 |
|---|---|---|
| OpenAI | `OPENAI_API_KEY` | `gpt-4o-mini` |
| DeepSeek | `DEEPSEEK_API_KEY` | `deepseek-v4-flash` |
| Kimi（Moonshot） | `MOONSHOT_API_KEY` | `kimi-k3` |
| 豆包（火山方舟） | `ARK_API_KEY` | `doubao-seed-2-1-pro-260628` |
| OpenRouter | `OPENROUTER_API_KEY` | 任意模型 slug |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-sonnet-4-5` |

## 通知系统

执行结果可通过三条通道送达，配置见 `configs/redteam.json` 的 `notifications` 段与 `.env`：

| 通道 | 配置 | 说明 |
|---|---|---|
| 本地系统通知 | `notifications.local` | macOS（osascript）、Linux（notify-send）、Windows 原生通知 |
| Telegram | `NOTIFICATION_PLATFORM=telegram` | 需 `TELEGRAM_BOT_TOKEN` 与 `TELEGRAM_CHAT_ID` |
| OpenClaw 微信 | `NOTIFICATION_PLATFORM=wechat` | 经本机 OpenClaw gateway 长连接投递，无需外部 API 凭据 |

## 日志系统

应用全程使用结构化日志器，每行日志包含 RFC3339 时间戳、级别、操作名（`op=`）、内存统计（`mem[]=`）与描述（`desc=`），输出到 stderr：

```
2026-08-17T03:57:25Z [信息] op=LoadDotenv desc=环境文件加载完成
2026-08-17T03:57:25Z [信息] op=LoadCatalog desc=目录校验通过：战术 25 条，技术 27 条
2026-08-17T03:57:25Z [信息] op=ResolveWordlist desc=选定字典 configs/wordlists/common.txt
2026-08-17T03:57:25Z [信息] op=Execute desc=按技术 BB05.001 执行
2026-08-17T03:57:25Z [信息] op=Summary desc=按技术执行完成：1/1 项成功
```

| 级别 | 中文标识 | 用途 |
|---|---|---|
| INFO | 信息 | 常规运行信息 |
| DEBUG | 调试 | 开发调试信息 |
| WARN | 警告 | 可恢复异常、用户取消 |
| ERROR | 错误 | 执行失败、需人工介入 |
| SECURITY | 安全 | 安全相关事件 |
| FATAL | 致命 | 致命错误并退出 |

## 项目结构

```
MITRERedTeam/
├── .githooks/
│   └── pre-commit              # 提交前质量门禁：gofmt → build → test → golangci-lint
├── catalog/
│   ├── tactics.json            # 战术目录数据：25 个战术（BB01–BB25）
│   └── techniques.json         # 技术目录数据：27 条技术，含 executor/mode/tools/mitre
├── cmd/
│   └── mitre_red_team/
│       └── main.go             # CLI 主程序入口
├── configs/
│   ├── wordlists/
│   │   └── common.txt          # 默认目录枚举字典
│   └── redteam.json            # 运行配置：工具路径、字典、通知偏好
├── docs/
│   └── mitre-mapping.md        # MITRE 编号映射说明
├── internal/
│   ├── agent/                  # AI 辅助执行：决策提示与多轮循环
│   ├── app/                    # 应用装配层
│   ├── catalog/                # 目录加载、校验、ID 索引注册表
│   ├── config/                 # 配置加载与工具可用性检查
│   ├── engine/                 # 执行编排：技术/战术/MITRE 调度
│   ├── llm/                    # 六家 LLM 供应商客户端
│   ├── model/                  # 数据契约：Tactic/Technique/Target/Execution
│   ├── technique/              # 执行器注册表与领域实现（enumeration 等）
│   └── utilities/              # 日志、通知、dotenv、内存监控
├── test/                       # 全部 Go 测试（外部测试包）
├── tools/                      # 外部工具适配层
│   ├── interface.go            # Result / Runner 统一执行器
│   ├── adapter.go              # Adapter 公共基座、默认超时
│   ├── ffuf/ nmap/ nuclei/     # 各工具适配器
│   ├── httpx/ subfinder/ sqlmap/
│   └── README.md               # 适配层文档与新增适配器指南
├── Makefile                    # 常用开发命令
└── go.mod                      # module mitre_red_team / go 1.26.1 / 零第三方依赖
```

## 编译与运行

### 前置条件

- Go 1.26 或更高版本
- 按需安装外部工具（ffuf、nmap、nuclei、httpx、subfinder、sqlmap），未安装时程序会提示缺失工具与安装指引

### 安装依赖

**macOS (Homebrew):**

```
brew install go ffuf nmap sqlmap
go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
go install github.com/projectdiscovery/httpx/cmd/httpx@latest
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
```

**Ubuntu / Debian:**

```
sudo apt-get install golang ffuf nmap sqlmap
go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
go install github.com/projectdiscovery/httpx/cmd/httpx@latest
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
```

### 方式一：直接运行

```
# 目录枚举
go run ./cmd/mitre_red_team --url https://example.com --technique BB05.001

# 执行整个战术
go run ./cmd/mitre_red_team --url https://example.com --tactic BB05

# 按 MITRE ID 执行
go run ./cmd/mitre_red_team --url https://example.com --mitre T1046
```

### 方式二：Make 构建

```
# 编译可执行文件到 build/mitre_red_team
make binary

# 运行（参数通过 ARGS 传递）
make run ARGS="--url https://example.com --technique BB05.001"

# 完整质量门禁
make all
```

## CLI 用法

| 参数 | 说明 |
|---|---|
| `--url <target>` | 目标 URL，必填 |
| `--technique <ID>` | 执行单条技术，如 `BB05.001` |
| `--tactic <ID>` | 执行整个战术，如 `BB05` |
| `--mitre <ID>` | 按 MITRE ATT&CK ID 执行，如 `T1046` |
| `--wordlist / -w <path>` | 自定义字典文件路径 |
| `--ai` | 启用 AI 辅助执行模式 |
| `--config <path>` | 配置文件路径（默认 `configs/redteam.json`） |

`--technique`、`--tactic`、`--mitre` 三选一；未指定任何选择器或未提供 `--url` 时会报错并打印完整用法。

## 快速上手

**目录枚举（BB05.001）：**

```
# 非交互环境（管道 / CI / IDE 内嵌）自动回退默认字典，不会阻塞等待输入
go run ./cmd/mitre_red_team --url https://example.com --technique BB05.001

# 使用自定义字典
go run ./cmd/mitre_red_team --url https://example.com --technique BB05.001 -w /path/to/words.txt
```

**端口发现（BB02.005）：**

端口发现执行器目前规划中。执行未实现的技术时引擎会明确提示，例如：

```
go run ./cmd/mitre_red_team --url https://example.com --mitre T1046
# 输出：错误: 所选技术均未实现执行器，无法执行
```

**AI 辅助执行：**

```
# 复制 .env.example 为 .env 并填入至少一家 LLM 供应商的 API Key
cp .env.example .env
go run ./cmd/mitre_red_team --url https://example.com --technique BB05.001 --ai
```

## 运行输出

```
2026-08-17T02:17:33Z [信息] op=LoadDotenv desc=环境文件加载完成
2026-08-17T02:17:33Z [信息] op=LoadConfig desc=已加载配置 configs/redteam.json，工具 6 项，字典 1 项
2026-08-17T02:17:33Z [信息] op=LoadCatalog desc=目录校验通过：战术 25 条，技术 27 条
2026-08-17T02:17:33Z [信息] op=ValidateFlags desc=命令行参数校验通过
2026-08-17T02:17:33Z [信息] op=CheckTools desc=全部依赖工具可用
2026-08-17T02:17:33Z [信息] op=ResolveWordlist desc=选定字典 configs/wordlists/common.txt
2026-08-17T02:17:33Z [信息] op=ParseTarget desc=目标解析完成：主机 example.com，协议 https，端口 0
2026-08-17T02:17:33Z [信息] op=Execute desc=按技术 BB05.001 执行
succeeded BB05.001: 目录枚举完成，未发现命中（匹配状态码 200/301/302/403）
2026-08-17T02:17:35Z [信息] op=Summary desc=按技术执行完成：1/1 项成功
```

## 配置说明

`configs/redteam.json` 声明工具路径、字典与通知偏好：

```json
{
  "tools": {
    "ffuf": "ffuf",
    "nmap": "nmap",
    "nuclei": "nuclei",
    "httpx": "httpx",
    "subfinder": "subfinder",
    "sqlmap": "sqlmap"
  },
  "wordlists": {
    "common": "configs/wordlists/common.txt"
  },
  "notifications": {
    "platform": "none",
    "local": true
  }
}
```

| 段 | 说明 |
|---|---|
| `tools` | 工具名到可执行文件的映射，命令名经 PATH 解析，绝对路径直接使用 |
| `wordlists` | 字典名到文件路径的映射，技术按名称引用，禁止硬编码路径 |
| `notifications` | `platform`（telegram / wechat / none）与本地通知开关 `local` |

环境变量（`.env`）用于通知平台凭据与 AI 供应商 API Key，敏感信息不进版本仓库。启动时 `.env` 文件不存在或条目为空会被静默忽略。

## 扩展指南

**新增一条技术：**

1. 在 `catalog/techniques.json` 添加条目，声明 `executor`、`mode`、`tools` 与 `mitre` 映射。
2. 在 `internal/technique/` 对应领域子目录实现执行器，经注册表注册。
3. 技术实现不得直接执行 `exec.Command`，统一经 `tools/` 适配层调用外部工具。

**新增一个工具适配器：** 按 `tools/README.md` 中的五步指南接入：新建 `tools/<name>/`、嵌入 `tools.Adapter` 公共基座、实现业务方法、声明配置、补测试。

**提交前质量门禁：** `.githooks/pre-commit` 强制 `gofmt -l` 为空、`go build ./...` 无错误、`go test ./...` 全部通过、`golangci-lint` 无错误。

## 安全最佳实践

- 仅在**授权范围**内使用，执行必须由用户显式选择，不自动扩大目标。
- 外部命令一律 `exec.CommandContext`，参数以切片传递，不做 shell 解释；禁止 `sh -c` 拼接。
- 所有外部调用带 context 超时；工具输出按行数上限约束，防止异常输出拖垮本地资源。
- 外部输入（CLI 参数、配置、工具输出）先校验长度与格式再进入业务逻辑。
- 凭据与 API Key 只存内存，不回显、不落盘、不写日志。
- 非交互环境（管道 / CI / IDE 内嵌）自动跳过交互询问，避免无声阻塞。

## 技术实现细节

- **CDRPA 架构**：目录驱动 + 注册模式 + 插件架构 + 分层 + 依赖倒置，核心引擎不依赖具体工具与技术实现。
- **索引化查询**：目录加载阶段建立 ID 索引，`GetTactic` / `GetTechnique` 为 O(1) 查表，替代线性扫描。
- **工具适配基座**：`tools.Adapter` 公共基座统一执行器与默认超时兜底，ffuf 参数经 `FuzzOptions` 集中构造。
- **流式加载**：catalog JSON 用 `json.Decoder` 流式消费，内存占用与单条记录成正比，不整文件常驻。
- **非交互检测**：`promptEnabled` 检测 stdin 是否为终端字符设备，非交互环境跳过询问直接回退默认字典。
- **内存约束**：RSS 上限 500MB，`GOMEMLIMIT=480MiB` 软限制，监控协程定期采样 `runtime.ReadMemStats`，连续超限即退出。
- **确定性执行**：技术按目录声明顺序执行，结果结构化输出，失败不静默。

## 依赖项

| 类型 | 内容 |
|---|---|
| 运行时 | Go 1.26.1，零第三方 Go 依赖（仅标准库） |
| 外部工具 | ffuf、nmap、nuclei、httpx、subfinder、sqlmap（按需安装） |
| AI 辅助 | 六家 LLM 供应商 API（OpenAI / DeepSeek / Kimi / 豆包 / OpenRouter / Anthropic，`--ai` 模式） |

## 许可证

本项目采用 [GPLv3](https://www.gnu.org/licenses/gpl-3.0.html) 许可证，作者：钟智强。
