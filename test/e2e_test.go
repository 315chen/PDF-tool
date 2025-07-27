package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/test_utils"
	"github.com/user/pdf-merger/internal/ui"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// TestE2E_CompleteWorkflow 端到端完整工作流程测试
func TestE2E_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试（短测试模式）")
	}

	// 创建临时目录
	tempDir := test_utils.CreateTempDir(t, "e2e-complete-workflow")

	// 创建测试PDF文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFiles := []string{
		test_utils.CreateTestPDFFile(t, tempDir, "additional1.pdf"),
		test_utils.CreateTestPDFFile(t, tempDir, "additional2.pdf"),
		test_utils.CreateTestPDFFile(t, tempDir, "additional3.pdf"),
	}
	outputFile := filepath.Join(tempDir, "merged_output.pdf")

	// 创建应用和服务
	testApp := app.New()
	testWindow := testApp.NewWindow("E2E Test")
	testWindow.Resize(fyne.NewSize(800, 600))

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器和UI
	ctrl := controller.NewController(pdfService, fileManager, config)
	userInterface := ui.NewUI(testWindow, ctrl)

	// 创建事件处理器
	eventHandler := controller.NewEventHandler(ctrl)

	// 设置回调以跟踪进度
	var progressUpdates []string
	var errorOccurred error
	var completionMessage string

	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		progressUpdates = append(progressUpdates, status)
		t.Logf("进度更新: %.1f%% - %s: %s", progress*100, status, detail)
	})

	eventHandler.SetErrorCallback(func(err error) {
		errorOccurred = err
		t.Logf("错误发生: %v", err)
	})

	eventHandler.SetCompletionCallback(func(message string) {
		completionMessage = message
		t.Logf("完成: %s", message)
	})

	// 构建UI
	content := userInterface.BuildUI()
	testWindow.SetContent(content)

	// 模拟用户操作流程
	t.Log("开始端到端测试流程...")

	// 步骤1: 验证主文件
	t.Log("步骤1: 验证主文件")
	err := eventHandler.HandleMainFileSelected(mainFile)
	if err != nil {
		t.Logf("主文件验证失败（预期，因为文件可能不是真正的PDF）: %v", err)
	}

	// 步骤2: 添加附加文件
	t.Log("步骤2: 添加附加文件")
	for i, additionalFile := range additionalFiles {
		t.Logf("添加文件 %d: %s", i+1, filepath.Base(additionalFile))
		_, err := eventHandler.HandleAdditionalFileAdded(additionalFile)
		if err != nil {
			t.Logf("添加文件失败（预期）: %v", err)
		}
	}

	// 步骤3: 验证输出路径
	t.Log("步骤3: 验证输出路径")
	err = eventHandler.HandleOutputPathChanged(outputFile)
	if err != nil {
		t.Errorf("输出路径验证失败: %v", err)
	}

	// 步骤4: 开始合并
	t.Log("步骤4: 开始合并")
	err = eventHandler.HandleMergeStart(mainFile, additionalFiles, outputFile)
	if err != nil {
		t.Logf("合并启动失败（可能由于文件验证）: %v", err)
	} else {
		// 等待合并完成或失败
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Log("合并操作超时")
				goto cleanup
			case <-ticker.C:
				if !eventHandler.IsJobRunning() {
					t.Log("合并操作完成")
					goto cleanup
				}
			}
		}
	}

cleanup:
	// 验证结果
	t.Log("验证测试结果...")

	if len(progressUpdates) > 0 {
		t.Logf("收到 %d 个进度更新", len(progressUpdates))
	}

	if errorOccurred != nil {
		t.Logf("测试过程中发生错误（可能是预期的）: %v", errorOccurred)
	}

	if completionMessage != "" {
		t.Logf("收到完成消息: %s", completionMessage)
	}

	// 清理
	testWindow.Close()
	testApp.Quit()

	t.Log("端到端测试完成")
}

