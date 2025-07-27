//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("PDF合并工具 - 控制器演示")
	fmt.Println("=============================")

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建服务实例
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建事件处理器
	eventHandler := controller.NewEventHandler(ctrl)

	// 设置回调函数
	setupCallbacks(eventHandler)

	// 演示功能
	demonstrateController(eventHandler)
}

func setupCallbacks(eventHandler *controller.EventHandler) {
	// 设置UI状态回调
	eventHandler.SetUIStateCallback(func(enabled bool) {
		if enabled {
			fmt.Println("✅ UI已启用")
		} else {
			fmt.Println("🔒 UI已禁用")
		}
	})

	// 设置进度更新回调
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		percentage := int(progress * 100)
		fmt.Printf("📊 进度: %d%% - %s: %s\n", percentage, status, detail)
	})

	// 设置错误回调
	eventHandler.SetErrorCallback(func(err error) {
		fmt.Printf("❌ 错误: %v\n", err)
	})

	// 设置完成回调
	eventHandler.SetCompletionCallback(func(message string) {
		fmt.Printf("🎉 完成: %s\n", message)
	})
}

func demonstrateController(eventHandler *controller.EventHandler) {
	fmt.Println("\n1. 演示文件验证")
	fmt.Println("----------------")

	// 测试文件验证（这些文件不存在，会产生错误）
	testFiles := []string{
		"test1.pdf",
		"test2.pdf",
		"nonexistent.pdf",
	}

	for _, file := range testFiles {
		fmt.Printf("验证文件: %s\n", file)
		_, err := eventHandler.HandleAdditionalFileAdded(file)
		if err != nil {
			fmt.Printf("  ❌ 验证失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 验证成功\n")
		}
	}

	fmt.Println("\n2. 演示任务状态检查")
	fmt.Println("------------------")

	// 检查初始状态
	fmt.Printf("初始任务状态: 运行中=%v\n", eventHandler.IsJobRunning())

	fmt.Println("\n3. 演示合并任务启动（会失败，因为文件不存在）")
	fmt.Println("----------------------------------------")

	// 尝试启动合并任务
	err := eventHandler.HandleMergeStart("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")
	if err != nil {
		fmt.Printf("启动任务失败: %v\n", err)
	} else {
		fmt.Println("任务已启动")

		// 检查任务状态
		fmt.Printf("任务状态: 运行中=%v\n", eventHandler.IsJobRunning())

		// 等待任务完成或失败
		time.Sleep(1 * time.Second)

		// 检查最终状态
		fmt.Printf("最终状态: 运行中=%v\n", eventHandler.IsJobRunning())
	}

	fmt.Println("\n4. 演示输出路径验证")
	fmt.Println("------------------")

	// 测试输出路径验证
	testOutputPaths := []string{
		"/tmp/output.pdf",
		"/nonexistent/path/output.pdf",
		"./output.pdf",
	}

	for _, path := range testOutputPaths {
		fmt.Printf("验证输出路径: %s\n", path)
		err := eventHandler.HandleOutputPathChanged(path)
		if err != nil {
			fmt.Printf("  ❌ 路径无效: %v\n", err)
		} else {
			fmt.Printf("  ✅ 路径有效\n")
		}
	}

	fmt.Println("\n演示完成！")
}