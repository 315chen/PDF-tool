package file

import (
	"crypto/md5"
	"fmt"
	"os"
	"sync"
	"time"
)

// TempFileInfo 临时文件信息
type TempFileInfo struct {
	Path         string
	Size         int64
	CreatedAt    time.Time
	LastAccessed time.Time
	Hash         string
	Tags         []string
}

// AdvancedTempManager 高级临时文件管理器
type AdvancedTempManager struct {
	*TempFileManager
	fileInfos    map[string]*TempFileInfo
	tagIndex     map[string][]string // tag -> file paths
	sizeLimit    int64               // 总大小限制
	currentSize  int64               // 当前总大小
	infoMutex    sync.RWMutex
}

// NewAdvancedTempManager 创建高级临时文件管理器
func NewAdvancedTempManager(baseDir string, sizeLimit int64) (*AdvancedTempManager, error) {
	baseTempManager, err := NewTempFileManager(baseDir)
	if err != nil {
		return nil, err
	}

	return &AdvancedTempManager{
		TempFileManager: baseTempManager,
		fileInfos:       make(map[string]*TempFileInfo),
		tagIndex:        make(map[string][]string),
		sizeLimit:       sizeLimit,
		currentSize:     0,
	}, nil
}

// CreateTempFileWithTags 创建带标签的临时文件
func (atm *AdvancedTempManager) CreateTempFileWithTags(prefix, suffix string, tags []string) (string, *os.File, error) {
	// 检查大小限制
	if err := atm.checkSizeLimit(0); err != nil {
		return "", nil, err
	}

	filePath, file, err := atm.TempFileManager.CreateTempFile(prefix, suffix)
	if err != nil {
		return "", nil, err
	}

	// 创建文件信息
	info := &TempFileInfo{
		Path:         filePath,
		Size:         0,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Tags:         make([]string, len(tags)),
	}
	copy(info.Tags, tags)

	atm.infoMutex.Lock()
	atm.fileInfos[filePath] = info
	
	// 更新标签索引
	for _, tag := range tags {
		atm.tagIndex[tag] = append(atm.tagIndex[tag], filePath)
	}
	atm.infoMutex.Unlock()

	return filePath, file, nil
}

// CreateTempFileWithContentAndTags 创建带内容和标签的临时文件
func (atm *AdvancedTempManager) CreateTempFileWithContentAndTags(prefix, suffix string, content []byte, tags []string) (string, error) {
	// 检查大小限制
	if err := atm.checkSizeLimit(int64(len(content))); err != nil {
		return "", err
	}

	filePath, file, err := atm.CreateTempFileWithTags(prefix, suffix, tags)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入内容
	if _, err := file.Write(content); err != nil {
		atm.RemoveFileAdvanced(filePath)
		return "", fmt.Errorf("无法写入临时文件: %v", err)
	}

	// 计算文件哈希
	hash := fmt.Sprintf("%x", md5.Sum(content))

	// 更新文件信息
	atm.infoMutex.Lock()
	if info, exists := atm.fileInfos[filePath]; exists {
		info.Size = int64(len(content))
		info.Hash = hash
		atm.currentSize += info.Size
	}
	atm.infoMutex.Unlock()

	return filePath, nil
}

// GetFilesByTag 根据标签获取文件列表
func (atm *AdvancedTempManager) GetFilesByTag(tag string) []string {
	atm.infoMutex.RLock()
	defer atm.infoMutex.RUnlock()

	if files, exists := atm.tagIndex[tag]; exists {
		result := make([]string, len(files))
		copy(result, files)
		return result
	}
	return nil
}

// GetFileInfo 获取文件详细信息
func (atm *AdvancedTempManager) GetFileInfo(filePath string) (*TempFileInfo, error) {
	atm.infoMutex.RLock()
	defer atm.infoMutex.RUnlock()

	if info, exists := atm.fileInfos[filePath]; exists {
		// 返回副本
		infoCopy := *info
		infoCopy.Tags = make([]string, len(info.Tags))
		copy(infoCopy.Tags, info.Tags)
		return &infoCopy, nil
	}

	return nil, fmt.Errorf("文件信息不存在: %s", filePath)
}

// UpdateLastAccessed 更新文件最后访问时间
func (atm *AdvancedTempManager) UpdateLastAccessed(filePath string) {
	atm.infoMutex.Lock()
	defer atm.infoMutex.Unlock()

	if info, exists := atm.fileInfos[filePath]; exists {
		info.LastAccessed = time.Now()
	}
}

