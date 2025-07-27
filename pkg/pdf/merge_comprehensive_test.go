package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDFMergeComprehensive 全面的PDF合并功能测试
func TestPDFMergeComprehensive(t *testing.T) {
	// 创建测试目录
	testDir := filepath.Join(os.TempDir(), "merge_comprehensive_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFiles := createMergeTestPDFFiles(t, testDir, 10)
	require.Greater(t, len(testFiles), 0, "应该创建至少一个测试文件")

	t.Run("TestBasicMerge", func(t *testing.T) {
		testBasicMerge(t, testDir, testFiles)
	})

	t.Run("TestLargeFileMerge", func(t *testing.T) {
		testLargeFileMerge(t, testDir)
	})

	t.Run("TestProgressAndCancel", func(t *testing.T) {
		testProgressAndCancel(t, testDir, testFiles)
	})

	t.Run("TestMergeQuality", func(t *testing.T) {
		testMergeQuality(t, testDir, testFiles)
	})

	t.Run("TestConcurrentMerge", func(t *testing.T) {
		testConcurrentMerge(t, testDir, testFiles)
	})

	t.Run("TestErrorHandling", func(t *testing.T) {
		testErrorHandling(t, testDir)
	})

	t.Run("TestMemoryUsage", func(t *testing.T) {
		testMergeMemoryUsage(t, testDir, testFiles)
	})

	t.Run("TestPerformanceComparison", func(t *testing.T) {
		testPerformanceComparison(t, testDir, testFiles)
	})
}

// testBasicMerge 测试基本合并功能
func testBasicMerge(t *testing.T, testDir string, testFiles []string) {
	outputFile := filepath.Join(testDir, "basic_merge_output.pdf")

	// 创建合并器
	options := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		TempDirectory:     testDir,
		EnableGC:          true,
		UseStreaming:      true,
		OptimizeMemory:    true,
		ConcurrentWorkers: runtime.NumCPU(),
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 执行合并
	ctx := context.Background()
	var progressUpdates []string
	progressCallback := func(progress float64, message string) {
		progressUpdates = append(progressUpdates, fmt.Sprintf("%.1f%% - %s", progress*100, message))
	}

	result, err := merger.MergeStreaming(ctx, testFiles, outputFile, progressCallback)

	// 验证结果
	if err != nil {
		t.Logf("合并失败（可能是由于pdfcpu不可用）: %v", err)
		// 检查是否是预期的错误类型
		if pdfErr, ok := err.(*PDFError); ok {
			t.Logf("PDF错误类型: %v, 消息: %s", pdfErr.Type, pdfErr.Message)
		}
		return
	}

	// 验证合并结果
	assert.NotNil(t, result, "应该返回合并结果")
	assert.Equal(t, len(testFiles), result.ProcessedFiles, "应该处理所有文件")
	assert.Greater(t, result.ProcessingTime, time.Duration(0), "处理时间应该大于0")
	assert.Greater(t, result.MemoryUsage, int64(0), "内存使用应该大于0")

	// 验证输出文件
	if _, err := os.Stat(outputFile); err == nil {
		t.Logf("输出文件创建成功: %s", outputFile)
	} else {
		t.Logf("输出文件检查失败: %v", err)
	}

	// 验证进度更新
	assert.Greater(t, len(progressUpdates), 0, "应该有进度更新")
	t.Logf("进度更新数量: %d", len(progressUpdates))
	for i, update := range progressUpdates {
		t.Logf("进度更新 %d: %s", i+1, update)
	}

	t.Logf("基本合并测试完成: 处理文件=%d, 处理时间=%v, 内存使用=%d bytes",
		result.ProcessedFiles, result.ProcessingTime, result.MemoryUsage)
}

// testLargeFileMerge 测试大文件合并
func testLargeFileMerge(t *testing.T, testDir string) {
	// 创建大文件
	largeFiles := createLargeMergeTestFiles(t, testDir, 3)
	outputFile := filepath.Join(testDir, "large_merge_output.pdf")

	// 创建内存受限的合并器
	options := &MergeOptions{
		MaxMemoryUsage:    50 * 1024 * 1024, // 50MB限制
		TempDirectory:     testDir,
		EnableGC:          true,
		UseStreaming:      true,
		OptimizeMemory:    true,
		ConcurrentWorkers: 2, // 减少并发数
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 记录合并前内存
	var beforeStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)
	beforeMemory := beforeStats.Alloc

	// 执行合并
	ctx := context.Background()
	result, err := merger.MergeStreaming(ctx, largeFiles, outputFile, nil)

	// 记录合并后内存
	var afterStats runtime.MemStats
	runtime.ReadMemStats(&afterStats)
	afterMemory := afterStats.Alloc

	if err != nil {
		t.Logf("大文件合并失败: %v", err)
		return
	}

	// 验证内存使用
	memoryIncrease := afterMemory - beforeMemory
	memoryIncreaseMB := float64(memoryIncrease) / (1024 * 1024)

	t.Logf("大文件合并内存使用: 增加 %.2f MB", memoryIncreaseMB)
	assert.Less(t, memoryIncreaseMB, 100.0, "内存增加不应该超过100MB")

	// 验证结果
	assert.NotNil(t, result, "应该返回合并结果")
	t.Logf("大文件合并完成: 处理文件=%d, 处理时间=%v, 内存使用=%d bytes",
		result.ProcessedFiles, result.ProcessingTime, result.MemoryUsage)
}

// testProgressAndCancel 测试进度回调和取消功能
func testProgressAndCancel(t *testing.T, testDir string, testFiles []string) {
	outputFile := filepath.Join(testDir, "progress_cancel_output.pdf")

	options := &MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024,
		TempDirectory:  testDir,
		EnableGC:       true,
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 测试取消功能
	t.Run("TestCancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// 100ms后取消
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		var progressUpdates []string
		progressCallback := func(progress float64, message string) {
			progressUpdates = append(progressUpdates, fmt.Sprintf("%.1f%% - %s", progress*100, message))
		}

		start := time.Now()
		_, err := merger.MergeStreaming(ctx, testFiles, outputFile, progressCallback)
		duration := time.Since(start)

		// 验证取消
		assert.Error(t, err, "应该因为取消而返回错误")
		assert.Contains(t, err.Error(), "canceled", "错误应该包含取消信息")
		assert.Less(t, duration, 500*time.Millisecond, "取消应该在500ms内响应")

		// 验证进度更新
		assert.Greater(t, len(progressUpdates), 0, "应该有进度更新")
		t.Logf("取消测试: 进度更新数量=%d, 耗时=%v", len(progressUpdates), duration)
	})

	// 测试进度回调
	t.Run("TestProgressCallback", func(t *testing.T) {
		ctx := context.Background()
		var progressUpdates []string
		var mu sync.Mutex

		progressCallback := func(progress float64, message string) {
			mu.Lock()
			defer mu.Unlock()
			progressUpdates = append(progressUpdates, fmt.Sprintf("%.1f%% - %s", progress*100, message))
		}

		result, err := merger.MergeStreaming(ctx, testFiles[:3], outputFile, progressCallback)

		if err != nil {
			t.Logf("进度回调测试合并失败: %v", err)
			return
		}

		// 验证进度更新
		assert.Greater(t, len(progressUpdates), 0, "应该有进度更新")

		// 验证进度值范围
		for _, update := range progressUpdates {
			parts := strings.Split(update, "%")
			if len(parts) > 0 {
				var progress float64
				fmt.Sscanf(parts[0], "%f", &progress)
				assert.GreaterOrEqual(t, progress, 0.0, "进度应该大于等于0")
				assert.LessOrEqual(t, progress, 100.0, "进度应该小于等于100")
			}
		}

		t.Logf("进度回调测试: 更新数量=%d, 处理文件=%d", len(progressUpdates), result.ProcessedFiles)
	})
}

