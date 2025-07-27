package test_utils

import (
	"testing"

	"github.com/user/pdf-merger/internal/model"
)

func TestMockDataGenerator_GenerateFileEntry(t *testing.T) {
	// 测试生成模拟文件条目
	generator := NewMockDataGenerator()

	entry := generator.GenerateFileEntry()

	// 验证基本属性
	if entry.Path == "" {
		t.Error("Expected non-empty path")
	}

	if entry.DisplayName == "" {
		t.Error("Expected non-empty display name")
	}

	if entry.Size < 0 {
		t.Error("Expected non-negative size")
	}

	if entry.PageCount <= 0 {
		t.Error("Expected positive page count")
	}

	// 验证Order字段
	if entry.Order < 0 {
		t.Error("Expected non-negative order")
	}
}

func TestMockDataGenerator_GenerateFileEntries(t *testing.T) {
	// 测试生成多个模拟文件条目
	generator := NewMockDataGenerator()
	count := 5

	entries := generator.GenerateFileEntries(count)

	// 验证数量
	if len(entries) != count {
		t.Errorf("Expected %d entries, got %d", count, len(entries))
	}

	// 验证每个条目
	for i, entry := range entries {
		if entry == nil {
			t.Errorf("Entry %d should not be nil", i)
		}

		if entry.Path == "" {
			t.Errorf("Entry %d should have non-empty path", i)
		}

		if entry.Order != i {
			t.Errorf("Entry %d should have order %d, got %d", i, i, entry.Order)
		}
	}
}

func TestMockDataGenerator_GenerateMergeJob(t *testing.T) {
	// 测试生成模拟合并任务
	generator := NewMockDataGenerator()

	job := generator.GenerateMergeJob()

	// 验证基本属性
	if job.ID == "" {
		t.Error("Expected non-empty job ID")
	}

	if job.MainFile == "" {
		t.Error("Expected non-empty main file")
	}

	if job.OutputPath == "" {
		t.Error("Expected non-empty output path")
	}

	if len(job.AdditionalFiles) == 0 {
		t.Error("Expected at least one additional file")
	}

	// 验证时间戳
	if job.CreatedAt.IsZero() {
		t.Error("Expected non-zero creation time")
	}
}

func TestMockDataGenerator_GenerateConfig(t *testing.T) {
	// 测试生成模拟配置
	generator := NewMockDataGenerator()

	config := generator.GenerateConfig()

	// 验证基本配置
	if config.TempDirectory == "" {
		t.Error("Expected non-empty temp directory")
	}

	if config.MaxMemoryUsage <= 0 {
		t.Error("Expected positive max memory usage")
	}

	// 验证其他配置字段存在
	if config.MaxMemoryUsage <= 0 {
		t.Error("Expected positive max memory usage")
	}
}

func TestMockDataGenerator_RandomValues(t *testing.T) {
	// 测试生成器的随机性
	generator := NewMockDataGenerator()

	// 生成多个文件条目，验证它们不完全相同
	entries := make([]*model.FileEntry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = generator.GenerateFileEntry()
	}

	// 验证至少有一些不同的值
	pathSet := make(map[string]bool)
	sizeSet := make(map[int64]bool)

	for _, entry := range entries {
		pathSet[entry.Path] = true
		sizeSet[entry.Size] = true
	}

	// 应该有多个不同的路径和大小
	if len(pathSet) < 2 {
		t.Error("Expected multiple different paths")
	}

	if len(sizeSet) < 2 {
		t.Error("Expected multiple different sizes")
	}
}

func TestMockDataGenerator_Consistency(t *testing.T) {
	// 测试生成器的一致性
	generator := NewMockDataGenerator()

	// 生成多个任务
	jobs := make([]*model.MergeJob, 5)
	for i := 0; i < 5; i++ {
		jobs[i] = generator.GenerateMergeJob()
	}

	// 验证每个任务都有必需的字段
	for i, job := range jobs {
		if job.ID == "" {
			t.Errorf("Job %d should have non-empty ID", i)
		}

		if job.MainFile == "" {
			t.Errorf("Job %d should have non-empty main file", i)
		}

		if job.OutputPath == "" {
			t.Errorf("Job %d should have non-empty output path", i)
		}

		if len(job.AdditionalFiles) == 0 {
			t.Errorf("Job %d should have additional files", i)
		}
	}
}

// 基准测试
func BenchmarkMockDataGenerator_GenerateFileEntry(b *testing.B) {
	generator := NewMockDataGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateFileEntry()
	}
}

func BenchmarkMockDataGenerator_GenerateMergeJob(b *testing.B) {
	generator := NewMockDataGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateMergeJob()
	}
}

func BenchmarkMockDataGenerator_GenerateFileEntries(b *testing.B) {
	generator := NewMockDataGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateFileEntries(10)
	}
}












