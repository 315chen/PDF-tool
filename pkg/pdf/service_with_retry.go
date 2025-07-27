package pdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	
	"github.com/user/pdf-merger/internal/model"
)

// ServiceWithRetry PDF服务的重试版本，集成了错误恢复机制
type ServiceWithRetry struct {
	baseService     PDFService
	recoveryManager *RecoveryManager
	retryManager    *RetryManager
}

// NewServiceWithRetry 创建带重试功能的PDF服务
func NewServiceWithRetry(baseService PDFService, maxMemoryMB int64) *ServiceWithRetry {
	recoveryManager := NewRecoveryManager(maxMemoryMB)
	retryManager := NewRetryManager(DefaultRetryConfig(), NewDefaultErrorHandler(3))
	
	return &ServiceWithRetry{
		baseService:     baseService,
		recoveryManager: recoveryManager,
		retryManager:    retryManager,
	}
}

// MergePDFs 合并PDF文件，带重试和恢复机制
func (s *ServiceWithRetry) MergePDFs(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error {
	operation := func() error {
		return s.baseService.MergePDFs(mainFile, additionalFiles, outputPath, progressWriter)
	}
	
	return s.recoveryManager.ExecuteWithRecovery(operation)
}

// MergePDFsWithContext 带上下文的PDF合并，支持取消和超时
func (s *ServiceWithRetry) MergePDFsWithContext(ctx context.Context, mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error {
	operation := func() error {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return NewPDFError(ErrorIO, "操作被取消", "", ctx.Err())
		default:
		}
		
		return s.baseService.MergePDFs(mainFile, additionalFiles, outputPath, progressWriter)
	}
	
	return s.retryManager.ExecuteWithContext(ctx, operation)
}

// ValidatePDF 验证PDF文件，带重试机制
func (s *ServiceWithRetry) ValidatePDF(filePath string) error {
	operation := func() error {
		return s.baseService.ValidatePDF(filePath)
	}
	
	return s.retryManager.Execute(operation)
}

// GetPDFInfo 获取PDF信息，带重试机制
func (s *ServiceWithRetry) GetPDFInfo(filePath string) (*PDFInfo, error) {
	var result *PDFInfo
	var err error
	
	operation := func() error {
		result, err = s.baseService.GetPDFInfo(filePath)
		return err
	}
	
	retryErr := s.retryManager.Execute(operation)
	if retryErr != nil {
		return nil, retryErr
	}
	
	return result, nil
}

// BatchMergePDFs 批量合并PDF文件，带错误收集和恢复
func (s *ServiceWithRetry) BatchMergePDFs(jobs []model.MergeJob) []error {
	var errors []error
	
	for i := range jobs {
		job := &jobs[i]
		job.Status = model.JobRunning
		
		err := s.MergePDFs(job.MainFile, job.AdditionalFiles, job.OutputPath, nil)
		if err != nil {
			job.Status = model.JobFailed
			job.Error = err
			errors = append(errors, err)
		} else {
			job.Status = model.JobCompleted
		}
	}
	
	return errors
}

// BatchValidatePDFs 批量验证PDF文件
func (s *ServiceWithRetry) BatchValidatePDFs(filePaths []string) map[string]error {
	results := make(map[string]error)
	
	for _, filePath := range filePaths {
		err := s.ValidatePDF(filePath)
		if err != nil {
			results[filePath] = err
		}
	}
	
	return results
}

// GetServiceStats 获取服务统计信息
func (s *ServiceWithRetry) GetServiceStats() map[string]interface{} {
	stats := s.recoveryManager.GetRecoveryStats()
	
	// 添加服务特定的统计信息
	stats["service_type"] = "PDFServiceWithRetry"
	stats["retry_enabled"] = true
	stats["recovery_enabled"] = true
	
	return stats
}

// ClearErrors 清空错误记录
func (s *ServiceWithRetry) ClearErrors() {
	s.recoveryManager.ClearErrors()
}

// GetErrors 获取收集的错误
func (s *ServiceWithRetry) GetErrors() []error {
	return s.recoveryManager.GetErrors()
}

// GetErrorSummary 获取错误摘要
func (s *ServiceWithRetry) GetErrorSummary() string {
	return s.recoveryManager.GetErrorSummary()
}

// RobustFileOperation 健壮的文件操作，带重试和恢复
func (s *ServiceWithRetry) RobustFileOperation(filePath string, operation func(string) error) error {
	robustOp := func() error {
		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return NewPDFError(ErrorInvalidFile, "文件不存在", filePath, err)
		}
		
		// 检查文件权限
		if file, err := os.Open(filePath); err != nil {
			return NewPDFError(ErrorPermission, "无法打开文件", filePath, err)
		} else {
			file.Close()
		}
		
		// 执行实际操作
		return operation(filePath)
	}
	
	return s.recoveryManager.ExecuteWithRecovery(robustOp)
}

// SafeOutputOperation 安全的输出操作，确保输出目录存在
func (s *ServiceWithRetry) SafeOutputOperation(outputPath string, operation func(string) error) error {
	safeOp := func() error {
		// 确保输出目录存在
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return NewPDFError(ErrorPermission, "无法创建输出目录", outputDir, err)
		}
		
		// 检查输出目录权限
		if file, err := os.OpenFile(filepath.Join(outputDir, ".test"), os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return NewPDFError(ErrorPermission, "输出目录无写权限", outputDir, err)
		} else {
			file.Close()
			os.Remove(filepath.Join(outputDir, ".test"))
		}
		
		// 执行实际操作
		return operation(outputPath)
	}
	
	return s.recoveryManager.ExecuteWithRecovery(safeOp)
}

// MemoryAwareMerge 内存感知的合并操作
func (s *ServiceWithRetry) MemoryAwareMerge(files []string, outputPath string, maxFilesPerBatch int) error {
	if len(files) == 0 {
		return NewPDFError(ErrorInvalidFile, "没有文件需要合并", "", nil)
	}
	
	// 如果文件数量较少，直接合并
	if len(files) <= maxFilesPerBatch {
		return s.MergePDFs(files[0], files[1:], outputPath, nil)
	}
	
	// 分批处理大量文件
	tempFiles := make([]string, 0)
	
	for i := 0; i < len(files); i += maxFilesPerBatch {
		end := i + maxFilesPerBatch
		if end > len(files) {
			end = len(files)
		}
		
		batch := files[i:end]
		tempOutput := fmt.Sprintf("%s.batch_%d.pdf", outputPath, i/maxFilesPerBatch)
		
		err := s.MergePDFs(batch[0], batch[1:], tempOutput, nil)
		if err != nil {
			// 清理已创建的临时文件
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return err
		}
		
		tempFiles = append(tempFiles, tempOutput)
	}
	
	// 合并所有批次文件
	err := s.MergePDFs(tempFiles[0], tempFiles[1:], outputPath, nil)
	
	// 清理临时文件
	for _, tempFile := range tempFiles {
		os.Remove(tempFile)
	}
	
	return err
}