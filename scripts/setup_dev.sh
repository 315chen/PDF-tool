#!/bin/bash

# PDF合并工具开发环境设置脚本

set -e

echo "🛠️  PDF合并工具 - 开发环境设置"
echo "=============================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
GO_MIN_VERSION="1.21"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 检查操作系统
detect_os() {
    case "$(uname -s)" in
        Darwin*)    OS="macos" ;;
        Linux*)     OS="linux" ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) OS="windows" ;;
        *)          OS="unknown" ;;
    esac
    echo "检测到操作系统: $OS"
}

# 检查Go环境
check_go() {
    echo "🔍 检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装${NC}"
        echo "请访问 https://golang.org/dl/ 下载安装Go"
        return 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}✅ Go版本: $go_version${NC}"
    
    # 检查版本是否满足要求
    if [ "$(printf '%s\n' "$GO_MIN_VERSION" "$go_version" | sort -V | head -n1)" != "$GO_MIN_VERSION" ]; then
        echo -e "${YELLOW}⚠️  建议使用Go $GO_MIN_VERSION或更高版本${NC}"
    fi
    
    # 显示Go环境信息
    echo "Go环境信息:"
    echo "  GOROOT: $(go env GOROOT)"
    echo "  GOPATH: $(go env GOPATH)"
    echo "  GOPROXY: $(go env GOPROXY)"
    
    return 0
}

# 安装开发工具
install_dev_tools() {
    echo "🔧 安装开发工具..."
    
    # 工具列表
    declare -A tools=(
        ["golangci-lint"]="github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        ["goimports"]="golang.org/x/tools/cmd/goimports@latest"
        ["govulncheck"]="golang.org/x/vuln/cmd/govulncheck@latest"
        ["staticcheck"]="honnef.co/go/tools/cmd/staticcheck@latest"
        ["air"]="github.com/cosmtrek/air@latest"
    )
    
    for tool in "${!tools[@]}"; do
        echo "安装 $tool..."
        if go install "${tools[$tool]}"; then
            echo -e "${GREEN}✅ $tool 安装成功${NC}"
        else
            echo -e "${YELLOW}⚠️  $tool 安装失败${NC}"
        fi
    done
}

# 安装系统依赖
install_system_deps() {
    echo "📦 安装系统依赖..."
    
    case $OS in
        "macos")
            if command -v brew &> /dev/null; then
                echo "使用Homebrew安装依赖..."
                brew install upx tree || echo "部分依赖安装失败"
            else
                echo -e "${YELLOW}⚠️  Homebrew未安装，跳过系统依赖安装${NC}"
                echo "建议安装Homebrew: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
            fi
            ;;
        "linux")
            if command -v apt-get &> /dev/null; then
                echo "使用apt安装依赖..."
                sudo apt-get update
                sudo apt-get install -y upx tree zip unzip || echo "部分依赖安装失败"
            elif command -v yum &> /dev/null; then
                echo "使用yum安装依赖..."
                sudo yum install -y upx tree zip unzip || echo "部分依赖安装失败"
            else
                echo -e "${YELLOW}⚠️  未检测到包管理器，请手动安装: upx, tree, zip, unzip${NC}"
            fi
            ;;
        "windows")
            echo -e "${YELLOW}⚠️  Windows系统请手动安装以下工具:${NC}"
            echo "  - UPX: https://upx.github.io/"
            echo "  - Git Bash或WSL"
            ;;
    esac
}

# 设置Git钩子
setup_git_hooks() {
    echo "🪝 设置Git钩子..."
    
    local hooks_dir="$PROJECT_ROOT/.git/hooks"
    
    if [ ! -d "$hooks_dir" ]; then
        echo -e "${YELLOW}⚠️  Git仓库未初始化，跳过Git钩子设置${NC}"
        return
    fi
    
    # 创建pre-commit钩子
    cat > "$hooks_dir/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook for PDF合并工具

echo "运行pre-commit检查..."

# 格式化代码
echo "格式化代码..."
go fmt ./...

# 运行静态检查
if command -v golangci-lint &> /dev/null; then
    echo "运行golangci-lint..."
    golangci-lint run
fi

# 运行测试
echo "运行测试..."
go test ./... -short

echo "Pre-commit检查完成"
EOF
    
    chmod +x "$hooks_dir/pre-commit"
    echo -e "${GREEN}✅ Pre-commit钩子已设置${NC}"
}

