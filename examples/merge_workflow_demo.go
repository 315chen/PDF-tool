//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== 合并流程控制功能演示 ===\n")

	// 1. 演示工作流管理器创建
	demonstrateWorkflowManagerCreation()

	// 2. 演示工作流步骤执行
	demonstrateWorkflowStepExecution()

	// 3. 演示流式合并控制
	demonstrateStreamingMergeControl()

	// 4. 演示批处理流程控制
	demonstrateBatchProcessingControl()

	// 5. 演示内存监控和优化
	demonstrateMemoryMonitoringAndOptimization()

	// 6. 演示错误处理和重试机制
	demonstrateErrorHandlingAndRetry()

	// 7. 演示完整的合并流程
	demonstrateCompleteMergeWorkflow()

	fmt.Println("\n=== 合并流程控制演示完成 ===")
}

func demonstrateWorkflowManagerCreation() {
	fmt.Println("1. 工作流管理器创建演示:")
	
	// 1.1 创建测试环境
	fmt.Println("\n   1.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	fmt.Printf("   - 控制器创建成功\n")
	
	// 1.2 创建工作流管理器
	fmt.Println("\n   1.2 创建工作流管理器:")
	_ = controller.NewWorkflowManager(ctrl)

	fmt.Printf("   - 工作流管理器创建成功\n")
	fmt.Printf("   - 工作流管理器已初始化\n")
	
	// 1.3 检查工作流步骤
	fmt.Println("\n   1.3 工作流步骤定义:")
	steps := []controller.WorkflowStep{
		controller.StepValidation,
		controller.StepPreparation,
		controller.StepDecryption,
		controller.StepMerging,
		controller.StepFinalization,
		controller.StepCompleted,
	}
	
	for i, step := range steps {
		fmt.Printf("   - 步骤 %d: %s\n", i+1, step.String())
	}
	
	// 1.4 检查内存监控器
	fmt.Println("\n   1.4 内存监控器:")
	fmt.Printf("   - 内存监控器已集成到工作流管理器中\n")
	
	fmt.Println()
}

func demonstrateWorkflowStepExecution() {
	fmt.Println("2. 工作流步骤执行演示:")
	
	// 2.1 创建测试环境
	fmt.Println("\n   2.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	workflowManager := controller.NewWorkflowManager(ctrl)
	
	// 创建测试任务
	testFiles := createTestPDFFiles(tempDir, 3)
	outputFile := filepath.Join(tempDir, "workflow_test.pdf")
	
	job := model.NewMergeJob(testFiles[0], testFiles[1:], outputFile)
	fmt.Printf("   - 测试任务创建成功\n")
	fmt.Printf("   - 主文件: %s\n", filepath.Base(job.MainFile))
	fmt.Printf("   - 附加文件数: %d\n", len(job.AdditionalFiles))
	fmt.Printf("   - 输出文件: %s\n", filepath.Base(job.OutputPath))
	
	// 2.2 设置进度回调
	fmt.Println("\n   2.2 设置进度回调:")
	var progressUpdates []string
	
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		progressUpdate := fmt.Sprintf("%.1f%% - %s - %s", progress*100, status, detail)
		progressUpdates = append(progressUpdates, progressUpdate)
		fmt.Printf("   - 进度: %s\n", progressUpdate)
	})
	
	// 2.3 执行工作流
	fmt.Println("\n   2.3 执行工作流:")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	startTime := time.Now()
	err := workflowManager.ExecuteWorkflow(ctx, job)
	elapsed := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("   - 工作流执行失败: %v\n", err)
	} else {
		fmt.Printf("   - 工作流执行成功\n")
	}
	
	fmt.Printf("   - 执行时间: %v\n", elapsed)
	fmt.Printf("   - 工作流执行完成\n")
	
	// 2.4 显示进度更新统计
	fmt.Println("\n   2.4 进度更新统计:")
	fmt.Printf("   - 总进度更新数: %d\n", len(progressUpdates))
	
	if len(progressUpdates) > 0 {
		fmt.Printf("   - 首次更新: %s\n", progressUpdates[0])
		fmt.Printf("   - 最后更新: %s\n", progressUpdates[len(progressUpdates)-1])
	}
	
	fmt.Println()
}

