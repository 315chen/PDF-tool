package pdf

import (
	"fmt"
	"os"
	"strings"
)

// PDFValidator 提供PDF文件验证功能
type PDFValidator struct{}

// NewPDFValidator 创建一个新的PDF验证器
func NewPDFValidator() *PDFValidator {
	return &PDFValidator{}
}

// ValidatePDFFile 验证PDF文件格式
func (v *PDFValidator) ValidatePDFFile(filePath string) error {
	// 尝试使用pdfcpu进行验证
	if err := v.validateWithPDFCPU(filePath); err == nil {
		// pdfcpu验证成功
		return nil
	}

	// pdfcpu验证失败或不可用，回退到基本验证
	return v.validateBasic(filePath)
}

// validateWithPDFCPU 使用pdfcpu进行验证
func (v *PDFValidator) validateWithPDFCPU(filePath string) error {
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return err // pdfcpu不可用
	}
	defer adapter.Close()

	return adapter.ValidateFile(filePath)
}

// validateBasic 基本验证方法（回退）
func (v *PDFValidator) validateBasic(filePath string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法打开文件",
			File:    filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 读取文件头部
	header := make([]byte, 8)
	n, err := file.Read(header)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法读取文件头部",
			File:    filePath,
			Cause:   err,
		}
	}

	if n < 4 {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件太小，不是有效的PDF文件",
			File:    filePath,
		}
	}

	// 检查PDF文件签名
	headerStr := string(header[:4])
	if headerStr != "%PDF" {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件不是有效的PDF格式",
			File:    filePath,
		}
	}

	// 检查PDF版本
	if n >= 8 {
		versionStr := string(header[4:8])
		if !v.isValidPDFVersion(versionStr) {
			return &PDFError{
				Type:    ErrorInvalidFile,
				Message: fmt.Sprintf("不支持的PDF版本: %s", versionStr),
				File:    filePath,
			}
		}
	}

	// 检查文件是否完整（查找EOF标记）
	if err := v.checkPDFIntegrity(file); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件可能已损坏",
			File:    filePath,
			Cause:   err,
		}
	}

	return nil
}

// isValidPDFVersion 检查PDF版本是否有效
func (v *PDFValidator) isValidPDFVersion(version string) bool {
	validVersions := []string{"-1.0", "-1.1", "-1.2", "-1.3", "-1.4", "-1.5", "-1.6", "-1.7", "-2.0"}
	for _, validVersion := range validVersions {
		if strings.HasPrefix(version, validVersion) {
			return true
		}
	}
	return false
}

// checkPDFIntegrity 检查PDF文件完整性
func (v *PDFValidator) checkPDFIntegrity(file *os.File) error {
	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := stat.Size()
	if fileSize < 100 { // PDF文件至少应该有100字节
		return fmt.Errorf("文件太小")
	}

	// 检查文件末尾是否有EOF标记
	// 读取文件末尾的1024字节
	bufferSize := int64(1024)
	if fileSize < bufferSize {
		bufferSize = fileSize
	}

	buffer := make([]byte, bufferSize)
	_, err = file.ReadAt(buffer, fileSize-bufferSize)
	if err != nil {
		return err
	}

	// 查找%%EOF标记
	content := string(buffer)
	if !strings.Contains(content, "%%EOF") {
		return fmt.Errorf("缺少PDF结束标记")
	}

	return nil
}

// GetBasicPDFInfo 获取PDF文件的基本信息
func (v *PDFValidator) GetBasicPDFInfo(filePath string) (*PDFInfo, error) {
	// 尝试使用pdfcpu获取详细信息
	adapter, err := NewPDFCPUAdapter(nil)
	if err == nil {
		defer adapter.Close()
		if info, err := adapter.GetFileInfo(filePath); err == nil {
			// pdfcpu获取信息成功
			return info, nil
		}
		// pdfcpu获取信息失败，继续使用基本方法
	}

	// 首先验证文件
	if err := v.ValidatePDFFile(filePath); err != nil {
		return nil, err
	}

	// 获取文件大小
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    filePath,
			Cause:   err,
		}
	}

	// 检查是否加密
	isEncrypted, err := v.isPDFEncrypted(filePath)
	if err != nil {
		// 如果无法确定加密状态，假设未加密
		isEncrypted = false
	}

	return &PDFInfo{
		PageCount:   -1, // 需要PDF库才能准确获取页数
		IsEncrypted: isEncrypted,
		FileSize:    stat.Size(),
		Title:       "", // 需要PDF库才能获取标题
	}, nil
}

