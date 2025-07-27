package controller

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// CancellationManager 取消操作管理器
type CancellationManager struct {
	controller     *Controller
	cancelRequests map[string]context.CancelFunc
	requestMutex   sync.RWMutex
	cleanupTasks   []CleanupTask
	cleanupMutex   sync.Mutex
}

// CleanupTask 清理任务接口
type CleanupTask interface {
	Execute() error
	Description() string
}

// NewCancellationManager 创建新的取消操作管理器
func NewCancellationManager(controller *Controller) *CancellationManager {
	return &CancellationManager{
		controller:     controller,
		cancelRequests: make(map[string]context.CancelFunc),
		cleanupTasks:   make([]CleanupTask, 0),
	}
}

// RegisterCancellation 注册取消操作
func (cm *CancellationManager) RegisterCancellation(jobID string, cancelFunc context.CancelFunc) {
	cm.requestMutex.Lock()
	defer cm.requestMutex.Unlock()
	cm.cancelRequests[jobID] = cancelFunc
}

// CancelJob 取消指定任务
func (cm *CancellationManager) CancelJob(jobID string) error {
	cm.requestMutex.Lock()
	cancelFunc, exists := cm.cancelRequests[jobID]
	if exists {
		delete(cm.cancelRequests, jobID)
	}
	cm.requestMutex.Unlock()
	
	if !exists {
		return fmt.Errorf("任务 %s 不存在或已完成", jobID)
	}
	
	// 执行取消
	cancelFunc()
	
	// 执行清理任务
	cm.executeCleanup()
	
	return nil
}

// CancelAllJobs 取消所有任务
func (cm *CancellationManager) CancelAllJobs() error {
	cm.requestMutex.Lock()
	cancelFuncs := make([]context.CancelFunc, 0, len(cm.cancelRequests))
	for _, cancelFunc := range cm.cancelRequests {
		cancelFuncs = append(cancelFuncs, cancelFunc)
	}
	cm.cancelRequests = make(map[string]context.CancelFunc)
	cm.requestMutex.Unlock()
	
	// 执行所有取消操作
	for _, cancelFunc := range cancelFuncs {
		cancelFunc()
	}
	
	// 执行清理任务
	cm.executeCleanup()
	
	return nil
}

// AddCleanupTask 添加清理任务
func (cm *CancellationManager) AddCleanupTask(task CleanupTask) {
	cm.cleanupMutex.Lock()
	defer cm.cleanupMutex.Unlock()
	cm.cleanupTasks = append(cm.cleanupTasks, task)
}

// executeCleanup 执行清理任务
func (cm *CancellationManager) executeCleanup() {
	cm.cleanupMutex.Lock()
	tasks := make([]CleanupTask, len(cm.cleanupTasks))
	copy(tasks, cm.cleanupTasks)
	cm.cleanupTasks = cm.cleanupTasks[:0] // 清空任务列表
	cm.cleanupMutex.Unlock()
	
	for _, task := range tasks {
		if err := task.Execute(); err != nil {
			// 清理失败不应该阻止其他清理任务
			fmt.Printf("清理任务失败 (%s): %v\n", task.Description(), err)
		}
	}
}

