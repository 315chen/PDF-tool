package pdf

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.MaxRetries)
	}
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay to be 100ms, got %v", config.InitialDelay)
	}
	if config.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor to be 2.0, got %f", config.BackoffFactor)
	}
}

func TestRetryManager_Execute_Success(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	rm := NewRetryManager(config, NewDefaultErrorHandler(3))

	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return NewPDFError(ErrorIO, "temporary failure", "test.pdf", nil)
		}
		return nil // 第三次成功
	}

	err := rm.Execute(operation)
	if err != nil {
		t.Errorf("Expected operation to succeed after retries, got error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, got %d", callCount)
	}
}

func TestRetryManager_Execute_MaxRetriesExceeded(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	rm := NewRetryManager(config, NewDefaultErrorHandler(3))

	callCount := 0
	operation := func() error {
		callCount++
		return NewPDFError(ErrorIO, "persistent failure", "test.pdf", nil)
	}

	err := rm.Execute(operation)
	if err == nil {
		t.Error("Expected operation to fail after max retries")
	}
	if callCount != 3 { // 初始调用 + 2次重试
		t.Errorf("Expected operation to be called 3 times, got %d", callCount)
	}
}

func TestRetryManager_Execute_NonRetryableError(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	rm := NewRetryManager(config, NewDefaultErrorHandler(3))

	callCount := 0
	operation := func() error {
		callCount++
		return NewPDFError(ErrorInvalidFile, "invalid file", "test.pdf", nil)
	}

	err := rm.Execute(operation)
	if err == nil {
		t.Error("Expected operation to fail immediately for non-retryable error")
	}
	if callCount != 1 {
		t.Errorf("Expected operation to be called only once, got %d", callCount)
	}
}

func TestRetryManager_ExecuteWithContext_Timeout(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:    5,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      500 * time.Millisecond,
		BackoffFactor: 2.0,
		Timeout:       200 * time.Millisecond,
	}

	rm := NewRetryManager(config, NewDefaultErrorHandler(5))

	ctx := context.Background()
	operation := func() error {
		return NewPDFError(ErrorIO, "slow operation", "test.pdf", nil)
	}

	start := time.Now()
	err := rm.ExecuteWithContext(ctx, operation)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected operation to fail due to timeout")
	}

	// 应该在超时时间附近结束
	if duration > 300*time.Millisecond {
		t.Errorf("Operation took too long: %v", duration)
	}
}

func TestRetryManager_ExecuteWithContext_Cancellation(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:    5,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      200 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	rm := NewRetryManager(config, NewDefaultErrorHandler(5))

	ctx, cancel := context.WithCancel(context.Background())

	// 100ms后取消上下文
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	operation := func() error {
		return NewPDFError(ErrorIO, "operation", "test.pdf", nil)
	}

	start := time.Now()
	err := rm.ExecuteWithContext(ctx, operation)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected operation to fail due to cancellation")
	}

	// 应该在取消时间附近结束
	if duration > 150*time.Millisecond {
		t.Errorf("Operation took too long after cancellation: %v", duration)
	}
}

func TestMemoryManager_CheckMemoryUsage(t *testing.T) {
	// 设置一个合理的内存限制用于测试
	mm := NewMemoryManager(100 * 1024 * 1024) // 100MB

	// 正常情况下应该不会超限
	err := mm.CheckMemoryUsage()
	if err != nil {
		t.Errorf("Unexpected memory error: %v", err)
	}
}

func TestMemoryManager_ForceGC(t *testing.T) {
	mm := NewMemoryManager(100 * 1024 * 1024) // 100MB

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 分配一些内存
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	mm.ForceGC()

	runtime.ReadMemStats(&m2)

	// GC后NumGC应该增加
	if m2.NumGC <= m1.NumGC {
		t.Error("Expected GC count to increase after ForceGC")
	}
}

func TestMemoryManager_GetMemoryStats(t *testing.T) {
	mm := NewMemoryManager(100 * 1024 * 1024) // 100MB

	stats := mm.GetMemoryStats()

	expectedKeys := []string{
		"alloc_mb", "total_alloc_mb", "sys_mb", "num_gc",
		"gc_cpu_fraction", "max_allowed_mb", "gc_threshold_mb",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected key %s to exist in memory stats", key)
		}
	}

	if stats["max_allowed_mb"] != int64(100) {
		t.Errorf("Expected max_allowed_mb to be 100, got %v", stats["max_allowed_mb"])
	}
}

