#!/bin/bash

# PDF合并工具主构建脚本
# 统一管理所有构建、测试、部署任务

set -e

echo "🚀 PDF合并工具 - 主构建系统"
echo "============================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"

# 显示帮助信息
show_help() {
    echo ""
    echo -e "${CYAN}PDF合并工具构建系统${NC}"
    echo "======================"
    echo ""
    echo "用法: $0 <命令> [选项]"
    echo ""
    echo -e "${YELLOW}可用命令:${NC}"
    echo ""
    echo -e "${GREEN}开发相关:${NC}"
    echo "  setup           - 设置开发环境"
    echo "  dev             - 开发模式运行"
    echo "  test            - 运行所有测试"
    echo "  test-unit       - 运行单元测试"
    echo "  test-integration - 运行集成测试"
    echo "  coverage        - 生成测试覆盖率报告"
    echo "  lint            - 代码静态分析"
    echo "  fmt             - 格式化代码"
    echo ""
    echo -e "${GREEN}构建相关:${NC}"
    echo "  build           - 构建当前平台版本"
    echo "  build-all       - 构建所有平台版本"
    echo "  build-release   - 构建发布版本"
    echo "  build-docker    - 构建Docker镜像"
    echo ""
    echo -e "${GREEN}CI/CD相关:${NC}"
    echo "  ci              - 运行CI构建流程"
    echo "  deploy          - 部署到服务器"
    echo "  install-deps    - 安装pdfcpu依赖"
    echo ""
    echo -e "${GREEN}维护相关:${NC}"
    echo "  clean           - 清理构建文件"
    echo "  clean-all       - 深度清理"
    echo "  status          - 显示项目状态"
    echo "  info            - 显示项目信息"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo "  --version <版本>  - 指定版本号"
    echo "  --env <环境>      - 指定环境 (dev/test/prod)"
    echo "  --verbose        - 详细输出"
    echo "  --help           - 显示帮助信息"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo "  $0 setup                    # 设置开发环境"
    echo "  $0 build --version 1.0.0   # 构建指定版本"
    echo "  $0 test --verbose           # 详细模式运行测试"
    echo "  $0 build-release            # 构建发布版本"
    echo "  $0 deploy --env prod        # 部署到生产环境"
    echo ""
}

# 检查脚本是否存在
check_script() {
    local script_name="$1"
    local script_path="$SCRIPTS_DIR/$script_name"
    
    if [ ! -f "$script_path" ]; then
        echo -e "${RED}❌ 脚本不存在: $script_path${NC}"
        return 1
    fi
    
    if [ ! -x "$script_path" ]; then
        echo "设置脚本执行权限: $script_path"
        chmod +x "$script_path"
    fi
    
    return 0
}

# 运行脚本
run_script() {
    local script_name="$1"
    shift
    local script_path="$SCRIPTS_DIR/$script_name"
    
    if check_script "$script_name"; then
        echo -e "${BLUE}🔧 运行脚本: $script_name${NC}"
        echo "参数: $@"
        echo ""
        "$script_path" "$@"
    else
        return 1
    fi
}

# 设置开发环境
setup_dev() {
    echo -e "${CYAN}🛠️  设置开发环境${NC}"
    run_script "setup_dev.sh" "$@"
}

# 开发模式运行
dev_run() {
    echo -e "${CYAN}🏃 开发模式运行${NC}"
    make dev
}

# 运行测试
run_tests() {
    local test_type="$1"
    shift
    
    case "$test_type" in
        "unit")
            echo -e "${CYAN}🧪 运行单元测试${NC}"
            go test ./internal/... ./pkg/... -v "$@"
            ;;
        "integration")
            echo -e "${CYAN}🔗 运行集成测试${NC}"
            run_script "run_integration_tests.sh" "$@"
            ;;
        "all"|"")
            echo -e "${CYAN}🧪 运行所有测试${NC}"
            make test
            ;;
        *)
            echo -e "${RED}❌ 未知测试类型: $test_type${NC}"
            return 1
            ;;
    esac
}

# 生成覆盖率报告
generate_coverage() {
    echo -e "${CYAN}📊 生成测试覆盖率报告${NC}"
    run_script "test_coverage.sh" "$@"
}

# 代码静态分析
run_lint() {
    echo -e "${CYAN}🔍 运行代码静态分析${NC}"
    make lint
}

# 格式化代码
format_code() {
    echo -e "${CYAN}✨ 格式化代码${NC}"
    make fmt
}

# 构建项目
build_project() {
    local build_type="$1"
    shift
    
    case "$build_type" in
        "current"|"")
            echo -e "${CYAN}🔨 构建当前平台版本${NC}"
            make build "$@"
            ;;
        "all")
            echo -e "${CYAN}🏗️  构建所有平台版本${NC}"
            make build-all "$@"
            ;;
        "release")
            echo -e "${CYAN}📦 构建发布版本${NC}"
            run_script "build_release.sh" "$@"
            ;;
        "docker")
            echo -e "${CYAN}🐳 构建Docker镜像${NC}"
            run_script "build_docker.sh" "$@"
            ;;
        *)
            echo -e "${RED}❌ 未知构建类型: $build_type${NC}"
            return 1
            ;;
    esac
}

