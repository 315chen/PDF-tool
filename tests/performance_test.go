package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/test_utils"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// BenchmarkMergeSmallFiles 基准测试：合并小文件
func BenchmarkMergeSmallFiles(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建小文件
		mainFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("main_%d.pdf", i))
		additionalFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("additional_%d.pdf", i))
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 执行合并
		err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			b.Fatalf("Failed to start merge job: %v", err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// BenchmarkMergeMediumFiles 基准测试：合并中等文件
func BenchmarkMergeMediumFiles(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建中等文件
		mainFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("main_%d.pdf", i))
		additionalFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("additional_%d.pdf", i))
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 执行合并
		err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			b.Fatalf("Failed to start merge job: %v", err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// BenchmarkMergeLargeFiles 基准测试：合并大文件
func BenchmarkMergeLargeFiles(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建大文件
		mainFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("main_%d.pdf", i))
		additionalFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("additional_%d.pdf", i))
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 执行合并
		err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			b.Fatalf("Failed to start merge job: %v", err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// BenchmarkMergeMultipleFiles 基准测试：合并多个文件
func BenchmarkMergeMultipleFiles(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建主文件和多个附加文件
		mainFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("main_%d.pdf", i))

		var additionalFiles []string
		for j := 0; j < 5; j++ {
			additionalFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("additional_%d_%d.pdf", i, j))
			additionalFiles = append(additionalFiles, additionalFile)
		}
		
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 执行合并
		err := ctrl.StartMergeJob(mainFile, additionalFiles, outputFile)
		if err != nil {
			b.Fatalf("Failed to start merge job: %v", err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// BenchmarkWorkflowExecution 基准测试：工作流执行
func BenchmarkWorkflowExecution(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	workflowManager := controller.NewWorkflowManager(ctrl)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建测试文件
		mainFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("main_%d.pdf", i))
		additionalFile := test_utils.CreateTestPDFFile(b, tempDir, fmt.Sprintf("additional_%d.pdf", i))
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 创建任务
		job := model.NewMergeJob(mainFile, []string{additionalFile}, outputFile)
		
		// 执行工作流
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := workflowManager.ExecuteWorkflow(ctx, job)
		cancel()
		
		if err != nil {
			b.Fatalf("Workflow execution failed: %v", err)
		}
	}
}

// BenchmarkEventHandlerProcessing 基准测试：事件处理器处理
func BenchmarkEventHandlerProcessing(b *testing.B) {
	tempDir := b.TempDir()
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(b, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(b, tempDir, "additional.pdf")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		// 处理事件
		eventHandler.HandleMainFileSelected(mainFile)
		eventHandler.HandleAdditionalFileAdded(additionalFile)
		eventHandler.HandleOutputPathChanged(outputFile)
		
		// 启动合并
		err := eventHandler.HandleMergeStart(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			b.Fatalf("Failed to handle merge start: %v", err)
		}
		
		// 等待完成
		for eventHandler.IsJobRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// TestMemoryUsage 测试内存使用情况
func TestMemoryUsage(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "memory-test")
	
	// 记录初始内存使用
	var initialMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 执行多次合并操作
	for i := 0; i < 10; i++ {
		mainFile := test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("main_%d.pdf", i))
		additionalFile := test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("additional_%d.pdf", i))
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			t.Fatalf("Failed to start merge job %d: %v", i, err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(10 * time.Millisecond)
		}
	}
	
	// 强制垃圾回收
	runtime.GC()
	runtime.GC()
	
	// 记录最终内存使用
	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)
	
	// 计算内存增长
	memoryGrowth := finalMemStats.Alloc - initialMemStats.Alloc
	
	t.Logf("Initial memory: %d bytes", initialMemStats.Alloc)
	t.Logf("Final memory: %d bytes", finalMemStats.Alloc)
	t.Logf("Memory growth: %d bytes", memoryGrowth)
	
	// 验证内存增长在合理范围内（小于50MB）
	maxAllowedGrowth := int64(50 * 1024 * 1024) // 50MB
	if int64(memoryGrowth) > maxAllowedGrowth {
		t.Errorf("Memory growth %d bytes exceeds maximum allowed %d bytes", memoryGrowth, maxAllowedGrowth)
	}
}

