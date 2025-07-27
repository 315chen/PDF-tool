package test_utils

import (
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func TestTestDataFactory_CreateValidFileEntry(t *testing.T) {
	factory := NewTestDataFactory()
	
	path := "/test/file.pdf"
	entry := factory.CreateValidFileEntry(path)
	
	if entry.Path != path {
		t.Errorf("Expected path %s, got %s", path, entry.Path)
	}
	
	if !entry.IsValid {
		t.Error("Expected entry to be valid")
	}
	
	if entry.IsEncrypted {
		t.Error("Expected entry to not be encrypted")
	}
	
	if entry.Size <= 0 {
		t.Error("Expected positive size")
	}
	
	if entry.PageCount <= 0 {
		t.Error("Expected positive page count")
	}
}

func TestTestDataFactory_CreateEncryptedFileEntry(t *testing.T) {
	factory := NewTestDataFactory()
	
	path := "/test/encrypted.pdf"
	entry := factory.CreateEncryptedFileEntry(path)
	
	if entry.Path != path {
		t.Errorf("Expected path %s, got %s", path, entry.Path)
	}
	
	if !entry.IsValid {
		t.Error("Expected entry to be valid")
	}
	
	if !entry.IsEncrypted {
		t.Error("Expected entry to be encrypted")
	}
}

func TestTestDataFactory_CreateInvalidFileEntry(t *testing.T) {
	factory := NewTestDataFactory()
	
	path := "/test/invalid.pdf"
	errorMsg := "File is corrupted"
	entry := factory.CreateInvalidFileEntry(path, errorMsg)
	
	if entry.Path != path {
		t.Errorf("Expected path %s, got %s", path, entry.Path)
	}
	
	if entry.IsValid {
		t.Error("Expected entry to be invalid")
	}
	
	if entry.Size != 0 {
		t.Errorf("Expected size 0, got %d", entry.Size)
	}
	
	if entry.PageCount != 0 {
		t.Errorf("Expected page count 0, got %d", entry.PageCount)
	}
}

func TestTestDataFactory_CreateLargeFileEntry(t *testing.T) {
	factory := NewTestDataFactory()
	
	path := "/test/large.pdf"
	sizeMB := 100
	entry := factory.CreateLargeFileEntry(path, sizeMB)
	
	expectedSize := int64(sizeMB) * 1024 * 1024
	if entry.Size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, entry.Size)
	}
	
	expectedPageCount := sizeMB * 10
	if entry.PageCount != expectedPageCount {
		t.Errorf("Expected page count %d, got %d", expectedPageCount, entry.PageCount)
	}
}

func TestTestDataFactory_CreateFileEntryList(t *testing.T) {
	factory := NewTestDataFactory()
	
	count := 5
	entries := factory.CreateFileEntryList(count)
	
	if len(entries) != count {
		t.Errorf("Expected %d entries, got %d", count, len(entries))
	}
	
	for i, entry := range entries {
		if entry.Order != i {
			t.Errorf("Expected order %d for entry %d, got %d", i, i, entry.Order)
		}
		
		if !entry.IsValid {
			t.Errorf("Expected entry %d to be valid", i)
		}
	}
}

func TestTestDataFactory_CreateMixedFileEntryList(t *testing.T) {
	factory := NewTestDataFactory()
	
	validCount := 3
	invalidCount := 2
	encryptedCount := 1
	
	entries := factory.CreateMixedFileEntryList(validCount, invalidCount, encryptedCount)
	
	totalCount := validCount + invalidCount + encryptedCount
	if len(entries) != totalCount {
		t.Errorf("Expected %d entries, got %d", totalCount, len(entries))
	}
	
	// 验证有效文件
	validFound := 0
	invalidFound := 0
	encryptedFound := 0
	
	for _, entry := range entries {
		if entry.IsValid && !entry.IsEncrypted {
			validFound++
		} else if !entry.IsValid {
			invalidFound++
		} else if entry.IsEncrypted {
			encryptedFound++
		}
	}
	
	if validFound != validCount {
		t.Errorf("Expected %d valid entries, got %d", validCount, validFound)
	}
	
	if invalidFound != invalidCount {
		t.Errorf("Expected %d invalid entries, got %d", invalidCount, invalidFound)
	}
	
	if encryptedFound != encryptedCount {
		t.Errorf("Expected %d encrypted entries, got %d", encryptedCount, encryptedFound)
	}
}