func TestRecoveryManager_ExecuteWithRecovery_Success(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 2 {
			return NewPDFError(ErrorIO, "temporary failure", "test.pdf", nil)
		}
		return nil
	}

	err := rm.ExecuteWithRecovery(operation)
	if err != nil {
		t.Errorf("Expected operation to succeed with recovery, got error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("Expected operation to be called at least 2 times, got %d", callCount)
	}
}

func TestRecoveryManager_ExecuteWithRecovery_MemoryError(t *testing.T) {
	// 设置很小的内存限制来触发内存错误
	rm := NewRecoveryManager(1) // 1MB，很容易超限

	operation := func() error {
		// 分配大量内存来触发内存错误
		data := make([]byte, 10*1024*1024) // 10MB
		_ = data
		return nil
	}

	err := rm.ExecuteWithRecovery(operation)
	// 由于内存限制很小，应该会失败
	if err == nil {
		t.Error("Expected memory error due to small memory limit")
	}

	pdfErr, ok := err.(*PDFError)
	if ok && pdfErr.Type != ErrorMemory {
		t.Errorf("Expected memory error, got %v", pdfErr.Type)
	}
}

func TestRecoveryManager_tryMemoryRecovery(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	// 尝试内存恢复
	success := rm.tryMemoryRecovery()

	// 在正常情况下应该成功
	if !success {
		t.Error("Expected memory recovery to succeed under normal conditions")
	}
}

func TestRecoveryManager_tryErrorRecovery(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	tests := []struct {
		name          string
		err           error
		shouldRecover bool
	}{
		{
			name:          "IO error should allow recovery",
			err:           NewPDFError(ErrorIO, "IO failure", "test.pdf", nil),
			shouldRecover: true,
		},
		{
			name:          "Permission error should not allow recovery",
			err:           NewPDFError(ErrorPermission, "Permission denied", "test.pdf", nil),
			shouldRecover: false,
		},
		{
			name:          "Regular error should not allow recovery",
			err:           errors.New("regular error"),
			shouldRecover: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recoveryErr := rm.tryErrorRecovery(tt.err)
			recovered := recoveryErr == nil

			if recovered != tt.shouldRecover {
				t.Errorf("Expected recovery success to be %t, got %t", tt.shouldRecover, recovered)
			}
		})
	}
}

func TestRecoveryManager_GetRecoveryStats(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	// 添加一些错误
	rm.errorCollector.Add(errors.New("test error 1"))
	rm.errorCollector.Add(errors.New("test error 2"))

	stats := rm.GetRecoveryStats()

	if stats["error_count"] != 2 {
		t.Errorf("Expected error_count to be 2, got %v", stats["error_count"])
	}
	if stats["has_errors"] != true {
		t.Errorf("Expected has_errors to be true, got %v", stats["has_errors"])
	}

	// 应该包含内存统计信息
	if _, exists := stats["alloc_mb"]; !exists {
		t.Error("Expected memory stats to be included")
	}
}

func TestRecoveryManager_ClearErrors(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	// 添加错误
	rm.errorCollector.Add(errors.New("test error"))

	if !rm.errorCollector.HasErrors() {
		t.Error("Expected to have errors before clearing")
	}

	rm.ClearErrors()

	if rm.errorCollector.HasErrors() {
		t.Error("Expected no errors after clearing")
	}
}

func TestRecoveryManager_GetErrors(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	rm.errorCollector.Add(err1)
	rm.errorCollector.Add(err2)

	errors := rm.GetErrors()

	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
	if errors[0] != err1 || errors[1] != err2 {
		t.Error("Errors not returned correctly")
	}
}

func TestRecoveryManager_GetErrorSummary(t *testing.T) {
	rm := NewRecoveryManager(100) // 100MB

	rm.errorCollector.Add(errors.New("test error"))

	summary := rm.GetErrorSummary()

	if summary == "没有错误" {
		t.Error("Expected error summary to contain error information")
	}
	if !contains(summary, "共发现 1 个错误") {
		t.Errorf("Expected summary to contain error count, got: %s", summary)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
