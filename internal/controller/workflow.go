package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// WorkflowStep 定义工作流程步骤
type WorkflowStep int

const (
	StepValidation WorkflowStep = iota
	StepPreparation
	StepDecryption
	StepMerging
	StepFinalization
	StepCompleted
)

// String 返回工作流程步骤的字符串表示
func (ws WorkflowStep) String() string {
	switch ws {
	case StepValidation:
		return "文件验证"
	case StepPreparation:
		return "准备合并"
	case StepDecryption:
		return "处理加密文件"
	case StepMerging:
		return "合并文件"
	case StepFinalization:
		return "完成处理"
	case StepCompleted:
		return "已完成"
	default:
		return "未知步骤"
	}
}

// WorkflowManager 管理合并工作流程
type WorkflowManager struct {
	controller     *Controller
	currentStep    WorkflowStep
	stepMutex      sync.RWMutex
	retryCount     map[string]int
	retryMutex     sync.RWMutex
	maxRetries     int
	memoryMonitor  *MemoryMonitor
}

// NewWorkflowManager 创建新的工作流程管理器
func NewWorkflowManager(controller *Controller) *WorkflowManager {
	return &WorkflowManager{
		controller:    controller,
		currentStep:   StepValidation,
		retryCount:    make(map[string]int),
		maxRetries:    3,
		memoryMonitor: NewMemoryMonitor(controller.Config.MaxMemoryUsage),
	}
}

// ExecuteWorkflow 执行完整的合并工作流程
func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context, job *model.MergeJob) error {
	// 重置状态
	wm.resetWorkflow()
	
	// 启动内存监控
	wm.memoryMonitor.Start()
	defer wm.memoryMonitor.Stop()
	
	// 执行各个步骤
	steps := []struct {
		step     WorkflowStep
		progress float64
		handler  func(context.Context, *model.MergeJob) error
	}{
		{StepValidation, 0.0, wm.executeValidation},
		{StepPreparation, 0.2, wm.executePreparation},
		{StepDecryption, 0.3, wm.executeDecryption},
		{StepMerging, 0.4, wm.executeMerging},
		{StepFinalization, 0.9, wm.executeFinalization},
	}
	
	for _, stepInfo := range steps {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// 设置当前步骤
		wm.setCurrentStep(stepInfo.step)
		
		// 更新进度
		wm.controller.notifyProgress(stepInfo.progress, stepInfo.step.String(), 
			fmt.Sprintf("正在执行: %s", stepInfo.step.String()))
		
		// 执行步骤
		if err := wm.executeStepWithRetry(ctx, job, stepInfo.handler); err != nil {
			return fmt.Errorf("%s失败: %v", stepInfo.step.String(), err)
		}
	}
	
	// 标记完成
	wm.setCurrentStep(StepCompleted)
	wm.controller.notifyProgress(1.0, StepCompleted.String(), "合并操作已完成")
	
	return nil
}

// executeStepWithRetry 执行步骤并支持重试
func (wm *WorkflowManager) executeStepWithRetry(ctx context.Context, job *model.MergeJob, 
	handler func(context.Context, *model.MergeJob) error) error {
	
	stepName := wm.getCurrentStep().String()
	
	for attempt := 0; attempt <= wm.maxRetries; attempt++ {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// 执行步骤
		err := handler(ctx, job)
		if err == nil {
			// 成功，重置重试计数
			wm.resetRetryCount(stepName)
			return nil
		}
		
		// 检查是否可以重试
		if !wm.shouldRetry(err) || attempt >= wm.maxRetries {
			return err
		}
		
		// 记录重试
		wm.incrementRetryCount(stepName)
		
		// 通知重试
		wm.controller.notifyProgress(
			wm.getStepProgress(),
			fmt.Sprintf("%s (重试 %d/%d)", stepName, attempt+1, wm.maxRetries),
			fmt.Sprintf("重试原因: %v", err),
		)
		
		// 等待后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
			// 继续重试
		}
	}
	
	return fmt.Errorf("步骤 %s 在 %d 次重试后仍然失败", stepName, wm.maxRetries)
}

// executeValidation 执行文件验证步骤
func (wm *WorkflowManager) executeValidation(ctx context.Context, job *model.MergeJob) error {
	allFiles := append([]string{job.MainFile}, job.AdditionalFiles...)
	totalFiles := len(allFiles)
	
	for i, filePath := range allFiles {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// 更新进度
		progress := 0.2 * float64(i) / float64(totalFiles)
		wm.controller.notifyProgress(progress, "验证文件", 
			fmt.Sprintf("正在验证: %s", filepath.Base(filePath)))
		
		// 验证文件
		if err := wm.controller.ValidateFile(filePath); err != nil {
			return fmt.Errorf("文件验证失败 %s: %v", filepath.Base(filePath), err)
		}
		
		// 模拟验证时间
		time.Sleep(10 * time.Millisecond)
	}
	
	return nil
}

