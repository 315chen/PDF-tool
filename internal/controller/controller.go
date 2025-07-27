package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// ProgressCallback 定义进度回调函数类型
type ProgressCallback func(progress float64, status string, detail string)

// ErrorCallback 定义错误回调函数类型
type ErrorCallback func(err error)

// CompletionCallback 定义完成回调函数类型
type CompletionCallback func(outputPath string)

// Controller 定义应用程序的主控制器
type Controller struct {
	PDFService  pdf.PDFService
	FileManager file.FileManager
	Config      *model.Config

	// 当前任务管理
	currentJob          *model.MergeJob
	jobMutex            sync.RWMutex
	cancelFunc          context.CancelFunc
	workflowManager     *WorkflowManager
	cancellationManager *CancellationManager

	// 回调函数
	progressCallback   ProgressCallback
	errorCallback      ErrorCallback
	completionCallback CompletionCallback
}

// NewController 创建一个新的控制器实例
func NewController(
	pdfService pdf.PDFService,
	fileManager file.FileManager,
	config *model.Config,
) *Controller {
	controller := &Controller{
		PDFService:  pdfService,
		FileManager: fileManager,
		Config:      config,
	}

	// 创建工作流程管理器
	controller.workflowManager = NewWorkflowManager(controller)

	// 创建取消管理器
	controller.cancellationManager = NewCancellationManager(controller)

	return controller
}

// SetProgressCallback 设置进度回调函数
func (c *Controller) SetProgressCallback(callback ProgressCallback) {
	c.progressCallback = callback
}

// SetErrorCallback 设置错误回调函数
func (c *Controller) SetErrorCallback(callback ErrorCallback) {
	c.errorCallback = callback
}

// SetCompletionCallback 设置完成回调函数
func (c *Controller) SetCompletionCallback(callback CompletionCallback) {
	c.completionCallback = callback
}

// ValidateFile 验证单个文件
func (c *Controller) ValidateFile(filePath string) error {
	// 首先验证文件是否存在和可访问
	if err := c.FileManager.ValidateFile(filePath); err != nil {
		return fmt.Errorf("文件访问失败: %v", err)
	}

	// 然后验证PDF格式
	return c.PDFService.ValidatePDF(filePath)
}

// ValidateFiles 验证多个文件
func (c *Controller) ValidateFiles(filePaths []string) map[string]error {
	results := make(map[string]error)

	for _, filePath := range filePaths {
		err := c.ValidateFile(filePath)
		if err != nil {
			results[filePath] = err
		}
	}

	return results
}

// GetPDFInfo 获取PDF文件信息
func (c *Controller) GetPDFInfo(filePath string) (*pdf.PDFInfo, error) {
	return c.PDFService.GetPDFInfo(filePath)
}

// GetCurrentJob 获取当前任务
func (c *Controller) GetCurrentJob() *model.MergeJob {
	c.jobMutex.RLock()
	defer c.jobMutex.RUnlock()
	return c.currentJob
}

// IsJobRunning 检查是否有任务正在运行
func (c *Controller) IsJobRunning() bool {
	c.jobMutex.RLock()
	defer c.jobMutex.RUnlock()
	// 修改：只要currentJob不为nil就认为有任务在运行，避免竞态条件
	return c.currentJob != nil
}

// StartMergeJob 开始合并任务（异步）
func (c *Controller) StartMergeJob(mainFile string, additionalFiles []string, outputPath string) error {
	// 检查是否已有任务在运行
	if c.IsJobRunning() {
		return fmt.Errorf("已有合并任务正在运行")
	}

	// 创建新任务
	job := model.NewMergeJob(mainFile, additionalFiles, outputPath)

	c.jobMutex.Lock()
	c.currentJob = job
	c.jobMutex.Unlock()

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	// 注册取消操作
	c.cancellationManager.RegisterCancellation(job.ID, cancel)

	// 添加清理任务
	c.cancellationManager.AddCleanupTask(NewTempFileCleanupTask(c.FileManager))
	c.cancellationManager.AddCleanupTask(NewMemoryCleanupTask())
	c.cancellationManager.AddCleanupTask(NewJobStateCleanupTask(c))

	// 异步执行合并
	go c.executeMergeJob(ctx, job)

	return nil
}

// CancelCurrentJob 取消当前任务
func (c *Controller) CancelCurrentJob() error {
	c.jobMutex.RLock()
	currentJob := c.currentJob
	c.jobMutex.RUnlock()

	if currentJob == nil {
		return fmt.Errorf("没有正在运行的任务")
	}

	// 使用取消管理器进行优雅取消
	return c.cancellationManager.GracefulCancellation(currentJob.ID, 5*time.Second)
}

