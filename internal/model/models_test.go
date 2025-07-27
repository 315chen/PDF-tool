package model

import (
	"fmt"
	"testing"
	"time"
)

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status   JobStatus
		expected string
	}{
		{JobPending, "等待中"},
		{JobRunning, "执行中"},
		{JobCompleted, "已完成"},
		{JobFailed, "失败"},
		{JobStatus(999), "未知状态"},
	}

	for _, test := range tests {
		if result := test.status.String(); result != test.expected {
			t.Errorf("JobStatus(%d).String() = %s, expected %s", test.status, result, test.expected)
		}
	}
}

func TestNewMergeJob(t *testing.T) {
	mainFile := "/path/to/main.pdf"
	additionalFiles := []string{"/path/to/file1.pdf", "/path/to/file2.pdf"}
	outputPath := "/path/to/output.pdf"

	job := NewMergeJob(mainFile, additionalFiles, outputPath)

	if job.MainFile != mainFile {
		t.Errorf("Expected MainFile %s, got %s", mainFile, job.MainFile)
	}

	if len(job.AdditionalFiles) != len(additionalFiles) {
		t.Errorf("Expected %d additional files, got %d", len(additionalFiles), len(job.AdditionalFiles))
	}

	if job.OutputPath != outputPath {
		t.Errorf("Expected OutputPath %s, got %s", outputPath, job.OutputPath)
	}

	if job.Status != JobPending {
		t.Errorf("Expected status JobPending, got %v", job.Status)
	}

	if job.Progress != 0.0 {
		t.Errorf("Expected progress 0.0, got %f", job.Progress)
	}

	if job.ID == "" {
		t.Error("Expected non-empty job ID")
	}
}

func TestMergeJob_SetCompleted(t *testing.T) {
	job := NewMergeJob("main.pdf", []string{"file1.pdf"}, "output.pdf")
	
	job.SetCompleted()

	if job.Status != JobCompleted {
		t.Errorf("Expected status JobCompleted, got %v", job.Status)
	}

	if job.Progress != 100.0 {
		t.Errorf("Expected progress 100.0, got %f", job.Progress)
	}

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestMergeJob_SetFailed(t *testing.T) {
	job := NewMergeJob("main.pdf", []string{"file1.pdf"}, "output.pdf")
	testError := fmt.Errorf("Test error")
	
	job.SetFailed(testError)

	if job.Status != JobFailed {
		t.Errorf("Expected status JobFailed, got %v", job.Status)
	}

	if job.Error != testError {
		t.Errorf("Expected error %v, got %v", testError, job.Error)
	}

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestMergeJob_UpdateProgress(t *testing.T) {
	job := NewMergeJob("main.pdf", []string{"file1.pdf"}, "output.pdf")

	// 测试正常进度
	job.UpdateProgress(50.0)
	if job.Progress != 50.0 {
		t.Errorf("Expected progress 50.0, got %f", job.Progress)
	}

	// 测试负数进度
	job.UpdateProgress(-10.0)
	if job.Progress != 0.0 {
		t.Errorf("Expected progress 0.0 for negative input, got %f", job.Progress)
	}

	// 测试超过100的进度
	job.UpdateProgress(150.0)
	if job.Progress != 100.0 {
		t.Errorf("Expected progress 100.0 for >100 input, got %f", job.Progress)
	}
}

func TestMergeJob_GetTotalFiles(t *testing.T) {
	job := NewMergeJob("main.pdf", []string{"file1.pdf", "file2.pdf", "file3.pdf"}, "output.pdf")
	
	expected := 4 // 1 main + 3 additional
	if total := job.GetTotalFiles(); total != expected {
		t.Errorf("Expected total files %d, got %d", expected, total)
	}
}

func TestNewFileEntry(t *testing.T) {
	path := "/path/to/test.pdf"
	order := 5

	entry := NewFileEntry(path, order)

	if entry.Path != path {
		t.Errorf("Expected Path %s, got %s", path, entry.Path)
	}

	if entry.Order != order {
		t.Errorf("Expected Order %d, got %d", order, entry.Order)
	}

	if entry.DisplayName != "test.pdf" {
		t.Errorf("Expected DisplayName 'test.pdf', got %s", entry.DisplayName)
	}

	if !entry.IsValid {
		t.Error("Expected IsValid to be true")
	}
}

func TestFileEntry_SetError(t *testing.T) {
	entry := NewFileEntry("/path/to/test.pdf", 1)
	errorMsg := "Test error message"

	entry.SetError(errorMsg)

	if entry.Error != errorMsg {
		t.Errorf("Expected Error %s, got %s", errorMsg, entry.Error)
	}

	if entry.IsValid {
		t.Error("Expected IsValid to be false after setting error")
	}
}

func TestFileEntry_GetSizeString(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{500, "500 B"},
		{1536, "1.5 KB"},
		{2097152, "2.0 MB"},
		{1048576, "1.0 MB"},
	}

	for _, test := range tests {
		entry := &FileEntry{Size: test.size}
		if result := entry.GetSizeString(); result != test.expected {
			t.Errorf("GetSizeString() for size %d = %s, expected %s", test.size, result, test.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxMemoryUsage != 100*1024*1024 {
		t.Errorf("Expected MaxMemoryUsage %d, got %d", 100*1024*1024, config.MaxMemoryUsage)
	}

	if !config.EnableAutoDecrypt {
		t.Error("Expected EnableAutoDecrypt to be true")
	}

	if config.WindowWidth != 800 {
		t.Errorf("Expected WindowWidth 800, got %d", config.WindowWidth)
	}

	if config.WindowHeight != 600 {
		t.Errorf("Expected WindowHeight 600, got %d", config.WindowHeight)
	}

	if len(config.CommonPasswords) == 0 {
		t.Error("Expected CommonPasswords to be non-empty")
	}
}

func TestGenerateJobID(t *testing.T) {
	id1 := generateJobID()
	time.Sleep(1 * time.Millisecond) // 确保时间戳不同
	id2 := generateJobID()

	if id1 == id2 {
		t.Error("Expected different job IDs")
	}

	if id1 == "" || id2 == "" {
		t.Error("Expected non-empty job IDs")
	}
}