# 架构说明

本文档说明项目的架构风格与高层数据流，作为理解代码库的上下文。与 `.trae/rules/` 下的规范文档配套：本文档回答"系统怎么组织与流转"，规则文档回答落地约束——`dir_structure.md` 管目录分层，`code_structure.md` 管代码写法，`security_structure.md` 管安全与内存。

## 1. 架构风格：CDRPA

项目采用**目录驱动的注册式插件架构**（Catalog-Driven Registry Plugin Architecture，CDRPA），组合以下五种模式：

| 模式 | 含义 |
|---|---|
| 目录驱动（Catalog-Driven） | 战术与技术清单由 `catalog/*.json` 数据定义，不写死在 Go 代码中 |
| 注册模式（Registry） | `executor` 标识通过注册表映射到具体 Go 实现 |
| 插件架构（Plugin） | 新增技术、工具集成时尽量不改动核心执行引擎 |
| 分层架构（Layered） | CLI、引擎、技术、工具各层职责分离，禁止越权 |
| 依赖倒置（Dependency Inversion） | 核心系统不依赖具体安全工具与具体技术实现，只依赖抽象 |

首要目标是可扩展性：新增战术、技术、工具集成时，应能在不修改核心执行引擎的前提下完成。核心系统必须保持与具体工具、具体技术实现相互独立。

## 2. 架构目标

架构设计用于支撑：

- 大量安全测试技术
- 漏洞赏金导向的安全工作流
- 适用场景下与 MITRE ATT&CK 对齐的技术
- 用户自行选择战术执行
- 用户自行选择技术执行
- 外部安全工具接入
- 数据驱动的技术元数据
- 相互独立的技术实现
- 相互独立的工具适配层
- 确定性执行
- 结构化结果
- 证据收集
- 发现（finding）生成
- 开源协作

同时必须避免：出现一个包含所有技术的巨型中心实现文件。这与 `dir_structure.md` 中"禁止巨型硬编码映射、数据驱动"的要求一致。

## 3. 分层职责与数据流

| 层 | 职责 |
|---|---|
| CLI | 解析用户请求（`--tactic` / `--technique` / `--target`），仅执行用户选定项 |
| Planner | 把用户请求解析为可执行的执行计划 |
| Engine | 编排执行：目标校验、目录查询、计划调度、结果汇总 |
| Technique Registry | 维护 `executor → 实现` 的注册表，按名称解析 |
| 领域技术 | 具体技术实现，按领域分目录（recon、enumeration、injection 等） |
| Capability | 领域技术之间可复用的原子能力层，位于技术与工具之间 |
| Tool 适配层 | 封装外部工具：路径解析、参数构造、超时、输出解析 |
| Result / Evidence / Finding | 执行结果、证据与最终发现的递进产出 |

工具层落地路径为 `tools/` 目录，工具路径统一由 `configs/redteam.json` 配置提供（详见 `dir_structure.md`）。技术实现不直接执行命令，统一经由工具适配层，安全与超时约束见 `security_structure.md`。

## 4. 高层架构图

```text
                              CLI
                               │
                               ▼
                         User Request
                               │
                               ▼
                         ┌──────────┐
                         │ Planner  │
                         └────┬─────┘
                              │
                              ▼
                       Execution Plan
                              │
                              ▼
                         ┌──────────┐
                         │  Engine  │
                         └────┬─────┘
                              │
                              ▼
                       Technique Registry
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
        Recon Technique  Web Technique  Injection Technique
              │               │               │
              └───────────────┼───────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
               Capability              Tool
                    │                   │
                    │         ┌─────────┼─────────┐
                    │         ▼         ▼         ▼
                    │       ffuf       nmap     nuclei
                    │
                    └─────────┬─────────┘
                              ▼
                           Result
                              │
                              ▼
                          Evidence
                              │
                              ▼
                           Finding
                              │
                              ▼
                           Output
```

完整链路：用户通过 CLI 提交请求 → Planner 生成执行计划 → Engine 校验目标并调度 → Technique Registry 按 `executor` 解析出领域技术实现 → 技术实现经 Capability 或 Tool 适配层调用外部工具（ffuf、nmap、nuclei 等）→ 产出结构化 Result → 沉淀为 Evidence → 汇总为 Finding → 由 Output 层输出。

## 5. 与规则文档的关系

- `dir_structure.md`：目录与分层规范，本架构的落地约束（`internal/` 分层、`tools/` 路径、注册模式）。
- `code_structure.md`：命名、注释与格式规范，适用于所有实现代码。
- `security_structure.md`：安全与内存约束（`exec.CommandContext`、超时、RSS 500MB 上限、cgo 准入）。
- `prompts/`：面向具体开发任务（添加战术、技术、工具，安全审查）的可执行提示，均在上述架构与规则约束下工作。
