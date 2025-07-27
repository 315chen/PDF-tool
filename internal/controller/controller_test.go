package controller

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// mockPDFService 模拟PDF服务
type mockPDFService struct {
	validateError error
	mergeError    error
	pdfInfo       *pdf.PDFInfo
}

func (m *mockPDFService) ValidatePDF(filePath string) error {
	return m.validateError
}

func (m *mockPDFService) GetPDFInfo(filePath string) (*pdf.PDFInfo, error) {
	if m.pdfInfo != nil {
		return m.pdfInfo, nil
	}
	return &pdf.PDFInfo{
		PageCount:   10,
		IsEncrypted: false,
		FileSize:    1024,
		Title:       "Test PDF",
	}, nil
}

func (m *mockPDFService) GetPDFMetadata(filePath string) (map[string]string, error) {
	return map[string]string{"Title": "Test"}, nil
}

func (m *mockPDFService) IsPDFEncrypted(filePath string) (bool, error) {
	return false, nil
}

func (m *mockPDFService) MergePDFs(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error {
	// 模拟合并过程
	time.Sleep(100 * time.Millisecond)
	return m.mergeError
}

// mockFileManager 模拟文件管理器
type mockFileManager struct {
	validateError error
	fileInfo      *file.FileInfo
}

func (m *mockFileManager) ValidateFile(filePath string) error {
	return m.validateError
}

func (m *mockFileManager) CreateTempFile() (string, error) {
	return "/tmp/test.pdf", nil
}

func (m *mockFileManager) CreateTempFileWithPrefix(prefix string, suffix string) (string, *os.File, error) {
	return "/tmp/test.pdf", nil, nil
}

func (m *mockFileManager) CreateTempFileWithContent(prefix string, suffix string, content []byte) (string, error) {
	return "/tmp/test.pdf", nil
}

func (m *mockFileManager) CopyToTempFile(sourcePath string, prefix string) (string, error) {
	return "/tmp/test.pdf", nil
}

func (m *mockFileManager) CleanupTempFiles() error {
	return nil
}

func (m *mockFileManager) RemoveTempFile(filePath string) error {
	return nil
}

func (m *mockFileManager) GetFileInfo(filePath string) (*file.FileInfo, error) {
	if m.fileInfo != nil {
		return m.fileInfo, nil
	}
	return &file.FileInfo{
		Name:    "test.pdf",
		Size:    1024,
		Path:    filePath,
		IsValid: true,
	}, nil
}

func (m *mockFileManager) EnsureDirectoryExists(dirPath string) error {
	return nil
}

func (m *mockFileManager) GetTempDir() string {
	return "/tmp"
}

func (m *mockFileManager) SetTempFileMaxAge(duration time.Duration) {
}

func (m *mockFileManager) CopyFile(sourcePath, destPath string) error {
	return nil
}

func (m *mockFileManager) WriteFile(filePath string, data []byte) error {
	return nil
}

func (m *mockFileManager) ReadFile(filePath string) ([]byte, error) {
	return []byte("test"), nil
}

func TestController_ValidateFile(t *testing.T) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	
	// 测试成功验证
	err := controller.ValidateFile("test.pdf")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 测试文件管理器错误
	mockFile.validateError = fmt.Errorf("file not found")
	err = controller.ValidateFile("test.pdf")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestController_StartMergeJob(t *testing.T) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	
	// 设置回调
	var progressCalled int32
	var completionCalled int32

	controller.SetProgressCallback(func(progress float64, status, detail string) {
		atomic.StoreInt32(&progressCalled, 1)
	})

	controller.SetCompletionCallback(func(outputPath string) {
		atomic.StoreInt32(&completionCalled, 1)
	})
	
	// 测试启动合并任务
	err := controller.StartMergeJob("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 等待任务完成
	time.Sleep(300 * time.Millisecond)
	
	// 验证回调被调用
	if atomic.LoadInt32(&progressCalled) == 0 {
		t.Error("Progress callback was not called")
	}

	if atomic.LoadInt32(&completionCalled) == 0 {
		t.Error("Completion callback was not called")
	}
	
	// 验证任务状态
	job := controller.GetCurrentJob()
	if job != nil {
		t.Error("Expected current job to be nil after completion")
	}
}

func TestController_CancelCurrentJob(t *testing.T) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	
	// 启动任务
	err := controller.StartMergeJob("main.pdf", []string{"add1.pdf"}, "output.pdf")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 立即取消
	err = controller.CancelCurrentJob()
	if err != nil {
		t.Errorf("Expected no error when canceling, got %v", err)
	}
	
	// 验证任务被取消
	job := controller.GetCurrentJob()
	if job != nil {
		t.Error("Expected current job to be nil after cancellation")
	}
}

func TestController_IsJobRunning(t *testing.T) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()
	
	controller := NewController(mockPDF, mockFile, config)
	
	// 初始状态应该没有任务运行
	if controller.IsJobRunning() {
		t.Error("Expected no job running initially")
	}
	
	// 启动任务
	err := controller.StartMergeJob("main.pdf", []string{"add1.pdf"}, "output.pdf")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 等待一小段时间让任务开始运行
	time.Sleep(50 * time.Millisecond)
	
	// 应该有任务在运行
	if !controller.IsJobRunning() {
		t.Error("Expected job to be running")
	}
	
	// 等待任务完成
	time.Sleep(300 * time.Millisecond)
	
	// 任务完成后应该没有任务运行
	if controller.IsJobRunning() {
		t.Error("Expected no job running after completion")
	}
}