#!/bin/bash

# PDF合并工具多平台构建脚本
# 专门用于构建可以成功构建的平台版本

set -e

echo "=== PDF合并工具多平台构建 ==="
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
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=$3
    local PLATFORM_NAME=$4
    
    echo "构建 ${PLATFORM_NAME} (${GOOS}/${GOARCH})..."
    
    # 设置环境变量
    export CGO_ENABLED=1
    
    # 构建
    if GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}" \
        ./cmd/pdfmerger 2>/dev/null; then
        
        echo "  ✓ 构建成功: pdf-merger-${PLATFORM_NAME}${EXT}"
        
        # 获取文件大小
        if [ -f "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}" ]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                SIZE=$(stat -f%z "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}")
            else
                SIZE=$(stat -c%s "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}")
            fi
            echo "  文件大小: $((SIZE / 1048576))MB"
        fi
        return 0
    else
        echo "  ✗ 构建失败: pdf-merger-${PLATFORM_NAME}${EXT}"
        return 1
    fi
    
    echo ""
}

# 尝试构建各平台版本
echo "开始多平台构建..."
echo ""

SUCCESSFUL_BUILDS=()
FAILED_BUILDS=()

# macOS Intel (通常能成功)
if build_platform "darwin" "amd64" "" "macos-intel"; then
    SUCCESSFUL_BUILDS+=("macOS Intel")
else
    FAILED_BUILDS+=("macOS Intel")
fi

# macOS Apple Silicon
if build_platform "darwin" "arm64" "" "macos-apple-silicon"; then
    SUCCESSFUL_BUILDS+=("macOS Apple Silicon")
else
    FAILED_BUILDS+=("macOS Apple Silicon")
fi

# Windows 64位 (可能需要特殊环境)
echo "尝试构建 Windows 64位版本..."
export CGO_ENABLED=0  # Windows禁用CGO尝试
if GOOS=windows GOARCH=amd64 go build \
    -ldflags="${LDFLAGS}" \
    -o "${RELEASE_DIR}/pdf-merger-windows-64bit.exe" \
    ./cmd/pdfmerger 2>/dev/null; then
    echo "  ✓ 构建成功: pdf-merger-windows-64bit.exe"
    SUCCESSFUL_BUILDS+=("Windows 64位")
else
    echo "  ✗ 构建失败: Windows版本需要在Windows系统上构建"
    FAILED_BUILDS+=("Windows 64位")
fi

# Linux 64位
echo "尝试构建 Linux 64位版本..."
export CGO_ENABLED=0  # Linux禁用CGO尝试
if GOOS=linux GOARCH=amd64 go build \
    -ldflags="${LDFLAGS}" \
    -o "${RELEASE_DIR}/pdf-merger-linux-64bit" \
    ./cmd/pdfmerger 2>/dev/null; then
    echo "  ✓ 构建成功: pdf-merger-linux-64bit"
    SUCCESSFUL_BUILDS+=("Linux 64位")
else
    echo "  ✗ 构建失败: Linux版本需要在Linux系统上构建"
    FAILED_BUILDS+=("Linux 64位")
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

echo "✅ 多平台构建完成！"
echo ""
echo "说明:"
echo "- 成功构建的版本可以直接使用"
echo "- 失败的版本需要在对应系统上构建"
echo "- Fyne GUI应用的跨平台构建需要特定的环境配置"