// executeMergeJob 执行合并任务的内部方法
func (c *Controller) executeMergeJob(ctx context.Context, job *model.MergeJob) {
	defer func() {
		c.jobMutex.Lock()
		// 只有在任务完成或失败时才清空当前任务
		if c.currentJob != nil && (c.currentJob.Status == model.JobCompleted || c.currentJob.Status == model.JobFailed) {
			c.currentJob = nil
		}
		c.cancelFunc = nil
		c.jobMutex.Unlock()
	}()

	// 标记任务开始
	c.jobMutex.Lock()
	job.SetRunning()
	c.jobMutex.Unlock()

	c.notifyProgress(0.0, "开始合并", "正在启动合并工作流程...")

	// 使用工作流程管理器执行完整的合并流程
	if err := c.workflowManager.ExecuteWorkflow(ctx, job); err != nil {
		c.jobMutex.Lock()
		job.SetFailed(err)
		c.jobMutex.Unlock()
		c.notifyError(err)
		return
	}

	// 检查取消
	if ctx.Err() != nil {
		return
	}

	// 标记任务完成
	c.jobMutex.Lock()
	job.SetCompleted()
	c.jobMutex.Unlock()

	c.notifyCompletion(job.OutputPath)
}

// validateJobFiles 验证任务中的所有文件
func (c *Controller) validateJobFiles(ctx context.Context, job *model.MergeJob) error {
	allFiles := append([]string{job.MainFile}, job.AdditionalFiles...)
	totalFiles := len(allFiles)

	for i, filePath := range allFiles {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 更新进度
		progress := 0.2 * float64(i) / float64(totalFiles)
		c.notifyProgress(progress, "验证文件", fmt.Sprintf("正在验证: %s", filePath))

		// 验证文件
		if err := c.ValidateFile(filePath); err != nil {
			return fmt.Errorf("文件验证失败 %s: %v", filePath, err)
		}

		// 减少模拟验证时间
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// performMerge 执行实际的合并操作
func (c *Controller) performMerge(ctx context.Context, job *model.MergeJob) error {
	// 创建进度写入器
	progressWriter := &progressWriter{
		controller:   c,
		baseProgress: 0.3,
		maxProgress:  0.9,
	}

	// 执行合并
	err := c.PDFService.MergePDFs(job.MainFile, job.AdditionalFiles, job.OutputPath, progressWriter)
	if err != nil {
		return fmt.Errorf("合并失败: %v", err)
	}

	return nil
}

// notifyProgress 通知进度更新
func (c *Controller) notifyProgress(progress float64, status, detail string) {
	if c.progressCallback != nil {
		c.progressCallback(progress, status, detail)
	}

	// 更新任务进度
	c.jobMutex.Lock()
	if c.currentJob != nil {
		c.currentJob.UpdateProgress(progress * 100)
	}
	c.jobMutex.Unlock()
}

// notifyError 通知错误
func (c *Controller) notifyError(err error) {
	if c.errorCallback != nil {
		c.errorCallback(err)
	}
}

// notifyCompletion 通知完成
func (c *Controller) notifyCompletion(outputPath string) {
	if c.completionCallback != nil {
		c.completionCallback(outputPath)
	}
}

// MergePDFs 执行PDF合并操作（同步版本，保持向后兼容）
func (c *Controller) MergePDFs(mainFile string, additionalFiles []string, outputPath string) error {
	// 验证主文件
	if err := c.ValidateFile(mainFile); err != nil {
		return fmt.Errorf("主文件验证失败: %v", err)
	}

	// 验证附加文件
	validationResults := c.ValidateFiles(additionalFiles)

	// 收集有效的文件
	validFiles := []string{mainFile}
	for _, filePath := range additionalFiles {
		if err, exists := validationResults[filePath]; !exists || err == nil {
			validFiles = append(validFiles, filePath)
		}
	}

	if len(validFiles) < 2 {
		return fmt.Errorf("至少需要两个有效的PDF文件进行合并")
	}

	// 执行合并
	return c.PDFService.MergePDFs(validFiles[0], validFiles[1:], outputPath, nil)
}

// progressWriter 实现io.Writer接口，用于接收合并进度
type progressWriter struct {
	controller   *Controller
	baseProgress float64
	maxProgress  float64
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	// 这里可以解析进度信息并更新进度
	// 目前简单地返回写入的字节数
	return len(p), nil
}
