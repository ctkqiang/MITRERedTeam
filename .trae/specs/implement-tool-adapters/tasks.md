# Tasks

- [x] Task 1: 重命名 ffuf 工具目录与文件，同步修正包名与文件内容
- [x] Task 2: 新增 `tools/interface.go`：`Result` 类型 + `Runner` 执行器（`exec.CommandContext` + 超时 + 切片参数 + stdout/stderr/退出码）
- [x] Task 3: 实现 `tools/nmap/nmap.go`：`Scanner` 适配器（`New` + `Scan`）
- [x] Task 4: 实现 `tools/ffuf/ffuf.go`：`Fuzzer` 适配器（`New` + `Fuzz`）
- [x] Task 5: 实现 `tools/nuclei/nuclei.go`：`Scanner` 适配器（`New` + `Scan`）
- [x] Task 6: 新增 `test/tools_test.go`：假命令 mock（成功输出、非零退出码 + stderr、超时返回错误）
- [x] Task 7: 全量验证：`make all` + `make lint` 全绿

# Task Dependencies

- Task 2 依赖 Task 1（接口与目录整理相互独立，可并行）。
- Task 3/4/5 依赖 Task 2（适配器复用 `Runner`），三者可并行。
- Task 6 依赖 Task 2/3/4/5。
- Task 7 依赖全部。
