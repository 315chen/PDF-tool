package pdf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// PerformanceTestSuite 性能测试套件
type PerformanceTestSuite struct {
	TestDir   string
	Results   []PerformanceTestResult
	Mutex     sync.RWMutex
	StartTime time.Time
	EndTime   time.Time
}

// PerformanceTestResult 性能测试结果
type PerformanceTestResult struct {
	TestName      string
	FileCount     int
	TotalPages    int
	TotalSize     int64
	Duration      time.Duration
	PeakMemory    uint64
	AvgMemory     uint64
	Success       bool
	Error         string
	TestType      string
	Configuration string
}

// NewPerformanceTestSuite 创建性能测试套件
func NewPerformanceTestSuite(testDir string) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		TestDir: testDir,
		Results: make([]PerformanceTestResult, 0),
	}
}

// RunComprehensivePerformanceTests 执行全面性能测试
func (pts *PerformanceTestSuite) RunComprehensivePerformanceTests(t *testing.T) {
	pts.StartTime = time.Now()
	defer func() {
		pts.EndTime = time.Now()
		pts.GenerateReport(t)
	}()

	// 1. 基础性能测试
	t.Run("BasicPerformance", pts.testBasicPerformance)

	// 2. 大规模文件测试
	t.Run("LargeScaleFiles", pts.testLargeScaleFiles)

	// 3. 内存压力测试
	t.Run("MemoryStress", pts.testMemoryStress)

	// 4. 并发性能测试
	t.Run("ConcurrencyPerformance", pts.testConcurrencyPerformance)

	// 5. 文件类型多样性测试
	t.Run("FileTypeDiversity", pts.testFileTypeDiversity)

	// 6. 边界条件测试
	t.Run("BoundaryConditions", pts.testBoundaryConditions)

	// 7. 性能对比测试
	t.Run("PerformanceComparison", pts.testPerformanceComparison)
}

// testBasicPerformance 基础性能测试
func (pts *PerformanceTestSuite) testBasicPerformance(t *testing.T) {
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"SmallFiles", 5, 10, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 2; return c }()},
		{"MediumFiles", 10, 25, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 4; return c }()},
		{"LargeFiles", 20, 50, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 8; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runSinglePerformanceTest(t, tc.name, tc.fileCount, tc.pagesPerFile, tc.config)
			pts.addResult(result)
		})
	}
}

// testLargeScaleFiles 大规模文件测试
func (pts *PerformanceTestSuite) testLargeScaleFiles(t *testing.T) {
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"MassiveFiles", 50, 100, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 16; return c }()},
		{"HugeFiles", 100, 200, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 32; return c }()},
		{"ExtremeFiles", 200, 500, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 64; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runSinglePerformanceTest(t, tc.name, tc.fileCount, tc.pagesPerFile, tc.config)
			pts.addResult(result)
		})
	}
}

// testMemoryStress 内存压力测试
func (pts *PerformanceTestSuite) testMemoryStress(t *testing.T) {
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"LowMemory", 10, 20, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 1; return c }()},
		{"VeryLowMemory", 5, 15, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 1; return c }()},
		{"ExtremeLowMemory", 3, 10, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 1; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runSinglePerformanceTest(t, tc.name, tc.fileCount, tc.pagesPerFile, tc.config)
			pts.addResult(result)
		})
	}
}

// testConcurrencyPerformance 并发性能测试
func (pts *PerformanceTestSuite) testConcurrencyPerformance(t *testing.T) {
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"HighConcurrency", 30, 30, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 16; return c }()},
		{"VeryHighConcurrency", 50, 20, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 32; return c }()},
		{"ExtremeConcurrency", 100, 10, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 64; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runSinglePerformanceTest(t, tc.name, tc.fileCount, tc.pagesPerFile, tc.config)
			pts.addResult(result)
		})
	}
}

// testFileTypeDiversity 文件类型多样性测试
func (pts *PerformanceTestSuite) testFileTypeDiversity(t *testing.T) {
	// 测试不同页面大小的文件混合
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"MixedSizes", 15, 0, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 8; return c }()}, // 0表示混合大小
		{"VariableSizes", 25, 0, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 12; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runMixedSizeTest(t, tc.name, tc.fileCount, tc.config)
			pts.addResult(result)
		})
	}
}