# 创建开发配置文件
create_dev_configs() {
    echo "📝 创建开发配置文件..."
    
    # 创建.air.toml配置文件（热重载）
    if [ ! -f "$PROJECT_ROOT/.air.toml" ]; then
        cat > "$PROJECT_ROOT/.air.toml" << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/pdfmerger"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "build", "dist"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
EOF
        echo -e "${GREEN}✅ .air.toml 配置文件已创建${NC}"
    fi
    
    # 创建.golangci.yml配置文件
    if [ ! -f "$PROJECT_ROOT/.golangci.yml" ]; then
        cat > "$PROJECT_ROOT/.golangci.yml" << 'EOF'
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - gocritic
    - gosec

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
    disabled-checks:
      - dupImport
      - ifElseChain
      - octalLiteral
      - whyNoLint
      - wrapperFunc

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - gocritic
EOF
        echo -e "${GREEN}✅ .golangci.yml 配置文件已创建${NC}"
    fi
}

# 设置IDE配置
setup_ide_configs() {
    echo "💻 设置IDE配置..."
    
    # VS Code配置
    local vscode_dir="$PROJECT_ROOT/.vscode"
    mkdir -p "$vscode_dir"
    
    # 创建settings.json
    if [ ! -f "$vscode_dir/settings.json" ]; then
        cat > "$vscode_dir/settings.json" << 'EOF'
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,128,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "files.exclude": {
        "**/tmp": true,
        "**/build": true,
        "**/dist": true,
        "**/*.exe": true
    }
}
EOF
        echo -e "${GREEN}✅ VS Code settings.json 已创建${NC}"
    fi
    
    # 创建launch.json
    if [ ! -f "$vscode_dir/launch.json" ]; then
        cat > "$vscode_dir/launch.json" << 'EOF'
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch PDF Merger",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/pdfmerger",
            "env": {},
            "args": []
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}",
            "env": {},
            "args": ["-test.v"]
        }
    ]
}
EOF
        echo -e "${GREEN}✅ VS Code launch.json 已创建${NC}"
    fi
}

# 验证环境
verify_environment() {
    echo "🔍 验证开发环境..."
    
    # 检查Go模块
    echo "检查Go模块..."
    cd "$PROJECT_ROOT"
    go mod tidy
    go mod verify
    
    # 运行快速测试
    echo "运行快速测试..."
    if go test ./... -short; then
        echo -e "${GREEN}✅ 测试通过${NC}"
    else
        echo -e "${RED}❌ 测试失败${NC}"
        return 1
    fi
    
    # 检查代码格式
    echo "检查代码格式..."
    if [ -n "$(gofmt -l .)" ]; then
        echo -e "${YELLOW}⚠️  代码格式需要调整${NC}"
        echo "运行: go fmt ./..."
    else
        echo -e "${GREEN}✅ 代码格式正确${NC}"
    fi
    
    # 检查静态分析
    if command -v golangci-lint &> /dev/null; then
        echo "运行静态分析..."
        if golangci-lint run --timeout=2m; then
            echo -e "${GREEN}✅ 静态分析通过${NC}"
        else
            echo -e "${YELLOW}⚠️  静态分析发现问题${NC}"
        fi
    fi
}

# 显示开发指南
show_dev_guide() {
    echo ""
    echo "📚 开发指南"
    echo "==========="
    echo ""
    echo "常用命令:"
    echo "  make help          - 显示所有可用命令"
    echo "  make dev           - 开发模式运行"
    echo "  make test          - 运行测试"
    echo "  make build         - 构建应用"
    echo "  make clean         - 清理构建文件"
    echo ""
    echo "开发工具:"
    echo "  air                - 热重载开发服务器"
    echo "  golangci-lint run  - 代码静态分析"
    echo "  go fmt ./...       - 格式化代码"
    echo "  go test ./...      - 运行所有测试"
    echo ""
    echo "项目结构:"
    echo "  cmd/               - 应用程序入口"
    echo "  internal/          - 内部包"
    echo "  pkg/               - 公共包"
    echo "  test/              - 集成测试"
    echo "  scripts/           - 构建脚本"
    echo "  docs/              - 文档"
    echo ""
    echo "开发流程:"
    echo "1. 创建功能分支: git checkout -b feature/xxx"
    echo "2. 编写代码和测试"
    echo "3. 运行测试: make test"
    echo "4. 提交代码: git commit -m 'feat: xxx'"
    echo "5. 推送分支: git push origin feature/xxx"
    echo "6. 创建Pull Request"
    echo ""
    echo -e "${GREEN}🎉 开发环境设置完成！${NC}"
    echo "开始愉快的开发吧！"
}

# 主函数
main() {
    echo "开始设置开发环境..."
    echo ""
    
    detect_os
    
    if ! check_go; then
        echo -e "${RED}❌ Go环境检查失败${NC}"
        exit 1
    fi
    
    install_dev_tools
    install_system_deps
    setup_git_hooks
    create_dev_configs
    setup_ide_configs
    
    if verify_environment; then
        show_dev_guide
    else
        echo -e "${RED}❌ 环境验证失败${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
