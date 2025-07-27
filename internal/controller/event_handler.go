package controller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/pdf"
)

// EventHandler 定义事件处理器，连接UI事件和控制器逻辑
type EventHandler struct {
	controller *Controller

	// UI状态回调
	onUIStateChanged func(enabled bool)
	onProgressUpdate func(progress float64, status, detail string)
	onError          func(err error)
	onCompletion     func(message string)
}

// NewEventHandler 创建新的事件处理器
func NewEventHandler(controller *Controller) *EventHandler {
	handler := &EventHandler{
		controller: controller,
	}

	// 设置控制器回调
	controller.SetProgressCallback(handler.handleProgress)
	controller.SetErrorCallback(handler.handleError)
	controller.SetCompletionCallback(handler.handleCompletion)

	return handler
}

// SetUIStateCallback 设置UI状态变更回调
func (eh *EventHandler) SetUIStateCallback(callback func(enabled bool)) {
	eh.onUIStateChanged = callback
}

// SetProgressUpdateCallback 设置进度更新回调
func (eh *EventHandler) SetProgressUpdateCallback(callback func(progress float64, status, detail string)) {
	eh.onProgressUpdate = callback
}

// SetErrorCallback 设置错误回调
func (eh *EventHandler) SetErrorCallback(callback func(err error)) {
	eh.onError = callback
}

// SetCompletionCallback 设置完成回调
func (eh *EventHandler) SetCompletionCallback(callback func(message string)) {
	eh.onCompletion = callback
}

// HandleMainFileSelected 处理主文件选择事件
func (eh *EventHandler) HandleMainFileSelected(filePath string) error {
	// 验证文件
	if err := eh.controller.ValidateFile(filePath); err != nil {
		return fmt.Errorf("主文件无效: %v", err)
	}

	return nil
}

// HandleAdditionalFileAdded 处理附加文件添加事件
func (eh *EventHandler) HandleAdditionalFileAdded(filePath string) (*model.FileEntry, error) {
	// 验证文件
	if err := eh.controller.ValidateFile(filePath); err != nil {
		return nil, fmt.Errorf("文件无效: %v", err)
	}

	// 获取文件信息
	fileInfo, err := eh.controller.FileManager.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 创建文件条目
	entry := &model.FileEntry{
		Path:        filePath,
		DisplayName: filepath.Base(filePath),
		Size:        fileInfo.Size,
		IsValid:     true,
	}

	// 获取PDF信息
	if pdfInfo, err := eh.controller.GetPDFInfo(filePath); err == nil {
		entry.PageCount = pdfInfo.PageCount
		entry.IsEncrypted = pdfInfo.IsEncrypted
	} else {
		entry.SetError(err.Error())
	}

	return entry, nil
}

// HandleFileValidation 处理文件验证事件
func (eh *EventHandler) HandleFileValidation(filePath string) (*model.FileEntry, error) {
	return eh.HandleAdditionalFileAdded(filePath)
}

// HandleMergeStart 处理合并开始事件
func (eh *EventHandler) HandleMergeStart(mainFile string, additionalFiles []string, outputPath string) error {
	// 检查是否已有任务在运行
	if eh.controller.IsJobRunning() {
		return fmt.Errorf("已有合并任务正在运行，请等待完成或取消当前任务")
	}

	// 基本验证
	if mainFile == "" {
		return fmt.Errorf("请选择主PDF文件")
	}

	if len(additionalFiles) == 0 {
		return fmt.Errorf("请至少添加一个附加PDF文件")
	}

	if outputPath == "" {
		return fmt.Errorf("请选择输出文件路径")
	}

	// 禁用UI
	eh.notifyUIStateChanged(false)

	// 开始异步合并任务
	if err := eh.controller.StartMergeJob(mainFile, additionalFiles, outputPath); err != nil {
		// 如果启动失败，重新启用UI
		eh.notifyUIStateChanged(true)
		return err
	}

	return nil
}

// HandleMergeCancel 处理合并取消事件
func (eh *EventHandler) HandleMergeCancel() error {
	if err := eh.controller.CancelCurrentJob(); err != nil {
		return err
	}

	// 重新启用UI
	eh.notifyUIStateChanged(true)

	return nil
}

// HandleFileRemoval 处理文件移除事件
func (eh *EventHandler) HandleFileRemoval(filePath string) error {
	// 这里可以添加文件移除的相关逻辑
	// 比如清理临时文件等
	return nil
}

// HandleOutputPathChanged 处理输出路径变更事件
func (eh *EventHandler) HandleOutputPathChanged(outputPath string) error {
	// 验证输出路径的目录是否存在
	dir := filepath.Dir(outputPath)
	if err := eh.controller.FileManager.EnsureDirectoryExists(dir); err != nil {
		return fmt.Errorf("输出目录无效: %v", err)
	}

	// 新增：检查目录写权限
	if err := checkDirectoryWritable(dir); err != nil {
		return &pdf.PDFError{
			Type:    pdf.ErrorPermission,
			Message: "输出目录不可写（只读目录）",
			File:    dir,
			Cause:   err,
		}
	}

	return nil
}

// GetJobStatus 获取当前任务状态
func (eh *EventHandler) GetJobStatus() *model.MergeJob {
	return eh.controller.GetCurrentJob()
}

// IsJobRunning 检查是否有任务正在运行
func (eh *EventHandler) IsJobRunning() bool {
	return eh.controller.IsJobRunning()
}

// 内部回调处理方法

// handleProgress 处理进度更新
func (eh *EventHandler) handleProgress(progress float64, status, detail string) {
	if eh.onProgressUpdate != nil {
		eh.onProgressUpdate(progress, status, detail)
	}
}

// handleError 处理错误
func (eh *EventHandler) handleError(err error) {
	// 重新启用UI
	eh.notifyUIStateChanged(true)

	if eh.onError != nil {
		eh.onError(err)
	}
}

// handleCompletion 处理完成
func (eh *EventHandler) handleCompletion(outputPath string) {
	// 重新启用UI
	eh.notifyUIStateChanged(true)

	message := fmt.Sprintf("PDF合并完成！\n输出文件: %s", outputPath)
	if eh.onCompletion != nil {
		eh.onCompletion(message)
	}
}

// notifyUIStateChanged 通知UI状态变更
func (eh *EventHandler) notifyUIStateChanged(enabled bool) {
	if eh.onUIStateChanged != nil {
		eh.onUIStateChanged(enabled)
	}
}

// ValidateAllFiles 验证所有文件
func (eh *EventHandler) ValidateAllFiles(mainFile string, additionalFiles []string) map[string]error {
	allFiles := append([]string{mainFile}, additionalFiles...)
	return eh.controller.ValidateFiles(allFiles)
}

// GetFileInfo 获取文件信息
func (eh *EventHandler) GetFileInfo(filePath string) (*model.FileEntry, error) {
	return eh.HandleFileValidation(filePath)
}

// checkDirectoryWritable 检查目录是否可写
func checkDirectoryWritable(dir string) error {
	// 尝试创建测试文件
	testFile := filepath.Join(dir, ".test_write")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}