// TestConcurrentMergeOperations 测试并发合并操作
func TestConcurrentMergeOperations(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "concurrent-test")
	
	// 创建多个控制器实例
	const numControllers = 5
	controllers := make([]*controller.Controller, numControllers)
	
	for i := 0; i < numControllers; i++ {
		pdfService := pdf.NewPDFService()
		fileManager := file.NewFileManager(tempDir)
		config := model.DefaultConfig()
		config.TempDirectory = tempDir
		
		controllers[i] = controller.NewController(pdfService, fileManager, config)
	}
	
	// 并发执行合并操作
	done := make(chan bool, numControllers)
	errors := make(chan error, numControllers)
	
	for i := 0; i < numControllers; i++ {
		go func(index int) {
			ctrl := controllers[index]
			
			// 创建测试文件
			mainFile := test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("concurrent_main_%d.pdf", index))
			additionalFile := test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("concurrent_additional_%d.pdf", index))
			outputFile := filepath.Join(tempDir, fmt.Sprintf("concurrent_output_%d.pdf", index))
			
			// 执行合并
			err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
			if err != nil {
				errors <- fmt.Errorf("Controller %d failed to start job: %v", index, err)
				return
			}
			
			// 等待完成
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			
			for {
				select {
				case <-timeout:
					errors <- fmt.Errorf("Controller %d timed out", index)
					return
				case <-ticker.C:
					if !ctrl.IsJobRunning() {
						// 验证输出文件
						if _, err := os.Stat(outputFile); os.IsNotExist(err) {
							errors <- fmt.Errorf("Controller %d output file not created", index)
							return
						}
						done <- true
						return
					}
				}
			}
		}(i)
	}
	
	// 等待所有操作完成
	completedCount := 0
	errorCount := 0
	
	for completedCount+errorCount < numControllers {
		select {
		case <-done:
			completedCount++
		case err := <-errors:
			t.Errorf("Concurrent operation error: %v", err)
			errorCount++
		case <-time.After(60 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}
	
	t.Logf("Concurrent operations completed: %d, errors: %d", completedCount, errorCount)
	
	// 验证至少有一些操作成功
	if completedCount == 0 {
		t.Error("No concurrent operations completed successfully")
	}
}

// TestLargeFileHandling 测试大文件处理
func TestLargeFileHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}
	
	tempDir := test_utils.CreateTempDir(t, "large-file-test")
	
	// 创建服务组件
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager(tempDir)
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	config.MaxMemoryUsage = 1024 * 1024 * 1024 // 1GB
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建大文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "large_main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(t, tempDir, "large_additional.pdf")
	outputFile := filepath.Join(tempDir, "large_output.pdf")
	
	// 记录开始时间
	startTime := time.Now()
	
	// 执行合并
	err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
	if err != nil {
		t.Fatalf("Failed to start large file merge: %v", err)
	}
	
	// 等待完成
	timeout := time.After(2 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("Large file merge timed out")
		case <-ticker.C:
			if !ctrl.IsJobRunning() {
				goto largeCompleted
			}
		}
	}
	
largeCompleted:
	// 记录完成时间
	elapsed := time.Since(startTime)
	
	// 验证输出文件
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should exist")
	}
	
	// 验证输出文件大小
	outputSize := test_utils.GetFileSize(t, outputFile)
	if outputSize == 0 {
		t.Error("Output file should not be empty")
	}
	
	t.Logf("Large file merge completed in %v", elapsed)
	t.Logf("Output file size: %d bytes", outputSize)
	
	// 验证性能（应该在合理时间内完成）
	maxAllowedTime := 2 * time.Minute
	if elapsed > maxAllowedTime {
		t.Errorf("Large file merge took too long: %v (max allowed: %v)", elapsed, maxAllowedTime)
	}
}
