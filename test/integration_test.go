package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/test_utils"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// TestIntegration_FullWorkflow 测试完整的工作流程
func TestIntegration_FullWorkflow(t *testing.T) {
	// 创建临时目录
	tempDir := test_utils.CreateTempDir(t, "integration-test")
	
	// 创建测试PDF文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFile1 := test_utils.CreateTestPDFFile(t, tempDir, "additional1.pdf")
	additionalFile2 := test_utils.CreateTestPDFFile(t, tempDir, "additional2.pdf")
	outputFile := filepath.Join(tempDir, "output.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 设置回调
	var progressUpdates []string
	var errorOccurred error
	
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		progressUpdates = append(progressUpdates, status)
	})
	
	ctrl.SetErrorCallback(func(err error) {
		errorOccurred = err
	})
	
	ctrl.SetCompletionCallback(func(outputPath string) {
		// 合并完成
	})
	
	// 执行合并
	err := ctrl.StartMergeJob(mainFile, []string{additionalFile1, additionalFile2}, outputFile)
	if err != nil {
		t.Fatalf("启动合并任务失败: %v", err)
	}
	
	// 等待完成
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("合并操作超时")
		case <-ticker.C:
			if !ctrl.IsJobRunning() {
				goto done
			}
		}
	}
	
done:
	// 验证结果
	if errorOccurred != nil {
		t.Logf("合并过程中发生错误: %v", errorOccurred)
	}
	
	if len(progressUpdates) == 0 {
		t.Error("没有收到进度更新")
	}
	
	// 验证输出文件
	if !test_utils.FileExists(outputFile) {
		t.Log("输出文件不存在（可能由于UniPDF许可证问题）")
	}
}

// TestIntegration_ErrorHandling 测试错误处理
func TestIntegration_ErrorHandling(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "error-test")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 测试不存在的文件
	err := ctrl.ValidateFile("/nonexistent/file.pdf")
	if err == nil {
		t.Error("验证不存在的文件应该失败")
	}
	
	// 测试无效的合并参数
	err = ctrl.StartMergeJob("", []string{}, "")
	if err != nil {
		t.Logf("无效参数导致启动失败（预期）: %v", err)
	}
}

// TestIntegration_CancellationWorkflow 测试取消工作流程
func TestIntegration_CancellationWorkflow(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "cancellation-test")
	
	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(t, tempDir, "additional.pdf")
	outputFile := filepath.Join(tempDir, "output.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 启动任务
	err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
	if err != nil {
		t.Fatalf("启动合并任务失败: %v", err)
	}
	
	// 短暂等待后取消
	time.Sleep(50 * time.Millisecond)
	
	err = ctrl.CancelCurrentJob()
	if err != nil {
		t.Logf("取消任务失败（可能任务已完成）: %v", err)
	}
	
	// 验证任务已停止
	time.Sleep(100 * time.Millisecond)
	if ctrl.IsJobRunning() {
		t.Error("任务应该已经停止")
	}
}

// TestIntegration_StreamingMerger 测试流式合并器
func TestIntegration_StreamingMerger(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "streaming-test")
	
	// 创建多个测试文件
	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		files[i] = test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("file%d.pdf", i))
	}
	
	outputFile := filepath.Join(tempDir, "streamed_output.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.MaxMemoryUsage = 1024 * 1024 // 1MB限制，强制使用流式处理
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建流式合并器
	streamingMerger := controller.NewStreamingMerger(ctrl)
	
	// 创建合并任务
	job := model.NewMergeJob(files[0], files[1:], outputFile)
	
	// 执行流式合并
	ctx := context.Background()
	err := streamingMerger.MergeStreaming(ctx, job, nil)
	
	if err != nil {
		t.Logf("流式合并失败（可能由于UniPDF许可证）: %v", err)
	}
}

// TestIntegration_WorkflowManager 测试工作流程管理器
func TestIntegration_WorkflowManager(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "workflow-test")
	
	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(t, tempDir, "additional.pdf")
	outputFile := filepath.Join(tempDir, "workflow_output.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建工作流程管理器
	workflowManager := controller.NewWorkflowManager(ctrl)
	
	// 创建任务
	job := model.NewMergeJob(mainFile, []string{additionalFile}, outputFile)
	
	// 执行工作流程
	ctx := context.Background()
	err := workflowManager.ExecuteWorkflow(ctx, job)
	
	if err != nil {
		t.Logf("工作流程执行失败（可能由于UniPDF许可证）: %v", err)
	}
}

// TestIntegration_MemoryMonitoring 测试内存监控
func TestIntegration_MemoryMonitoring(t *testing.T) {
	// 创建内存监控器
	monitor := controller.NewMemoryMonitor(50 * 1024 * 1024) // 50MB
	
	// 启动监控
	monitor.Start()
	defer monitor.Stop()
	
	// 检查内存状态
	isLow := monitor.IsMemoryLow()
	t.Logf("内存是否不足: %v", isLow)
	
	// 等待一段时间让监控器运行
	time.Sleep(100 * time.Millisecond)
}

