//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF服务重试和恢复机制演示 ===\n")
	
	// 创建基础PDF服务
	baseService := pdf.NewPDFService()
	
	// 创建带重试功能的服务，限制内存为100MB
	serviceWithRetry := pdf.NewServiceWithRetry(baseService, 100)
	
	// 1. 基本重试功能演示
	demonstrateBasicRetry(serviceWithRetry)
	
	// 2. 上下文取消演示
	demonstrateContextCancellation(serviceWithRetry)
	
	// 3. 批量处理演示
	demonstrateBatchProcessing(serviceWithRetry)
	
	// 4. 内存感知合并演示
	demonstrateMemoryAwareMerge(serviceWithRetry)
	
	// 5. 错误管理演示
	demonstrateErrorManagement(serviceWithRetry)
	
	// 6. 服务统计信息演示
	demonstrateServiceStats(serviceWithRetry)
}

func demonstrateBasicRetry(service *pdf.ServiceWithRetry) {
	fmt.Println("1. 基本重试功能演示:")
	
	// 模拟验证一个可能暂时不可用的文件
	fmt.Printf("  验证PDF文件...\n")
	err := service.ValidatePDF("test.pdf")
	if err != nil {
		fmt.Printf("    验证失败: %v\n", err)
	} else {
		fmt.Printf("    验证成功\n")
	}
	
	// 获取PDF信息
	fmt.Printf("  获取PDF信息...\n")
	info, err := service.GetPDFInfo("test.pdf")
	if err != nil {
		fmt.Printf("    获取信息失败: %v\n", err)
	} else {
		fmt.Printf("    获取信息成功: %+v\n", info)
	}
	
	fmt.Println()
}

func demonstrateContextCancellation(service *pdf.ServiceWithRetry) {
	fmt.Println("2. 上下文取消演示:")
	
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	fmt.Printf("  开始合并操作（2秒超时）...\n")
	start := time.Now()
	
	err := service.MergePDFsWithContext(ctx, "main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf", nil)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("    操作失败: %v (耗时: %v)\n", err, duration)
	} else {
		fmt.Printf("    操作成功 (耗时: %v)\n", duration)
	}
	
	fmt.Println()
}

func demonstrateBatchProcessing(service *pdf.ServiceWithRetry) {
	fmt.Println("3. 批量处理演示:")
	
	// 创建批量合并任务
	jobs := []model.MergeJob{
		*model.NewMergeJob("doc1.pdf", []string{"add1.pdf"}, "output1.pdf"),
		*model.NewMergeJob("doc2.pdf", []string{"add2.pdf"}, "output2.pdf"),
		*model.NewMergeJob("doc3.pdf", []string{"add3.pdf"}, "output3.pdf"),
	}
	
	fmt.Printf("  执行 %d 个合并任务...\n", len(jobs))
	errors := service.BatchMergePDFs(jobs)
	
	// 统计结果
	successCount := 0
	for _, job := range jobs {
		if job.Status == model.JobCompleted {
			successCount++
		}
	}
	
	fmt.Printf("  批量处理结果:\n")
	fmt.Printf("    成功: %d/%d\n", successCount, len(jobs))
	fmt.Printf("    失败: %d/%d\n", len(errors), len(jobs))
	
	if len(errors) > 0 {
		fmt.Printf("  错误详情:\n")
		for i, err := range errors {
			fmt.Printf("    %d. %v\n", i+1, err)
		}
	}
	
	// 批量验证文件
	fmt.Printf("  批量验证文件...\n")
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf"}
	validationResults := service.BatchValidatePDFs(files)
	
	fmt.Printf("  验证结果:\n")
	for _, file := range files {
		if err, exists := validationResults[file]; exists {
			fmt.Printf("    %s: 失败 - %v\n", file, err)
		} else {
			fmt.Printf("    %s: 成功\n", file)
		}
	}
	
	fmt.Println()
}

func demonstrateMemoryAwareMerge(service *pdf.ServiceWithRetry) {
	fmt.Println("4. 内存感知合并演示:")
	
	// 模拟大量文件合并
	files := []string{
		"doc1.pdf", "doc2.pdf", "doc3.pdf", "doc4.pdf", "doc5.pdf",
		"doc6.pdf", "doc7.pdf", "doc8.pdf", "doc9.pdf", "doc10.pdf",
	}
	
	fmt.Printf("  合并 %d 个文件（分批处理，每批最多3个）...\n", len(files))
	
	err := service.MemoryAwareMerge(files, "large_output.pdf", 3)
	if err != nil {
		fmt.Printf("    内存感知合并失败: %v\n", err)
	} else {
		fmt.Printf("    内存感知合并成功\n")
	}
	
	fmt.Println()
}

