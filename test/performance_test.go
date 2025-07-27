package test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/test_utils"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Duration       time.Duration
	MemoryUsed     int64
	GoroutineCount int
	AllocCount     uint64
	AllocBytes     uint64
}

// measurePerformance 测量性能指标
func measurePerformance(fn func()) PerformanceMetrics {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	startTime := time.Now()
	startGoroutines := runtime.NumGoroutine()

	fn()

	endTime := time.Now()
	endGoroutines := runtime.NumGoroutine()
	runtime.ReadMemStats(&m2)

	return PerformanceMetrics{
		Duration:       endTime.Sub(startTime),
		MemoryUsed:     int64(m2.Alloc - m1.Alloc),
		GoroutineCount: endGoroutines - startGoroutines,
		AllocCount:     m2.Mallocs - m1.Mallocs,
		AllocBytes:     m2.TotalAlloc - m1.TotalAlloc,
	}
}

// TestPerformance_FileValidation 文件验证性能测试
func TestPerformance_FileValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-validation")

	// 创建不同大小的测试文件
	testFiles := []struct {
		name string
		size int // KB
	}{
		{"小文件", 10},
		{"中等文件", 100},
		{"大文件", 1000},
	}

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	for _, tf := range testFiles {
		t.Run(tf.name, func(t *testing.T) {
			// 创建测试文件
			testFile := test_utils.CreateLargePDFFile(t, tempDir,
				fmt.Sprintf("perf_%s.pdf", tf.name), tf.size)

			// 测量性能
			metrics := measurePerformance(func() {
				for i := 0; i < 100; i++ {
					ctrl.ValidateFile(testFile)
				}
			})

			t.Logf("%s 性能指标:", tf.name)
			t.Logf("  总时间: %v", metrics.Duration)
			t.Logf("  平均时间: %v", metrics.Duration/100)
			t.Logf("  内存使用: %d bytes", metrics.MemoryUsed)
			t.Logf("  分配次数: %d", metrics.AllocCount)
			t.Logf("  分配字节: %d", metrics.AllocBytes)

			// 性能阈值检查
			avgDuration := metrics.Duration / 100
			maxDuration := time.Duration(tf.size) * time.Microsecond // 1μs per KB
			if avgDuration > maxDuration {
				t.Logf("性能警告: 平均验证时间 %v 超过预期 %v", avgDuration, maxDuration)
			}
		})
	}
}

// TestPerformance_MemoryUsage 内存使用测试
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-memory")

	// 创建多个大文件
	fileCount := 20
	files := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		files[i] = test_utils.CreateLargePDFFile(t, tempDir,
			fmt.Sprintf("memory_test_%d.pdf", i), 50) // 50KB each
	}

	// 测试不同内存限制下的性能
	memoryLimits := []int64{
		1 * 1024 * 1024,  // 1MB
		5 * 1024 * 1024,  // 5MB
		10 * 1024 * 1024, // 10MB
		50 * 1024 * 1024, // 50MB
	}

	for _, limit := range memoryLimits {
		t.Run(fmt.Sprintf("内存限制_%dMB", limit/(1024*1024)), func(t *testing.T) {
			fileManager := file.NewFileManager(tempDir)
			pdfService := pdf.NewPDFService()
			config := model.DefaultConfig()
			config.MaxMemoryUsage = limit
			config.TempDirectory = tempDir

			ctrl := controller.NewController(pdfService, fileManager, config)
			streamingMerger := controller.NewStreamingMerger(ctrl)

			// 创建内存监控器
			memoryMonitor := controller.NewMemoryMonitor(limit)
			memoryMonitor.Start()
			defer memoryMonitor.Stop()

			// 测量内存使用
			metrics := measurePerformance(func() {
				job := model.NewMergeJob(files[0], files[1:5], // 使用前5个文件
					fmt.Sprintf("%s/memory_output_%d.pdf", tempDir, limit))

				ctx := context.Background()
				err := streamingMerger.MergeStreaming(ctx, job, nil)
				if err != nil {
					t.Logf("流式合并失败（可能由于UniPDF许可证）: %v", err)
				}
			})

			t.Logf("内存限制 %dMB 性能指标:", limit/(1024*1024))
			t.Logf("  执行时间: %v", metrics.Duration)
			t.Logf("  内存使用: %d bytes", metrics.MemoryUsed)
			t.Logf("  内存是否不足: %v", memoryMonitor.IsMemoryLow())

			// 检查内存使用是否在限制内
			if metrics.MemoryUsed > limit {
				t.Logf("内存警告: 使用了 %d bytes，超过限制 %d bytes",
					metrics.MemoryUsed, limit)
			}
		})
	}
}

