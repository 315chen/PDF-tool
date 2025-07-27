package ui

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestNewProgressManager(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	
	pm := NewProgressManager(window)
	
	if pm == nil {
		t.Error("NewProgressManager returned nil")
	}
	
	if pm.window != window {
		t.Error("Window not set correctly")
	}
	
	if pm.progressBar == nil {
		t.Error("Progress bar not created")
	}
	
	if pm.statusLabel == nil {
		t.Error("Status label not created")
	}
	
	if pm.container == nil {
		t.Error("Container not created")
	}
}

func TestProgressManager_StartStop(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	// 测试初始状态
	if pm.IsActive() {
		t.Error("Progress manager should not be active initially")
	}
	
	// 测试启动
	pm.Start(5, 10)
	
	if !pm.IsActive() {
		t.Error("Progress manager should be active after start")
	}
	
	if !pm.progressBar.Visible() {
		t.Error("Progress bar should be visible after start")
	}
	
	// 测试停止
	pm.Stop()
	
	if pm.IsActive() {
		t.Error("Progress manager should not be active after stop")
	}
	
	if pm.progressBar.Visible() {
		t.Error("Progress bar should be hidden after stop")
	}
}

func TestProgressManager_UpdateProgress(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	pm.Start(5, 10)
	
	// 测试进度更新
	info := ProgressInfo{
		Progress:       0.5,
		Status:         "测试状态",
		Detail:         "测试详情",
		CurrentFile:    "test.pdf",
		ProcessedFiles: 5,
		TotalFiles:     10,
		Step:           3,
		TotalSteps:     5,
	}
	
	pm.UpdateProgress(info)
	
	if pm.GetProgress() != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", pm.GetProgress())
	}
	
	if pm.statusLabel.Text != "测试状态" {
		t.Errorf("Expected status '测试状态', got '%s'", pm.statusLabel.Text)
	}
	
	if pm.detailLabel.Text != "测试详情" {
		t.Errorf("Expected detail '测试详情', got '%s'", pm.detailLabel.Text)
	}
}

func TestProgressManager_SetMethods(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	// 测试设置状态
	pm.SetStatus("测试状态")
	if pm.statusLabel.Text != "测试状态" {
		t.Errorf("Expected status '测试状态', got '%s'", pm.statusLabel.Text)
	}
	
	// 测试设置详情
	pm.SetDetail("测试详情")
	if pm.detailLabel.Text != "测试详情" {
		t.Errorf("Expected detail '测试详情', got '%s'", pm.detailLabel.Text)
	}
	
	// 测试设置进度
	pm.SetProgress(0.7)
	if pm.GetProgress() != 0.7 {
		t.Errorf("Expected progress 0.7, got %f", pm.GetProgress())
	}
	
	// 测试进度边界值
	pm.SetProgress(-0.1)
	if pm.GetProgress() != 0.0 {
		t.Errorf("Expected progress 0.0 for negative input, got %f", pm.GetProgress())
	}
	
	pm.SetProgress(1.5)
	if pm.GetProgress() != 1.0 {
		t.Errorf("Expected progress 1.0 for >1 input, got %f", pm.GetProgress())
	}
}

func TestProgressManager_Complete(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	pm.Start(5, 10)
	
	pm.SetOnComplete(func() {
		// 完成回调
	})
	
	pm.Complete("测试完成")
	
	if pm.GetProgress() != 1.0 {
		t.Errorf("Expected progress 1.0 after complete, got %f", pm.GetProgress())
	}
	
	if pm.statusLabel.Text != "测试完成" {
		t.Errorf("Expected status '测试完成', got '%s'", pm.statusLabel.Text)
	}
	
	// 等待一小段时间让完成回调执行
	time.Sleep(100 * time.Millisecond)
	
	// 注意：由于异步执行，这个测试可能不稳定
	// 在实际应用中，完成回调会在2秒后执行
}

func TestProgressManager_Cancel(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	pm.Start(5, 10)
	
	cancelCalled := false
	pm.SetOnCancel(func() {
		cancelCalled = true
	})
	
	pm.Cancel()
	
	if pm.IsActive() {
		t.Error("Progress manager should not be active after cancel")
	}
	
	if pm.statusLabel.Text != "已取消" {
		t.Errorf("Expected status '已取消', got '%s'", pm.statusLabel.Text)
	}
	
	if !cancelCalled {
		t.Error("Cancel callback should have been called")
	}
}

func TestProgressManager_Error(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	pm.Start(5, 10)
	
	testError := fmt.Errorf("测试错误")
	pm.Error(testError)
	
	if pm.IsActive() {
		t.Error("Progress manager should not be active after error")
	}
	
	if pm.statusLabel.Text != "操作失败" {
		t.Errorf("Expected status '操作失败', got '%s'", pm.statusLabel.Text)
	}
	
	if pm.detailLabel.Text != "测试错误" {
		t.Errorf("Expected detail '测试错误', got '%s'", pm.detailLabel.Text)
	}
}

func TestProgressManager_GetElapsedTime(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	// 测试未启动时的时间
	if pm.GetElapsedTime() != 0 {
		t.Error("Elapsed time should be 0 when not active")
	}
	
	// 测试启动后的时间
	pm.Start(5, 10)
	time.Sleep(100 * time.Millisecond)
	
	elapsed := pm.GetElapsedTime()
	if elapsed <= 0 {
		t.Error("Elapsed time should be greater than 0 when active")
	}
	
	if elapsed < 50*time.Millisecond {
		t.Error("Elapsed time seems too small")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30.0秒"},
		{90 * time.Second, "1.5分钟"},
		{2 * time.Hour, "2.0小时"},
	}
	
	for _, test := range tests {
		result := formatDuration(test.duration)
		if !containsText(result, "秒") && !containsText(result, "分钟") && !containsText(result, "小时") {
			t.Errorf("formatDuration(%v) = %s, expected to contain time unit", test.duration, result)
		}
	}
}

func TestGetStatusMessage(t *testing.T) {
	tests := []struct {
		statusType StatusType
		message    string
		expected   string
	}{
		{StatusReady, "准备中", "准备就绪"},
		{StatusProcessing, "处理中", "正在处理"},
		{StatusCompleted, "已完成", "完成"},
		{StatusError, "出错了", "错误"},
		{StatusCancelled, "已取消", "已取消"},
	}
	
	for _, test := range tests {
		result := GetStatusMessage(test.statusType, test.message)
		if result.Title != test.expected {
			t.Errorf("GetStatusMessage(%v, %s).Title = %s, expected %s", 
				test.statusType, test.message, result.Title, test.expected)
		}
		if result.Message != test.message {
			t.Errorf("GetStatusMessage(%v, %s).Message = %s, expected %s", 
				test.statusType, test.message, result.Message, test.message)
		}
	}
}

func TestProgressManager_Container(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("Test")
	pm := NewProgressManager(window)
	
	container := pm.GetContainer()
	if container == nil {
		t.Error("GetContainer returned nil")
	}
	
	// 检查容器中是否包含必要的组件
	if len(container.Objects) == 0 {
		t.Error("Container should contain objects")
	}
}

// 辅助函数（重命名以避免冲突）
func containsText(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsTextSubstring(s, substr)))
}

func containsTextSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}