func demonstrateErrorManagement(service *pdf.ServiceWithRetry) {
	fmt.Println("5. 错误管理演示:")
	
	// 执行一些可能失败的操作
	fmt.Printf("  执行可能失败的操作...\n")
	
	service.ValidatePDF("nonexistent.pdf")
	service.MergePDFs("invalid.pdf", []string{"also_invalid.pdf"}, "output.pdf", nil)
	
	// 检查收集的错误
	errors := service.GetErrors()
	fmt.Printf("  收集到 %d 个错误:\n", len(errors))
	
	if len(errors) > 0 {
		summary := service.GetErrorSummary()
		fmt.Printf("  错误摘要:\n%s", summary)
	}
	
	// 清空错误
	fmt.Printf("  清空错误记录...\n")
	service.ClearErrors()
	
	errors = service.GetErrors()
	fmt.Printf("  清空后错误数量: %d\n", len(errors))
	
	fmt.Println()
}

func demonstrateServiceStats(service *pdf.ServiceWithRetry) {
	fmt.Println("6. 服务统计信息演示:")
	
	stats := service.GetServiceStats()
	
	fmt.Printf("  服务统计信息:\n")
	fmt.Printf("    服务类型: %v\n", stats["service_type"])
	fmt.Printf("    重试启用: %v\n", stats["retry_enabled"])
	fmt.Printf("    恢复启用: %v\n", stats["recovery_enabled"])
	fmt.Printf("    当前内存使用: %v MB\n", stats["alloc_mb"])
	fmt.Printf("    最大允许内存: %v MB\n", stats["max_allowed_mb"])
	fmt.Printf("    错误数量: %v\n", stats["error_count"])
	fmt.Printf("    是否有错误: %v\n", stats["has_errors"])
	fmt.Printf("    GC次数: %v\n", stats["num_gc"])
	
	fmt.Println()
}

// 演示健壮的文件操作
func demonstrateRobustFileOperation(service *pdf.ServiceWithRetry) {
	fmt.Println("7. 健壮文件操作演示:")
	
	// 创建一个测试文件
	testFile := "test_robust.pdf"
	err := createTestPDFFile(testFile)
	if err != nil {
		fmt.Printf("  创建测试文件失败: %v\n", err)
		return
	}
	defer os.Remove(testFile)
	
	// 使用健壮的文件操作
	fmt.Printf("  执行健壮的文件操作...\n")
	err = service.RobustFileOperation(testFile, func(path string) error {
		fmt.Printf("    处理文件: %s\n", path)
		return nil
	})
	
	if err != nil {
		fmt.Printf("    健壮操作失败: %v\n", err)
	} else {
		fmt.Printf("    健壮操作成功\n")
	}
	
	// 测试不存在的文件
	fmt.Printf("  测试不存在的文件...\n")
	err = service.RobustFileOperation("nonexistent.pdf", func(path string) error {
		return nil
	})
	
	if err != nil {
		fmt.Printf("    预期的错误: %v\n", err)
	}
	
	fmt.Println()
}

// 演示安全输出操作
func demonstrateSafeOutputOperation(service *pdf.ServiceWithRetry) {
	fmt.Println("8. 安全输出操作演示:")
	
	outputPath := "output/safe/test_output.pdf"
	
	fmt.Printf("  执行安全输出操作到: %s\n", outputPath)
	err := service.SafeOutputOperation(outputPath, func(path string) error {
		fmt.Printf("    写入文件: %s\n", path)
		// 模拟写入操作
		return os.WriteFile(path, []byte("test content"), 0644)
	})
	
	if err != nil {
		fmt.Printf("    安全输出操作失败: %v\n", err)
	} else {
		fmt.Printf("    安全输出操作成功\n")
		// 清理创建的文件和目录
		os.Remove(outputPath)
		os.RemoveAll("output")
	}
	
	fmt.Println()
}

// 创建测试PDF文件
func createTestPDFFile(filename string) error {
	// 创建一个简单的PDF文件内容（这不是真正的PDF，只是用于演示）
	content := `%PDF-1.4
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
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
200
%%EOF`
	
	return os.WriteFile(filename, []byte(content), 0644)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}