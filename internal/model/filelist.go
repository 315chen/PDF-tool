package model

import (
	"sort"
	"sync"
)

// FileList 定义文件列表管理器
type FileList struct {
	mu      sync.RWMutex
	files   []*FileEntry
	mainFile *FileEntry
}

// NewFileList 创建一个新的文件列表
func NewFileList() *FileList {
	return &FileList{
		files: make([]*FileEntry, 0),
	}
}

// SetMainFile 设置主文件
func (fl *FileList) SetMainFile(path string) *FileEntry {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	fl.mainFile = NewFileEntry(path, 0)
	return fl.mainFile
}

// GetMainFile 获取主文件
func (fl *FileList) GetMainFile() *FileEntry {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	return fl.mainFile
}

// AddFile 添加文件到列表
func (fl *FileList) AddFile(path string) *FileEntry {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	// 检查文件是否已存在
	for _, file := range fl.files {
		if file.Path == path {
			return file
		}
	}
	
	// 创建新的文件条目
	order := len(fl.files) + 1
	fileEntry := NewFileEntry(path, order)
	fl.files = append(fl.files, fileEntry)
	
	return fileEntry
}

// RemoveFile 从列表中移除文件
func (fl *FileList) RemoveFile(path string) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	for i, file := range fl.files {
		if file.Path == path {
			// 移除文件
			fl.files = append(fl.files[:i], fl.files[i+1:]...)
			
			// 重新排序
			fl.reorderFiles()
			return true
		}
	}
	
	return false
}

// MoveFile 移动文件位置
func (fl *FileList) MoveFile(path string, newOrder int) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	// 找到要移动的文件
	var targetFile *FileEntry
	targetIndex := -1
	
	for i, file := range fl.files {
		if file.Path == path {
			targetFile = file
			targetIndex = i
			break
		}
	}
	
	if targetFile == nil || newOrder < 1 || newOrder > len(fl.files) {
		return false
	}
	
	// 移除文件
	fl.files = append(fl.files[:targetIndex], fl.files[targetIndex+1:]...)
	
	// 插入到新位置
	newIndex := newOrder - 1
	if newIndex >= len(fl.files) {
		fl.files = append(fl.files, targetFile)
	} else {
		fl.files = append(fl.files[:newIndex], append([]*FileEntry{targetFile}, fl.files[newIndex:]...)...)
	}
	
	// 重新排序
	fl.reorderFiles()
	return true
}

// GetFiles 获取所有附加文件
func (fl *FileList) GetFiles() []*FileEntry {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	// 返回副本以避免并发修改
	files := make([]*FileEntry, len(fl.files))
	copy(files, fl.files)
	return files
}

// GetAllFiles 获取所有文件（包括主文件）
func (fl *FileList) GetAllFiles() []*FileEntry {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	allFiles := make([]*FileEntry, 0, len(fl.files)+1)
	
	if fl.mainFile != nil {
		allFiles = append(allFiles, fl.mainFile)
	}
	
	allFiles = append(allFiles, fl.files...)
	return allFiles
}

// GetFilePaths 获取所有附加文件路径
func (fl *FileList) GetFilePaths() []string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	paths := make([]string, len(fl.files))
	for i, file := range fl.files {
		paths[i] = file.Path
	}
	return paths
}

// GetAllFilePaths 获取所有文件路径（包括主文件）
func (fl *FileList) GetAllFilePaths() []string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	paths := make([]string, 0, len(fl.files)+1)
	
	if fl.mainFile != nil {
		paths = append(paths, fl.mainFile.Path)
	}
	
	for _, file := range fl.files {
		paths = append(paths, file.Path)
	}
	
	return paths
}

// Clear 清空文件列表
func (fl *FileList) Clear() {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	fl.files = fl.files[:0]
	fl.mainFile = nil
}

// Count 获取附加文件数量
func (fl *FileList) Count() int {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	return len(fl.files)
}

// TotalCount 获取总文件数量（包括主文件）
func (fl *FileList) TotalCount() int {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	count := len(fl.files)
	if fl.mainFile != nil {
		count++
	}
	return count
}

// IsEmpty 检查是否为空
func (fl *FileList) IsEmpty() bool {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	return len(fl.files) == 0 && fl.mainFile == nil
}

// HasMainFile 检查是否有主文件
func (fl *FileList) HasMainFile() bool {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	return fl.mainFile != nil
}

// GetValidFiles 获取所有有效的文件
func (fl *FileList) GetValidFiles() []*FileEntry {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	validFiles := make([]*FileEntry, 0)
	
	if fl.mainFile != nil && fl.mainFile.IsValid {
		validFiles = append(validFiles, fl.mainFile)
	}
	
	for _, file := range fl.files {
		if file.IsValid {
			validFiles = append(validFiles, file)
		}
	}
	
	return validFiles
}

// reorderFiles 重新排序文件（内部方法，调用时需要已加锁）
func (fl *FileList) reorderFiles() {
	for i, file := range fl.files {
		file.Order = i + 1
	}
	
	// 按顺序排序
	sort.Slice(fl.files, func(i, j int) bool {
		return fl.files[i].Order < fl.files[j].Order
	})
}