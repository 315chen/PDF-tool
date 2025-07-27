//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	
	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/ui"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	log.Println("=== 进度显示和状态反馈演示 ===")
	
	// 创建应用程序实例
	a := app.New()
	a.SetIcon(nil)
	
	w := a.NewWindow("PDF合并工具 - 进度显示演示")
	w.Resize(fyne.NewSize(1400, 1000))
	w.CenterOnScreen()
	
	// 初始化服务
	tempDir := createTempDir()
	log.Printf("临时目录: %s", tempDir)
	
	// 创建测试PDF文件
	createTestPDFFiles(tempDir)
	
	// 创建服务实例
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	
	// 创建配置
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	config.WindowWidth = 1400
	config.WindowHeight = 1000
	
	log.Println("初始化服务完成")
	
	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建UI
	userInterface := ui.NewUI(w, ctrl)
	
	// 设置主窗口内容
	mainContent := userInterface.BuildUI()
	
	// 创建演示控制面板
	demoPanel := createProgressDemoPanel(userInterface, tempDir, w)
	
	// 组合布局
	mainLayout := container.NewHSplit(
		mainContent,
		demoPanel,
	)
	mainLayout.SetOffset(0.65) // 65%给主界面，35%给演示面板
	
	w.SetContent(mainLayout)
	
	log.Println("UI构建完成")
	
	// 显示欢迎信息
	go func() {
		time.Sleep(500 * time.Millisecond)
		showProgressDemoWelcome(w, tempDir)
	}()
	
	// 添加应用程序关闭时的清理操作
	w.SetCloseIntercept(func() {
		log.Println("正在清理资源...")
		
		// 清理临时文件
		if err := fileManager.CleanupTempFiles(); err != nil {
			log.Printf("清理临时文件时发生错误: %v", err)
		}
		
		// 清理临时目录
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("清理临时目录时发生错误: %v", err)
		}
		
		log.Println("应用程序正在关闭...")
		a.Quit()
	})
	
	log.Println("启动GUI界面...")
	
	// 运行应用程序
	w.ShowAndRun()
}

