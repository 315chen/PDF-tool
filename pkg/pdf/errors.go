package pdf

import (
	"fmt"
	"strings"
)

// ErrorType 定义PDF处理中可能出现的错误类型
type ErrorType int

const (
	// ErrorInvalidFile 表示文件格式无效或已损坏
	ErrorInvalidFile ErrorType = iota
	// ErrorEncrypted 表示文件已加密
	ErrorEncrypted
	// ErrorCorrupted 表示文件已损坏
	ErrorCorrupted
	// ErrorPermission 表示没有访问文件的权限
	ErrorPermission
	// ErrorMemory 表示内存不足
	ErrorMemory
	// ErrorIO 表示文件读写错误
	ErrorIO
	// ErrorValidation 表示PDF验证失败
	ErrorValidation
	// ErrorProcessing 表示PDF处理失败
	ErrorProcessing
	// ErrorInvalidInput 表示输入参数无效
	ErrorInvalidInput
)

// PDFError 定义PDF处理错误的结构
type PDFError struct {
	Type    ErrorType
	Message string
	File    string
	Cause   error
}

// Error 实现error接口
func (e *PDFError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (file: %s): %v", e.typeString(), e.Message, e.File, e.Cause)
	}
	return fmt.Sprintf("%s: %s (file: %s)", e.typeString(), e.Message, e.File)
}

// typeString 返回错误类型的字符串表示
func (e *PDFError) typeString() string {
	switch e.Type {
	case ErrorInvalidFile:
		return "Invalid File"
	case ErrorEncrypted:
		return "Encrypted File"
	case ErrorCorrupted:
		return "Corrupted File"
	case ErrorPermission:
		return "Permission Error"
	case ErrorMemory:
		return "Memory Error"
	case ErrorIO:
		return "IO Error"
	case ErrorValidation:
		return "Validation Error"
	case ErrorProcessing:
		return "Processing Error"
	case ErrorInvalidInput:
		return "Invalid Input"
	default:
		return "Unknown Error"
	}
}

// ErrorMessages 定义用户友好的错误消息
var ErrorMessages = map[ErrorType]string{
	ErrorInvalidFile:  "文件格式无效或已损坏",
	ErrorEncrypted:    "文件已加密，需要密码",
	ErrorCorrupted:    "文件已损坏，无法处理",
	ErrorPermission:   "没有访问文件的权限",
	ErrorMemory:       "内存不足，请关闭其他程序后重试",
	ErrorIO:           "文件读写错误，请检查磁盘空间",
	ErrorValidation:   "PDF文件验证失败",
	ErrorProcessing:   "PDF文件处理失败",
	ErrorInvalidInput: "输入参数无效",
}

// NewPDFError 创建一个新的PDFError
func NewPDFError(errorType ErrorType, message, file string, cause error) *PDFError {
	return &PDFError{
		Type:    errorType,
		Message: message,
		File:    file,
		Cause:   cause,
	}
}

// GetUserMessage 获取用户友好的错误消息
func (e *PDFError) GetUserMessage() string {
	if msg, exists := ErrorMessages[e.Type]; exists {
		return msg
	}
	return "未知错误"
}

// GetDetailedMessage 获取详细的错误消息，包含文件信息
func (e *PDFError) GetDetailedMessage() string {
	userMsg := e.GetUserMessage()
	if e.File != "" {
		return fmt.Sprintf("%s (文件: %s)", userMsg, e.File)
	}
	return userMsg
}

// IsRetryable 判断错误是否可以重试
func (e *PDFError) IsRetryable() bool {
	switch e.Type {
	case ErrorIO, ErrorMemory:
		return true
	case ErrorInvalidFile, ErrorCorrupted, ErrorPermission:
		return false
	case ErrorEncrypted:
		return false // 加密错误需要特殊处理，不是简单重试
	default:
		return false
	}
}

// GetSeverity 获取错误严重程度
func (e *PDFError) GetSeverity() string {
	switch e.Type {
	case ErrorMemory, ErrorIO:
		return "high"
	case ErrorPermission, ErrorCorrupted:
		return "medium"
	case ErrorInvalidFile, ErrorEncrypted:
		return "low"
	default:
		return "unknown"
	}
}

// Unwrap 实现errors.Unwrap接口，用于错误链
func (e *PDFError) Unwrap() error {
	return e.Cause
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err error) error
	ShouldRetry(err error) bool
	GetUserFriendlyMessage(err error) string
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	maxRetries int
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(maxRetries int) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		maxRetries: maxRetries,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err error) error {
	if err == nil {
		return nil
	}

	if pdfErr, ok := err.(*PDFError); ok {
		return pdfErr
	}

	// 将普通错误转换为PDFError
	return NewPDFError(ErrorIO, err.Error(), "", err)
}

// ShouldRetry 判断是否应该重试
func (h *DefaultErrorHandler) ShouldRetry(err error) bool {
	if pdfErr, ok := err.(*PDFError); ok {
		return pdfErr.IsRetryable()
	}
	return false
}

// GetUserFriendlyMessage 获取用户友好的错误消息
func (h *DefaultErrorHandler) GetUserFriendlyMessage(err error) string {
	if pdfErr, ok := err.(*PDFError); ok {
		return pdfErr.GetDetailedMessage()
	}
	return "处理过程中发生未知错误"
}

// ErrorCollector 错误收集器，用于收集批量处理中的错误
type ErrorCollector struct {
	errors []error
}

// NewErrorCollector 创建新的错误收集器
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Add 添加错误到收集器
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// HasErrors 检查是否有错误
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// GetErrors 获取所有错误
func (ec *ErrorCollector) GetErrors() []error {
	return ec.errors
}

// GetErrorCount 获取错误数量
func (ec *ErrorCollector) GetErrorCount() int {
	return len(ec.errors)
}

// GetSummary 获取错误摘要
func (ec *ErrorCollector) GetSummary() string {
	if !ec.HasErrors() {
		return "没有错误"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("共发现 %d 个错误:\n", len(ec.errors)))

	for i, err := range ec.errors {
		summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Error()))
	}

	return summary.String()
}

// Clear 清空错误收集器
func (ec *ErrorCollector) Clear() {
	ec.errors = ec.errors[:0]
}
