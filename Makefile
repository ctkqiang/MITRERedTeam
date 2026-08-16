# MITRERedTeam 常用开发命令
#
# 用法示例：
#   make fmt      检查代码格式（gofmt -l 必须为空）
#   make vet      静态分析
#   make build    编译（排除 test 目录）
#   make test     运行全部测试
#   make lint     golangci-lint 静态检查（未安装时跳过）
#   make vuln     依赖漏洞扫描（govulncheck，未安装时跳过）
#   make clean    清理构建缓存
#   make all      完整质量门禁：fmt + vet + build + test

.PHONY: help run all fmt vet build test lint vuln clean

# 构建目标排除 test 目录：该目录仅存放 *_test.go，不产生构建产物
BUILD_TARGETS := $(shell go list ./... | grep -v '/test$$')

# 显示可用目标与说明（make 无参数时默认执行此目标）
help:
	@echo "MITRERedTeam 常用开发命令："
	@echo ""
	@echo "  make fmt      检查代码格式（gofmt -l 必须为空）"
	@echo "  make vet      静态分析"
	@echo "  make build    编译（排除 test 目录）"
	@echo "  make run      运行 CLI（可传参，如 make run ARGS=\"--technique BB05.001\")"
	@echo "  make test     运行全部测试"
	@echo "  make lint     golangci-lint 静态检查（未安装时跳过）"
	@echo "  make vuln     依赖漏洞扫描（govulncheck，未安装时跳过）"
	@echo "  make clean    清理构建缓存"
	@echo "  make all      完整质量门禁：fmt + vet + build + test"

# 完整质量门禁，与 .githooks/pre-commit 保持一致
all: fmt vet build test

# 运行 CLI。参数通过 ARGS 传递，如 make run ARGS="--tactic BB05"
run:
	go run ./cmd/mitre_red_team $(ARGS)

# 格式化检查：gofmt -l 输出为空才算通过
fmt:
	@if [ -n "$$(gofmt -l .)" ]; then \
		gofmt -l .; \
		echo "==> 存在未格式化文件，请运行 go fmt ./..."; \
		exit 1; \
	fi
	@echo "==> gofmt 检查通过"

# 静态分析
vet:
	go vet ./...
	@echo "==> go vet 通过"

# 编译（排除 test 目录）
build:
	go build $(BUILD_TARGETS)
	@echo "==> go build 通过"

# 单元测试
test:
	go test -v ./...
	@echo "==> go test 通过"

# 静态代码检查（golangci-lint 未安装时跳过）
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "==> golangci-lint 未安装，跳过 lint"; \
	fi

# 依赖漏洞扫描（govulncheck 未安装时跳过）
vuln:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "==> govulncheck 未安装，跳过漏洞扫描"; \
	fi

# 清理构建缓存
clean:
	go clean
