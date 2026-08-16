# Checklist

- [x] `tools/ffuf/` 目录与文件已重命名，全库无旧拼写残留引用
- [x] `Runner` 使用 `exec.CommandContext`，参数切片传递，无 shell 拼接
- [x] `Runner` 设置超时，超时返回错误
- [x] `Runner` 捕获 stdout 与 stderr 并记录退出码，stderr 未被吞
- [x] `nmap` / `ffuf` / `nuclei` 适配器经构造器接收可执行路径，无硬编码路径
- [x] `test/tools_test.go` 覆盖成功输出、非零退出码、超时三种场景
- [x] 全量门禁 `make all` + `make lint` 通过
