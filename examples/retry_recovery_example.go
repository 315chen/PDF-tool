//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF重试和恢复机制演示 ===\n")
	
	// 1. 重试机制演示
	demonstrateRetryMechanism()
	
	// 2. 内存管理演示
	demonstrateMemoryManagement()
	
	// 3. 恢复管理器演示
	demonstrateRecoveryManager()
	
	// 4. 实际应用场景演示
	demonstrateRealWorldScenario()
}

func demonstrateRetryMechanism() {
	fmt.Println("1. 重试机制演示:")
	
	// 创建重试配置
	config := &pdf.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       10 * time.Second,
	}
	
	// 创建重试管理器
	retryManager := pdf.NewRetryManager(config, pdf.NewDefaultErrorHandler(3))
	
	// 模拟一个会失败几次然后成功的操作
	attemptCount := 0
	operation := func() error {
		attemptCount++
		fmt.Printf("  尝试 %d: ", attemptCount)
		
		if attemptCount < 3 {
			fmt.Println("失败 - IO错误")
			return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", "test.pdf", nil)
		}
		
		fmt.Println("成功")
		return nil
	}
	
	start := time.Now()
	err := retryManager.Execute(operation)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("  最终结果: 失败 - %v\n", err)
	} else {
		fmt.Printf("  最终结果: 成功 (耗时: %v)\n", duration)
	}
	fmt.Println()
}

func demonstrateMemoryManagement() {
	fmt.Println("2. 内存管理演示:")
	
	// 创建内存管理器，限制为50MB
	memoryManager := pdf.NewMemoryManager(50 * 1024 * 1024)
	
	// 显示初始内存状态
	stats := memoryManager.GetMemoryStats()
	fmt.Printf("  初始内存状态:\n")
	fmt.Printf("    已分配: %d MB\n", stats["alloc_mb"])
	fmt.Printf("    系统内存: %d MB\n", stats["sys_mb"])
	fmt.Printf("    最大允许: %d MB\n", stats["max_allowed_mb"])
	fmt.Printf("    GC阈值: %d MB\n", stats["gc_threshold_mb"])
	
	// 检查内存使用情况
	err := memoryManager.CheckMemoryUsage()
	if err != nil {
		fmt.Printf("  内存检查: 失败 - %v\n", err)
	} else {
		fmt.Printf("  内存检查: 通过\n")
	}
	
	// 强制垃圾回收
	fmt.Printf("  执行垃圾回收...\n")
	memoryManager.ForceGC()
	
	// 显示GC后的内存状态
	stats = memoryManager.GetMemoryStats()
	fmt.Printf("  GC后内存状态:\n")
	fmt.Printf("    已分配: %d MB\n", stats["alloc_mb"])
	fmt.Printf("    GC次数: %v\n", stats["num_gc"])
	fmt.Println()
}

func demonstrateRecoveryManager() {
	fmt.Println("3. 恢复管理器演示:")
	
	// 创建恢复管理器，限制内存为100MB
	recoveryManager := pdf.NewRecoveryManager(100)
	
	// 模拟不同类型的错误和恢复
	scenarios := []struct {
		name      string
		operation func() error
	}{
		{
			name: "IO错误恢复",
			operation: func() error {
				// 模拟临时IO错误
				return pdf.NewPDFError(pdf.ErrorIO, "临时网络错误", "remote.pdf", nil)
			},
		},
		{
			name: "成功操作",
			operation: func() error {
				return nil
			},
		},
	}
	
	for _, scenario := range scenarios {
		fmt.Printf("  场景: %s\n", scenario.name)
		
		err := recoveryManager.ExecuteWithRecovery(scenario.operation)
		if err != nil {
			fmt.Printf("    结果: 失败 - %v\n", err)
		} else {
			fmt.Printf("    结果: 成功\n")
		}
		
		// 显示恢复统计信息
		stats := recoveryManager.GetRecoveryStats()
		fmt.Printf("    错误数量: %v\n", stats["error_count"])
		fmt.Printf("    内存使用: %v MB\n", stats["alloc_mb"])
	}
	
	// 显示错误摘要
	if recoveryManager.GetErrors() != nil && len(recoveryManager.GetErrors()) > 0 {
		fmt.Printf("  错误摘要:\n%s", recoveryManager.GetErrorSummary())
	}
	
	// 清空错误
	recoveryManager.ClearErrors()
	fmt.Printf("  错误已清空\n")
	fmt.Println()
}

