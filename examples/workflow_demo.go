//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("PDF合并工具 - 工作流程演示")
	fmt.Println("==============================")

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "pdf-merger-workflow-demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建服务实例
	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 演示功能
	demonstrateWorkflow(ctrl)
}

func demonstrateWorkflow(ctrl *controller.Controller) {
	fmt.Println("\n1. 演示完整的合并工作流程")
	fmt.Println("==========================")

	// 设置回调函数
	setupCallbacks(ctrl)

	// 创建测试任务
	job := model.NewMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")

	fmt.Printf("任务ID: %s\n", job.ID)
	fmt.Printf("主文件: %s\n", job.MainFile)
	fmt.Printf("附加文件: %v\n", job.AdditionalFiles)
	fmt.Printf("输出路径: %s\n", job.OutputPath)

	// 启动合并任务
	fmt.Println("\n启动合并任务...")
	err := ctrl.StartMergeJob(job.MainFile, job.AdditionalFiles, job.OutputPath)
	if err != nil {
		fmt.Printf("❌ 启动任务失败: %v\n", err)
	} else {
		fmt.Println("✅ 任务已启动")
	}

	// 等待任务完成
	fmt.Println("\n等待任务完成...")
	waitForJobCompletion(ctrl, 5*time.Second)

	fmt.Println("\n2. 演示任务取消功能")
	fmt.Println("==================")

	// 启动另一个任务
	job2 := model.NewMergeJob("main2.pdf", []string{"add3.pdf"}, "output2.pdf")
	err = ctrl.StartMergeJob(job2.MainFile, job2.AdditionalFiles, job2.OutputPath)
	if err != nil {
		fmt.Printf("❌ 启动任务失败: %v\n", err)
	} else {
		fmt.Println("✅ 第二个任务已启动")
	}

	// 短暂等待后取消
	time.Sleep(100 * time.Millisecond)
	fmt.Println("正在取消任务...")
	err = ctrl.CancelCurrentJob()
	if err != nil {
		fmt.Printf("❌ 取消任务失败: %v\n", err)
	} else {
		fmt.Println("✅ 任务已取消")
	}

	fmt.Println("\n3. 演示工作流程管理器")
	fmt.Println("====================")

	// 直接使用工作流程管理器
	workflowManager := controller.NewWorkflowManager(ctrl)
	job3 := model.NewMergeJob("main3.pdf", []string{"add4.pdf", "add5.pdf"}, "output3.pdf")

	ctx := context.Background()
	fmt.Println("执行工作流程...")
	err = workflowManager.ExecuteWorkflow(ctx, job3)
	if err != nil {
		fmt.Printf("❌ 工作流程执行失败: %v\n", err)
	} else {
		fmt.Println("✅ 工作流程执行完成")
	}

	fmt.Println("\n4. 演示流式合并器")
	fmt.Println("================")

	// 创建流式合并器
	streamingMerger := controller.NewStreamingMerger(ctrl)
	job4 := model.NewMergeJob("main4.pdf", []string{"add6.pdf", "add7.pdf"}, "output4.pdf")

	fmt.Println("执行流式合并...")
	err = streamingMerger.MergeStreaming(ctx, job4, nil)
	if err != nil {
		fmt.Printf("❌ 流式合并失败: %v\n", err)
	} else {
		fmt.Println("✅ 流式合并完成")
	}

	fmt.Println("\n5. 演示批处理器")
	fmt.Println("==============")

	// 创建批处理器
	batchProcessor := controller.NewBatchProcessor(streamingMerger)
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf", "file4.pdf", "file5.pdf"}

	fmt.Printf("批量处理 %d 个文件...\n", len(files))
	err = batchProcessor.ProcessBatch(ctx, files, "batch_output.pdf", nil)
	if err != nil {
		fmt.Printf("❌ 批处理失败: %v\n", err)
	} else {
		fmt.Println("✅ 批处理完成")
	}

	fmt.Println("\n6. 演示取消管理器")
	fmt.Println("================")

	// 创建取消管理器
	cancelManager := controller.NewCancellationManager(ctrl)

	// 创建测试上下文
	_, cancel := context.WithCancel(context.Background())
	jobID := "demo-job-123"

	// 注册取消操作
	cancelManager.RegisterCancellation(jobID, cancel)

	// 添加清理任务
	cleanupExecuted := false
	cleanupTask := controller.NewResourceCleanupTask("demo-cleanup", func() error {
		cleanupExecuted = true
		fmt.Println("🧹 执行清理任务")
		return nil
	})
	cancelManager.AddCleanupTask(cleanupTask)

	// 执行优雅取消
	fmt.Println("执行优雅取消...")
	err = cancelManager.GracefulCancellation(jobID, 1*time.Second)
	if err != nil {
		fmt.Printf("❌ 优雅取消失败: %v\n", err)
	} else {
		fmt.Println("✅ 优雅取消完成")
	}

	if cleanupExecuted {
		fmt.Println("✅ 清理任务已执行")
	} else {
		fmt.Println("❌ 清理任务未执行")
	}

	fmt.Println("\n7. 演示内存监控器")
	fmt.Println("================")

	// 创建内存监控器
	memoryMonitor := controller.NewMemoryMonitor(100 * 1024 * 1024) // 100MB

	fmt.Println("启动内存监控...")
	memoryMonitor.Start()

	// 检查内存状态
	isLow := memoryMonitor.IsMemoryLow()
	fmt.Printf("内存是否不足: %v\n", isLow)

	// 停止监控
	memoryMonitor.Stop()
	fmt.Println("✅ 内存监控已停止")

	fmt.Println("\n演示完成！")
}

func setupCallbacks(ctrl *controller.Controller) {
	// 设置进度回调
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		percentage := int(progress * 100)
		fmt.Printf("📊 进度: %d%% - %s: %s\n", percentage, status, detail)
	})

	// 设置错误回调
	ctrl.SetErrorCallback(func(err error) {
		fmt.Printf("❌ 错误: %v\n", err)
	})

	// 设置完成回调
	ctrl.SetCompletionCallback(func(outputPath string) {
		fmt.Printf("🎉 完成: 输出文件 %s\n", outputPath)
	})
}

func waitForJobCompletion(ctrl *controller.Controller, timeout time.Duration) {
	start := time.Now()
	for {
		if !ctrl.IsJobRunning() {
			fmt.Println("✅ 任务已完成")
			return
		}

		if time.Since(start) > timeout {
			fmt.Println("⏰ 等待超时")
			return
		}

		time.Sleep(100 * time.Millisecond)
	}
}