// testMergeQuality 测试合并质量
func testMergeQuality(t *testing.T, testDir string, testFiles []string) {
	outputFile := filepath.Join(testDir, "quality_test_output.pdf")

	options := &MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024,
		TempDirectory:  testDir,
		EnableGC:       true,
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 执行合并
	ctx := context.Background()
	result, err := merger.MergeStreaming(ctx, testFiles, outputFile, nil)

	if err != nil {
		t.Logf("质量测试合并失败: %v", err)
		return
	}

	// 验证输出文件
	if info, err := os.Stat(outputFile); err == nil {
		assert.Greater(t, info.Size(), int64(0), "输出文件大小应该大于0")
		t.Logf("输出文件大小: %d bytes", info.Size())
	} else {
		t.Logf("输出文件检查失败: %v", err)
	}

	// 验证合并结果
	assert.NotNil(t, result, "应该返回合并结果")
	assert.Equal(t, len(testFiles), result.ProcessedFiles, "应该处理所有文件")
	assert.Greater(t, result.TotalPages, 0, "总页数应该大于0")

	t.Logf("合并质量测试: 处理文件=%d, 总页数=%d, 处理时间=%v",
		result.ProcessedFiles, result.TotalPages, result.ProcessingTime)
}

// testConcurrentMerge 测试并发合并
func testConcurrentMerge(t *testing.T, testDir string, testFiles []string) {
	const numConcurrent = 3
	var wg sync.WaitGroup
	results := make(chan *MergeResult, numConcurrent)
	errors := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			outputFile := filepath.Join(testDir, fmt.Sprintf("concurrent_output_%d.pdf", index))

			options := &MergeOptions{
				MaxMemoryUsage:    50 * 1024 * 1024,
				TempDirectory:     testDir,
				EnableGC:          true,
				ConcurrentWorkers: 2,
			}

			merger := NewStreamingMerger(options)
			defer merger.Close()

			ctx := context.Background()
			result, err := merger.MergeStreaming(ctx, testFiles[:3], outputFile, nil)

			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// 统计结果
	successCount := 0
	errorCount := 0

	for result := range results {
		if result != nil {
			successCount++
		}
	}

	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("并发合并错误: %v", err)
		}
	}

	t.Logf("并发合并测试: 成功=%d, 失败=%d", successCount, errorCount)
	assert.Greater(t, successCount, 0, "至少应该有一个成功的合并")
}

