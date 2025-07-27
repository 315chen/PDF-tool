package model

import (
	"testing"
)

func TestValidator_ValidateMergeJob(t *testing.T) {
	validator := NewValidator()

	// 测试有效的合并任务
	validJob := NewMergeJob("/path/to/main.pdf", []string{"/path/to/file1.pdf"}, "/path/to/output.pdf")
	err := validator.ValidateMergeJob(validJob)
	if err != nil {
		t.Errorf("Expected valid job to pass validation, got error: %v", err)
	}

	// 测试nil任务
	err = validator.ValidateMergeJob(nil)
	if err == nil {
		t.Error("Expected error for nil job")
	}

	// 测试空ID
	invalidJob := &MergeJob{
		ID:       "",
		MainFile: "/path/to/main.pdf",
	}
	err = validator.ValidateMergeJob(invalidJob)
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// 测试空主文件
	invalidJob = &MergeJob{
		ID:       "test-id",
		MainFile: "",
	}
	err = validator.ValidateMergeJob(invalidJob)
	if err == nil {
		t.Error("Expected error for empty main file")
	}

	// 测试无效进度
	invalidJob = NewMergeJob("/path/to/main.pdf", []string{}, "/path/to/output.pdf")
	invalidJob.Progress = -10
	err = validator.ValidateMergeJob(invalidJob)
	if err == nil {
		t.Error("Expected error for negative progress")
	}

	invalidJob.Progress = 150
	err = validator.ValidateMergeJob(invalidJob)
	if err == nil {
		t.Error("Expected error for progress > 100")
	}
}

func TestValidator_ValidateFileEntry(t *testing.T) {
	validator := NewValidator()

	// 测试有效的文件条目
	validEntry := NewFileEntry("/path/to/test.pdf", 1)
	validEntry.Size = 1024
	validEntry.PageCount = 10
	err := validator.ValidateFileEntry(validEntry)
	if err != nil {
		t.Errorf("Expected valid entry to pass validation, got error: %v", err)
	}

	// 测试nil条目
	err = validator.ValidateFileEntry(nil)
	if err == nil {
		t.Error("Expected error for nil entry")
	}

	// 测试空路径
	invalidEntry := &FileEntry{
		Path: "",
	}
	err = validator.ValidateFileEntry(invalidEntry)
	if err == nil {
		t.Error("Expected error for empty path")
	}

	// 测试负数大小
	invalidEntry = NewFileEntry("/path/to/test.pdf", 1)
	invalidEntry.Size = -100
	err = validator.ValidateFileEntry(invalidEntry)
	if err == nil {
		t.Error("Expected error for negative size")
	}

	// 测试负数页数
	invalidEntry = NewFileEntry("/path/to/test.pdf", 1)
	invalidEntry.PageCount = -5
	err = validator.ValidateFileEntry(invalidEntry)
	if err == nil {
		t.Error("Expected error for negative page count")
	}
}

func TestValidator_ValidateConfig(t *testing.T) {
	validator := NewValidator()

	// 测试有效配置
	validConfig := DefaultConfig()
	err := validator.ValidateConfig(validConfig)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// 测试nil配置
	err = validator.ValidateConfig(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// 测试无效内存使用量
	invalidConfig := DefaultConfig()
	invalidConfig.MaxMemoryUsage = 0
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for zero memory usage")
	}

	invalidConfig.MaxMemoryUsage = -100
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for negative memory usage")
	}

	// 测试过大的内存使用量
	invalidConfig = DefaultConfig()
	invalidConfig.MaxMemoryUsage = 10 * 1024 * 1024 * 1024 // 10GB
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for excessive memory usage")
	}

	// 测试无效窗口大小
	invalidConfig = DefaultConfig()
	invalidConfig.WindowWidth = 0
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for zero window width")
	}

	invalidConfig = DefaultConfig()
	invalidConfig.WindowWidth = 5000
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for excessive window width")
	}

	// 测试过长的密码
	invalidConfig = DefaultConfig()
	longPassword := make([]byte, 101)
	for i := range longPassword {
		longPassword[i] = 'a'
	}
	invalidConfig.CommonPasswords = append(invalidConfig.CommonPasswords, string(longPassword))
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for overly long password")
	}
}

func TestValidator_ValidateFileList(t *testing.T) {
	validator := NewValidator()

	// 测试有效的文件列表
	validList := NewFileList()
	validList.SetMainFile("/path/to/main.pdf")
	validList.AddFile("/path/to/file1.pdf")
	err := validator.ValidateFileList(validList)
	if err != nil {
		t.Errorf("Expected valid file list to pass validation, got error: %v", err)
	}

	// 测试nil文件列表
	err = validator.ValidateFileList(nil)
	if err == nil {
		t.Error("Expected error for nil file list")
	}

	// 测试重复文件路径
	duplicateList := NewFileList()
	duplicateList.SetMainFile("/path/to/same.pdf")
	duplicateList.AddFile("/path/to/same.pdf") // 与主文件相同
	err = validator.ValidateFileList(duplicateList)
	if err == nil {
		t.Error("Expected error for duplicate file paths")
	}
}

func TestValidator_isValidFilePath(t *testing.T) {
	validator := NewValidator()

	validPaths := []string{
		"/absolute/path/to/file.pdf",
		"./relative/path/to/file.pdf",
		"../parent/file.pdf",
		"simple-file.pdf",
		"C:\\Windows\\path\\file.pdf", // Windows路径在某些系统上可能有效
	}

	for _, path := range validPaths {
		if !validator.isValidFilePath(path) {
			t.Errorf("Expected path '%s' to be valid", path)
		}
	}

	invalidPaths := []string{
		"",
		"path/with<invalid>chars.pdf",
		"path/with|pipe.pdf",
		"path/with\"quote.pdf",
	}

	for _, path := range invalidPaths {
		if validator.isValidFilePath(path) {
			t.Errorf("Expected path '%s' to be invalid", path)
		}
	}
}

func TestValidator_isValidOutputPath(t *testing.T) {
	validator := NewValidator()

	validPaths := []string{
		"/path/to/output.pdf",
		"./output.pdf",
		"../output.PDF", // 大写扩展名也应该有效
	}

	for _, path := range validPaths {
		if !validator.isValidOutputPath(path) {
			t.Errorf("Expected output path '%s' to be valid", path)
		}
	}

	invalidPaths := []string{
		"/path/to/output.txt",
		"/path/to/output",
		"",
		"output.doc",
	}

	for _, path := range invalidPaths {
		if validator.isValidOutputPath(path) {
			t.Errorf("Expected output path '%s' to be invalid", path)
		}
	}
}

func TestValidator_ValidateProgressTracker(t *testing.T) {
	validator := NewValidator()

	// 测试有效的进度跟踪器
	validTracker := NewProgressTracker(5)
	validTracker.SetCurrentStep(2, "Step 2")
	validTracker.UpdateStepProgress(50, "Half done")
	err := validator.ValidateProgressTracker(validTracker)
	if err != nil {
		t.Errorf("Expected valid progress tracker to pass validation, got error: %v", err)
	}

	// 测试nil跟踪器
	err = validator.ValidateProgressTracker(nil)
	if err == nil {
		t.Error("Expected error for nil progress tracker")
	}

	// 测试无效的总步数
	invalidTracker := NewProgressTracker(0)
	err = validator.ValidateProgressTracker(invalidTracker)
	if err == nil {
		t.Error("Expected error for zero total steps")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "TestField",
		Message: "test message",
	}

	expected := "validation error for field 'TestField': test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}