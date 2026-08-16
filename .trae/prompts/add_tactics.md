# 添加战术

在 `catalog/tactics.json` 中新增一个战术（tactic）条目。战术是相关技术的逻辑分组，编号形如 `BB05`。

## 前置检查

动手前先阅读并对照：

- `.trae/rules/dir_structure.md`：架构分层与目录约定
- `.trae/rules/code_structure.md`：命名与注释规范
- `catalog/tactics.json`：现有战术与编号，确认不重复
- `catalog/techniques.json`：现有技术及其所属战术

## 数据要求

`tactics.json` 中每个战术条目包含：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 战术编号，格式 `BBxx`，取现有最大编号 +1，禁止跳号与重复 |
| `name` | string | 战术名称，中文 |
| `description` | string | 战术描述，中文 |
| `techniques` | array[string] | 该战术下的技术 ID 列表，新增战术可先为空 |

示例：

```json
{
  "id": "BB26",
  "name": "示例战术",
  "description": "该战术的用途说明。",
  "techniques": []
}
```

## 规则

- 只修改 `catalog/tactics.json`，不触碰其他文件。
- 编号连续递增，不与现有战术重复。
- 名称与描述使用中文，风格与 `BB01`–`BB25` 保持一致。
- `techniques` 数组引用的每个技术 ID 必须真实存在于 `catalog/techniques.json`，禁止悬挂引用。

## 验证

```bash
python3 -c "import json; json.load(open('catalog/tactics.json'))"
```

确认 JSON 合法、编号唯一，且所有 `techniques` 引用在 `techniques.json` 中均有对应条目。