// executePreparation 执行准备步骤
func (wm *WorkflowManager) executePreparation(ctx context.Context, job *model.MergeJob) error {
	// 检查输出目录
	outputDir := filepath.Dir(job.OutputPath)
	if err := wm.controller.FileManager.EnsureDirectoryExists(outputDir); err != nil {
		return fmt.Errorf("无法创建输出目录: %v", err)
	}
	
	// 检查内存使用情况
	if wm.memoryMonitor.IsMemoryLow() {
		wm.controller.notifyProgress(0.25, "内存优化", "内存使用较高，启用流式处理模式")
		// 这里可以设置流式处理标志
	}
	
	// 预估合并时间和资源需求
	totalSize := int64(0)
	for _, filePath := range append([]string{job.MainFile}, job.AdditionalFiles...) {
		if info, err := wm.controller.FileManager.GetFileInfo(filePath); err == nil {
			totalSize += info.Size
		}
	}
	
	wm.controller.notifyProgress(0.3, "准备合并", 
		fmt.Sprintf("准备合并 %d 个文件，总大小: %.2f MB", 
			len(job.AdditionalFiles)+1, float64(totalSize)/(1024*1024)))
	
	return nil
}

// executeDecryption 执行解密步骤
func (wm *WorkflowManager) executeDecryption(ctx context.Context, job *model.MergeJob) error {
	allFiles := append([]string{job.MainFile}, job.AdditionalFiles...)
	encryptedFiles := []string{}
	
	// 检查哪些文件需要解密
	for _, filePath := range allFiles {
		if encrypted, err := wm.controller.PDFService.IsPDFEncrypted(filePath); err == nil && encrypted {
			encryptedFiles = append(encryptedFiles, filePath)
		}
	}
	
	if len(encryptedFiles) == 0 {
		wm.controller.notifyProgress(0.4, "跳过解密", "没有加密文件需要处理")
		return nil
	}
	
	// 处理加密文件
	for i, filePath := range encryptedFiles {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		progress := 0.3 + (0.1 * float64(i) / float64(len(encryptedFiles)))
		wm.controller.notifyProgress(progress, "处理加密文件", 
			fmt.Sprintf("正在处理: %s", filepath.Base(filePath)))
		
		// 这里应该调用解密服务，但由于我们还没有实现完整的解密功能，
		// 暂时跳过实际解密，只是记录需要处理的文件
		wm.controller.notifyProgress(progress+0.01, "解密文件", 
			fmt.Sprintf("文件 %s 需要密码", filepath.Base(filePath)))
	}
	
	return nil
}

// executeMerging 执行合并步骤
func (wm *WorkflowManager) executeMerging(ctx context.Context, job *model.MergeJob) error {
	// 创建进度写入器
	progressWriter := &WorkflowProgressWriter{
		workflow:     wm,
		baseProgress: 0.4,
		maxProgress:  0.9,
		totalFiles:   len(job.AdditionalFiles) + 1,
	}
	
	// 检查内存使用情况，决定使用流式处理还是常规处理
	if wm.memoryMonitor.IsMemoryLow() {
		wm.controller.notifyProgress(0.5, "流式合并", "使用内存优化模式进行合并")
		return wm.executeStreamingMerge(ctx, job, progressWriter)
	} else {
		wm.controller.notifyProgress(0.5, "标准合并", "使用标准模式进行合并")
		return wm.executeStandardMerge(ctx, job, progressWriter)
	}
}

// executeStreamingMerge 执行流式合并
func (wm *WorkflowManager) executeStreamingMerge(ctx context.Context, job *model.MergeJob, 
	progressWriter *WorkflowProgressWriter) error {
	
	// 创建流式合并器
	streamingMerger := NewStreamingMerger(wm.controller)
	
	// 执行流式合并
	return streamingMerger.MergeStreaming(ctx, job, progressWriter)
}

// executeStandardMerge 执行标准合并
func (wm *WorkflowManager) executeStandardMerge(ctx context.Context, job *model.MergeJob, 
	progressWriter *WorkflowProgressWriter) error {
	
	// 执行合并
	err := wm.controller.PDFService.MergePDFs(job.MainFile, job.AdditionalFiles, job.OutputPath, progressWriter)
	if err != nil {
		return fmt.Errorf("合并失败: %v", err)
	}
	
	return nil
}

// executeFinalization 执行完成步骤
func (wm *WorkflowManager) executeFinalization(ctx context.Context, job *model.MergeJob) error {
	// 验证输出文件
	if err := wm.controller.ValidateFile(job.OutputPath); err != nil {
		return fmt.Errorf("输出文件验证失败: %v", err)
	}
	
	// 获取输出文件信息
	if info, err := wm.controller.FileManager.GetFileInfo(job.OutputPath); err == nil {
		wm.controller.notifyProgress(0.95, "验证输出", 
			fmt.Sprintf("输出文件大小: %.2f MB", float64(info.Size)/(1024*1024)))
	}
	
	// 清理临时文件
	if err := wm.controller.FileManager.CleanupTempFiles(); err != nil {
		// 清理失败不应该导致整个操作失败，只记录警告
		wm.controller.notifyProgress(0.98, "清理警告", 
			fmt.Sprintf("临时文件清理失败: %v", err))
	}
	
	return nil
}

