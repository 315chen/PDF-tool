//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== 错误类型定义和处理功能演示 ===\n")

	// 1. 演示错误类型定义
	demonstrateErrorTypes()

	// 2. 演示错误创建和处理
	demonstrateErrorCreationHandling()

	// 3. 演示错误收集器
	demonstrateErrorCollector()

	// 4. 演示重试机制
	demonstrateRetryMechanism()

	// 5. 演示恢复管理器
	demonstrateRecoveryManager()

	// 6. 演示错误严重程度分析
	demonstrateErrorSeverityAnalysis()

	// 7. 演示完整的错误处理流程
	demonstrateCompleteErrorHandlingFlow()

	fmt.Println("\n=== 错误类型定义和处理演示完成 ===")
}

func demonstrateErrorTypes() {
	fmt.Println("1. 错误类型定义演示:")
	
	// 1.1 显示所有错误类型
	fmt.Println("\n   1.1 支持的错误类型:")
	errorTypes := []pdf.ErrorType{
		pdf.ErrorInvalidFile,
		pdf.ErrorEncrypted,
		pdf.ErrorCorrupted,
		pdf.ErrorPermission,
		pdf.ErrorMemory,
		pdf.ErrorIO,
		pdf.ErrorValidation,
		pdf.ErrorProcessing,
		pdf.ErrorInvalidInput,
	}
	
	for i, errorType := range errorTypes {
		// 创建示例错误
		sampleError := pdf.NewPDFError(errorType, "示例错误消息", "sample.pdf", nil)
		fmt.Printf("   %d. %s: %s\n", i+1, sampleError.Error()[:strings.Index(sampleError.Error(), ":")], sampleError.GetUserMessage())
	}
	
	// 1.2 错误严重程度分类
	fmt.Println("\n   1.2 错误严重程度分类:")
	severityGroups := map[string][]pdf.ErrorType{
		"高严重程度": {pdf.ErrorMemory, pdf.ErrorIO},
		"中等严重程度": {pdf.ErrorPermission, pdf.ErrorCorrupted},
		"低严重程度": {pdf.ErrorInvalidFile, pdf.ErrorEncrypted},
	}
	
	for severity, types := range severityGroups {
		fmt.Printf("   %s:\n", severity)
		for _, errorType := range types {
			sampleError := pdf.NewPDFError(errorType, "示例", "test.pdf", nil)
			fmt.Printf("     - %s\n", sampleError.GetUserMessage())
		}
	}
	
	// 1.3 可重试错误分类
	fmt.Println("\n   1.3 可重试错误分类:")
	retryableErrors := []pdf.ErrorType{}
	nonRetryableErrors := []pdf.ErrorType{}
	
	for _, errorType := range errorTypes {
		sampleError := pdf.NewPDFError(errorType, "测试", "test.pdf", nil)
		if sampleError.IsRetryable() {
			retryableErrors = append(retryableErrors, errorType)
		} else {
			nonRetryableErrors = append(nonRetryableErrors, errorType)
		}
	}
	
	fmt.Printf("   可重试错误 (%d个):\n", len(retryableErrors))
	for _, errorType := range retryableErrors {
		sampleError := pdf.NewPDFError(errorType, "测试", "test.pdf", nil)
		fmt.Printf("     - %s\n", sampleError.GetUserMessage())
	}
	
	fmt.Printf("   不可重试错误 (%d个):\n", len(nonRetryableErrors))
	for _, errorType := range nonRetryableErrors {
		sampleError := pdf.NewPDFError(errorType, "测试", "test.pdf", nil)
		fmt.Printf("     - %s\n", sampleError.GetUserMessage())
	}
	
	fmt.Println()
}