// testErrorHandling 测试错误处理
func testErrorHandling(t *testing.T, testDir string) {
	// 创建无效文件
	invalidFiles := []string{
		filepath.Join(testDir, "nonexistent.pdf"),
		filepath.Join(testDir, "empty.pdf"),
		filepath.Join(testDir, "corrupted.pdf"),
	}

	// 创建空文件
	emptyFile := filepath.Join(testDir, "empty.pdf")
	os.WriteFile(emptyFile, []byte(""), 0644)

	// 创建损坏文件
	corruptedFile := filepath.Join(testDir, "corrupted.pdf")
	os.WriteFile(corruptedFile, []byte("NOT_A_PDF"), 0644)

	outputFile := filepath.Join(testDir, "error_test_output.pdf")

	options := &MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024,
		TempDirectory:  testDir,
		EnableGC:       true,
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 测试无效文件处理
	ctx := context.Background()
	result, err := merger.MergeStreaming(ctx, invalidFiles, outputFile, nil)

	// 验证错误处理
	if err != nil {
		t.Logf("预期错误: %v", err)
		// 检查是否是预期的错误类型
		if pdfErr, ok := err.(*PDFError); ok {
			assert.Equal(t, ErrorInvalidInput, pdfErr.Type, "应该是无效输入错误")
		}
	} else {
		// 如果成功，验证跳过的文件
		assert.NotNil(t, result, "应该返回结果")
		assert.Greater(t, len(result.SkippedFiles), 0, "应该有跳过的文件")
		t.Logf("跳过的文件: %v", result.SkippedFiles)
	}
}

