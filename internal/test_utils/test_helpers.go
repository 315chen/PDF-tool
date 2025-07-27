package test_utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// TestHelper 测试助手
type TestHelper struct {
	t       *testing.T
	tempDir string
	cleanup []func()
	mutex   sync.Mutex
}

// NewTestHelper 创建新的测试助手
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:       t,
		cleanup: make([]func(), 0),
	}
}

// GetTempDir 获取临时目录
func (h *TestHelper) GetTempDir() string {
	if h.tempDir == "" {
		h.tempDir = h.t.TempDir()
	}
	return h.tempDir
}

// CreateTestFile 创建测试文件
func (h *TestHelper) CreateTestFile(filename string, content []byte) string {
	tempDir := h.GetTempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		h.t.Fatalf("Failed to create test file %s: %v", filePath, err)
	}

	return filePath
}

// CreateTestPDF 创建测试PDF文件
func (h *TestHelper) CreateTestPDF(filename string) string {
	content := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000074 00000 n \n0000000120 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n179\n%%EOF")
	return h.CreateTestFile(filename, content)
}

// CreateTestDirectory 创建测试目录
func (h *TestHelper) CreateTestDirectory(dirname string) string {
	tempDir := h.GetTempDir()
	dirPath := filepath.Join(tempDir, dirname)

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		h.t.Fatalf("Failed to create test directory %s: %v", dirPath, err)
	}

	return dirPath
}

// AddCleanup 添加清理函数
func (h *TestHelper) AddCleanup(cleanup func()) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.cleanup = append(h.cleanup, cleanup)
}

// Cleanup 执行清理
func (h *TestHelper) Cleanup() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}
	h.cleanup = h.cleanup[:0]
}

// AssertNoError 断言没有错误
func (h *TestHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
	if err != nil {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected no error, but got: %v. %v", err, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected no error, but got: %v", err)
		}
	}
}

// AssertError 断言有错误
func (h *TestHelper) AssertError(err error, msgAndArgs ...interface{}) {
	if err == nil {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected error, but got nil. %v", msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected error, but got nil")
		}
	}
}

// AssertEqual 断言相等
func (h *TestHelper) AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected %v, but got %v. %v", expected, actual, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected %v, but got %v", expected, actual)
		}
	}
}

// AssertNotEqual 断言不相等
func (h *TestHelper) AssertNotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected == actual {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected %v to not equal %v. %v", expected, actual, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected %v to not equal %v", expected, actual)
		}
	}
}

// AssertTrue 断言为真
func (h *TestHelper) AssertTrue(condition bool, msgAndArgs ...interface{}) {
	if !condition {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected true, but got false. %v", msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected true, but got false")
		}
	}
}

// AssertFalse 断言为假
func (h *TestHelper) AssertFalse(condition bool, msgAndArgs ...interface{}) {
	if condition {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected false, but got true. %v", msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected false, but got true")
		}
	}
}

// AssertContains 断言包含
func (h *TestHelper) AssertContains(haystack, needle string, msgAndArgs ...interface{}) {
	if !strings.Contains(haystack, needle) {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected '%s' to contain '%s'. %v", haystack, needle, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected '%s' to contain '%s'", haystack, needle)
		}
	}
}

// AssertNotContains 断言不包含
func (h *TestHelper) AssertNotContains(haystack, needle string, msgAndArgs ...interface{}) {
	if strings.Contains(haystack, needle) {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected '%s' to not contain '%s'. %v", haystack, needle, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected '%s' to not contain '%s'", haystack, needle)
		}
	}
}

// AssertFileExists 断言文件存在
func (h *TestHelper) AssertFileExists(filePath string, msgAndArgs ...interface{}) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected file to exist: %s. %v", filePath, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected file to exist: %s", filePath)
		}
	}
}

// AssertFileNotExists 断言文件不存在
func (h *TestHelper) AssertFileNotExists(filePath string, msgAndArgs ...interface{}) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		if len(msgAndArgs) > 0 {
			h.t.Fatalf("Expected file to not exist: %s. %v", filePath, msgAndArgs[0])
		} else {
			h.t.Fatalf("Expected file to not exist: %s", filePath)
		}
	}
}