func demonstrateErrorCreationHandling() {
	fmt.Println("2. 错误创建和处理演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "error-demo")
	defer os.RemoveAll(tempDir)
	
	// 2.1 创建不同类型的错误
	fmt.Println("\n   2.1 创建不同类型的错误:")
	
	testErrors := []struct {
		name        string
		errorType   pdf.ErrorType
		message     string
		file        string
		causeError  error
	}{
		{
			name:      "文件不存在错误",
			errorType: pdf.ErrorInvalidFile,
			message:   "指定的PDF文件不存在",
			file:      "nonexistent.pdf",
			causeError: fmt.Errorf("file not found"),
		},
		{
			name:      "权限错误",
			errorType: pdf.ErrorPermission,
			message:   "没有读取文件的权限",
			file:      "protected.pdf",
			causeError: fmt.Errorf("permission denied"),
		},
		{
			name:      "内存不足错误",
			errorType: pdf.ErrorMemory,
			message:   "处理大文件时内存不足",
			file:      "large.pdf",
			causeError: fmt.Errorf("out of memory"),
		},
		{
			name:      "IO错误",
			errorType: pdf.ErrorIO,
			message:   "磁盘读写错误",
			file:      "corrupted.pdf",
			causeError: fmt.Errorf("disk error"),
		},
	}
	
	for i, testCase := range testErrors {
		pdfError := pdf.NewPDFError(testCase.errorType, testCase.message, testCase.file, testCase.causeError)
		
		fmt.Printf("   %d. %s:\n", i+1, testCase.name)
		fmt.Printf("      完整错误: %s\n", pdfError.Error())
		fmt.Printf("      用户消息: %s\n", pdfError.GetUserMessage())
		fmt.Printf("      详细消息: %s\n", pdfError.GetDetailedMessage())
		fmt.Printf("      严重程度: %s\n", pdfError.GetSeverity())
		fmt.Printf("      可重试: %t\n", pdfError.IsRetryable())
		
		// 测试错误链
		if pdfError.Unwrap() != nil {
			fmt.Printf("      底层错误: %v\n", pdfError.Unwrap())
		}
		fmt.Println()
	}
	
	// 2.2 错误处理器演示
	fmt.Println("   2.2 错误处理器演示:")
	errorHandler := pdf.NewDefaultErrorHandler(3)
	
	// 测试普通错误转换
	normalError := fmt.Errorf("普通的Go错误")
	handledError := errorHandler.HandleError(normalError)
	fmt.Printf("   普通错误转换: %s\n", handledError.Error())
	
	// 测试重试判断
	for _, testCase := range testErrors {
		pdfError := pdf.NewPDFError(testCase.errorType, testCase.message, testCase.file, testCase.causeError)
		shouldRetry := errorHandler.ShouldRetry(pdfError)
		userMessage := errorHandler.GetUserFriendlyMessage(pdfError)
		
		fmt.Printf("   %s - 应该重试: %t, 用户消息: %s\n", 
			testCase.name, shouldRetry, userMessage)
	}
	
	fmt.Println()
}

func demonstrateErrorCollector() {
	fmt.Println("3. 错误收集器演示:")
	
	// 3.1 创建错误收集器
	fmt.Println("\n   3.1 创建和使用错误收集器:")
	errorCollector := pdf.NewErrorCollector()
	
	// 添加多个错误
	errors := []error{
		pdf.NewPDFError(pdf.ErrorInvalidFile, "文件1格式错误", "file1.pdf", nil),
		pdf.NewPDFError(pdf.ErrorEncrypted, "文件2需要密码", "file2.pdf", nil),
		pdf.NewPDFError(pdf.ErrorIO, "文件3读取失败", "file3.pdf", fmt.Errorf("disk error")),
		nil, // 测试nil错误
		pdf.NewPDFError(pdf.ErrorMemory, "内存不足", "file4.pdf", nil),
	}
	
	fmt.Printf("   添加 %d 个错误到收集器:\n", len(errors))
	for i, err := range errors {
		errorCollector.Add(err)
		if err != nil {
			fmt.Printf("   %d. %s\n", i+1, err.Error())
		} else {
			fmt.Printf("   %d. (nil错误，已忽略)\n", i+1)
		}
	}
	
	// 3.2 错误统计
	fmt.Println("\n   3.2 错误统计:")
	fmt.Printf("   错误数量: %d\n", errorCollector.GetErrorCount())
	fmt.Printf("   是否有错误: %t\n", errorCollector.HasErrors())
	
	// 3.3 获取错误摘要
	fmt.Println("\n   3.3 错误摘要:")
	summary := errorCollector.GetSummary()
	fmt.Printf("   %s\n", summary)
	
	// 3.4 获取所有错误
	fmt.Println("\n   3.4 所有收集的错误:")
	allErrors := errorCollector.GetErrors()
	for i, err := range allErrors {
		if pdfErr, ok := err.(*pdf.PDFError); ok {
			fmt.Printf("   %d. [%s] %s\n", i+1, pdfErr.GetSeverity(), pdfErr.GetDetailedMessage())
		} else {
			fmt.Printf("   %d. %s\n", i+1, err.Error())
		}
	}
	
	// 3.5 清空收集器
	fmt.Println("\n   3.5 清空错误收集器:")
	errorCollector.Clear()
	fmt.Printf("   清空后错误数量: %d\n", errorCollector.GetErrorCount())
	
	fmt.Println()
}

func demonstrateRetryMechanism() {
	fmt.Println("4. 重试机制演示:")
	
	// 4.1 创建重试管理器
	fmt.Println("\n   4.1 创建重试管理器:")
	retryConfig := pdf.DefaultRetryConfig()
	errorHandler := pdf.NewDefaultErrorHandler(3)
	retryManager := pdf.NewRetryManager(retryConfig, errorHandler)
	
	fmt.Printf("   重试配置:\n")
	fmt.Printf("   - 最大重试次数: %d\n", retryConfig.MaxRetries)
	fmt.Printf("   - 初始延迟: %v\n", retryConfig.InitialDelay)
	fmt.Printf("   - 最大延迟: %v\n", retryConfig.MaxDelay)
	fmt.Printf("   - 退避因子: %.1f\n", retryConfig.BackoffFactor)
	
	// 4.2 模拟重试场景
	fmt.Println("\n   4.2 重试场景演示:")
	
	scenarios := []struct {
		name      string
		operation func() error
		expectRetry bool
	}{
		{
			name: "成功操作",
			operation: func() error {
				return nil
			},
			expectRetry: false,
		},
		{
			name: "IO错误(可重试)",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", "test.pdf", nil)
			},
			expectRetry: true,
		},
		{
			name: "权限错误(不可重试)",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorPermission, "权限被拒绝", "test.pdf", nil)
			},
			expectRetry: false,
		},
	}
	
	for i, scenario := range scenarios {
		fmt.Printf("   场景 %d: %s\n", i+1, scenario.name)
		
		startTime := time.Now()
		err := retryManager.Execute(scenario.operation)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("   - 结果: 失败 - %s\n", err.Error())
		} else {
			fmt.Printf("   - 结果: 成功\n")
		}
		fmt.Printf("   - 用时: %v\n", duration)
		fmt.Printf("   - 预期重试: %t\n", scenario.expectRetry)
		fmt.Println()
	}
	
	fmt.Println()
}

