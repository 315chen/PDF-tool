package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	
	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
)

// UI 定义用户界面组件
type UI struct {
	window              fyne.Window
	controller          *controller.Controller
	eventHandler        interface{} // 将在后续更新中使用具体类型
	mainFileEntry       *widget.Entry
	mainFileBrowseBtn   *widget.Button
	fileListManager     *FileListManager
	fileInfoLabel       *widget.Label
	addFileBtn          *widget.Button
	removeFileBtn       *widget.Button
	clearFilesBtn       *widget.Button
	moveUpBtn           *widget.Button
	moveDownBtn         *widget.Button
	refreshBtn          *widget.Button
	outputPathEntry     *widget.Entry
	outputBrowseBtn     *widget.Button
	progressManager     *ProgressManager
	mergeButton         *widget.Button
	cancelButton        *widget.Button
	
	// 数据
	mainFilePath string
	outputPath   string
}

// NewUI 创建一个新的UI实例
func NewUI(window fyne.Window, controller *controller.Controller) *UI {
	ui := &UI{
		window:     window,
		controller: controller,
	}
	
	// 创建文件列表管理器
	ui.fileListManager = NewFileListManager()
	
	// 创建进度管理器
	ui.progressManager = NewProgressManager(window)
	
	// 设置回调
	ui.fileListManager.SetOnFileChanged(ui.onFileListChanged)
	ui.fileListManager.SetOnFileInfo(ui.getFileInfo)
	ui.progressManager.SetOnCancel(ui.onProgressCancel)
	ui.progressManager.SetOnComplete(ui.onProgressComplete)
	
	return ui
}

// BuildUI 构建用户界面
func (u *UI) BuildUI() fyne.CanvasObject {
	// 创建主文件选择区域
	mainFileSection := u.createMainFileSection()
	
	// 创建附加文件列表区域
	additionalFilesSection := u.createAdditionalFilesSection()
	
	// 创建输出文件选择区域
	outputSection := u.createOutputSection()
	
	// 创建进度和控制区域
	controlSection := u.createControlSection()
	
	// 构建主布局
	content := container.NewVBox(
		mainFileSection,
		widget.NewSeparator(),
		additionalFilesSection,
		widget.NewSeparator(),
		outputSection,
		widget.NewSeparator(),
		controlSection,
	)
	
	// 设置初始状态
	u.updateUI()
	
	return content
}

// createMainFileSection 创建主文件选择区域
func (u *UI) createMainFileSection() *fyne.Container {
	// 主文件输入框
	u.mainFileEntry = widget.NewEntry()
	u.mainFileEntry.SetPlaceHolder("请选择主PDF文件...")
	u.mainFileEntry.Disable() // 只读，通过浏览按钮选择
	
	// 主文件浏览按钮
	u.mainFileBrowseBtn = widget.NewButton(BrowseButton, u.onMainFileBrowse)
	
	// 布局
	fileRow := container.NewBorder(nil, nil, nil, u.mainFileBrowseBtn, u.mainFileEntry)
	
	return container.NewVBox(
		widget.NewRichTextFromMarkdown("## 主PDF文件"),
		fileRow,
	)
}

// createAdditionalFilesSection 创建附加文件列表区域
func (u *UI) createAdditionalFilesSection() *fyne.Container {
	// 文件信息标签
	u.fileInfoLabel = widget.NewLabel(NoFilesLabel)
	u.fileInfoLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 主要操作按钮
	u.addFileBtn = widget.NewButtonWithIcon(AddFileButton, theme.ContentAddIcon(), u.onAddFiles)
	u.removeFileBtn = widget.NewButtonWithIcon(RemoveFileButton, theme.DeleteIcon(), u.onRemoveSelected)
	u.clearFilesBtn = widget.NewButtonWithIcon(ClearFilesButton, theme.ContentClearIcon(), u.onClearFiles)

	// 排序按钮
	u.moveUpBtn = widget.NewButtonWithIcon(MoveUpButton, theme.MoveUpIcon(), u.onMoveUp)
	u.moveDownBtn = widget.NewButtonWithIcon(MoveDownButton, theme.MoveDownIcon(), u.onMoveDown)
	u.refreshBtn = widget.NewButtonWithIcon(RefreshButton, theme.ViewRefreshIcon(), u.onRefreshFiles)
	
	// 按钮行
	mainButtonRow := container.NewHBox(
		u.addFileBtn,
		u.removeFileBtn,
		u.clearFilesBtn,
	)
	
	sortButtonRow := container.NewHBox(
		u.moveUpBtn,
		u.moveDownBtn,
		u.refreshBtn,
	)
	
	buttonContainer := container.NewVBox(
		mainButtonRow,
		sortButtonRow,
	)
	
	// 文件列表容器
	listWidget := u.fileListManager.GetWidget()
	listContainer := container.NewBorder(
		u.fileInfoLabel,
		buttonContainer,
		nil, nil,
		listWidget,
	)
	
	return container.NewVBox(
		widget.NewRichTextFromMarkdown("## 附加PDF文件"),
		listContainer,
	)
}

