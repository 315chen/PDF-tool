#!/bin/bash

# PDF合并工具CI/CD构建脚本
# 适用于GitHub Actions, GitLab CI, Jenkins等CI/CD系统

set -e

echo "🚀 PDF合并工具 - CI/CD构建"
echo "=========================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 环境变量
CI_COMMIT_SHA=${CI_COMMIT_SHA:-${GITHUB_SHA:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}}
CI_COMMIT_REF_NAME=${CI_COMMIT_REF_NAME:-${GITHUB_REF_NAME:-$(git branch --show-current 2>/dev/null || echo "unknown")}}
CI_PIPELINE_ID=${CI_PIPELINE_ID:-${GITHUB_RUN_ID:-$(date +%s)}}
CI_JOB_ID=${CI_JOB_ID:-${GITHUB_RUN_NUMBER:-"1"}}

# 配置变量
APP_NAME="pdf-merger"
VERSION=${VERSION:-"dev-$(echo $CI_COMMIT_SHA | cut -c1-8)"}
BUILD_DIR="build"
ARTIFACTS_DIR="artifacts"
REPORTS_DIR="reports"

# 构建阶段标志
RUN_TESTS=${RUN_TESTS:-true}
RUN_LINT=${RUN_LINT:-true}
RUN_SECURITY_SCAN=${RUN_SECURITY_SCAN:-true}
BUILD_BINARIES=${BUILD_BINARIES:-true}
BUILD_DOCKER=${BUILD_DOCKER:-false}
UPLOAD_ARTIFACTS=${UPLOAD_ARTIFACTS:-true}

# 初始化CI环境
init_ci_environment() {
    echo "🔧 初始化CI环境..."
    
    # 创建必要目录
    mkdir -p "$BUILD_DIR" "$ARTIFACTS_DIR" "$REPORTS_DIR"
    
    # 显示环境信息
    echo "CI环境信息:"
    echo "  提交SHA: $CI_COMMIT_SHA"
    echo "  分支: $CI_COMMIT_REF_NAME"
    echo "  流水线ID: $CI_PIPELINE_ID"
    echo "  任务ID: $CI_JOB_ID"
    echo "  版本: $VERSION"
    echo "  Go版本: $(go version)"
    echo "  操作系统: $(uname -a)"
    
    # 设置Go环境
    export CGO_ENABLED=0
    export GOPROXY=${GOPROXY:-"https://proxy.golang.org,direct"}
    export GOSUMDB=${GOSUMDB:-"sum.golang.org"}
    
    echo -e "${GREEN}✅ CI环境初始化完成${NC}"
}

# 下载依赖
download_dependencies() {
    echo "📦 下载依赖..."
    
    # 验证go.mod和go.sum
    go mod verify
    
    # 下载依赖
    go mod download
    
    # 整理依赖
    go mod tidy
    
    # 检查是否有未提交的变更
    if [ -n "$(git status --porcelain go.mod go.sum 2>/dev/null)" ]; then
        echo -e "${YELLOW}⚠️  go.mod或go.sum有未提交的变更${NC}"
        git diff go.mod go.sum
    fi
    
    echo -e "${GREEN}✅ 依赖下载完成${NC}"
}

# 代码质量检查
run_quality_checks() {
    if [ "$RUN_LINT" != "true" ]; then
        echo "跳过代码质量检查"
        return 0
    fi
    
    echo "🔍 运行代码质量检查..."
    
    # 格式检查
    echo "检查代码格式..."
    if [ -n "$(gofmt -l .)" ]; then
        echo -e "${RED}❌ 代码格式不正确${NC}"
        echo "未格式化的文件:"
        gofmt -l .
        return 1
    fi
    
    # 导入检查
    if command -v goimports &> /dev/null; then
        echo "检查导入格式..."
        if [ -n "$(goimports -l .)" ]; then
            echo -e "${YELLOW}⚠️  导入格式需要调整${NC}"
            goimports -l .
        fi
    fi
    
    # 静态分析
    if command -v golangci-lint &> /dev/null; then
        echo "运行golangci-lint..."
        golangci-lint run --out-format=junit-xml > "$REPORTS_DIR/golangci-lint.xml" || true
        golangci-lint run
    else
        echo -e "${YELLOW}⚠️  golangci-lint未安装，跳过静态分析${NC}"
    fi
    
    # 代码复杂度检查
    if command -v gocyclo &> /dev/null; then
        echo "检查代码复杂度..."
        gocyclo -over 15 . > "$REPORTS_DIR/complexity.txt" || true
    fi
    
    echo -e "${GREEN}✅ 代码质量检查完成${NC}"
}

