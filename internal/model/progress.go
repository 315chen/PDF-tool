package model

import (
	"sync"
	"time"
)

// ProgressTracker 定义进度跟踪器
type ProgressTracker struct {
	mu            sync.RWMutex
	currentStep   int
	totalSteps    int
	stepProgress  float64
	message       string
	startTime     time.Time
	lastUpdate    time.Time
	isCompleted   bool
	isCancelled   bool
	callbacks     []ProgressCallback
}

// ProgressCallback 定义进度回调函数类型
type ProgressCallback func(progress float64, message string)

// ProgressInfo 定义进度信息
type ProgressInfo struct {
	CurrentStep  int
	TotalSteps   int
	StepProgress float64
	TotalProgress float64
	Message      string
	ElapsedTime  time.Duration
	IsCompleted  bool
	IsCancelled  bool
}

// NewProgressTracker 创建一个新的进度跟踪器
func NewProgressTracker(totalSteps int) *ProgressTracker {
	return &ProgressTracker{
		totalSteps: totalSteps,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// SetCurrentStep 设置当前步骤
func (pt *ProgressTracker) SetCurrentStep(step int, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.currentStep = step
	pt.stepProgress = 0
	pt.message = message
	pt.lastUpdate = time.Now()
	
	pt.notifyCallbacks()
}

// UpdateStepProgress 更新当前步骤的进度
func (pt *ProgressTracker) UpdateStepProgress(progress float64, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	
	pt.stepProgress = progress
	if message != "" {
		pt.message = message
	}
	pt.lastUpdate = time.Now()
	
	pt.notifyCallbacks()
}

// Complete 标记进度为完成
func (pt *ProgressTracker) Complete(message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.currentStep = pt.totalSteps
	pt.stepProgress = 100
	pt.isCompleted = true
	if message != "" {
		pt.message = message
	}
	pt.lastUpdate = time.Now()
	
	pt.notifyCallbacks()
}

// Cancel 取消进度
func (pt *ProgressTracker) Cancel(message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.isCancelled = true
	if message != "" {
		pt.message = message
	}
	pt.lastUpdate = time.Now()
	
	pt.notifyCallbacks()
}

// GetProgress 获取当前进度信息
func (pt *ProgressTracker) GetProgress() ProgressInfo {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	totalProgress := 0.0
	if pt.totalSteps > 0 {
		totalProgress = (float64(pt.currentStep-1) + pt.stepProgress/100.0) / float64(pt.totalSteps) * 100.0
	}
	
	return ProgressInfo{
		CurrentStep:   pt.currentStep,
		TotalSteps:    pt.totalSteps,
		StepProgress:  pt.stepProgress,
		TotalProgress: totalProgress,
		Message:       pt.message,
		ElapsedTime:   time.Since(pt.startTime),
		IsCompleted:   pt.isCompleted,
		IsCancelled:   pt.isCancelled,
	}
}

// AddCallback 添加进度回调
func (pt *ProgressTracker) AddCallback(callback ProgressCallback) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.callbacks = append(pt.callbacks, callback)
}

// notifyCallbacks 通知所有回调函数
func (pt *ProgressTracker) notifyCallbacks() {
	info := pt.getProgressUnsafe()
	for _, callback := range pt.callbacks {
		go callback(info.TotalProgress, info.Message)
	}
}

// getProgressUnsafe 获取进度信息（不加锁）
func (pt *ProgressTracker) getProgressUnsafe() ProgressInfo {
	totalProgress := 0.0
	if pt.totalSteps > 0 {
		totalProgress = (float64(pt.currentStep-1) + pt.stepProgress/100.0) / float64(pt.totalSteps) * 100.0
	}
	
	return ProgressInfo{
		CurrentStep:   pt.currentStep,
		TotalSteps:    pt.totalSteps,
		StepProgress:  pt.stepProgress,
		TotalProgress: totalProgress,
		Message:       pt.message,
		ElapsedTime:   time.Since(pt.startTime),
		IsCompleted:   pt.isCompleted,
		IsCancelled:   pt.isCancelled,
	}
}