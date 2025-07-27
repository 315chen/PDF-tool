package test_utils

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMockPDFService_Basic(t *testing.T) {
	service := NewMockPDFService()

	// 测试设置和获取合并结果
	outputPath := "/test/output.pdf"
	expectedError := fmt.Errorf("test error")
	service.SetMergeResult(outputPath, expectedError)

	// 测试合并操作
	ctx := context.Background()
	err := service.Merge(ctx, "/test/main.pdf", []string{"/test/file1.pdf"}, outputPath, nil)

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	// 验证调用次数
	if count := service.GetCallCount("Merge"); count != 1 {
		t.Errorf("Expected 1 call to Merge, got %d", count)
	}
}

func TestMockPDFService_WithProgress(t *testing.T) {
	service := NewMockPDFService()
	service.SetMergeDelay(100 * time.Millisecond)

	var progressUpdates []float64
	var mu sync.Mutex
	progressCallback := func(progress float64) {
		mu.Lock()
		progressUpdates = append(progressUpdates, progress)
		mu.Unlock()
	}

	ctx := context.Background()
	err := service.Merge(ctx, "/test/main.pdf", []string{"/test/file1.pdf"}, "/test/output.pdf", progressCallback)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 等待进度更新完成
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	updateCount := len(progressUpdates)
	mu.Unlock()

	if updateCount == 0 {
		t.Error("Expected progress updates, got none")
	}
}

func TestMockPDFService_Validate(t *testing.T) {
	service := NewMockPDFService()

	// 测试设置验证结果
	filePath := "/test/file.pdf"
	expectedError := fmt.Errorf("validation error")
	service.SetValidateResult(filePath, expectedError)

	err := service.Validate(filePath)

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	// 验证调用次数
	if count := service.GetCallCount("Validate"); count != 1 {
		t.Errorf("Expected 1 call to Validate, got %d", count)
	}
}

func TestMockPDFService_GetInfo(t *testing.T) {
	service := NewMockPDFService()

	// 测试默认信息
	info, err := service.GetInfo("/test/file.pdf")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if info.PageCount != 10 {
		t.Errorf("Expected page count 10, got %d", info.PageCount)
	}

	// 验证调用次数
	if count := service.GetCallCount("GetInfo"); count != 1 {
		t.Errorf("Expected 1 call to GetInfo, got %d", count)
	}
}

func TestMockFileManager_Basic(t *testing.T) {
	manager := NewMockFileManager()

	// 测试添加文件
	filePath := "/test/file.pdf"
	content := []byte("test content")
	manager.AddFile(filePath, content)

	// 测试验证文件
	entry, err := manager.ValidateFile(filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if entry.Path != filePath {
		t.Errorf("Expected path %s, got %s", filePath, entry.Path)
	}

	if entry.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), entry.Size)
	}

	// 验证调用次数
	if count := manager.GetCallCount("ValidateFile"); count != 1 {
		t.Errorf("Expected 1 call to ValidateFile, got %d", count)
	}
}