// createProgressDemoPanel 创建进度演示面板
func createProgressDemoPanel(userInterface *ui.UI, tempDir string, window fyne.Window) *fyne.Container {
	// 创建独立的进度管理器用于演示
	demoProgressManager := ui.NewProgressManager(window)
	
	// 演示按钮
	startProgressBtn := widget.NewButtonWithIcon("开始进度演示", theme.MediaPlayIcon(), func() {
		demonstrateProgress(demoProgressManager)
	})
	
	simulateErrorBtn := widget.NewButtonWithIcon("模拟错误", theme.ErrorIcon(), func() {
		demonstrateError(demoProgressManager)
	})
	
	simulateCancelBtn := widget.NewButtonWithIcon("模拟取消", theme.CancelIcon(), func() {
		demonstrateCancel(demoProgressManager)
	})
	
	showInfoDialogBtn := widget.NewButtonWithIcon("信息对话框", theme.InfoIcon(), func() {
		demoProgressManager.ShowInfoDialog("信息", "这是一个信息对话框演示")
	})
	
	showErrorDialogBtn := widget.NewButtonWithIcon("错误对话框", theme.ErrorIcon(), func() {
		demoProgressManager.ShowErrorDialog("错误", "这是一个错误对话框演示")
	})
	
	showConfirmDialogBtn := widget.NewButtonWithIcon("确认对话框", theme.QuestionIcon(), func() {
		demoProgressManager.ShowConfirmDialog("确认", "您确定要执行此操作吗？", func(confirmed bool) {
			if confirmed {
				log.Println("用户确认了操作")
			} else {
				log.Println("用户取消了操作")
			}
		})
	})
	
	// 状态演示按钮
	statusButtons := container.NewVBox(
		widget.NewLabel("状态演示:"),
		widget.NewButton("准备状态", func() {
			msg := ui.GetStatusMessage(ui.StatusReady, "系统准备就绪")
			demoProgressManager.SetStatus(msg.Title + ": " + msg.Message)
		}),
		widget.NewButton("处理状态", func() {
			msg := ui.GetStatusMessage(ui.StatusProcessing, "正在处理文件")
			demoProgressManager.SetStatus(msg.Title + ": " + msg.Message)
		}),
		widget.NewButton("完成状态", func() {
			msg := ui.GetStatusMessage(ui.StatusCompleted, "操作已完成")
			demoProgressManager.SetStatus(msg.Title + ": " + msg.Message)
		}),
		widget.NewButton("错误状态", func() {
			msg := ui.GetStatusMessage(ui.StatusError, "发生了错误")
			demoProgressManager.SetStatus(msg.Title + ": " + msg.Message)
		}),
	)
	
	// 进度控制
	progressSlider := widget.NewSlider(0, 1)
	progressSlider.OnChanged = func(value float64) {
		demoProgressManager.SetProgress(value)
		demoProgressManager.SetDetail(fmt.Sprintf("进度: %.1f%%", value*100))
	}
	
	progressControls := container.NewVBox(
		widget.NewLabel("手动进度控制:"),
		progressSlider,
	)
	
	// 实时信息显示
	infoText := widget.NewRichText()
	infoText.Wrapping = fyne.TextWrapWord
	
	infoScroll := container.NewScroll(infoText)
	infoScroll.SetMinSize(fyne.NewSize(350, 200))
	
	// 定时更新信息
	go func() {
		for {
			time.Sleep(1 * time.Second)
			updateDemoInfo(infoText, demoProgressManager)
		}
	}()
	
	// 创建面板布局
	panel := container.NewVBox(
		widget.NewRichTextFromMarkdown("## 进度演示控制面板"),
		widget.NewSeparator(),
		
		widget.NewLabel("基本演示:"),
		startProgressBtn,
		simulateErrorBtn,
		simulateCancelBtn,
		
		widget.NewSeparator(),
		widget.NewLabel("对话框演示:"),
		showInfoDialogBtn,
		showErrorDialogBtn,
		showConfirmDialogBtn,
		
		widget.NewSeparator(),
		statusButtons,
		
		widget.NewSeparator(),
		progressControls,
		
		widget.NewSeparator(),
		widget.NewLabel("演示进度管理器状态:"),
		demoProgressManager.GetContainer(),
		
		widget.NewSeparator(),
		widget.NewLabel("实时信息:"),
		infoScroll,
	)
	
	return panel
}

// demonstrateProgress 演示进度功能
func demonstrateProgress(pm *ui.ProgressManager) {
	pm.Start(5, 10)
	
	go func() {
		steps := []struct {
			progress float64
			status   string
			detail   string
			file     string
		}{
			{0.1, "初始化", "正在初始化系统...", ""},
			{0.3, "验证文件", "正在验证PDF文件...", "document1.pdf"},
			{0.5, "处理文件", "正在处理PDF内容...", "document2.pdf"},
			{0.7, "合并文件", "正在合并PDF文件...", "document3.pdf"},
			{0.9, "保存文件", "正在保存合并结果...", "merged.pdf"},
			{1.0, "完成", "所有操作已完成", ""},
		}
		
		for i, step := range steps {
			if !pm.IsActive() {
				return // 用户取消了
			}
			
			pm.UpdateProgress(ui.ProgressInfo{
				Progress:       step.progress,
				Status:         step.status,
				Detail:         step.detail,
				CurrentFile:    step.file,
				ProcessedFiles: i + 1,
				TotalFiles:     len(steps),
				Step:           i + 1,
				TotalSteps:     len(steps),
			})
			
			time.Sleep(1 * time.Second)
		}
		
		pm.Complete("演示完成！")
	}()
}

// demonstrateError 演示错误功能
func demonstrateError(pm *ui.ProgressManager) {
	pm.Start(3, 5)
	
	go func() {
		pm.UpdateProgress(ui.ProgressInfo{
			Progress: 0.3,
			Status:   "处理中",
			Detail:   "正在处理文件...",
			Step:     1,
		})
		
		time.Sleep(1 * time.Second)
		
		pm.UpdateProgress(ui.ProgressInfo{
			Progress: 0.6,
			Status:   "遇到问题",
			Detail:   "检测到潜在错误...",
			Step:     2,
		})
		
		time.Sleep(1 * time.Second)
		
		// 模拟错误
		pm.Error(fmt.Errorf("演示错误：文件处理失败"))
	}()
}

