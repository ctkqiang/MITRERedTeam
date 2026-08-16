---
alwaysApply: true
scene: dir_structure
---

# MITRERedTeam 目录结构与架构规范

本文件是项目架构与目录结构的权威规范。所有新增、修改代码的操作必须先对照本文件：盘点现有实现，遵循既有分层与命名，不得绕过约束。

## 1. 项目概述

MITRERedTeam 是一个授权范围内的红队 / 漏洞赏金 CLI 安全工具，使用 Go 编写。

**核心交互模型**：用户显式触发单一战术或技术，应用不自动执行全部内容。

```
redteam run --technique BB05.001 --target https://example.com
redteam run --tactic BB05 --target https://example.com
```

**硬性边界**：

- 应用**不是** AI agent，**不得**自动编排执行所有战术。
- 执行必须由用户显式选择，且严格限定在用户声明的目标范围内。
- 战术与技术清单**数据驱动**，存放在 `catalog/` 外部 JSON 中，不得硬编码进 Go 代码。

**编号体系**：项目使用自研稳定编号作为主键。

- `BBxx`：战术（如 `BB05` Web 应用枚举）。
- `BBxx.xxx`：技术（如 `BB05.001` 目录枚举）。
- MITRE ATT&CK ID（如 `T1083`）仅为元数据，用于对齐行业标准，**不**作为主标识符。

## 2. 现状目录结构（权威）

以下结构基于对仓库的实际盘点，是当前唯一准确的状态。`internal/model/` 与 `catalog/` 为权威起点，不得擅自改动其字段语义。

```
MITRERedTeam/
├── .githooks/
│   └── pre-commit              # 提交前质量门禁：gofmt → build → test → golangci-lint
├── .trae/
│   └── rules/
│       └── dir_structure.md    # 本文档
├── catalog/
│   ├── tactics.json            # 战术目录数据：25 个战术（BB01–BB25）
│   └── techniques.json         # 技术目录数据：27 条技术，含 executor/mode/tools/mitre
├── cmd/
│   ├── mitre_red_team/
│   │   └── main.go             # CLI 主程序入口（当前为空壳，package 声明有误，见 §7）
│   └── mitre_red_team_agent/
│       └── main.go             # Agent 程序入口（当前为空壳，package 声明有误，见 §7）
├── internal/
│   ├── app/
│   │   ├── app.go              # 应用装配层（当前为空壳）
│   │   └── bootstrap.go        # 启动引导（当前为空壳）
│   └── model/
│       ├── tactics.go          # Tactic 模型：ID/Name/Description/Techniques
│       ├── techniques.go       # Technique 模型 + TechniqueMode(passive/active/manual)
│       └── target.go           # Target 模型：ID/Host/Port/Scheme/Metadata
├── tools/
│   ├── fuff/
│   │   └── fuff.go             # 外部工具适配层骨架（目录名拼写有误，应为 ffuf，见 §7）
│   ├── nmap/
│   │   └── nmap.go             # 外部工具适配层骨架（未实现）
│   └── nuclei/
│       └── nuclei.go           # 外部工具适配层骨架（未实现）
├── README.md                   # 项目说明（当前为空）
└── go.mod                      # module mitre_red_team / go 1.26.1 / 零第三方依赖
```

## 3. 目标架构与分层职责

目标架构在现有结构之上补齐缺失组件，分层职责如下。层与层之间不得越权。

```
CLI（cmd/）
  │
  ▼
Engine（internal/engine/）        编排：目标校验 → 目录查询 → 执行计划 → 执行 → 结果
  │
  ├── Catalog（internal/catalog/）加载 catalog JSON → 校验 → 按 ID 解析
  ├── Technique（internal/technique/）executor 字符串 → Go 实现（注册模式）
  ├── Tool（tools/）              外部工具封装：路径解析、参数构造、执行、结果解析
  └── Output（internal/output/）  结果输出：console / json / report
```

各层职责：

