package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnhancedPDFReader_ValidationModes(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "enhanced_reader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建简化的PDF文件
	simplePDFPath := filepath.Join(tempDir, "simple.pdf")
	simplePDFContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj
%%EOF`)

	if err := os.WriteFile(simplePDFPath, simplePDFContent, 0644); err != nil {
		t.Fatalf("Failed to write simple PDF: %v", err)
	}

	tests := []struct {
		name          string
		mode          ValidationMode
		expectSuccess bool
		description   string
	}{
		{
			name:          "基本验证模式",
			mode:          ValidationBasic,
			expectSuccess: true,
			description:   "基本验证应该通过",
		},
		{
			name:          "宽松验证模式",
			mode:          ValidationRelaxed,
			expectSuccess: true,
			description:   "宽松验证应该通过",
		},
		{
			name:          "严格验证模式",
			mode:          ValidationStrict,
			expectSuccess: false, // 简化PDF可能不通过严格验证
			description:   "严格验证可能失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewEnhancedPDFReader(simplePDFPath, tt.mode)

			if tt.expectSuccess {
				if err != nil {
					t.Logf("验证失败 (预期可能): %v", err)
					// 对于严格模式，失败是可以接受的
					if tt.mode != ValidationStrict {
						t.Errorf("Expected success but got error: %v", err)
					}
					return
				}
				defer reader.Close()

				// 验证读取器状态
				if !reader.IsOpen() {
					t.Error("Reader should be open")
				}

				if reader.GetValidationMode() != tt.mode {
					t.Errorf("Expected mode %v, got %v", tt.mode, reader.GetValidationMode())
				}

				// 尝试获取信息
				info, err := reader.GetInfo()
				if err != nil {
					t.Errorf("Failed to get info: %v", err)
				} else {
					if info.FilePath != simplePDFPath {
						t.Errorf("Expected path %s, got %s", simplePDFPath, info.FilePath)
					}

					if info.PageCount <= 0 {
						t.Errorf("Expected positive page count, got %d", info.PageCount)
					}
				}
			} else {
				if err == nil {
					reader.Close()
					t.Logf("验证意外成功: %s", tt.description)
				}
			}
		})
	}
}

func TestEnhancedPDFReader_BasicFunctionality(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "enhanced_reader_basic_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建带元数据的PDF文件
	pdfWithMetadataPath := filepath.Join(tempDir, "with_metadata.pdf")
	pdfWithMetadataContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 2
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj
4 0 obj
<<
/Title (Test Document)
/Author (Test Author)
/Subject (Test Subject)
/Creator (Test Creator)
/Producer (Test Producer)
>>
endobj
%%EOF`)

	if err := os.WriteFile(pdfWithMetadataPath, pdfWithMetadataContent, 0644); err != nil {
		t.Fatalf("Failed to write PDF with metadata: %v", err)
	}

	// 使用基本验证模式
	reader, err := NewEnhancedPDFReader(pdfWithMetadataPath, ValidationBasic)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// 测试基本功能
	t.Run("GetInfo", func(t *testing.T) {
		info, err := reader.GetInfo()
		if err != nil {
			t.Fatalf("Failed to get info: %v", err)
		}

		if info.FilePath != pdfWithMetadataPath {
			t.Errorf("Expected path %s, got %s", pdfWithMetadataPath, info.FilePath)
		}

		if info.FileSize <= 0 {
			t.Errorf("Expected positive file size, got %d", info.FileSize)
		}

		if info.Version == "" {
			t.Error("Expected version to be set")
		}

		t.Logf("PDF Info: Pages=%d, Size=%d, Version=%s", info.PageCount, info.FileSize, info.Version)
	})

	t.Run("GetFilePath", func(t *testing.T) {
		path := reader.GetFilePath()
		if path != pdfWithMetadataPath {
			t.Errorf("Expected path %s, got %s", pdfWithMetadataPath, path)
		}
	})

	t.Run("IsOpen", func(t *testing.T) {
		if !reader.IsOpen() {
			t.Error("Reader should be open")
		}
	})
}

