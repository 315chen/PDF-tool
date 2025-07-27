package pdf

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// MockPDFService 模拟PDF服务用于测试
type MockPDFService struct {
	mergeCallCount    int
	validateCallCount int
	infoCallCount     int
	metadataCallCount int
	shouldFail        bool
	failureCount      int
}

func (m *MockPDFService) MergePDFs(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error {
	m.mergeCallCount++
	if m.shouldFail && m.mergeCallCount <= m.failureCount {
		return NewPDFError(ErrorIO, "模拟IO错误", mainFile, nil)
	}
	return nil
}

func (m *MockPDFService) ValidatePDF(filePath string) error {
	m.validateCallCount++
	if m.shouldFail && m.validateCallCount <= m.failureCount {
		return NewPDFError(ErrorIO, "模拟验证错误", filePath, nil)
	}
	return nil
}

func (m *MockPDFService) GetPDFInfo(filePath string) (*PDFInfo, error) {
	m.infoCallCount++
	if m.shouldFail && m.infoCallCount <= m.failureCount {
		return nil, NewPDFError(ErrorIO, "模拟信息获取错误", filePath, nil)
	}
	return &PDFInfo{
		PageCount:   10,
		IsEncrypted: false,
		FileSize:    1024,
		Title:       "Test PDF",
	}, nil
}

func (m *MockPDFService) GetPDFMetadata(filePath string) (map[string]string, error) {
	m.metadataCallCount++
	if m.shouldFail && m.metadataCallCount <= m.failureCount {
		return nil, NewPDFError(ErrorIO, "模拟元数据获取错误", filePath, nil)
	}
	return map[string]string{
		"Title":  "Test PDF",
		"Author": "Test Author",
	}, nil
}

func (m *MockPDFService) IsPDFEncrypted(filePath string) (bool, error) {
	if m.shouldFail {
		return false, NewPDFError(ErrorIO, "模拟加密检查错误", filePath, nil)
	}
	return false, nil
}

func TestNewServiceWithRetry(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	if service.baseService != mockService {
		t.Error("Base service not set correctly")
	}
	if service.recoveryManager == nil {
		t.Error("Recovery manager not initialized")
	}
	if service.retryManager == nil {
		t.Error("Retry manager not initialized")
	}
}

func TestServiceWithRetry_MergePDFs_Success(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	err := service.MergePDFs("main.pdf", []string{"add1.pdf", "add2.pdf"}, "output.pdf", nil)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if mockService.mergeCallCount != 1 {
		t.Errorf("Expected 1 merge call, got %d", mockService.mergeCallCount)
	}
}

func TestServiceWithRetry_MergePDFs_WithRetry(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 2, // 前两次失败，第三次成功
	}
	service := NewServiceWithRetry(mockService, 100)

	err := service.MergePDFs("main.pdf", []string{"add1.pdf"}, "output.pdf", nil)

	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}
	if mockService.mergeCallCount < 3 {
		t.Errorf("Expected at least 3 merge calls, got %d", mockService.mergeCallCount)
	}
}

func TestServiceWithRetry_MergePDFsWithContext_Success(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	ctx := context.Background()
	err := service.MergePDFsWithContext(ctx, "main.pdf", []string{"add1.pdf"}, "output.pdf", nil)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestServiceWithRetry_MergePDFsWithContext_Cancellation(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 10, // 持续失败
	}
	service := NewServiceWithRetry(mockService, 100)

	ctx, cancel := context.WithCancel(context.Background())

	// 100ms后取消
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := service.MergePDFsWithContext(ctx, "main.pdf", []string{"add1.pdf"}, "output.pdf", nil)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected error due to cancellation")
	}

	// 应该在取消时间附近结束
	if duration > 200*time.Millisecond {
		t.Errorf("Operation took too long after cancellation: %v", duration)
	}
}

func TestServiceWithRetry_ValidatePDF(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 1, // 第一次失败，第二次成功
	}
	service := NewServiceWithRetry(mockService, 100)

	err := service.ValidatePDF("test.pdf")

	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}
	if mockService.validateCallCount < 2 {
		t.Errorf("Expected at least 2 validate calls, got %d", mockService.validateCallCount)
	}
}

func TestServiceWithRetry_GetPDFInfo(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 1, // 第一次失败，第二次成功
	}
	service := NewServiceWithRetry(mockService, 100)

	info, err := service.GetPDFInfo("test.pdf")

	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}
	if info == nil {
		t.Error("Expected PDF info to be returned")
	}
	if info.PageCount != 10 {
		t.Errorf("Expected page count 10, got %d", info.PageCount)
	}
	if mockService.infoCallCount < 2 {
		t.Errorf("Expected at least 2 info calls, got %d", mockService.infoCallCount)
	}
}

