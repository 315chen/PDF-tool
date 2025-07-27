//go:build ignore
// +build ignore
package main

import (
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
	fmt.Println("=== 主控制器和事件处理功能演示 ===\n")

	// 1. 演示控制器创建和初始化
	demonstrateControllerCreation()

	// 2. 演示事件处理器创建
	demonstrateEventHandlerCreation()

	// 3. 演示文件验证事件处理
	demonstrateFileValidationEvents()

	// 4. 演示合并任务事件处理
	demonstrateMergeJobEvents()

	// 5. 演示进度和状态回调
	demonstrateProgressAndStatusCallbacks()

	// 6. 演示错误处理和恢复
	demonstrateErrorHandlingAndRecovery()

	// 7. 演示完整的控制器事件流程
	demonstrateCompleteControllerEventFlow()

	fmt.Println("\n=== 主控制器和事件处理演示完成 ===")
}

func demonstrateControllerCreation() {
	fmt.Println("1. 控制器创建和初始化演示:")
	
	// 1.1 创建临时目录
	fmt.Println("\n   1.1 创建临时目录:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	fmt.Printf("   - 临时目录: %s\n", tempDir)
	
	// 1.2 创建服务组件
	fmt.Println("\n   1.2 创建服务组件:")
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	
	fmt.Printf("   - 文件管理器创建成功\n")
	fmt.Printf("   - PDF服务创建成功\n")
	
	// 1.3 创建配置
	fmt.Println("\n   1.3 创建配置:")
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	config.MaxMemoryUsage = 100 * 1024 * 1024 // 100MB
	
	fmt.Printf("   - 配置创建成功\n")
	fmt.Printf("   - 临时目录: %s\n", config.TempDirectory)
	fmt.Printf("   - 最大内存: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	
	// 1.4 创建控制器
	fmt.Println("\n   1.4 创建控制器:")
	ctrl := controller.NewController(pdfService, fileManager, config)
	
	fmt.Printf("   - 控制器创建成功\n")
	fmt.Printf("   - 初始任务状态: %t\n", ctrl.IsJobRunning())
	fmt.Printf("   - 当前任务: %v\n", ctrl.GetCurrentJob())
	
	// 1.5 设置回调函数
	fmt.Println("\n   1.5 设置回调函数:")
	
	progressCallbackCalled := false
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		progressCallbackCalled = true
		fmt.Printf("   - 进度回调: %.1f%% - %s\n", progress*100, status)
	})
	
	errorCallbackCalled := false
	ctrl.SetErrorCallback(func(err error) {
		errorCallbackCalled = true
		fmt.Printf("   - 错误回调: %v\n", err)
	})
	
	completionCallbackCalled := false
	ctrl.SetCompletionCallback(func(outputPath string) {
		completionCallbackCalled = true
		fmt.Printf("   - 完成回调: %s\n", outputPath)
	})
	
	fmt.Printf("   - 回调函数设置完成\n")
	fmt.Printf("   - 进度回调已设置: %t\n", progressCallbackCalled)
	fmt.Printf("   - 错误回调已设置: %t\n", errorCallbackCalled)
	fmt.Printf("   - 完成回调已设置: %t\n", completionCallbackCalled)
	
	fmt.Println()
}

func demonstrateEventHandlerCreation() {
	fmt.Println("2. 事件处理器创建演示:")
	
	// 2.1 创建控制器
	fmt.Println("\n   2.1 创建控制器:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	fmt.Printf("   - 控制器创建成功\n")
	
	// 2.2 创建事件处理器
	fmt.Println("\n   2.2 创建事件处理器:")
	eventHandler := controller.NewEventHandler(ctrl)
	
	fmt.Printf("   - 事件处理器创建成功\n")
	fmt.Printf("   - 任务运行状态: %t\n", eventHandler.IsJobRunning())
	
	// 2.3 设置UI回调
	fmt.Println("\n   2.3 设置UI回调:")
	
	uiStateChanged := false
	eventHandler.SetUIStateCallback(func(enabled bool) {
		uiStateChanged = true
		fmt.Printf("   - UI状态变更: %t\n", enabled)
	})
	
	progressUpdated := false
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		progressUpdated = true
		fmt.Printf("   - 进度更新: %.1f%% - %s - %s\n", progress*100, status, detail)
	})
	
	errorOccurred := false
	eventHandler.SetErrorCallback(func(err error) {
		errorOccurred = true
		fmt.Printf("   - 错误发生: %v\n", err)
	})
	
	completionOccurred := false
	eventHandler.SetCompletionCallback(func(message string) {
		completionOccurred = true
		fmt.Printf("   - 完成通知: %s\n", message)
	})
	
	fmt.Printf("   - UI回调设置完成\n")
	fmt.Printf("   - UI状态回调: %t\n", uiStateChanged)
	fmt.Printf("   - 进度回调: %t\n", progressUpdated)
	fmt.Printf("   - 错误回调: %t\n", errorOccurred)
	fmt.Printf("   - 完成回调: %t\n", completionOccurred)
	
	fmt.Println()
}

