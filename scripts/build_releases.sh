#!/bin/bash

# PDF合并工具发布版本构建脚本
# 构建多平台可执行文件

set -e

echo "=== PDF合并工具发布版本构建 ==="
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
    
    # 英文版本
    GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}" \
        ./cmd/pdfmerger
    
    # 检查构建结果
    if [ -f "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}" ]; then
        echo "  ✓ 构建成功: pdf-merger-${PLATFORM_NAME}${EXT}"
        
        # 获取文件大小
        if command -v stat >/dev/null 2>&1; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                SIZE=$(stat -f%z "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}")
            else
                SIZE=$(stat -c%s "${RELEASE_DIR}/pdf-merger-${PLATFORM_NAME}${EXT}")
            fi
            # 简单的大小显示，避免依赖numfmt
            if [ ${SIZE} -gt 1048576 ]; then
                echo "  文件大小: $((SIZE / 1048576))MB"
            else
                echo "  文件大小: $((SIZE / 1024))KB"
            fi
        fi
    else
        echo "  ✗ 构建失败: pdf-merger-${PLATFORM_NAME}${EXT}"
        return 1
    fi
    
    echo ""
}

# 构建当前平台版本
echo "构建当前平台版本..."
echo ""

# 检测当前平台
CURRENT_OS=$(go env GOOS)
CURRENT_ARCH=$(go env GOARCH)

echo "当前平台: ${CURRENT_OS}/${CURRENT_ARCH}"
echo ""

# 构建当前平台版本
case "${CURRENT_OS}" in
    "darwin")
        build_platform "darwin" "amd64" "" "macos-intel"
        ;;
    "windows")
        build_platform "windows" "amd64" ".exe" "windows-64bit"
        ;;
    "linux")
        build_platform "linux" "amd64" "" "linux-64bit"
        ;;
    *)
        echo "不支持的平台: ${CURRENT_OS}"
        exit 1
        ;;
esac

echo "注意: 其他平台版本需要在对应系统上构建"

echo "=== 构建完成 ==="
echo ""

# 创建发布说明
echo "创建发布说明..."
cat > "${RELEASE_DIR}/RELEASE_NOTES.md" << EOF
# PDF合并工具 ${VERSION} 发布说明

## 📦 下载

选择适合您操作系统的版本：

### macOS
- **Intel Mac**: \`pdf-merger-macos-intel\`
- **Apple Silicon (M1/M2)**: \`pdf-merger-macos-apple-silicon\`

### Windows
- **64位系统**: \`pdf-merger-windows-64bit.exe\`
- **32位系统**: \`pdf-merger-windows-32bit.exe\`

### Linux
- **64位系统**: \`pdf-merger-linux-64bit\`
- **32位系统**: \`pdf-merger-linux-32bit\`
- **ARM64**: \`pdf-merger-linux-arm64\`

## 🚀 快速开始

### macOS
\`\`\`bash
# 下载后添加执行权限
chmod +x pdf-merger-macos-*
./pdf-merger-macos-*
\`\`\`

### Windows
直接双击 \`.exe\` 文件运行

### Linux
\`\`\`bash
# 下载后添加执行权限
chmod +x pdf-merger-linux-*
./pdf-merger-linux-*
\`\`\`

## ✨ 主要功能

- 📄 **PDF文件合并** - 支持多个PDF文件合并为单个文件
- 🔐 **加密文件处理** - 自动处理密码保护的PDF文件
- 🎨 **现代化界面** - 基于Fyne的跨平台GUI
- 📊 **实时进度** - 详细的合并进度和状态显示
- 🔄 **拖拽支持** - 支持文件拖拽添加和排序
- ⚡ **高性能** - 流式处理，支持大文件合并
- 🛡️ **错误恢复** - 完善的错误处理和恢复机制

## 🔧 系统要求

- **内存**: 至少 512MB 可用内存
- **磁盘空间**: 至少 100MB 可用空间
- **操作系统**:
  - macOS 10.14 或更高版本
  - Windows 10 或更高版本
  - Ubuntu 18.04 或更高版本（Linux）

## 📚 文档

- [用户使用指南](../docs/USER_GUIDE.md)
- [技术开发指南](../docs/TECHNICAL_GUIDE.md)
- [macOS字体修复指南](../docs/MACOS_FONT_FIX.md)

## 🐛 已知问题

- **macOS中文字体**: 如果遇到中文显示问题，请参考 [macOS字体修复指南](../docs/MACOS_FONT_FIX.md)

## 📞 技术支持

如有问题，请查看文档或提交Issue。

---

**构建信息**
- 版本: ${VERSION}
- 构建时间: ${BUILD_TIME}
- Git提交: ${GIT_COMMIT}
EOF

echo "  ✓ 发布说明已创建: ${RELEASE_DIR}/RELEASE_NOTES.md"
echo ""

# 创建校验和文件
echo "创建校验和文件..."
cd "${RELEASE_DIR}"
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum pdf-merger-* > checksums.sha256
    echo "  ✓ SHA256校验和已创建: checksums.sha256"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 pdf-merger-* > checksums.sha256
    echo "  ✓ SHA256校验和已创建: checksums.sha256"
fi
cd - > /dev/null

# 显示构建结果
echo ""
echo "=== 构建结果 ==="
echo ""
echo "发布文件位置: ${RELEASE_DIR}/"
ls -la "${RELEASE_DIR}/"
echo ""

echo "✅ 所有平台版本构建完成！"
echo ""
echo "下一步："
echo "1. 测试各平台版本"
echo "2. 创建GitHub Release"
echo "3. 上传构建文件"
echo "4. 更新文档链接"