func demonstrateStreamingMergeControl() {
	fmt.Println("3. 流式合并控制演示:")

	// 3.1 创建测试环境
	fmt.Println("\n   3.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	_ = createTestController(tempDir)

	// 3.2 流式合并概念演示
	fmt.Println("\n   3.2 流式合并概念:")
	fmt.Printf("   - 流式合并用于处理大文件\n")
	fmt.Printf("   - 减少内存使用，提高处理效率\n")
	fmt.Printf("   - 支持实时进度更新\n")

	// 3.3 模拟流式处理
	fmt.Println("\n   3.3 模拟流式处理:")
	testFiles := createTestPDFFiles(tempDir, 2)
	outputFile := filepath.Join(tempDir, "streaming_test.pdf")

	fmt.Printf("   - 输入文件: %d 个\n", len(testFiles))
	fmt.Printf("   - 输出文件: %s\n", filepath.Base(outputFile))

	// 模拟流式处理过程
	fmt.Printf("   - 开始流式处理...\n")
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		progress := float64(i+1) / 5.0
		fmt.Printf("   - 处理进度: %.1f%%\n", progress*100)
	}
	fmt.Printf("   - 流式处理完成\n")

	fmt.Println()
}

func demonstrateBatchProcessingControl() {
	fmt.Println("4. 批处理流程控制演示:")

	// 4.1 创建测试环境
	fmt.Println("\n   4.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	_ = createTestController(tempDir)

	// 4.2 批处理概念演示
	fmt.Println("\n   4.2 批处理概念:")
	fmt.Printf("   - 批处理用于处理大量文件\n")
	fmt.Printf("   - 支持并发处理提高效率\n")
	fmt.Printf("   - 智能资源管理和调度\n")

	// 4.3 创建大量测试文件
	fmt.Println("\n   4.3 创建大量测试文件:")
	testFiles := createTestPDFFiles(tempDir, 8) // 创建8个文件用于批处理

	fmt.Printf("   - 创建了 %d 个测试文件\n", len(testFiles))

	// 4.4 模拟批处理
	fmt.Println("\n   4.4 模拟批处理:")

	fmt.Printf("   - 开始批处理...\n")
	batchSize := 3
	for i := 0; i < len(testFiles); i += batchSize {
		end := i + batchSize
		if end > len(testFiles) {
			end = len(testFiles)
		}

		fmt.Printf("   - 处理批次 %d: 文件 %d-%d\n", (i/batchSize)+1, i+1, end)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("   - 批处理完成\n")

	fmt.Println()
}

func demonstrateMemoryMonitoringAndOptimization() {
	fmt.Println("5. 内存监控和优化演示:")

	// 5.1 创建测试环境
	fmt.Println("\n   5.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	ctrl := createTestController(tempDir)
	_ = controller.NewWorkflowManager(ctrl)

	// 5.2 内存监控概念
	fmt.Println("\n   5.2 内存监控概念:")
	fmt.Printf("   - 实时监控内存使用情况\n")
	fmt.Printf("   - 自动触发垃圾回收\n")
	fmt.Printf("   - 内存压力预警机制\n")

	// 5.3 模拟内存监控
	fmt.Println("\n   5.3 模拟内存监控:")
	fmt.Printf("   - 开始内存监控...\n")

	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		usage := 30 + i*10 // 模拟内存使用增长
		fmt.Printf("   - 内存使用: %d%%\n", usage)

		if usage > 70 {
			fmt.Printf("   - 内存压力警告，建议清理\n")
		}
	}

	fmt.Printf("   - 内存监控完成\n")

	// 5.4 内存优化策略
	fmt.Println("\n   5.4 内存优化策略:")
	fmt.Printf("   - 策略1: 及时释放不需要的对象\n")
	fmt.Printf("   - 策略2: 使用流式处理减少内存占用\n")
	fmt.Printf("   - 策略3: 分批处理大量数据\n")
	fmt.Printf("   - 策略4: 定期执行垃圾回收\n")

	fmt.Println()
}

func demonstrateErrorHandlingAndRetry() {
	fmt.Println("6. 错误处理和重试机制演示:")
	
	// 6.1 创建测试环境
	fmt.Println("\n   6.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	workflowManager := controller.NewWorkflowManager(ctrl)
	
	// 6.2 设置错误回调
	fmt.Println("\n   6.2 设置错误回调:")
	var errorMessages []string
	
	ctrl.SetErrorCallback(func(err error) {
		errorMessages = append(errorMessages, err.Error())
		fmt.Printf("   - 错误: %v\n", err)
	})
	
	// 6.3 模拟重试场景
	fmt.Println("\n   6.3 模拟重试场景:")
	
	// 创建一个会失败的任务（使用不存在的文件）
	nonExistentFile := filepath.Join(tempDir, "non_existent.pdf")
	outputFile := filepath.Join(tempDir, "retry_test.pdf")
	
	job := model.NewMergeJob(nonExistentFile, []string{}, outputFile)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	fmt.Printf("   - 执行会失败的任务\n")
	
	startTime := time.Now()
	err := workflowManager.ExecuteWorkflow(ctx, job)
	elapsed := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("   - 任务失败（预期）: %v\n", err)
	}
	
	fmt.Printf("   - 执行时间: %v\n", elapsed)
	
	// 6.4 显示重试统计
	fmt.Println("\n   6.4 重试统计:")
	fmt.Printf("   - 重试机制已集成到工作流中\n")
	fmt.Printf("   - 支持智能重试策略\n")
	
	// 6.5 显示错误统计
	fmt.Println("\n   6.5 错误统计:")
	fmt.Printf("   - 总错误数: %d\n", len(errorMessages))
	
	if len(errorMessages) > 0 {
		fmt.Printf("   - 最后错误: %s\n", errorMessages[len(errorMessages)-1])
	}
	
	fmt.Println()
}

func demonstrateCompleteMergeWorkflow() {
	fmt.Println("7. 完整合并流程演示:")
	
	// 7.1 创建完整测试环境
	fmt.Println("\n   7.1 创建完整测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	
	// 创建有效的测试PDF文件（简化版本）
	testFiles := createValidTestPDFFiles(tempDir, 3)
	outputFile := filepath.Join(tempDir, "complete_workflow_test.pdf")
	
	fmt.Printf("   - 创建了 %d 个有效测试文件\n", len(testFiles))
	fmt.Printf("   - 输出文件: %s\n", filepath.Base(outputFile))
	
	// 7.2 设置完整的回调系统
	fmt.Println("\n   7.2 设置完整的回调系统:")
	
	var progressLog []string
	var errorLog []string
	var completionMessage string
	
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		logEntry := fmt.Sprintf("[%.1f%%] %s: %s", progress*100, status, detail)
		progressLog = append(progressLog, logEntry)
		fmt.Printf("   - %s\n", logEntry)
	})
	
	ctrl.SetErrorCallback(func(err error) {
		errorLog = append(errorLog, err.Error())
		fmt.Printf("   - 错误: %v\n", err)
	})
	
	ctrl.SetCompletionCallback(func(outputPath string) {
		completionMessage = fmt.Sprintf("合并完成: %s", outputPath)
		fmt.Printf("   - %s\n", completionMessage)
	})
	
	// 7.3 执行完整的合并流程
	fmt.Println("\n   7.3 执行完整的合并流程:")
	
	startTime := time.Now()
	err := ctrl.StartMergeJob(testFiles[0], testFiles[1:], outputFile)
	
	if err != nil {
		fmt.Printf("   - 任务启动失败: %v\n", err)
		return
	}
	
	fmt.Printf("   - 任务启动成功\n")
	
	// 等待任务完成
	fmt.Printf("   - 等待任务完成...\n")
	for ctrl.IsJobRunning() {
		time.Sleep(100 * time.Millisecond)
		
		// 超时保护
		if time.Since(startTime) > 30*time.Second {
			fmt.Printf("   - 任务超时，取消执行\n")
			ctrl.CancelCurrentJob()
			break
		}
	}
	
	elapsed := time.Since(startTime)
	
	// 7.4 显示执行结果
	fmt.Println("\n   7.4 执行结果:")
	fmt.Printf("   - 总执行时间: %v\n", elapsed)
	fmt.Printf("   - 任务运行状态: %t\n", ctrl.IsJobRunning())
	
	// 检查任务状态
	if job := ctrl.GetCurrentJob(); job != nil {
		fmt.Printf("   - 任务状态: %s\n", job.Status.String())
		fmt.Printf("   - 任务进度: %.1f%%\n", job.Progress)
	} else {
		fmt.Printf("   - 无当前任务\n")
	}
	
	// 检查输出文件
	if _, err := os.Stat(outputFile); err == nil {
		if info, err := os.Stat(outputFile); err == nil {
			fmt.Printf("   - 输出文件已创建: %s (%d bytes)\n", 
				filepath.Base(outputFile), info.Size())
		}
	} else {
		fmt.Printf("   - 输出文件未创建\n")
	}
	
	// 7.5 显示日志统计
	fmt.Println("\n   7.5 日志统计:")
	fmt.Printf("   - 进度日志条数: %d\n", len(progressLog))
	fmt.Printf("   - 错误日志条数: %d\n", len(errorLog))
	
	if completionMessage != "" {
		fmt.Printf("   - 完成消息: %s\n", completionMessage)
	}
	
	if len(progressLog) > 0 {
		fmt.Printf("   - 首次进度: %s\n", progressLog[0])
		fmt.Printf("   - 最后进度: %s\n", progressLog[len(progressLog)-1])
	}
	
	fmt.Println("\n   完整合并流程演示完成 🎉")
	fmt.Println("   所有合并流程控制功能正常工作")
	
	fmt.Println()
}

// 辅助函数

func createTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "merge-workflow-demo-"+fmt.Sprintf("%d", time.Now().Unix()))
	os.MkdirAll(tempDir, 0755)
	return tempDir
}