// testBoundaryConditions 边界条件测试
func (pts *PerformanceTestSuite) testBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		config       *StreamingConfig
	}{
		{"SingleFile", 1, 100, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 1; return c }()},
		{"TwoFiles", 2, 50, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 1; return c }()},
		{"ManySmallFiles", 100, 1, func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 16; return c }()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pts.runSinglePerformanceTest(t, tc.name, tc.fileCount, tc.pagesPerFile, tc.config)
			pts.addResult(result)
		})
	}
}

// testPerformanceComparison 性能对比测试
func (pts *PerformanceTestSuite) testPerformanceComparison(t *testing.T) {
	// 对比不同配置下的性能
	testCases := []struct {
		name         string
		fileCount    int
		pagesPerFile int
		configs      []*StreamingConfig
	}{
		{"ConfigComparison", 20, 30, []*StreamingConfig{
			func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 4; return c }(),
			func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 8; return c }(),
			func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 16; return c }(),
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, config := range tc.configs {
				result := pts.runSinglePerformanceTest(t,
					fmt.Sprintf("%s_Config%d", tc.name, i+1),
					tc.fileCount, tc.pagesPerFile, config)
				pts.addResult(result)
			}
		})
	}
}

// runSinglePerformanceTest 运行单个性能测试
func (pts *PerformanceTestSuite) runSinglePerformanceTest(t *testing.T, testName string, fileCount, pagesPerFile int, config *StreamingConfig) PerformanceTestResult {
	result := PerformanceTestResult{
		TestName:      testName,
		FileCount:     fileCount,
		Configuration: fmt.Sprintf("Concurrent:%d", config.MaxConcurrentChunks),
		TestType:      "SingleTest",
	}

	// 创建测试文件
	testFiles, totalPages, totalSize, err := pts.createTestFiles(t, fileCount, pagesPerFile)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}
	defer pts.cleanupTestFiles(testFiles)

	result.TotalPages = totalPages
	result.TotalSize = totalSize

	// 记录开始时间和内存
	startTime := time.Now()
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	// 执行合并测试
	options := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		TempDirectory:     pts.TestDir,
		EnableGC:          true,
		ChunkSize:         10,
		UseStreaming:      true,
		OptimizeMemory:    true,
		ConcurrentWorkers: config.MaxConcurrentChunks,
	}
	merger := NewStreamingMerger(options)
	outputPath := filepath.Join(pts.TestDir, fmt.Sprintf("output_%s.pdf", testName))

	ctx := context.Background()
	_, mergeErr := merger.MergeStreaming(ctx, testFiles, outputPath, nil)

	// 记录结束时间和内存
	endTime := time.Now()
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)

	result.Duration = endTime.Sub(startTime)
	result.PeakMemory = endMem.Alloc - startMem.Alloc
	result.AvgMemory = (startMem.Alloc + endMem.Alloc) / 2
	result.Success = mergeErr == nil
	if mergeErr != nil {
		result.Error = mergeErr.Error()
	}

	// 清理输出文件
	os.Remove(outputPath)

	return result
}

// runMixedSizeTest 运行混合大小文件测试
func (pts *PerformanceTestSuite) runMixedSizeTest(t *testing.T, testName string, fileCount int, config *StreamingConfig) PerformanceTestResult {
	result := PerformanceTestResult{
		TestName:      testName,
		FileCount:     fileCount,
		Configuration: fmt.Sprintf("Concurrent:%d", config.MaxConcurrentChunks),
		TestType:      "MixedSizeTest",
	}

	// 创建不同大小的测试文件
	testFiles, totalPages, totalSize, err := pts.createMixedSizeTestFiles(t, fileCount)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}
	defer pts.cleanupTestFiles(testFiles)

	result.TotalPages = totalPages
	result.TotalSize = totalSize

	// 执行测试
	startTime := time.Now()
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	options := &MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		TempDirectory:     pts.TestDir,
		EnableGC:          true,
		ChunkSize:         10,
		UseStreaming:      true,
		OptimizeMemory:    true,
		ConcurrentWorkers: config.MaxConcurrentChunks,
	}
	merger := NewStreamingMerger(options)
	outputPath := filepath.Join(pts.TestDir, fmt.Sprintf("output_%s.pdf", testName))

	ctx := context.Background()
	_, mergeErr := merger.MergeStreaming(ctx, testFiles, outputPath, nil)

	endTime := time.Now()
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)

	result.Duration = endTime.Sub(startTime)
	result.PeakMemory = endMem.Alloc - startMem.Alloc
	result.AvgMemory = (startMem.Alloc + endMem.Alloc) / 2
	result.Success = mergeErr == nil
	if mergeErr != nil {
		result.Error = mergeErr.Error()
	}

	os.Remove(outputPath)
	return result
}

