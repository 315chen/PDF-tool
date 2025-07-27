package ui

import (
	"log"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ChineseTheme 支持中文显示的主题
type ChineseTheme struct {
	fyne.Theme
}

// NewChineseTheme 创建新的中文主题
func NewChineseTheme() fyne.Theme {
	return &ChineseTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Font 返回支持中文的字体资源
func (t *ChineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 在不同操作系统上使用不同的字体策略
	switch runtime.GOOS {
	case "darwin": // macOS
		return t.getMacOSFont(style)
	case "windows":
		return t.getWindowsFont(style)
	default:
		return t.getLinuxFont(style)
	}
}

// getMacOSFont 获取macOS系统字体
func (t *ChineseTheme) getMacOSFont(style fyne.TextStyle) fyne.Resource {
	// 在macOS上，使用系统默认字体通常能正确显示中文
	// 如果仍有问题，可以尝试指定特定字体
	return theme.DefaultTheme().Font(style)
}

// getWindowsFont 获取Windows系统字体
func (t *ChineseTheme) getWindowsFont(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// getLinuxFont 获取Linux系统字体
func (t *ChineseTheme) getLinuxFont(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// SetupFontEnvironment 设置字体环境
func SetupFontEnvironment() {
	switch runtime.GOOS {
	case "darwin": // macOS
		setupMacOSFont()
	case "windows":
		setupWindowsFont()
	default:
		setupLinuxFont()
	}
}

// setupMacOSFont 设置macOS字体环境
func setupMacOSFont() {
	// 设置语言环境
	os.Setenv("LANG", "zh_CN.UTF-8")
	os.Setenv("LC_ALL", "zh_CN.UTF-8")
	os.Setenv("LC_CTYPE", "zh_CN.UTF-8")

	// 尝试设置中文字体
	macOSFonts := []string{
		"/System/Library/Fonts/PingFang.ttc",          // PingFang SC (系统默认中文字体)
		"/System/Library/Fonts/STHeiti Light.ttc",     // 华文黑体
		"/System/Library/Fonts/Helvetica.ttc",         // Helvetica (备用)
		"/Library/Fonts/Arial Unicode MS.ttf",         // Arial Unicode MS
		"/System/Library/Fonts/Apple Color Emoji.ttc", // 表情符号字体
	}

	for _, fontPath := range macOSFonts {
		if _, err := os.Stat(fontPath); err == nil {
			os.Setenv("FYNE_FONT", fontPath)
			log.Printf("设置macOS字体: %s", fontPath)
			break
		}
	}

	// 设置字体回退
	os.Setenv("FONTCONFIG_PATH", "/etc/fonts")
}

// setupWindowsFont 设置Windows字体环境
func setupWindowsFont() {
	os.Setenv("LANG", "zh_CN.UTF-8")
	os.Setenv("LC_ALL", "zh_CN.UTF-8")

	// Windows通常能自动处理中文字体
	log.Println("Windows字体环境设置完成")
}

// setupLinuxFont 设置Linux字体环境
func setupLinuxFont() {
	os.Setenv("LANG", "zh_CN.UTF-8")
	os.Setenv("LC_ALL", "zh_CN.UTF-8")
	os.Setenv("LC_CTYPE", "zh_CN.UTF-8")

	// 常见的Linux中文字体路径
	linuxFonts := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
	}

	for _, fontPath := range linuxFonts {
		if _, err := os.Stat(fontPath); err == nil {
			os.Setenv("FYNE_FONT", fontPath)
			log.Printf("设置Linux字体: %s", fontPath)
			break
		}
	}
}

// ApplyChineseTheme 应用中文主题到应用程序
func ApplyChineseTheme(app fyne.App) {
	// 首先设置字体环境
	SetupFontEnvironment()

	// 使用默认主题，环境变量设置应该足够解决字体问题
	app.Settings().SetTheme(theme.DefaultTheme())

	log.Printf("中文字体环境已设置 (操作系统: %s)", runtime.GOOS)
}
