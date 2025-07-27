#!/bin/bash

# macOS PDF合并工具启动脚本
# 此脚本解决macOS上的中文字体显示问题

echo "正在启动PDF合并工具..."

# 设置语言环境
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8
export LC_CTYPE=zh_CN.UTF-8

# 设置字体配置
export FONTCONFIG_PATH=/etc/fonts

# 尝试设置中文字体路径
MACOS_FONTS=(
    "/System/Library/Fonts/PingFang.ttc"
    "/System/Library/Fonts/STHeiti Light.ttc"
    "/Library/Fonts/Arial Unicode MS.ttf"
    "/System/Library/Fonts/Helvetica.ttc"
)

for font in "${MACOS_FONTS[@]}"; do
    if [ -f "$font" ]; then
        export FYNE_FONT="$font"
        echo "使用字体: $font"
        break
    fi
done

# 设置Fyne特定的环境变量
export FYNE_THEME=light
export FYNE_SCALE=1.0

# 检查可执行文件是否存在
if [ ! -f "./pdf-merger-font-fix" ]; then
    echo "错误: 找不到可执行文件 pdf-merger-font-fix"
    echo "请先运行: go build -o pdf-merger-font-fix ./cmd/pdfmerger"
    exit 1
fi

# 启动应用程序
echo "启动PDF合并工具..."
./pdf-merger-font-fix

echo "PDF合并工具已退出"
