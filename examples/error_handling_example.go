//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"log"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	// 创建错误处理器
	errorHandler := pdf.NewDefaultErrorHandler(3)
	
	// 创建错误收集器
	errorCollector := pdf.NewErrorCollector()
	
	// 模拟一些错误情况
	demonstrateErrorHandling(errorHandler, errorCollector)
}

func demonstrateErrorHandling(handler *pdf.DefaultErrorHandler, collector *pdf.ErrorCollector) {
	fmt.Println("=== PDF错误处理系统演示 ===\n")
	
	// 1. 创建不同类型的PDF错误
	errors := []error{
		pdf.NewPDFError(pdf.ErrorInvalidFile, "文件头部损坏", "document1.pdf", nil),
		pdf.NewPDFError(pdf.ErrorEncrypted, "需要密码解密", "secure.pdf", nil),
		pdf.NewPDFError(pdf.ErrorIO, "磁盘空间不足", "output.pdf", fmt.Errorf("disk full")),
		pdf.NewPDFError(pdf.ErrorMemory, "内存分配失败", "large.pdf", fmt.Errorf("out of memory")),
	}
	
	fmt.Println("1. 错误类型和消息演示:")
	for i, err := range errors {
		pdfErr := err.(*pdf.PDFError)
		fmt.Printf("错误 %d:\n", i+1)
		fmt.Printf("  类型: %s\n", pdfErr.Error())
		fmt.Printf("  用户消息: %s\n", pdfErr.GetUserMessage())
		fmt.Printf("  详细消息: %s\n", pdfErr.GetDetailedMessage())
		fmt.Printf("  严重程度: %s\n", pdfErr.GetSeverity())
		fmt.Printf("  可重试: %t\n", pdfErr.IsRetryable())
		fmt.Println()
	}
	
	// 2. 错误处理器演示
	fmt.Println("2. 错误处理器演示:")
	for i, err := range errors {
		fmt.Printf("处理错误 %d:\n", i+1)
		handledErr := handler.HandleError(err)
		fmt.Printf("  处理后: %s\n", handledErr.Error())
		fmt.Printf("  应该重试: %t\n", handler.ShouldRetry(handledErr))
		fmt.Printf("  用户友好消息: %s\n", handler.GetUserFriendlyMessage(handledErr))
		fmt.Println()
	}
	
	// 3. 错误收集器演示
	fmt.Println("3. 错误收集器演示:")
	
	// 添加错误到收集器
	for _, err := range errors {
		collector.Add(err)
	}
	
	fmt.Printf("收集到的错误数量: %d\n", collector.GetErrorCount())
	fmt.Printf("是否有错误: %t\n", collector.HasErrors())
	fmt.Println("\n错误摘要:")
	fmt.Println(collector.GetSummary())
	
	// 4. 批量处理错误演示
	fmt.Println("4. 批量处理错误演示:")
	processBatchFiles(handler, collector)
}

func processBatchFiles(handler *pdf.DefaultErrorHandler, collector *pdf.ErrorCollector) {
	// 清空收集器
	collector.Clear()
	
	// 模拟批量处理文件
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf", "file4.pdf"}
	
	for _, file := range files {
		err := processFile(file)
		if err != nil {
			// 使用错误处理器处理错误
			handledErr := handler.HandleError(err)
			collector.Add(handledErr)
			
			// 检查是否应该重试
			if handler.ShouldRetry(handledErr) {
				fmt.Printf("文件 %s 处理失败，将重试: %s\n", file, handler.GetUserFriendlyMessage(handledErr))
			} else {
				fmt.Printf("文件 %s 处理失败，跳过: %s\n", file, handler.GetUserFriendlyMessage(handledErr))
			}
		} else {
			fmt.Printf("文件 %s 处理成功\n", file)
		}
	}
	
	// 显示批量处理结果
	if collector.HasErrors() {
		fmt.Printf("\n批量处理完成，发现 %d 个错误:\n", collector.GetErrorCount())
		fmt.Println(collector.GetSummary())
	} else {
		fmt.Println("\n批量处理完成，没有错误")
	}
}

// 模拟文件处理函数
func processFile(filename string) error {
	switch filename {
	case "file1.pdf":
		return nil // 成功
	case "file2.pdf":
		return pdf.NewPDFError(pdf.ErrorEncrypted, "文件已加密", filename, nil)
	case "file3.pdf":
		return pdf.NewPDFError(pdf.ErrorIO, "读取失败", filename, fmt.Errorf("permission denied"))
	case "file4.pdf":
		return pdf.NewPDFError(pdf.ErrorCorrupted, "文件损坏", filename, nil)
	default:
		return nil
	}
}

// 演示错误链功能
func demonstrateErrorChaining() {
	fmt.Println("=== 错误链演示 ===")
	
	// 创建一个带有原因的错误
	originalErr := fmt.Errorf("底层系统错误")
	pdfErr := pdf.NewPDFError(pdf.ErrorIO, "文件操作失败", "test.pdf", originalErr)
	
	fmt.Printf("PDF错误: %s\n", pdfErr.Error())
	
	// 使用errors.Unwrap获取原始错误
	if unwrapped := pdfErr.Unwrap(); unwrapped != nil {
		fmt.Printf("原始错误: %s\n", unwrapped.Error())
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}