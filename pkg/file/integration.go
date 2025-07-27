package file

import (
	"github.com/user/pdf-merger/pkg/pdf"
)

// FileValidator 集成文件验证功能
type FileValidator struct {
	fileManager   FileManager
	pdfValidator  *pdf.PDFValidator
}

// NewFileValidator 创建一个新的文件验证器
func NewFileValidator(tempDir string) *FileValidator {
	return &FileValidator{
		fileManager:  NewFileManager(tempDir),
		pdfValidator: pdf.NewPDFValidator(),
	}
}

// ValidateAndGetInfo 验证文件并获取详细信息
func (fv *FileValidator) ValidateAndGetInfo(filePath string) (*FileValidationResult, error) {
	result := &FileValidationResult{
		FilePath: filePath,
		IsValid:  false,
	}

	// 基本文件验证
	if err := fv.fileManager.ValidateFile(filePath); err != nil {
		result.Error = err
		return result, err
	}

	// 获取文件基本信息
	fileInfo, err := fv.fileManager.GetFileInfo(filePath)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.FileInfo = fileInfo

	// PDF格式验证
	if err := fv.pdfValidator.ValidatePDFFile(filePath); err != nil {
		result.Error = err
		return result, err
	}

	// 获取PDF详细信息
	pdfInfo, err := fv.pdfValidator.GetBasicPDFInfo(filePath)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.PDFInfo = pdfInfo
	result.IsValid = true

	return result, nil
}

// FileValidationResult 文件验证结果
type FileValidationResult struct {
	FilePath string
	IsValid  bool
	Error    error
	FileInfo *FileInfo
	PDFInfo  *pdf.PDFInfo
}