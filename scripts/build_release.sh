#!/bin/bash

# PDF合并工具发布构建脚本
# 用于构建跨平台发布版本

set -e

echo "🚀 PDF合并工具 - 发布构建"
echo "========================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
APP_NAME="pdf-merger"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="build"
DIST_DIR="dist"
CMD_DIR="./cmd/pdfmerger"

# 构建标志
BUILD_FLAGS="-ldflags=-s -w -X main.version=$VERSION -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
CGO_ENABLED=0

# 支持的平台
PLATFORMS="linux-amd64:linux:amd64 linux-arm64:linux:arm64 darwin-amd64:darwin:amd64 darwin-arm64:darwin:arm64 windows-amd64:windows:amd64 windows-arm64:windows:arm64"

# 清理函数
cleanup() {
    echo "🧹 清理构建文件..."
    rm -rf "$BUILD_DIR"
    rm -rf "$DIST_DIR"
}

# 检查依赖
check_dependencies() {
    echo "🔍 检查构建依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Go版本: $(go version)${NC}"
    
    # 检查zip命令
    if ! command -v zip &> /dev/null; then
        echo -e "${YELLOW}⚠️  zip命令未找到，将跳过压缩包创建${NC}"
        CREATE_ZIP=false
    else
        CREATE_ZIP=true
    fi
    
    # 检查upx（可选的压缩工具）
    if command -v upx &> /dev/null; then
        echo -e "${GREEN}✅ UPX可用，将压缩二进制文件${NC}"
        USE_UPX=true
    else
        echo -e "${YELLOW}⚠️  UPX未安装，跳过二进制压缩${NC}"
        USE_UPX=false
    fi
}

# 准备构建环境
prepare_build() {
    echo "📁 准备构建环境..."
    
    # 创建构建目录
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    # 下载依赖
    echo "📦 下载依赖..."
    go mod download
    go mod tidy
    
    # 运行测试
    echo "🧪 运行测试..."
    if ! go test ./... -short; then
        echo -e "${RED}❌ 测试失败，停止构建${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 测试通过${NC}"
}

# 构建单个平台
build_platform() {
    local platform=$1
    local goos=$2
    local goarch=$3
    
    echo -e "${BLUE}🔨 构建 $platform...${NC}"
    
    # 设置输出文件名
    local output_name="$APP_NAME"
    if [ "$goos" = "windows" ]; then
        output_name="$APP_NAME.exe"
    fi
    
    local output_path="$BUILD_DIR/$platform/$output_name"
    
    # 创建平台目录
    mkdir -p "$BUILD_DIR/$platform"
    
    # 构建
    env GOOS=$goos GOARCH=$goarch CGO_ENABLED=$CGO_ENABLED \
        go build $BUILD_FLAGS -o "$output_path" "$CMD_DIR"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $platform 构建成功${NC}"
        
        # 显示文件大小
        local size=$(ls -lh "$output_path" | awk '{print $5}')
        echo "   文件大小: $size"
        
        # 使用UPX压缩（如果可用且不是macOS）
        if [ "$USE_UPX" = true ] && [ "$goos" != "darwin" ]; then
            echo "   🗜️  使用UPX压缩..."
            upx --best --lzma "$output_path" 2>/dev/null || echo "   ⚠️  UPX压缩失败"
            local compressed_size=$(ls -lh "$output_path" | awk '{print $5}')
            echo "   压缩后大小: $compressed_size"
        fi
        
        # 复制相关文件
        cp README.md "$BUILD_DIR/$platform/" 2>/dev/null || true
        cp LICENSE "$BUILD_DIR/$platform/" 2>/dev/null || true
        
        # 创建压缩包
        if [ "$CREATE_ZIP" = true ]; then
            local zip_name="$APP_NAME-$VERSION-$platform.zip"
            echo "   📦 创建压缩包: $zip_name"
            
            cd "$BUILD_DIR"
            zip -r "../$DIST_DIR/$zip_name" "$platform/" > /dev/null
            cd ..
            
            echo -e "${GREEN}   ✅ 压缩包已创建${NC}"
        fi
        
        return 0
    else
        echo -e "${RED}❌ $platform 构建失败${NC}"
        return 1
    fi
}

