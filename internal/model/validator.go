package model

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError 定义验证错误
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现error接口
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Message)
}

// Validator 定义数据验证器
type Validator struct{}

// NewValidator 创建一个新的验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMergeJob 验证合并任务
func (v *Validator) ValidateMergeJob(job *MergeJob) error {
	if job == nil {
		return errors.New("merge job cannot be nil")
	}

	if job.ID == "" {
		return &ValidationError{Field: "ID", Message: "cannot be empty"}
	}

	if job.MainFile == "" {
		return &ValidationError{Field: "MainFile", Message: "cannot be empty"}
	}

	if !v.isValidFilePath(job.MainFile) {
		return &ValidationError{Field: "MainFile", Message: "invalid file path"}
	}

	if job.OutputPath == "" {
		return &ValidationError{Field: "OutputPath", Message: "cannot be empty"}
	}

	if !v.isValidOutputPath(job.OutputPath) {
		return &ValidationError{Field: "OutputPath", Message: "invalid output path"}
	}

	if job.Progress < 0 || job.Progress > 100 {
		return &ValidationError{Field: "Progress", Message: "must be between 0 and 100"}
	}

	// 验证附加文件
	for i, file := range job.AdditionalFiles {
		if file == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("AdditionalFiles[%d]", i),
				Message: "cannot be empty",
			}
		}
		if !v.isValidFilePath(file) {
			return &ValidationError{
				Field:   fmt.Sprintf("AdditionalFiles[%d]", i),
				Message: "invalid file path",
			}
		}
	}

	return nil
}

// ValidateFileEntry 验证文件条目
func (v *Validator) ValidateFileEntry(entry *FileEntry) error {
	if entry == nil {
		return errors.New("file entry cannot be nil")
	}

	if entry.Path == "" {
		return &ValidationError{Field: "Path", Message: "cannot be empty"}
	}

	if !v.isValidFilePath(entry.Path) {
		return &ValidationError{Field: "Path", Message: "invalid file path"}
	}

	if entry.DisplayName == "" {
		return &ValidationError{Field: "DisplayName", Message: "cannot be empty"}
	}

	if entry.Size < 0 {
		return &ValidationError{Field: "Size", Message: "cannot be negative"}
	}

	if entry.PageCount < 0 {
		return &ValidationError{Field: "PageCount", Message: "cannot be negative"}
	}

	if entry.Order < 0 {
		return &ValidationError{Field: "Order", Message: "cannot be negative"}
	}

	return nil
}

// ValidateConfig 验证配置
func (v *Validator) ValidateConfig(config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.MaxMemoryUsage <= 0 {
		return &ValidationError{Field: "MaxMemoryUsage", Message: "must be positive"}
	}

	// 检查内存使用量是否合理（不超过系统内存的80%）
	if config.MaxMemoryUsage > 8*1024*1024*1024 { // 8GB
		return &ValidationError{Field: "MaxMemoryUsage", Message: "exceeds reasonable limit (8GB)"}
	}

	if config.TempDirectory != "" && !v.isValidDirectoryPath(config.TempDirectory) {
		return &ValidationError{Field: "TempDirectory", Message: "invalid directory path"}
	}

	if config.OutputDirectory != "" && !v.isValidDirectoryPath(config.OutputDirectory) {
		return &ValidationError{Field: "OutputDirectory", Message: "invalid directory path"}
	}

	if config.WindowWidth <= 0 {
		return &ValidationError{Field: "WindowWidth", Message: "must be positive"}
	}

	if config.WindowHeight <= 0 {
		return &ValidationError{Field: "WindowHeight", Message: "must be positive"}
	}

	// 检查窗口大小是否合理
	if config.WindowWidth < 400 || config.WindowWidth > 4000 {
		return &ValidationError{Field: "WindowWidth", Message: "must be between 400 and 4000"}
	}

	if config.WindowHeight < 300 || config.WindowHeight > 3000 {
		return &ValidationError{Field: "WindowHeight", Message: "must be between 300 and 3000"}
	}

	// 验证密码列表
	for i, password := range config.CommonPasswords {
		if len(password) > 100 {
			return &ValidationError{
				Field:   fmt.Sprintf("CommonPasswords[%d]", i),
				Message: "password too long (max 100 characters)",
			}
		}
	}

	return nil
}

