# 添加外部工具

在 `tools/` 下新增一个外部安全工具适配层，供技术实现通过统一接口调用外部工具。

## 前置检查

动手前先阅读并对照：

- `.trae/rules/dir_structure.md`：工具层职责与分层约定
- `.trae/rules/code_structure.md`：命名与注释规范
- `.trae/rules/security_structure.md`：命令执行、内存与安全约束
- `tools/` 现有适配器（`nmap`、`nuclei`、`ffuf`）的写法，作为参照
- `configs/redteam.json`：工具路径配置，未创建时按约定创建

## 适配层职责

工具适配层必须处理：

- 可执行文件解析与配置路径读取
- 命令参数构造（以切片传递，不做 shell 解释）
- context 取消与超时
- stdout / stderr 处理（stderr 不得吞掉）
- 退出码判断
- 必要时的结构化输出解析

## 编码要求

- 工具名与目录名一致（小写）；目录已存在时复用，禁止重复创建。
- 参数构造是适配层的核心：参数名与取值来自配置或调用方，不拼接用户输入。
- 执行方式统一为：

```go
command := exec.CommandContext(ctx, executablePath, arguments...)
```

- 解析工具输出前先限制读取上限，防止无界增长（依据 `security_structure.md` 内存约束）。
- 新增工具的路径在 `configs/redteam.json` 的 `tools` 段登记，禁止硬编码。

## 验证

```bash
gofmt -l .
go build ./...
go test -v ./...
go vet ./...
```

已安装 `golangci-lint` 时追加 `golangci-lint run`。
