package model

import (
	"fmt"
	"path/filepath"
	"time"
)

// JobStatus 定义合并任务的状态
type JobStatus int

const (
	// JobPending 表示任务等待执行
	JobPending JobStatus = iota
	// JobRunning 表示任务正在执行
	JobRunning
	// JobCompleted 表示任务已完成
	JobCompleted
	// JobFailed 表示任务失败
	JobFailed
)

// String 返回JobStatus的字符串表示
func (js JobStatus) String() string {
	switch js {
	case JobPending:
		return "等待中"
	case JobRunning:
		return "执行中"
	case JobCompleted:
		return "已完成"
	case JobFailed:
		return "失败"
	default:
		return "未知状态"
	}
}

// MergeJob 定义PDF合并任务
type MergeJob struct {
	ID              string
	MainFile        string
	AdditionalFiles []string
	OutputPath      string
	Status          JobStatus
	Progress        float64
	Error           error
	CreatedAt       time.Time
	CompletedAt     *time.Time
}

// NewMergeJob 创建一个新的合并任务
func NewMergeJob(mainFile string, additionalFiles []string, outputPath string) *MergeJob {
	return &MergeJob{
		ID:              generateJobID(),
		MainFile:        mainFile,
		AdditionalFiles: additionalFiles,
		OutputPath:      outputPath,
		Status:          JobPending,
		Progress:        0.0,
		CreatedAt:       time.Now(),
	}
}

// SetCompleted 标记任务为已完成
func (mj *MergeJob) SetCompleted() {
	mj.Status = JobCompleted
	mj.Progress = 100.0
	now := time.Now()
	mj.CompletedAt = &now
}

// SetFailed 标记任务为失败
func (mj *MergeJob) SetFailed(err error) {
	mj.Status = JobFailed
	mj.Error = err
	now := time.Now()
	mj.CompletedAt = &now
}

// SetRunning 标记任务为运行中
func (mj *MergeJob) SetRunning() {
	mj.Status = JobRunning
}

// UpdateProgress 更新任务进度
func (mj *MergeJob) UpdateProgress(progress float64) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	mj.Progress = progress
}

// GetTotalFiles 获取总文件数
func (mj *MergeJob) GetTotalFiles() int {
	return 1 + len(mj.AdditionalFiles) // 主文件 + 附加文件
}

// FileEntry 定义文件列表中的条目
type FileEntry struct {
	Path        string
	DisplayName string
	Size        int64
	PageCount   int
	IsEncrypted bool
	IsValid     bool
	Order       int
	Error       string // 文件处理错误信息
}

// NewFileEntry 创建一个新的文件条目
func NewFileEntry(path string, order int) *FileEntry {
	return &FileEntry{
		Path:        path,
		DisplayName: filepath.Base(path),
		Order:       order,
		IsValid:     true,
	}
}

// SetError 设置文件错误信息
func (fe *FileEntry) SetError(err string) {
	fe.Error = err
	fe.IsValid = false
}

// GetSizeString 获取文件大小的字符串表示
func (fe *FileEntry) GetSizeString() string {
	if fe.Size < 1024 {
		return fmt.Sprintf("%d B", fe.Size)
	} else if fe.Size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(fe.Size)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(fe.Size)/(1024*1024))
	}
}

// Config 定义应用程序配置
type Config struct {
	MaxMemoryUsage    int64    // 最大内存使用量 (bytes)
	TempDirectory     string   // 临时文件目录
	CommonPasswords   []string // 常用密码列表
	OutputDirectory   string   // 默认输出目录
	EnableAutoDecrypt bool     // 是否启用自动解密
	WindowWidth       int      // 窗口宽度
	WindowHeight      int      // 窗口高度
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		TempDirectory:     "",                 // 将使用系统临时目录
		CommonPasswords:   getDefaultPasswords(),
		OutputDirectory:   "",                 // 将使用用户文档目录
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
	}
}

// getDefaultPasswords 返回默认的常用密码列表
func getDefaultPasswords() []string {
	return []string{
		"", // 空密码
		"123456",
		"password",
		"123456789",
		"12345678",
		"12345",
		"1234567",
		"admin",
		"123123",
		"qwerty",
		"abc123",
		"Password",
		"123",
		"1234",
		"pdf",
		"PDF",
	}
}

// generateJobID 生成唯一的任务ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}