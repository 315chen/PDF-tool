#!/bin/bash

# PDF合并工具命令行版本构建脚本
# 用于构建不依赖GUI的跨平台版本

set -e

echo "=== PDF合并工具命令行版本构建 ==="
echo ""

# 版本信息
VERSION="v1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# 创建发布目录
RELEASE_DIR="releases/${VERSION}"
mkdir -p "${RELEASE_DIR}"

echo "版本: ${VERSION}"
echo "构建时间: ${BUILD_TIME}"
echo "Git提交: ${GIT_COMMIT}"
echo "发布目录: ${RELEASE_DIR}"
echo ""

# 构建函数
build_cli_platform() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=$3
    local PLATFORM_NAME=$4
    
    echo "构建 ${PLATFORM_NAME} 命令行版本 (${GOOS}/${GOARCH})..."
    
    # 禁用CGO以实现真正的跨平台构建
    export CGO_ENABLED=0
    
    # 构建
    if GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${RELEASE_DIR}/pdf-merger-cli-${PLATFORM_NAME}${EXT}" \
        ./cmd/pdfmerger-cli; then
        
        echo "  ✓ 构建成功: pdf-merger-cli-${PLATFORM_NAME}${EXT}"
        
        # 获取文件大小
        if [ -f "${RELEASE_DIR}/pdf-merger-cli-${PLATFORM_NAME}${EXT}" ]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                SIZE=$(stat -f%z "${RELEASE_DIR}/pdf-merger-cli-${PLATFORM_NAME}${EXT}")
            else
                SIZE=$(stat -c%s "${RELEASE_DIR}/pdf-merger-cli-${PLATFORM_NAME}${EXT}")
            fi
            echo "  文件大小: $((SIZE / 1048576))MB"
        fi
        return 0
    else
        echo "  ✗ 构建失败: pdf-merger-cli-${PLATFORM_NAME}${EXT}"
        return 1
    fi
    
    echo ""
}

# 构建所有平台的命令行版本
echo "开始构建命令行版本..."
echo ""

SUCCESSFUL_BUILDS=()
FAILED_BUILDS=()

# Windows 64位
if build_cli_platform "windows" "amd64" ".exe" "windows-64bit"; then
    SUCCESSFUL_BUILDS+=("Windows 64位 CLI")
else
    FAILED_BUILDS+=("Windows 64位 CLI")
fi

# Windows 32位
if build_cli_platform "windows" "386" ".exe" "windows-32bit"; then
    SUCCESSFUL_BUILDS+=("Windows 32位 CLI")
else
    FAILED_BUILDS+=("Windows 32位 CLI")
fi

# Linux 64位
if build_cli_platform "linux" "amd64" "" "linux-64bit"; then
    SUCCESSFUL_BUILDS+=("Linux 64位 CLI")
else
    FAILED_BUILDS+=("Linux 64位 CLI")
fi

# Linux ARM64
if build_cli_platform "linux" "arm64" "" "linux-arm64"; then
    SUCCESSFUL_BUILDS+=("Linux ARM64 CLI")
else
    FAILED_BUILDS+=("Linux ARM64 CLI")
fi

# macOS Intel (命令行版本)
if build_cli_platform "darwin" "amd64" "" "macos-intel"; then
    SUCCESSFUL_BUILDS+=("macOS Intel CLI")
else
    FAILED_BUILDS+=("macOS Intel CLI")
fi

# macOS Apple Silicon (命令行版本)
if build_cli_platform "darwin" "arm64" "" "macos-apple-silicon"; then
    SUCCESSFUL_BUILDS+=("macOS Apple Silicon CLI")
else
    FAILED_BUILDS+=("macOS Apple Silicon CLI")
fi

echo ""
echo "=== 构建结果 ==="
echo ""

echo "✅ 成功构建的平台:"
for platform in "${SUCCESSFUL_BUILDS[@]}"; do
    echo "  - $platform"
done

if [ ${#FAILED_BUILDS[@]} -gt 0 ]; then
    echo ""
    echo "❌ 构建失败的平台:"
    for platform in "${FAILED_BUILDS[@]}"; do
        echo "  - $platform"
    done
fi

# 更新校验和文件
echo ""
echo "更新校验和文件..."
cd "${RELEASE_DIR}"
if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 pdf-merger-* > checksums.sha256 2>/dev/null || true
    echo "  ✓ SHA256校验和已更新"
fi
cd - > /dev/null

# 显示最终结果
echo ""
echo "发布文件位置: ${RELEASE_DIR}/"
ls -la "${RELEASE_DIR}/" | grep pdf-merger
echo ""

echo "✅ 命令行版本构建完成！"
echo ""
echo "使用说明:"
echo "- GUI版本: pdf-merger-macos-* (仅macOS)"
echo "- CLI版本: pdf-merger-cli-* (所有平台)"
echo "- CLI用法: ./pdf-merger-cli-* -input file1.pdf,file2.pdf -output merged.pdf"