// GracefulCancellation 优雅取消操作
func (cm *CancellationManager) GracefulCancellation(jobID string, timeout time.Duration) error {
	// 首先尝试正常取消
	if err := cm.CancelJob(jobID); err != nil {
		return err
	}
	
	// 等待任务完成清理
	done := make(chan bool, 1)
	go func() {
		// 等待任务状态变为非运行状态
		for {
			if !cm.controller.IsJobRunning() {
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("取消操作超时")
	}
}

// 具体的清理任务实现

// TempFileCleanupTask 临时文件清理任务
type TempFileCleanupTask struct {
	fileManager interface {
		CleanupTempFiles() error
	}
}

func NewTempFileCleanupTask(fileManager interface {
	CleanupTempFiles() error
}) *TempFileCleanupTask {
	return &TempFileCleanupTask{fileManager: fileManager}
}

func (t *TempFileCleanupTask) Execute() error {
	return t.fileManager.CleanupTempFiles()
}

func (t *TempFileCleanupTask) Description() string {
	return "清理临时文件"
}

// MemoryCleanupTask 内存清理任务
type MemoryCleanupTask struct{}

func NewMemoryCleanupTask() *MemoryCleanupTask {
	return &MemoryCleanupTask{}
}

func (m *MemoryCleanupTask) Execute() error {
	// 触发垃圾回收
	runtime.GC()
	return nil
}

func (m *MemoryCleanupTask) Description() string {
	return "清理内存"
}

// JobStateCleanupTask 任务状态清理任务
type JobStateCleanupTask struct {
	controller *Controller
}

func NewJobStateCleanupTask(controller *Controller) *JobStateCleanupTask {
	return &JobStateCleanupTask{controller: controller}
}

func (j *JobStateCleanupTask) Execute() error {
	j.controller.jobMutex.Lock()
	defer j.controller.jobMutex.Unlock()
	
	if j.controller.currentJob != nil {
		j.controller.currentJob.Status = model.JobFailed
		j.controller.currentJob.Error = fmt.Errorf("任务被用户取消")
		j.controller.currentJob = nil
	}
	
	return nil
}

func (j *JobStateCleanupTask) Description() string {
	return "清理任务状态"
}

// ResourceCleanupTask 资源清理任务
type ResourceCleanupTask struct {
	resources []func() error
	name      string
}

func NewResourceCleanupTask(name string, resources ...func() error) *ResourceCleanupTask {
	return &ResourceCleanupTask{
		resources: resources,
		name:      name,
	}
}

func (r *ResourceCleanupTask) Execute() error {
	var lastError error
	for _, resource := range r.resources {
		if err := resource(); err != nil {
			lastError = err
		}
	}
	return lastError
}

func (r *ResourceCleanupTask) Description() string {
	return fmt.Sprintf("清理资源: %s", r.name)
}

// CancellationContext 取消上下文包装器
type CancellationContext struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	jobID          string
	cancelManager  *CancellationManager
	cleanupTasks   []CleanupTask
}

// NewCancellationContext 创建新的取消上下文
func NewCancellationContext(parent context.Context, jobID string, 
	cancelManager *CancellationManager) *CancellationContext {
	
	ctx, cancel := context.WithCancel(parent)
	
	cc := &CancellationContext{
		ctx:           ctx,
		cancelFunc:    cancel,
		jobID:         jobID,
		cancelManager: cancelManager,
		cleanupTasks:  make([]CleanupTask, 0),
	}
	
	// 注册到取消管理器
	cancelManager.RegisterCancellation(jobID, cc.Cancel)
	
	return cc
}

// Context 返回上下文
func (cc *CancellationContext) Context() context.Context {
	return cc.ctx
}

// Cancel 取消操作
func (cc *CancellationContext) Cancel() {
	cc.cancelFunc()
	
	// 执行本地清理任务
	for _, task := range cc.cleanupTasks {
		if err := task.Execute(); err != nil {
			fmt.Printf("本地清理任务失败 (%s): %v\n", task.Description(), err)
		}
	}
}

// AddCleanupTask 添加清理任务
func (cc *CancellationContext) AddCleanupTask(task CleanupTask) {
	cc.cleanupTasks = append(cc.cleanupTasks, task)
}

// IsCancelled 检查是否已取消
func (cc *CancellationContext) IsCancelled() bool {
	select {
	case <-cc.ctx.Done():
		return true
	default:
		return false
	}
}

// WaitForCancellation 等待取消信号
func (cc *CancellationContext) WaitForCancellation() <-chan struct{} {
	return cc.ctx.Done()
}

// CancellationAwareOperation 支持取消的操作接口
type CancellationAwareOperation interface {
	Execute(ctx *CancellationContext) error
	CanBeCancelled() bool
	EstimatedDuration() time.Duration
}

// OperationExecutor 操作执行器
type OperationExecutor struct {
	cancelManager *CancellationManager
}

// NewOperationExecutor 创建新的操作执行器
func NewOperationExecutor(cancelManager *CancellationManager) *OperationExecutor {
	return &OperationExecutor{cancelManager: cancelManager}
}

// ExecuteWithCancellation 执行支持取消的操作
func (oe *OperationExecutor) ExecuteWithCancellation(parent context.Context, jobID string, 
	operation CancellationAwareOperation) error {
	
	// 创建取消上下文
	cancelCtx := NewCancellationContext(parent, jobID, oe.cancelManager)
	defer cancelCtx.Cancel()
	
	// 如果操作不支持取消，直接执行
	if !operation.CanBeCancelled() {
		return operation.Execute(cancelCtx)
	}
	
	// 在单独的goroutine中执行操作
	errChan := make(chan error, 1)
	go func() {
		errChan <- operation.Execute(cancelCtx)
	}()
	
	// 等待操作完成或取消
	select {
	case err := <-errChan:
		return err
	case <-cancelCtx.WaitForCancellation():
		return fmt.Errorf("操作被取消")
	}
}