// createOutputSection 创建输出文件选择区域
func (u *UI) createOutputSection() *fyne.Container {
	// 输出路径输入框
	u.outputPathEntry = widget.NewEntry()
	u.outputPathEntry.SetPlaceHolder("请选择输出文件路径...")
	u.outputPathEntry.OnChanged = func(text string) {
		u.outputPath = text
		u.updateUI()
	}
	
	// 输出路径浏览按钮
	u.outputBrowseBtn = widget.NewButton(BrowseButton, u.onOutputBrowse)
	
	// 布局
	outputRow := container.NewBorder(nil, nil, nil, u.outputBrowseBtn, u.outputPathEntry)
	
	return container.NewVBox(
		widget.NewRichTextFromMarkdown("## 输出文件"),
		outputRow,
	)
}

// createControlSection 创建进度和控制区域
func (u *UI) createControlSection() *fyne.Container {
	// 控制按钮
	u.mergeButton = widget.NewButtonWithIcon(StartMergeButton, theme.MediaPlayIcon(), u.onMerge)
	u.cancelButton = widget.NewButtonWithIcon(CancelButton, theme.CancelIcon(), u.onCancel)
	u.cancelButton.Hide() // 初始隐藏
	
	buttonRow := container.NewHBox(
		u.mergeButton,
		u.cancelButton,
	)
	
	// 获取进度管理器容器
	progressContainer := u.progressManager.GetContainer()
	
	return container.NewVBox(
		progressContainer,
		buttonRow,
	)
}

// 事件处理方法

// onMainFileBrowse 主文件浏览按钮点击处理
func (u *UI) onMainFileBrowse() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		
		uri := reader.URI()
		if uri == nil {
			return
		}
		
		path := uri.Path()
		if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			dialog.ShowError(fmt.Errorf("请选择PDF文件"), u.window)
			return
		}
		
		u.mainFilePath = path
		u.mainFileEntry.SetText(filepath.Base(path))
		u.updateUI()
		
	}, u.window)
	
	// 设置文件过滤器
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	fileDialog.Show()
}

// onAddFiles 添加文件按钮点击处理
func (u *UI) onAddFiles() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		
		uri := reader.URI()
		if uri == nil {
			return
		}
		
		path := uri.Path()
		if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			dialog.ShowError(fmt.Errorf("请选择PDF文件"), u.window)
			return
		}
		
		// 添加到文件列表管理器
		if err := u.fileListManager.AddFile(path); err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		
	}, u.window)
	
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	fileDialog.Show()
}

// onRemoveSelected 移除选中文件按钮点击处理
func (u *UI) onRemoveSelected() {
	if !u.fileListManager.HasFiles() {
		dialog.ShowInformation("提示", "没有文件可以移除", u.window)
		return
	}
	
	if u.fileListManager.GetSelectedIndex() < 0 {
		dialog.ShowInformation("提示", "请先选择要移除的文件", u.window)
		return
	}
	
	u.fileListManager.RemoveSelected()
}

// onClearFiles 清空文件列表按钮点击处理
func (u *UI) onClearFiles() {
	if !u.fileListManager.HasFiles() {
		return
	}
	
	dialog.ShowConfirm("确认", "确定要清空所有附加文件吗？", func(confirmed bool) {
		if confirmed {
			u.fileListManager.Clear()
		}
	}, u.window)
}

// onOutputBrowse 输出路径浏览按钮点击处理
func (u *UI) onOutputBrowse() {
	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		
		uri := writer.URI()
		if uri == nil {
			return
		}
		
		path := uri.Path()
		if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			path += ".pdf"
		}
		
		u.outputPath = path
		u.outputPathEntry.SetText(path)
		u.updateUI()
		
	}, u.window)
	
	fileDialog.SetFileName("merged.pdf")
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	fileDialog.Show()
}

