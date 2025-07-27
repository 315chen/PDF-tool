package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProgressManager 进度管理器
type ProgressManager struct {
	window      fyne.Window
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
	detailLabel *widget.Label
	timeLabel   *widget.Label
	speedLabel  *widget.Label
	container   *fyne.Container

	// 进度状态
	isActive       bool
	startTime      time.Time
	currentStep    int
	totalSteps     int
	currentFile    string
	processedFiles int
	totalFiles     int

	// 回调函数
	onCancel   func()
	onComplete func()
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	Progress       float64 // 0.0 - 1.0
	Status         string
	Detail         string
	CurrentFile    string
	ProcessedFiles int
	TotalFiles     int
	Step           int
	TotalSteps     int
}

// NewProgressManager 创建新的进度管理器
func NewProgressManager(window fyne.Window) *ProgressManager {
	pm := &ProgressManager{
		window: window,
	}

	pm.createComponents()
	return pm
}

// createComponents 创建进度组件
func (pm *ProgressManager) createComponents() {
	// 创建进度条
	pm.progressBar = widget.NewProgressBar()
	pm.progressBar.SetValue(0)
	pm.progressBar.Hide()

	// 创建状态标签
	pm.statusLabel = widget.NewLabel("准备就绪")
	pm.statusLabel.Alignment = fyne.TextAlignCenter
	pm.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建详细信息标签
	pm.detailLabel = widget.NewLabel("")
	pm.detailLabel.Alignment = fyne.TextAlignCenter
	pm.detailLabel.TextStyle = fyne.TextStyle{Italic: true}
	pm.detailLabel.Hide()

	// 创建时间标签
	pm.timeLabel = widget.NewLabel("")
	pm.timeLabel.Alignment = fyne.TextAlignLeading
	pm.timeLabel.Hide()

	// 创建速度标签
	pm.speedLabel = widget.NewLabel("")
	pm.speedLabel.Alignment = fyne.TextAlignTrailing
	pm.speedLabel.Hide()

	// 创建信息行容器
	infoRow := container.NewBorder(nil, nil, pm.timeLabel, pm.speedLabel, nil)

	// 创建主容器
	pm.container = container.NewVBox(
		pm.progressBar,
		pm.statusLabel,
		pm.detailLabel,
		infoRow,
	)
}

// GetContainer 获取进度容器
func (pm *ProgressManager) GetContainer() *fyne.Container {
	return pm.container
}

// Start 开始进度显示
func (pm *ProgressManager) Start(totalSteps int, totalFiles int) {
	pm.isActive = true
	pm.startTime = time.Now()
	pm.totalSteps = totalSteps
	pm.totalFiles = totalFiles
	pm.currentStep = 0
	pm.processedFiles = 0

	pm.progressBar.SetValue(0)
	pm.progressBar.Show()
	pm.detailLabel.Show()
	pm.timeLabel.Show()
	pm.speedLabel.Show()

	pm.updateDisplay()

	// 启动定时更新
	go pm.startTimer()
}

// Stop 停止进度显示
func (pm *ProgressManager) Stop() {
	pm.isActive = false
	pm.progressBar.Hide()
	pm.detailLabel.Hide()
	pm.timeLabel.Hide()
	pm.speedLabel.Hide()

	pm.statusLabel.SetText("准备就绪")
}

// UpdateProgress 更新进度
func (pm *ProgressManager) UpdateProgress(info ProgressInfo) {
	if !pm.isActive {
		return
	}

	// 更新进度值
	pm.progressBar.SetValue(info.Progress)

	// 更新状态信息
	if info.Status != "" {
		pm.statusLabel.SetText(info.Status)
	}

	if info.Detail != "" {
		pm.detailLabel.SetText(info.Detail)
	}

	// 更新文件信息
	if info.CurrentFile != "" {
		pm.currentFile = info.CurrentFile
	}

	if info.ProcessedFiles > 0 {
		pm.processedFiles = info.ProcessedFiles
	}

	if info.TotalFiles > 0 {
		pm.totalFiles = info.TotalFiles
	}

	// 更新步骤信息
	if info.Step > 0 {
		pm.currentStep = info.Step
	}

	if info.TotalSteps > 0 {
		pm.totalSteps = info.TotalSteps
	}

	pm.updateDisplay()
}

// SetStatus 设置状态文本
func (pm *ProgressManager) SetStatus(status string) {
	pm.statusLabel.SetText(status)
}

// SetDetail 设置详细信息
func (pm *ProgressManager) SetDetail(detail string) {
	pm.detailLabel.SetText(detail)
}

// SetProgress 设置进度值
func (pm *ProgressManager) SetProgress(progress float64) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	pm.progressBar.SetValue(progress)
}

// updateDisplay 更新显示信息
func (pm *ProgressManager) updateDisplay() {
	if !pm.isActive {
		return
	}

	// 更新时间信息
	elapsed := time.Since(pm.startTime)
	pm.timeLabel.SetText(fmt.Sprintf("已用时: %s", formatDuration(elapsed)))

	// 更新速度信息
	if pm.processedFiles > 0 && elapsed.Seconds() > 0 {
		speed := float64(pm.processedFiles) / elapsed.Seconds()
		pm.speedLabel.SetText(fmt.Sprintf("速度: %.1f 文件/秒", speed))
	}

	// 更新详细信息（如果没有自定义详细信息）
	if pm.detailLabel.Text == "" {
		if pm.currentFile != "" {
			pm.detailLabel.SetText(fmt.Sprintf("正在处理: %s", pm.currentFile))
		} else if pm.totalFiles > 0 {
			pm.detailLabel.SetText(fmt.Sprintf("文件进度: %d/%d", pm.processedFiles, pm.totalFiles))
		}
	}
}