func demonstrateFileValidationEvents() {
	fmt.Println("3. 文件验证事件处理演示:")
	
	// 3.1 创建测试环境
	fmt.Println("\n   3.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 创建测试文件
	testFiles := createTestPDFFiles(tempDir, 3)
	fmt.Printf("   - 创建了 %d 个测试文件\n", len(testFiles))
	
	// 3.2 演示主文件选择事件
	fmt.Println("\n   3.2 演示主文件选择事件:")
	
	mainFile := testFiles[0]
	err := eventHandler.HandleMainFileSelected(mainFile)
	if err != nil {
		fmt.Printf("   - 主文件选择失败: %v\n", err)
	} else {
		fmt.Printf("   - 主文件选择成功: %s\n", filepath.Base(mainFile))
	}
	
	// 3.3 演示附加文件添加事件
	fmt.Println("\n   3.3 演示附加文件添加事件:")
	
	for i, additionalFile := range testFiles[1:] {
		fileEntry, err := eventHandler.HandleAdditionalFileAdded(additionalFile)
		if err != nil {
			fmt.Printf("   - 附加文件 %d 添加失败: %v\n", i+1, err)
		} else {
			fmt.Printf("   - 附加文件 %d 添加成功: %s\n", i+1, fileEntry.DisplayName)
			fmt.Printf("     大小: %s, 页数: %d, 有效: %t\n", 
				fileEntry.GetSizeString(), fileEntry.PageCount, fileEntry.IsValid)
		}
	}
	
	// 3.4 演示文件验证事件
	fmt.Println("\n   3.4 演示文件验证事件:")
	
	for i, testFile := range testFiles {
		fileEntry, err := eventHandler.HandleFileValidation(testFile)
		if err != nil {
			fmt.Printf("   - 文件 %d 验证失败: %v\n", i+1, err)
		} else {
			fmt.Printf("   - 文件 %d 验证成功: %s\n", i+1, fileEntry.DisplayName)
			fmt.Printf("     路径: %s\n", fileEntry.Path)
			fmt.Printf("     大小: %s\n", fileEntry.GetSizeString())
			fmt.Printf("     页数: %d\n", fileEntry.PageCount)
			fmt.Printf("     加密: %t\n", fileEntry.IsEncrypted)
			fmt.Printf("     有效: %t\n", fileEntry.IsValid)
		}
	}
	
	// 3.5 演示批量文件验证
	fmt.Println("\n   3.5 演示批量文件验证:")
	
	validationResults := eventHandler.ValidateAllFiles(testFiles[0], testFiles[1:])
	fmt.Printf("   - 验证结果数量: %d\n", len(validationResults))
	
	for filePath, err := range validationResults {
		if err != nil {
			fmt.Printf("   - %s: 验证失败 - %v\n", filepath.Base(filePath), err)
		} else {
			fmt.Printf("   - %s: 验证成功\n", filepath.Base(filePath))
		}
	}
	
	fmt.Println()
}

func demonstrateMergeJobEvents() {
	fmt.Println("4. 合并任务事件处理演示:")
	
	// 4.1 创建测试环境
	fmt.Println("\n   4.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 创建测试文件
	testFiles := createTestPDFFiles(tempDir, 3)
	outputFile := filepath.Join(tempDir, "merged_output.pdf")
	
	fmt.Printf("   - 主文件: %s\n", filepath.Base(testFiles[0]))
	fmt.Printf("   - 附加文件数: %d\n", len(testFiles)-1)
	fmt.Printf("   - 输出文件: %s\n", filepath.Base(outputFile))
	
	// 4.2 演示输出路径变更事件
	fmt.Println("\n   4.2 演示输出路径变更事件:")
	
	err := eventHandler.HandleOutputPathChanged(outputFile)
	if err != nil {
		fmt.Printf("   - 输出路径验证失败: %v\n", err)
	} else {
		fmt.Printf("   - 输出路径验证成功: %s\n", filepath.Base(outputFile))
	}
	
	// 4.3 演示合并开始事件
	fmt.Println("\n   4.3 演示合并开始事件:")
	
	// 设置进度回调
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		fmt.Printf("   - 进度: %.1f%% - %s\n", progress*100, status)
	})
	
	err = eventHandler.HandleMergeStart(testFiles[0], testFiles[1:], outputFile)
	if err != nil {
		fmt.Printf("   - 合并开始失败: %v\n", err)
	} else {
		fmt.Printf("   - 合并开始成功\n")
		fmt.Printf("   - 任务运行状态: %t\n", eventHandler.IsJobRunning())
		
		// 等待任务完成
		fmt.Printf("   - 等待任务完成...\n")
		for eventHandler.IsJobRunning() {
			time.Sleep(100 * time.Millisecond)
			
			// 显示任务状态
			if job := eventHandler.GetJobStatus(); job != nil {
				fmt.Printf("   - 任务状态: %s, 进度: %.1f%%\n", 
					job.Status.String(), job.Progress)
			}
		}
		
		fmt.Printf("   - 任务完成\n")
	}
	
	// 4.4 演示任务状态查询
	fmt.Println("\n   4.4 演示任务状态查询:")
	
	job := eventHandler.GetJobStatus()
	if job != nil {
		fmt.Printf("   - 任务ID: %s\n", job.ID)
		fmt.Printf("   - 状态: %s\n", job.Status.String())
		fmt.Printf("   - 进度: %.1f%%\n", job.Progress)
		fmt.Printf("   - 主文件: %s\n", filepath.Base(job.MainFile))
		fmt.Printf("   - 附加文件数: %d\n", len(job.AdditionalFiles))
		fmt.Printf("   - 输出文件: %s\n", filepath.Base(job.OutputPath))
		fmt.Printf("   - 开始时间: %v\n", job.CreatedAt.Format("15:04:05"))
		if job.CompletedAt != nil {
			fmt.Printf("   - 完成时间: %v\n", job.CompletedAt.Format("15:04:05"))
			fmt.Printf("   - 用时: %v\n", job.CompletedAt.Sub(job.CreatedAt))
		}
	} else {
		fmt.Printf("   - 无当前任务\n")
	}
	
	fmt.Println()
}