// onMerge 合并按钮点击处理
func (u *UI) onMerge() {
	// 如果有事件处理器，使用事件处理器
	if u.eventHandler != nil {
		// 这里需要类型断言，但为了保持兼容性，先使用原有逻辑
		additionalFiles := u.fileListManager.GetFilePaths()
		
		// 基本验证
		if u.mainFilePath == "" {
			dialog.ShowError(fmt.Errorf("请选择主PDF文件"), u.window)
			return
		}
		
		if len(additionalFiles) == 0 {
			dialog.ShowError(fmt.Errorf("请至少添加一个附加PDF文件"), u.window)
			return
		}
		
		if u.outputPath == "" {
			dialog.ShowError(fmt.Errorf("请选择输出文件路径"), u.window)
			return
		}
		
		// 开始异步合并
		u.startAsyncMerge()
		return
	}
	
	// 原有的同步合并逻辑（向后兼容）
	// 验证输入
	if u.mainFilePath == "" {
		dialog.ShowError(fmt.Errorf("请选择主PDF文件"), u.window)
		return
	}
	
	if !u.fileListManager.HasFiles() {
		dialog.ShowError(fmt.Errorf("请至少添加一个附加PDF文件"), u.window)
		return
	}
	
	if u.outputPath == "" {
		dialog.ShowError(fmt.Errorf("请选择输出文件路径"), u.window)
		return
	}
	
	// 开始合并
	u.startMerge()
}

// onCancel 取消按钮点击处理
func (u *UI) onCancel() {
	// 取消合并操作
	u.cancelMerge()
}

// onMoveUp 上移文件按钮点击处理
func (u *UI) onMoveUp() {
	u.fileListManager.MoveSelectedUp()
}

// onMoveDown 下移文件按钮点击处理
func (u *UI) onMoveDown() {
	u.fileListManager.MoveSelectedDown()
}

// onRefreshFiles 刷新文件信息按钮点击处理
func (u *UI) onRefreshFiles() {
	u.fileListManager.RefreshFileInfo()
	u.updateFileInfo()
}

// onFileListChanged 文件列表变更回调
func (u *UI) onFileListChanged() {
	u.updateFileInfo()
	u.updateUI()
}

// getFileInfo 获取文件信息回调
func (u *UI) getFileInfo(filePath string) (*model.FileEntry, error) {
	// 创建文件条目
	fileEntry := &model.FileEntry{
		Path:        filePath,
		DisplayName: filepath.Base(filePath),
		IsValid:     true,
	}
	
	// 获取文件大小
	if fileInfo, err := os.Stat(filePath); err == nil {
		fileEntry.Size = fileInfo.Size()
	}
	
	// 获取PDF信息
	if u.controller != nil {
		if pdfInfo, err := u.controller.GetPDFInfo(filePath); err == nil {
			fileEntry.PageCount = pdfInfo.PageCount
			fileEntry.IsEncrypted = pdfInfo.IsEncrypted
		} else {
			fileEntry.IsValid = false
			fileEntry.Error = err.Error()
		}
	}
	
	return fileEntry, nil
}

// updateFileInfo 更新文件信息显示
func (u *UI) updateFileInfo() {
	if u.fileInfoLabel != nil {
		u.fileInfoLabel.SetText(u.fileListManager.GetFileInfo())
	}
}

// disableInputControls 禁用输入控件
func (u *UI) disableInputControls() {
	u.mainFileBrowseBtn.Disable()
	u.addFileBtn.Disable()
	u.removeFileBtn.Disable()
	u.clearFilesBtn.Disable()
	u.moveUpBtn.Disable()
	u.moveDownBtn.Disable()
	u.refreshBtn.Disable()
	u.outputBrowseBtn.Disable()
}

// enableInputControls 启用输入控件
func (u *UI) enableInputControls() {
	u.mainFileBrowseBtn.Enable()
	u.addFileBtn.Enable()
	u.removeFileBtn.Enable()
	u.clearFilesBtn.Enable()
	u.moveUpBtn.Enable()
	u.moveDownBtn.Enable()
	u.refreshBtn.Enable()
	u.outputBrowseBtn.Enable()
	
	// 重新应用按钮状态逻辑
	u.updateUI()
}

