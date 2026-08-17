---
alwaysApply: true
scene: code_structure
---
# MITRERedTeam Go 代码结构与编码规范

本文件约束仓库内所有 Go 代码的写法。与 `dir_structure.md` 配套使用：后者定义目录与分层，本文档定义文件内部的组织、命名、注释与格式。`.githooks/pre-commit` 负责自动化门禁，本文档是门禁之外需要人工遵循的规范。

## 1. 适用范围与定位

- 适用于 `cmd/`、`internal/`、`tools/` 下所有 Go 代码。
- 与 `dir_structure.md` 冲突时，以"代码能通过 gofmt 与 go vet"为最低门槛，架构问题遵循 `dir_structure.md`。
- 文档注释面向读代码的人，目标是让人不看代码也能理解意图。

## 2. Go 版本与工具链

- 项目目标版本为 `go 1.26`（见 `go.mod`），新增代码不得使用低于该版本的语法特性。
- 提交前必须通过四道门禁（`.githooks/pre-commit` 已强制）：
  1. `gofmt -l` 输出为空；
  2. `go build ./...` 无错误；
  3. `go test -v ./...` 全部通过；
  4. 若已安装 `golangci-lint`，`golangci-lint run` 无错误。
- 优先使用标准库；新增第三方依赖必须说明理由并在 PR 中体现。
- 常规工作流：改完代码依次执行 `gofmt -l` → `go vet ./...` → `go test ./...` → `golangci-lint run`。

## 3. 命名规范

### 变量

变量名必须完整、有意义、能自解释用途。禁止使用单字母或难以辨认的缩写。

```go
// 反例
v := "example.com"      // v 是什么？
ret := parseTarget(url) // ret 无法说明返回内容
tmp := strings.Split(s, "/") // tmp 用途不明

// 正例
targetHost := "example.com"
parsedTarget, err := parseTarget(requestURL)
pathSegments := strings.Split(requestURL.Path, "/")
```

惯例例外仅限 Go 社区广泛接受的极小作用域用法，如 `for` 循环内自增的循环变量，且不得扩散到函数级其他变量：

```go
for index := range pathSegments { // 优先完整命名
	// ...
}

// 仅当循环体极短且上下文中语义唯一时，可接受惯例写法
for i := 0; i < len(pathSegments); i++ {
	_ = pathSegments[i]
}
```

### 常量

导出常量使用 MixedCaps，与 Go 惯例一致，不使用 ALL_CAPS。未导出的包级常量同样遵循驼峰。

```go
const (
	TechniquePassive      TechniqueMode = "passive"
	DefaultTimeoutSeconds               = 30
)
```

### 函数与方法

函数名用动作性名词短语，MixedCaps。返回布尔值的函数用 `is`、`has`、`can` 等前缀体现判断语义。

```go
func LoadCatalog(path string) (*Catalog, error)
func ValidateTarget(target Target) error
func IsInScope(host string, scope []string) bool
```

### 类型与结构体

类型名为名词，结构体字段名与 JSON tag 保持语义对应。参考 `internal/model/` 的既有写法：

```go
type Target struct {
	ID       string            `json:"id"`
	Host     string            `json:"host"`
	Port     int               `json:"port,omitempty"`
	Scheme   string            `json:"scheme,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

### 包名

包名使用小写单词，不加下划线、不加复数。`internal/` 下的包名遵循 `dir_structure.md` 的分层命名。

### 缩写词

`ID`、`URL`、`API`、`HTTP` 等缩写词在命名中保持大小写一致，不混用 `Id`、`Url` 等变体。全大写或全小写二选一，全项目统一。

## 4. 注释规范

### 基本要求

所有注释使用中文，自然、人性化，像同事之间交流那样讲清楚。注释要解释参数含义、返回值说明与实现逻辑，而不是复述代码本身。

### 导出标识符

导出的类型、函数、常量、字段必须有文档注释，以标识符本身开头，便于 `go doc` 生成说明。

```go
// ParseTarget 从 rawTarget 中解析出协议、主机与端口。
// rawTarget 支持形如 https://example.com:8443 的输入。
// 返回 Target 结构；协议或主机缺失时返回错误。
func ParseTarget(rawTarget string) (Target, error) {
	// ...
}
```

### 函数注释结构