func demonstrateRecoveryManager() {
	fmt.Println("5. 恢复管理器演示:")
	
	// 5.1 创建恢复管理器
	fmt.Println("\n   5.1 创建恢复管理器:")
	recoveryManager := pdf.NewRecoveryManager(100) // 100MB内存限制
	
	fmt.Printf("   恢复管理器已创建，内存限制: 100MB\n")
	
	// 5.2 模拟恢复场景
	fmt.Println("\n   5.2 恢复场景演示:")
	
	recoveryScenarios := []struct {
		name      string
		operation func() error
	}{
		{
			name: "正常操作",
			operation: func() error {
				return nil
			},
		},
		{
			name: "IO错误恢复",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", "test.pdf", nil)
			},
		},
		{
			name: "内存错误恢复",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorMemory, "内存不足", "large.pdf", nil)
			},
		},
	}
	
	for i, scenario := range recoveryScenarios {
		fmt.Printf("   场景 %d: %s\n", i+1, scenario.name)
		
		startTime := time.Now()
		err := recoveryManager.ExecuteWithRecovery(scenario.operation)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("   - 结果: 失败 - %s\n", err.Error())
		} else {
			fmt.Printf("   - 结果: 成功\n")
		}
		fmt.Printf("   - 用时: %v\n", duration)
		
		// 显示恢复统计
		stats := recoveryManager.GetRecoveryStats()
		fmt.Printf("   - 内存使用: %v MB\n", stats["alloc_mb"])
		fmt.Printf("   - 错误数量: %v\n", stats["error_count"])
		fmt.Println()
	}
	
	// 5.3 错误摘要
	fmt.Println("   5.3 恢复管理器错误摘要:")
	if recoveryManager.GetErrorSummary() != "" {
		fmt.Printf("   %s\n", recoveryManager.GetErrorSummary())
	} else {
		fmt.Printf("   无错误记录\n")
	}
	
	// 清空错误
	recoveryManager.ClearErrors()
	fmt.Printf("   错误记录已清空\n")
	
	fmt.Println()
}