func demonstrateProgressAndStatusCallbacks() {
	fmt.Println("5. 进度和状态回调演示:")
	
	// 5.1 创建测试环境
	fmt.Println("\n   5.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 5.2 设置详细的回调函数
	fmt.Println("\n   5.2 设置详细的回调函数:")
	
	var progressUpdates []string
	var statusUpdates []string
	var errorMessages []string
	var completionMessages []string
	
	eventHandler.SetUIStateCallback(func(enabled bool) {
		if enabled {
			statusUpdates = append(statusUpdates, "UI已启用")
		} else {
			statusUpdates = append(statusUpdates, "UI已禁用")
		}
	})
	
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		progressUpdate := fmt.Sprintf("%.1f%% - %s - %s", progress*100, status, detail)
		progressUpdates = append(progressUpdates, progressUpdate)
	})
	
	eventHandler.SetErrorCallback(func(err error) {
		errorMessages = append(errorMessages, err.Error())
	})
	
	eventHandler.SetCompletionCallback(func(message string) {
		completionMessages = append(completionMessages, message)
	})
	
	fmt.Printf("   - 回调函数设置完成\n")
	
	// 5.3 模拟合并操作
	fmt.Println("\n   5.3 模拟合并操作:")
	
	testFiles := createTestPDFFiles(tempDir, 2)
	outputFile := filepath.Join(tempDir, "callback_test.pdf")
	
	err := eventHandler.HandleMergeStart(testFiles[0], testFiles[1:], outputFile)
	if err != nil {
		fmt.Printf("   - 合并启动失败: %v\n", err)
	} else {
		// 等待完成
		for eventHandler.IsJobRunning() {
			time.Sleep(50 * time.Millisecond)
		}
	}
	
	// 5.4 显示回调结果
	fmt.Println("\n   5.4 回调结果统计:")
	
	fmt.Printf("   - 状态更新数量: %d\n", len(statusUpdates))
	for i, update := range statusUpdates {
		fmt.Printf("     %d. %s\n", i+1, update)
	}
	
	fmt.Printf("   - 进度更新数量: %d\n", len(progressUpdates))
	for i, update := range progressUpdates {
		if i < 5 { // 只显示前5个
			fmt.Printf("     %d. %s\n", i+1, update)
		}
	}
	if len(progressUpdates) > 5 {
		fmt.Printf("     ... (还有 %d 个更新)\n", len(progressUpdates)-5)
	}
	
	fmt.Printf("   - 错误消息数量: %d\n", len(errorMessages))
	for i, message := range errorMessages {
		fmt.Printf("     %d. %s\n", i+1, message)
	}
	
	fmt.Printf("   - 完成消息数量: %d\n", len(completionMessages))
	for i, message := range completionMessages {
		fmt.Printf("     %d. %s\n", i+1, message)
	}
	
	fmt.Println()
}