按"用途 → 关键参数 → 返回值 → 错误行为 → 副作用"的顺序组织，不要求全部出现，只写有信息量的部分。

```go
// RunCatalogLoad 加载并校验 catalog 目录数据。
// 入参 path 指向 tactics.json 所在目录，会同时读取 techniques.json。
// 校验失败（ID 冲突、引用悬空）时返回 error，catalog 不会部分生效。
func RunCatalogLoad(path string) (*Catalog, error)
```

### 实现逻辑注释

解释"为什么"和思路，放在难以一眼看懂的代码附近：复杂分支、并发、偏移计算、边界处理。

```go
// 先解析主机再回退默认端口，避免 DNS 解析失败时直接报错
// 导致用户无法获知主机本身是否有效。
host := parsedTarget.Host
if parsedTarget.Port == 0 {
	host = net.JoinHostPort(parsedTarget.Host, "443")
}
```

### 禁止事项

- 分割线形式的注释：`// --------`、`// ========`、`// =====` 一律禁止。
- 步骤式注释：`// step 1`、`// 第一步：`、`// ===== steps =====` 一律禁止。
- AI 生成式套话：`该函数用于…`、`此代码段将…` 这类无信息量的句式。
- 注释与代码不一致：修改代码时必须同步更新相关注释。
- 过度注释：对 `x := 1` 这类自明语句加注释。

### 特殊注解

`@param`、`@return`、`pragma`、`nolint` 等特殊注解保留英文形式，不翻译。

```go
//nolint:gosec // 此处需要直接拼接参数，已确认无注入面
```

### TODO / FIXME

允许使用，但必须写明原因与上下文，方便后续处理。

```go
// TODO(redteam): 兼容 IPv6 字面量地址，当前实现只处理 IPv4。
```

## 5. 代码结构规范

### 文件组织

一个文件只承载一个清晰职责，不堆砌大文件。`internal/technique/` 下的实现按领域拆分子目录，与 `dir_structure.md` 保持一致。

### 测试文件

所有以 `_test.go` 结尾的 Go 测试文件必须放置在项目根目录的 `test/` 目录下，与被测包目录分离。

- 命名模式：`<被测对象>_test.go`，如 `logger_test.go`、`catalog_test.go`。
- 测试文件统一使用外部测试包 `package tests`（因 `test/` 目录受 Go"单目录单包"约束），只能访问被测包的导出 API；未导出成员的验证通过公开 API 的输出间接断言。

- `test/` 目录仅存放测试文件，不产生构建产物；构建门禁已排除该目录（见 `.githooks/pre-commit`）。
- 新增测试文件必须同时满足本文件其余规范：中文注释、完整命名、gofmt 合规、无第三方依赖。

### 错误处理

- 不忽略错误；不需要该返回值时显式丢弃并说明原因（`_` 或注释）。
- 用 `fmt.Errorf` 与 `%w` 包装错误，携带调用上下文，便于 `errors.Is` / `errors.As` 判断。
- 外部命令的 stderr 不得吞掉，必须透传或记录。

```go
configFile, err := os.ReadFile(configPath)
if err != nil {
	return nil, fmt.Errorf("读取配置文件 %s: %w", configPath, err)
}
```

### 函数设计

- 函数保持单一职责，长度以能一眼把握为限。
- 优先提前返回（early return）减少嵌套层级。

```go
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("配置路径不能为空")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// ...
}
```

### context 与超时

context 贯穿调用链，从 CLI 入口到工具执行不得中断。外部命令必须使用 `exec.CommandContext` 并设置超时。

```go
ctx, cancel := context.WithTimeout(parentContext, 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, executable, arguments...)
```

### 并发

- 优先标准库 `sync` 原语，避免引入额外并发库。
- 避免共享可变状态；并发写集合时先复制或使用同步原语。
- channel 的关闭与所有权必须有明确归属，防止发送方关闭后仍写入。

### 依赖注入

构造器显式接收依赖，避免包级全局变量。接口定义在使用方，保持最小。

```go
type CatalogLoader interface {
	Load(path string) (*Catalog, error)
}

func NewEngine(loader CatalogLoader, runner ToolRunner) *Engine {
	return &Engine{loader: loader, runner: runner}
}
```

### 反冗余与代码复用