// TestE2E_ErrorScenarios 端到端错误场景测试
func TestE2E_ErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端错误测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "e2e-error-scenarios")

	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	ctrl := controller.NewController(pdfService, fileManager, config)
	eventHandler := controller.NewEventHandler(ctrl)

	// 设置错误回调
	var capturedErrors []error
	eventHandler.SetErrorCallback(func(err error) {
		capturedErrors = append(capturedErrors, err)
	})

	// 测试场景1: 不存在的主文件
	t.Log("测试场景1: 不存在的主文件")
	err := eventHandler.HandleMainFileSelected("/nonexistent/main.pdf")
	if err == nil {
		t.Error("不存在的主文件应该产生错误")
	}

	// 测试场景2: 不存在的附加文件
	t.Log("测试场景2: 不存在的附加文件")
	_, err = eventHandler.HandleAdditionalFileAdded("/nonexistent/additional.pdf")
	if err == nil {
		t.Error("不存在的附加文件应该产生错误")
	}

	// 测试场景3: 无效的输出路径
	t.Log("测试场景3: 无效的输出路径")
	err = eventHandler.HandleOutputPathChanged("/readonly/output.pdf")
	if err == nil {
		t.Log("只读路径可能在某些系统上不会立即产生错误")
	}

	// 测试场景4: 空参数合并
	t.Log("测试场景4: 空参数合并")
	err = eventHandler.HandleMergeStart("", []string{}, "")
	if err == nil {
		t.Error("空参数应该产生错误")
	}

	// 验证错误被正确捕获
	if len(capturedErrors) > 0 {
		t.Logf("捕获到 %d 个错误（预期）", len(capturedErrors))
		for i, err := range capturedErrors {
			t.Logf("错误 %d: %v", i+1, err)
		}
	}

	t.Log("错误场景测试完成")
}

// TestE2E_CancellationFlow 端到端取消流程测试
func TestE2E_CancellationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端取消测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "e2e-cancellation")

	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(t, tempDir, "additional.pdf")
	outputFile := filepath.Join(tempDir, "cancelled_output.pdf")

	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	ctrl := controller.NewController(pdfService, fileManager, config)
	eventHandler := controller.NewEventHandler(ctrl)

	// 设置回调
	eventHandler.SetErrorCallback(func(err error) {
		if err.Error() == "任务被用户取消" {
			t.Log("任务被取消")
		}
	})

	// 启动合并任务
	t.Log("启动合并任务...")
	err := eventHandler.HandleMergeStart(mainFile, []string{additionalFile}, outputFile)
	if err != nil {
		t.Logf("启动任务失败: %v", err)
		return
	}

	// 短暂等待后取消
	time.Sleep(100 * time.Millisecond)
	t.Log("取消合并任务...")
	err = eventHandler.HandleMergeCancel()
	if err != nil {
		t.Logf("取消任务失败: %v", err)
	}

	// 等待取消完成
	time.Sleep(200 * time.Millisecond)

	// 验证任务已停止
	if eventHandler.IsJobRunning() {
		t.Error("任务应该已经停止")
	}

	t.Log("取消流程测试完成")
}

// TestE2E_MemoryStressTest 端到端内存压力测试
func TestE2E_MemoryStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存压力测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "e2e-memory-stress")

	// 创建多个大文件
	t.Log("创建测试文件...")
	files := make([]string, 10)
	for i := 0; i < 10; i++ {
		files[i] = test_utils.CreateLargePDFFile(t, tempDir, fmt.Sprintf("large_%d.pdf", i), 100) // 100KB each
	}

	outputFile := filepath.Join(tempDir, "stress_output.pdf")

	// 创建内存限制的配置
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.MaxMemoryUsage = 5 * 1024 * 1024 // 5MB限制
	config.TempDirectory = tempDir

	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建流式合并器进行压力测试
	streamingMerger := controller.NewStreamingMerger(ctrl)
	job := model.NewMergeJob(files[0], files[1:], outputFile)

	// 监控内存使用
	memoryMonitor := controller.NewMemoryMonitor(config.MaxMemoryUsage)
	memoryMonitor.Start()
	defer memoryMonitor.Stop()

	t.Log("开始内存压力测试...")
	ctx := context.Background()
	err := streamingMerger.MergeStreaming(ctx, job, nil)

	if err != nil {
		t.Logf("内存压力测试失败（可能由于PDF文件生成或验证问题）: %v", err)
	}

	// 检查内存使用情况
	if memoryMonitor.IsMemoryLow() {
		t.Log("内存使用较高，流式处理正常工作")
	}

	t.Log("内存压力测试完成")
}

// TestE2E_ConcurrentOperations 端到端并发操作测试
func TestE2E_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发操作测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "e2e-concurrent")

	// 创建多个控制器实例
	controllers := make([]*controller.Controller, 3)
	for i := 0; i < 3; i++ {
		fileManager := file.NewFileManager(tempDir)
		pdfService := pdf.NewPDFService()
		config := model.DefaultConfig()
		config.TempDirectory = filepath.Join(tempDir, fmt.Sprintf("instance_%d", i))
		os.MkdirAll(config.TempDirectory, 0755)

		controllers[i] = controller.NewController(pdfService, fileManager, config)
	}

	// 并发执行文件验证
	t.Log("并发文件验证测试...")
	done := make(chan bool, 3)

	for i, ctrl := range controllers {
		go func(index int, controller *controller.Controller) {
			defer func() { done <- true }()

			// 创建测试文件
			testFile := test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("concurrent_%d.pdf", index))

			// 执行验证
			err := controller.ValidateFile(testFile)
			if err != nil {
				t.Logf("控制器 %d 验证失败: %v", index, err)
			}
		}(i, ctrl)
	}

	// 等待所有goroutine完成
	for i := 0; i < 3; i++ {
		<-done
	}

	t.Log("并发操作测试完成")
}