# 构建所有平台
build_all_platforms() {
    echo "🏗️  开始构建所有平台..."

    local success_count=0
    local total_count=0

    for platform_info in $PLATFORMS; do
        total_count=$((total_count + 1))
        IFS=':' read -r platform goos goarch <<< "$platform_info"

        if build_platform "$platform" "$goos" "$goarch"; then
            success_count=$((success_count + 1))
        fi

        echo ""
    done

    echo "📊 构建统计:"
    echo "   成功: $success_count/$total_count"

    if [ $success_count -eq $total_count ]; then
        echo -e "${GREEN}🎉 所有平台构建成功！${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  部分平台构建失败${NC}"
        return 1
    fi
}

# 生成校验和
generate_checksums() {
    echo "🔐 生成校验和文件..."
    
    cd "$DIST_DIR"
    
    # 生成SHA256校验和
    if command -v sha256sum &> /dev/null; then
        sha256sum *.zip > checksums.sha256 2>/dev/null || true
        echo -e "${GREEN}✅ SHA256校验和已生成${NC}"
    elif command -v shasum &> /dev/null; then
        shasum -a 256 *.zip > checksums.sha256 2>/dev/null || true
        echo -e "${GREEN}✅ SHA256校验和已生成${NC}"
    else
        echo -e "${YELLOW}⚠️  无法生成校验和文件${NC}"
    fi
    
    cd ..
}

# 生成发布信息
generate_release_info() {
    echo "📄 生成发布信息..."
    
    local release_info="$DIST_DIR/release-info.txt"
    
    cat > "$release_info" << EOF
PDF合并工具 v$VERSION
==================

构建时间: $(date)
Go版本: $(go version)
构建机器: $(uname -a)

支持平台:
EOF
    
    for platform in "${!PLATFORMS[@]}"; do
        echo "- $platform" >> "$release_info"
    done
    
    cat >> "$release_info" << EOF

安装说明:
1. 下载对应平台的压缩包
2. 解压到目标目录
3. 运行可执行文件

注意事项:
- 首次运行时可能需要安装系统依赖
- Windows用户可能需要安装Visual C++运行库
- macOS用户可能需要在安全设置中允许运行

更多信息请访问项目主页。
EOF
    
    echo -e "${GREEN}✅ 发布信息已生成: $release_info${NC}"
}

# 显示构建结果
show_build_results() {
    echo ""
    echo "📋 构建结果:"
    echo "============"
    
    if [ -d "$DIST_DIR" ]; then
        echo "发布文件:"
        ls -lh "$DIST_DIR"
        
        echo ""
        echo "总大小:"
        du -sh "$DIST_DIR"
    fi
    
    echo ""
    echo "构建目录结构:"
    tree "$BUILD_DIR" 2>/dev/null || find "$BUILD_DIR" -type f
}

# 主函数
main() {
    echo "开始发布构建流程..."
    echo "版本: $VERSION"
    echo ""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --clean)
                cleanup
                shift
                ;;
            --platform)
                SINGLE_PLATFORM="$2"
                shift 2
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --version <版本>    设置版本号"
                echo "  --clean            清理构建文件"
                echo "  --platform <平台>  只构建指定平台"
                echo "  --help             显示帮助信息"
                echo ""
                echo "支持的平台:"
                for platform_info in $PLATFORMS; do
                    IFS=':' read -r platform goos goarch <<< "$platform_info"
                    echo "  $platform"
                done
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 执行构建流程
    check_dependencies
    prepare_build
    
    if [ -n "$SINGLE_PLATFORM" ]; then
        local found=false
        for platform_info in $PLATFORMS; do
            IFS=':' read -r platform goos goarch <<< "$platform_info"
            if [ "$platform" = "$SINGLE_PLATFORM" ]; then
                build_platform "$SINGLE_PLATFORM" "$goos" "$goarch"
                found=true
                break
            fi
        done

        if [ "$found" = false ]; then
            echo -e "${RED}❌ 不支持的平台: $SINGLE_PLATFORM${NC}"
            exit 1
        fi
    else
        build_all_platforms
    fi
    
    generate_checksums
    generate_release_info
    show_build_results
    
    echo ""
    echo -e "${GREEN}🎊 发布构建完成！${NC}"
    echo "发布文件位于: $DIST_DIR"
}

# 捕获中断信号
trap 'echo -e "\n${YELLOW}构建被中断${NC}"; exit 1' INT TERM

# 运行主函数
main "$@"
