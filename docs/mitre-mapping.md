# MITRE ATT&CK 工具映射与执行工作流

这篇文章讲 MITRERedTeam 怎么把 MITRE ATT&CK 的战术和技术映射起来、再拿去跑工具。映射数据存在 `catalog/techniques.json` 的 `mitre` 字段里，catalog 层负责解析，engine 层负责执行。

## 1. 映射机制

每条技术（`BBxx.xxx`）在 `catalog/techniques.json` 中声明其对应的 MITRE ATT&CK ID 与依赖工具：

```json
{
  "id": "BB02.005",
  "name": "端口发现",
  "executor": "port-discovery",
  "mode": "active",
  "tools": ["nmap"],
  "mitre": ["T1046"]
}
```

| 字段 | 含义 |
|---|---|
| `mitre` | 该技术对应的 MITRE ATT&CK 技术 ID（如 `T1046` 网络服务发现） |
| `tools` | 执行该技术所需的外部安全工具 |
| `executor` | 映射到 `internal/technique/` 注册表中的 Go 实现 |

一个 MITRE ID 可以对应多条技术，一条技术也能声明多个 MITRE ID。MITRE ID 在这里**只是元数据和查询入口**，真正的主标识还是 `BBxx.xxx`（见 `dir_structure.md`）。

## 2. 组件职责

| 组件 | 职责 |
|---|---|
| `catalog/techniques.json` | 定义工具-TTP 映射数据（配置机制） |
| `internal/catalog` `TechniquesByMitreID()` | 检测组件：从 MITRE ID 反查对应技术 |
| `internal/technique` 注册表 | 按 `executor` 解析到实际实现 |
| `internal/engine` `ExecuteByMitre()` | 执行引擎：查询 → 解析 → 执行关联工具 |
| `tools/` 适配层 | 调用外部工具（路径来自 `configs/redteam.json`） |

## 3. 执行工作流

```
CLI --mitre T1046
   │
   ▼
catalog.TechniquesByMitreID("T1046")
   │   → [BB02.005 端口发现, BB05.003 端点发现]
   ▼
engine.ExecuteByMitre()
   │   逐个技术：technique.Get(executor)
   │   未注册的跳过（规划中技术）
   ▼
工具适配层（nmap / ffuf / httpx）
   │   exec.CommandContext + 超时
   ▼
[]model.ExecutionResult 输出
```

## 4. 使用方式

```bash
# 按 MITRE ATT&CK 技术 ID 触发
go run ./cmd/mitre_red_team --url https://example.com --mitre T1046

# 与按技术/战术触发并行：三选一
go run ./cmd/mitre_red_team --url https://example.com --technique BB05.001
go run ./cmd/mitre_red_team --url https://example.com --tactic BB05
```

执行要求：

- `--url` 目标必填。
- `--technique` / `--tactic` / `--mitre` 三者选一。
- 目标工具得先装好，路径在 `configs/redteam.json` 里声明；没装的话命令会明确报错，不会闷头继续。
- 这是授权范围内的**手动触发**工具，不会自动展开，也不会悄悄扩大目标。

## 5. 维护映射

新增或调整 MITRE 映射，改 `catalog/techniques.json` 里的 `mitre` 字段就行，然后跑一下校验确认数据没写错：

```bash
go test ./test/ -run TestTechniquesByMitreID -v
```

MITRE ATT&CK 版本更新时（比如新增、合并了技术 ID），跟着改 `mitre` 字段即可，代码一个字都不用动——数据驱动的好处就在这。
