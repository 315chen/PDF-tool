package test_utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// MockDataGenerator 模拟数据生成器
type MockDataGenerator struct {
	rand *rand.Rand
}

// NewMockDataGenerator 创建新的模拟数据生成器
func NewMockDataGenerator() *MockDataGenerator {
	return &MockDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateFileEntry 生成模拟文件条目
func (m *MockDataGenerator) GenerateFileEntry() *model.FileEntry {
	return &model.FileEntry{
		Path:        fmt.Sprintf("/tmp/test_%d.pdf", m.rand.Intn(1000)),
		DisplayName: fmt.Sprintf("test_%d.pdf", m.rand.Intn(1000)),
		Size:        int64(m.rand.Intn(10000000)), // 0-10MB
		PageCount:   m.rand.Intn(100) + 1,        // 1-100页
		IsEncrypted: m.rand.Float32() < 0.2,      // 20%概率加密
		IsValid:     m.rand.Float32() < 0.9,      // 90%概率有效
		Order:       m.rand.Intn(10),
	}
}

// GenerateFileEntries 生成多个模拟文件条目
func (m *MockDataGenerator) GenerateFileEntries(count int) []*model.FileEntry {
	entries := make([]*model.FileEntry, count)
	for i := 0; i < count; i++ {
		entries[i] = m.GenerateFileEntry()
		entries[i].Order = i
	}
	return entries
}

// GenerateMergeJob 生成模拟合并任务
func (m *MockDataGenerator) GenerateMergeJob() *model.MergeJob {
	additionalCount := m.rand.Intn(5) + 1 // 1-5个附加文件
	additionalFiles := make([]string, additionalCount)
	
	for i := 0; i < additionalCount; i++ {
		additionalFiles[i] = fmt.Sprintf("/tmp/additional_%d.pdf", i)
	}
	
	job := model.NewMergeJob(
		"/tmp/main.pdf",
		additionalFiles,
		"/tmp/output.pdf",
	)
	
	// 随机设置任务状态
	statuses := []model.JobStatus{
		model.JobPending,
		model.JobRunning,
		model.JobCompleted,
		model.JobFailed,
	}
	job.Status = statuses[m.rand.Intn(len(statuses))]
	job.Progress = m.rand.Float64() * 100
	
	return job
}

// GenerateConfig 生成模拟配置
func (m *MockDataGenerator) GenerateConfig() *model.Config {
	config := model.DefaultConfig()
	
	// 随机调整一些配置值
	config.MaxMemoryUsage = int64(m.rand.Intn(500)+50) * 1024 * 1024 // 50-550MB
	config.TempDirectory = fmt.Sprintf("/tmp/pdf-merger-%d", m.rand.Intn(1000))
	config.OutputDirectory = fmt.Sprintf("/tmp/output-%d", m.rand.Intn(1000))
	config.EnableAutoDecrypt = m.rand.Float32() < 0.8 // 80%概率启用
	config.WindowWidth = m.rand.Intn(400) + 600       // 600-1000
	config.WindowHeight = m.rand.Intn(300) + 400      // 400-700
	
	return config
}

// GenerateTestFiles 生成测试文件路径列表
func (m *MockDataGenerator) GenerateTestFiles(count int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		files[i] = fmt.Sprintf("/tmp/test_file_%d.pdf", i)
	}
	return files
}

// GenerateRandomString 生成随机字符串
func (m *MockDataGenerator) GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[m.rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateRandomError 生成随机错误
func (m *MockDataGenerator) GenerateRandomError() error {
	errors := []string{
		"文件不存在",
		"权限被拒绝",
		"文件已损坏",
		"内存不足",
		"网络错误",
		"超时",
		"无效格式",
		"加密文件",
	}
	
	return fmt.Errorf(errors[m.rand.Intn(len(errors))])
}

// TestScenario 测试场景
type TestScenario struct {
	Name        string
	Description string
	Setup       func() interface{}
	Execute     func(interface{}) error
	Verify      func(interface{}, error) bool
}

// GenerateTestScenarios 生成测试场景
func (m *MockDataGenerator) GenerateTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "正常合并场景",
			Description: "测试正常的PDF合并流程",
			Setup: func() interface{} {
				return m.GenerateMergeJob()
			},
			Execute: func(data interface{}) error {
				// 模拟执行
				return nil
			},
			Verify: func(data interface{}, err error) bool {
				return err == nil
			},
		},
		{
			Name:        "文件不存在场景",
			Description: "测试文件不存在时的错误处理",
			Setup: func() interface{} {
				job := m.GenerateMergeJob()
				job.MainFile = "/nonexistent/file.pdf"
				return job
			},
			Execute: func(data interface{}) error {
				return fmt.Errorf("文件不存在")
			},
			Verify: func(data interface{}, err error) bool {
				return err != nil
			},
		},
		{
			Name:        "内存不足场景",
			Description: "测试内存不足时的处理",
			Setup: func() interface{} {
				config := m.GenerateConfig()
				config.MaxMemoryUsage = 1024 // 很小的内存限制
				return config
			},
			Execute: func(data interface{}) error {
				return fmt.Errorf("内存不足")
			},
			Verify: func(data interface{}, err error) bool {
				return err != nil
			},
		},
	}
}

