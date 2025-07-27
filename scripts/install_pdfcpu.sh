#!/bin/bash

# install_pdfcpu.sh - 安装pdfcpu依赖的脚本

set -e

echo "🚀 开始安装pdfcpu依赖..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go未安装或不在PATH中"
    exit 1
fi

echo "✅ Go版本: $(go version)"

# 进入项目目录
cd "$(dirname "$0")/.."

echo "📁 当前目录: $(pwd)"

# 备份当前的go.mod
cp go.mod go.mod.backup
echo "💾 已备份go.mod文件"

# 尝试添加pdfcpu依赖
echo "📦 尝试添加pdfcpu依赖..."

# 尝试不同版本的pdfcpu
VERSIONS=("v0.8.0" "v0.7.0" "v0.6.0" "v0.5.0" "v0.4.0")

for version in "${VERSIONS[@]}"; do
    echo "🔄 尝试版本 $version..."
    
    # 添加依赖到go.mod
    if ! grep -q "github.com/pdfcpu/pdfcpu" go.mod; then
        # 如果不存在，添加到require块中
        sed -i.tmp '/require (/a\
	github.com/pdfcpu/pdfcpu '"$version"'
' go.mod && rm go.mod.tmp
    else
        # 如果存在，更新版本
        sed -i.tmp "s|github.com/pdfcpu/pdfcpu.*|github.com/pdfcpu/pdfcpu $version|" go.mod && rm go.mod.tmp
    fi
    
    # 尝试下载依赖
    if timeout 60 go mod download github.com/pdfcpu/pdfcpu; then
        echo "✅ 成功下载pdfcpu $version"
        
        # 运行go mod tidy
        if timeout 60 go mod tidy; then
            echo "✅ 成功运行go mod tidy"
            
            # 测试编译
            if go build ./pkg/pdf; then
                echo "✅ 成功编译PDF包"
                echo "🎉 pdfcpu $version 安装成功！"
                
                # 更新pdfcpu_adapter.go以启用真正的pdfcpu功能
                echo "🔧 更新pdfcpu适配器..."
                update_adapter
                
                # 运行测试验证
                echo "🧪 运行测试验证..."
                if go test -v ./pkg/pdf -run TestPDFServiceCompatibility -timeout 30s; then
                    echo "✅ 测试通过！"
                    echo "🎊 pdfcpu迁移准备完成！"
                    exit 0
                else
                    echo "⚠️  测试失败，但依赖已安装"
                    exit 0
                fi
            else
                echo "❌ 编译失败，尝试下一个版本..."
            fi
        else
            echo "❌ go mod tidy失败，尝试下一个版本..."
        fi
    else
        echo "❌ 下载失败，尝试下一个版本..."
    fi
    
    # 恢复备份
    cp go.mod.backup go.mod
done

echo "❌ 所有版本都安装失败"
echo "💡 可能的解决方案:"
echo "   1. 检查网络连接"
echo "   2. 尝试使用VPN"
echo "   3. 设置Go代理: go env -w GOPROXY=https://goproxy.cn,direct"
echo "   4. 手动下载pdfcpu源码"

# 恢复备份
cp go.mod.backup go.mod
rm go.mod.backup

exit 1

# 更新适配器函数
update_adapter() {
    local adapter_file="pkg/pdf/pdfcpu_adapter.go"
    local availability_file="pkg/pdf/pdfcpu_availability.go"
    
    echo "📝 更新 $adapter_file..."
    
    # 取消注释pdfcpu导入
    sed -i.tmp 's|// "github.com/pdfcpu/pdfcpu/pkg/api"|"github.com/pdfcpu/pdfcpu/pkg/api"|' "$adapter_file" && rm "$adapter_file.tmp"
    sed -i.tmp 's|// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"|"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"|' "$adapter_file" && rm "$adapter_file.tmp"
    
    echo "📝 更新 $availability_file..."
    
    # 更新可用性检查
    cat > "$availability_file" << 'EOF'
package pdf

import (
	"fmt"
	"log"
	
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// PDFCPUAvailability 检查pdfcpu库的可用性
type PDFCPUAvailability struct {
	isAvailable bool
	version     string
	error       error
}

// CheckPDFCPUAvailability 检查pdfcpu是否可用
func CheckPDFCPUAvailability() *PDFCPUAvailability {
	availability := &PDFCPUAvailability{
		isAvailable: true,
		version:     pdfcpu.VersionStr,
	}

	return availability
}

// IsAvailable 返回pdfcpu是否可用
func (a *PDFCPUAvailability) IsAvailable() bool {
	return a.isAvailable
}

// GetVersion 返回pdfcpu版本
func (a *PDFCPUAvailability) GetVersion() string {
	return a.version
}

// GetError 返回错误信息
func (a *PDFCPUAvailability) GetError() error {
	return a.error
}

// LogStatus 记录pdfcpu状态
func (a *PDFCPUAvailability) LogStatus(logger *log.Logger) {
	if a.isAvailable {
		logger.Printf("pdfcpu is available (version: %s)", a.version)
	} else {
		logger.Printf("pdfcpu is not available: %v", a.error)
	}
}

// GetFallbackMessage 获取回退消息
func (a *PDFCPUAvailability) GetFallbackMessage() string {
	if a.isAvailable {
		return ""
	}
	return "Using placeholder implementation. Install pdfcpu for full functionality."
}

// ShouldUseFallback 是否应该使用回退实现
func (a *PDFCPUAvailability) ShouldUseFallback() bool {
	return !a.isAvailable
}
EOF
    
    echo "✅ 适配器更新完成"
}