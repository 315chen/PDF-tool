package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OutputManager 输出路径管理器
type OutputManager struct {
	baseDir         string
	defaultFileName string
	autoIncrement   bool
	timestampSuffix bool
	backupEnabled   bool
}

// OutputOptions 输出选项
type OutputOptions struct {
	BaseDirectory   string // 基础输出目录
	DefaultFileName string // 默认文件名
	AutoIncrement   bool   // 自动递增文件名
	TimestampSuffix bool   // 添加时间戳后缀
	BackupEnabled   bool   // 启用备份
}

// OutputInfo 输出信息
type OutputInfo struct {
	FinalPath     string
	OriginalPath  string
	BackupPath    string
	IsIncremented bool
	HasTimestamp  bool
}

// NewOutputManager 创建输出路径管理器
func NewOutputManager(options *OutputOptions) *OutputManager {
	if options == nil {
		options = &OutputOptions{
			BaseDirectory:   ".",
			DefaultFileName: "merged_output.pdf",
			AutoIncrement:   true,
			TimestampSuffix: false,
			BackupEnabled:   true,
		}
	}

	return &OutputManager{
		baseDir:         options.BaseDirectory,
		defaultFileName: options.DefaultFileName,
		autoIncrement:   options.AutoIncrement,
		timestampSuffix: options.TimestampSuffix,
		backupEnabled:   options.BackupEnabled,
	}
}

// ResolveOutputPath 解析输出路径
func (om *OutputManager) ResolveOutputPath(requestedPath string) (*OutputInfo, error) {
	info := &OutputInfo{
		OriginalPath: requestedPath,
	}

	// 如果没有提供路径，使用默认路径
	if requestedPath == "" {
		requestedPath = filepath.Join(om.baseDir, om.defaultFileName)
	}

	// 确保是绝对路径
	if !filepath.IsAbs(requestedPath) {
		requestedPath = filepath.Join(om.baseDir, requestedPath)
	}

	// 验证路径
	if err := om.validatePath(requestedPath); err != nil {
		return nil, err
	}

	// 处理时间戳后缀
	if om.timestampSuffix {
		requestedPath = om.addTimestampSuffix(requestedPath)
		info.HasTimestamp = true
	}

	// 处理自动递增
	if om.autoIncrement {
		finalPath, incremented := om.resolveAutoIncrement(requestedPath)
		info.FinalPath = finalPath
		info.IsIncremented = incremented
	} else {
		info.FinalPath = requestedPath
	}

	// 处理备份
	if om.backupEnabled && fileExists(info.FinalPath) {
		info.BackupPath = om.generateBackupPath(info.FinalPath)
	}

	return info, nil
}

// validatePath 验证路径
func (om *OutputManager) validatePath(path string) error {
	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "输出文件必须是PDF格式",
			File:    path,
		}
	}

	// 检查目录权限
	dir := filepath.Dir(path)
	if err := om.ensureDirectoryWritable(dir); err != nil {
		return &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写",
			File:    dir,
			Cause:   err,
		}
	}

	return nil
}

// addTimestampSuffix 添加时间戳后缀
func (om *OutputManager) addTimestampSuffix(path string) string {
	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(path, ext)
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
}

// resolveAutoIncrement 解析自动递增
func (om *OutputManager) resolveAutoIncrement(path string) (string, bool) {
	if !fileExists(path) {
		return path, false
	}

	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(path, ext)

	for i := 1; i <= 9999; i++ {
		newPath := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		if !fileExists(newPath) {
			return newPath, true
		}
	}

	// 如果找不到可用的递增名称，使用时间戳
	timestamp := time.Now().Format("20060102_150405_000")
	return fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext), true
}

// generateBackupPath 生成备份路径
func (om *OutputManager) generateBackupPath(path string) string {
	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(path, ext)
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_backup_%s%s", nameWithoutExt, timestamp, ext)
}

// ensureDirectoryWritable 确保目录可写
func (om *OutputManager) ensureDirectoryWritable(dir string) error {
	// 创建目录（如果不存在）
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 测试写入权限
	testFile := filepath.Join(dir, ".write_test_"+time.Now().Format("20060102150405"))
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// CreateBackup 创建备份文件
func (om *OutputManager) CreateBackup(originalPath, backupPath string) error {
	if !fileExists(originalPath) {
		return nil // 原文件不存在，无需备份
	}

	return copyFile(originalPath, backupPath)
}

// RestoreBackup 恢复备份文件
func (om *OutputManager) RestoreBackup(backupPath, targetPath string) error {
	if !fileExists(backupPath) {
		return &PDFError{
			Type:    ErrorIO,
			Message: "备份文件不存在",
			File:    backupPath,
		}
	}

	return copyFile(backupPath, targetPath)
}

// CleanupBackup 清理备份文件
func (om *OutputManager) CleanupBackup(backupPath string) error {
	if backupPath != "" && fileExists(backupPath) {
		return os.Remove(backupPath)
	}
	return nil
}

// GetSuggestedPath 获取建议的输出路径
func (om *OutputManager) GetSuggestedPath(inputFiles []string) string {
	if len(inputFiles) == 0 {
		return filepath.Join(om.baseDir, om.defaultFileName)
	}

	// 基于第一个输入文件生成建议路径
	firstFile := inputFiles[0]
	fileName := filepath.Base(firstFile)
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	suggestedName := fmt.Sprintf("%s_merged.pdf", nameWithoutExt)
	return filepath.Join(om.baseDir, suggestedName)
}

// ValidateOutputPath 验证输出路径是否有效
func (om *OutputManager) ValidateOutputPath(path string) error {
	return om.validatePath(path)
}

// GetOutputDirectory 获取输出目录
func (om *OutputManager) GetOutputDirectory() string {
	return om.baseDir
}

// SetOutputDirectory 设置输出目录
func (om *OutputManager) SetOutputDirectory(dir string) error {
	if err := om.ensureDirectoryWritable(dir); err != nil {
		return err
	}
	om.baseDir = dir
	return nil
}

// GetDefaultFileName 获取默认文件名
func (om *OutputManager) GetDefaultFileName() string {
	return om.defaultFileName
}

// SetDefaultFileName 设置默认文件名
func (om *OutputManager) SetDefaultFileName(fileName string) error {
	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "默认文件名必须是PDF格式",
			File:    fileName,
		}
	}
	om.defaultFileName = fileName
	return nil
}