// performMerge 执行合并操作
func (u *UI) performMerge() {
	defer func() {
		// 确保在操作完成后恢复UI状态
		u.mergeButton.Show()
		u.cancelButton.Hide()
		u.enableInputControls()
	}()
	
	// 步骤1: 验证文件
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: 0.1,
		Status:   "验证文件",
		Detail:   "正在验证PDF文件...",
		Step:     1,
	})
	
	if !u.validateFiles() {
		return
	}
	
	// 步骤2: 准备合并
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: 0.3,
		Status:   "准备合并",
		Detail:   "正在准备合并操作...",
		Step:     2,
	})
	
	// 步骤3: 执行合并
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: 0.5,
		Status:   "合并文件",
		Detail:   "正在合并PDF文件...",
		Step:     3,
	})
	
	if !u.executeMerge() {
		return
	}
	
	// 步骤4: 保存文件
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: 0.8,
		Status:   "保存文件",
		Detail:   "正在保存合并后的文件...",
		Step:     4,
	})
	
	// 步骤5: 完成
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: 1.0,
		Status:   "完成",
		Detail:   "合并操作已完成",
		Step:     5,
	})
	
	u.progressManager.Complete("PDF合并完成！")
}

// validateFiles 验证文件
func (u *UI) validateFiles() bool {
	// 验证主文件
	if err := u.controller.ValidateFile(u.mainFilePath); err != nil {
		u.progressManager.Error(fmt.Errorf("主文件验证失败: %v", err))
		return false
	}
	
	// 验证附加文件
	additionalFiles := u.fileListManager.GetFilePaths()
	for i, filePath := range additionalFiles {
		u.progressManager.UpdateProgress(ProgressInfo{
			Progress:       0.1 + (0.2 * float64(i) / float64(len(additionalFiles))),
			CurrentFile:    filepath.Base(filePath),
			ProcessedFiles: i,
			TotalFiles:     len(additionalFiles),
		})
		
		if err := u.controller.ValidateFile(filePath); err != nil {
			u.progressManager.Error(fmt.Errorf("文件 %s 验证失败: %v", filepath.Base(filePath), err))
			return false
		}
	}
	
	return true
}

// executeMerge 执行合并
func (u *UI) executeMerge() bool {
	additionalFiles := u.fileListManager.GetFilePaths()
	
	// 模拟合并过程
	for i := 0; i < len(additionalFiles)+1; i++ {
		if !u.progressManager.IsActive() {
			return false // 用户取消了操作
		}
		
		progress := 0.5 + (0.3 * float64(i) / float64(len(additionalFiles)+1))
		
		var currentFile string
		if i == 0 {
			currentFile = filepath.Base(u.mainFilePath)
		} else {
			currentFile = filepath.Base(additionalFiles[i-1])
		}
		
		u.progressManager.UpdateProgress(ProgressInfo{
			Progress:       progress,
			CurrentFile:    currentFile,
			ProcessedFiles: i,
			TotalFiles:     len(additionalFiles) + 1,
		})
		
		// 模拟处理时间
		time.Sleep(500 * time.Millisecond)
	}
	
	// 实际的合并逻辑
	if u.controller != nil {
		err := u.controller.MergePDFs(u.mainFilePath, additionalFiles, u.outputPath)
		if err != nil {
			u.progressManager.Error(fmt.Errorf("合并失败: %v", err))
			return false
		}
	}
	
	return true
}

// onProgressCancel 进度取消回调
func (u *UI) onProgressCancel() {
	// 这里可以添加取消合并的逻辑
}

// onProgressComplete 进度完成回调
func (u *UI) onProgressComplete() {
	// 显示完成对话框
	u.progressManager.ShowInfoDialog("完成", "PDF文件合并完成！")
}

// startMerge 开始合并操作
func (u *UI) startMerge() {
	// 更新UI状态
	u.mergeButton.Hide()
	u.cancelButton.Show()
	
	// 禁用输入控件
	u.disableInputControls()
	
	// 获取文件信息
	additionalFiles := u.fileListManager.GetFilePaths()
	totalFiles := len(additionalFiles) + 1 // 包括主文件
	
	// 启动进度显示
	u.progressManager.Start(5, totalFiles) // 5个主要步骤
	
	// 异步执行合并操作
	go u.performMerge()
}

// startAsyncMerge 开始异步合并操作
func (u *UI) startAsyncMerge() {
	// 更新UI状态
	u.mergeButton.Hide()
	u.cancelButton.Show()
	
	// 禁用输入控件
	u.disableInputControls()
	
	// 获取文件信息
	additionalFiles := u.fileListManager.GetFilePaths()
	totalFiles := len(additionalFiles) + 1 // 包括主文件
	
	// 启动进度显示
	u.progressManager.Start(5, totalFiles) // 5个主要步骤
	
	// 通过控制器开始异步合并
	if u.controller != nil {
		err := u.controller.StartMergeJob(u.mainFilePath, additionalFiles, u.outputPath)
		if err != nil {
			dialog.ShowError(err, u.window)
			u.cancelAsyncMerge()
		}
	}
}

