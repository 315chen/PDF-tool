package pdf

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int           // 最大重试次数
	InitialDelay  time.Duration // 初始延迟
	MaxDelay      time.Duration // 最大延迟
	BackoffFactor float64       // 退避因子
	Timeout       time.Duration // 总超时时间
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       30 * time.Second,
	}
}

// RetryableOperation 可重试的操作函数类型
type RetryableOperation func() error

// RetryManager 重试管理器
type RetryManager struct {
	config       *RetryConfig
	errorHandler ErrorHandler
}

// NewRetryManager 创建新的重试管理器
func NewRetryManager(config *RetryConfig, errorHandler ErrorHandler) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}
	if errorHandler == nil {
		errorHandler = NewDefaultErrorHandler(config.MaxRetries)
	}

	return &RetryManager{
		config:       config,
		errorHandler: errorHandler,
	}
}

// Execute 执行可重试的操作
func (rm *RetryManager) Execute(operation RetryableOperation) error {
	return rm.ExecuteWithContext(context.Background(), operation)
}

// ExecuteWithContext 带上下文的执行可重试操作
func (rm *RetryManager) ExecuteWithContext(ctx context.Context, operation RetryableOperation) error {
	// 创建带超时的上下文
	if rm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, rm.config.Timeout)
		defer cancel()
	}

	var lastErr error
	delay := rm.config.InitialDelay

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return NewPDFError(ErrorIO, "操作超时或被取消", "", ctx.Err())
		default:
		}

		// 执行操作
		err := operation()
		if err == nil {
			return nil // 成功
		}

		// 处理错误
		handledErr := rm.errorHandler.HandleError(err)
		lastErr = handledErr

		// 检查是否应该重试
		if !rm.errorHandler.ShouldRetry(handledErr) {
			return handledErr
		}

		// 如果是最后一次尝试，不再延迟
		if attempt == rm.config.MaxRetries {
			break
		}

		// 等待重试延迟
		select {
		case <-ctx.Done():
			return NewPDFError(ErrorIO, "操作超时或被取消", "", ctx.Err())
		case <-time.After(delay):
		}

		// 计算下次延迟（指数退避）
		delay = time.Duration(float64(delay) * rm.config.BackoffFactor)
		if delay > rm.config.MaxDelay {
			delay = rm.config.MaxDelay
		}
	}

	return lastErr
}

// MemoryManager 内存管理器，用于处理内存不足的情况
type MemoryManager struct {
	maxMemoryUsage int64 // 最大内存使用量（字节）
	gcThreshold    int64 // 触发GC的阈值
}

// NewMemoryManager 创建新的内存管理器
func NewMemoryManager(maxMemoryUsage int64) *MemoryManager {
	return &MemoryManager{
		maxMemoryUsage: maxMemoryUsage,
		gcThreshold:    maxMemoryUsage / 2, // 50%时触发GC
	}
}

// CheckMemoryUsage 检查内存使用情况
func (mm *MemoryManager) CheckMemoryUsage() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentUsage := int64(m.Alloc)

	if currentUsage > mm.maxMemoryUsage {
		return NewPDFError(ErrorMemory,
			fmt.Sprintf("内存使用超限: %d MB > %d MB",
				currentUsage/1024/1024, mm.maxMemoryUsage/1024/1024),
			"", nil)
	}

	// 如果内存使用超过阈值，触发GC
	if currentUsage > mm.gcThreshold {
		runtime.GC()
		runtime.ReadMemStats(&m)

		// GC后再次检查
		if int64(m.Alloc) > mm.maxMemoryUsage {
			return NewPDFError(ErrorMemory,
				fmt.Sprintf("GC后内存仍然超限: %d MB > %d MB",
					int64(m.Alloc)/1024/1024, mm.maxMemoryUsage/1024/1024),
				"", nil)
		}
	}

	return nil
}

// ForceGC 强制垃圾回收
func (mm *MemoryManager) ForceGC() {
	runtime.GC()
	runtime.GC() // 执行两次确保彻底清理
}

