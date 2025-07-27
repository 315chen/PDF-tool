package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func TestCancellationManager_CancelAllJobs(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	// 创建控制器和取消管理器
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	// 创建多个测试上下文
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	
	// 注册多个取消操作
	cancelManager.RegisterCancellation("job1", cancel1)
	cancelManager.RegisterCancellation("job2", cancel2)
	
	// 取消所有任务
	err := cancelManager.CancelAllJobs()
	if err != nil {
		t.Errorf("取消所有任务失败: %v", err)
	}
	
	// 验证所有上下文都被取消
	select {
	case <-ctx1.Done():
		// 正确，上下文1已被取消
	default:
		t.Error("上下文1未被取消")
	}
	
	select {
	case <-ctx2.Done():
		// 正确，上下文2已被取消
	default:
		t.Error("上下文2未被取消")
	}
}

func TestTempFileCleanupTask(t *testing.T) {
	mockFile := &mockFileManager{}
	
	// 创建清理任务
	task := NewTempFileCleanupTask(mockFile)
	
	// 测试描述
	description := task.Description()
	if description != "清理临时文件" {
		t.Errorf("期望描述为 '清理临时文件'，实际为 '%s'", description)
	}
	
	// 执行清理任务
	err := task.Execute()
	if err != nil {
		t.Errorf("清理任务执行失败: %v", err)
	}
}

func TestMemoryCleanupTask(t *testing.T) {
	// 创建内存清理任务
	task := NewMemoryCleanupTask()
	
	// 测试描述
	description := task.Description()
	if description != "清理内存" {
		t.Errorf("期望描述为 '清理内存'，实际为 '%s'", description)
	}
	
	// 执行清理任务
	err := task.Execute()
	if err != nil {
		t.Errorf("内存清理任务执行失败: %v", err)
	}
}

func TestJobStateCleanupTask(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)
	
	// 设置当前任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf"}, "output.pdf")
	controller.currentJob = job
	
	// 创建任务状态清理任务
	task := NewJobStateCleanupTask(controller)
	
	// 测试描述
	description := task.Description()
	if description != "清理任务状态" {
		t.Errorf("期望描述为 '清理任务状态'，实际为 '%s'", description)
	}
	
	// 执行清理任务
	err := task.Execute()
	if err != nil {
		t.Errorf("任务状态清理失败: %v", err)
	}
	
	// 验证任务状态被清理
	if controller.currentJob != nil {
		t.Error("当前任务应该被清理")
	}
}

func TestResourceCleanupTask(t *testing.T) {
	// 创建测试资源
	resource1Called := false
	resource2Called := false
	
	resource1 := func() error {
		resource1Called = true
		return nil
	}
	
	resource2 := func() error {
		resource2Called = true
		return nil
	}
	
	// 创建资源清理任务
	task := NewResourceCleanupTask("test-resources", resource1, resource2)
	
	// 测试描述
	description := task.Description()
	expected := "清理资源: test-resources"
	if description != expected {
		t.Errorf("期望描述为 '%s'，实际为 '%s'", expected, description)
	}
	
	// 执行清理任务
	err := task.Execute()
	if err != nil {
		t.Errorf("资源清理任务执行失败: %v", err)
	}
	
	// 验证所有资源都被清理
	if !resource1Called {
		t.Error("资源1未被清理")
	}
	
	if !resource2Called {
		t.Error("资源2未被清理")
	}
}

func TestCancellationContext(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	// 创建控制器和取消管理器
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	// 创建取消上下文
	parent := context.Background()
	jobID := "test-job-456"
	cancelCtx := NewCancellationContext(parent, jobID, cancelManager)
	
	// 测试初始状态
	if cancelCtx.IsCancelled() {
		t.Error("初始状态不应该被取消")
	}
	
	// 添加清理任务
	cleanupExecuted := false
	cleanupTask := NewResourceCleanupTask("test", func() error {
		cleanupExecuted = true
		return nil
	})
	cancelCtx.AddCleanupTask(cleanupTask)
	
	// 取消操作
	cancelCtx.Cancel()
	
	// 验证取消状态
	if !cancelCtx.IsCancelled() {
		t.Error("应该被标记为已取消")
	}
	
	// 验证清理任务被执行
	if !cleanupExecuted {
		t.Error("清理任务未被执行")
	}
	
	// 验证上下文被取消
	select {
	case <-cancelCtx.Context().Done():
		// 正确，上下文已被取消
	default:
		t.Error("上下文未被取消")
	}
}

func TestOperationExecutor(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	// 创建控制器和取消管理器
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	// 创建操作执行器
	executor := NewOperationExecutor(cancelManager)
	
	// 创建测试操作
	operation := &testOperation{
		canBeCancelled: true,
		duration:       100 * time.Millisecond,
		shouldFail:     false,
	}
	
	// 执行操作
	parent := context.Background()
	jobID := "test-operation-789"
	
	err := executor.ExecuteWithCancellation(parent, jobID, operation)
	if err != nil {
		t.Errorf("操作执行失败: %v", err)
	}
	
	// 验证操作被执行
	if !operation.executed {
		t.Error("操作未被执行")
	}
}

func TestOperationExecutor_Cancellation(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	// 创建控制器和取消管理器
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	// 创建操作执行器
	executor := NewOperationExecutor(cancelManager)
	
	// 创建长时间运行的测试操作
	operation := &testOperation{
		canBeCancelled: true,
		duration:       2 * time.Second,
		shouldFail:     false,
	}
	
	// 创建可取消的上下文
	parent, cancel := context.WithCancel(context.Background())
	jobID := "test-cancel-operation"
	
	// 在短时间后取消
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	// 执行操作
	err := executor.ExecuteWithCancellation(parent, jobID, operation)
	
	// 应该返回取消错误
	if err == nil {
		t.Error("期望取消错误，但操作成功完成")
	}
	
	if err.Error() != "操作被取消" {
		t.Errorf("期望取消错误，实际错误: %v", err)
	}
}

// 测试操作实现
type testOperation struct {
	canBeCancelled bool
	duration       time.Duration
	shouldFail     bool
	executed       bool
}

func (to *testOperation) Execute(ctx *CancellationContext) error {
	to.executed = true
	
	if to.shouldFail {
		return fmt.Errorf("测试操作失败")
	}
	
	// 模拟操作执行时间
	select {
	case <-time.After(to.duration):
		return nil
	case <-ctx.WaitForCancellation():
		return fmt.Errorf("操作被取消")
	}
}

func (to *testOperation) CanBeCancelled() bool {
	return to.canBeCancelled
}

func (to *testOperation) EstimatedDuration() time.Duration {
	return to.duration
}

// 基准测试

func BenchmarkCancellationManager_RegisterCancellation(b *testing.B) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		jobID := fmt.Sprintf("job-%d", i)
		cancelManager.RegisterCancellation(jobID, cancel)
		ctx.Done() // 避免内存泄漏
	}
}

func BenchmarkCancellationManager_GracefulCancellation(b *testing.B) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	cancelManager := NewCancellationManager(controller)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		jobID := fmt.Sprintf("job-%d", i)
		cancelManager.RegisterCancellation(jobID, cancel)
		
		err := cancelManager.GracefulCancellation(jobID, 100*time.Millisecond)
		if err != nil {
			b.Errorf("优雅取消失败: %v", err)
		}
		
		ctx.Done() // 避免内存泄漏
	}
}