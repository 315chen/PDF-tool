package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// TempFileManager 专门负责临时文件的管理
type TempFileManager struct {
	baseDir      string
	sessionDir   string
	files        map[string]time.Time
	maxAge       time.Duration
	cleanupTimer *time.Timer
	mutex        sync.RWMutex
}

// NewTempFileManager 创建一个新的临时文件管理器
func NewTempFileManager(baseDir string) (*TempFileManager, error) {
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// 创建以应用名称为前缀的临时目录
	baseDir = filepath.Join(baseDir, "pdf-merger-temp")
	
	// 创建会话特定的目录（使用时间戳确保唯一性）
	sessionDir := filepath.Join(baseDir, fmt.Sprintf("session_%d", time.Now().UnixNano()))
	
	// 确保目录存在
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建临时目录: %v", err)
	}

	manager := &TempFileManager{
		baseDir:    baseDir,
		sessionDir: sessionDir,
		files:      make(map[string]time.Time),
		maxAge:     1 * time.Hour, // 默认临时文件最长保留1小时
	}

	// 设置清理定时器
	manager.startCleanupTimer()
	
	// 设置终结器，确保在对象被垃圾回收时清理临时文件
	runtime.SetFinalizer(manager, func(m *TempFileManager) {
		m.Cleanup()
	})

	return manager, nil
}

// CreateTempFile 创建一个新的临时文件
func (tm *TempFileManager) CreateTempFile(prefix string, suffix string) (string, *os.File, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if prefix == "" {
		prefix = "pdf_"
	}
	
	if suffix == "" {
		suffix = ".tmp"
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(tm.sessionDir, prefix+"*"+suffix)
	if err != nil {
		return "", nil, fmt.Errorf("无法创建临时文件: %v", err)
	}

	// 记录文件创建时间
	tm.files[tempFile.Name()] = time.Now()

	return tempFile.Name(), tempFile, nil
}

// CreateTempFileWithContent 创建一个带有指定内容的临时文件
func (tm *TempFileManager) CreateTempFileWithContent(prefix string, suffix string, content []byte) (string, error) {
	filePath, file, err := tm.CreateTempFile(prefix, suffix)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入内容
	if _, err := file.Write(content); err != nil {
		os.Remove(filePath) // 如果写入失败，删除文件
		return "", fmt.Errorf("无法写入临时文件: %v", err)
	}

	return filePath, nil
}

// CopyToTempFile 将源文件复制到临时文件
func (tm *TempFileManager) CopyToTempFile(sourcePath string, prefix string) (string, error) {
	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("无法打开源文件: %v", err)
	}
	defer sourceFile.Close()

	// 获取源文件扩展名
	ext := filepath.Ext(sourcePath)
	if ext == "" {
		ext = ".tmp"
	}

	// 创建临时文件
	destPath, destFile, err := tm.CreateTempFile(prefix, ext)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		os.Remove(destPath) // 如果复制失败，删除临时文件
		return "", fmt.Errorf("无法复制文件内容: %v", err)
	}

	return destPath, nil
}

// RemoveFile 删除指定的临时文件
func (tm *TempFileManager) RemoveFile(filePath string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 检查文件是否在我们的管理范围内
	if !tm.isOwnedFile(filePath) {
		return fmt.Errorf("文件不在临时目录中: %s", filePath)
	}

	// 检查文件是否存在
	if _, ok := tm.files[filePath]; !ok {
		// 检查文件系统中是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", filePath)
		}
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", filePath)
		}
		return fmt.Errorf("无法删除临时文件: %v", err)
	}

	// 从记录中移除
	delete(tm.files, filePath)

	return nil
}

// Cleanup 清理所有临时文件
func (tm *TempFileManager) Cleanup() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 删除会话目录中的所有文件
	if err := os.RemoveAll(tm.sessionDir); err != nil {
		fmt.Printf("警告: 无法删除临时目录 %s: %v\n", tm.sessionDir, err)
	}

	// 清空文件记录
	tm.files = make(map[string]time.Time)
}

// CleanupExpired 清理过期的临时文件
func (tm *TempFileManager) CleanupExpired() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	for filePath, creationTime := range tm.files {
		if now.Sub(creationTime) > tm.maxAge {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("警告: 无法删除过期临时文件 %s: %v\n", filePath, err)
			}
			delete(tm.files, filePath)
		}
	}

	// 清理其他会话的过期目录
	tm.cleanupOldSessions()
}

// cleanupOldSessions 清理旧的会话目录
func (tm *TempFileManager) cleanupOldSessions() {
	// 获取基础目录中的所有条目
	entries, err := os.ReadDir(tm.baseDir)
	if err != nil {
		return
	}

	currentSession := filepath.Base(tm.sessionDir)
	now := time.Now()

	for _, entry := range entries {
		// 跳过当前会话目录
		if entry.Name() == currentSession || !entry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(tm.baseDir, entry.Name())
		info, err := os.Stat(sessionPath)
		if err != nil {
			continue
		}

		// 如果目录超过最大年龄，则删除
		if now.Sub(info.ModTime()) > tm.maxAge {
			os.RemoveAll(sessionPath)
		}
	}
}

// GetSessionDir 获取当前会话的临时目录
func (tm *TempFileManager) GetSessionDir() string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.sessionDir
}

// GetFileCount 获取当前管理的临时文件数量
func (tm *TempFileManager) GetFileCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.files)
}

// SetMaxAge 设置临时文件的最大保留时间
func (tm *TempFileManager) SetMaxAge(duration time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.maxAge = duration
}

// startCleanupTimer 启动定期清理定时器
func (tm *TempFileManager) startCleanupTimer() {
	tm.cleanupTimer = time.AfterFunc(10*time.Minute, func() {
		tm.CleanupExpired()
		tm.startCleanupTimer() // 重新设置定时器
	})
}

// isOwnedFile 检查文件是否在当前会话的临时目录中
func (tm *TempFileManager) isOwnedFile(filePath string) bool {
	// 规范化路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	
	// 检查文件是否在会话目录中
	return strings.HasPrefix(absPath, tm.sessionDir)
}

// Close 关闭临时文件管理器，清理所有资源
func (tm *TempFileManager) Close() {
	if tm.cleanupTimer != nil {
		tm.cleanupTimer.Stop()
	}
	tm.Cleanup()
}