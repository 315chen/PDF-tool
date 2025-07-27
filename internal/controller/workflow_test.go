package controller

import (
	"context"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func TestWorkflowManager_ExecuteWorkflow(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)

	// 创建工作流程管理器
	workflowManager := NewWorkflowManager(controller)

	// 创建测试任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")

	// 设置回调以跟踪进度
	var progressUpdates []string
	controller.SetProgressCallback(func(progress float64, status, detail string) {
		progressUpdates = append(progressUpdates, status)
	})

	// 执行工作流程
	ctx := context.Background()
	err := workflowManager.ExecuteWorkflow(ctx, job)

	if err != nil {
		t.Errorf("工作流程执行失败: %v", err)
	}

	// 验证进度更新
	expectedSteps := []string{"文件验证", "准备合并", "处理加密文件", "合并文件", "完成处理", "已完成"}

	if len(progressUpdates) < len(expectedSteps) {
		t.Errorf("进度更新不足，期望至少 %d 个步骤，实际 %d 个", len(expectedSteps), len(progressUpdates))
	}

	// 检查是否包含关键步骤
	stepFound := make(map[string]bool)
	for _, update := range progressUpdates {
		stepFound[update] = true
	}

	for _, expectedStep := range expectedSteps {
		if !stepFound[expectedStep] {
			t.Errorf("缺少预期的步骤: %s", expectedStep)
		}
	}
}

func TestWorkflowManager_CancellationSupport(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)

	// 创建工作流程管理器
	workflowManager := NewWorkflowManager(controller)

	// 创建测试任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf"}, "output.pdf")

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 在短时间后取消
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// 执行工作流程
	err := workflowManager.ExecuteWorkflow(ctx, job)

	// 应该返回取消错误
	if err == nil {
		t.Error("期望取消错误，但工作流程成功完成")
	}

	if err != context.Canceled {
		t.Errorf("期望取消错误，实际错误: %v", err)
	}
}

func TestStreamingMerger_MergeStreaming(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)

	// 创建流式合并器
	streamingMerger := NewStreamingMerger(controller)

	// 创建测试任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")

	// 设置回调以跟踪进度
	var progressUpdates []string
	controller.SetProgressCallback(func(progress float64, status, detail string) {
		progressUpdates = append(progressUpdates, status)
	})

	// 执行流式合并
	ctx := context.Background()
	err := streamingMerger.MergeStreaming(ctx, job, nil)

	if err != nil {
		t.Errorf("流式合并失败: %v", err)
	}

	// 验证进度更新
	if len(progressUpdates) == 0 {
		t.Log("警告：没有收到进度更新，这可能是因为使用了标准合并而不是流式合并")
	}
}

func TestCancellationManager_GracefulCancellation(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)

	// 创建取消管理器
	cancelManager := NewCancellationManager(controller)

	// 创建测试上下文
	ctx, cancel := context.WithCancel(context.Background())
	jobID := "test-job-123"

	// 注册取消操作
	cancelManager.RegisterCancellation(jobID, cancel)

	// 添加清理任务
	cleanupExecuted := false
	cleanupTask := NewResourceCleanupTask("test", func() error {
		cleanupExecuted = true
		return nil
	})
	cancelManager.AddCleanupTask(cleanupTask)

	// 执行优雅取消
	err := cancelManager.GracefulCancellation(jobID, 1*time.Second)

	if err != nil {
		t.Errorf("优雅取消失败: %v", err)
	}

	// 验证清理任务被执行
	if !cleanupExecuted {
		t.Error("清理任务未被执行")
	}

	// 验证上下文被取消
	select {
	case <-ctx.Done():
		// 正确，上下文已被取消
	default:
		t.Error("上下文未被取消")
	}
}

func TestMemoryMonitor_IsMemoryLow(t *testing.T) {
	// 创建内存监控器
	monitor := NewMemoryMonitor(100 * 1024 * 1024) // 100MB

	// 启动监控
	monitor.Start()
	defer monitor.Stop()

	// 检查内存状态
	isLow := monitor.IsMemoryLow()

	// 这个测试结果取决于当前系统内存使用情况
	// 我们只验证方法不会崩溃
	t.Logf("内存是否不足: %v", isLow)
}

func TestBatchProcessor_ProcessBatch(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)

	// 创建流式合并器和批处理器
	streamingMerger := NewStreamingMerger(controller)
	batchProcessor := NewBatchProcessor(streamingMerger)

	// 创建测试文件列表
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf"}
	outputPath := "batch_output.pdf"

	// 执行批处理
	ctx := context.Background()
	err := batchProcessor.ProcessBatch(ctx, files, outputPath, nil)

	if err != nil {
		t.Errorf("批处理失败: %v", err)
	}
}

// 基准测试

func BenchmarkWorkflowManager_ExecuteWorkflow(b *testing.B) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器
	controller := NewController(mockPDF, mockFile, config)
	workflowManager := NewWorkflowManager(controller)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := model.NewMergeJob("main.pdf", []string{"add1.pdf"}, "output.pdf")
		ctx := context.Background()

		err := workflowManager.ExecuteWorkflow(ctx, job)
		if err != nil {
			b.Errorf("工作流程执行失败: %v", err)
		}
	}
}

func BenchmarkStreamingMerger_MergeStreaming(b *testing.B) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := model.NewMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")
		ctx := context.Background()

		err := streamingMerger.MergeStreaming(ctx, job, nil)
		if err != nil {
			b.Errorf("流式合并失败: %v", err)
		}
	}
}