// shouldRetry 判断错误是否可以重试
func (wm *WorkflowManager) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	
	// 检查错误类型，某些错误不应该重试
	errorStr := err.Error()
	
	// 不可重试的错误
	nonRetryableErrors := []string{
		"文件不存在",
		"权限被拒绝",
		"无效的PDF格式",
		"用户取消",
	}
	
	for _, nonRetryable := range nonRetryableErrors {
		if contains(errorStr, nonRetryable) {
			return false
		}
	}
	
	// 可重试的错误
	retryableErrors := []string{
		"网络错误",
		"临时文件",
		"内存不足",
		"IO错误",
		"超时",
	}
	
	for _, retryable := range retryableErrors {
		if contains(errorStr, retryable) {
			return true
		}
	}
	
	// 默认可以重试
	return true
}

// 辅助方法

func (wm *WorkflowManager) resetWorkflow() {
	wm.stepMutex.Lock()
	defer wm.stepMutex.Unlock()
	wm.currentStep = StepValidation
	
	wm.retryMutex.Lock()
	defer wm.retryMutex.Unlock()
	wm.retryCount = make(map[string]int)
}

func (wm *WorkflowManager) setCurrentStep(step WorkflowStep) {
	wm.stepMutex.Lock()
	defer wm.stepMutex.Unlock()
	wm.currentStep = step
}

func (wm *WorkflowManager) getCurrentStep() WorkflowStep {
	wm.stepMutex.RLock()
	defer wm.stepMutex.RUnlock()
	return wm.currentStep
}

func (wm *WorkflowManager) getStepProgress() float64 {
	step := wm.getCurrentStep()
	switch step {
	case StepValidation:
		return 0.1
	case StepPreparation:
		return 0.25
	case StepDecryption:
		return 0.35
	case StepMerging:
		return 0.65
	case StepFinalization:
		return 0.95
	case StepCompleted:
		return 1.0
	default:
		return 0.0
	}
}

func (wm *WorkflowManager) incrementRetryCount(step string) {
	wm.retryMutex.Lock()
	defer wm.retryMutex.Unlock()
	wm.retryCount[step]++
}

func (wm *WorkflowManager) resetRetryCount(step string) {
	wm.retryMutex.Lock()
	defer wm.retryMutex.Unlock()
	delete(wm.retryCount, step)
}

// WorkflowProgressWriter 工作流程进度写入器
type WorkflowProgressWriter struct {
	workflow     *WorkflowManager
	baseProgress float64
	maxProgress  float64
	totalFiles   int
	currentFile  int
}

func (wpw *WorkflowProgressWriter) Write(p []byte) (n int, err error) {
	// 更新进度
	wpw.currentFile++
	progress := wpw.baseProgress + 
		(wpw.maxProgress-wpw.baseProgress)*float64(wpw.currentFile)/float64(wpw.totalFiles)
	
	wpw.workflow.controller.notifyProgress(progress, "合并文件", 
		fmt.Sprintf("正在处理第 %d/%d 个文件", wpw.currentFile, wpw.totalFiles))
	
	return len(p), nil
}

// MemoryMonitor 内存监控器
type MemoryMonitor struct {
	maxMemory     int64
	checkInterval time.Duration
	stopChan      chan bool
	isRunning     bool
	mutex         sync.RWMutex
}

// NewMemoryMonitor 创建新的内存监控器
func NewMemoryMonitor(maxMemory int64) *MemoryMonitor {
	return &MemoryMonitor{
		maxMemory:     maxMemory,
		checkInterval: 1 * time.Second,
		stopChan:      make(chan bool),
	}
}

// Start 启动内存监控
func (mm *MemoryMonitor) Start() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	if mm.isRunning {
		return
	}
	
	mm.isRunning = true
	go mm.monitor()
}

// Stop 停止内存监控
func (mm *MemoryMonitor) Stop() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	if !mm.isRunning {
		return
	}
	
	mm.isRunning = false
	mm.stopChan <- true
}

// IsMemoryLow 检查内存是否不足
func (mm *MemoryMonitor) IsMemoryLow() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	currentMemory := int64(m.Alloc)
	return currentMemory > mm.maxMemory*80/100 // 超过80%认为内存不足
}

// monitor 监控内存使用情况
func (mm *MemoryMonitor) monitor() {
	ticker := time.NewTicker(mm.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if mm.IsMemoryLow() {
				runtime.GC() // 触发垃圾回收
			}
		case <-mm.stopChan:
			return
		}
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
				len(s) > len(substr)*2)))
}