func TestServiceWithRetry_BatchMergePDFs(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	jobs := []model.MergeJob{
		{
			ID:              "job1",
			MainFile:        "main1.pdf",
			AdditionalFiles: []string{"add1.pdf"},
			OutputPath:      "output1.pdf",
			Status:          model.JobPending,
		},
		{
			ID:              "job2",
			MainFile:        "main2.pdf",
			AdditionalFiles: []string{"add2.pdf"},
			OutputPath:      "output2.pdf",
			Status:          model.JobPending,
		},
	}

	errors := service.BatchMergePDFs(jobs)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	for _, job := range jobs {
		if job.Status != model.JobCompleted {
			t.Errorf("Expected job %s to be completed, got status %v", job.ID, job.Status)
		}
	}
}

func TestServiceWithRetry_BatchMergePDFs_WithErrors(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 10, // 持续失败
	}
	service := NewServiceWithRetry(mockService, 100)

	jobs := []model.MergeJob{
		{
			ID:              "job1",
			MainFile:        "main1.pdf",
			AdditionalFiles: []string{"add1.pdf"},
			OutputPath:      "output1.pdf",
			Status:          model.JobPending,
		},
	}

	errors := service.BatchMergePDFs(jobs)

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if jobs[0].Status != model.JobFailed {
		t.Errorf("Expected job to be failed, got status %v", jobs[0].Status)
	}
	if jobs[0].Error == nil {
		t.Error("Expected job to have error set")
	}
}

func TestServiceWithRetry_BatchValidatePDFs(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	files := []string{"file1.pdf", "file2.pdf", "file3.pdf"}
	results := service.BatchValidatePDFs(files)

	if len(results) != 0 {
		t.Errorf("Expected no validation errors, got %d", len(results))
	}
}

func TestServiceWithRetry_BatchValidatePDFs_WithErrors(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 10, // 持续失败
	}
	service := NewServiceWithRetry(mockService, 100)

	files := []string{"file1.pdf", "file2.pdf"}
	results := service.BatchValidatePDFs(files)

	if len(results) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(results))
	}

	for _, file := range files {
		if _, exists := results[file]; !exists {
			t.Errorf("Expected error for file %s", file)
		}
	}
}

func TestServiceWithRetry_GetServiceStats(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	stats := service.GetServiceStats()

	expectedKeys := []string{
		"service_type", "retry_enabled", "recovery_enabled",
		"error_count", "has_errors", "alloc_mb",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected key %s to exist in stats", key)
		}
	}

	if stats["service_type"] != "PDFServiceWithRetry" {
		t.Errorf("Expected service_type to be PDFServiceWithRetry, got %v", stats["service_type"])
	}
	if stats["retry_enabled"] != true {
		t.Errorf("Expected retry_enabled to be true, got %v", stats["retry_enabled"])
	}
}

func TestServiceWithRetry_ErrorManagement(t *testing.T) {
	mockService := &MockPDFService{
		shouldFail:   true,
		failureCount: 10, // 持续失败
	}
	service := NewServiceWithRetry(mockService, 100)

	// 执行一些会失败的操作
	service.MergePDFs("main.pdf", []string{"add.pdf"}, "output.pdf", nil)
	service.ValidatePDF("test.pdf")

	// 检查错误收集
	errors := service.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected errors to be collected")
	}

	summary := service.GetErrorSummary()
	if summary == "没有错误" {
		t.Error("Expected error summary to contain errors")
	}

	// 清空错误
	service.ClearErrors()
	errors = service.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected no errors after clearing, got %d", len(errors))
	}
}

func TestServiceWithRetry_RobustFileOperation(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	// 测试不存在的文件
	err := service.RobustFileOperation("/nonexistent/file.pdf", func(path string) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	pdfErr, ok := err.(*PDFError)
	if !ok || pdfErr.Type != ErrorInvalidFile {
		t.Errorf("Expected ErrorInvalidFile, got %v", err)
	}
}

func TestServiceWithRetry_MemoryAwareMerge_SmallBatch(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	files := []string{"file1.pdf", "file2.pdf", "file3.pdf"}
	err := service.MemoryAwareMerge(files, "output.pdf", 5)

	if err != nil {
		t.Errorf("Expected success for small batch, got error: %v", err)
	}
	if mockService.mergeCallCount != 1 {
		t.Errorf("Expected 1 merge call for small batch, got %d", mockService.mergeCallCount)
	}
}

func TestServiceWithRetry_MemoryAwareMerge_EmptyFiles(t *testing.T) {
	mockService := &MockPDFService{}
	service := NewServiceWithRetry(mockService, 100)

	files := []string{}
	err := service.MemoryAwareMerge(files, "output.pdf", 5)

	if err == nil {
		t.Error("Expected error for empty files")
	}

	pdfErr, ok := err.(*PDFError)
	if !ok || pdfErr.Type != ErrorInvalidFile {
		t.Errorf("Expected ErrorInvalidFile, got %v", err)
	}
}
