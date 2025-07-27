package test_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func TestTestHelper_Basic(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// 测试获取临时目录
	tempDir := helper.GetTempDir()
	if tempDir == "" {
		t.Error("Expected non-empty temp directory")
	}

	// 验证目录存在
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory should exist: %s", tempDir)
	}
}

func TestTestHelper_CreateTestFile(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	filename := "test.txt"
	content := []byte("test content")

	filePath := helper.CreateTestFile(filename, content)

	// 验证文件路径
	expectedPath := filepath.Join(helper.GetTempDir(), filename)
	if filePath != expectedPath {
		t.Errorf("Expected file path %s, got %s", expectedPath, filePath)
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Test file should exist: %s", filePath)
	}

	// 验证文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Expected content %s, got %s", string(content), string(data))
	}
}

func TestTestHelper_CreateTestPDF(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	filename := "test.pdf"
	filePath := helper.CreateTestPDF(filename)

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Test PDF should exist: %s", filePath)
	}

	// 验证PDF内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test PDF: %v", err)
	}

	content := string(data)
	if !contains(content, "%PDF-") {
		t.Error("Test PDF should contain PDF header")
	}

	if !contains(content, "%%EOF") {
		t.Error("Test PDF should contain EOF marker")
	}
}

func TestTestHelper_CreateTestDirectory(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	dirname := "testdir"
	dirPath := helper.CreateTestDirectory(dirname)

	// 验证目录路径
	expectedPath := filepath.Join(helper.GetTempDir(), dirname)
	if dirPath != expectedPath {
		t.Errorf("Expected directory path %s, got %s", expectedPath, dirPath)
	}

	// 验证目录存在
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("Test directory should exist: %s", dirPath)
	}

	if !info.IsDir() {
		t.Errorf("Path should be a directory: %s", dirPath)
	}
}

func TestTestHelper_Assertions(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// 测试AssertNoError
	helper.AssertNoError(nil)

	// 测试AssertEqual
	helper.AssertEqual(42, 42)
	helper.AssertEqual("test", "test")

	// 测试AssertNotEqual
	helper.AssertNotEqual(42, 43)
	helper.AssertNotEqual("test", "other")

	// 测试AssertTrue
	helper.AssertTrue(true)

	// 测试AssertFalse
	helper.AssertFalse(false)

	// 测试AssertContains
	helper.AssertContains("hello world", "world")

	// 测试AssertNotContains
	helper.AssertNotContains("hello world", "xyz")
}

func TestTestHelper_FileAssertions(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// 创建测试文件
	filePath := helper.CreateTestFile("exists.txt", []byte("content"))

	// 测试AssertFileExists
	helper.AssertFileExists(filePath)

	// 测试AssertFileNotExists
	nonExistentPath := filepath.Join(helper.GetTempDir(), "nonexistent.txt")
	helper.AssertFileNotExists(nonExistentPath)
}

func TestTestHelper_WaitForCondition(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// 测试立即满足的条件
	start := time.Now()
	helper.WaitForCondition(func() bool {
		return true
	}, time.Second, "immediate condition")
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		t.Errorf("Expected immediate return, took %v", duration)
	}

	// 测试延迟满足的条件
	counter := 0
	start = time.Now()
	helper.WaitForCondition(func() bool {
		counter++
		return counter >= 5
	}, time.Second, "delayed condition")
	duration = time.Since(start)

	if duration < 40*time.Millisecond {
		t.Errorf("Expected some delay, took %v", duration)
	}

	if counter < 5 {
		t.Errorf("Expected counter to be at least 5, got %d", counter)
	}
}

func TestTestHelper_WaitForJobCompletion(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// 创建一个已完成的任务
	job := model.NewMergeJob("/test/main.pdf", []string{"/test/file1.pdf"}, "/test/output.pdf")
	job.Status = model.JobCompleted

	// 应该立即返回
	start := time.Now()
	helper.WaitForJobCompletion(job, time.Second)
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		t.Errorf("Expected immediate return for completed job, took %v", duration)
	}
}