// ValidateFileList 验证文件列表
func (v *Validator) ValidateFileList(fileList *FileList) error {
	if fileList == nil {
		return errors.New("file list cannot be nil")
	}

	// 验证主文件
	if mainFile := fileList.GetMainFile(); mainFile != nil {
		if err := v.ValidateFileEntry(mainFile); err != nil {
			return fmt.Errorf("main file validation failed: %w", err)
		}
	}

	// 验证附加文件
	files := fileList.GetFiles()
	for i, file := range files {
		if err := v.ValidateFileEntry(file); err != nil {
			return fmt.Errorf("additional file %d validation failed: %w", i, err)
		}
	}

	// 检查文件路径是否重复
	pathMap := make(map[string]bool)

	if mainFile := fileList.GetMainFile(); mainFile != nil {
		pathMap[mainFile.Path] = true
	}

	for i, file := range files {
		if pathMap[file.Path] {
			return &ValidationError{
				Field:   fmt.Sprintf("Files[%d].Path", i),
				Message: "duplicate file path: " + file.Path,
			}
		}
		pathMap[file.Path] = true
	}

	return nil
}

// isValidFilePath 检查文件路径是否有效
func (v *Validator) isValidFilePath(path string) bool {
	if path == "" {
		return false
	}

	// 特殊处理Windows路径
	if strings.HasPrefix(path, "C:\\") || strings.HasPrefix(path, "c:\\") {
		// Windows路径格式，我们认为它是有效的
		return true
	}

	// 检查路径中是否包含非法字符（排除Windows路径中的冒号）
	invalidChars := "<>\"|?*"
	if !strings.HasPrefix(path, "C:\\") && !strings.HasPrefix(path, "c:\\") {
		invalidChars = "<>:\"|?*"
	}

	if strings.ContainsAny(path, invalidChars) {
		return false
	}

	// 检查是否为绝对路径或相对路径
	if !filepath.IsAbs(path) && !strings.HasPrefix(path, ".") {
		// 相对路径应该以 . 或 .. 开头，或者是简单的文件名
		if !strings.Contains(path, string(filepath.Separator)) {
			return true // 简单文件名
		}
	}

	return filepath.IsAbs(path) || strings.HasPrefix(path, ".")
}

// isValidOutputPath 检查输出路径是否有效
func (v *Validator) isValidOutputPath(path string) bool {
	if !v.isValidFilePath(path) {
		return false
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pdf"
}

// isValidDirectoryPath 检查目录路径是否有效
func (v *Validator) isValidDirectoryPath(path string) bool {
	if path == "" {
		return false
	}

	// 检查路径中是否包含非法字符
	if strings.ContainsAny(path, "<>:\"|?*") {
		return false
	}

	// 尝试创建目录来验证路径是否有效
	if err := os.MkdirAll(path, 0755); err != nil {
		return false
	}

	// 清理测试目录（如果是新创建的）
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		// 只有当目录为空时才删除
		if entries, err := os.ReadDir(path); err == nil && len(entries) == 0 {
			os.Remove(path)
		}
	}

	return true
}

// ValidateProgressTracker 验证进度跟踪器
func (v *Validator) ValidateProgressTracker(tracker *ProgressTracker) error {
	if tracker == nil {
		return errors.New("progress tracker cannot be nil")
	}

	info := tracker.GetProgress()

	if info.TotalSteps <= 0 {
		return &ValidationError{Field: "TotalSteps", Message: "must be positive"}
	}

	if info.CurrentStep < 0 || info.CurrentStep > info.TotalSteps {
		return &ValidationError{Field: "CurrentStep", Message: "must be between 0 and TotalSteps"}
	}

	if info.StepProgress < 0 || info.StepProgress > 100 {
		return &ValidationError{Field: "StepProgress", Message: "must be between 0 and 100"}
	}

	if info.TotalProgress < 0 || info.TotalProgress > 100 {
		return &ValidationError{Field: "TotalProgress", Message: "must be between 0 and 100"}
	}

	return nil
}