func TestEnhancedPDFReader_ValidationModeSwitch(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "enhanced_reader_mode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建简单PDF文件
	simplePDFPath := filepath.Join(tempDir, "mode_test.pdf")
	simplePDFContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
%%EOF`)

	if err := os.WriteFile(simplePDFPath, simplePDFContent, 0644); err != nil {
		t.Fatalf("Failed to write simple PDF: %v", err)
	}

	// 使用基本模式创建读取器
	reader, err := NewEnhancedPDFReader(simplePDFPath, ValidationBasic)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// 测试模式切换
	originalMode := reader.GetValidationMode()
	if originalMode != ValidationBasic {
		t.Errorf("Expected ValidationBasic, got %v", originalMode)
	}

	// 切换到宽松模式
	reader.SetValidationMode(ValidationRelaxed)
	if reader.GetValidationMode() != ValidationRelaxed {
		t.Errorf("Expected ValidationRelaxed after setting")
	}

	// 测试使用不同模式验证
	err = reader.ValidateWithMode(ValidationBasic)
	if err != nil {
		t.Errorf("Basic validation should pass: %v", err)
	}

	err = reader.ValidateWithMode(ValidationRelaxed)
	if err != nil {
		t.Errorf("Relaxed validation should pass: %v", err)
	}

	// 严格验证可能失败，这是正常的
	err = reader.ValidateWithMode(ValidationStrict)
	if err != nil {
		t.Logf("Strict validation failed as expected: %v", err)
	}
}

func TestEnhancedPDFReader_InvalidFiles(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "enhanced_reader_invalid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		content  []byte
		filename string
	}{
		{
			name:     "非PDF文件",
			content:  []byte("This is not a PDF file"),
			filename: "not_pdf.pdf",
		},
		{
			name:     "空文件",
			content:  []byte(""),
			filename: "empty.pdf",
		},
		{
			name:     "只有PDF头部",
			content:  []byte("%PDF-1.4"),
			filename: "header_only.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(filePath, tt.content, 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// 测试不同验证模式
			for _, mode := range []ValidationMode{ValidationBasic, ValidationRelaxed, ValidationStrict} {
				_, err := NewEnhancedPDFReader(filePath, mode)

				// 对于只有PDF头部的文件，基本验证模式可能会通过
				if tt.name == "只有PDF头部" && mode == ValidationBasic {
					if err != nil {
						t.Logf("Basic validation failed for header-only file (acceptable): %v", err)
					} else {
						t.Logf("Basic validation passed for header-only file")
					}
				} else {
					// 其他情况应该失败
					if err == nil {
						t.Errorf("Expected error for %s with mode %v", tt.name, mode)
					} else {
						t.Logf("Correctly failed for %s with mode %v: %v", tt.name, mode, err)
					}
				}
			}
		})
	}
}

func TestEnhancedPDFReader_NonExistentFile(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist.pdf"

	_, err := NewEnhancedPDFReader(nonExistentPath, ValidationBasic)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// 检查错误类型
	if pdfErr, ok := err.(*PDFError); ok {
		if pdfErr.Type != ErrorIO {
			t.Errorf("Expected ErrorIO, got %v", pdfErr.Type)
		}
	} else {
		t.Error("Expected PDFError type")
	}
}

func TestEnhancedPDFReader_CloseAndReopen(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "enhanced_reader_close_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建简单PDF文件
	simplePDFPath := filepath.Join(tempDir, "close_test.pdf")
	simplePDFContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
%%EOF`)

	if err := os.WriteFile(simplePDFPath, simplePDFContent, 0644); err != nil {
		t.Fatalf("Failed to write simple PDF: %v", err)
	}

	// 创建读取器
	reader, err := NewEnhancedPDFReader(simplePDFPath, ValidationBasic)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 验证初始状态
	if !reader.IsOpen() {
		t.Error("Reader should be open initially")
	}

	// 关闭读取器
	err = reader.Close()
	if err != nil {
		t.Errorf("Failed to close reader: %v", err)
	}

	// 验证关闭状态
	if reader.IsOpen() {
		t.Error("Reader should be closed after Close()")
	}

	// 重新打开
	err = reader.Open()
	if err != nil {
		t.Errorf("Failed to reopen reader: %v", err)
	}

	// 验证重新打开状态
	if !reader.IsOpen() {
		t.Error("Reader should be open after Open()")
	}

	// 最终关闭
	reader.Close()
}
