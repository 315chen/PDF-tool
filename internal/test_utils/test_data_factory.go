package test_utils

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// TestDataFactory 测试数据工厂
type TestDataFactory struct {
	rand *rand.Rand
}

// NewTestDataFactory 创建新的测试数据工厂
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateValidFileEntry 创建有效的文件条目
func (f *TestDataFactory) CreateValidFileEntry(path string) *model.FileEntry {
	return &model.FileEntry{
		Path:        path,
		DisplayName: filepath.Base(path),
		Size:        int64(f.rand.Intn(10000000) + 1000), // 1KB-10MB
		PageCount:   f.rand.Intn(100) + 1,               // 1-100页
		IsValid:     true,
		IsEncrypted: false,
		Order:       0,
	}
}

// CreateEncryptedFileEntry 创建加密的文件条目
func (f *TestDataFactory) CreateEncryptedFileEntry(path string) *model.FileEntry {
	entry := f.CreateValidFileEntry(path)
	entry.IsEncrypted = true
	return entry
}

// CreateInvalidFileEntry 创建无效的文件条目
func (f *TestDataFactory) CreateInvalidFileEntry(path string, errorMsg string) *model.FileEntry {
	entry := &model.FileEntry{
		Path:        path,
		DisplayName: filepath.Base(path),
		Size:        0,
		PageCount:   0,
		IsValid:     false,
		IsEncrypted: false,
		Order:       0,
	}
	if errorMsg != "" {
		entry.SetError(errorMsg)
	}
	return entry
}

// CreateLargeFileEntry 创建大文件条目
func (f *TestDataFactory) CreateLargeFileEntry(path string, sizeMB int) *model.FileEntry {
	entry := f.CreateValidFileEntry(path)
	entry.Size = int64(sizeMB) * 1024 * 1024
	entry.PageCount = sizeMB * 10 // 假设每MB约10页
	return entry
}

// CreateFileEntryList 创建文件条目列表
func (f *TestDataFactory) CreateFileEntryList(count int) []*model.FileEntry {
	entries := make([]*model.FileEntry, count)
	for i := 0; i < count; i++ {
		path := fmt.Sprintf("/test/file_%d.pdf", i)
		entries[i] = f.CreateValidFileEntry(path)
		entries[i].Order = i
	}
	return entries
}

// CreateMixedFileEntryList 创建混合文件条目列表（包含有效、无效、加密文件）
func (f *TestDataFactory) CreateMixedFileEntryList(validCount, invalidCount, encryptedCount int) []*model.FileEntry {
	totalCount := validCount + invalidCount + encryptedCount
	entries := make([]*model.FileEntry, totalCount)
	
	index := 0
	
	// 添加有效文件
	for i := 0; i < validCount; i++ {
		path := fmt.Sprintf("/test/valid_%d.pdf", i)
		entries[index] = f.CreateValidFileEntry(path)
		entries[index].Order = index
		index++
	}
	
	// 添加无效文件
	for i := 0; i < invalidCount; i++ {
		path := fmt.Sprintf("/test/invalid_%d.pdf", i)
		entries[index] = f.CreateInvalidFileEntry(path, "文件损坏")
		entries[index].Order = index
		index++
	}
	
	// 添加加密文件
	for i := 0; i < encryptedCount; i++ {
		path := fmt.Sprintf("/test/encrypted_%d.pdf", i)
		entries[index] = f.CreateEncryptedFileEntry(path)
		entries[index].Order = index
		index++
	}
	
	return entries
}

// CreatePendingMergeJob 创建待处理的合并任务
func (f *TestDataFactory) CreatePendingMergeJob(mainFile string, additionalFiles []string, outputPath string) *model.MergeJob {
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)
	job.Status = model.JobPending
	job.Progress = 0.0
	return job
}

// CreateRunningMergeJob 创建运行中的合并任务
func (f *TestDataFactory) CreateRunningMergeJob(mainFile string, additionalFiles []string, outputPath string, progress float64) *model.MergeJob {
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)
	job.Status = model.JobRunning
	job.Progress = progress
	// StartedAt字段不存在，使用CreatedAt代替
	return job
}

// CreateCompletedMergeJob 创建已完成的合并任务
func (f *TestDataFactory) CreateCompletedMergeJob(mainFile string, additionalFiles []string, outputPath string) *model.MergeJob {
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)
	job.Status = model.JobCompleted
	job.Progress = 100.0
	now := time.Now()
	job.CreatedAt = now.Add(-2 * time.Minute)
	job.CompletedAt = &now
	return job
}

// CreateFailedMergeJob 创建失败的合并任务
func (f *TestDataFactory) CreateFailedMergeJob(mainFile string, additionalFiles []string, outputPath string, errorMsg string) *model.MergeJob {
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)
	job.Status = model.JobFailed
	job.Progress = f.rand.Float64() * 100 // 随机进度
	now := time.Now()
	job.CreatedAt = now.Add(-time.Minute)
	job.CompletedAt = &now
	if errorMsg != "" {
		job.Error = fmt.Errorf(errorMsg)
	}
	return job
}

// CreateCancelledMergeJob 创建已取消的合并任务
func (f *TestDataFactory) CreateCancelledMergeJob(mainFile string, additionalFiles []string, outputPath string) *model.MergeJob {
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)
	job.Status = model.JobFailed // 使用JobFailed代替JobCancelled
	job.Progress = f.rand.Float64() * 100 // 随机进度
	now := time.Now()
	job.CreatedAt = now.Add(-time.Minute)
	job.CompletedAt = &now
	return job
}