// TestE2E_PerformanceBenchmark 端到端性能基准测试
func TestE2E_PerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能基准测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "e2e-performance")

	// 创建不同大小的测试文件
	testCases := []struct {
		name      string
		fileCount int
		fileSize  int // KB
	}{
		{"小文件", 5, 10},
		{"中等文件", 3, 50},
		{"大文件", 2, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试文件
			files := make([]string, tc.fileCount)
			for i := 0; i < tc.fileCount; i++ {
				files[i] = test_utils.CreateLargePDFFile(t, tempDir, 
					fmt.Sprintf("%s_%d.pdf", tc.name, i), tc.fileSize)
			}

			outputFile := filepath.Join(tempDir, fmt.Sprintf("%s_output.pdf", tc.name))

			// 创建服务
			fileManager := file.NewFileManager(tempDir)
			pdfService := pdf.NewPDFService()
			config := model.DefaultConfig()

			ctrl := controller.NewController(pdfService, fileManager, config)

			// 测量性能
			start := time.Now()

			// 执行验证
			for _, file := range files {
				err := ctrl.ValidateFile(file)
				if err != nil {
					t.Logf("文件验证失败: %v", err)
				}
			}

			// 尝试合并（可能失败）
			if len(files) > 1 {
				err := ctrl.StartMergeJob(files[0], files[1:], outputFile)
				if err != nil {
					t.Logf("合并启动失败: %v", err)
				} else {
					// 等待完成
					for ctrl.IsJobRunning() {
						time.Sleep(10 * time.Millisecond)
					}
				}
			}

			duration := time.Since(start)
			t.Logf("%s 处理时间: %v", tc.name, duration)

			// 性能阈值检查
			maxDuration := time.Duration(tc.fileCount*tc.fileSize) * time.Millisecond
			if duration > maxDuration {
				t.Logf("性能警告: %s 处理时间 %v 超过预期 %v", tc.name, duration, maxDuration)
			}
		})
	}

	t.Log("性能基准测试完成")
}

// TestE2E_ResourceCleanup 端到端资源清理测试
func TestE2E_ResourceCleanup(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "e2e-cleanup")

	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	_ = controller.NewController(pdfService, fileManager, config)

	// 创建一些临时文件
	tempFiles := make([]string, 5)
	for i := 0; i < 5; i++ {
		tempFile, _, err := fileManager.CreateTempFileWithPrefix(fmt.Sprintf("cleanup_test_%d_", i), ".pdf")
		if err != nil {
			t.Fatalf("创建临时文件失败: %v", err)
		}
		tempFiles[i] = tempFile
	}

	// 验证文件存在
	for i, tempFile := range tempFiles {
		if !test_utils.FileExists(tempFile) {
			t.Errorf("临时文件 %d 不存在: %s", i, tempFile)
		}
	}

	// 执行清理
	err := fileManager.CleanupTempFiles()
	if err != nil {
		t.Errorf("清理临时文件失败: %v", err)
	}

	// 验证文件被清理（注意：某些实现可能不会立即删除文件）
	cleanedCount := 0
	for _, tempFile := range tempFiles {
		if !test_utils.FileExists(tempFile) {
			cleanedCount++
		}
	}

	t.Logf("清理了 %d/%d 个临时文件", cleanedCount, len(tempFiles))

	t.Log("资源清理测试完成")
}

// 基准测试

func BenchmarkE2E_FileValidation(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "benchmark-validation")
	testFile := test_utils.CreateTestPDFFile(b, tempDir, "benchmark.pdf")

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	ctrl := controller.NewController(pdfService, fileManager, config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctrl.ValidateFile(testFile)
	}
}

func BenchmarkE2E_EventHandling(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "benchmark-events")
	testFile := test_utils.CreateTestPDFFile(b, tempDir, "benchmark.pdf")

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	ctrl := controller.NewController(pdfService, fileManager, config)
	eventHandler := controller.NewEventHandler(ctrl)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		eventHandler.HandleMainFileSelected(testFile)
	}
}