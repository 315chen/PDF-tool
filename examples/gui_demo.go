//go:build ignore
// +build ignore
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
	log.Println("启动PDF合并工具GUI演示...")
	
	// 创建应用程序实例
	a := app.New()
	a.SetIcon(nil)
	
	w := a.NewWindow("PDF合并工具 - 演示版")
	w.Resize(fyne.NewSize(900, 700))
	w.CenterOnScreen()
	
	// 初始化服务
	tempDir := createTempDir()
	log.Printf("临时目录: %s", tempDir)
	
	// 创建服务实例
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	
	// 创建配置
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	config.WindowWidth = 900
	config.WindowHeight = 700
	
	log.Println("初始化服务完成")
	
	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建UI
	userInterface := ui.NewUI(w, ctrl)
	
	// 设置主窗口内容
	content := userInterface.BuildUI()
	w.SetContent(content)
	
	log.Println("UI构建完成")
	
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

// createTempDir 创建临时目录
func createTempDir() string {
	// 使用系统临时目录下的应用特定子目录
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	return tempDir
}