// WaitForCondition 等待条件满足
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// WaitForJobCompletion 等待任务完成
func (h *TestHelper) WaitForJobCompletion(job *model.MergeJob, timeout time.Duration) {
	h.WaitForCondition(func() bool {
		return job.Status == model.JobCompleted || job.Status == model.JobFailed
	}, timeout, fmt.Sprintf("job %s to complete", job.ID))
}

// MockTimeProvider 模拟时间提供者
type MockTimeProvider struct {
	currentTime time.Time
	mutex       sync.RWMutex
}

// NewMockTimeProvider 创建新的模拟时间提供者
func NewMockTimeProvider(initialTime time.Time) *MockTimeProvider {
	return &MockTimeProvider{
		currentTime: initialTime,
	}
}

// Now 获取当前时间
func (m *MockTimeProvider) Now() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentTime
}

// SetTime 设置时间
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.currentTime = t
}

// AddTime 增加时间
func (m *MockTimeProvider) AddTime(d time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.currentTime = m.currentTime.Add(d)
}

// TestRunner 测试运行器
type TestRunner struct {
	helper    *TestHelper
	scenarios []TestScenario
}

// NewTestRunner 创建新的测试运行器
func NewTestRunner(t *testing.T) *TestRunner {
	return &TestRunner{
		helper:    NewTestHelper(t),
		scenarios: make([]TestScenario, 0),
	}
}

// AddScenario 添加测试场景
func (r *TestRunner) AddScenario(scenario TestScenario) {
	r.scenarios = append(r.scenarios, scenario)
}

// RunScenarios 运行所有测试场景
func (r *TestRunner) RunScenarios() {
	for _, scenario := range r.scenarios {
		r.helper.t.Run(scenario.Name, func(t *testing.T) {
			data := scenario.Setup()
			err := scenario.Execute(data)
			if !scenario.Verify(data, err) {
				t.Fatalf("Scenario verification failed: %s", scenario.Description)
			}
		})
	}
}

// Cleanup 清理测试运行器
func (r *TestRunner) Cleanup() {
	r.helper.Cleanup()
}

// BenchmarkRunner 基准测试运行器
type BenchmarkRunner struct {
	benchmarks []BenchmarkData
}

// NewBenchmarkRunner 创建新的基准测试运行器
func NewBenchmarkRunner() *BenchmarkRunner {
	return &BenchmarkRunner{
		benchmarks: make([]BenchmarkData, 0),
	}
}

// AddBenchmark 添加基准测试
func (r *BenchmarkRunner) AddBenchmark(benchmark BenchmarkData) {
	r.benchmarks = append(r.benchmarks, benchmark)
}

// RunBenchmarks 运行所有基准测试
func (r *BenchmarkRunner) RunBenchmarks(b *testing.B) {
	for _, benchmark := range r.benchmarks {
		b.Run(benchmark.Name, func(b *testing.B) {
			data := benchmark.Setup(benchmark.DataSize)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				err := benchmark.Operation(data)
				if err != nil {
					b.Fatalf("Benchmark operation failed: %v", err)
				}
			}
		})
	}
}

// PerformanceProfiler 性能分析器
type PerformanceProfiler struct {
	startTime time.Time
	endTime   time.Time
	metrics   map[string]interface{}
}

// NewPerformanceProfiler 创建新的性能分析器
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		metrics: make(map[string]interface{}),
	}
}

// Start 开始性能分析
func (p *PerformanceProfiler) Start() {
	p.startTime = time.Now()
}

// Stop 停止性能分析
func (p *PerformanceProfiler) Stop() {
	p.endTime = time.Now()
}

// GetDuration 获取执行时间
func (p *PerformanceProfiler) GetDuration() time.Duration {
	return p.endTime.Sub(p.startTime)
}

// SetMetric 设置指标
func (p *PerformanceProfiler) SetMetric(name string, value interface{}) {
	p.metrics[name] = value
}

// GetMetric 获取指标
func (p *PerformanceProfiler) GetMetric(name string) interface{} {
	return p.metrics[name]
}

// GetAllMetrics 获取所有指标
func (p *PerformanceProfiler) GetAllMetrics() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range p.metrics {
		result[k] = v
	}
	return result
}