func TestTestDataFactory_CreatePendingMergeJob(t *testing.T) {
	factory := NewTestDataFactory()
	
	mainFile := "/test/main.pdf"
	additionalFiles := []string{"/test/file1.pdf", "/test/file2.pdf"}
	outputPath := "/test/output.pdf"
	
	job := factory.CreatePendingMergeJob(mainFile, additionalFiles, outputPath)
	
	if job.MainFile != mainFile {
		t.Errorf("Expected main file %s, got %s", mainFile, job.MainFile)
	}
	
	if len(job.AdditionalFiles) != len(additionalFiles) {
		t.Errorf("Expected %d additional files, got %d", len(additionalFiles), len(job.AdditionalFiles))
	}
	
	if job.OutputPath != outputPath {
		t.Errorf("Expected output path %s, got %s", outputPath, job.OutputPath)
	}
	
	if job.Status != model.JobPending {
		t.Errorf("Expected status %v, got %v", model.JobPending, job.Status)
	}
	
	if job.Progress != 0.0 {
		t.Errorf("Expected progress 0.0, got %f", job.Progress)
	}
}

func TestTestDataFactory_CreateRunningMergeJob(t *testing.T) {
	factory := NewTestDataFactory()
	
	mainFile := "/test/main.pdf"
	additionalFiles := []string{"/test/file1.pdf"}
	outputPath := "/test/output.pdf"
	progress := 50.0
	
	job := factory.CreateRunningMergeJob(mainFile, additionalFiles, outputPath, progress)
	
	if job.Status != model.JobRunning {
		t.Errorf("Expected status %v, got %v", model.JobRunning, job.Status)
	}
	
	if job.Progress != progress {
		t.Errorf("Expected progress %f, got %f", progress, job.Progress)
	}
	
	// StartedAt字段不存在，跳过检查
}

func TestTestDataFactory_CreateCompletedMergeJob(t *testing.T) {
	factory := NewTestDataFactory()
	
	mainFile := "/test/main.pdf"
	additionalFiles := []string{"/test/file1.pdf"}
	outputPath := "/test/output.pdf"
	
	job := factory.CreateCompletedMergeJob(mainFile, additionalFiles, outputPath)
	
	if job.Status != model.JobCompleted {
		t.Errorf("Expected status %v, got %v", model.JobCompleted, job.Status)
	}
	
	if job.Progress != 100.0 {
		t.Errorf("Expected progress 100.0, got %f", job.Progress)
	}
	
	if job.CompletedAt == nil {
		t.Error("Expected completed time to be set")
	}

	if job.CompletedAt.Before(job.CreatedAt) {
		t.Error("Completed time should be after created time")
	}
}

func TestTestDataFactory_CreateFailedMergeJob(t *testing.T) {
	factory := NewTestDataFactory()
	
	mainFile := "/test/main.pdf"
	additionalFiles := []string{"/test/file1.pdf"}
	outputPath := "/test/output.pdf"
	errorMsg := "Test error message"
	
	job := factory.CreateFailedMergeJob(mainFile, additionalFiles, outputPath, errorMsg)
	
	if job.Status != model.JobFailed {
		t.Errorf("Expected status %v, got %v", model.JobFailed, job.Status)
	}
	
	if job.Error == nil {
		t.Error("Expected error to be set")
	}

	if job.Error.Error() != errorMsg {
		t.Errorf("Expected error message %s, got %s", errorMsg, job.Error.Error())
	}
	
	if job.CompletedAt == nil {
		t.Error("Expected completed time to be set")
	}
}

func TestTestDataFactory_CreateCancelledMergeJob(t *testing.T) {
	factory := NewTestDataFactory()
	
	mainFile := "/test/main.pdf"
	additionalFiles := []string{"/test/file1.pdf"}
	outputPath := "/test/output.pdf"
	
	job := factory.CreateCancelledMergeJob(mainFile, additionalFiles, outputPath)
	
	if job.Status != model.JobFailed {
		t.Errorf("Expected status %v, got %v", model.JobFailed, job.Status)
	}
	
	if job.CompletedAt == nil {
		t.Error("Expected completed time to be set")
	}
}

func TestTestDataFactory_CreateTestConfig(t *testing.T) {
	factory := NewTestDataFactory()
	
	config := factory.CreateTestConfig()
	
	if config.TempDirectory != "/tmp/test" {
		t.Errorf("Expected temp directory '/tmp/test', got %s", config.TempDirectory)
	}
	
	if config.OutputDirectory != "/tmp/output" {
		t.Errorf("Expected output directory '/tmp/output', got %s", config.OutputDirectory)
	}
	
	if config.MaxMemoryUsage != 256*1024*1024 {
		t.Errorf("Expected max memory 256MB, got %d", config.MaxMemoryUsage)
	}
	
	if !config.EnableAutoDecrypt {
		t.Error("Expected auto decrypt to be enabled")
	}
}

func TestTestDataFactory_CreateCustomConfig(t *testing.T) {
	factory := NewTestDataFactory()
	
	tempDir := "/custom/temp"
	outputDir := "/custom/output"
	maxMemory := int64(512 * 1024 * 1024) // 512MB
	
	config := factory.CreateCustomConfig(tempDir, outputDir, maxMemory)
	
	if config.TempDirectory != tempDir {
		t.Errorf("Expected temp directory %s, got %s", tempDir, config.TempDirectory)
	}
	
	if config.OutputDirectory != outputDir {
		t.Errorf("Expected output directory %s, got %s", outputDir, config.OutputDirectory)
	}
	
	if config.MaxMemoryUsage != maxMemory {
		t.Errorf("Expected max memory %d, got %d", maxMemory, config.MaxMemoryUsage)
	}
}

