# 添加技术

在 `catalog/techniques.json` 中新增一条技术（technique），并在技术注册表中挂载对应实现。技术是具体的安全测试动作，编号形如 `BB05.001`。

## 前置检查

动手前先阅读并对照：

- `.trae/rules/dir_structure.md`：分层职责与 executor 注册模式
- `.trae/rules/code_structure.md`：命名与注释规范
- `.trae/rules/security_structure.md`：命令执行、内存与安全约束
- `internal/model/techniques.go`：Technique 模型，字段语义的权威起点
- `catalog/techniques.json`：现有技术与编号，确认不重复
- 同战术下相似技术的实现，作为写法参照
- `tools/`：所需外部工具的适配层是否已存在

## 数据要求

`techniques.json` 中每条技术包含：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 技术编号，格式 `BBxx.xxx`，取所属战术内可用序号 |
| `name` | string | 技术名称，中文 |
| `tactic_id` | string | 所属战术编号，必须真实存在于 `tactics.json` |
| `description` | string | 技术描述，中文 |
| `executor` | string | 执行器标识，映射到 `internal/technique/` 下的实现 |
| `mode` | string | `passive` / `active` / `manual` 之一 |
| `tools` | array[string] | 依赖的外部工具名，需有对应适配层 |
| `mitre` | array[string] | MITRE ATT&CK ID，仅元数据，可为空 |
| `tags` | array[string] | 分类标签 |

示例：

```json
{
  "id": "BB05.002",
  "name": "参数枚举",
  "tactic_id": "BB05",
  "description": "枚举授权 Web 目标的参数。",
  "executor": "parameter-enumeration",
  "mode": "active",
  "tools": ["ffuf"],
  "mitre": [],
  "tags": ["web", "parameter"]
}
```

## 实现要求

- catalog 定义与 Go 实现分离：数据进 `catalog/techniques.json`，实现进对应的 technique 包。
- 实现必须注册到技术注册表，通过 `executor` 名称解析；禁止巨型 switch 分发。
- 复用现有抽象与工具适配层；不得复制相似技术实现到新文件。
- 工具调用统一走 `tools/` 适配层，禁止在技术实现中直接执行 `exec.Command`。

## 安全要求

依据 `security_structure.md`：

- 外部命令使用 `exec.CommandContext` 并设置超时；禁止 shell 拼接与参数注入。
- 工具输出与外部输入先校验长度与边界，再进入逻辑。
- 工具路径不得硬编码，统一由 `configs/redteam.json` 提供。

## 验证

```bash
gofmt -l .
go build ./...
go test -v ./...
go vet ./...
```

已安装 `golangci-lint` 时追加 `golangci-lint run`。