// startTimer 启动定时器
func (pm *ProgressManager) startTimer() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for pm.isActive {
		select {
		case <-ticker.C:
			pm.updateDisplay()
		}
	}
}

// Complete 完成进度
func (pm *ProgressManager) Complete(message string) {
	pm.progressBar.SetValue(1.0)
	pm.statusLabel.SetText(message)

	elapsed := time.Since(pm.startTime)
	pm.detailLabel.SetText(fmt.Sprintf("完成！总用时: %s", formatDuration(elapsed)))

	// 延迟隐藏进度信息
	go func() {
		time.Sleep(2 * time.Second)
		if pm.onComplete != nil {
			pm.onComplete()
		}
		pm.Stop()
	}()
}

// Error 显示错误
func (pm *ProgressManager) Error(err error) {
	pm.isActive = false
	pm.progressBar.Hide()

	pm.statusLabel.SetText("操作失败")
	pm.detailLabel.SetText(err.Error())
	pm.detailLabel.Show()

	// 显示错误对话框
	pm.ShowErrorDialog("操作失败", err.Error())

	// 延迟重置状态
	go func() {
		time.Sleep(3 * time.Second)
		pm.Stop()
	}()
}

// Cancel 取消操作
func (pm *ProgressManager) Cancel() {
	pm.isActive = false
	pm.statusLabel.SetText("已取消")
	pm.detailLabel.SetText("操作已被用户取消")

	if pm.onCancel != nil {
		pm.onCancel()
	}

	// 延迟重置状态
	go func() {
		time.Sleep(2 * time.Second)
		pm.Stop()
	}()
}

// SetOnCancel 设置取消回调
func (pm *ProgressManager) SetOnCancel(callback func()) {
	pm.onCancel = callback
}

// SetOnComplete 设置完成回调
func (pm *ProgressManager) SetOnComplete(callback func()) {
	pm.onComplete = callback
}

// IsActive 检查是否活跃
func (pm *ProgressManager) IsActive() bool {
	return pm.isActive
}

// GetProgress 获取当前进度
func (pm *ProgressManager) GetProgress() float64 {
	return pm.progressBar.Value
}

// GetElapsedTime 获取已用时间
func (pm *ProgressManager) GetElapsedTime() time.Duration {
	if pm.isActive {
		return time.Since(pm.startTime)
	}
	return 0
}

// ShowErrorDialog 显示错误对话框
func (pm *ProgressManager) ShowErrorDialog(title, message string) {
	dialog.ShowError(fmt.Errorf("%s", message), pm.window)
}

// ShowInfoDialog 显示信息对话框
func (pm *ProgressManager) ShowInfoDialog(title, message string) {
	dialog.ShowInformation(title, message, pm.window)
}

// ShowConfirmDialog 显示确认对话框
func (pm *ProgressManager) ShowConfirmDialog(title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, pm.window)
}

// ShowProgressDialog 显示进度对话框
func (pm *ProgressManager) ShowProgressDialog(title, message string, onCancel func()) {
	// 创建进度对话框内容
	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel(message)

	cancelBtn := widget.NewButton("取消", func() {
		if onCancel != nil {
			onCancel()
		}
	})

	content := container.NewVBox(
		statusLabel,
		progressBar,
		cancelBtn,
	)

	// 创建对话框
	progressDialog := dialog.NewCustom(title, "关闭", content, pm.window)
	progressDialog.Show()

	// 更新进度的函数
	updateProgress := func(progress float64, status string) {
		progressBar.SetValue(progress)
		statusLabel.SetText(status)
	}

	// 这里可以添加进度更新逻辑
	_ = updateProgress
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
}

// StatusType 状态类型
type StatusType int

const (
	StatusReady StatusType = iota
	StatusProcessing
	StatusCompleted
	StatusError
	StatusCancelled
)

// StatusMessage 状态消息
type StatusMessage struct {
	Type    StatusType
	Title   string
	Message string
	Details string
	Icon    fyne.Resource
}

// GetStatusMessage 获取状态消息
func GetStatusMessage(statusType StatusType, message string) StatusMessage {
	switch statusType {
	case StatusReady:
		return StatusMessage{
			Type:    StatusReady,
			Title:   "准备就绪",
			Message: message,
			Icon:    theme.InfoIcon(),
		}
	case StatusProcessing:
		return StatusMessage{
			Type:    StatusProcessing,
			Title:   "正在处理",
			Message: message,
			Icon:    theme.ComputerIcon(),
		}
	case StatusCompleted:
		return StatusMessage{
			Type:    StatusCompleted,
			Title:   "完成",
			Message: message,
			Icon:    theme.ConfirmIcon(),
		}
	case StatusError:
		return StatusMessage{
			Type:    StatusError,
			Title:   "错误",
			Message: message,
			Icon:    theme.ErrorIcon(),
		}
	case StatusCancelled:
		return StatusMessage{
			Type:    StatusCancelled,
			Title:   "已取消",
			Message: message,
			Icon:    theme.CancelIcon(),
		}
	default:
		return StatusMessage{
			Type:    StatusReady,
			Title:   "未知状态",
			Message: message,
			Icon:    theme.QuestionIcon(),
		}
	}
}