func demonstrateErrorHandlingAndRecovery() {
	fmt.Println("6. 错误处理和恢复演示:")
	
	// 6.1 创建测试环境
	fmt.Println("\n   6.1 创建测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 6.2 演示文件验证错误
	fmt.Println("\n   6.2 演示文件验证错误:")
	
	// 尝试验证不存在的文件
	nonExistentFile := filepath.Join(tempDir, "non_existent.pdf")
	err := eventHandler.HandleMainFileSelected(nonExistentFile)
	if err != nil {
		fmt.Printf("   - 预期错误: %v\n", err)
	}
	
	// 6.3 演示输出路径错误
	fmt.Println("\n   6.3 演示输出路径错误:")
	
	// 尝试设置只读目录作为输出路径
	readOnlyDir := "/System" // macOS系统目录
	readOnlyFile := filepath.Join(readOnlyDir, "test.pdf")
	err = eventHandler.HandleOutputPathChanged(readOnlyFile)
	if err != nil {
		fmt.Printf("   - 预期错误: %v\n", err)
	}
	
	// 6.4 演示合并参数错误
	fmt.Println("\n   6.4 演示合并参数错误:")
	
	// 尝试没有主文件的合并
	err = eventHandler.HandleMergeStart("", []string{}, "")
	if err != nil {
		fmt.Printf("   - 预期错误: %v\n", err)
	}
	
	// 尝试没有附加文件的合并
	testFiles := createTestPDFFiles(tempDir, 1)
	err = eventHandler.HandleMergeStart(testFiles[0], []string{}, "")
	if err != nil {
		fmt.Printf("   - 预期错误: %v\n", err)
	}
	
	// 6.5 演示任务取消
	fmt.Println("\n   6.5 演示任务取消:")
	
	// 设置错误回调
	var errorOccurred bool
	eventHandler.SetErrorCallback(func(err error) {
		errorOccurred = true
		fmt.Printf("   - 错误回调: %v\n", err)
	})
	
	// 启动一个任务然后立即取消
	testFiles = createTestPDFFiles(tempDir, 2)
	outputFile := filepath.Join(tempDir, "cancel_test.pdf")
	
	err = eventHandler.HandleMergeStart(testFiles[0], testFiles[1:], outputFile)
	if err != nil {
		fmt.Printf("   - 任务启动失败: %v\n", err)
	} else {
		fmt.Printf("   - 任务启动成功\n")
		
		// 等待一小段时间然后取消
		time.Sleep(100 * time.Millisecond)
		
		err = eventHandler.HandleMergeCancel()
		if err != nil {
			fmt.Printf("   - 取消失败: %v\n", err)
		} else {
			fmt.Printf("   - 取消成功\n")
		}
		
		// 等待取消完成
		time.Sleep(200 * time.Millisecond)
		
		fmt.Printf("   - 任务运行状态: %t\n", eventHandler.IsJobRunning())
		fmt.Printf("   - 错误是否发生: %t\n", errorOccurred)
	}
	
	fmt.Println()
}

func demonstrateCompleteControllerEventFlow() {
	fmt.Println("7. 完整控制器事件流程演示:")
	
	// 7.1 创建完整测试环境
	fmt.Println("\n   7.1 创建完整测试环境:")
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	
	ctrl := createTestController(tempDir)
	eventHandler := controller.NewEventHandler(ctrl)
	
	// 创建测试文件
	testFiles := createTestPDFFiles(tempDir, 4)
	outputFile := filepath.Join(tempDir, "complete_flow_test.pdf")
	
	fmt.Printf("   - 测试文件数: %d\n", len(testFiles))
	fmt.Printf("   - 输出文件: %s\n", filepath.Base(outputFile))
	
	// 7.2 设置完整的事件监听
	fmt.Println("\n   7.2 设置完整的事件监听:")
	
	var eventLog []string
	
	eventHandler.SetUIStateCallback(func(enabled bool) {
		if enabled {
			eventLog = append(eventLog, "UI状态: 已启用")
		} else {
			eventLog = append(eventLog, "UI状态: 已禁用")
		}
	})
	
	eventHandler.SetProgressUpdateCallback(func(progress float64, status, detail string) {
		eventLog = append(eventLog, fmt.Sprintf("进度: %.1f%% - %s", progress*100, status))
	})
	
	eventHandler.SetErrorCallback(func(err error) {
		eventLog = append(eventLog, fmt.Sprintf("错误: %v", err))
	})
	
	eventHandler.SetCompletionCallback(func(message string) {
		eventLog = append(eventLog, fmt.Sprintf("完成: %s", message))
	})
	
	// 7.3 执行完整流程
	fmt.Println("\n   7.3 执行完整流程:")
	
	// 步骤1: 选择主文件
	fmt.Printf("   步骤1: 选择主文件\n")
	err := eventHandler.HandleMainFileSelected(testFiles[0])
	if err != nil {
		fmt.Printf("   - 主文件选择失败: %v\n", err)
		return
	}
	eventLog = append(eventLog, "主文件选择: 成功")
	
	// 步骤2: 添加附加文件
	fmt.Printf("   步骤2: 添加附加文件\n")
	for i, additionalFile := range testFiles[1:] {
		_, err := eventHandler.HandleAdditionalFileAdded(additionalFile)
		if err != nil {
			fmt.Printf("   - 附加文件 %d 添加失败: %v\n", i+1, err)
		} else {
			eventLog = append(eventLog, fmt.Sprintf("附加文件 %d: 添加成功", i+1))
		}
	}
	
	// 步骤3: 设置输出路径
	fmt.Printf("   步骤3: 设置输出路径\n")
	err = eventHandler.HandleOutputPathChanged(outputFile)
	if err != nil {
		fmt.Printf("   - 输出路径设置失败: %v\n", err)
		return
	}
	eventLog = append(eventLog, "输出路径: 设置成功")
	
	// 步骤4: 开始合并
	fmt.Printf("   步骤4: 开始合并\n")
	err = eventHandler.HandleMergeStart(testFiles[0], testFiles[1:], outputFile)
	if err != nil {
		fmt.Printf("   - 合并开始失败: %v\n", err)
		return
	}
	eventLog = append(eventLog, "合并任务: 开始")
	
	// 步骤5: 等待完成
	fmt.Printf("   步骤5: 等待完成\n")
	startTime := time.Now()
	for eventHandler.IsJobRunning() {
		time.Sleep(50 * time.Millisecond)
		
		// 超时保护
		if time.Since(startTime) > 30*time.Second {
			fmt.Printf("   - 任务超时，强制取消\n")
			eventHandler.HandleMergeCancel()
			break
		}
	}
	
	// 7.4 显示事件日志
	fmt.Println("\n   7.4 事件日志:")
	fmt.Printf("   - 总事件数: %d\n", len(eventLog))
	
	for i, event := range eventLog {
		fmt.Printf("   %d. %s\n", i+1, event)
	}
	
	// 7.5 显示最终状态
	fmt.Println("\n   7.5 最终状态:")
	
	job := eventHandler.GetJobStatus()
	if job != nil {
		fmt.Printf("   - 任务状态: %s\n", job.Status.String())
		fmt.Printf("   - 最终进度: %.1f%%\n", job.Progress)
		if job.CompletedAt != nil {
			fmt.Printf("   - 总用时: %v\n", job.CompletedAt.Sub(job.CreatedAt))
		}
	}
	
	fmt.Printf("   - 任务运行状态: %t\n", eventHandler.IsJobRunning())
	
	// 检查输出文件
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("   - 输出文件已创建: %s\n", filepath.Base(outputFile))
	} else {
		fmt.Printf("   - 输出文件未创建\n")
	}
	
	fmt.Println("\n   完整控制器事件流程演示完成 🎉")
	fmt.Println("   所有控制器和事件处理功能正常工作")
	
	fmt.Println()
}

// 辅助函数

func createTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "controller-demo-"+fmt.Sprintf("%d", time.Now().Unix()))
	os.MkdirAll(tempDir, 0755)
	return tempDir
}

func createTestController(tempDir string) *controller.Controller {
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir
	
	return controller.NewController(pdfService, fileManager, config)
}

func createTestPDFFiles(tempDir string, count int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("test_%d.pdf", i+1)
		filepath := filepath.Join(tempDir, filename)
		
		// 创建简单的测试PDF内容
		content := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000074 00000 n \n0000000120 00000 n \ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n179\n%%%%EOF\n")
		os.WriteFile(filepath, []byte(content), 0644)
		
		files[i] = filepath
	}
	return files
}