# 安全扫描
run_security_scan() {
    if [ "$RUN_SECURITY_SCAN" != "true" ]; then
        echo "跳过安全扫描"
        return 0
    fi
    
    echo "🔒 运行安全扫描..."
    
    # 漏洞扫描
    if command -v govulncheck &> /dev/null; then
        echo "运行漏洞扫描..."
        govulncheck ./... > "$REPORTS_DIR/vulnerabilities.txt" || true
    else
        echo -e "${YELLOW}⚠️  govulncheck未安装，跳过漏洞扫描${NC}"
    fi
    
    # 安全检查
    if command -v gosec &> /dev/null; then
        echo "运行安全检查..."
        gosec -fmt=junit-xml -out="$REPORTS_DIR/security.xml" ./... || true
        gosec ./...
    else
        echo -e "${YELLOW}⚠️  gosec未安装，跳过安全检查${NC}"
    fi
    
    echo -e "${GREEN}✅ 安全扫描完成${NC}"
}

# 运行测试
run_tests() {
    if [ "$RUN_TESTS" != "true" ]; then
        echo "跳过测试"
        return 0
    fi
    
    echo "🧪 运行测试..."
    
    # 单元测试
    echo "运行单元测试..."
    go test -v -race -coverprofile="$REPORTS_DIR/coverage.out" \
        -covermode=atomic \
        -timeout=10m \
        ./internal/... ./pkg/... \
        2>&1 | tee "$REPORTS_DIR/unit-tests.log"
    
    # 生成覆盖率报告
    if [ -f "$REPORTS_DIR/coverage.out" ]; then
        go tool cover -html="$REPORTS_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"
        
        # 显示覆盖率统计
        local coverage=$(go tool cover -func="$REPORTS_DIR/coverage.out" | tail -1 | awk '{print $3}')
        echo "代码覆盖率: $coverage"
        
        # 检查覆盖率目标
        local coverage_num=$(echo $coverage | sed 's/%//')
        if (( $(echo "$coverage_num >= 70" | bc -l) )); then
            echo -e "${GREEN}✅ 覆盖率达标: $coverage${NC}"
        else
            echo -e "${YELLOW}⚠️  覆盖率未达标: $coverage < 70%${NC}"
        fi
    fi
    
    # 集成测试
    echo "运行集成测试..."
    go test -v -timeout=5m ./test/... \
        2>&1 | tee "$REPORTS_DIR/integration-tests.log" || true
    
    # 基准测试
    echo "运行基准测试..."
    go test -bench=. -benchmem -timeout=5m ./... \
        > "$REPORTS_DIR/benchmarks.txt" 2>&1 || true
    
    echo -e "${GREEN}✅ 测试完成${NC}"
}

# 构建二进制文件
build_binaries() {
    if [ "$BUILD_BINARIES" != "true" ]; then
        echo "跳过二进制构建"
        return 0
    fi
    
    echo "🔨 构建二进制文件..."
    
    # 构建标志
    local ldflags="-s -w -X main.version=$VERSION -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.gitCommit=$CI_COMMIT_SHA"
    
    # 支持的平台
    declare -A platforms=(
        ["linux-amd64"]="linux amd64"
        ["linux-arm64"]="linux arm64"
        ["darwin-amd64"]="darwin amd64"
        ["darwin-arm64"]="darwin arm64"
        ["windows-amd64"]="windows amd64"
    )
    
    for platform in "${!platforms[@]}"; do
        IFS=' ' read -r goos goarch <<< "${platforms[$platform]}"
        
        echo "构建 $platform..."
        
        local output_name="$APP_NAME"
        if [ "$goos" = "windows" ]; then
            output_name="$APP_NAME.exe"
        fi
        
        local output_path="$BUILD_DIR/$platform/$output_name"
        mkdir -p "$BUILD_DIR/$platform"
        
        env GOOS=$goos GOARCH=$goarch go build \
            -ldflags="$ldflags" \
            -o "$output_path" \
            ./cmd/pdfmerger
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ $platform 构建成功${NC}"
            
            # 创建压缩包
            cd "$BUILD_DIR"
            tar -czf "../$ARTIFACTS_DIR/$APP_NAME-$VERSION-$platform.tar.gz" "$platform/"
            cd ..
        else
            echo -e "${RED}❌ $platform 构建失败${NC}"
        fi
    done
    
    echo -e "${GREEN}✅ 二进制构建完成${NC}"
}

