package model

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewProgressTracker(t *testing.T) {
	totalSteps := 5
	tracker := NewProgressTracker(totalSteps)

	if tracker.totalSteps != totalSteps {
		t.Errorf("Expected totalSteps %d, got %d", totalSteps, tracker.totalSteps)
	}

	if tracker.currentStep != 0 {
		t.Errorf("Expected currentStep 0, got %d", tracker.currentStep)
	}

	if tracker.stepProgress != 0 {
		t.Errorf("Expected stepProgress 0, got %f", tracker.stepProgress)
	}
}

func TestProgressTracker_SetCurrentStep(t *testing.T) {
	tracker := NewProgressTracker(3)
	step := 2
	message := "Processing step 2"

	tracker.SetCurrentStep(step, message)

	info := tracker.GetProgress()
	if info.CurrentStep != step {
		t.Errorf("Expected CurrentStep %d, got %d", step, info.CurrentStep)
	}

	if info.Message != message {
		t.Errorf("Expected Message %s, got %s", message, info.Message)
	}

	if info.StepProgress != 0 {
		t.Errorf("Expected StepProgress 0, got %f", info.StepProgress)
	}
}

func TestProgressTracker_UpdateStepProgress(t *testing.T) {
	tracker := NewProgressTracker(2)
	tracker.SetCurrentStep(1, "Step 1")

	// 测试正常进度更新
	progress := 50.0
	message := "Half done"
	tracker.UpdateStepProgress(progress, message)

	info := tracker.GetProgress()
	if info.StepProgress != progress {
		t.Errorf("Expected StepProgress %f, got %f", progress, info.StepProgress)
	}

	if info.Message != message {
		t.Errorf("Expected Message %s, got %s", message, info.Message)
	}

	// 测试负数进度
	tracker.UpdateStepProgress(-10, "")
	info = tracker.GetProgress()
	if info.StepProgress != 0 {
		t.Errorf("Expected StepProgress 0 for negative input, got %f", info.StepProgress)
	}

	// 测试超过100的进度
	tracker.UpdateStepProgress(150, "")
	info = tracker.GetProgress()
	if info.StepProgress != 100 {
		t.Errorf("Expected StepProgress 100 for >100 input, got %f", info.StepProgress)
	}
}

func TestProgressTracker_Complete(t *testing.T) {
	tracker := NewProgressTracker(3)
	message := "All done"

	tracker.Complete(message)

	info := tracker.GetProgress()
	if !info.IsCompleted {
		t.Error("Expected IsCompleted to be true")
	}

	if info.CurrentStep != tracker.totalSteps {
		t.Errorf("Expected CurrentStep %d, got %d", tracker.totalSteps, info.CurrentStep)
	}

	if info.StepProgress != 100 {
		t.Errorf("Expected StepProgress 100, got %f", info.StepProgress)
	}

	if info.Message != message {
		t.Errorf("Expected Message %s, got %s", message, info.Message)
	}
}

func TestProgressTracker_Cancel(t *testing.T) {
	tracker := NewProgressTracker(3)
	message := "Cancelled by user"

	tracker.Cancel(message)

	info := tracker.GetProgress()
	if !info.IsCancelled {
		t.Error("Expected IsCancelled to be true")
	}

	if info.Message != message {
		t.Errorf("Expected Message %s, got %s", message, info.Message)
	}
}

func TestProgressTracker_TotalProgress(t *testing.T) {
	tracker := NewProgressTracker(4)

	// 第一步完成50%
	tracker.SetCurrentStep(1, "Step 1")
	tracker.UpdateStepProgress(50, "")

	info := tracker.GetProgress()
	expectedTotal := (0.0 + 0.5) / 4.0 * 100.0 // (step-1 + stepProgress/100) / totalSteps * 100
	if info.TotalProgress != expectedTotal {
		t.Errorf("Expected TotalProgress %f, got %f", expectedTotal, info.TotalProgress)
	}

	// 第二步完成100%
	tracker.SetCurrentStep(2, "Step 2")
	tracker.UpdateStepProgress(100, "")

	info = tracker.GetProgress()
	expectedTotal = (1.0 + 1.0) / 4.0 * 100.0 // (step-1 + stepProgress/100) / totalSteps * 100
	if info.TotalProgress != expectedTotal {
		t.Errorf("Expected TotalProgress %f, got %f", expectedTotal, info.TotalProgress)
	}
}

func TestProgressTracker_AddCallback(t *testing.T) {
	tracker := NewProgressTracker(2)
	var callbackCalled int32
	var receivedMessage string
	var mu sync.Mutex

	callback := func(progress float64, message string) {
		atomic.StoreInt32(&callbackCalled, 1)
		mu.Lock()
		receivedMessage = message
		mu.Unlock()
		// 我们不检查进度值，只检查消息
	}

	tracker.AddCallback(callback)
	tracker.SetCurrentStep(1, "Test message")

	// 给回调一些时间执行（因为是异步的）
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&callbackCalled) == 0 {
		t.Error("Expected callback to be called")
	}

	mu.Lock()
	msg := receivedMessage
	mu.Unlock()

	if msg != "Test message" {
		t.Errorf("Expected callback message 'Test message', got %s", msg)
	}
}

func TestProgressTracker_ElapsedTime(t *testing.T) {
	tracker := NewProgressTracker(1)

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	info := tracker.GetProgress()
	if info.ElapsedTime < 10*time.Millisecond {
		t.Errorf("Expected ElapsedTime >= 10ms, got %v", info.ElapsedTime)
	}
}