// GetMemoryStats 获取内存统计信息
func (mm *MemoryManager) GetMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":        int64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb":  int64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":          int64(m.Sys) / 1024 / 1024,
		"num_gc":          m.NumGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"max_allowed_mb":  mm.maxMemoryUsage / 1024 / 1024,
		"gc_threshold_mb": mm.gcThreshold / 1024 / 1024,
	}
}

// RecoveryManager 恢复管理器，处理各种错误恢复策略
type RecoveryManager struct {
	retryManager   *RetryManager
	memoryManager  *MemoryManager
	errorCollector *ErrorCollector
}

// NewRecoveryManager 创建新的恢复管理器
func NewRecoveryManager(maxMemoryMB int64) *RecoveryManager {
	maxMemoryBytes := maxMemoryMB * 1024 * 1024

	return &RecoveryManager{
		retryManager:   NewRetryManager(DefaultRetryConfig(), NewDefaultErrorHandler(3)),
		memoryManager:  NewMemoryManager(maxMemoryBytes),
		errorCollector: NewErrorCollector(),
	}
}

// ExecuteWithRecovery 执行带恢复机制的操作
func (rm *RecoveryManager) ExecuteWithRecovery(operation RetryableOperation) error {
	// 检查内存使用情况
	if err := rm.memoryManager.CheckMemoryUsage(); err != nil {
		rm.errorCollector.Add(err)

		// 尝试内存恢复
		if rm.tryMemoryRecovery() {
			// 内存恢复成功，继续执行
		} else {
			return err
		}
	}

	// 执行带重试的操作
	err := rm.retryManager.Execute(func() error {
		// 在每次重试前检查内存
		if memErr := rm.memoryManager.CheckMemoryUsage(); memErr != nil {
			return memErr
		}

		return operation()
	})

	if err != nil {
		rm.errorCollector.Add(err)

		// 尝试错误恢复
		if recoveredErr := rm.tryErrorRecovery(err); recoveredErr == nil {
			// 恢复成功，重新执行
			return rm.retryManager.Execute(operation)
		}

		return err
	}

	return nil
}

// tryMemoryRecovery 尝试内存恢复
func (rm *RecoveryManager) tryMemoryRecovery() bool {
	// 强制垃圾回收
	rm.memoryManager.ForceGC()

	// 等待一小段时间让GC完成
	time.Sleep(100 * time.Millisecond)

	// 再次检查内存使用情况
	err := rm.memoryManager.CheckMemoryUsage()
	return err == nil
}

// tryErrorRecovery 尝试错误恢复
func (rm *RecoveryManager) tryErrorRecovery(err error) error {
	pdfErr, ok := err.(*PDFError)
	if !ok {
		return err
	}

	switch pdfErr.Type {
	case ErrorMemory:
		// 内存错误：尝试内存恢复
		if rm.tryMemoryRecovery() {
			return nil // 恢复成功
		}
		return err

	case ErrorIO:
		// IO错误：等待一段时间后重试
		time.Sleep(500 * time.Millisecond)
		return nil // 允许重试

	case ErrorPermission:
		// 权限错误：无法自动恢复
		return err

	default:
		return err
	}
}

// GetRecoveryStats 获取恢复统计信息
func (rm *RecoveryManager) GetRecoveryStats() map[string]interface{} {
	stats := rm.memoryManager.GetMemoryStats()
	stats["error_count"] = rm.errorCollector.GetErrorCount()
	stats["has_errors"] = rm.errorCollector.HasErrors()

	return stats
}

// ClearErrors 清空错误收集器
func (rm *RecoveryManager) ClearErrors() {
	rm.errorCollector.Clear()
}

// GetErrors 获取收集的错误
func (rm *RecoveryManager) GetErrors() []error {
	return rm.errorCollector.GetErrors()
}

// GetErrorSummary 获取错误摘要
func (rm *RecoveryManager) GetErrorSummary() string {
	return rm.errorCollector.GetSummary()
}