# 构建Docker镜像
build_docker_image() {
    if [ "$BUILD_DOCKER" != "true" ]; then
        echo "跳过Docker构建"
        return 0
    fi
    
    echo "🐳 构建Docker镜像..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${YELLOW}⚠️  Docker未安装，跳过镜像构建${NC}"
        return 0
    fi
    
    # 构建镜像
    local image_tag="$APP_NAME:$VERSION"
    
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --build-arg GIT_COMMIT="$CI_COMMIT_SHA" \
        -t "$image_tag" \
        .
    
    # 保存镜像
    docker save "$image_tag" | gzip > "$ARTIFACTS_DIR/$APP_NAME-$VERSION-docker.tar.gz"
    
    echo -e "${GREEN}✅ Docker镜像构建完成${NC}"
}

# 生成构建报告
generate_build_report() {
    echo "📊 生成构建报告..."
    
    local report_file="$REPORTS_DIR/build-report.md"
    
    cat > "$report_file" << EOF
# PDF合并工具构建报告

## 构建信息
- **版本**: $VERSION
- **提交**: $CI_COMMIT_SHA
- **分支**: $CI_COMMIT_REF_NAME
- **构建时间**: $(date)
- **流水线ID**: $CI_PIPELINE_ID

## 构建结果
EOF
    
    # 添加测试结果
    if [ -f "$REPORTS_DIR/coverage.out" ]; then
        local coverage=$(go tool cover -func="$REPORTS_DIR/coverage.out" | tail -1 | awk '{print $3}')
        echo "- **代码覆盖率**: $coverage" >> "$report_file"
    fi
    
    # 添加构建产物
    echo "" >> "$report_file"
    echo "## 构建产物" >> "$report_file"
    if [ -d "$ARTIFACTS_DIR" ]; then
        ls -la "$ARTIFACTS_DIR" | tail -n +2 | while read line; do
            echo "- $line" >> "$report_file"
        done
    fi
    
    echo -e "${GREEN}✅ 构建报告已生成: $report_file${NC}"
}

# 上传构建产物
upload_artifacts() {
    if [ "$UPLOAD_ARTIFACTS" != "true" ]; then
        echo "跳过产物上传"
        return 0
    fi
    
    echo "📤 准备上传构建产物..."
    
    # 这里可以根据不同的CI系统实现不同的上传逻辑
    # GitHub Actions: 使用actions/upload-artifact
    # GitLab CI: 使用artifacts配置
    # Jenkins: 使用archiveArtifacts
    
    echo "构建产物列表:"
    find "$ARTIFACTS_DIR" -type f -exec ls -lh {} \;
    
    echo "测试报告列表:"
    find "$REPORTS_DIR" -type f -exec ls -lh {} \;
    
    echo -e "${GREEN}✅ 构建产物准备完成${NC}"
}

# 清理环境
cleanup() {
    echo "🧹 清理构建环境..."
    
    # 清理临时文件
    go clean -cache -testcache -modcache 2>/dev/null || true
    
    # 清理Docker（如果需要）
    if [ "$BUILD_DOCKER" = "true" ] && command -v docker &> /dev/null; then
        docker system prune -f 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✅ 环境清理完成${NC}"
}

# 主函数
main() {
    echo "开始CI/CD构建流程..."
    echo ""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-tests)
                RUN_TESTS=false
                shift
                ;;
            --skip-lint)
                RUN_LINT=false
                shift
                ;;
            --skip-security)
                RUN_SECURITY_SCAN=false
                shift
                ;;
            --docker)
                BUILD_DOCKER=true
                shift
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            *)
                echo "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    # 设置错误处理
    trap cleanup EXIT
    
    # 执行构建流程
    init_ci_environment
    download_dependencies
    run_quality_checks
    run_security_scan
    run_tests
    build_binaries
    build_docker_image
    generate_build_report
    upload_artifacts
    
    echo ""
    echo -e "${GREEN}🎉 CI/CD构建完成！${NC}"
    echo "版本: $VERSION"
    echo "构建产物: $ARTIFACTS_DIR"
    echo "测试报告: $REPORTS_DIR"
}

# 运行主函数
main "$@"