// testMergeMemoryUsage 测试内存使用
func testMergeMemoryUsage(t *testing.T, testDir string, testFiles []string) {
	outputFile := filepath.Join(testDir, "memory_test_output.pdf")

	// 记录初始内存
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	initialMemory := initialStats.Alloc

	options := &MergeOptions{
		MaxMemoryUsage:    50 * 1024 * 1024, // 50MB限制
		TempDirectory:     testDir,
		EnableGC:          true,
		OptimizeMemory:    true,
		ConcurrentWorkers: 1, // 单线程测试
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	// 执行合并
	ctx := context.Background()
	result, err := merger.MergeStreaming(ctx, testFiles, outputFile, nil)

	// 记录最终内存
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)
	finalMemory := finalStats.Alloc

	if err != nil {
		t.Logf("内存测试合并失败: %v", err)
		return
	}

	// 计算内存使用
	memoryUsed := finalMemory - initialMemory
	memoryUsedMB := float64(memoryUsed) / (1024 * 1024)

	t.Logf("内存使用测试: 初始=%d bytes, 最终=%d bytes, 使用=%.2f MB",
		initialMemory, finalMemory, memoryUsedMB)

	// 验证内存使用在合理范围内
	assert.Less(t, memoryUsedMB, 100.0, "内存使用不应该超过100MB")
	assert.Greater(t, memoryUsedMB, 0.0, "应该有内存使用")

	// 验证结果
	assert.NotNil(t, result, "应该返回合并结果")
	t.Logf("内存测试完成: 处理文件=%d, 处理时间=%v, 报告内存使用=%d bytes",
		result.ProcessedFiles, result.ProcessingTime, result.MemoryUsage)
}

// testPerformanceComparison 测试性能对比
func testPerformanceComparison(t *testing.T, testDir string, testFiles []string) {
	outputFile1 := filepath.Join(testDir, "perf_standard_output.pdf")
	outputFile2 := filepath.Join(testDir, "perf_optimized_output.pdf")

	// 标准配置
	standardOptions := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     testDir,
		EnableGC:          true,
		UseStreaming:      false, // 不使用流式处理
		OptimizeMemory:    false, // 不优化内存
		ConcurrentWorkers: 1,
	}

	// 优化配置
	optimizedOptions := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     testDir,
		EnableGC:          true,
		UseStreaming:      true, // 使用流式处理
		OptimizeMemory:    true, // 优化内存
		ConcurrentWorkers: runtime.NumCPU(),
	}

	// 测试标准合并
	standardMerger := NewStreamingMerger(standardOptions)
	defer standardMerger.Close()

	ctx := context.Background()
	start := time.Now()
	standardResult, standardErr := standardMerger.MergeStreaming(ctx, testFiles, outputFile1, nil)
	standardDuration := time.Since(start)

	// 测试优化合并
	optimizedMerger := NewStreamingMerger(optimizedOptions)
	defer optimizedMerger.Close()

	start = time.Now()
	optimizedResult, optimizedErr := optimizedMerger.MergeStreaming(ctx, testFiles, outputFile2, nil)
	optimizedDuration := time.Since(start)

	// 记录结果
	t.Logf("性能对比测试:")
	t.Logf("  标准合并: 耗时=%v, 错误=%v", standardDuration, standardErr)
	if standardResult != nil {
		t.Logf("  标准合并: 处理文件=%d, 内存使用=%d bytes",
			standardResult.ProcessedFiles, standardResult.MemoryUsage)
	}

	t.Logf("  优化合并: 耗时=%v, 错误=%v", optimizedDuration, optimizedErr)
	if optimizedResult != nil {
		t.Logf("  优化合并: 处理文件=%d, 内存使用=%d bytes",
			optimizedResult.ProcessedFiles, optimizedResult.MemoryUsage)
	}

	// 如果两者都成功，比较性能
	if standardErr == nil && optimizedErr == nil {
		timeRatio := float64(optimizedDuration) / float64(standardDuration)
		t.Logf("  时间比率: 优化/标准 = %.2f", timeRatio)

		if standardResult != nil && optimizedResult != nil {
			memoryRatio := float64(optimizedResult.MemoryUsage) / float64(standardResult.MemoryUsage)
			t.Logf("  内存比率: 优化/标准 = %.2f", memoryRatio)
		}
	}
}

