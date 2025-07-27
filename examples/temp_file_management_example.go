//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/user/pdf-merger/pkg/file"
)

func main() {
	fmt.Println("临时文件管理示例")
	fmt.Println("=================")

	// 设置自动资源清理
	file.SetupDefaultAutoCleaner()

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-example")
	fmt.Printf("创建临时目录: %s\n", tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}

	// 将临时目录添加到自动清理
	file.AddDirectoryToAutoClean(tempDir, 10)

	// 创建文件管理器
	fileManager := file.NewFileManager(tempDir)
	fmt.Printf("临时文件目录: %s\n", fileManager.GetTempDir())

	// 创建临时文件
	tempFile, err := fileManager.CreateTempFile()
	if err != nil {
		fmt.Printf("创建临时文件失败: %v\n", err)
		return
	}
	fmt.Printf("创建临时文件: %s\n", tempFile)

	// 创建带前缀的临时文件
	prefixFile, fileObj, err := fileManager.CreateTempFileWithPrefix("prefix_", ".txt")
	if err != nil {
		fmt.Printf("创建带前缀的临时文件失败: %v\n", err)
		return
	}
	fileObj.Close()
	fmt.Printf("创建带前缀的临时文件: %s\n", prefixFile)

	// 创建带内容的临时文件
	content := []byte("这是临时文件的内容")
	contentFile, err := fileManager.CreateTempFileWithContent("content_", ".txt", content)
	if err != nil {
		fmt.Printf("创建带内容的临时文件失败: %v\n", err)
		return
	}
	fmt.Printf("创建带内容的临时文件: %s\n", contentFile)

	// 创建源文件
	sourceFile := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("源文件内容"), 0644); err != nil {
		fmt.Printf("创建源文件失败: %v\n", err)
		return
	}
	fmt.Printf("创建源文件: %s\n", sourceFile)

	// 将源文件复制到临时文件
	copyFile, err := fileManager.CopyToTempFile(sourceFile, "copy_")
	if err != nil {
		fmt.Printf("复制到临时文件失败: %v\n", err)
		return
	}
	fmt.Printf("复制到临时文件: %s\n", copyFile)

	// 创建资源管理器
	resourceManager := file.NewResourceManager()

	// 添加自定义资源
	resourceManager.AddCustom(func() error {
		fmt.Println("执行自定义清理操作")
		return nil
	}, 1)

	// 添加文件资源
	customFile := filepath.Join(tempDir, "custom.txt")
	if err := os.WriteFile(customFile, []byte("自定义文件内容"), 0644); err != nil {
		fmt.Printf("创建自定义文件失败: %v\n", err)
		return
	}
	resourceManager.AddFile(customFile, 2)
	fmt.Printf("添加文件资源: %s\n", customFile)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("\n程序正在运行...")
	fmt.Println("按Ctrl+C退出并清理资源")

	// 等待信号或超时
	select {
	case <-sigChan:
		fmt.Println("\n收到中断信号，正在清理资源...")
	case <-time.After(30 * time.Second):
		fmt.Println("\n超时，正在清理资源...")
	}

	// 清理资源
	if errors := resourceManager.Cleanup(); len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("清理资源时发生错误: %v\n", err)
		}
	} else {
		fmt.Println("资源管理器清理完成")
	}

	// 清理临时文件
	if err := fileManager.CleanupTempFiles(); err != nil {
		fmt.Printf("清理临时文件失败: %v\n", err)
	} else {
		fmt.Println("临时文件清理完成")
	}

	// 清理所有自动注册的资源
	if errors := file.CleanupAll(); len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("清理自动资源时发生错误: %v\n", err)
		}
	} else {
		fmt.Println("自动资源清理完成")
	}

	fmt.Println("程序退出")
}