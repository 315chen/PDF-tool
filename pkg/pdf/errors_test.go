package pdf

import (
	"errors"
	"strings"
	"testing"
)

func TestPDFError_Error(t *testing.T) {
	tests := []struct {
		name     string
		pdfError *PDFError
		want     string
	}{
		{
			name: "error with cause",
			pdfError: &PDFError{
				Type:    ErrorInvalidFile,
				Message: "test message",
				File:    "test.pdf",
				Cause:   errors.New("underlying error"),
			},
			want: "Invalid File: test message (file: test.pdf): underlying error",
		},
		{
			name: "error without cause",
			pdfError: &PDFError{
				Type:    ErrorEncrypted,
				Message: "encrypted file",
				File:    "encrypted.pdf",
				Cause:   nil,
			},
			want: "Encrypted File: encrypted file (file: encrypted.pdf)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pdfError.Error(); got != tt.want {
				t.Errorf("PDFError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPDFError_GetUserMessage(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		want      string
	}{
		{
			name:      "invalid file error",
			errorType: ErrorInvalidFile,
			want:      "文件格式无效或已损坏",
		},
		{
			name:      "encrypted error",
			errorType: ErrorEncrypted,
			want:      "文件已加密，需要密码",
		},
		{
			name:      "unknown error type",
			errorType: ErrorType(999),
			want:      "未知错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfErr := &PDFError{Type: tt.errorType}
			if got := pdfErr.GetUserMessage(); got != tt.want {
				t.Errorf("PDFError.GetUserMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPDFError_GetDetailedMessage(t *testing.T) {
	tests := []struct {
		name     string
		pdfError *PDFError
		want     string
	}{
		{
			name: "with file",
			pdfError: &PDFError{
				Type: ErrorInvalidFile,
				File: "test.pdf",
			},
			want: "文件格式无效或已损坏 (文件: test.pdf)",
		},
		{
			name: "without file",
			pdfError: &PDFError{
				Type: ErrorMemory,
				File: "",
			},
			want: "内存不足，请关闭其他程序后重试",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pdfError.GetDetailedMessage(); got != tt.want {
				t.Errorf("PDFError.GetDetailedMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPDFError_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		want      bool
	}{
		{
			name:      "IO error is retryable",
			errorType: ErrorIO,
			want:      true,
		},
		{
			name:      "Memory error is retryable",
			errorType: ErrorMemory,
			want:      true,
		},
		{
			name:      "Invalid file error is not retryable",
			errorType: ErrorInvalidFile,
			want:      false,
		},
		{
			name:      "Encrypted error is not retryable",
			errorType: ErrorEncrypted,
			want:      false,
		},
		{
			name:      "Permission error is not retryable",
			errorType: ErrorPermission,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfErr := &PDFError{Type: tt.errorType}
			if got := pdfErr.IsRetryable(); got != tt.want {
				t.Errorf("PDFError.IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPDFError_GetSeverity(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		want      string
	}{
		{
			name:      "memory error has high severity",
			errorType: ErrorMemory,
			want:      "high",
		},
		{
			name:      "IO error has high severity",
			errorType: ErrorIO,
			want:      "high",
		},
		{
			name:      "permission error has medium severity",
			errorType: ErrorPermission,
			want:      "medium",
		},
		{
			name:      "invalid file error has low severity",
			errorType: ErrorInvalidFile,
			want:      "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfErr := &PDFError{Type: tt.errorType}
			if got := pdfErr.GetSeverity(); got != tt.want {
				t.Errorf("PDFError.GetSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPDFError(t *testing.T) {
	cause := errors.New("test cause")
	pdfErr := NewPDFError(ErrorInvalidFile, "test message", "test.pdf", cause)

	if pdfErr.Type != ErrorInvalidFile {
		t.Errorf("Expected Type to be ErrorInvalidFile, got %v", pdfErr.Type)
	}
	if pdfErr.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got %v", pdfErr.Message)
	}
	if pdfErr.File != "test.pdf" {
		t.Errorf("Expected File to be 'test.pdf', got %v", pdfErr.File)
	}
	if pdfErr.Cause != cause {
		t.Errorf("Expected Cause to be the provided error, got %v", pdfErr.Cause)
	}
}

func TestDefaultErrorHandler_HandleError(t *testing.T) {
	handler := NewDefaultErrorHandler(3)

	t.Run("handle PDFError", func(t *testing.T) {
		originalErr := NewPDFError(ErrorInvalidFile, "test", "test.pdf", nil)
		handledErr := handler.HandleError(originalErr)

		if handledErr != originalErr {
			t.Errorf("Expected PDFError to be returned as-is")
		}
	})

	t.Run("handle regular error", func(t *testing.T) {
		originalErr := errors.New("regular error")
		handledErr := handler.HandleError(originalErr)

		pdfErr, ok := handledErr.(*PDFError)
		if !ok {
			t.Errorf("Expected regular error to be converted to PDFError")
		}
		if pdfErr.Type != ErrorIO {
			t.Errorf("Expected error type to be ErrorIO, got %v", pdfErr.Type)
		}
		if pdfErr.Cause != originalErr {
			t.Errorf("Expected cause to be the original error")
		}
	})
}

func TestDefaultErrorHandler_ShouldRetry(t *testing.T) {
	handler := NewDefaultErrorHandler(3)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "retryable PDFError",
			err:  NewPDFError(ErrorIO, "test", "test.pdf", nil),
			want: true,
		},
		{
			name: "non-retryable PDFError",
			err:  NewPDFError(ErrorInvalidFile, "test", "test.pdf", nil),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.ShouldRetry(tt.err); got != tt.want {
				t.Errorf("DefaultErrorHandler.ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultErrorHandler_GetUserFriendlyMessage(t *testing.T) {
	handler := NewDefaultErrorHandler(3)

	t.Run("PDFError", func(t *testing.T) {
		pdfErr := NewPDFError(ErrorInvalidFile, "test", "test.pdf", nil)
		msg := handler.GetUserFriendlyMessage(pdfErr)
		expected := "文件格式无效或已损坏 (文件: test.pdf)"
		if msg != expected {
			t.Errorf("Expected message '%s', got '%s'", expected, msg)
		}
	})

	t.Run("regular error", func(t *testing.T) {
		err := errors.New("regular error")
		msg := handler.GetUserFriendlyMessage(err)
		expected := "处理过程中发生未知错误"
		if msg != expected {
			t.Errorf("Expected message '%s', got '%s'", expected, msg)
		}
	})
}

func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector()

	t.Run("empty collector", func(t *testing.T) {
		if collector.HasErrors() {
			t.Error("Expected empty collector to have no errors")
		}
		if collector.GetErrorCount() != 0 {
			t.Errorf("Expected error count to be 0, got %d", collector.GetErrorCount())
		}
		if collector.GetSummary() != "没有错误" {
			t.Errorf("Expected summary to be '没有错误', got '%s'", collector.GetSummary())
		}
	})

	t.Run("add errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		collector.Add(err1)
		collector.Add(err2)
		collector.Add(nil) // should be ignored

		if !collector.HasErrors() {
			t.Error("Expected collector to have errors")
		}
		if collector.GetErrorCount() != 2 {
			t.Errorf("Expected error count to be 2, got %d", collector.GetErrorCount())
		}

		errors := collector.GetErrors()
		if len(errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(errors))
		}
		if errors[0] != err1 || errors[1] != err2 {
			t.Error("Errors not stored correctly")
		}
	})

	t.Run("summary", func(t *testing.T) {
		collector.Clear()
		collector.Add(errors.New("first error"))
		collector.Add(errors.New("second error"))

		summary := collector.GetSummary()
		if !strings.Contains(summary, "共发现 2 个错误") {
			t.Errorf("Summary should contain error count, got: %s", summary)
		}
		if !strings.Contains(summary, "1. first error") {
			t.Errorf("Summary should contain first error, got: %s", summary)
		}
		if !strings.Contains(summary, "2. second error") {
			t.Errorf("Summary should contain second error, got: %s", summary)
		}
	})

	t.Run("clear", func(t *testing.T) {
		collector.Add(errors.New("test error"))
		collector.Clear()

		if collector.HasErrors() {
			t.Error("Expected collector to be empty after clear")
		}
		if collector.GetErrorCount() != 0 {
			t.Errorf("Expected error count to be 0 after clear, got %d", collector.GetErrorCount())
		}
	})
}

func TestPDFError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	pdfErr := NewPDFError(ErrorInvalidFile, "test", "test.pdf", cause)

	unwrapped := pdfErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be the cause, got %v", unwrapped)
	}

	// Test with nil cause
	pdfErrNoCause := NewPDFError(ErrorInvalidFile, "test", "test.pdf", nil)
	unwrappedNil := pdfErrNoCause.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("Expected unwrapped error to be nil, got %v", unwrappedNil)
	}
}
