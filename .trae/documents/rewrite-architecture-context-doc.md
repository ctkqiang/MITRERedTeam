# 重写 `.trae/context/architecture.md` 计划

## Summary

将英文的 `.trae/context/architecture.md` 重写为中文专业架构说明文档，保持其「架构上下文参考」定位（区别于 `rules/` 下的强制性规范），并与已就绪的 `dir_structure.md`、`code_structure.md`、`security_structure.md` 三份规则文档衔接一致。风格沿用此前标准：中文书写、术语保留英文、去 AI 味、内容准确。

本计划只改一个文件：`.trae/context/architecture.md`。

## Current State Analysis（Phase 1 探索结论）

### 现有内容（97 行英文，已完整读取）

| 章节 | 内容 | 评价 |
|---|---|---|
| Architecture Style | CDRPA（Catalog-Driven Registry Plugin Architecture），组合 5 种模式，首要目标可扩展性 | 核心概念准确，保留 |
| Architectural Goals | 14 项能力目标 + 禁止巨型中心文件 | 清单有价值，需中文化 |
| High-Level Architecture | ASCII 图：CLI → Planner → Engine → Technique Registry → 领域技术 → Capability/Tool → ffuf/nmap/nuclei → Result → Evidence → Finding → Output | 结构清晰，保留并加中文标注 |

### 与现有规则文档的一致性检查

- `dir_structure.md` 分层（engine/technique/tool）与架构图一致 ✓
- executor 注册模式、禁止巨型 switch 与规则一致 ✓
- 工具路径 `tools/`（`fuff` 拼写错误已在 `dir_structure.md` 标注）与图中 ffuf/nmap/nuclei 一致 ✓
- 图中有 `Capability` 概念，规则文档未提及——保留并在文档中说明其位置（技术可复用的原子能力）
- 安全约束（`exec.CommandContext`、超时、500MB 内存）在 `security_structure.md` 已有，本文档只引用不重复

### 定位差异

`context/architecture.md` 是给 AI/开发者理解系统组织的**参考文档**；`rules/dir_structure.md` 是强制性的**现状+目标规范**。重写后保持二者分工：本文档讲"系统怎么组织与流转"，规则文档讲"落地约束"。

## Proposed Changes

### 唯一改动文件

`.trae/context/architecture.md` —— 全量重写（保留文件路径与定位）。

### 新文档章节大纲（5 节，中文，去 AI 味）

1. **架构风格：CDRPA**
   - 说明组合的 5 种模式（目录驱动、注册、插件、分层、依赖倒置），每条 1 句定义。
   - 明确首要目标：可扩展性——新增战术/技术/工具集成不改动核心执行引擎；核心系统不依赖具体工具与技术实现。

2. **架构目标**
   - 保留原 14 项能力目标，中文表述，列成清单。
   - 明确避免：出现包含所有技术的巨型中心文件（与 `dir_structure.md`"禁止巨型硬编码映射"呼应）。

3. **分层职责与数据流**
   - 表格：CLI / Planner / Engine / Technique Registry / 领域技术 / Tool 适配层 / Result-Evidence-Finding-Output，每层职责 1–2 句。
   - 说明 `Capability` 概念：领域技术可复用的原子能力层，位于技术与工具之间。
   - 说明工具层落地路径：`tools/` 目录 + `configs/redteam.json` 配置（引用 `dir_structure.md`）。

4. **高层架构图**
   - 保留原 ASCII 图，组件名保持英文（与未来 Go 包名一致），关键节点加中文说明注释。
   - 配一段中文文字说明完整链路：用户请求 → 计划 → 注册表解析 → 技术执行 → 工具调用 → 结构化结果。

5. **与规则文档的关系**
   - 引用 `rules/` 三份文档的职责分工：本文档说明组织方式，`dir_structure.md` 管目录分层，`code_structure.md` 管写法，`security_structure.md` 管安全与内存。

### 写作要求

- 中文书写；专有名词保留英文：CDRPA、Planner、Engine、Registry、executor、Capability、MITRE ATT&CK、CLI、JSON、RSS。
- 无分割线、无步骤式表述、无 AI 味套话；语气客观陈述。
- 架构图内 ASCII 边框保留，图中注释用中文。
- 不与三份规则文档冲突；引用其条款而非重复。

## Assumptions & Decisions

1. **文档语言为中文**（与用户偏好及现有规则文档一致），术语保留英文。
2. **保留 CDRPA 架构风格命名与整体结构**，只做中文化与专业化重写，不发明新架构。
3. **保留 ASCII 架构图**并加中文标注；组件名保留英文以便与未来包名对应。
4. **保留 `Capability` 概念**并在文档中给出定位说明，不与规则文档冲突。
5. **文档定位为参考文档**：不改写为强制性规范，不并入 `rules/`。

## Verification

1. 重读 `.trae/context/architecture.md`，核对：
   - CDRPA 五模式、14 项目标、完整架构图均保留；
   - 与 `dir_structure.md` 分层（engine/technique/tool）、`tools/` 路径、executor 注册模式无冲突；
   - 无分割线/步骤式/AI 味表述。
2. `git status` 确认除 `.trae/context/architecture.md` 外无其他文件变更。