func TestTestDataFactory_CreateFileEntrySlice(t *testing.T) {
	factory := NewTestDataFactory()

	filePaths := []string{"/test/main.pdf", "/test/file1.pdf", "/test/file2.pdf"}
	entries := factory.CreateFileEntrySlice(filePaths)

	if len(entries) != len(filePaths) {
		t.Errorf("Expected %d entries, got %d", len(filePaths), len(entries))
	}

	for i, entry := range entries {
		if entry.Path != filePaths[i] {
			t.Errorf("Expected path %s at index %d, got %s", filePaths[i], i, entry.Path)
		}

		if entry.Order != i {
			t.Errorf("Expected order %d at index %d, got %d", i, i, entry.Order)
		}
	}
}

func TestTestScenarioBuilder_BuildNormalMergeScenario(t *testing.T) {
	builder := NewTestScenarioBuilder()
	
	scenario := builder.BuildNormalMergeScenario()
	
	if scenario.Name != "正常合并场景" {
		t.Errorf("Expected scenario name '正常合并场景', got %s", scenario.Name)
	}
	
	// 测试场景设置
	data := scenario.Setup()
	if data == nil {
		t.Error("Expected scenario data to be set")
	}
	
	// 测试场景执行
	err := scenario.Execute(data)
	if err != nil {
		t.Errorf("Unexpected error in normal scenario: %v", err)
	}
	
	// 测试场景验证
	if !scenario.Verify(data, err) {
		t.Error("Normal scenario verification failed")
	}
}

func TestTestScenarioBuilder_BuildErrorScenario(t *testing.T) {
	builder := NewTestScenarioBuilder()
	
	errorType := "文件不存在"
	scenario := builder.BuildErrorScenario(errorType)
	
	expectedName := errorType + "错误场景"
	if scenario.Name != expectedName {
		t.Errorf("Expected scenario name %s, got %s", expectedName, scenario.Name)
	}
	
	// 测试场景设置
	data := scenario.Setup()
	if data == nil {
		t.Error("Expected scenario data to be set")
	}
	
	// 测试场景执行
	err := scenario.Execute(data)
	if err == nil {
		t.Error("Expected error in error scenario")
	}
	
	// 测试场景验证
	if !scenario.Verify(data, err) {
		t.Error("Error scenario verification failed")
	}
}

func TestTestScenarioBuilder_BuildPerformanceScenario(t *testing.T) {
	builder := NewTestScenarioBuilder()
	
	fileCount := 5
	expectedDuration := 100 * time.Millisecond
	scenario := builder.BuildPerformanceScenario(fileCount, expectedDuration)
	
	expectedName := "性能测试场景_5文件"
	if scenario.Name != expectedName {
		t.Errorf("Expected scenario name %s, got %s", expectedName, scenario.Name)
	}
	
	// 测试场景设置
	data := scenario.Setup()
	if data == nil {
		t.Error("Expected scenario data to be set")
	}
	
	// 测试场景执行（应该有延迟）
	start := time.Now()
	err := scenario.Execute(data)
	duration := time.Since(start)
	
	if err != nil {
		t.Errorf("Unexpected error in performance scenario: %v", err)
	}
	
	if duration < expectedDuration {
		t.Errorf("Expected duration at least %v, got %v", expectedDuration, duration)
	}
	
	// 测试场景验证
	if !scenario.Verify(data, err) {
		t.Error("Performance scenario verification failed")
	}
}

func TestTestScenarioBuilder_BuildConcurrencyScenario(t *testing.T) {
	builder := NewTestScenarioBuilder()
	
	concurrentCount := 3
	scenario := builder.BuildConcurrencyScenario(concurrentCount)
	
	expectedName := "并发测试场景_3并发"
	if scenario.Name != expectedName {
		t.Errorf("Expected scenario name %s, got %s", expectedName, scenario.Name)
	}
	
	// 测试场景设置
	data := scenario.Setup()
	if data == nil {
		t.Error("Expected scenario data to be set")
	}
	
	// 验证数据类型
	jobs, ok := data.([]*model.MergeJob)
	if !ok {
		t.Error("Expected data to be slice of MergeJob")
	}
	
	if len(jobs) != concurrentCount {
		t.Errorf("Expected %d jobs, got %d", concurrentCount, len(jobs))
	}
	
	// 测试场景执行
	err := scenario.Execute(data)
	if err != nil {
		t.Errorf("Unexpected error in concurrency scenario: %v", err)
	}
	
	// 测试场景验证
	if !scenario.Verify(data, err) {
		t.Error("Concurrency scenario verification failed")
	}
}