func demonstrateErrorSeverityAnalysis() {
	fmt.Println("6. 错误严重程度分析演示:")
	
	// 6.1 创建不同严重程度的错误
	fmt.Println("\n   6.1 错误严重程度分析:")
	
	testErrors := []struct {
		errorType pdf.ErrorType
		message   string
		file      string
	}{
		{pdf.ErrorMemory, "内存分配失败", "large.pdf"},
		{pdf.ErrorIO, "磁盘写入失败", "output.pdf"},
		{pdf.ErrorPermission, "文件访问被拒绝", "protected.pdf"},
		{pdf.ErrorCorrupted, "文件结构损坏", "broken.pdf"},
		{pdf.ErrorInvalidFile, "文件格式不正确", "invalid.pdf"},
		{pdf.ErrorEncrypted, "文件已加密", "secure.pdf"},
		{pdf.ErrorValidation, "PDF验证失败", "test.pdf"},
		{pdf.ErrorProcessing, "处理过程失败", "complex.pdf"},
	}
	
	// 按严重程度分组
	severityGroups := make(map[string][]string)
	
	for _, testCase := range testErrors {
		pdfError := pdf.NewPDFError(testCase.errorType, testCase.message, testCase.file, nil)
		severity := pdfError.GetSeverity()
		
		if severityGroups[severity] == nil {
			severityGroups[severity] = make([]string, 0)
		}
		
		errorInfo := fmt.Sprintf("%s (%s)", pdfError.GetUserMessage(), testCase.file)
		severityGroups[severity] = append(severityGroups[severity], errorInfo)
	}
	
	// 按严重程度显示
	severityOrder := []string{"high", "medium", "low", "unknown"}
	severityNames := map[string]string{
		"high":    "高严重程度",
		"medium":  "中等严重程度", 
		"low":     "低严重程度",
		"unknown": "未知严重程度",
	}
	
	for _, severity := range severityOrder {
		if errors, exists := severityGroups[severity]; exists && len(errors) > 0 {
			fmt.Printf("   %s (%d个):\n", severityNames[severity], len(errors))
			for i, errorInfo := range errors {
				fmt.Printf("     %d. %s\n", i+1, errorInfo)
			}
			fmt.Println()
		}
	}
	
	// 6.2 错误处理建议
	fmt.Println("   6.2 错误处理建议:")
	suggestions := map[string][]string{
		"high": {
			"立即停止当前操作",
			"释放系统资源",
			"通知用户并提供解决方案",
			"记录详细错误日志",
		},
		"medium": {
			"尝试替代方案",
			"提示用户检查文件状态",
			"记录警告日志",
		},
		"low": {
			"提供用户友好的错误提示",
			"建议用户检查文件格式",
			"记录信息日志",
		},
	}
	
	for severity, suggestionList := range suggestions {
		fmt.Printf("   %s错误处理建议:\n", severityNames[severity])
		for i, suggestion := range suggestionList {
			fmt.Printf("     %d. %s\n", i+1, suggestion)
		}
		fmt.Println()
	}
	
	fmt.Println()
}

