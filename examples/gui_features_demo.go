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
	"fyne.io/fyne/v2/dialog"
	
	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/ui"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	log.Println("=== PDF合并工具GUI功能演示 ===")
	
	// 创建应用程序实例
	a := app.New()
	a.SetIcon(nil)
	
	w := a.NewWindow("PDF合并工具 - 功能演示")
	w.Resize(fyne.NewSize(1000, 800))
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
	config.WindowWidth = 1000
	config.WindowHeight = 800
	
	log.Println("初始化服务完成")
	
	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建UI
	userInterface := ui.NewUI(w, ctrl)
	
	// 设置主窗口内容
	content := userInterface.BuildUI()
	w.SetContent(content)
	
	log.Println("UI构建完成")
	
	// 显示欢迎信息
	go func() {
		time.Sleep(500 * time.Millisecond)
		showWelcomeDialog(w, tempDir)
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

// createTempDir 创建临时目录
func createTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-features-demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	return tempDir
}

// createTestPDFFiles 创建测试用的PDF文件
func createTestPDFFiles(tempDir string) {
	log.Println("创建测试PDF文件...")
	
	// 创建简单的PDF文件内容（这不是真正的PDF，只是用于演示）
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
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello World) Tj
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
300
%%EOF`

	// 创建多个测试文件
	testFiles := []string{
		"main_document.pdf",
		"appendix_a.pdf",
		"appendix_b.pdf",
		"references.pdf",
	}
	
	for i, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("%s\n%% Test file %d: %s", pdfContent, i+1, filename)
		
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			log.Printf("创建测试文件 %s 失败: %v", filename, err)
		} else {
			log.Printf("创建测试文件: %s", filePath)
		}
	}
}

// showWelcomeDialog 显示欢迎对话框
func showWelcomeDialog(w fyne.Window, tempDir string) {
	welcomeText := fmt.Sprintf(`欢迎使用PDF合并工具！

这是一个功能演示版本，展示了以下特性：

🔹 主界面布局
  - 主PDF文件选择
  - 附加PDF文件列表管理
  - 输出文件路径设置

🔹 文件操作
  - 文件浏览和选择
  - 文件列表管理（添加、移除、清空）
  - 文件信息显示

🔹 用户界面
  - 响应式布局设计
  - 进度显示和状态反馈
  - 错误提示和信息对话框

🔹 测试文件
已在以下目录创建了测试PDF文件：
%s

您可以使用这些文件来测试合并功能。

注意：由于unidoc许可证限制，实际的PDF合并功能可能无法正常工作，但界面功能完全可用。`, tempDir)

	dialog.ShowInformation("欢迎", welcomeText, w)
}

// 演示功能的辅助函数

// simulateMergeProcess 模拟合并过程
func simulateMergeProcess(ui *ui.UI) {
	go func() {
		// 模拟合并过程
		for i := 0; i <= 100; i += 10 {
			time.Sleep(200 * time.Millisecond)
			progress := float64(i) / 100.0
			ui.SetProgress(progress)
			ui.SetStatus(fmt.Sprintf("正在合并... %d%%", i))
		}
		
		// 完成
		ui.SetStatus("合并完成！")
		time.Sleep(1 * time.Second)
		
		// 重置状态
		ui.SetProgress(0)
		ui.SetStatus("准备就绪")
	}()
}