// demonstrateCancel 演示取消功能
func demonstrateCancel(pm *ui.ProgressManager) {
	pm.Start(5, 10)
	
	go func() {
		for i := 0; i < 5; i++ {
			if !pm.IsActive() {
				return
			}
			
			pm.UpdateProgress(ui.ProgressInfo{
				Progress: float64(i) * 0.2,
				Status:   fmt.Sprintf("步骤 %d", i+1),
				Detail:   "正在处理...",
				Step:     i + 1,
			})
			
			time.Sleep(800 * time.Millisecond)
		}
		
		// 模拟取消
		pm.Cancel()
	}()
}

// updateDemoInfo 更新演示信息
func updateDemoInfo(infoText *widget.RichText, pm *ui.ProgressManager) {
	info := fmt.Sprintf("## 进度管理器状态\n\n")
	info += fmt.Sprintf("**活跃状态**: %t\n\n", pm.IsActive())
	info += fmt.Sprintf("**当前进度**: %.1f%%\n\n", pm.GetProgress()*100)
	
	if pm.IsActive() {
		elapsed := pm.GetElapsedTime()
		info += fmt.Sprintf("**已用时间**: %v\n\n", elapsed)
	}
	
	info += "### 功能特性\n\n"
	info += "- ✅ 实时进度更新\n"
	info += "- ✅ 状态消息显示\n"
	info += "- ✅ 详细信息展示\n"
	info += "- ✅ 时间和速度统计\n"
	info += "- ✅ 错误处理和显示\n"
	info += "- ✅ 取消操作支持\n"
	info += "- ✅ 完成状态处理\n"
	info += "- ✅ 多种对话框类型\n"
	
	infoText.ParseMarkdown(info)
}

// createTempDir 创建临时目录
func createTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-progress-demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	return tempDir
}

// createTestPDFFiles 创建测试用的PDF文件
func createTestPDFFiles(tempDir string) {
	log.Println("创建测试PDF文件...")
	
	// 创建简单的PDF文件内容
	pdfContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 50
>>
stream
BT
/F1 12 Tf
100 700 Td
(Progress Demo Document) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
0000000200 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
320
%%EOF`

	// 创建多个测试文件
	testFiles := []string{
		"progress_test_1.pdf",
		"progress_test_2.pdf",
		"progress_test_3.pdf",
		"progress_test_4.pdf",
		"progress_test_5.pdf",
	}
	
	for i, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("%s\n%% Test file %d for progress demo", pdfContent, i+1)
		
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			log.Printf("创建测试文件 %s 失败: %v", filename, err)
		} else {
			log.Printf("创建测试文件: %s", filePath)
		}
	}
}

// showProgressDemoWelcome 显示进度演示欢迎对话框
func showProgressDemoWelcome(w fyne.Window, tempDir string) {
	welcomeText := fmt.Sprintf(`# 进度显示和状态反馈演示

欢迎使用PDF合并工具的进度显示功能演示！

## 🎯 演示功能

### 进度显示
- 📊 实时进度条更新
- ⏱️ 时间统计和速度显示
- 📝 详细状态信息
- 📁 当前处理文件显示

### 状态反馈
- ✅ 成功状态指示
- ❌ 错误状态处理
- ⏸️ 取消操作支持
- 🔄 实时状态更新

### 对话框系统
- 💬 信息提示对话框
- ⚠️ 错误警告对话框
- ❓ 确认选择对话框
- 📋 自定义内容对话框

### 用户体验
- 🎨 现代化界面设计
- 📱 响应式布局
- 🔄 平滑动画效果
- 💡 直观操作反馈

## 📁 测试文件

已在以下目录创建了测试PDF文件：
%s

## 🚀 使用方法

### 右侧演示面板：
1. **基本演示** - 体验完整的进度流程
2. **错误演示** - 查看错误处理效果
3. **取消演示** - 测试取消操作
4. **对话框演示** - 体验各种对话框
5. **状态演示** - 查看不同状态效果
6. **手动控制** - 手动调节进度

### 左侧主界面：
- 添加测试文件到列表
- 设置输出路径
- 点击"开始合并"查看实际进度

开始探索进度显示的强大功能吧！`, tempDir)

	dialog.ShowInformation("欢迎", welcomeText, w)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}