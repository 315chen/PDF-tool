# PDF合并工具 Makefile

.PHONY: build test clean install deps help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := pdf-merger
CMD_DIR := ./cmd/pdfmerger
BUILD_FLAGS := -ldflags="-s -w"
TEST_FLAGS := -v -race -coverprofile=coverage.out

# 检测操作系统
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    OUTPUT_NAME := $(APP_NAME)-linux
endif
ifeq ($(UNAME_S),Darwin)
    OUTPUT_NAME := $(APP_NAME)-mac
endif
ifeq ($(OS),Windows_NT)
    OUTPUT_NAME := $(APP_NAME).exe
endif

# 默认输出名称
ifndef OUTPUT_NAME
    OUTPUT_NAME := $(APP_NAME)
endif

## help: 显示帮助信息
help:
	@echo "PDF合并工具构建系统"
	@echo ""
	@echo "可用命令:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## deps: 下载并整理依赖
deps:
	@echo "下载依赖包..."
	go mod download
	go mod tidy

## test: 运行所有测试
test:
	@echo "运行测试..."
	go test $(TEST_FLAGS) ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage: test
	@echo "生成覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

## build: 构建应用程序
build: deps
	@echo "构建应用程序..."
	go build $(BUILD_FLAGS) -o $(OUTPUT_NAME) $(CMD_DIR)
	@echo "构建完成: $(OUTPUT_NAME)"

## build-all: 为所有平台构建
build-all: deps
	@echo "为所有平台构建..."
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(APP_NAME)-linux $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(APP_NAME)-mac $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(APP_NAME).exe $(CMD_DIR)
	@echo "所有平台构建完成"

## build-release: 构建发布版本（优化）
build-release: deps
	@echo "构建发布版本..."
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -a -installsuffix cgo -o $(OUTPUT_NAME) $(CMD_DIR)
	@echo "发布版本构建完成: $(OUTPUT_NAME)"

## install: 安装到系统路径
install: build
	@echo "安装到系统路径..."
	go install $(CMD_DIR)
	@echo "安装完成"

## clean: 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f $(APP_NAME) $(APP_NAME)-* *.exe
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "清理完成"

## fmt: 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "代码格式化完成"

## lint: 运行代码检查
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
		echo "安装方法: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## dev: 开发模式运行
dev: deps
	@echo "开发模式运行..."
	go run $(CMD_DIR)

## mod-update: 更新所有依赖到最新版本
mod-update:
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy

## size: 显示构建文件大小
size: build
	@echo "构建文件大小:"
	@ls -lh $(OUTPUT_NAME) | awk '{print $$5 "\t" $$9}'

## info: 显示项目信息
info:
	@echo "项目信息:"
	@echo "  名称: PDF合并工具"
	@echo "  Go版本: $(shell go version)"
	@echo "  操作系统: $(UNAME_S)"
	@echo "  输出文件: $(OUTPUT_NAME)"
	@echo "  模块路径: $(shell go list -m)"