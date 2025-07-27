package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileManagerImpl 实现FileManager接口
type FileManagerImpl struct {
	tempManager *TempFileManager
}

// NewFileManager 创建一个新的文件管理器实例
func NewFileManager(tempDir string) FileManager {
	// 创建临时文件管理器
	tempManager, err := NewTempFileManager(tempDir)
	if err != nil {
		// 如果创建失败，使用默认临时目录
		tempManager, _ = NewTempFileManager("")
	}

	return &FileManagerImpl{
		tempManager: tempManager,
	}
}

// ValidateFile 验证文件是否存在且可访问
func (fm *FileManagerImpl) ValidateFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("无法访问文件: %v", err)
	}

	// 检查是否为文件而不是目录
	if info.IsDir() {
		return fmt.Errorf("路径指向目录而不是文件: %s", filePath)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".pdf" {
		return fmt.Errorf("不支持的文件格式: %s (仅支持PDF文件)", ext)
	}

	// 检查文件大小
	if info.Size() == 0 {
		return fmt.Errorf("文件为空: %s", filePath)
	}

	// 检查文件是否可读
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法读取文件: %v", err)
	}
	file.Close()

	return nil
}

// CreateTempFile 创建临时文件
func (fm *FileManagerImpl) CreateTempFile() (string, error) {
	filePath, _, err := fm.tempManager.CreateTempFile("pdf_", ".tmp")
	return filePath, err
}

// CreateTempFileWithPrefix 创建带前缀的临时文件
func (fm *FileManagerImpl) CreateTempFileWithPrefix(prefix string, suffix string) (string, *os.File, error) {
	return fm.tempManager.CreateTempFile(prefix, suffix)
}

// CreateTempFileWithContent 创建带内容的临时文件
func (fm *FileManagerImpl) CreateTempFileWithContent(prefix string, suffix string, content []byte) (string, error) {
	return fm.tempManager.CreateTempFileWithContent(prefix, suffix, content)
}

// CopyToTempFile 将源文件复制到临时文件
func (fm *FileManagerImpl) CopyToTempFile(sourcePath string, prefix string) (string, error) {
	return fm.tempManager.CopyToTempFile(sourcePath, prefix)
}

// CleanupTempFiles 清理所有临时文件
func (fm *FileManagerImpl) CleanupTempFiles() error {
	fm.tempManager.Cleanup()
	return nil
}

// RemoveTempFile 删除指定的临时文件
func (fm *FileManagerImpl) RemoveTempFile(filePath string) error {
	return fm.tempManager.RemoveFile(filePath)
}

// GetFileInfo 获取文件的基本信息
func (fm *FileManagerImpl) GetFileInfo(filePath string) (*FileInfo, error) {
	// 首先验证文件
	if err := fm.ValidateFile(filePath); err != nil {
		return &FileInfo{
			Name:    filepath.Base(filePath),
			Size:    0,
			Path:    filePath,
			IsValid: false,
		}, err
	}

	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %v", err)
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Path:    filePath,
		IsValid: true,
	}, nil
}

// EnsureDirectoryExists 确保目录存在，如不存在则创建
func (fm *FileManagerImpl) EnsureDirectoryExists(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	// 检查目录是否已存在
	if DirExists(dirPath) {
		return nil
	}

	// 创建目录
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("无法创建目录 %s: %v", dirPath, err)
	}

	return nil
}

// GetTempDir 获取临时目录路径
func (fm *FileManagerImpl) GetTempDir() string {
	return fm.tempManager.GetSessionDir()
}

// SetTempFileMaxAge 设置临时文件的最大保留时间
func (fm *FileManagerImpl) SetTempFileMaxAge(duration time.Duration) {
	fm.tempManager.SetMaxAge(duration)
}

// CopyFile 复制文件
func (fm *FileManagerImpl) CopyFile(sourcePath, destPath string) error {
	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %v", err)
	}
	defer sourceFile.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("无法创建目标文件: %v", err)
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("无法复制文件内容: %v", err)
	}

	return nil
}

// WriteFile 写入文件
func (fm *FileManagerImpl) WriteFile(filePath string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := fm.EnsureDirectoryExists(dir); err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("无法写入文件: %v", err)
	}

	return nil
}

// ReadFile 读取文件
func (fm *FileManagerImpl) ReadFile(filePath string) ([]byte, error) {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件: %v", err)
	}

	return data, nil
}