func TestMockTimeProvider(t *testing.T) {
	initialTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	provider := NewMockTimeProvider(initialTime)

	// 测试获取时间
	now := provider.Now()
	if !now.Equal(initialTime) {
		t.Errorf("Expected time %v, got %v", initialTime, now)
	}

	// 测试设置时间
	newTime := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)
	provider.SetTime(newTime)

	now = provider.Now()
	if !now.Equal(newTime) {
		t.Errorf("Expected time %v, got %v", newTime, now)
	}

	// 测试增加时间
	duration := time.Hour
	provider.AddTime(duration)

	expectedTime := newTime.Add(duration)
	now = provider.Now()
	if !now.Equal(expectedTime) {
		t.Errorf("Expected time %v, got %v", expectedTime, now)
	}
}

func TestTestRunner(t *testing.T) {
	runner := NewTestRunner(t)
	defer runner.Cleanup()

	// 添加测试场景
	scenario1 := TestScenario{
		Name:        "Test Scenario 1",
		Description: "First test scenario",
		Setup: func() interface{} {
			return "test data"
		},
		Execute: func(data interface{}) error {
			return nil
		},
		Verify: func(data interface{}, err error) bool {
			return err == nil && data == "test data"
		},
	}

	scenario2 := TestScenario{
		Name:        "Test Scenario 2",
		Description: "Second test scenario",
		Setup: func() interface{} {
			return 42
		},
		Execute: func(data interface{}) error {
			return fmt.Errorf("test error")
		},
		Verify: func(data interface{}, err error) bool {
			return err != nil && data == 42
		},
	}

	runner.AddScenario(scenario1)
	runner.AddScenario(scenario2)

	// 运行场景
	runner.RunScenarios()
}

func TestBenchmarkRunner(t *testing.T) {
	runner := NewBenchmarkRunner()

	// 添加基准测试
	benchmark1 := BenchmarkData{
		Name:     "Benchmark 1",
		DataSize: 100,
		Setup: func(size int) interface{} {
			return make([]int, size)
		},
		Operation: func(data interface{}) error {
			slice := data.([]int)
			for i := range slice {
				slice[i] = i
			}
			return nil
		},
	}

	benchmark2 := BenchmarkData{
		Name:     "Benchmark 2",
		DataSize: 50,
		Setup: func(size int) interface{} {
			return make(map[int]int, size)
		},
		Operation: func(data interface{}) error {
			m := data.(map[int]int)
			for i := 0; i < 50; i++ {
				m[i] = i * 2
			}
			return nil
		},
	}

	runner.AddBenchmark(benchmark1)
	runner.AddBenchmark(benchmark2)

	// 运行基准测试
	testing.Benchmark(func(b *testing.B) {
		runner.RunBenchmarks(b)
	})
}

func TestPerformanceProfiler(t *testing.T) {
	profiler := NewPerformanceProfiler()

	// 测试性能分析
	profiler.Start()

	// 模拟一些工作
	time.Sleep(10 * time.Millisecond)

	profiler.Stop()

	// 验证执行时间
	duration := profiler.GetDuration()
	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration at least 10ms, got %v", duration)
	}

	// 测试指标
	profiler.SetMetric("operations", 100)
	profiler.SetMetric("memory_usage", 1024*1024)

	operations := profiler.GetMetric("operations")
	if operations != 100 {
		t.Errorf("Expected operations 100, got %v", operations)
	}

	memoryUsage := profiler.GetMetric("memory_usage")
	if memoryUsage != 1024*1024 {
		t.Errorf("Expected memory usage 1MB, got %v", memoryUsage)
	}

	// 测试获取所有指标
	allMetrics := profiler.GetAllMetrics()
	if len(allMetrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(allMetrics))
	}

	if allMetrics["operations"] != 100 {
		t.Errorf("Expected operations 100 in all metrics, got %v", allMetrics["operations"])
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAt(s, substr, 1))))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) > len(s) {
		return containsAt(s, substr, start+1)
	}
	if s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}
