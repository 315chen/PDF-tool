#!/bin/bash

# PDF合并工具中文字体修复脚本
# 此脚本专门解决macOS上的中文字体显示问题

echo "=== PDF合并工具中文字体修复脚本 ==="
echo ""

# 检查操作系统
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "此脚本专为macOS设计，当前系统: $OSTYPE"
    echo "在其他系统上，中文字体问题可能不存在或需要不同的解决方案"
    exit 1
fi

echo "检测到macOS系统，开始修复中文字体问题..."
echo ""

# 1. 检查系统字体
echo "1. 检查系统中文字体..."
CHINESE_FONTS=(
    "/System/Library/Fonts/PingFang.ttc"
    "/System/Library/Fonts/STHeiti Light.ttc"
    "/Library/Fonts/Arial Unicode MS.ttf"
    "/System/Library/Fonts/Helvetica.ttc"
    "/System/Library/Fonts/Apple Color Emoji.ttc"
)

FOUND_FONT=""
for font in "${CHINESE_FONTS[@]}"; do
    if [ -f "$font" ]; then
        echo "   ✓ 找到字体: $font"
        if [ -z "$FOUND_FONT" ]; then
            FOUND_FONT="$font"
        fi
    else
        echo "   ✗ 未找到: $font"
    fi
done

if [ -z "$FOUND_FONT" ]; then
    echo "   ⚠️  警告: 未找到合适的中文字体"
else
    echo "   ✓ 将使用字体: $FOUND_FONT"
fi

echo ""

# 2. 创建字体配置
echo "2. 创建字体配置..."

# 创建fontconfig目录
FONTCONFIG_DIR="$HOME/.config/fontconfig"
mkdir -p "$FONTCONFIG_DIR"

# 创建字体配置文件
cat > "$FONTCONFIG_DIR/fonts.conf" << 'EOF'
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
    <!-- 中文字体配置 -->
    <alias>
        <family>sans-serif</family>
        <prefer>
            <family>PingFang SC</family>
            <family>STHeiti</family>
            <family>Helvetica</family>
            <family>Arial Unicode MS</family>
        </prefer>
    </alias>
    
    <alias>
        <family>serif</family>
        <prefer>
            <family>PingFang SC</family>
            <family>STSong</family>
            <family>Times</family>
        </prefer>
    </alias>
    
    <alias>
        <family>monospace</family>
        <prefer>
            <family>SF Mono</family>
            <family>Menlo</family>
            <family>Monaco</family>
        </prefer>
    </alias>
    
    <!-- 强制使用中文字体显示中文字符 -->
    <match target="pattern">
        <test name="lang">
            <string>zh-cn</string>
        </test>
        <edit name="family" mode="prepend" binding="strong">
            <string>PingFang SC</string>
        </edit>
    </match>
</fontconfig>
EOF

echo "   ✓ 字体配置文件已创建: $FONTCONFIG_DIR/fonts.conf"
echo ""

# 3. 创建启动脚本
echo "3. 创建优化的启动脚本..."

cat > "run_with_chinese_font.sh" << EOF
#!/bin/bash

# PDF合并工具中文字体启动脚本
echo "启动PDF合并工具 (中文字体优化版)..."

# 设置语言环境
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8
export LC_CTYPE=zh_CN.UTF-8

# 设置字体环境
export FONTCONFIG_PATH="\$HOME/.config/fontconfig"
export FONTCONFIG_FILE="\$HOME/.config/fontconfig/fonts.conf"

# 如果找到了中文字体，设置FYNE_FONT
if [ -f "$FOUND_FONT" ]; then
    export FYNE_FONT="$FOUND_FONT"
    echo "使用字体: $FOUND_FONT"
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
EOF

chmod +x "run_with_chinese_font.sh"
echo "   ✓ 启动脚本已创建: run_with_chinese_font.sh"
echo ""

# 4. 创建中文版本的字符串文件
echo "4. 恢复中文界面文本..."

# 备份当前的英文版本
if [ -f "internal/ui/strings.go" ]; then
    cp "internal/ui/strings.go" "internal/ui/strings_english.go.bak"
fi

# 创建中文版本
cat > "internal/ui/strings_chinese.go" << 'EOF'
package ui

// 中文界面文本常量
const (
	// 窗口标题
	WindowTitleChinese = "PDF合并工具"
	
	// 按钮文本
	BrowseButtonChinese        = "浏览..."
	AddFileButtonChinese       = "添加文件"
	RemoveFileButtonChinese    = "移除选中"
	ClearFilesButtonChinese    = "清空列表"
	MoveUpButtonChinese        = "上移"
	MoveDownButtonChinese      = "下移"
	RefreshButtonChinese       = "刷新"
	StartMergeButtonChinese    = "开始合并"
	CancelButtonChinese        = "取消"
	
	// 标签文本
	MainFileLabelChinese       = "主PDF文件:"
	AdditionalFilesLabelChinese = "附加PDF文件:"
	OutputPathLabelChinese     = "输出路径:"
	NoFilesLabelChinese        = "没有文件"
	ProgressLabelChinese       = "进度:"
	StatusLabelChinese         = "状态:"
	
	// 状态消息
	StatusReadyTextChinese     = "就绪"
	StatusMergingChinese       = "正在合并..."
	StatusCompletedTextChinese = "合并完成"
	StatusCancelledTextChinese = "已取消"
	StatusErrorTextChinese     = "发生错误"
)
EOF

echo "   ✓ 中文字符串文件已创建: internal/ui/strings_chinese.go"
echo ""

# 5. 提供使用说明
echo "=== 修复完成 ==="
echo ""
echo "现在您有以下选项来解决中文字体问题:"
echo ""
echo "选项1: 使用英文界面 (推荐)"
echo "   运行: ./pdf-merger-english"
echo "   这个版本使用英文界面，避免了字体问题"
echo ""
echo "选项2: 使用中文界面 + 字体修复"
echo "   1. 重新构建中文版本:"
echo "      go build -o pdf-merger-chinese ./cmd/pdfmerger"
echo "   2. 使用修复脚本启动:"
echo "      ./run_with_chinese_font.sh"
echo ""
echo "选项3: 手动设置环境变量"
echo "   export LANG=zh_CN.UTF-8"
echo "   export LC_ALL=zh_CN.UTF-8"
if [ -n "$FOUND_FONT" ]; then
echo "   export FYNE_FONT=\"$FOUND_FONT\""
fi
echo "   ./pdf-merger-font-fix"
echo ""
echo "建议: 如果您主要使用英文环境，选择选项1最简单有效"
echo "如果您需要中文界面，请尝试选项2"
echo ""
echo "如果问题仍然存在，请检查:"
echo "1. 系统是否安装了中文字体"
echo "2. 系统语言设置是否正确"
echo "3. 终端的字符编码设置"
EOF

chmod +x fix_chinese_font.sh