// createTestFiles 创建测试文件
func (pts *PerformanceTestSuite) createTestFiles(t *testing.T, fileCount, pagesPerFile int) ([]string, int, int64, error) {
	var testFiles []string
	totalPages := 0
	totalSize := int64(0)

	for i := 0; i < fileCount; i++ {
		fileName := fmt.Sprintf("test_%d.pdf", i)
		filePath := filepath.Join(pts.TestDir, fileName)

		// 创建测试PDF内容
		content := createPerformanceTestPDFContent(pagesPerFile)
		err := ioutil.WriteFile(filePath, content, 0644)
		if err != nil {
			return nil, 0, 0, err
		}

		testFiles = append(testFiles, filePath)
		totalPages += pagesPerFile

		// 获取文件大小
		if info, err := os.Stat(filePath); err == nil {
			totalSize += info.Size()
		}
	}

	return testFiles, totalPages, totalSize, nil
}

// createMixedSizeTestFiles 创建混合大小的测试文件
func (pts *PerformanceTestSuite) createMixedSizeTestFiles(t *testing.T, fileCount int) ([]string, int, int64, error) {
	var testFiles []string
	totalPages := 0
	totalSize := int64(0)

	for i := 0; i < fileCount; i++ {
		fileName := fmt.Sprintf("mixed_test_%d.pdf", i)
		filePath := filepath.Join(pts.TestDir, fileName)

		// 根据索引创建不同大小的文件
		pagesPerFile := 5 + (i % 20) // 5-24页
		content := createPerformanceTestPDFContent(pagesPerFile)
		err := ioutil.WriteFile(filePath, content, 0644)
		if err != nil {
			return nil, 0, 0, err
		}

		testFiles = append(testFiles, filePath)
		totalPages += pagesPerFile

		if info, err := os.Stat(filePath); err == nil {
			totalSize += info.Size()
		}
	}

	return testFiles, totalPages, totalSize, nil
}

// cleanupTestFiles 清理测试文件
func (pts *PerformanceTestSuite) cleanupTestFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

// addResult 添加测试结果
func (pts *PerformanceTestSuite) addResult(result PerformanceTestResult) {
	pts.Mutex.Lock()
	defer pts.Mutex.Unlock()
	pts.Results = append(pts.Results, result)
}

// GenerateReport 生成性能测试报告
func (pts *PerformanceTestSuite) GenerateReport(t *testing.T) {
	pts.Mutex.RLock()
	defer pts.Mutex.RUnlock()

	t.Logf("=== 全面性能测试报告 ===")
	t.Logf("测试时间: %s", pts.EndTime.Sub(pts.StartTime))
	t.Logf("总测试数: %d", len(pts.Results))

	// 统计成功和失败
	successCount := 0
	totalDuration := time.Duration(0)
	totalMemory := uint64(0)
	totalPages := 0
	totalSize := int64(0)

	for _, result := range pts.Results {
		if result.Success {
			successCount++
			totalDuration += result.Duration
			totalMemory += result.PeakMemory
			totalPages += result.TotalPages
			totalSize += result.TotalSize
		}
	}

	t.Logf("成功率: %.2f%% (%d/%d)", float64(successCount)/float64(len(pts.Results))*100, successCount, len(pts.Results))
	t.Logf("总处理页数: %d", totalPages)
	t.Logf("总处理大小: %.2f MB", float64(totalSize)/1024/1024)
	if successCount > 0 {
		t.Logf("平均处理时间: %v", totalDuration/time.Duration(successCount))
		t.Logf("平均内存使用: %.2f MB", float64(totalMemory)/1024/1024/float64(successCount))
	} else {
		t.Logf("平均处理时间: N/A (无成功测试)")
		t.Logf("平均内存使用: N/A (无成功测试)")
	}

	// 按测试类型分组统计
	typeStats := make(map[string][]PerformanceTestResult)
	for _, result := range pts.Results {
		typeStats[result.TestType] = append(typeStats[result.TestType], result)
	}

	for testType, results := range typeStats {
		t.Logf("\n=== %s 统计 ===", testType)
		success := 0
		var avgDuration time.Duration
		var avgMemory uint64

		for _, result := range results {
			if result.Success {
				success++
				avgDuration += result.Duration
				avgMemory += result.PeakMemory
			}
		}

		if success > 0 {
			t.Logf("成功率: %.2f%% (%d/%d)", float64(success)/float64(len(results))*100, success, len(results))
			t.Logf("平均处理时间: %v", avgDuration/time.Duration(success))
			t.Logf("平均内存使用: %.2f MB", float64(avgMemory)/1024/1024/float64(success))
		}
	}
}