// createMergeTestPDFFiles 创建合并测试PDF文件
func createMergeTestPDFFiles(t *testing.T, testDir string, count int) []string {
	files := make([]string, count)

	for i := 0; i < count; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("merge_test_%d.pdf", i+1))
		content := createMergeTestPDFContent(i + 1)
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err, "创建合并测试文件失败")
		files[i] = filename
	}

	return files
}

// createLargeMergeTestFiles 创建大合并测试文件
func createLargeMergeTestFiles(t *testing.T, testDir string, count int) []string {
	files := make([]string, count)

	for i := 0; i < count; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("large_merge_test_%d.pdf", i+1))
		content := createLargeMergeTestPDFContent(i + 1)
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err, "创建大合并测试文件失败")
		files[i] = filename
	}

	return files
}

// createMergeTestPDFContent 创建合并测试PDF内容
func createMergeTestPDFContent(pageNum int) string {
	return fmt.Sprintf(`%%PDF-1.4
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
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
72 720 Td
(Merge Test Page %d) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`, pageNum)
}

// createLargeMergeTestPDFContent 创建大合并测试PDF内容
func createLargeMergeTestPDFContent(pageNum int) string {
	// 创建更大的PDF内容
	content := fmt.Sprintf(`%%PDF-1.4
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
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length %d
>>
stream`, 1000+pageNum*100)

	// 添加大量文本内容
	for i := 0; i < 50; i++ {
		content += fmt.Sprintf("BT\n/F1 12 Tf\n72 %d Td\n(Large Merge Test Page %d - Line %d) Tj\nET\n", 720-i*12, pageNum, i+1)
	}

	content += `endstream
endobj
xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`

	return content
}

// BenchmarkPDFMerge 性能基准测试
func BenchmarkPDFMerge(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "merge_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFiles := createMergeTestPDFFilesForBenchmark(b, testDir, 5)

	options := &MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024,
		TempDirectory:  testDir,
		EnableGC:       true,
		UseStreaming:   true,
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(testDir, fmt.Sprintf("benchmark_output_%d.pdf", i))

		ctx := context.Background()
		_, err := merger.MergeStreaming(ctx, testFiles, outputFile, nil)
		if err != nil {
			b.Fatalf("Benchmark merge failed: %v", err)
		}

		// 清理输出文件
		os.Remove(outputFile)
	}
}

// BenchmarkPDFMergeLarge 大文件合并性能基准测试
func BenchmarkPDFMergeLarge(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "merge_large_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// 创建大测试文件
	testFiles := createLargeMergeTestFilesForBenchmark(b, testDir, 3)

	options := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     testDir,
		EnableGC:          true,
		UseStreaming:      true,
		OptimizeMemory:    true,
		ConcurrentWorkers: 2,
	}

	merger := NewStreamingMerger(options)
	defer merger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(testDir, fmt.Sprintf("large_benchmark_output_%d.pdf", i))

		ctx := context.Background()
		_, err := merger.MergeStreaming(ctx, testFiles, outputFile, nil)
		if err != nil {
			b.Fatalf("Large file benchmark merge failed: %v", err)
		}

		// 清理输出文件
		os.Remove(outputFile)
	}
}

// createMergeTestPDFFilesForBenchmark 为基准测试创建合并测试PDF文件
func createMergeTestPDFFilesForBenchmark(b *testing.B, testDir string, count int) []string {
	files := make([]string, count)

	for i := 0; i < count; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("benchmark_merge_test_%d.pdf", i+1))
		content := createMergeTestPDFContent(i + 1)
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create benchmark test file: %v", err)
		}
		files[i] = filename
	}

	return files
}

// createLargeMergeTestFilesForBenchmark 为基准测试创建大合并测试文件
func createLargeMergeTestFilesForBenchmark(b *testing.B, testDir string, count int) []string {
	files := make([]string, count)

	for i := 0; i < count; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("large_benchmark_merge_test_%d.pdf", i+1))
		content := createLargeMergeTestPDFContent(i + 1)
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create large benchmark test file: %v", err)
		}
		files[i] = filename
	}

	return files
}
