package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/user/pdf-merger/internal/model"
)

// FileListManager 文件列表管理器
type FileListManager struct {
	files         []model.FileEntry
	list          *widget.List
	selectedIndex int
	onFileChanged func()
	onFileInfo    func(string) (*model.FileEntry, error)
}

// NewFileListManager 创建新的文件列表管理器
func NewFileListManager() *FileListManager {
	flm := &FileListManager{
		files:         make([]model.FileEntry, 0),
		selectedIndex: -1,
	}
	
	flm.createList()
	return flm
}

// createList 创建文件列表组件
func (flm *FileListManager) createList() {
	flm.list = widget.NewList(
		func() int {
			return len(flm.files)
		},
		func() fyne.CanvasObject {
			return flm.createListItem()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			flm.updateListItem(id, obj)
		},
	)
	
	// 设置列表选择处理
	flm.list.OnSelected = func(id widget.ListItemID) {
		flm.selectedIndex = id
	}
	
	flm.list.OnUnselected = func(id widget.ListItemID) {
		if flm.selectedIndex == id {
			flm.selectedIndex = -1
		}
	}
}

// createListItem 创建列表项模板
func (flm *FileListManager) createListItem() fyne.CanvasObject {
	// 简化的列表项，避免复杂的嵌套容器
	fileIcon := widget.NewIcon(theme.DocumentIcon())
	nameLabel := widget.NewLabel("文件名")
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	sizeLabel := widget.NewLabel("大小")
	statusLabel := widget.NewLabel("状态")
	
	return container.NewHBox(
		fileIcon,
		nameLabel,
		sizeLabel,
		statusLabel,
	)
}

// updateListItem 更新列表项内容
func (flm *FileListManager) updateListItem(id widget.ListItemID, obj fyne.CanvasObject) {
	if id >= len(flm.files) {
		return
	}
	
	file := flm.files[id]
	
	// 简化的列表项更新，避免复杂的容器结构
	// 由于Fyne的List组件限制，我们使用简单的布局
	container := obj.(*fyne.Container)
	if len(container.Objects) < 4 {
		return
	}
	
	// 更新文件图标
	if icon, ok := container.Objects[0].(*widget.Icon); ok {
		if file.IsValid {
			icon.SetResource(theme.DocumentIcon())
		} else {
			icon.SetResource(theme.ErrorIcon())
		}
	}
	
	// 更新文件名
	if nameLabel, ok := container.Objects[1].(*widget.Label); ok {
		nameLabel.SetText(file.DisplayName)
	}
	
	// 更新文件大小
	if sizeLabel, ok := container.Objects[2].(*widget.Label); ok {
		sizeLabel.SetText(file.GetSizeString())
	}
	
	// 更新状态
	if statusLabel, ok := container.Objects[3].(*widget.Label); ok {
		statusLabel.SetText(flm.getStatusText(file))
	}
}

// getStatusText 获取状态文本
func (flm *FileListManager) getStatusText(file model.FileEntry) string {
	if !file.IsValid {
		if file.Error != "" {
			return "错误"
		}
		return "无效"
	}
	
	if file.IsEncrypted {
		return "已加密"
	}
	
	return "正常"
}

// AddFile 添加文件到列表
func (flm *FileListManager) AddFile(filePath string) error {
	// 检查文件是否已存在
	for _, file := range flm.files {
		if file.Path == filePath {
			return fmt.Errorf("文件已存在于列表中")
		}
	}
	
	// 创建文件条目
	fileEntry := model.NewFileEntry(filePath, len(flm.files))
	
	// 获取文件信息
	if flm.onFileInfo != nil {
		if info, err := flm.onFileInfo(filePath); err == nil {
			fileEntry.Size = info.Size
			fileEntry.PageCount = info.PageCount
			fileEntry.IsEncrypted = info.IsEncrypted
			fileEntry.IsValid = info.IsValid
			fileEntry.Error = info.Error
		}
	}
	
	// 添加到列表
	flm.files = append(flm.files, *fileEntry)
	flm.list.Refresh()
	
	if flm.onFileChanged != nil {
		flm.onFileChanged()
	}
	
	return nil
}

// RemoveFile 移除指定索引的文件
func (flm *FileListManager) removeFile(index int) {
	if index < 0 || index >= len(flm.files) {
		return
	}
	
	// 移除文件
	flm.files = append(flm.files[:index], flm.files[index+1:]...)
	
	// 重新设置Order
	for i := range flm.files {
		flm.files[i].Order = i
	}
	
	// 调整选中索引
	if flm.selectedIndex == index {
		flm.selectedIndex = -1
	} else if flm.selectedIndex > index {
		flm.selectedIndex--
	}
	
	flm.list.Refresh()
	
	if flm.onFileChanged != nil {
		flm.onFileChanged()
	}
}

// RemoveSelected 移除选中的文件
func (flm *FileListManager) RemoveSelected() {
	if flm.selectedIndex >= 0 {
		flm.removeFile(flm.selectedIndex)
	}
}

