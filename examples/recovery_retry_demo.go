//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== 错误恢复和重试机制功能演示 ===\n")

	// 1. 演示重试配置和策略
	demonstrateRetryConfiguration()

	// 2. 演示指数退避重试
	demonstrateExponentialBackoff()

	// 3. 演示内存管理和恢复
	demonstrateMemoryManagement()

	// 4. 演示恢复管理器
	demonstrateRecoveryManager()

	// 5. 演示断路器模式
	demonstrateCircuitBreaker()

	// 6. 演示上下文取消和超时
	demonstrateContextCancellation()

	// 7. 演示完整的恢复重试流程
	demonstrateCompleteRecoveryFlow()

	fmt.Println("\n=== 错误恢复和重试机制演示完成 ===")
}

func demonstrateRetryConfiguration() {
	fmt.Println("1. 重试配置和策略演示:")
	
	// 1.1 默认重试配置
	fmt.Println("\n   1.1 默认重试配置:")
	defaultConfig := pdf.DefaultRetryConfig()
	fmt.Printf("   - 最大重试次数: %d\n", defaultConfig.MaxRetries)
	fmt.Printf("   - 初始延迟: %v\n", defaultConfig.InitialDelay)
	fmt.Printf("   - 最大延迟: %v\n", defaultConfig.MaxDelay)
	fmt.Printf("   - 退避因子: %.1f\n", defaultConfig.BackoffFactor)
	fmt.Printf("   - 总超时时间: %v\n", defaultConfig.Timeout)
	
	// 1.2 自定义重试配置
	fmt.Println("\n   1.2 自定义重试配置:")
	customConfigs := []*pdf.RetryConfig{
		{
			MaxRetries:    5,
			InitialDelay:  50 * time.Millisecond,
			MaxDelay:      2 * time.Second,
			BackoffFactor: 1.5,
			Timeout:       15 * time.Second,
		},
		{
			MaxRetries:    2,
			InitialDelay:  200 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffFactor: 3.0,
			Timeout:       5 * time.Second,
		},
	}
	
	for i, config := range customConfigs {
		fmt.Printf("   配置 %d:\n", i+1)
		fmt.Printf("   - 最大重试: %d, 初始延迟: %v, 退避因子: %.1f\n", 
			config.MaxRetries, config.InitialDelay, config.BackoffFactor)
		
		// 计算重试延迟序列
		delays := calculateRetryDelays(config)
		fmt.Printf("   - 重试延迟序列: %v\n", delays)
	}
	
	// 1.3 重试策略比较
	fmt.Println("\n   1.3 重试策略比较:")
	strategies := map[string]*pdf.RetryConfig{
		"快速重试": {MaxRetries: 5, InitialDelay: 10 * time.Millisecond, BackoffFactor: 1.2},
		"标准重试": {MaxRetries: 3, InitialDelay: 100 * time.Millisecond, BackoffFactor: 2.0},
		"保守重试": {MaxRetries: 2, InitialDelay: 500 * time.Millisecond, BackoffFactor: 3.0},
	}
	
	for name, config := range strategies {
		totalTime := calculateTotalRetryTime(config)
		fmt.Printf("   - %s: 总时间约 %v\n", name, totalTime)
	}
	
	fmt.Println()
}

