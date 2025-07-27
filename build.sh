#!/bin/bash

# PDF合并工具构建脚本

set -e

echo "开始构建PDF合并工具..."

# 检查Go是否已安装
if ! command -v go &> /dev/null; then
    echo "错误: Go语言未安装。请先安装Go语言环境。"
    echo "访问 https://golang.org/dl/ 下载安装Go"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "警告: 建议使用Go $REQUIRED_VERSION 或更高版本，当前版本: $GO_VERSION"
fi

# 下载依赖
echo "下载依赖包..."
go mod tidy

# 运行测试
echo "运行测试..."
go test ./...

# 构建应用程序
echo "构建应用程序..."
OUTPUT_NAME="pdf-merger"

# 根据操作系统设置输出文件名
case "$(uname -s)" in
    Darwin*)    OUTPUT_NAME="pdf-merger-mac" ;;
    Linux*)     OUTPUT_NAME="pdf-merger-linux" ;;
    CYGWIN*|MINGW32*|MSYS*|MINGW*) OUTPUT_NAME="pdf-merger.exe" ;;
esac

# 构建
go build -ldflags="-s -w" -o "$OUTPUT_NAME" ./cmd/pdfmerger

echo "构建完成！"
echo "可执行文件: $OUTPUT_NAME"
echo ""
echo "使用方法:"
echo "  ./$OUTPUT_NAME"
echo ""
echo "注意: 首次运行时，Fyne可能需要下载额外的系统依赖。"