// RemoveFileAdvanced 删除文件（高级版本）
func (atm *AdvancedTempManager) RemoveFileAdvanced(filePath string) error {
	atm.infoMutex.Lock()
	defer atm.infoMutex.Unlock()

	// 获取文件信息
	info, exists := atm.fileInfos[filePath]
	if !exists {
		return atm.TempFileManager.RemoveFile(filePath)
	}

	// 从标签索引中移除
	for _, tag := range info.Tags {
		if files, tagExists := atm.tagIndex[tag]; tagExists {
			for i, f := range files {
				if f == filePath {
					atm.tagIndex[tag] = append(files[:i], files[i+1:]...)
					break
				}
			}
			// 如果标签下没有文件了，删除标签
			if len(atm.tagIndex[tag]) == 0 {
				delete(atm.tagIndex, tag)
			}
		}
	}

	// 更新总大小
	atm.currentSize -= info.Size

	// 删除文件信息
	delete(atm.fileInfos, filePath)

	// 删除实际文件
	return atm.TempFileManager.RemoveFile(filePath)
}

// CleanupByTag 根据标签清理文件
func (atm *AdvancedTempManager) CleanupByTag(tag string) error {
	files := atm.GetFilesByTag(tag)
	var lastError error

	for _, filePath := range files {
		if err := atm.RemoveFileAdvanced(filePath); err != nil {
			lastError = err
		}
	}

	return lastError
}

// CleanupOldFiles 清理旧文件（基于最后访问时间）
func (atm *AdvancedTempManager) CleanupOldFiles(maxAge time.Duration) int {
	atm.infoMutex.Lock()
	defer atm.infoMutex.Unlock()

	now := time.Now()
	var filesToRemove []string

	for filePath, info := range atm.fileInfos {
		if now.Sub(info.LastAccessed) > maxAge {
			filesToRemove = append(filesToRemove, filePath)
		}
	}

	cleanedCount := 0
	for _, filePath := range filesToRemove {
		if err := atm.removeFileInternal(filePath); err == nil {
			cleanedCount++
		}
	}

	return cleanedCount
}

// CleanupLargeFiles 清理大文件
func (atm *AdvancedTempManager) CleanupLargeFiles(maxSize int64) int {
	atm.infoMutex.Lock()
	defer atm.infoMutex.Unlock()

	var filesToRemove []string

	for filePath, info := range atm.fileInfos {
		if info.Size > maxSize {
			filesToRemove = append(filesToRemove, filePath)
		}
	}

	cleanedCount := 0
	for _, filePath := range filesToRemove {
		if err := atm.removeFileInternal(filePath); err == nil {
			cleanedCount++
		}
	}

	return cleanedCount
}

// GetStatistics 获取统计信息
func (atm *AdvancedTempManager) GetStatistics() map[string]interface{} {
	atm.infoMutex.RLock()
	defer atm.infoMutex.RUnlock()

	stats := map[string]interface{}{
		"total_files":    len(atm.fileInfos),
		"total_size":     atm.currentSize,
		"size_limit":     atm.sizeLimit,
		"usage_percent":  float64(atm.currentSize) / float64(atm.sizeLimit) * 100,
		"total_tags":     len(atm.tagIndex),
		"session_dir":    atm.GetSessionDir(),
	}

	// 按标签统计
	tagStats := make(map[string]int)
	for tag, files := range atm.tagIndex {
		tagStats[tag] = len(files)
	}
	stats["tag_stats"] = tagStats

	return stats
}

// checkSizeLimit 检查大小限制
func (atm *AdvancedTempManager) checkSizeLimit(additionalSize int64) error {
	atm.infoMutex.RLock()
	defer atm.infoMutex.RUnlock()

	if atm.sizeLimit > 0 && atm.currentSize+additionalSize > atm.sizeLimit {
		return fmt.Errorf("超出大小限制: 当前 %d + 新增 %d > 限制 %d", 
			atm.currentSize, additionalSize, atm.sizeLimit)
	}
	return nil
}

// removeFileInternal 内部文件删除方法（不加锁）
func (atm *AdvancedTempManager) removeFileInternal(filePath string) error {
	info, exists := atm.fileInfos[filePath]
	if !exists {
		return fmt.Errorf("文件信息不存在: %s", filePath)
	}

	// 从标签索引中移除
	for _, tag := range info.Tags {
		if files, tagExists := atm.tagIndex[tag]; tagExists {
			for i, f := range files {
				if f == filePath {
					atm.tagIndex[tag] = append(files[:i], files[i+1:]...)
					break
				}
			}
			if len(atm.tagIndex[tag]) == 0 {
				delete(atm.tagIndex, tag)
			}
		}
	}

	// 更新总大小
	atm.currentSize -= info.Size

	// 删除文件信息
	delete(atm.fileInfos, filePath)

	// 删除实际文件
	return os.Remove(filePath)
}

// Close 关闭高级临时文件管理器
func (atm *AdvancedTempManager) Close() {
	atm.infoMutex.Lock()
	defer atm.infoMutex.Unlock()

	// 清空所有信息
	atm.fileInfos = make(map[string]*TempFileInfo)
	atm.tagIndex = make(map[string][]string)
	atm.currentSize = 0

	// 关闭基础管理器
	atm.TempFileManager.Close()
}