func TestMockFileManager_FileNotExists(t *testing.T) {
	manager := NewMockFileManager()

	// 测试不存在的文件
	_, err := manager.ValidateFile("/nonexistent/file.pdf")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestMockFileManager_TempFiles(t *testing.T) {
	manager := NewMockFileManager()

	// 测试创建临时文件
	tempFile, err := manager.CreateTempFile("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if tempFile == "" {
		t.Error("Expected non-empty temp file path")
	}

	// 测试创建带内容的临时文件
	content := []byte("temp content")
	tempFileWithContent, err := manager.CreateTempFileWithContent("test", content)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 验证内容
	readContent, err := manager.ReadFile(tempFileWithContent)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Expected content %s, got %s", string(content), string(readContent))
	}

	// 测试清理临时文件
	err = manager.CleanupTempFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMockProgressCallback(t *testing.T) {
	callback := NewMockProgressCallback()

	// 测试进度更新
	callback.OnProgress(25.0, "processing", "step 1")
	callback.OnProgress(50.0, "processing", "step 2")
	callback.OnProgress(100.0, "completed", "done")

	// 验证更新
	updates := callback.GetUpdates()
	if len(updates) != 3 {
		t.Errorf("Expected 3 updates, got %d", len(updates))
	}

	if updates[0] != 25.0 {
		t.Errorf("Expected first update 25.0, got %f", updates[0])
	}

	statuses := callback.GetStatuses()
	if statuses[2] != "completed" {
		t.Errorf("Expected last status 'completed', got %s", statuses[2])
	}

	// 验证调用次数
	if count := callback.GetCallCount(); count != 3 {
		t.Errorf("Expected 3 calls, got %d", count)
	}

	// 测试重置
	callback.Reset()
	if count := callback.GetCallCount(); count != 0 {
		t.Errorf("Expected 0 calls after reset, got %d", count)
	}
}

func TestMockErrorHandler(t *testing.T) {
	handler := NewMockErrorHandler()

	// 测试错误处理
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	handler.OnError(err1)
	handler.OnError(err2)

	// 验证错误
	errors := handler.GetErrors()
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	if errors[0].Error() != "error 1" {
		t.Errorf("Expected first error 'error 1', got %s", errors[0].Error())
	}

	// 验证错误数量
	if count := handler.GetErrorCount(); count != 2 {
		t.Errorf("Expected 2 errors, got %d", count)
	}

	// 验证是否有错误
	if !handler.HasError() {
		t.Error("Expected to have errors")
	}

	// 验证最后一个错误
	lastError := handler.GetLastError()
	if lastError.Error() != "error 2" {
		t.Errorf("Expected last error 'error 2', got %s", lastError.Error())
	}

	// 测试重置
	handler.Reset()
	if handler.HasError() {
		t.Error("Expected no errors after reset")
	}
}

func TestMockUIStateHandler(t *testing.T) {
	handler := NewMockUIStateHandler()

	// 测试状态变更
	handler.OnUIStateChange(true)
	handler.OnUIStateChange(false)
	handler.OnUIStateChange(true)

	// 验证状态
	states := handler.GetStates()
	if len(states) != 3 {
		t.Errorf("Expected 3 state changes, got %d", len(states))
	}

	if !states[0] {
		t.Error("Expected first state to be true")
	}

	if states[1] {
		t.Error("Expected second state to be false")
	}

	// 验证当前状态
	if !handler.GetCurrentState() {
		t.Error("Expected current state to be true")
	}

	// 验证调用次数
	if count := handler.GetCallCount(); count != 3 {
		t.Errorf("Expected 3 calls, got %d", count)
	}

	// 测试重置
	handler.Reset()
	if handler.GetCurrentState() {
		t.Error("Expected current state to be false after reset")
	}
}

func TestMockCompletionHandler(t *testing.T) {
	handler := NewMockCompletionHandler()

	// 测试完成处理
	handler.OnCompletion("Task 1 completed")
	handler.OnCompletion("Task 2 completed")

	// 验证消息
	messages := handler.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0] != "Task 1 completed" {
		t.Errorf("Expected first message 'Task 1 completed', got %s", messages[0])
	}

	// 验证完成状态
	if !handler.IsCompleted() {
		t.Error("Expected to be completed")
	}

	// 验证最后一条消息
	lastMessage := handler.GetLastMessage()
	if lastMessage != "Task 2 completed" {
		t.Errorf("Expected last message 'Task 2 completed', got %s", lastMessage)
	}

	// 验证调用次数
	if count := handler.GetCallCount(); count != 2 {
		t.Errorf("Expected 2 calls, got %d", count)
	}

	// 测试重置
	handler.Reset()
	if handler.IsCompleted() {
		t.Error("Expected not to be completed after reset")
	}
}

func TestMockReader(t *testing.T) {
	data := []byte("test data")
	reader := NewMockReader(data)

	// 测试读取
	buffer := make([]byte, 5)
	n, err := reader.Read(buffer)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}

	if string(buffer) != "test " {
		t.Errorf("Expected 'test ', got %s", string(buffer))
	}

	// 测试关闭
	err = reader.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 测试关闭后读取
	_, err = reader.Read(buffer)
	if err == nil {
		t.Error("Expected error when reading from closed reader")
	}
}

func TestMockWriter(t *testing.T) {
	writer := NewMockWriter()

	// 测试写入
	data := []byte("test data")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}

	// 验证数据
	writtenData := writer.GetData()
	if string(writtenData) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(writtenData))
	}

	// 测试关闭
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 测试关闭后写入
	_, err = writer.Write([]byte("more data"))
	if err == nil {
		t.Error("Expected error when writing to closed writer")
	}

	// 测试重置
	writer.Reset()
	resetData := writer.GetData()
	if len(resetData) != 0 {
		t.Errorf("Expected empty data after reset, got %d bytes", len(resetData))
	}
}