- **model**（已有）：定义 `Tactic`、`Technique`、`Target` 等数据契约，新增 `execution.go`、`result.go`。
- **catalog**（新增）：加载 `catalog/*.json`，校验字段合法性与引用完整性，提供按 ID 查询的注册表。
- **engine**（新增）：接收 CLI 传入的 `--tactic` / `--technique` 与 `--target`，解析为执行计划，调度执行并汇总结果。仅执行用户选定项。
- **technique**（新增）：以注册模式维护 `executor → 实现` 的映射。按领域分子目录（recon、enumeration、authentication、authorization、injection、clientside、server、businesslogic、cloud、infrastructure）。技术实现不得直接执行 `exec.Command`。
- **tool**（已有 `tools/`，待实现）：封装外部安全工具（ffuf、nmap、nuclei、httpx、subfinder、sqlmap）。负责可执行文件解析、参数构造、`context.Context` 与超时、stdout/stderr 处理、退出码与结构化输出。工具路径必须可配置，不得硬编码。
- **config**（新增）：加载 `configs/redteam.json`，提供工具路径等运行配置。
- **output**（新增）：将执行结果格式化为 console、JSON 或报告。

## 4. 现状 → 目标映射

| 现有组件 | 状态 | 处理 | 理由 |
|---|---|---|---|
| `go.mod` | 可用 | Keep | 模块名与 Go 版本有效，零依赖符合工程约束 |
| `internal/model/{tactics,techniques,target}.go` | 权威起点 | Keep | 数据契约已定义，作为后续所有层的依赖基础 |
| `catalog/{tactics,techniques}.json` | 数据可用 | Keep | 数据驱动来源，后续仅按需增补条目 |
| `.githooks/pre-commit` | 已完成 | Keep | 质量门禁已生效 |
| `cmd/mitre_red_team/main.go` | 空壳、有缺陷 | Modify | 修正 package 声明，接入 engine 与 CLI 解析 |
| `cmd/mitre_red_team_agent/main.go` | 空壳、有缺陷 | Modify | 修正 package 声明，按 agent 定位实现 |
| `internal/app/{app,bootstrap}.go` | 空壳 | Modify | 作为装配入口承接 engine 初始化 |
| `tools/{nmap,nuclei}/*.go` | 空壳 | Modify | 实现工具适配层 |
| `tools/fuff/` | 空壳、拼写有误 | Modify | 重命名为 `ffuf` 后实现 |
| `README.md` | 空 | Modify | 补充项目说明 |
| `internal/model/{execution,result}.go` | 缺失 | New | 执行计划与结果模型 |
| `internal/catalog/` | 缺失 | New | loader / registry / validator |
| `internal/engine/` | 缺失 | New | engine / planner / executor |
| `internal/technique/` | 缺失 | New | 接口 + 注册表 + 领域实现 |
| `internal/config/` | 缺失 | New | 配置加载 |
| `internal/output/` | 缺失 | New | console / json / report 输出 |
| `configs/redteam.json` | 缺失 | New | 运行配置（含工具路径） |
| `tests/` | 缺失 | New | 集成测试与 fixtures |

## 5. 核心设计原则

**数据驱动**

战术与技术定义于 `catalog/*.json`，由 `internal/catalog/` 加载。禁止在 Go 中维护巨型硬编码映射。

**Executor 注册模式**

`catalog` 中 `executor` 字段（如 `directory-enumeration`）映射到 `internal/technique/` 中的 Go 实现，通过注册表按名称解析。禁止使用巨型 `switch` 分发。

数据流：

```
catalog（BB05.001）
  → Technique 元数据（executor = directory-enumeration）
  → Technique Registry
  → DirectoryEnumeration 实现
  → Tool Registry
  → ffuf 适配层
  → Execution Result
```

**工具抽象层**

技术实现不得裸调 `exec.Command("ffuf", ...)`。工具调用统一走 `tools/` 适配层，职责包括：可执行文件解析、参数构造、context 与超时、stdout/stderr、退出码、结构化输出、错误处理。

**安全要求**

- 执行必须显式且限定目标范围，不得静默扩大目标。
- 禁止执行任意 shell 字符串；使用 `exec.CommandContext(ctx, executable, args...)`，不得用 `sh -c "..."`。
- 禁止将用户输入拼接入 shell 命令。
- 使用 context 取消与超时。
- 错误必须正常返回，不得吞掉 stderr，关键执行失败不得静默继续。