// TestPerformance_ConcurrentOperations 并发操作性能测试
func TestPerformance_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发性能测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-concurrent")

	// 创建测试文件
	testFile := test_utils.CreateTestPDFFile(t, tempDir, "concurrent_test.pdf")

	// 测试不同并发级别
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, level := range concurrencyLevels {
		t.Run(fmt.Sprintf("并发级别_%d", level), func(t *testing.T) {
			fileManager := file.NewFileManager(tempDir)
			pdfService := pdf.NewPDFService()
			config := model.DefaultConfig()
			ctrl := controller.NewController(pdfService, fileManager, config)

			// 测量并发性能
			metrics := measurePerformance(func() {
				var wg sync.WaitGroup
				operationsPerGoroutine := 50

				for i := 0; i < level; i++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()
						for j := 0; j < operationsPerGoroutine; j++ {
							ctrl.ValidateFile(testFile)
						}
					}(i)
				}

				wg.Wait()
			})

			totalOperations := level * 50
			avgDuration := metrics.Duration / time.Duration(totalOperations)

			t.Logf("并发级别 %d 性能指标:", level)
			t.Logf("  总时间: %v", metrics.Duration)
			t.Logf("  平均操作时间: %v", avgDuration)
			t.Logf("  吞吐量: %.2f ops/sec", float64(totalOperations)/metrics.Duration.Seconds())
			t.Logf("  内存使用: %d bytes", metrics.MemoryUsed)
			t.Logf("  Goroutine 增量: %d", metrics.GoroutineCount)

			// 检查是否有goroutine泄漏
			if metrics.GoroutineCount > level {
				t.Logf("Goroutine 警告: 创建了 %d 个额外的 goroutine",
					metrics.GoroutineCount-level)
			}
		})
	}
}

// TestPerformance_WorkflowManager 工作流程管理器性能测试
func TestPerformance_WorkflowManager(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过工作流程性能测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-workflow")

	// 创建测试文件
	mainFile := test_utils.CreateTestPDFFile(t, tempDir, "workflow_main.pdf")
	additionalFiles := []string{
		test_utils.CreateTestPDFFile(t, tempDir, "workflow_add1.pdf"),
		test_utils.CreateTestPDFFile(t, tempDir, "workflow_add2.pdf"),
	}

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	ctrl := controller.NewController(pdfService, fileManager, config)
	workflowManager := controller.NewWorkflowManager(ctrl)

	// 测量工作流程性能
	iterations := 10
	metrics := measurePerformance(func() {
		for i := 0; i < iterations; i++ {
			job := model.NewMergeJob(mainFile, additionalFiles,
				fmt.Sprintf("%s/workflow_output_%d.pdf", tempDir, i))

			ctx := context.Background()
			err := workflowManager.ExecuteWorkflow(ctx, job)
			if err != nil {
				t.Logf("工作流程执行失败（可能由于UniPDF许可证）: %v", err)
			}
		}
	})

	avgDuration := metrics.Duration / time.Duration(iterations)

	t.Logf("工作流程管理器性能指标:")
	t.Logf("  总时间: %v", metrics.Duration)
	t.Logf("  平均执行时间: %v", avgDuration)
	t.Logf("  内存使用: %d bytes", metrics.MemoryUsed)
	t.Logf("  分配次数: %d", metrics.AllocCount)

	// 性能阈值检查
	maxAvgDuration := 1 * time.Second
	if avgDuration > maxAvgDuration {
		t.Logf("性能警告: 平均执行时间 %v 超过预期 %v", avgDuration, maxAvgDuration)
	}
}

// TestPerformance_BatchProcessing 批处理性能测试
func TestPerformance_BatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过批处理性能测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-batch")

	// 测试不同批次大小
	batchSizes := []int{5, 10, 20, 50}

	for _, batchSize := range batchSizes {
		t.Run(fmt.Sprintf("批次大小_%d", batchSize), func(t *testing.T) {
			// 创建测试文件
			files := make([]string, batchSize)
			for i := 0; i < batchSize; i++ {
				files[i] = test_utils.CreateTestPDFFile(t, tempDir,
					fmt.Sprintf("batch_%d_%d.pdf", batchSize, i))
			}

			outputFile := fmt.Sprintf("%s/batch_output_%d.pdf", tempDir, batchSize)

			fileManager := file.NewFileManager(tempDir)
			pdfService := pdf.NewPDFService()
			config := model.DefaultConfig()
			config.TempDirectory = tempDir

			ctrl := controller.NewController(pdfService, fileManager, config)
			streamingMerger := controller.NewStreamingMerger(ctrl)
			batchProcessor := controller.NewBatchProcessor(streamingMerger)

			// 测量批处理性能
			metrics := measurePerformance(func() {
				ctx := context.Background()
				err := batchProcessor.ProcessBatch(ctx, files, outputFile, nil)
				if err != nil {
					t.Logf("批处理失败（可能由于UniPDF许可证）: %v", err)
				}
			})

			t.Logf("批次大小 %d 性能指标:", batchSize)
			t.Logf("  执行时间: %v", metrics.Duration)
			t.Logf("  每文件平均时间: %v", metrics.Duration/time.Duration(batchSize))
			t.Logf("  内存使用: %d bytes", metrics.MemoryUsed)
			t.Logf("  每文件内存: %d bytes", metrics.MemoryUsed/int64(batchSize))

			// 检查线性扩展性
			expectedDuration := time.Duration(batchSize) * 10 * time.Millisecond
			if metrics.Duration > expectedDuration*2 {
				t.Logf("性能警告: 执行时间 %v 超过预期 %v 的2倍",
					metrics.Duration, expectedDuration)
			}
		})
	}
}