# CI构建
run_ci() {
    echo -e "${CYAN}🚀 运行CI构建流程${NC}"
    run_script "ci_build.sh" "$@"
}

# 部署项目
deploy_project() {
    echo -e "${CYAN}🚀 部署项目${NC}"
    run_script "deploy.sh" "$@"
}

# 安装依赖
install_dependencies() {
    echo -e "${CYAN}📦 安装pdfcpu依赖${NC}"
    run_script "install_pdfcpu.sh" "$@"
}

# 清理项目
clean_project() {
    local clean_type="$1"
    
    case "$clean_type" in
        "all")
            echo -e "${CYAN}🧹 深度清理项目${NC}"
            make clean
            go clean -cache -testcache -modcache
            docker system prune -f 2>/dev/null || true
            ;;
        ""|"basic")
            echo -e "${CYAN}🧹 清理构建文件${NC}"
            make clean
            ;;
        *)
            echo -e "${RED}❌ 未知清理类型: $clean_type${NC}"
            return 1
            ;;
    esac
}

# 显示项目状态
show_status() {
    echo -e "${CYAN}📊 项目状态${NC}"
    echo "============="
    echo ""
    
    # Git状态
    if git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${YELLOW}Git状态:${NC}"
        echo "  分支: $(git branch --show-current 2>/dev/null || echo 'unknown')"
        echo "  提交: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        echo "  状态: $(git status --porcelain | wc -l | tr -d ' ') 个未提交变更"
        echo ""
    fi
    
    # Go环境
    echo -e "${YELLOW}Go环境:${NC}"
    echo "  版本: $(go version | awk '{print $3}')"
    echo "  GOPATH: $(go env GOPATH)"
    echo "  GOPROXY: $(go env GOPROXY)"
    echo ""
    
    # 项目信息
    echo -e "${YELLOW}项目信息:${NC}"
    echo "  模块: $(go list -m 2>/dev/null || echo 'unknown')"
    echo "  路径: $PROJECT_ROOT"
    echo ""
    
    # 构建产物
    echo -e "${YELLOW}构建产物:${NC}"
    if [ -d "build" ]; then
        find build -name "*" -type f | head -5 | while read file; do
            echo "  $file ($(ls -lh "$file" | awk '{print $5}'))"
        done
    else
        echo "  无构建产物"
    fi
    echo ""
    
    # 测试覆盖率
    if [ -f "coverage.out" ]; then
        local coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
        echo -e "${YELLOW}测试覆盖率:${NC} $coverage"
    fi
}

# 显示项目信息
show_info() {
    echo -e "${CYAN}📋 项目信息${NC}"
    make info
}

# 主函数
main() {
    # 检查是否在项目根目录
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}❌ 请在项目根目录运行此脚本${NC}"
        exit 1
    fi
    
    # 解析全局选项
    local verbose=false
    local version=""
    local env="dev"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --verbose)
                verbose=true
                shift
                ;;
            --version)
                version="$2"
                export VERSION="$version"
                shift 2
                ;;
            --env)
                env="$2"
                export DEPLOY_ENV="$env"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            -*)
                echo -e "${RED}❌ 未知选项: $1${NC}"
                show_help
                exit 1
                ;;
            *)
                break
                ;;
        esac
    done
    
    # 设置详细输出
    if [ "$verbose" = true ]; then
        set -x
    fi
    
    # 获取命令
    local command="$1"
    shift || true
    
    # 执行命令
    case "$command" in
        "setup")
            setup_dev "$@"
            ;;
        "dev")
            dev_run "$@"
            ;;
        "test")
            run_tests "all" "$@"
            ;;
        "test-unit")
            run_tests "unit" "$@"
            ;;
        "test-integration")
            run_tests "integration" "$@"
            ;;
        "coverage")
            generate_coverage "$@"
            ;;
        "lint")
            run_lint "$@"
            ;;
        "fmt")
            format_code "$@"
            ;;
        "build")
            build_project "current" "$@"
            ;;
        "build-all")
            build_project "all" "$@"
            ;;
        "build-release")
            build_project "release" "$@"
            ;;
        "build-docker")
            build_project "docker" "$@"
            ;;
        "ci")
            run_ci "$@"
            ;;
        "deploy")
            deploy_project "$@"
            ;;
        "install-deps")
            install_dependencies "$@"
            ;;
        "clean")
            clean_project "basic" "$@"
            ;;
        "clean-all")
            clean_project "all" "$@"
            ;;
        "status")
            show_status "$@"
            ;;
        "info")
            show_info "$@"
            ;;
        "help"|"")
            show_help
            ;;
        *)
            echo -e "${RED}❌ 未知命令: $command${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