func demonstrateRealWorldScenario() {
	fmt.Println("4. 实际应用场景演示:")
	
	// 模拟PDF文件合并过程中的错误处理
	recoveryManager := pdf.NewRecoveryManager(100)
	
	files := []string{"doc1.pdf", "doc2.pdf", "doc3.pdf", "doc4.pdf"}
	successCount := 0
	
	for i, file := range files {
		fmt.Printf("  处理文件 %d/%d: %s\n", i+1, len(files), file)
		
		operation := func() error {
			return simulateFileProcessing(file)
		}
		
		err := recoveryManager.ExecuteWithRecovery(operation)
		if err != nil {
			fmt.Printf("    失败: %v\n", err)
		} else {
			fmt.Printf("    成功\n")
			successCount++
		}
		
		// 显示当前内存使用情况
		stats := recoveryManager.GetRecoveryStats()
		fmt.Printf("    内存使用: %v MB\n", stats["alloc_mb"])
	}
	
	fmt.Printf("\n  批量处理结果:\n")
	fmt.Printf("    成功: %d/%d\n", successCount, len(files))
	fmt.Printf("    失败: %d/%d\n", len(files)-successCount, len(files))
	
	// 显示最终统计信息
	stats := recoveryManager.GetRecoveryStats()
	fmt.Printf("    总错误数: %v\n", stats["error_count"])
	fmt.Printf("    最终内存使用: %v MB\n", stats["alloc_mb"])
	
	if recoveryManager.GetErrors() != nil && len(recoveryManager.GetErrors()) > 0 {
		fmt.Printf("\n  详细错误报告:\n")
		fmt.Printf("%s", recoveryManager.GetErrorSummary())
	}
}

// 模拟文件处理函数
func simulateFileProcessing(filename string) error {
	switch filename {
	case "doc1.pdf":
		return nil // 成功
	case "doc2.pdf":
		// 模拟临时IO错误，可以重试
		return pdf.NewPDFError(pdf.ErrorIO, "网络连接超时", filename, nil)
	case "doc3.pdf":
		// 模拟内存错误，可以通过GC恢复
		allocateMemory() // 分配一些内存
		return pdf.NewPDFError(pdf.ErrorMemory, "内存不足", filename, nil)
	case "doc4.pdf":
		// 模拟文件损坏，无法恢复
		return pdf.NewPDFError(pdf.ErrorCorrupted, "文件已损坏", filename, nil)
	default:
		return nil
	}
}

// 分配内存来模拟内存压力
func allocateMemory() {
	data := make([][]byte, 100)
	for i := range data {
		data[i] = make([]byte, 1024*1024) // 1MB each
	}
	runtime.KeepAlive(data)
}

// 演示上下文取消
func demonstrateContextCancellation() {
	fmt.Println("5. 上下文取消演示:")
	
	config := pdf.DefaultRetryConfig()
	retryManager := pdf.NewRetryManager(config, pdf.NewDefaultErrorHandler(3))
	
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	
	// 2秒后取消操作
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Printf("  取消操作...\n")
		cancel()
	}()
	
	// 模拟长时间运行的操作
	operation := func() error {
		time.Sleep(500 * time.Millisecond) // 模拟处理时间
		return pdf.NewPDFError(pdf.ErrorIO, "持续失败", "slow.pdf", nil)
	}
	
	start := time.Now()
	err := retryManager.ExecuteWithContext(ctx, operation)
	duration := time.Since(start)
	
	fmt.Printf("  操作结果: %v\n", err)
	fmt.Printf("  总耗时: %v\n", duration)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}