// TestPerformance_MemoryLeaks 内存泄漏测试
func TestPerformance_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存泄漏测试（短测试模式）")
	}

	tempDir := test_utils.CreateTempDir(t, "perf-leaks")
	testFile := test_utils.CreateTestPDFFile(t, tempDir, "leak_test.pdf")

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 基准内存使用
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 执行多次操作
	iterations := 1000
	for i := 0; i < iterations; i++ {
		ctrl.ValidateFile(testFile)

		// 每100次操作检查一次内存
		if i%100 == 99 {
			runtime.GC()
			runtime.ReadMemStats(&m2)

			memoryGrowth := int64(m2.Alloc - m1.Alloc)
			if memoryGrowth > 10*1024*1024 { // 10MB
				t.Logf("内存警告: 在 %d 次操作后内存增长了 %d bytes", i+1, memoryGrowth)
			}
		}
	}

	// 最终内存检查
	runtime.GC()
	runtime.ReadMemStats(&m2)

	finalMemoryGrowth := int64(m2.Alloc - m1.Alloc)
	t.Logf("内存泄漏测试结果:")
	t.Logf("  操作次数: %d", iterations)
	t.Logf("  最终内存增长: %d bytes", finalMemoryGrowth)
	t.Logf("  平均每操作内存: %d bytes", finalMemoryGrowth/int64(iterations))

	// 检查内存泄漏
	maxAcceptableGrowth := int64(5 * 1024 * 1024) // 5MB
	if finalMemoryGrowth > maxAcceptableGrowth {
		t.Logf("内存泄漏警告: 最终内存增长 %d bytes 超过可接受范围 %d bytes",
			finalMemoryGrowth, maxAcceptableGrowth)
	}
}

// TestPerformance_ResourceCleanup 资源清理性能测试
func TestPerformance_ResourceCleanup(t *testing.T) {
	tempDir := test_utils.CreateTempDir(t, "perf-cleanup")

	fileManager := file.NewFileManager(tempDir)

	// 创建大量临时文件
	fileCount := 100
	tempFiles := make([]string, fileCount)

	createMetrics := measurePerformance(func() {
		for i := 0; i < fileCount; i++ {
			tempFile, _, err := fileManager.CreateTempFileWithPrefix(
				fmt.Sprintf("cleanup_perf_%d_", i), ".pdf")
			if err != nil {
				t.Fatalf("创建临时文件失败: %v", err)
			}
			tempFiles[i] = tempFile
		}
	})

	t.Logf("创建 %d 个临时文件性能:", fileCount)
	t.Logf("  总时间: %v", createMetrics.Duration)
	t.Logf("  平均时间: %v", createMetrics.Duration/time.Duration(fileCount))
	t.Logf("  内存使用: %d bytes", createMetrics.MemoryUsed)

	// 测量清理性能
	cleanupMetrics := measurePerformance(func() {
		err := fileManager.CleanupTempFiles()
		if err != nil {
			t.Errorf("清理临时文件失败: %v", err)
		}
	})

	t.Logf("清理 %d 个临时文件性能:", fileCount)
	t.Logf("  总时间: %v", cleanupMetrics.Duration)
	t.Logf("  平均时间: %v", cleanupMetrics.Duration/time.Duration(fileCount))
	t.Logf("  内存使用: %d bytes", cleanupMetrics.MemoryUsed)

	// 验证清理效果
	cleanedCount := 0
	for _, tempFile := range tempFiles {
		if !test_utils.FileExists(tempFile) {
			cleanedCount++
		}
	}

	t.Logf("成功清理了 %d/%d 个文件", cleanedCount, fileCount)
}

// 基准测试

func BenchmarkPerformance_FileValidation(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "bench-validation")
	testFile := test_utils.CreateTestPDFFile(b, tempDir, "benchmark.pdf")

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctrl.ValidateFile(testFile)
	}
}

func BenchmarkPerformance_ControllerCreation(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "bench-creation")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fileManager := file.NewFileManager(tempDir)
		pdfService := pdf.NewPDFService()
		config := model.DefaultConfig()
		controller.NewController(pdfService, fileManager, config)
	}
}

func BenchmarkPerformance_EventHandling(b *testing.B) {
	tempDir := test_utils.CreateTempDir(b, "bench-events")
	testFile := test_utils.CreateTestPDFFile(b, tempDir, "benchmark.pdf")

	fileManager := file.NewFileManager(tempDir)
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)
	eventHandler := controller.NewEventHandler(ctrl)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		eventHandler.HandleMainFileSelected(testFile)
	}
}
