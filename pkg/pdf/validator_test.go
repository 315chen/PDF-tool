package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPDFValidator_ValidatePDFFile(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPDFValidator()

	tests := []struct {
		name        string
		setupFile   func() string
		expectError bool
		errorType   ErrorType
	}{
		{
			name: "有效的PDF文件",
			setupFile: func() string {
				file := filepath.Join(tempDir, "valid.pdf")
				content := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000010 00000 n \n0000000079 00000 n \n0000000173 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n253\n%%EOF"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: false,
		},
		{
			name: "无效的文件头",
			setupFile: func() string {
				file := filepath.Join(tempDir, "invalid_header.pdf")
				content := "NOT_PDF-1.4\nsome content\n%%EOF"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: true,
			errorType:   ErrorInvalidFile,
		},
		{
			name: "文件太小",
			setupFile: func() string {
				file := filepath.Join(tempDir, "too_small.pdf")
				content := "%P"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: true,
			errorType:   ErrorInvalidFile,
		},
		{
			name: "缺少EOF标记",
			setupFile: func() string {
				file := filepath.Join(tempDir, "no_eof.pdf")
				content := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: true,
			errorType:   ErrorCorrupted,
		},
		{
			name: "不支持的PDF版本",
			setupFile: func() string {
				file := filepath.Join(tempDir, "unsupported_version.pdf")
				content := "%PDF-3.0\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\n%%EOF"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: true,
			errorType:   ErrorInvalidFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			err := validator.ValidatePDFFile(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				} else {
					if pdfErr, ok := err.(*PDFError); ok {
						if pdfErr.Type != tt.errorType {
							t.Errorf("错误类型不匹配，期望: %v, 实际: %v", tt.errorType, pdfErr.Type)
						}
					} else {
						t.Errorf("期望PDFError类型，实际: %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("不期望出现错误，但出现了: %v", err)
				}
			}
		})
	}
}

func TestPDFValidator_isValidPDFVersion(t *testing.T) {
	validator := NewPDFValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"-1.4", true},
		{"-1.7", true},
		{"-2.0", true},
		{"-1.0", true},
		{"-3.0", false},
		{"-0.9", false},
		{"1.4", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := validator.isValidPDFVersion(tt.version)
			if result != tt.valid {
				t.Errorf("版本 %s 的验证结果不正确，期望: %v, 实际: %v", tt.version, tt.valid, result)
			}
		})
	}
}

func TestPDFValidator_GetBasicPDFInfo(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPDFValidator()

	// 创建有效的PDF文件
	validPDFContent := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000010 00000 n \n0000000079 00000 n \n0000000173 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n253\n%%EOF"
	validFile := filepath.Join(tempDir, "valid.pdf")
	err := os.WriteFile(validFile, []byte(validPDFContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试获取PDF信息
	info, err := validator.GetBasicPDFInfo(validFile)
	if err != nil {
		t.Fatalf("获取PDF信息失败: %v", err)
	}

	// 验证信息
	if info.FileSize != int64(len(validPDFContent)) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", len(validPDFContent), info.FileSize)
	}

	// 使用pdfcpu后，我们可以获取真实的页数
	if info.PageCount <= 0 {
		t.Errorf("页数应该大于0，实际: %d", info.PageCount)
	}
}

func TestPDFValidator_isPDFEncrypted(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPDFValidator()

	tests := []struct {
		name      string
		content   string
		encrypted bool
	}{
		{
			name:      "未加密PDF",
			content:   "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\n%%EOF",
			encrypted: false,
		},
		{
			name:      "包含加密标记的PDF",
			content:   "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Encrypt 2 0 R\n>>\nendobj\n%%EOF",
			encrypted: true,
		},
		{
			name:      "包含Filter标记的PDF",
			content:   "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Filter /Standard\n>>\nendobj\n%%EOF",
			encrypted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(tempDir, tt.name+".pdf")
			err := os.WriteFile(file, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("创建测试文件失败: %v", err)
			}

			encrypted, err := validator.isPDFEncrypted(file)
			if err != nil {
				t.Errorf("检查加密状态失败: %v", err)
			}

			if encrypted != tt.encrypted {
				t.Errorf("加密状态不匹配，期望: %v, 实际: %v", tt.encrypted, encrypted)
			}
		})
	}
}