// moveFileUp 向上移动文件
func (flm *FileListManager) moveFileUp(index int) {
	if index <= 0 || index >= len(flm.files) {
		return
	}
	
	// 交换文件位置
	flm.files[index], flm.files[index-1] = flm.files[index-1], flm.files[index]
	
	// 更新Order
	flm.files[index-1].Order = index - 1
	flm.files[index].Order = index
	
	// 更新选中索引
	if flm.selectedIndex == index {
		flm.selectedIndex = index - 1
		flm.list.Select(flm.selectedIndex)
	}
	
	flm.list.Refresh()
	
	if flm.onFileChanged != nil {
		flm.onFileChanged()
	}
}

// moveFileDown 向下移动文件
func (flm *FileListManager) moveFileDown(index int) {
	if index < 0 || index >= len(flm.files)-1 {
		return
	}
	
	// 交换文件位置
	flm.files[index], flm.files[index+1] = flm.files[index+1], flm.files[index]
	
	// 更新Order
	flm.files[index].Order = index
	flm.files[index+1].Order = index + 1
	
	// 更新选中索引
	if flm.selectedIndex == index {
		flm.selectedIndex = index + 1
		flm.list.Select(flm.selectedIndex)
	}
	
	flm.list.Refresh()
	
	if flm.onFileChanged != nil {
		flm.onFileChanged()
	}
}

// MoveSelectedUp 向上移动选中的文件
func (flm *FileListManager) MoveSelectedUp() {
	if flm.selectedIndex > 0 {
		flm.moveFileUp(flm.selectedIndex)
	}
}

// MoveSelectedDown 向下移动选中的文件
func (flm *FileListManager) MoveSelectedDown() {
	if flm.selectedIndex >= 0 && flm.selectedIndex < len(flm.files)-1 {
		flm.moveFileDown(flm.selectedIndex)
	}
}

// Clear 清空文件列表
func (flm *FileListManager) Clear() {
	flm.files = make([]model.FileEntry, 0)
	flm.selectedIndex = -1
	flm.list.Refresh()
	
	if flm.onFileChanged != nil {
		flm.onFileChanged()
	}
}

// GetFiles 获取文件列表
func (flm *FileListManager) GetFiles() []model.FileEntry {
	return flm.files
}

// GetFilePaths 获取文件路径列表
func (flm *FileListManager) GetFilePaths() []string {
	paths := make([]string, len(flm.files))
	for i, file := range flm.files {
		paths[i] = file.Path
	}
	return paths
}

// GetSelectedIndex 获取选中的索引
func (flm *FileListManager) GetSelectedIndex() int {
	return flm.selectedIndex
}

// HasFiles 检查是否有文件
func (flm *FileListManager) HasFiles() bool {
	return len(flm.files) > 0
}

// GetFileCount 获取文件数量
func (flm *FileListManager) GetFileCount() int {
	return len(flm.files)
}

// GetWidget 获取列表组件
func (flm *FileListManager) GetWidget() *widget.List {
	return flm.list
}

// SetOnFileChanged 设置文件变更回调
func (flm *FileListManager) SetOnFileChanged(callback func()) {
	flm.onFileChanged = callback
}

// SetOnFileInfo 设置文件信息获取回调
func (flm *FileListManager) SetOnFileInfo(callback func(string) (*model.FileEntry, error)) {
	flm.onFileInfo = callback
}

// RefreshFileInfo 刷新文件信息
func (flm *FileListManager) RefreshFileInfo() {
	if flm.onFileInfo == nil {
		return
	}
	
	for i := range flm.files {
		if info, err := flm.onFileInfo(flm.files[i].Path); err == nil {
			flm.files[i].Size = info.Size
			flm.files[i].PageCount = info.PageCount
			flm.files[i].IsEncrypted = info.IsEncrypted
			flm.files[i].IsValid = info.IsValid
			flm.files[i].Error = info.Error
		}
	}
	
	flm.list.Refresh()
}

// GetFileInfo 获取指定文件的信息摘要
func (flm *FileListManager) GetFileInfo() string {
	if len(flm.files) == 0 {
		return "没有文件"
	}
	
	totalFiles := len(flm.files)
	validFiles := 0
	encryptedFiles := 0
	totalPages := 0
	totalSize := int64(0)
	
	for _, file := range flm.files {
		if file.IsValid {
			validFiles++
			totalPages += file.PageCount
		}
		if file.IsEncrypted {
			encryptedFiles++
		}
		totalSize += file.Size
	}
	
	var info strings.Builder
	info.WriteString(fmt.Sprintf("文件: %d个", totalFiles))
	
	if validFiles != totalFiles {
		info.WriteString(fmt.Sprintf(" (有效: %d个)", validFiles))
	}
	
	if encryptedFiles > 0 {
		info.WriteString(fmt.Sprintf(" (加密: %d个)", encryptedFiles))
	}
	
	if totalPages > 0 {
		info.WriteString(fmt.Sprintf(", 总页数: %d页", totalPages))
	}
	
	info.WriteString(fmt.Sprintf(", 总大小: %s", formatFileSize(totalSize)))
	
	return info.String()
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}
}