// CreateTestConfig 创建测试配置
func (f *TestDataFactory) CreateTestConfig() *model.Config {
	config := model.DefaultConfig()
	config.TempDirectory = "/tmp/test"
	config.OutputDirectory = "/tmp/output"
	config.MaxMemoryUsage = 256 * 1024 * 1024 // 256MB
	config.EnableAutoDecrypt = true
	config.WindowWidth = 800
	config.WindowHeight = 600
	return config
}

// CreateCustomConfig 创建自定义配置
func (f *TestDataFactory) CreateCustomConfig(tempDir, outputDir string, maxMemory int64) *model.Config {
	config := f.CreateTestConfig()
	config.TempDirectory = tempDir
	config.OutputDirectory = outputDir
	config.MaxMemoryUsage = maxMemory
	return config
}

// CreateFileEntrySlice 创建文件条目切片
func (f *TestDataFactory) CreateFileEntrySlice(filePaths []string) []*model.FileEntry {
	entries := make([]*model.FileEntry, len(filePaths))
	for i, path := range filePaths {
		entries[i] = f.CreateValidFileEntry(path)
		entries[i].Order = i
	}
	return entries
}

// TestScenarioBuilder 测试场景构建器
type TestScenarioBuilder struct {
	factory *TestDataFactory
}

// NewTestScenarioBuilder 创建新的测试场景构建器
func NewTestScenarioBuilder() *TestScenarioBuilder {
	return &TestScenarioBuilder{
		factory: NewTestDataFactory(),
	}
}

// BuildNormalMergeScenario 构建正常合并场景
func (b *TestScenarioBuilder) BuildNormalMergeScenario() *TestScenario {
	return &TestScenario{
		Name:        "正常合并场景",
		Description: "测试正常的PDF合并流程",
		Setup: func() interface{} {
			mainFile := "/test/main.pdf"
			additionalFiles := []string{"/test/file1.pdf", "/test/file2.pdf"}
			outputPath := "/test/output.pdf"
			return b.factory.CreatePendingMergeJob(mainFile, additionalFiles, outputPath)
		},
		Execute: func(data interface{}) error {
			// 模拟正常执行
			return nil
		},
		Verify: func(data interface{}, err error) bool {
			return err == nil
		},
	}
}

// BuildErrorScenario 构建错误场景
func (b *TestScenarioBuilder) BuildErrorScenario(errorType string) *TestScenario {
	return &TestScenario{
		Name:        fmt.Sprintf("%s错误场景", errorType),
		Description: fmt.Sprintf("测试%s时的错误处理", errorType),
		Setup: func() interface{} {
			mainFile := "/nonexistent/main.pdf"
			additionalFiles := []string{"/nonexistent/file1.pdf"}
			outputPath := "/test/output.pdf"
			return b.factory.CreatePendingMergeJob(mainFile, additionalFiles, outputPath)
		},
		Execute: func(data interface{}) error {
			return fmt.Errorf("%s", errorType)
		},
		Verify: func(data interface{}, err error) bool {
			return err != nil
		},
	}
}

// BuildPerformanceScenario 构建性能测试场景
func (b *TestScenarioBuilder) BuildPerformanceScenario(fileCount int, expectedDuration time.Duration) *TestScenario {
	return &TestScenario{
		Name:        fmt.Sprintf("性能测试场景_%d文件", fileCount),
		Description: fmt.Sprintf("测试%d个文件的合并性能", fileCount),
		Setup: func() interface{} {
			mainFile := "/test/main.pdf"
			additionalFiles := make([]string, fileCount-1)
			for i := 0; i < fileCount-1; i++ {
				additionalFiles[i] = fmt.Sprintf("/test/file_%d.pdf", i+1)
			}
			outputPath := "/test/output.pdf"
			return b.factory.CreatePendingMergeJob(mainFile, additionalFiles, outputPath)
		},
		Execute: func(data interface{}) error {
			// 模拟处理时间
			time.Sleep(expectedDuration)
			return nil
		},
		Verify: func(data interface{}, err error) bool {
			return err == nil
		},
	}
}

// BuildConcurrencyScenario 构建并发测试场景
func (b *TestScenarioBuilder) BuildConcurrencyScenario(concurrentCount int) *TestScenario {
	return &TestScenario{
		Name:        fmt.Sprintf("并发测试场景_%d并发", concurrentCount),
		Description: fmt.Sprintf("测试%d个并发操作", concurrentCount),
		Setup: func() interface{} {
			jobs := make([]*model.MergeJob, concurrentCount)
			for i := 0; i < concurrentCount; i++ {
				mainFile := fmt.Sprintf("/test/main_%d.pdf", i)
				additionalFiles := []string{fmt.Sprintf("/test/file_%d.pdf", i)}
				outputPath := fmt.Sprintf("/test/output_%d.pdf", i)
				jobs[i] = b.factory.CreatePendingMergeJob(mainFile, additionalFiles, outputPath)
			}
			return jobs
		},
		Execute: func(data interface{}) error {
			// 模拟并发执行
			return nil
		},
		Verify: func(data interface{}, err error) bool {
			return err == nil
		},
	}
}