func demonstrateExponentialBackoff() {
	fmt.Println("2. 指数退避重试演示:")
	
	// 2.1 创建重试管理器
	fmt.Println("\n   2.1 创建重试管理器:")
	config := &pdf.RetryConfig{
		MaxRetries:    4,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       10 * time.Second,
	}
	
	errorHandler := pdf.NewDefaultErrorHandler(config.MaxRetries)
	retryManager := pdf.NewRetryManager(config, errorHandler)
	
	fmt.Printf("   重试管理器创建完成\n")
	
	// 2.2 模拟不同的重试场景
	fmt.Println("\n   2.2 重试场景演示:")
	
	scenarios := []struct {
		name        string
		operation   func() error
		expectRetry bool
	}{
		{
			name: "立即成功",
			operation: func() error {
				return nil
			},
			expectRetry: false,
		},
		{
			name: "第3次成功",
			operation: func() func() error {
				attempt := 0
				return func() error {
					attempt++
					if attempt < 3 {
						return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", "test.pdf", nil)
					}
					return nil
				}
			}(),
			expectRetry: true,
		},
		{
			name: "权限错误(不重试)",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorPermission, "权限被拒绝", "test.pdf", nil)
			},
			expectRetry: false,
		},
		{
			name: "内存错误(重试)",
			operation: func() func() error {
				attempt := 0
				return func() error {
					attempt++
					if attempt < 2 {
						return pdf.NewPDFError(pdf.ErrorMemory, "内存不足", "test.pdf", nil)
					}
					return nil
				}
			}(),
			expectRetry: true,
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

func demonstrateMemoryManagement() {
	fmt.Println("3. 内存管理和恢复演示:")
	
	// 3.1 创建内存管理器
	fmt.Println("\n   3.1 创建内存管理器:")
	memoryManager := pdf.NewMemoryManager(50 * 1024 * 1024) // 50MB限制
	
	fmt.Printf("   内存管理器创建完成，限制: 50MB\n")
	
	// 3.2 内存状态检查
	fmt.Println("\n   3.2 内存状态检查:")
	initialStats := memoryManager.GetMemoryStats()
	fmt.Printf("   初始内存状态:\n")
	printMemoryStats(initialStats)
	
	// 3.3 内存使用检查
	fmt.Println("\n   3.3 内存使用检查:")
	err := memoryManager.CheckMemoryUsage()
	if err != nil {
		fmt.Printf("   内存检查失败: %v\n", err)
	} else {
		fmt.Printf("   内存检查通过 ✓\n")
	}
	
	// 3.4 强制垃圾回收
	fmt.Println("\n   3.4 强制垃圾回收:")
	beforeGC := memoryManager.GetMemoryStats()
	fmt.Printf("   GC前: %d MB, GC次数: %v\n", 
		beforeGC["alloc_mb"], beforeGC["num_gc"])
	
	memoryManager.ForceGC()
	
	afterGC := memoryManager.GetMemoryStats()
	fmt.Printf("   GC后: %d MB, GC次数: %v\n", 
		afterGC["alloc_mb"], afterGC["num_gc"])
	
	// 3.5 内存压力模拟
	fmt.Println("\n   3.5 内存压力模拟:")
	fmt.Printf("   分配大块内存进行测试...\n")
	
	// 分配一些内存来模拟压力
	data := make([][]byte, 10)
	for i := range data {
		data[i] = make([]byte, 1024*1024) // 1MB each
	}
	
	pressureStats := memoryManager.GetMemoryStats()
	fmt.Printf("   压力测试后: %d MB\n", pressureStats["alloc_mb"])
	
	// 释放内存
	data = nil
	runtime.GC()
	
	finalStats := memoryManager.GetMemoryStats()
	fmt.Printf("   释放后: %d MB\n", finalStats["alloc_mb"])
	
	fmt.Println()
}

func demonstrateRecoveryManager() {
	fmt.Println("4. 恢复管理器演示:")
	
	// 4.1 创建恢复管理器
	fmt.Println("\n   4.1 创建恢复管理器:")
	recoveryManager := pdf.NewRecoveryManager(100) // 100MB限制
	
	fmt.Printf("   恢复管理器创建完成\n")
	
	// 4.2 恢复场景测试
	fmt.Println("\n   4.2 恢复场景测试:")
	
	recoveryScenarios := []struct {
		name      string
		operation func() error
		expectRecovery bool
	}{
		{
			name: "正常操作",
			operation: func() error {
				return nil
			},
			expectRecovery: false,
		},
		{
			name: "IO错误恢复",
			operation: func() func() error {
				attempt := 0
				return func() error {
					attempt++
					if attempt == 1 {
						return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", "test.pdf", nil)
					}
					return nil
				}
			}(),
			expectRecovery: true,
		},
		{
			name: "权限错误(无法恢复)",
			operation: func() error {
				return pdf.NewPDFError(pdf.ErrorPermission, "权限被拒绝", "test.pdf", nil)
			},
			expectRecovery: false,
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
		fmt.Printf("   - 预期恢复: %t\n", scenario.expectRecovery)
		
		// 显示恢复统计
		stats := recoveryManager.GetRecoveryStats()
		fmt.Printf("   - 当前内存: %v MB\n", stats["alloc_mb"])
		fmt.Println()
	}
	
	// 4.3 错误统计
	fmt.Println("   4.3 恢复管理器错误统计:")
	errorSummary := recoveryManager.GetErrorSummary()
	if errorSummary != "" {
		fmt.Printf("   %s\n", errorSummary)
	} else {
		fmt.Printf("   无错误记录\n")
	}
	
	// 清空错误
	recoveryManager.ClearErrors()
	fmt.Printf("   错误记录已清空\n")
	
	fmt.Println()
}

func demonstrateCircuitBreaker() {
	fmt.Println("5. 断路器模式演示:")
	
	// 5.1 模拟断路器行为
	fmt.Println("\n   5.1 断路器状态演示:")
	
	// 创建一个简单的断路器模拟
	circuitBreaker := &SimpleCircuitBreaker{
		failureThreshold: 3,
		timeout:         2 * time.Second,
	}
	
	fmt.Printf("   断路器创建: 失败阈值=%d, 超时=%v\n", 
		circuitBreaker.failureThreshold, circuitBreaker.timeout)
	
	// 5.2 测试断路器状态变化
	fmt.Println("\n   5.2 断路器状态变化:")
	
	operations := []struct {
		name    string
		success bool
	}{
		{"操作1", true},
		{"操作2", false},
		{"操作3", false},
		{"操作4", false}, // 这里应该触发断路器
		{"操作5", true},  // 断路器打开，直接失败
		{"操作6", true},  // 等待超时后重试
	}
	
	for i, op := range operations {
		fmt.Printf("   %s: ", op.name)
		
		if i == 5 {
			// 等待断路器超时
			time.Sleep(circuitBreaker.timeout + 100*time.Millisecond)
		}
		
		err := circuitBreaker.Execute(func() error {
			if op.success {
				return nil
			}
			return pdf.NewPDFError(pdf.ErrorIO, "模拟失败", "test.pdf", nil)
		})
		
		if err != nil {
			fmt.Printf("失败 - %s (状态: %s)\n", err.Error(), circuitBreaker.GetState())
		} else {
			fmt.Printf("成功 (状态: %s)\n", circuitBreaker.GetState())
		}
	}
	
	fmt.Println()
}

func demonstrateContextCancellation() {
	fmt.Println("6. 上下文取消和超时演示:")
	
	// 6.1 超时控制
	fmt.Println("\n   6.1 超时控制演示:")
	
	config := &pdf.RetryConfig{
		MaxRetries:    5,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       1 * time.Second, // 1秒超时
	}
	
	retryManager := pdf.NewRetryManager(config, pdf.NewDefaultErrorHandler(5))
	
	// 模拟一个会一直失败的操作
	slowOperation := func() error {
		time.Sleep(300 * time.Millisecond) // 每次尝试300ms
		return pdf.NewPDFError(pdf.ErrorIO, "持续失败", "test.pdf", nil)
	}
	
	fmt.Printf("   执行会超时的操作 (超时: %v)...\n", config.Timeout)
	startTime := time.Now()
	
	ctx := context.Background()
	err := retryManager.ExecuteWithContext(ctx, slowOperation)
	duration := time.Since(startTime)
	
	fmt.Printf("   结果: %v\n", err)
	fmt.Printf("   用时: %v\n", duration)
	
	// 6.2 手动取消
	fmt.Println("\n   6.2 手动取消演示:")
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// 启动一个goroutine在1秒后取消
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
		fmt.Printf("   上下文已取消\n")
	}()
	
	fmt.Printf("   执行会被取消的操作...\n")
	startTime = time.Now()
	
	err = retryManager.ExecuteWithContext(ctx, slowOperation)
	duration = time.Since(startTime)
	
	fmt.Printf("   结果: %v\n", err)
	fmt.Printf("   用时: %v\n", duration)
	
	fmt.Println()
}

func demonstrateCompleteRecoveryFlow() {
	fmt.Println("7. 完整恢复重试流程演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "recovery-flow-demo")
	defer os.RemoveAll(tempDir)
	
	// 7.1 初始化组件
	fmt.Println("\n   7.1 初始化恢复重试组件:")
	
	recoveryManager := pdf.NewRecoveryManager(100)
	
	fmt.Printf("   - 恢复管理器初始化完成\n")
	
	// 7.2 创建测试文件
	fmt.Println("\n   7.2 创建测试文件:")
	testFile := filepath.Join(tempDir, "recovery_test.pdf")
	os.WriteFile(testFile, []byte("test content"), 0644)
	fmt.Printf("   - 测试文件: %s\n", filepath.Base(testFile))
	
	// 7.3 执行完整恢复流程
	fmt.Println("\n   7.3 执行完整恢复重试流程:")
	
	// 模拟复杂的PDF处理操作
	complexOperation := func() func() error {
		step := 0
		return func() error {
			step++
			fmt.Printf("   执行步骤 %d: ", step)
			
			switch step {
			case 1:
				fmt.Printf("文件验证\n")
				return nil
			case 2:
				fmt.Printf("内存检查 - 失败\n")
				return pdf.NewPDFError(pdf.ErrorMemory, "内存不足", testFile, nil)
			case 3:
				fmt.Printf("内存检查 - 成功\n")
				return nil
			case 4:
				fmt.Printf("IO操作 - 失败\n")
				return pdf.NewPDFError(pdf.ErrorIO, "临时IO错误", testFile, nil)
			case 5:
				fmt.Printf("IO操作 - 成功\n")
				return nil
			case 6:
				fmt.Printf("最终处理\n")
				return nil
			default:
				fmt.Printf("完成\n")
				return nil
			}
		}
	}()
	
	// 执行恢复流程
	fmt.Printf("   开始执行恢复重试流程...\n")
	startTime := time.Now()
	
	err := recoveryManager.ExecuteWithRecovery(complexOperation)
	duration := time.Since(startTime)
	
	// 7.4 结果分析
	fmt.Printf("\n   7.4 流程结果分析:\n")
	if err != nil {
		fmt.Printf("   - 最终结果: 失败 - %s\n", err.Error())
	} else {
		fmt.Printf("   - 最终结果: 成功 ✓\n")
	}
	fmt.Printf("   - 总用时: %v\n", duration)
	
	// 7.5 统计信息
	fmt.Printf("   7.5 统计信息:\n")
	stats := recoveryManager.GetRecoveryStats()
	fmt.Printf("   - 当前内存: %v MB\n", stats["alloc_mb"])
	fmt.Printf("   - 总分配: %v MB\n", stats["total_alloc_mb"])
	fmt.Printf("   - GC次数: %v\n", stats["num_gc"])
	fmt.Printf("   - GC CPU占用: %.2f%%\n", stats["gc_cpu_fraction"].(float64)*100)
	
	// 7.6 错误摘要
	fmt.Printf("   7.6 错误摘要:\n")
	errorSummary := recoveryManager.GetErrorSummary()
	if errorSummary != "" {
		fmt.Printf("   %s\n", errorSummary)
	} else {
		fmt.Printf("   无错误记录\n")
	}
	
	fmt.Println("\n   完整恢复重试流程演示完成 🎉")
	fmt.Println("   所有恢复重试组件协同工作正常")
	
	fmt.Println()
}

// 辅助函数和结构

func calculateRetryDelays(config *pdf.RetryConfig) []time.Duration {
	delays := make([]time.Duration, config.MaxRetries)
	delay := config.InitialDelay
	
	for i := 0; i < config.MaxRetries; i++ {
		delays[i] = delay
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}
	
	return delays
}

func calculateTotalRetryTime(config *pdf.RetryConfig) time.Duration {
	delays := calculateRetryDelays(config)
	total := time.Duration(0)
	for _, delay := range delays {
		total += delay
	}
	return total
}

func printMemoryStats(stats map[string]interface{}) {
	fmt.Printf("     - 当前分配: %v MB\n", stats["alloc_mb"])
	fmt.Printf("     - 总分配: %v MB\n", stats["total_alloc_mb"])
	fmt.Printf("     - 系统内存: %v MB\n", stats["sys_mb"])
	fmt.Printf("     - GC次数: %v\n", stats["num_gc"])
	fmt.Printf("     - 最大允许: %v MB\n", stats["max_allowed_mb"])
}

// 简单断路器实现
type SimpleCircuitBreaker struct {
	failureCount     int
	failureThreshold int
	lastFailureTime  time.Time
	timeout          time.Duration
	state            string
}

func (cb *SimpleCircuitBreaker) Execute(operation func() error) error {
	// 检查断路器状态
	if cb.state == "open" {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = "half-open"
		} else {
			return pdf.NewPDFError(pdf.ErrorIO, "断路器打开", "", nil)
		}
	}
	
	// 执行操作
	err := operation()
	
	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		
		if cb.failureCount >= cb.failureThreshold {
			cb.state = "open"
		}
		
		return err
	}
	
	// 成功时重置
	cb.failureCount = 0
	cb.state = "closed"
	return nil
}

func (cb *SimpleCircuitBreaker) GetState() string {
	if cb.state == "" {
		return "closed"
	}
	return cb.state
}