// createPerformanceTestPDFContent 创建性能测试用的PDF内容
func createPerformanceTestPDFContent(pages int) []byte {
	// 创建基本的PDF内容用于性能测试
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
/Kids [%s]
/Count %d
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
(Performance Test Page) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000210 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
%d
%%EOF`, strings.Repeat("3 0 R ", pages), pages, 300+pages*50)

	return []byte(content)
}

// TestComprehensivePerformance 执行全面性能测试
func TestComprehensivePerformance(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "performance_test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建性能测试套件
	suite := NewPerformanceTestSuite(testDir)

	// 执行全面性能测试
	suite.RunComprehensivePerformanceTests(t)
}

// BenchmarkStreamingMerger 基准测试
func BenchmarkStreamingMerger(b *testing.B) {
	// 创建测试文件
	testDir, err := ioutil.TempDir("", "benchmark_test")
	if err != nil {
		b.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("bench_test_%d.pdf", i)
		filePath := filepath.Join(testDir, fileName)
		content := createPerformanceTestPDFContent(20)
		err := ioutil.WriteFile(filePath, content, 0644)
		if err != nil {
			b.Fatalf("创建测试文件失败: %v", err)
		}
		testFiles[i] = filePath
	}

	config := func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 4; return c }()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		options := &MergeOptions{
			MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
			TempDirectory:     testDir,
			EnableGC:          true,
			ChunkSize:         10,
			UseStreaming:      true,
			OptimizeMemory:    true,
			ConcurrentWorkers: config.MaxConcurrentChunks,
		}
		merger := NewStreamingMerger(options)
		outputPath := filepath.Join(testDir, fmt.Sprintf("bench_output_%d.pdf", i))

		ctx := context.Background()
		_, err := merger.MergeStreaming(ctx, testFiles, outputPath, nil)
		if err != nil {
			b.Fatalf("合并失败: %v", err)
		}

		os.Remove(outputPath)
	}
}

// BenchmarkMemoryUsage 内存使用基准测试
func BenchmarkMemoryUsage(b *testing.B) {
	testDir, err := ioutil.TempDir("", "memory_benchmark")
	if err != nil {
		b.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建大文件进行内存测试
	testFiles := make([]string, 5)
	for i := 0; i < 5; i++ {
		fileName := fmt.Sprintf("memory_test_%d.pdf", i)
		filePath := filepath.Join(testDir, fileName)
		content := createPerformanceTestPDFContent(100) // 大文件
		err := ioutil.WriteFile(filePath, content, 0644)
		if err != nil {
			b.Fatalf("创建测试文件失败: %v", err)
		}
		testFiles[i] = filePath
	}

	config := func() *StreamingConfig { c := DefaultStreamingConfig(); c.MaxConcurrentChunks = 2; return c }()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		options := &MergeOptions{
			MaxMemoryUsage:    50 * 1024 * 1024, // 50MB
			TempDirectory:     testDir,
			EnableGC:          true,
			ChunkSize:         10,
			UseStreaming:      true,
			OptimizeMemory:    true,
			ConcurrentWorkers: config.MaxConcurrentChunks,
		}
		merger := NewStreamingMerger(options)
		outputPath := filepath.Join(testDir, fmt.Sprintf("memory_output_%d.pdf", i))

		ctx := context.Background()
		_, err := merger.MergeStreaming(ctx, testFiles, outputPath, nil)
		if err != nil {
			b.Fatalf("合并失败: %v", err)
		}

		runtime.ReadMemStats(&m2)
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB")

		os.Remove(outputPath)
	}
}
