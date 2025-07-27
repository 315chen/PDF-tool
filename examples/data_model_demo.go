//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func main() {
	fmt.Println("=== PDF合并工具数据模型演示 ===\n")

	// 1. 演示MergeJob数据模型
	demonstrateMergeJob()

	// 2. 演示FileList数据模型
	demonstrateFileList()

	// 3. 演示ProgressTracker数据模型
	demonstrateProgressTracker()

	// 4. 演示Config数据模型
	demonstrateConfig()

	// 5. 演示Validator数据模型
	demonstrateValidator()

	fmt.Println("\n=== 数据模型演示完成 ===")
}

func demonstrateMergeJob() {
	fmt.Println("1. MergeJob 数据模型演示:")
	
	// 创建合并任务
	job := model.NewMergeJob(
		"/path/to/main.pdf",
		[]string{"/path/to/file1.pdf", "/path/to/file2.pdf"},
		"/path/to/output.pdf",
	)
	
	fmt.Printf("   任务ID: %s\n", job.ID)
	fmt.Printf("   主文件: %s\n", job.MainFile)
	fmt.Printf("   附加文件数量: %d\n", len(job.AdditionalFiles))
	fmt.Printf("   总文件数: %d\n", job.GetTotalFiles())
	fmt.Printf("   初始状态: %s\n", job.Status.String())
	
	// 模拟任务执行过程
	job.SetRunning()
	fmt.Printf("   运行状态: %s\n", job.Status.String())
	
	job.UpdateProgress(50.0)
	fmt.Printf("   进度更新: %.1f%%\n", job.Progress)
	
	job.SetCompleted()
	fmt.Printf("   完成状态: %s\n", job.Status.String())
	fmt.Printf("   完成时间: %s\n", job.CompletedAt.Format("2006-01-02 15:04:05"))
	
	fmt.Println()
}

func demonstrateFileList() {
	fmt.Println("2. FileList 数据模型演示:")
	
	// 创建文件列表
	fileList := model.NewFileList()
	
	// 设置主文件
	mainFile := fileList.SetMainFile("/path/to/main.pdf")
	mainFile.Size = 1024 * 1024 // 1MB
	mainFile.PageCount = 10
	mainFile.IsEncrypted = false
	
	fmt.Printf("   主文件: %s (大小: %s, 页数: %d)\n", 
		mainFile.DisplayName, mainFile.GetSizeString(), mainFile.PageCount)
	
	// 添加附加文件
	file1 := fileList.AddFile("/path/to/file1.pdf")
	file1.Size = 512 * 1024 // 512KB
	file1.PageCount = 5
	
	file2 := fileList.AddFile("/path/to/file2.pdf")
	file2.Size = 2 * 1024 * 1024 // 2MB
	file2.PageCount = 20
	file2.IsEncrypted = true
	
	fmt.Printf("   附加文件1: %s (大小: %s, 页数: %d)\n", 
		file1.DisplayName, file1.GetSizeString(), file1.PageCount)
	fmt.Printf("   附加文件2: %s (大小: %s, 页数: %d, 加密: %t)\n", 
		file2.DisplayName, file2.GetSizeString(), file2.PageCount, file2.IsEncrypted)
	
	fmt.Printf("   总文件数: %d\n", fileList.TotalCount())
	fmt.Printf("   有效文件数: %d\n", len(fileList.GetValidFiles()))
	
	// 演示文件移动
	fmt.Printf("   移动文件2到位置1: %t\n", fileList.MoveFile("/path/to/file2.pdf", 1))
	
	fmt.Println()
}

func demonstrateProgressTracker() {
	fmt.Println("3. ProgressTracker 数据模型演示:")
	
	// 创建进度跟踪器
	tracker := model.NewProgressTracker(3)
	
	// 添加进度回调
	tracker.AddCallback(func(progress float64, message string) {
		fmt.Printf("   [回调] 进度: %.1f%%, 消息: %s\n", progress, message)
	})
	
	// 模拟进度更新
	tracker.SetCurrentStep(1, "开始验证文件...")
	time.Sleep(100 * time.Millisecond)
	
	tracker.UpdateStepProgress(50, "验证文件中...")
	time.Sleep(100 * time.Millisecond)
	
	tracker.UpdateStepProgress(100, "文件验证完成")
	time.Sleep(100 * time.Millisecond)
	
	tracker.SetCurrentStep(2, "开始合并PDF...")
	time.Sleep(100 * time.Millisecond)
	
	tracker.UpdateStepProgress(75, "合并进行中...")
	time.Sleep(100 * time.Millisecond)
	
	tracker.Complete("PDF合并完成!")
	
	// 获取最终进度信息
	info := tracker.GetProgress()
	fmt.Printf("   最终进度: %.1f%%, 耗时: %v\n", info.TotalProgress, info.ElapsedTime)
	
	fmt.Println()
}

func demonstrateConfig() {
	fmt.Println("4. Config 数据模型演示:")
	
	// 获取默认配置
	config := model.DefaultConfig()
	
	fmt.Printf("   最大内存使用: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	fmt.Printf("   窗口大小: %dx%d\n", config.WindowWidth, config.WindowHeight)
	fmt.Printf("   自动解密: %t\n", config.EnableAutoDecrypt)
	fmt.Printf("   常用密码数量: %d\n", len(config.CommonPasswords))
	fmt.Printf("   前3个常用密码: %v\n", config.CommonPasswords[:3])
	
	fmt.Println()
}

func demonstrateValidator() {
	fmt.Println("5. Validator 数据模型演示:")
	
	validator := model.NewValidator()
	
	// 验证合并任务
	job := model.NewMergeJob(
		"/path/to/main.pdf",
		[]string{"/path/to/file1.pdf"},
		"/path/to/output.pdf",
	)
	
	if err := validator.ValidateMergeJob(job); err != nil {
		fmt.Printf("   合并任务验证失败: %v\n", err)
	} else {
		fmt.Printf("   合并任务验证通过 ✓\n")
	}
	
	// 验证文件条目
	fileEntry := model.NewFileEntry("/path/to/test.pdf", 1)
	fileEntry.Size = 1024
	fileEntry.PageCount = 5
	
	if err := validator.ValidateFileEntry(fileEntry); err != nil {
		fmt.Printf("   文件条目验证失败: %v\n", err)
	} else {
		fmt.Printf("   文件条目验证通过 ✓\n")
	}
	
	// 验证配置
	config := model.DefaultConfig()
	
	if err := validator.ValidateConfig(config); err != nil {
		fmt.Printf("   配置验证失败: %v\n", err)
	} else {
		fmt.Printf("   配置验证通过 ✓\n")
	}
	
	// 验证文件列表
	fileList := model.NewFileList()
	fileList.SetMainFile("/path/to/main.pdf")
	fileList.AddFile("/path/to/file1.pdf")
	
	if err := validator.ValidateFileList(fileList); err != nil {
		fmt.Printf("   文件列表验证失败: %v\n", err)
	} else {
		fmt.Printf("   文件列表验证通过 ✓\n")
	}
	
	// 验证进度跟踪器
	tracker := model.NewProgressTracker(3)
	
	if err := validator.ValidateProgressTracker(tracker); err != nil {
		fmt.Printf("   进度跟踪器验证失败: %v\n", err)
	} else {
		fmt.Printf("   进度跟踪器验证通过 ✓\n")
	}
	
	fmt.Println()
}
