package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/ui"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	// 创建应用程序实例
	a := app.New()
	a.SetIcon(nil) // 可以设置应用图标

	// 应用中文字体支持
	ui.ApplyChineseTheme(a)

	w := a.NewWindow("PDF Merger Tool")
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()

	// 初始化服务
	tempDir := createTempDir()

	// 创建服务实例
	fileManager := createFileManager(tempDir)
	pdfService := createPDFService()

	// 创建配置
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建事件处理器
	eventHandler := controller.NewEventHandler(ctrl)

	// 创建UI
	userInterface := ui.NewUI(w, ctrl)

	// 连接事件处理器和UI
	setupEventHandling(userInterface, eventHandler)

	// 设置主窗口内容
	w.SetContent(userInterface.BuildUI())

	// 添加应用程序关闭时的清理操作
	w.SetCloseIntercept(func() {
		// 清理临时文件
		if err := fileManager.CleanupTempFiles(); err != nil {
			log.Printf("清理临时文件时发生错误: %v", err)
		}

		log.Println("应用程序正在关闭...")
		a.Quit()
	})

	// 运行应用程序
	w.ShowAndRun()
}

// createTempDir 创建临时目录
func createTempDir() string {
	// 使用系统临时目录下的应用特定子目录
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-"+filepath.Base(os.Args[0])+"-"+filepath.Base(os.TempDir()))
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	return tempDir
}

// createFileManager 创建文件管理器实例
func createFileManager(tempDir string) file.FileManager {
	return file.NewFileManager(tempDir)
}

// createPDFService 创建PDF服务实例
func createPDFService() pdf.PDFService {
	return pdf.NewPDFService()
}

// setupEventHandling 设置事件处理
func setupEventHandling(ui *ui.UI, eventHandler *controller.EventHandler) {
	// 设置UI状态回调
	eventHandler.SetUIStateCallback(func(enabled bool) {
		if enabled {
			ui.EnableControls()
		} else {
			ui.DisableControls()
		}
	})

	// 设置进度更新回调
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		ui.UpdateProgressWithStrings(progress, status, detail)
	})

	// 设置错误回调
	eventHandler.SetErrorCallback(func(err error) {
		ui.ShowError(err)
	})

	// 设置完成回调
	eventHandler.SetCompletionCallback(func(message string) {
		ui.ShowCompletion(message)
	})

	// 设置UI的事件处理器
	ui.SetEventHandler(eventHandler)
}