// PerformanceTestData 性能测试数据
type PerformanceTestData struct {
	FileCount    int
	FileSizes    []int64
	ExpectedTime time.Duration
}

// GeneratePerformanceTestData 生成性能测试数据
func (m *MockDataGenerator) GeneratePerformanceTestData() []PerformanceTestData {
	return []PerformanceTestData{
		{
			FileCount:    2,
			FileSizes:    []int64{1024 * 1024, 2 * 1024 * 1024}, // 1MB, 2MB
			ExpectedTime: 1 * time.Second,
		},
		{
			FileCount:    5,
			FileSizes:    []int64{5 * 1024 * 1024, 10 * 1024 * 1024, 3 * 1024 * 1024, 7 * 1024 * 1024, 4 * 1024 * 1024},
			ExpectedTime: 5 * time.Second,
		},
		{
			FileCount:    10,
			FileSizes:    m.generateRandomSizes(10),
			ExpectedTime: 10 * time.Second,
		},
	}
}

// generateRandomSizes 生成随机文件大小
func (m *MockDataGenerator) generateRandomSizes(count int) []int64 {
	sizes := make([]int64, count)
	for i := 0; i < count; i++ {
		sizes[i] = int64(m.rand.Intn(20)+1) * 1024 * 1024 // 1-20MB
	}
	return sizes
}

// ErrorTestCase 错误测试用例
type ErrorTestCase struct {
	Name          string
	InputData     interface{}
	ExpectedError string
	ShouldRetry   bool
}

// GenerateErrorTestCases 生成错误测试用例
func (m *MockDataGenerator) GenerateErrorTestCases() []ErrorTestCase {
	return []ErrorTestCase{
		{
			Name:          "文件不存在",
			InputData:     "/nonexistent/file.pdf",
			ExpectedError: "文件不存在",
			ShouldRetry:   false,
		},
		{
			Name:          "权限被拒绝",
			InputData:     "/root/protected.pdf",
			ExpectedError: "权限被拒绝",
			ShouldRetry:   false,
		},
		{
			Name:          "网络错误",
			InputData:     "http://example.com/file.pdf",
			ExpectedError: "网络错误",
			ShouldRetry:   true,
		},
		{
			Name:          "内存不足",
			InputData:     m.GenerateConfig(),
			ExpectedError: "内存不足",
			ShouldRetry:   true,
		},
	}
}

// BenchmarkData 基准测试数据
type BenchmarkData struct {
	Name      string
	DataSize  int
	Setup     func(int) interface{}
	Operation func(interface{}) error
}

// GenerateBenchmarkData 生成基准测试数据
func (m *MockDataGenerator) GenerateBenchmarkData() []BenchmarkData {
	return []BenchmarkData{
		{
			Name:     "小文件合并",
			DataSize: 10,
			Setup: func(size int) interface{} {
				return m.GenerateTestFiles(size)
			},
			Operation: func(data interface{}) error {
				// 模拟合并操作
				time.Sleep(time.Millisecond)
				return nil
			},
		},
		{
			Name:     "大文件合并",
			DataSize: 100,
			Setup: func(size int) interface{} {
				return m.GenerateTestFiles(size)
			},
			Operation: func(data interface{}) error {
				// 模拟合并操作
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		},
		{
			Name:     "内存优化合并",
			DataSize: 1000,
			Setup: func(size int) interface{} {
				return m.GenerateTestFiles(size)
			},
			Operation: func(data interface{}) error {
				// 模拟流式合并操作
				time.Sleep(100 * time.Millisecond)
				return nil
			},
		},
	}
}