代码出现重复时，不得复制粘贴了事。重复的信号包括：结构相同、仅参数或值不同的代码块；需要在多处同步修改的字面量与分支；逐处手写相同的查找或转换逻辑。对这些模式，用映射结构（map）、循环（for/range）或提取公共函数集中处理，保证增删修改只落在一处。

```go
// 反例：两处手写相同的线性查找，新增字段时需同步改两处
func findTactic(tactics []Tactic, id string) (Tactic, bool) {
	for _, tactic := range tactics {
		if tactic.ID == id {
			return tactic, true
		}
	}
	return Tactic{}, false
}

func findTechnique(techniques []Technique, id string) (Technique, bool) {
	for _, technique := range techniques {
		if technique.ID == id {
			return technique, true
		}
	}
	return Technique{}, false
}

// 正例：加载阶段统一建立 ID 索引，查找集中在一处且为 O(1)
type Catalog struct {
	byTacticID    map[string]Tactic
	byTechniqueID map[string]Technique
}

func (c *Catalog) index() {
	c.byTacticID = make(map[string]Tactic, len(c.tactics))
	for _, tactic := range c.tactics {
		c.byTacticID[tactic.ID] = tactic
	}
	// byTechniqueID 同理，循环遍历一次完成建表
}

func (c *Catalog) GetTactic(id string) (Tactic, bool) {
	tactic, found := c.byTacticID[id]
	return tactic, found
}
```

```go
// 反例：同一键值映射散落在各函数里，维护时容易漏改
if toolName == "nmap" {
	toolPath = "/usr/local/bin/nmap"
}
// ... 另一处
if toolName == "nmap" {
	toolPath = "/usr/local/bin/nmap"
}

// 正例：映射表集中声明，查表取代逐条判断
var toolDefaultPaths = map[string]string{
	"nmap":   "/usr/local/bin/nmap",
	"nuclei": "/usr/local/bin/nuclei",
}

func defaultToolPath(toolName string) string {
	return toolDefaultPaths[toolName]
}
```

注意边界：反冗余不等同于强行抽象。只有模式真正重复（出现两处及以上且会同步演化）才集中化；一次性或语义差异过大的相似代码，保留直白写法更清晰。

## 6. 格式规范

- 以 `gofmt` 输出为准，`gofmt -l` 必须为空。
- 缩进使用 Tab；导入按 标准库 / 第三方 / 本地模块 分组，组间空行分隔。

```go
import (
	"context"
	"fmt"
	"os/exec"

	"mitre_red_team/internal/model"
)
```

- 长行优先通过换行保持可读，不截断语义；结构体字段对齐交给 gofmt。
- `golangci-lint` 建议启用 `govet`、`staticcheck`、`ineffassign`、`unused` 等默认检查项。

### 可读性排版

gofmt 保证语法格式统一，但逻辑分段需要人工维护。具体要求：

- 用空行分隔不同职责的逻辑块与函数；同一函数内，变量准备、错误处理、结果汇总等段落之间用空行隔开，避免整段粘连。
- 嵌套结构保持一致的 Tab 缩进，每层一个 Tab，不得混用空格；分支层级必须能一眼分辨。
- 运算符、逗号、冒号两侧按 gofmt 规则留适当空白；关键表达式周围允许用空格提升可读性。
- 整体要求：单屏内能分辨函数边界与分支层级，段落分明、不粘连。

```go
// 反例：错误处理与正常逻辑挤在一起，段落无法分辨
results := engine.Execute(ctx, request)
if err != nil {
	return nil, err
}
summary := summarize(results)
return summary, nil

// 正例：错误处理独立成段，正常路径清晰可读
results, err := engine.Execute(ctx, request)
if err != nil {
	return nil, fmt.Errorf("执行失败: %w", err)
}

summary := summarize(results)
return summary, nil
```

## 7. 提交前自查清单

- 变量名完整且自解释，无单字母/难懂缩写。
- 导出的标识符有中文文档注释，非导出的关键逻辑有中文注释。
- 无分割线、步骤式、AI 味套话注释。
- 错误均被处理或显式说明，stderr 未被吞。
- 无重复代码段：重复模式已用映射结构、循环或公共函数集中，无散落的多处手写拷贝。
- 排版可读：不同职责的逻辑块之间有换行分隔，嵌套缩进一致，段落不粘连。
- `gofmt -l` 为空，`go vet` 无告警。
- `go test ./...` 与 `golangci-lint run` 通过。
