#!/bin/bash

# PDF合并工具中文字体启动脚本
echo "启动PDF合并工具 (中文字体优化版)..."

# 设置语言环境
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8
export LC_CTYPE=zh_CN.UTF-8

# 设置字体环境
export FONTCONFIG_PATH="$HOME/.config/fontconfig"
export FONTCONFIG_FILE="$HOME/.config/fontconfig/fonts.conf"

# 如果找到了中文字体，设置FYNE_FONT
if [ -f "/System/Library/Fonts/STHeiti Light.ttc" ]; then
    export FYNE_FONT="/System/Library/Fonts/STHeiti Light.ttc"
    echo "使用字体: /System/Library/Fonts/STHeiti Light.ttc"
fi

# 设置Fyne特定环境变量
export FYNE_THEME=light
export FYNE_SCALE=1.0

# 检查可执行文件
if [ ! -f "./pdf-merger-font-fix" ]; then
    echo "错误: 找不到可执行文件 pdf-merger-font-fix"
    echo "请先运行: go build -o pdf-merger-font-fix ./cmd/pdfmerger"
    exit 1
fi

# 启动应用程序
echo "正在启动应用程序..."
./pdf-merger-font-fix

echo "应用程序已退出"