// isPDFEncrypted 检查PDF是否加密（简单检查）
func (v *PDFValidator) isPDFEncrypted(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 读取文件内容的一部分来查找加密标记
	buffer := make([]byte, 4096)
	_, err = file.Read(buffer)
	if err != nil {
		return false, err
	}

	content := string(buffer)
	// 查找常见的加密相关关键字
	encryptionKeywords := []string{"/Encrypt", "/Filter", "/V", "/R"}

	for _, keyword := range encryptionKeywords {
		if strings.Contains(content, keyword) {
			return true, nil
		}
	}

	return false, nil
}

// ValidateWithStrictMode 使用严格模式验证PDF文件
func (v *PDFValidator) ValidateWithStrictMode(filePath string) error {
	// 尝试使用pdfcpu的严格验证模式
	adapter, err := NewPDFCPUAdapter(&PDFCPUConfig{
		ValidationMode: "strict",
	})
	if err == nil {
		defer adapter.Close()
		if err := adapter.ValidateFile(filePath); err == nil {
			return nil
		}
		// 严格验证失败，返回错误
		return &PDFError{
			Type:    ErrorValidation,
			Message: "PDF文件未通过严格验证",
			File:    filePath,
			Cause:   err,
		}
	}

	// pdfcpu不可用，回退到基本验证
	return v.validateBasic(filePath)
}

// GetValidationReport 获取详细的验证报告
func (v *PDFValidator) GetValidationReport(filePath string) (*ValidationReport, error) {
	report := &ValidationReport{
		FilePath: filePath,
		IsValid:  false,
		Errors:   []string{},
		Warnings: []string{},
		Details:  make(map[string]interface{}),
	}

	// 基本文件检查
	if err := v.validateBasic(filePath); err != nil {
		report.Errors = append(report.Errors, err.Error())
		return report, nil
	}

	// 尝试使用pdfcpu获取详细信息
	adapter, err := NewPDFCPUAdapter(nil)
	if err == nil {
		defer adapter.Close()

		// 验证文件
		if err := adapter.ValidateFile(filePath); err != nil {
			report.Errors = append(report.Errors, "pdfcpu验证失败: "+err.Error())
		} else {
			report.IsValid = true
		}

		// 获取文件信息
		if info, err := adapter.GetFileInfo(filePath); err == nil {
			report.Details["pageCount"] = info.PageCount
			report.Details["fileSize"] = info.FileSize
			report.Details["isEncrypted"] = info.IsEncrypted
			report.Details["title"] = info.Title
		}
	} else {
		report.Warnings = append(report.Warnings, "pdfcpu不可用，使用基本验证")
		report.IsValid = true // 基本验证已通过
	}

	return report, nil
}

// CheckPermissions 检查PDF文件权限
func (v *PDFValidator) CheckPermissions(filePath string) (*PDFPermissions, error) {
	// 尝试使用pdfcpu获取权限信息
	adapter, err := NewPDFCPUAdapter(nil)
	if err == nil {
		defer adapter.Close()

		// 获取文件信息
		if info, err := adapter.GetFileInfo(filePath); err == nil {
			permissions := &PDFPermissions{
				CanPrint:    true, // 默认权限
				CanCopy:     true,
				CanModify:   true,
				CanAnnotate: true,
				IsEncrypted: info.IsEncrypted,
			}

			// 如果文件加密，权限可能受限
			if info.IsEncrypted {
				permissions.CanPrint = false
				permissions.CanCopy = false
				permissions.CanModify = false
				permissions.CanAnnotate = false
			}

			return permissions, nil
		}
	}

	// 回退到基本权限检查
	isEncrypted, _ := v.isPDFEncrypted(filePath)
	return &PDFPermissions{
		CanPrint:    !isEncrypted,
		CanCopy:     !isEncrypted,
		CanModify:   !isEncrypted,
		CanAnnotate: !isEncrypted,
		IsEncrypted: isEncrypted,
	}, nil
}

// ValidationReport 验证报告结构
type ValidationReport struct {
	FilePath string                 `json:"filePath"`
	IsValid  bool                   `json:"isValid"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
	Details  map[string]interface{} `json:"details"`
}

// PDFPermissions PDF权限结构
type PDFPermissions struct {
	CanPrint    bool `json:"canPrint"`
	CanCopy     bool `json:"canCopy"`
	CanModify   bool `json:"canModify"`
	CanAnnotate bool `json:"canAnnotate"`
	IsEncrypted bool `json:"isEncrypted"`
}