**Go 工程约束**

- 优先标准库：`context`、`encoding/json`、`errors`、`fmt`、`io`、`os`、`os/exec`、`path/filepath`、`strings`、`sync`、`time`。
- 避免不必要依赖；不引入框架。
- 接口保持最小化；用依赖注入提升可测性。
- 文件按职责拆分，禁止将大量技术堆进单个大文件。

## 6. 数据契约

### catalog/tactics.json

每个条目代表一个战术，`techniques` 数组引用该战术下的技术 ID。

```json
{
  "id": "BB05",
  "name": "Web 应用枚举",
  "description": "发现 Web 应用的路由、资源、参数和接口。",
  "techniques": ["BB05.001", "BB05.003"]
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 战术编号（`BBxx`），主键 |
| `name` | string | 战术名称，中文 |
| `description` | string | 战术描述，中文 |
| `techniques` | array[string] | 该战术下的技术 ID 列表 |

### catalog/techniques.json

每条技术声明其执行器、执行模式、依赖工具与 ATT&CK 映射。

```json
{
  "id": "BB10.001",
  "name": "SQL 注入",
  "tactic_id": "BB10",
  "description": "评估授权应用输入中的 SQL 注入漏洞。",
  "executor": "sql-injection",
  "mode": "active",
  "tools": ["sqlmap"],
  "mitre": ["T1190"],
  "tags": ["injection", "sql"]
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 技术编号（`BBxx.xxx`），主键 |
| `name` | string | 技术名称，中文 |
| `tactic_id` | string | 所属战术编号 |
| `description` | string | 技术描述，中文 |
| `executor` | string | 执行器标识，映射到 `internal/technique/` 实现 |
| `mode` | string | 执行模式：`passive` / `active` / `manual` |
| `tools` | array[string] | 依赖的外部工具名 |
| `mitre` | array[string] | MITRE ATT&CK ID，仅元数据 |
| `tags` | array[string] | 分类标签 |

### configs/redteam.json

运行配置，工具路径必须在此声明，禁止硬编码进代码。

```json
{
  "tools": {
    "ffuf": "/usr/local/bin/ffuf",
    "nmap": "/usr/local/bin/nmap",
    "nuclei": "/usr/local/bin/nuclei",
    "httpx": "/usr/local/bin/httpx",
    "subfinder": "/usr/local/bin/subfinder",
    "sqlmap": "/usr/local/bin/sqlmap"
  }
}
```

## 7. 已知问题与待办

以下问题如实记录，便于后续按模块处理。本文档仅记录，不在本文件范围内修复。

| 问题 | 位置 | 影响 | 处理建议 |
|---|---|---|---|
| `package` 声明非 `main` | 两个 `cmd/**/main.go` | `go build` 无法通过 | 修正为 `package main`，并接入 CLI 入口 |
| 目录拼写错误 | `tools/fuff/` | 与工具 `ffuf` 不对应 | 重命名为 `tools/ffuf/` |
| 空壳文件 | `internal/app/`、`tools/*/` | 无实际功能 | 按 §3 分层实现 |
| 无 CLI 实现 | `cmd/mitre_red_team/` | 无法执行任何命令 | 实现 `--tactic` / `--technique` / `--target` 解析 |
| 无测试 | `tests/` 缺失 | 无回归保障 | 补充单元与集成测试 |
| 无配置 | `configs/` 缺失 | 工具路径不可配置 | 补充 `configs/redteam.json` |
| `README.md` 为空 | 根目录 | 项目无说明 | 补充安装、用法、架构说明 |

## 8. 变更纪律

适用于后续所有 AI / 开发者对仓库的改动：

1. **盘点优先**：动手前先对照 §2 的权威结构，识别已有实现，不重复造轮子。
2. **禁止重建**：不从零新建项目，不替换既有目录结构。
3. **禁止删除**：不删除已有文件。
4. **禁止无理由重命名**：除非确有需要（如 §7 中的拼写错误），不改动既有包名与路径。
5. **禁止为重构而重写**：不因架构偏好重写可用代码。
6. **先计划后实现**：结构层面的变更必须先产出集成计划并与维护者确认，再写实现代码。
