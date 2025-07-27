package file

import (
	"os"
	"time"
)

// FileInfo 包含文件的基本信息
type FileInfo struct {
	Name    string
	Size    int64
	Path    string
	IsValid bool
}

// FileManager 定义文件管理的核心功能接口
type FileManager interface {
	// ValidateFile 验证文件是否存在且可访问
	ValidateFile(filePath string) error

	// CreateTempFile 创建临时文件
	CreateTempFile() (string, error)

	// CreateTempFileWithPrefix 创建带前缀的临时文件
	CreateTempFileWithPrefix(prefix string, suffix string) (string, *os.File, error)

	// CreateTempFileWithContent 创建带内容的临时文件
	CreateTempFileWithContent(prefix string, suffix string, content []byte) (string, error)

	// CopyToTempFile 将源文件复制到临时文件
	CopyToTempFile(sourcePath string, prefix string) (string, error)

	// CleanupTempFiles 清理所有临时文件
	CleanupTempFiles() error

	// RemoveTempFile 删除指定的临时文件
	RemoveTempFile(filePath string) error

	// GetFileInfo 获取文件的基本信息
	GetFileInfo(filePath string) (*FileInfo, error)

	// EnsureDirectoryExists 确保目录存在，如不存在则创建
	EnsureDirectoryExists(dirPath string) error

	// GetTempDir 获取临时目录路径
	GetTempDir() string

	// SetTempFileMaxAge 设置临时文件的最大保留时间
	SetTempFileMaxAge(duration time.Duration)

	// CopyFile 复制文件
	CopyFile(sourcePath, destPath string) error

	// WriteFile 写入文件
	WriteFile(filePath string, data []byte) error

	// ReadFile 读取文件
	ReadFile(filePath string) ([]byte, error)
}