func demonstrateCompleteErrorHandlingFlow() {
	fmt.Println("7. 完整错误处理流程演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "complete-error-demo")
	defer os.RemoveAll(tempDir)
	
	// 7.1 初始化错误处理组件
	fmt.Println("\n   7.1 初始化错误处理组件:")
	
	errorHandler := pdf.NewDefaultErrorHandler(3)
	errorCollector := pdf.NewErrorCollector()
	recoveryManager := pdf.NewRecoveryManager(100)
	
	fmt.Printf("   - 错误处理器初始化完成\n")
	fmt.Printf("   - 错误收集器初始化完成\n")
	fmt.Printf("   - 恢复管理器初始化完成\n")
	
	// 7.2 模拟复杂的错误处理场景
	fmt.Println("\n   7.2 执行复杂错误处理流程:")
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	os.WriteFile(testFile, []byte("test content"), 0644)
	
	// 模拟PDF处理操作
	fmt.Printf("   处理文件: %s\n", filepath.Base(testFile))
	
	// 步骤1: 文件验证
	fmt.Printf("   步骤1: 文件验证\n")
	validationErr := pdf.NewPDFError(pdf.ErrorInvalidFile, "文件格式验证失败", testFile, fmt.Errorf("invalid header"))
	errorCollector.Add(validationErr)
	fmt.Printf("   - 验证失败: %s\n", validationErr.GetDetailedMessage())
	
	// 步骤2: 错误处理和重试
	fmt.Printf("   步骤2: 错误处理和重试\n")
	handledErr := errorHandler.HandleError(validationErr)
	shouldRetry := errorHandler.ShouldRetry(handledErr)
	fmt.Printf("   - 错误处理结果: %s\n", handledErr.Error())
	fmt.Printf("   - 是否可重试: %t\n", shouldRetry)
	
	// 步骤3: 恢复尝试
	fmt.Printf("   步骤3: 恢复尝试\n")
	recoveryErr := recoveryManager.ExecuteWithRecovery(func() error {
		// 模拟恢复后的成功操作
		return nil
	})
	
	if recoveryErr != nil {
		fmt.Printf("   - 恢复失败: %s\n", recoveryErr.Error())
		errorCollector.Add(recoveryErr)
	} else {
		fmt.Printf("   - 恢复成功 ✓\n")
	}
	
	// 步骤4: 最终错误统计
	fmt.Printf("   步骤4: 最终错误统计\n")
	fmt.Printf("   - 收集的错误数量: %d\n", errorCollector.GetErrorCount())
	fmt.Printf("   - 错误摘要: %s\n", errorCollector.GetSummary())
	
	// 步骤5: 生成错误报告
	fmt.Printf("   步骤5: 生成错误报告\n")
	if errorCollector.HasErrors() {
		fmt.Printf("   错误详情:\n")
		for i, err := range errorCollector.GetErrors() {
			if pdfErr, ok := err.(*pdf.PDFError); ok {
				fmt.Printf("     %d. [%s] %s\n", i+1, pdfErr.GetSeverity(), pdfErr.GetDetailedMessage())
			} else {
				fmt.Printf("     %d. %s\n", i+1, err.Error())
			}
		}
	}
	
	// 步骤6: 清理和总结
	fmt.Printf("   步骤6: 清理和总结\n")
	recoveryStats := recoveryManager.GetRecoveryStats()
	fmt.Printf("   - 内存使用: %v MB\n", recoveryStats["alloc_mb"])
	fmt.Printf("   - GC次数: %v\n", recoveryStats["num_gc"])
	
	errorCollector.Clear()
	recoveryManager.ClearErrors()
	fmt.Printf("   - 错误记录已清理 ✓\n")
	
	fmt.Println("\n   完整错误处理流程演示完成 🎉")
	fmt.Println("   所有错误处理组件协同工作正常")
	
	fmt.Println()
}


