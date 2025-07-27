//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/pkg/file"
)

func main() {
	fmt.Println("=== PDF合并工具临时文件管理演示 ===\n")

	// 1. 演示临时文件管理器基本功能
	demonstrateTempFileManager()

	// 2. 演示资源管理器功能
	demonstrateResourceManager()

	// 3. 演示自动清理器功能
	demonstrateAutoCleaner()

	// 4. 演示综合临时文件处理流程
	demonstrateComprehensiveFlow()

	fmt.Println("\n=== 临时文件管理演示完成 ===")
}

func demonstrateTempFileManager() {
	fmt.Println("1. 临时文件管理器基本功能演示:")
	
	// 创建临时文件管理器
	tempManager, err := file.NewTempFileManager("")
	if err != nil {
		fmt.Printf("   创建临时文件管理器失败: %v\n", err)
		return
	}
	defer tempManager.Close()
	
	fmt.Printf("   会话目录: %s\n", tempManager.GetSessionDir())
	
	// 1.1 创建临时文件
	fmt.Println("\n   1.1 创建临时文件:")
	tempPath1, tempFile1, err := tempManager.CreateTempFile("pdf_", ".pdf")
	if err != nil {
		fmt.Printf("   创建临时文件失败: %v\n", err)
		return
	}
	tempFile1.Close()
	
	fmt.Printf("   - 创建临时文件: %s\n", filepath.Base(tempPath1))
	fmt.Printf("   - 当前文件数量: %d\n", tempManager.GetFileCount())
	
	// 1.2 创建带内容的临时文件
	fmt.Println("\n   1.2 创建带内容的临时文件:")
	content := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\n%%EOF")
	tempPath2, err := tempManager.CreateTempFileWithContent("content_", ".pdf", content)
	if err != nil {
		fmt.Printf("   创建带内容的临时文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("   - 创建带内容的临时文件: %s\n", filepath.Base(tempPath2))
	fmt.Printf("   - 文件大小: %d 字节\n", len(content))
	fmt.Printf("   - 当前文件数量: %d\n", tempManager.GetFileCount())
	
	// 1.3 复制文件到临时文件
	fmt.Println("\n   1.3 复制文件到临时文件:")
	
	// 先创建一个源文件
	sourceDir, _ := os.MkdirTemp("", "demo-source")
	defer os.RemoveAll(sourceDir)
	
	sourcePath := filepath.Join(sourceDir, "source.pdf")
	os.WriteFile(sourcePath, content, 0644)
	
	tempPath3, err := tempManager.CopyToTempFile(sourcePath, "copied_")
	if err != nil {
		fmt.Printf("   复制文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("   - 源文件: %s\n", filepath.Base(sourcePath))
	fmt.Printf("   - 复制到: %s\n", filepath.Base(tempPath3))
	fmt.Printf("   - 当前文件数量: %d\n", tempManager.GetFileCount())
	
	// 1.4 删除特定临时文件
	fmt.Println("\n   1.4 删除特定临时文件:")
	err = tempManager.RemoveFile(tempPath1)
	if err != nil {
		fmt.Printf("   删除文件失败: %v\n", err)
	} else {
		fmt.Printf("   - 成功删除: %s\n", filepath.Base(tempPath1))
		fmt.Printf("   - 当前文件数量: %d\n", tempManager.GetFileCount())
	}
	
	// 1.5 设置文件最大保留时间
	fmt.Println("\n   1.5 设置文件最大保留时间:")
	tempManager.SetMaxAge(5 * time.Second)
	fmt.Println("   - 设置最大保留时间为5秒")
	
	// 等待一段时间后清理过期文件
	fmt.Println("   - 等待6秒后清理过期文件...")
	time.Sleep(6 * time.Second)
	
	tempManager.CleanupExpired()
	fmt.Printf("   - 清理后文件数量: %d\n", tempManager.GetFileCount())
	
	fmt.Println()
}

func demonstrateResourceManager() {
	fmt.Println("2. 资源管理器功能演示:")
	
	// 创建资源管理器
	resourceManager := file.NewResourceManager()
	
	// 创建测试文件和目录
	testDir, _ := os.MkdirTemp("", "resource-demo")
	
	testFile1 := filepath.Join(testDir, "test1.txt")
	testFile2 := filepath.Join(testDir, "test2.txt")
	testSubDir := filepath.Join(testDir, "subdir")
	
	os.WriteFile(testFile1, []byte("test content 1"), 0644)
	os.WriteFile(testFile2, []byte("test content 2"), 0644)
	os.MkdirAll(testSubDir, 0755)
	
	fmt.Printf("   创建测试目录: %s\n", testDir)
	
	// 2.1 添加文件资源
	fmt.Println("\n   2.1 添加资源到管理器:")
	resourceManager.AddFile(testFile1, 1)
	resourceManager.AddFile(testFile2, 2)
	resourceManager.AddDirectory(testSubDir, 3)
	
	// 添加自定义资源
	customCleanupCalled := false
	resourceManager.AddCustom(func() error {
		customCleanupCalled = true
		fmt.Println("   - 执行自定义清理函数")
		return nil
	}, 4)
	
	fmt.Printf("   - 添加文件资源: %s (优先级: 1)\n", filepath.Base(testFile1))
	fmt.Printf("   - 添加文件资源: %s (优先级: 2)\n", filepath.Base(testFile2))
	fmt.Printf("   - 添加目录资源: %s (优先级: 3)\n", filepath.Base(testSubDir))
	fmt.Printf("   - 添加自定义资源 (优先级: 4)\n")
	fmt.Printf("   - 当前资源数量: %d\n", resourceManager.GetResourceCount())
	
	// 2.2 清理特定资源
	fmt.Println("\n   2.2 清理特定资源:")
	err := resourceManager.CleanupResource(testFile1)
	if err != nil {
		fmt.Printf("   清理资源失败: %v\n", err)
	} else {
		fmt.Printf("   - 成功清理: %s\n", filepath.Base(testFile1))
		fmt.Printf("   - 剩余资源数量: %d\n", resourceManager.GetResourceCount())
	}
	
	// 2.3 清理所有资源
	fmt.Println("\n   2.3 清理所有资源 (按优先级从高到低):")
	errors := resourceManager.Cleanup()
	
	if len(errors) > 0 {
		fmt.Printf("   清理过程中出现 %d 个错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("   - %v\n", err)
		}
	} else {
		fmt.Println("   - 所有资源清理成功")
	}
	
	fmt.Printf("   - 自定义清理函数是否被调用: %t\n", customCleanupCalled)
	fmt.Printf("   - 最终资源数量: %d\n", resourceManager.GetResourceCount())
	
	// 清理测试目录
	os.RemoveAll(testDir)
	
	fmt.Println()
}

func demonstrateAutoCleaner() {
	fmt.Println("3. 自动清理器功能演示:")
	
	// 创建自动清理器
	autoCleaner := file.NewAutoCleaner()
	
	// 创建测试资源
	testDir, _ := os.MkdirTemp("", "auto-cleaner-demo")
	testFile := filepath.Join(testDir, "auto_test.txt")
	os.WriteFile(testFile, []byte("auto cleaner test"), 0644)
	
	fmt.Printf("   创建测试文件: %s\n", testFile)
	
	// 3.1 添加资源到自动清理器
	fmt.Println("\n   3.1 添加资源到自动清理器:")
	autoCleaner.AddFile(testFile, 1)
	autoCleaner.AddDirectory(testDir, 2)
	
	// 添加自定义清理任务
	cleanupLog := ""
	autoCleaner.AddCustom(func() error {
		cleanupLog = "自动清理器执行了自定义清理任务"
		fmt.Println("   - 执行自定义清理任务")
		return nil
	}, 3)
	
	fmt.Printf("   - 添加文件: %s\n", filepath.Base(testFile))
	fmt.Printf("   - 添加目录: %s\n", filepath.Base(testDir))
	fmt.Printf("   - 添加自定义任务\n")
	fmt.Printf("   - 当前资源数量: %d\n", autoCleaner.GetResourceCount())
	
	// 3.2 手动触发清理
	fmt.Println("\n   3.2 手动触发清理:")
	errors := autoCleaner.Cleanup()
	
	if len(errors) > 0 {
		fmt.Printf("   清理过程中出现 %d 个错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("   - %v\n", err)
		}
	} else {
		fmt.Println("   - 所有资源清理成功")
	}
	
	fmt.Printf("   - 清理日志: %s\n", cleanupLog)
	fmt.Printf("   - 最终资源数量: %d\n", autoCleaner.GetResourceCount())
	
	fmt.Println()
}

func demonstrateComprehensiveFlow() {
	fmt.Println("4. 综合临时文件处理流程演示:")
	
	// 4.1 初始化管理器
	fmt.Println("   4.1 初始化管理器:")
	tempManager, err := file.NewTempFileManager("")
	if err != nil {
		fmt.Printf("   初始化失败: %v\n", err)
		return
	}
	defer tempManager.Close()
	
	resourceManager := file.NewResourceManager()
	
	fmt.Println("   - 临时文件管理器初始化完成")
	fmt.Println("   - 资源管理器初始化完成")
	
	// 4.2 模拟PDF处理流程
	fmt.Println("\n   4.2 模拟PDF处理流程:")
	
	// 创建主PDF文件
	mainPDFContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Count 1\n>>\nendobj\n%%EOF")
	mainPDFPath, err := tempManager.CreateTempFileWithContent("main_", ".pdf", mainPDFContent)
	if err != nil {
		fmt.Printf("   创建主PDF失败: %v\n", err)
		return
	}
	resourceManager.AddFile(mainPDFPath, 1)
	fmt.Printf("   - 创建主PDF: %s\n", filepath.Base(mainPDFPath))
	
	// 创建附加PDF文件
	additionalPDFs := make([]string, 3)
	for i := 0; i < 3; i++ {
		content := fmt.Sprintf("%%PDF-1.4\n%% Additional PDF %d\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\n%%%%EOF", i+1)
		path, err := tempManager.CreateTempFileWithContent(fmt.Sprintf("additional_%d_", i+1), ".pdf", []byte(content))
		if err != nil {
			fmt.Printf("   创建附加PDF %d失败: %v\n", i+1, err)
			continue
		}
		additionalPDFs[i] = path
		resourceManager.AddFile(path, 2)
		fmt.Printf("   - 创建附加PDF %d: %s\n", i+1, filepath.Base(path))
	}
	
	// 创建输出文件
	outputPath, outputFile, err := tempManager.CreateTempFile("merged_", ".pdf")
	if err != nil {
		fmt.Printf("   创建输出文件失败: %v\n", err)
		return
	}
	outputFile.Close()
	resourceManager.AddFile(outputPath, 3)
	fmt.Printf("   - 创建输出文件: %s\n", filepath.Base(outputPath))
	
	// 4.3 显示处理状态
	fmt.Println("\n   4.3 处理状态:")
	fmt.Printf("   - 临时文件数量: %d\n", tempManager.GetFileCount())
	fmt.Printf("   - 资源管理器中的资源数量: %d\n", resourceManager.GetResourceCount())
	fmt.Printf("   - 会话目录: %s\n", tempManager.GetSessionDir())
	
	// 4.4 模拟处理完成后的清理
	fmt.Println("\n   4.4 处理完成，开始清理:")
	
	// 首先清理资源管理器中的资源
	errors := resourceManager.Cleanup()
	if len(errors) > 0 {
		fmt.Printf("   资源清理出现 %d 个错误\n", len(errors))
	} else {
		fmt.Println("   - 资源管理器清理完成")
	}
	
	// 然后清理临时文件管理器
	tempManager.Cleanup()
	fmt.Println("   - 临时文件管理器清理完成")
	
	fmt.Printf("   - 最终临时文件数量: %d\n", tempManager.GetFileCount())
	fmt.Printf("   - 最终资源数量: %d\n", resourceManager.GetResourceCount())
	
	fmt.Println("\n   综合处理流程完成 🎉")
	fmt.Println("   所有临时资源已安全清理")
	
	fmt.Println()
}