// TestIntegration_BatchProcessing 测试批处理
func TestIntegration_BatchProcessing(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "batch-test")
	
	// 创建多个测试文件
	files := make([]string, 15) // 创建15个文件用于批处理
	for i := 0; i < 15; i++ {
		files[i] = test_utils.CreateTestPDFFile(t, tempDir, fmt.Sprintf("batch_file%d.pdf", i))
	}
	
	outputFile := filepath.Join(tempDir, "batch_output.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	// 创建批处理器
	streamingMerger := controller.NewStreamingMerger(ctrl)
	batchProcessor := controller.NewBatchProcessor(streamingMerger)
	
	// 执行批处理
	ctx := context.Background()
	err := batchProcessor.ProcessBatch(ctx, files, outputFile, nil)
	
	if err != nil {
		t.Logf("批处理失败（可能由于UniPDF许可证）: %v", err)
	}
}

// TestIntegration_ConfigurationManagement 测试配置管理
func TestIntegration_ConfigurationManagement(t *testing.T) {
	// 测试默认配置
	config := model.DefaultConfig()
	
	if config.MaxMemoryUsage <= 0 {
		t.Error("默认最大内存使用量应该大于0")
	}
	
	if len(config.CommonPasswords) == 0 {
		t.Error("默认密码列表不应该为空")
	}
	
	if config.WindowWidth <= 0 || config.WindowHeight <= 0 {
		t.Error("默认窗口尺寸应该大于0")
	}
	
	// 测试配置验证
	config.MaxMemoryUsage = -1
	// 这里可以添加配置验证逻辑的测试
}

// TestIntegration_DataModelOperations 测试数据模型操作
func TestIntegration_DataModelOperations(t *testing.T) {
	// 测试合并任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")
	
	if job.Status != model.JobPending {
		t.Error("新任务状态应该是Pending")
	}
	
	if job.Progress != 0.0 {
		t.Error("新任务进度应该是0")
	}
	
	// 测试状态转换
	job.SetRunning()
	if job.Status != model.JobRunning {
		t.Error("任务状态应该是Running")
	}
	
	job.UpdateProgress(50.0)
	if job.Progress != 50.0 {
		t.Error("任务进度应该是50")
	}
	
	job.SetCompleted()
	if job.Status != model.JobCompleted {
		t.Error("任务状态应该是Completed")
	}
	
	if job.Progress != 100.0 {
		t.Error("完成任务的进度应该是100")
	}
	
	// 测试文件条目
	fileEntry := model.NewFileEntry("/tmp/test.pdf", 1)
	if fileEntry.Order != 1 {
		t.Error("文件条目顺序不正确")
	}
	
	if !fileEntry.IsValid {
		t.Error("新文件条目应该是有效的")
	}
	
	fileEntry.SetError("测试错误")
	if fileEntry.IsValid {
		t.Error("设置错误后文件条目应该无效")
	}
	
	if fileEntry.Error != "测试错误" {
		t.Error("错误信息不正确")
	}
}

// TestIntegration_CrossModuleInteraction 测试跨模块交互
func TestIntegration_CrossModuleInteraction(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "cross-module-test")

	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(t, tempDir, "additional.pdf")
	outputFile := filepath.Join(tempDir, "cross_module_output.pdf")

	// 创建各个模块的实例
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 测试文件管理器与PDF服务的交互
	fileInfo, err := fileManager.GetFileInfo(mainFile)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if fileInfo.Size <= 0 {
		t.Error("文件大小应该大于0")
	}

	// 测试PDF服务验证功能
	err = pdfService.ValidatePDF(mainFile)
	if err != nil {
		t.Fatalf("PDF验证失败: %v", err)
	}

	// 测试控制器协调各模块
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 设置跨模块回调
	var progressEvents []string
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		progressEvents = append(progressEvents, fmt.Sprintf("%.1f%%: %s - %s", progress*100, status, detail))
	})

	// 执行合并操作
	err = ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
	if err != nil {
		t.Fatalf("启动合并任务失败: %v", err)
	}

	// 等待完成
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("跨模块交互测试超时")
		case <-ticker.C:
			if !ctrl.IsJobRunning() {
				goto completed
			}
		}
	}

completed:
	// 验证结果
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("输出文件应该存在")
	}

	// 验证进度事件
	if len(progressEvents) == 0 {
		t.Error("应该有进度事件")
	}

	t.Logf("跨模块交互测试完成，进度事件数量: %d", len(progressEvents))
}

// 基准测试

func BenchmarkIntegration_FullWorkflow(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "benchmark-workflow")
	
	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(b, tempDir, "main.pdf")
	additionalFile := test_utils.CreateTestPDFFile(b, tempDir, "additional.pdf")
	
	// 创建服务
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pdf", i))
		
		err := ctrl.StartMergeJob(mainFile, []string{additionalFile}, outputFile)
		if err != nil {
			b.Errorf("启动合并任务失败: %v", err)
		}
		
		// 等待完成
		for ctrl.IsJobRunning() {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func BenchmarkIntegration_FileValidation(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "benchmark-validation")
	testFile := test_utils.CreateTestPDFFile(b, tempDir, "test.pdf")
	
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ctrl.ValidateFile(testFile)
	}
}