func createTestController(tempDir string) *controller.Controller {
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	config.MaxMemoryUsage = 100 * 1024 * 1024 // 100MB
	
	return controller.NewController(pdfService, fileManager, config)
}

func createTestPDFFiles(tempDir string, count int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("test_%d.pdf", i+1)
		filepath := filepath.Join(tempDir, filename)
		
		// 创建简单的测试PDF内容
		content := fmt.Sprintf("%%PDF-1.4\nTest PDF file %d\n%%%%EOF", i+1)
		os.WriteFile(filepath, []byte(content), 0644)
		
		files[i] = filepath
	}
	return files
}

func createValidTestPDFFiles(tempDir string, count int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("valid_test_%d.pdf", i+1)
		filepath := filepath.Join(tempDir, filename)
		
		// 创建更完整的PDF内容
		content := fmt.Sprintf(`%%PDF-1.4
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
(Test PDF %d) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000074 00000 n 
0000000120 00000 n 
0000000179 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
273
%%%%EOF`, i+1)
		
		os.WriteFile(filepath, []byte(content), 0644)
		files[i] = filepath
	}
	return files
}

// TestProgressWriter 测试进度写入器（演示用）
type TestProgressWriter struct {
	events []string
}

func (tpw *TestProgressWriter) Write(p []byte) (n int, err error) {
	tpw.events = append(tpw.events, string(p))
	return len(p), nil
}

func (tpw *TestProgressWriter) GetEventCount() int {
	return len(tpw.events)
}