// cancelAsyncMerge 取消异步合并操作
func (u *UI) cancelAsyncMerge() {
	// 通过控制器取消任务
	if u.controller != nil {
		u.controller.CancelCurrentJob()
	}
	
	// 取消进度管理器
	u.progressManager.Cancel()
	
	// 恢复UI状态
	u.mergeButton.Show()
	u.cancelButton.Hide()
	
	// 启用输入控件
	u.enableInputControls()
}

// cancelMerge 取消合并操作
func (u *UI) cancelMerge() {
	// 如果有事件处理器，使用异步取消
	if u.eventHandler != nil {
		u.cancelAsyncMerge()
		return
	}
	
	// 原有的同步取消逻辑
	// 取消进度管理器
	u.progressManager.Cancel()
	
	// 恢复UI状态
	u.mergeButton.Show()
	u.cancelButton.Hide()
	
	// 启用输入控件
	u.enableInputControls()
}

// updateUI 更新UI状态
func (u *UI) updateUI() {
	// 更新按钮状态
	canMerge := u.mainFilePath != "" && u.fileListManager.HasFiles() && u.outputPath != ""
	
	if u.mergeButton.Visible() {
		if canMerge {
			u.mergeButton.Enable()
		} else {
			u.mergeButton.Disable()
		}
	}
	
	// 更新文件操作按钮状态
	hasFiles := u.fileListManager.HasFiles()
	hasSelection := u.fileListManager.GetSelectedIndex() >= 0
	
	if hasFiles {
		u.clearFilesBtn.Enable()
	} else {
		u.clearFilesBtn.Disable()
	}
	
	if hasSelection {
		u.removeFileBtn.Enable()
		u.moveUpBtn.Enable()
		u.moveDownBtn.Enable()
	} else {
		u.removeFileBtn.Disable()
		u.moveUpBtn.Disable()
		u.moveDownBtn.Disable()
	}
	
	if hasFiles {
		u.refreshBtn.Enable()
	} else {
		u.refreshBtn.Disable()
	}
}

// GetMainFilePath 获取主文件路径
func (u *UI) GetMainFilePath() string {
	return u.mainFilePath
}

// GetAdditionalFiles 获取附加文件列表
func (u *UI) GetAdditionalFiles() []model.FileEntry {
	return u.fileListManager.GetFiles()
}

// GetAdditionalFilePaths 获取附加文件路径列表
func (u *UI) GetAdditionalFilePaths() []string {
	return u.fileListManager.GetFilePaths()
}

// GetOutputPath 获取输出路径
func (u *UI) GetOutputPath() string {
	return u.outputPath
}

// SetProgress 设置进度
func (u *UI) SetProgress(progress float64) {
	u.progressManager.SetProgress(progress)
}

// SetStatus 设置状态文本
func (u *UI) SetStatus(status string) {
	u.progressManager.SetStatus(status)
}

// SetDetail 设置详细信息
func (u *UI) SetDetail(detail string) {
	u.progressManager.SetDetail(detail)
}

// UpdateProgress 更新进度信息
func (u *UI) UpdateProgress(info ProgressInfo) {
	u.progressManager.UpdateProgress(info)
}

// ShowError 显示错误对话框
func (u *UI) ShowError(err error) {
	dialog.ShowError(err, u.window)
}

// ShowInfo 显示信息对话框
func (u *UI) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, u.window)
}

// SetEventHandler 设置事件处理器
func (u *UI) SetEventHandler(handler interface{}) {
	u.eventHandler = handler
}

// EnableControls 启用控件
func (u *UI) EnableControls() {
	u.enableInputControls()
}

// DisableControls 禁用控件
func (u *UI) DisableControls() {
	u.disableInputControls()
}

// UpdateProgressWithStrings 更新进度（支持字符串参数）
func (u *UI) UpdateProgressWithStrings(progress float64, status, detail string) {
	u.progressManager.UpdateProgress(ProgressInfo{
		Progress: progress,
		Status:   status,
		Detail:   detail,
	})
}

// ShowCompletion 显示完成消息
func (u *UI) ShowCompletion(message string) {
	u.progressManager.Complete(message)
	dialog.ShowInformation("完成", message, u.window)
}