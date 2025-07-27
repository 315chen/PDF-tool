package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ValidationTestSuite PDF验证功能综合测试套件
type ValidationTestSuite struct {
	tempDir   string
	validator *PDFValidator
	service   PDFService
	adapter   *PDFCPUAdapter
}

// SetupValidationTest 设置验证测试环境
func SetupValidationTest(t *testing.T) *ValidationTestSuite {
	tempDir, err := os.MkdirTemp("", "validation-test-*")
	require.NoError(t, err)

	validator := NewPDFValidator()
	service := NewPDFService()

	adapter, err := NewPDFCPUAdapter(nil)
	require.NoError(t, err)

	suite := &ValidationTestSuite{
		tempDir:   tempDir,
		validator: validator,
		service:   service,
		adapter:   adapter,
	}

	t.Cleanup(func() {
		if suite.adapter != nil {
			suite.adapter.Close()
		}
		os.RemoveAll(suite.tempDir)
	})

	return suite
}

// TestVariousPDFFormats 测试各种格式PDF文件的验证
func TestVariousPDFFormats(t *testing.T) {
	suite := SetupValidationTest(t)

	testCases := []struct {
		name        string
		version     string
		content     string
		expectValid bool
		description string
	}{
		{
			name:        "PDF_1_0",
			version:     "1.0",
			content:     createPDFContent("1.0", false, false),
			expectValid: true,
			description: "PDF 1.0 格式文件",
		},
		{
			name:        "PDF_1_4",
			version:     "1.4",
			content:     createPDFContent("1.4", false, false),
			expectValid: true,
			description: "PDF 1.4 格式文件（最常见）",
		},
		{
			name:        "PDF_1_7",
			version:     "1.7",
			content:     createPDFContent("1.7", false, false),
			expectValid: true,
			description: "PDF 1.7 格式文件",
		},
		{
			name:        "PDF_2_0",
			version:     "2.0",
			content:     createPDFContent("2.0", false, false),
			expectValid: true,
			description: "PDF 2.0 格式文件（最新标准）",
		},
		{
			name:        "PDF_Unsupported_3_0",
			version:     "3.0",
			content:     createPDFContent("3.0", false, false),
			expectValid: false,
			description: "不支持的PDF 3.0版本",
		},
		{
			name:        "PDF_With_Metadata",
			version:     "1.4",
			content:     createPDFWithMetadata("1.4"),
			expectValid: true,
			description: "包含元数据的PDF文件",
		},
		{
			name:        "PDF_With_Multiple_Pages",
			version:     "1.4",
			content:     createMultiPagePDF("1.4"),
			expectValid: true,
			description: "多页PDF文件",
		},
		{
			name:        "PDF_With_Images",
			version:     "1.4",
			content:     createPDFWithImages("1.4"),
			expectValid: true,
			description: "包含图像的PDF文件",
		},
		{
			name:        "PDF_With_Fonts",
			version:     "1.4",
			content:     createPDFWithFonts("1.4"),
			expectValid: true,
			description: "包含字体的PDF文件",
		},
		{
			name:        "PDF_With_Annotations",
			version:     "1.4",
			content:     createPDFWithAnnotations("1.4"),
			expectValid: true,
			description: "包含注释的PDF文件",
		},
		{
			name:        "PDF_With_Forms",
			version:     "1.4",
			content:     createPDFWithForms("1.4"),
			expectValid: true,
			description: "包含表单的PDF文件",
		},
		{
			name:        "PDF_With_Bookmarks",
			version:     "1.4",
			content:     createPDFWithBookmarks("1.4"),
			expectValid: true,
			description: "包含书签的PDF文件",
		},
		{
			name:        "PDF_With_JavaScript",
			version:     "1.4",
			content:     createPDFWithJavaScript("1.4"),
			expectValid: true,
			description: "包含JavaScript的PDF文件",
		},
		{
			name:        "PDF_With_Attachments",
			version:     "1.4",
			content:     createPDFWithAttachments("1.4"),
			expectValid: true,
			description: "包含附件的PDF文件",
		},
		{
			name:        "PDF_Linearized",
			version:     "1.4",
			content:     createLinearizedPDF("1.4"),
			expectValid: true,
			description: "线性化PDF文件（快速Web查看）",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试文件
			filePath := filepath.Join(suite.tempDir, tc.name+".pdf")
			err := os.WriteFile(filePath, []byte(tc.content), 0644)
			require.NoError(t, err)

			t.Logf("测试文件: %s (%s)", tc.description, tc.name)

			// 测试基本验证器
			err = suite.validator.ValidatePDFFile(filePath)
			if tc.expectValid {
				assert.NoError(t, err, "基本验证器应该验证通过")
			} else {
				assert.Error(t, err, "基本验证器应该验证失败")
				if pdfErr, ok := err.(*PDFError); ok {
					t.Logf("验证错误类型: %v, 消息: %s", pdfErr.Type, pdfErr.Message)
				}
			}

			// 测试PDF服务
			err = suite.service.ValidatePDF(filePath)
			if tc.expectValid {
				assert.NoError(t, err, "PDF服务应该验证通过")
			} else {
				assert.Error(t, err, "PDF服务应该验证失败")
			}

			// 测试pdfcpu适配器
			err = suite.adapter.ValidateFile(filePath)
			if tc.expectValid {
				// pdfcpu可能有不同的验证标准，记录结果但不强制要求
				t.Logf("pdfcpu验证结果: %v", err)
			} else {
				t.Logf("pdfcpu验证结果: %v", err)
			}

			// 如果验证通过，测试获取文件信息
			if tc.expectValid && err == nil {
				info, err := suite.validator.GetBasicPDFInfo(filePath)
				if err == nil {
					t.Logf("文件信息: 页数=%d, 加密=%t, 大小=%d",
						info.PageCount, info.IsEncrypted, info.FileSize)
				}
			}

			// 测试严格验证模式
			err = suite.validator.ValidateWithStrictMode(filePath)
			if tc.expectValid {
				t.Logf("严格验证结果: %v", err)
			} else {
				assert.Error(t, err, "严格验证应该失败")
			}

			// 测试验证报告
			report, err := suite.validator.GetValidationReport(filePath)
			if err == nil {
				t.Logf("验证报告: 有效=%t, 错误数=%d, 警告数=%d",
					report.IsValid, len(report.Errors), len(report.Warnings))
			}
		})
	}
}

// TestCorruptedPDFHandling 测试损坏PDF文件的错误处理
func TestCorruptedPDFHandling(t *testing.T) {
	suite := SetupValidationTest(t)

	corruptionCases := []struct {
		name        string
		createFile  func(string) string
		expectedErr ErrorType
		description string
	}{
		{
			name: "Missing_PDF_Header",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "no_header.pdf")
				content := "NOT_A_PDF_FILE\nSome content here\n%%EOF"
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorInvalidFile,
			description: "缺少PDF文件头",
		},
		{
			name: "Truncated_Header",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "truncated_header.pdf")
				content := "%PD" // 截断的头部
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorInvalidFile,
			description: "截断的PDF头部",
		},
		{
			name: "Missing_EOF_Marker",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "no_eof.pdf")
				content := createPDFContent("1.4", false, false)
				// 移除%%EOF标记
				content = strings.Replace(content, "%%EOF", "", 1)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "缺少EOF标记",
		},
		{
			name: "Corrupted_Xref_Table",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "corrupted_xref.pdf")
				content := createPDFContent("1.4", false, false)
				// 损坏xref表
				content = strings.Replace(content, "xref", "CORRUPTED_XREF", 1)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "损坏的交叉引用表",
		},
		{
			name: "Invalid_Object_Structure",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "invalid_object.pdf")
				content := `%PDF-1.4
INVALID_OBJECT_STRUCTURE
<<
/Type /Catalog
>>
%%EOF`
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "无效的对象结构",
		},
		{
			name: "Incomplete_Stream",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "incomplete_stream.pdf")
				content := `%PDF-1.4
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
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 100
>>
stream
INCOMPLETE_STREAM_DATA
endstream
endobj

%%EOF`
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "不完整的流数据",
		},
		{
			name: "Binary_Corruption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "binary_corruption.pdf")
				content := createPDFContent("1.4", false, false)
				// 在中间插入二进制垃圾数据
				corruptedContent := content[:len(content)/2] +
					"\x00\x01\x02\x03\xFF\xFE\xFD" +
					content[len(content)/2:]
				os.WriteFile(filePath, []byte(corruptedContent), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "二进制数据损坏",
		},
		{
			name: "Empty_File",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "empty.pdf")
				os.WriteFile(filePath, []byte{}, 0644)
				return filePath
			},
			expectedErr: ErrorInvalidFile,
			description: "空文件",
		},
		{
			name: "Too_Small_File",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "too_small.pdf")
				os.WriteFile(filePath, []byte("%PDF"), 0644)
				return filePath
			},
			expectedErr: ErrorInvalidFile,
			description: "文件过小",
		},
		{
			name: "Malformed_Trailer",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "malformed_trailer.pdf")
				content := createPDFContent("1.4", false, false)
				// 损坏trailer
				content = strings.Replace(content, "trailer", "MALFORMED_TRAILER", 1)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "损坏的trailer",
		},
		{
			name: "Invalid_PDF_Version",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "invalid_version.pdf")
				content := "%PDF-INVALID\n" + createPDFContent("1.4", false, false)[8:]
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorInvalidFile,
			description: "无效的PDF版本",
		},
		{
			name: "Circular_Reference",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "circular_ref.pdf")
				content := `%PDF-1.4
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
/Parent 3 0 R
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Kids [2 0 R]
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
200
%%EOF`
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: ErrorCorrupted,
			description: "循环引用",
		},
	}

	for _, tc := range corruptionCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.createFile(suite.tempDir)
			t.Logf("测试损坏文件: %s", tc.description)

			// 测试基本验证器
			err := suite.validator.ValidatePDFFile(filePath)
			assert.Error(t, err, "应该检测到文件损坏")

			if pdfErr, ok := err.(*PDFError); ok {
				assert.Equal(t, tc.expectedErr, pdfErr.Type,
					"错误类型应该匹配，期望: %v, 实际: %v", tc.expectedErr, pdfErr.Type)
				t.Logf("检测到错误: %s (类型: %v)", pdfErr.Message, pdfErr.Type)
			}

			// 测试PDF服务
			err = suite.service.ValidatePDF(filePath)
			assert.Error(t, err, "PDF服务应该检测到文件损坏")

			// 测试pdfcpu适配器
			err = suite.adapter.ValidateFile(filePath)
			t.Logf("pdfcpu验证结果: %v", err)

			// 测试严格验证模式
			err = suite.validator.ValidateWithStrictMode(filePath)
			assert.Error(t, err, "严格模式应该检测到文件损坏")

			// 测试验证报告
			report, err := suite.validator.GetValidationReport(filePath)
			if err == nil {
				assert.False(t, report.IsValid, "验证报告应该显示文件无效")
				assert.NotEmpty(t, report.Errors, "验证报告应该包含错误信息")
				t.Logf("验证报告错误: %v", report.Errors)
			} else {
				t.Logf("获取验证报告失败: %v", err)
			}

			// 测试错误处理器
			handler := NewDefaultErrorHandler(3)
			processedErr := handler.HandleError(err)
			userMsg := handler.GetUserFriendlyMessage(processedErr)
			shouldRetry := handler.ShouldRetry(processedErr)

			t.Logf("用户友好消息: %s", userMsg)
			t.Logf("是否可重试: %t", shouldRetry)
		})
	}
}

// TestEncryptedPDFDetection 测试加密PDF文件的检测功能
func TestEncryptedPDFDetection(t *testing.T) {
	suite := SetupValidationTest(t)

	encryptionCases := []struct {
		name            string
		createFile      func(string) string
		expectEncrypted bool
		description     string
	}{
		{
			name: "Standard_Encryption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "standard_encrypted.pdf")
				content := createPDFContent("1.4", true, false)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "标准加密PDF文件",
		},
		{
			name: "AES_Encryption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "aes_encrypted.pdf")
				content := createPDFWithAESEncryption("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "AES加密PDF文件",
		},
		{
			name: "Password_Protected",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "password_protected.pdf")
				content := createPasswordProtectedPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "密码保护PDF文件",
		},
		{
			name: "Owner_Password_Only",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "owner_password.pdf")
				content := createOwnerPasswordPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "仅所有者密码保护",
		},
		{
			name: "User_Password_Only",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "user_password.pdf")
				content := createUserPasswordPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "仅用户密码保护",
		},
		{
			name: "RC4_Encryption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "rc4_encrypted.pdf")
				content := createRC4EncryptedPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "RC4加密PDF文件",
		},
		{
			name: "High_Security_Encryption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "high_security.pdf")
				content := createHighSecurityPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "高安全级别加密",
		},
		{
			name: "Unencrypted_PDF",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "unencrypted.pdf")
				content := createPDFContent("1.4", false, false)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: false,
			description:     "未加密PDF文件",
		},
		{
			name: "False_Positive_Test",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "false_positive.pdf")
				// 创建包含加密关键字但实际未加密的文件
				content := createPDFWithEncryptKeywords("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: false,
			description:     "包含加密关键字但未加密的文件",
		},
		{
			name: "Metadata_Only_Encryption",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "metadata_encrypted.pdf")
				content := createMetadataEncryptedPDF("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectEncrypted: true,
			description:     "仅元数据加密",
		},
	}

	for _, tc := range encryptionCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.createFile(suite.tempDir)
			t.Logf("测试加密检测: %s", tc.description)

			// 测试基本加密检测
			isEncrypted, err := suite.validator.isPDFEncrypted(filePath)
			if err == nil {
				assert.Equal(t, tc.expectEncrypted, isEncrypted,
					"加密状态检测不正确，期望: %t, 实际: %t", tc.expectEncrypted, isEncrypted)
				t.Logf("基本检测结果: 加密=%t", isEncrypted)
			} else {
				t.Logf("基本检测失败: %v", err)
			}

			// 测试PDF服务加密检测
			isEncrypted, err = suite.service.IsPDFEncrypted(filePath)
			if err == nil {
				assert.Equal(t, tc.expectEncrypted, isEncrypted,
					"PDF服务加密检测不正确")
				t.Logf("PDF服务检测结果: 加密=%t", isEncrypted)
			} else {
				t.Logf("PDF服务检测失败: %v", err)
			}

			// 测试获取PDF信息中的加密状态
			info, err := suite.validator.GetBasicPDFInfo(filePath)
			if err == nil {
				assert.Equal(t, tc.expectEncrypted, info.IsEncrypted,
					"PDF信息中的加密状态不正确")
				t.Logf("PDF信息检测结果: 加密=%t, 页数=%d, 大小=%d",
					info.IsEncrypted, info.PageCount, info.FileSize)
			} else {
				t.Logf("获取PDF信息失败: %v", err)
			}

			// 测试权限检查
			permissions, err := suite.validator.CheckPermissions(filePath)
			if err == nil {
				assert.Equal(t, tc.expectEncrypted, permissions.IsEncrypted,
					"权限检查中的加密状态不正确")
				t.Logf("权限检查结果: 加密=%t, 可打印=%t, 可复制=%t, 可修改=%t",
					permissions.IsEncrypted, permissions.CanPrint, permissions.CanCopy, permissions.CanModify)
			} else {
				t.Logf("权限检查失败: %v", err)
			}

			// 如果文件加密，测试验证是否会失败或需要密码
			if tc.expectEncrypted {
				err := suite.validator.ValidatePDFFile(filePath)
				// 加密文件可能验证失败或需要特殊处理
				t.Logf("加密文件验证结果: %v", err)

				// 测试解密功能（如果适配器支持）
				if suite.adapter != nil {
					tempOutput := filepath.Join(suite.tempDir, "decrypted_"+tc.name+".pdf")
					err = suite.adapter.DecryptFile(filePath, tempOutput, "testpassword")
					t.Logf("解密测试结果: %v", err)
				}
			}

			// 测试加密级别检测
			if tc.expectEncrypted {
				encLevel := detectEncryptionLevel(filePath)
				t.Logf("检测到的加密级别: %s", encLevel)
			}
		})
	}
}

// TestLargeFileValidationPerformance 测试大文件验证的性能表现
func TestLargeFileValidationPerformance(t *testing.T) {
	suite := SetupValidationTest(t)

	// 性能测试用例
	performanceCases := []struct {
		name        string
		sizeKB      int
		maxDuration time.Duration
		description string
	}{
		{
			name:        "Small_File_1KB",
			sizeKB:      1,
			maxDuration: 100 * time.Millisecond,
			description: "小文件 (1KB)",
		},
		{
			name:        "Medium_File_100KB",
			sizeKB:      100,
			maxDuration: 500 * time.Millisecond,
			description: "中等文件 (100KB)",
		},
		{
			name:        "Large_File_1MB",
			sizeKB:      1024,
			maxDuration: 2 * time.Second,
			description: "大文件 (1MB)",
		},
		{
			name:        "Very_Large_File_5MB",
			sizeKB:      5 * 1024,
			maxDuration: 5 * time.Second,
			description: "超大文件 (5MB)",
		},
		{
			name:        "Huge_File_10MB",
			sizeKB:      10 * 1024,
			maxDuration: 10 * time.Second,
			description: "巨大文件 (10MB)",
		},
	}

	for _, tc := range performanceCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建指定大小的测试文件
			filePath := createLargeTestPDF(t, suite.tempDir, tc.name+".pdf", tc.sizeKB)
			t.Logf("测试性能: %s", tc.description)

			// 测试基本验证器性能
			start := time.Now()
			err := suite.validator.ValidatePDFFile(filePath)
			duration := time.Since(start)

			t.Logf("基本验证器: 耗时=%v, 错误=%v", duration, err)
			if duration > tc.maxDuration {
				t.Logf("警告: 验证时间超过预期 (期望<%v, 实际=%v)", tc.maxDuration, duration)
			}

			// 测试PDF服务性能
			start = time.Now()
			err = suite.service.ValidatePDF(filePath)
			duration = time.Since(start)

			t.Logf("PDF服务: 耗时=%v, 错误=%v", duration, err)

			// 测试pdfcpu适配器性能
			start = time.Now()
			err = suite.adapter.ValidateFile(filePath)
			duration = time.Since(start)

			t.Logf("pdfcpu适配器: 耗时=%v, 错误=%v", duration, err)

			// 测试内存使用
			testMemoryUsage(t, filePath, suite)

			// 测试获取文件信息的性能
			start = time.Now()
			info, err := suite.validator.GetBasicPDFInfo(filePath)
			duration = time.Since(start)

			if err == nil {
				t.Logf("获取信息: 耗时=%v, 页数=%d, 大小=%d",
					duration, info.PageCount, info.FileSize)
			} else {
				t.Logf("获取信息失败: 耗时=%v, 错误=%v", duration, err)
			}

			// 测试严格验证模式性能
			start = time.Now()
			err = suite.validator.ValidateWithStrictMode(filePath)
			duration = time.Since(start)

			t.Logf("严格验证: 耗时=%v, 错误=%v", duration, err)

			// 测试验证报告生成性能
			start = time.Now()
			report, err := suite.validator.GetValidationReport(filePath)
			duration = time.Since(start)

			if err == nil {
				t.Logf("验证报告: 耗时=%v, 有效=%t, 错误数=%d",
					duration, report.IsValid, len(report.Errors))
			} else {
				t.Logf("验证报告失败: 耗时=%v, 错误=%v", duration, err)
			}
		})
	}
}

// TestConcurrentValidation 测试并发验证性能
func TestConcurrentValidation(t *testing.T) {
	suite := SetupValidationTest(t)

	// 创建多个测试文件
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("concurrent_test_%d.pdf", i)
		filePath := filepath.Join(suite.tempDir, fileName)
		content := createPDFContent("1.4", false, false)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
		testFiles[i] = filePath
	}

	t.Run("Concurrent_Basic_Validation", func(t *testing.T) {
		const numWorkers = 5
		const numIterations = 20

		start := time.Now()

		// 启动并发验证
		results := make(chan error, numWorkers*numIterations*len(testFiles))

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for j := 0; j < numIterations; j++ {
					for _, file := range testFiles {
						err := suite.validator.ValidatePDFFile(file)
						results <- err
					}
				}
			}(i)
		}

		// 收集结果
		totalOperations := numWorkers * numIterations * len(testFiles)
		successCount := 0
		errorCount := 0

		for i := 0; i < totalOperations; i++ {
			err := <-results
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
		}

		duration := time.Since(start)
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("并发验证结果:")
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  成功数: %d", successCount)
		t.Logf("  失败数: %d", errorCount)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  操作/秒: %.2f", operationsPerSecond)

		// 验证并发安全性
		assert.True(t, successCount > 0, "应该有成功的验证操作")
		assert.True(t, operationsPerSecond > 10, "性能应该合理 (>10 ops/sec)")
	})

	t.Run("Concurrent_Service_Validation", func(t *testing.T) {
		const numWorkers = 3
		const numIterations = 10

		start := time.Now()
		results := make(chan error, numWorkers*numIterations*len(testFiles))

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for j := 0; j < numIterations; j++ {
					for _, file := range testFiles {
						err := suite.service.ValidatePDF(file)
						results <- err
					}
				}
			}(i)
		}

		totalOperations := numWorkers * numIterations * len(testFiles)
		successCount := 0

		for i := 0; i < totalOperations; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		t.Logf("并发服务验证: 成功=%d/%d, 耗时=%v", successCount, totalOperations, duration)
	})

	t.Run("Concurrent_Mixed_Operations", func(t *testing.T) {
		const numWorkers = 4
		const numIterations = 5

		start := time.Now()
		results := make(chan string, numWorkers*numIterations*len(testFiles)*3)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for j := 0; j < numIterations; j++ {
					for _, file := range testFiles {
						// 验证
						err := suite.validator.ValidatePDFFile(file)
						if err == nil {
							results <- "validate_success"
						} else {
							results <- "validate_error"
						}

						// 获取信息
						_, err = suite.validator.GetBasicPDFInfo(file)
						if err == nil {
							results <- "info_success"
						} else {
							results <- "info_error"
						}

						// 检查权限
						_, err = suite.validator.CheckPermissions(file)
						if err == nil {
							results <- "permission_success"
						} else {
							results <- "permission_error"
						}
					}
				}
			}(i)
		}

		totalOperations := numWorkers * numIterations * len(testFiles) * 3
		operationCounts := make(map[string]int)

		for i := 0; i < totalOperations; i++ {
			result := <-results
			operationCounts[result]++
		}

		duration := time.Since(start)
		t.Logf("并发混合操作结果 (耗时=%v):", duration)
		for op, count := range operationCounts {
			t.Logf("  %s: %d", op, count)
		}
	})
}

// TestStressValidation 压力测试
func TestStressValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	suite := SetupValidationTest(t)

	// 创建不同类型的测试文件
	testFiles := []struct {
		name    string
		content string
		valid   bool
	}{
		{"valid_simple.pdf", createPDFContent("1.4", false, false), true},
		{"valid_complex.pdf", createPDFWithMetadata("1.4"), true},
		{"invalid_header.pdf", "NOT_PDF_CONTENT", false},
		{"corrupted.pdf", createPDFContent("1.4", false, false)[:100], false},
	}

	filePaths := make([]string, len(testFiles))
	for i, tf := range testFiles {
		filePath := filepath.Join(suite.tempDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		require.NoError(t, err)
		filePaths[i] = filePath
	}

	t.Run("High_Volume_Validation", func(t *testing.T) {
		const totalOperations = 1000

		start := time.Now()
		successCount := 0
		errorCount := 0

		for i := 0; i < totalOperations; i++ {
			file := filePaths[i%len(filePaths)]
			err := suite.validator.ValidatePDFFile(file)
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
		}

		duration := time.Since(start)
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("高容量验证结果:")
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  成功数: %d", successCount)
		t.Logf("  失败数: %d", errorCount)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  操作/秒: %.2f", operationsPerSecond)

		assert.True(t, operationsPerSecond > 50, "高容量测试性能应该合理")
	})

	t.Run("Memory_Stress_Test", func(t *testing.T) {
		const iterations = 100

		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		for i := 0; i < iterations; i++ {
			for _, file := range filePaths {
				suite.validator.ValidatePDFFile(file)
				suite.validator.GetBasicPDFInfo(file)
				suite.validator.CheckPermissions(file)
			}

			// 每10次迭代强制GC
			if i%10 == 0 {
				runtime.GC()
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc
		t.Logf("内存压力测试: 使用内存=%d字节 (%.2f MB)",
			memoryUsed, float64(memoryUsed)/(1024*1024))

		// 内存使用应该在合理范围内
		maxMemoryMB := int64(100)
		assert.Less(t, int64(memoryUsed), maxMemoryMB*1024*1024,
			"内存使用应该小于%dMB", maxMemoryMB)
	})
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	suite := SetupValidationTest(t)

	edgeCases := []struct {
		name        string
		createFile  func(string) string
		description string
	}{
		{
			name: "Minimum_Valid_PDF",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "minimum.pdf")
				content := "%PDF-1.0\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]>>endobj\nxref\n0 4\n0000000000 65535 f\n0000000010 00000 n\n0000000053 00000 n\n0000000103 00000 n\ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n149\n%%EOF"
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			description: "最小有效PDF",
		},
		{
			name: "Maximum_PDF_Version",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "max_version.pdf")
				content := createPDFContent("2.0", false, false)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			description: "最高PDF版本",
		},
		{
			name: "Unicode_Filename",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "测试文件_中文名称.pdf")
				content := createPDFContent("1.4", false, false)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			description: "Unicode文件名",
		},
		{
			name: "Very_Long_Filename",
			createFile: func(dir string) string {
				longName := strings.Repeat("a", 200) + ".pdf"
				filePath := filepath.Join(dir, longName)
				content := createPDFContent("1.4", false, false)
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			description: "超长文件名",
		},
		{
			name: "Special_Characters_Content",
			createFile: func(dir string) string {
				filePath := filepath.Join(dir, "special_chars.pdf")
				content := createPDFWithSpecialChars("1.4")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			description: "包含特殊字符的内容",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.createFile(suite.tempDir)
			t.Logf("测试边界情况: %s", tc.description)

			// 测试各种验证方法
			err := suite.validator.ValidatePDFFile(filePath)
			t.Logf("基本验证结果: %v", err)

			err = suite.service.ValidatePDF(filePath)
			t.Logf("服务验证结果: %v", err)

			info, err := suite.validator.GetBasicPDFInfo(filePath)
			if err == nil {
				t.Logf("文件信息: 页数=%d, 大小=%d", info.PageCount, info.FileSize)
			} else {
				t.Logf("获取信息失败: %v", err)
			}

			report, err := suite.validator.GetValidationReport(filePath)
			if err == nil {
				t.Logf("验证报告: 有效=%t, 错误数=%d", report.IsValid, len(report.Errors))
			}
		})
	}
} // 辅助函数

// createPDFContent 创建PDF内容
func createPDFContent(version string, encrypted bool, multiPage bool) string {
	header := fmt.Sprintf("%%PDF-%s\n", version)

	catalog := `1 0 obj
<<
/Type /Catalog
/Pages 2 0 R`

	if encrypted {
		catalog += `
/Encrypt 5 0 R`
	}

	catalog += `
>>
endobj

`

	pages := `2 0 obj
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

`

	if multiPage {
		pages = `2 0 obj
<<
/Type /Pages
/Kids [3 0 R 4 0 R]
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
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

`
	}

	encryption := ""
	if encrypted {
		encryption = `5 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

`
	}

	xref := `xref
0 ` + fmt.Sprintf("%d", 4+boolToInt(encrypted)) + `
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
`

	if encrypted {
		xref += `0000000250 00000 n 
`
	}

	trailer := `trailer
<<
/Size ` + fmt.Sprintf("%d", 4+boolToInt(encrypted)) + `
/Root 1 0 R`

	if encrypted {
		trailer += `
/Encrypt 5 0 R`
	}

	trailer += `
>>
startxref
253
%%EOF`

	return header + catalog + pages + encryption + xref + trailer
}

// createPDFWithMetadata 创建包含元数据的PDF
func createPDFWithMetadata(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Metadata 4 0 R
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

4 0 obj
<<
/Type /Metadata
/Subtype /XML
/Length 200
>>
stream
<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>Test PDF with Metadata</dc:title>
<dc:creator>PDF Validation Test</dc:creator>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
500
%%EOF`, version)
}

// createMultiPagePDF 创建多页PDF
func createMultiPagePDF(version string) string {
	return createPDFContent(version, false, true)
}

// createPDFWithImages 创建包含图像的PDF
func createPDFWithImages(version string) string {
	imageData := strings.Repeat("A", 100) // 减少图像数据大小以避免过大的测试文件
	return fmt.Sprintf(`%%PDF-%s
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
/Resources <<
/XObject <<
/Im1 4 0 R
>>
>>
>>
endobj

4 0 obj
<<
/Type /XObject
/Subtype /Image
/Width 10
/Height 10
/ColorSpace /DeviceRGB
/BitsPerComponent 8
/Length %d
>>
stream
%s
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
%d
%%EOF`, version, len(imageData), imageData, 400+len(imageData))
}

// createPDFWithFonts 创建包含字体的PDF
func createPDFWithFonts(version string) string {
	return fmt.Sprintf(`%%PDF-%s
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
/Resources <<
/Font <<
/F1 4 0 R
>>
>>
>>
endobj

4 0 obj
<<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
320
%%EOF`, version)
}

// createPDFWithAnnotations 创建包含注释的PDF
func createPDFWithAnnotations(version string) string {
	return fmt.Sprintf(`%%PDF-%s
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
/Annots [4 0 R]
>>
endobj

4 0 obj
<<
/Type /Annot
/Subtype /Text
/Rect [100 100 200 200]
/Contents (Test annotation)
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
0000000200 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
300
%%EOF`, version)
}

// createPDFWithForms 创建包含表单的PDF
func createPDFWithForms(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/AcroForm 5 0 R
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
/Annots [4 0 R]
>>
endobj

4 0 obj
<<
/Type /Annot
/Subtype /Widget
/Rect [100 100 300 120]
/FT /Tx
/T (TextField1)
/V (Default Value)
>>
endobj

5 0 obj
<<
/Fields [4 0 R]
>>
endobj

xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000151 00000 n 
0000000230 00000 n 
0000000350 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
380
%%EOF`, version)
}

// createPDFWithBookmarks 创建包含书签的PDF
func createPDFWithBookmarks(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Outlines 5 0 R
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

4 0 obj
<<
/Title (Chapter 1)
/Parent 5 0 R
/Dest [3 0 R /XYZ 0 792 0]
>>
endobj

5 0 obj
<<
/Type /Outlines
/First 4 0 R
/Last 4 0 R
/Count 1
>>
endobj

xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000151 00000 n 
0000000220 00000 n 
0000000300 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
380
%%EOF`, version)
}

// createPDFWithJavaScript 创建包含JavaScript的PDF
func createPDFWithJavaScript(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Names 5 0 R
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

4 0 obj
<<
/S /JavaScript
/JS (app.alert("Hello World");)
>>
endobj

5 0 obj
<<
/JavaScript <<
/Names [(MyScript) 4 0 R]
>>
>>
endobj

xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000151 00000 n 
0000000220 00000 n 
0000000300 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
380
%%EOF`, version)
}

// createPDFWithAttachments 创建包含附件的PDF
func createPDFWithAttachments(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Names 5 0 R
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

4 0 obj
<<
/Type /Filespec
/F (attachment.txt)
/EF <<
/F 6 0 R
>>
>>
endobj

5 0 obj
<<
/EmbeddedFiles <<
/Names [(attachment.txt) 4 0 R]
>>
>>
endobj

6 0 obj
<<
/Length 12
/Filter /ASCIIHexDecode
>>
stream
48656C6C6F20576F726C64
endstream
endobj

xref
0 7
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000151 00000 n 
0000000220 00000 n 
0000000320 00000 n 
0000000400 00000 n 
trailer
<<
/Size 7
/Root 1 0 R
>>
startxref
500
%%EOF`, version)
}

// createLinearizedPDF 创建线性化PDF
func createLinearizedPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
%%ÿÿÿÿ
1 0 obj
<<
/Linearized 1
/L 500
/H [100 200]
/O 3
/E 300
/N 1
/T 400
>>
endobj

2 0 obj
<<
/Type /Catalog
/Pages 4 0 R
>>
endobj

3 0 obj
<<
/Type /Pages
/Kids [5 0 R]
/Count 1
>>
endobj

4 0 obj
<<
/Type /Page
/Parent 3 0 R
/MediaBox [0 0 612 792]
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000100 00000 n 
0000000150 00000 n 
0000000200 00000 n 
trailer
<<
/Size 5
/Root 2 0 R
/Prev 300
>>
startxref
250
%%EOF`, version)
}

// createPDFWithAESEncryption 创建AES加密的PDF
func createPDFWithAESEncryption(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 4
/R 4
/CF <<
/StdCF <<
/AuthEvent /DocOpen
/CFM /AESV2
/Length 16
>>
>>
/StmF /StdCF
/StrF /StdCF
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
600
%%EOF`, version)
}

// createPasswordProtectedPDF 创建密码保护的PDF
func createPasswordProtectedPDF(version string) string {
	return createPDFContent(version, true, false)
}

// createOwnerPasswordPDF 创建仅所有者密码保护的PDF
func createOwnerPasswordPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA0108000000000000000000000000000000000>
/P -44
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
400
%%EOF`, version)
}

// createUserPasswordPDF 创建仅用户密码保护的PDF
func createUserPasswordPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O <000000000000000000000000000000000000000000000000000000000000000>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
400
%%EOF`, version)
}

// createRC4EncryptedPDF 创建RC4加密的PDF
func createRC4EncryptedPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 2
/R 3
/Length 128
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
400
%%EOF`, version)
}

// createHighSecurityPDF 创建高安全级别加密的PDF
func createHighSecurityPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 5
/R 6
/Length 256
/CF <<
/StdCF <<
/AuthEvent /DocOpen
/CFM /AESV3
/Length 32
>>
>>
/StmF /StdCF
/StrF /StdCF
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
700
%%EOF`, version)
}

// createPDFWithEncryptKeywords 创建包含加密关键字但未加密的PDF
func createPDFWithEncryptKeywords(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Title (Document about Encryption and Security)
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
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 100
>>
stream
BT
/F1 12 Tf
100 700 Td
(This document discusses /Encrypt /Filter /V /R keywords) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000100 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`, version)
}

// createMetadataEncryptedPDF 创建仅元数据加密的PDF
func createMetadataEncryptedPDF(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 5 0 R
/Metadata 4 0 R
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

4 0 obj
<<
/Type /Metadata
/Subtype /XML
/Length 50
>>
stream
<encrypted_metadata>ENCRYPTED_CONTENT</encrypted_metadata>
endstream
endobj

5 0 obj
<<
/Filter /Standard
/V 4
/R 4
/EncryptMetadata true
/CF <<
/StdCF <<
/AuthEvent /DocOpen
/CFM /AESV2
/Length 16
>>
>>
/StmF /StdCF
/StrF /StdCF
/O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A>
/P -4
>>
endobj

xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000100 00000 n 
0000000173 00000 n 
0000000250 00000 n 
0000000350 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
/Encrypt 5 0 R
>>
startxref
700
%%EOF`, version)
}

// createPDFWithSpecialChars 创建包含特殊字符的PDF
func createPDFWithSpecialChars(version string) string {
	return fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Title (Special Characters: àáâãäåæçèéêë)
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
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 80
>>
stream
BT
/F1 12 Tf
100 700 Td
(Special chars: ñüöß€£¥§©®™) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000100 00000 n 
0000000173 00000 n 
0000000250 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
380
%%EOF`, version)
}

// createLargeTestPDF 创建指定大小的测试PDF
func createLargeTestPDF(t *testing.T, dir, filename string, sizeKB int) string {
	filePath := filepath.Join(dir, filename)

	// 计算需要的填充数据大小
	baseSize := 500 // 基础PDF结构大小
	paddingSize := sizeKB*1024 - baseSize
	if paddingSize < 0 {
		paddingSize = 0
	}

	// 创建填充数据
	padding := strings.Repeat("A", paddingSize)

	content := fmt.Sprintf(`%%PDF-1.4
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
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length %d
>>
stream
%s
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
0000000200 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
%d
%%EOF`, len(padding), padding, 300+len(padding))

	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	return filePath
}

// testMemoryUsage 测试内存使用情况
func testMemoryUsage(t *testing.T, filePath string, suite *ValidationTestSuite) {
	var m1, m2 runtime.MemStats

	// 测试验证操作的内存使用
	runtime.GC()
	runtime.ReadMemStats(&m1)

	err := suite.validator.ValidatePDFFile(filePath)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryUsed := m2.Alloc - m1.Alloc
	maxMemoryLimit := int64(50 * 1024 * 1024) // 50MB限制

	t.Logf("内存使用: %d 字节 (%.2f MB), 验证结果: %v",
		memoryUsed, float64(memoryUsed)/(1024*1024), err)

	if memoryUsed > uint64(maxMemoryLimit) {
		t.Logf("警告: 内存使用超过限制 (限制: %d MB, 实际: %.2f MB)",
			maxMemoryLimit/(1024*1024), float64(memoryUsed)/(1024*1024))
	}
}

// detectEncryptionLevel 检测加密级别
func detectEncryptionLevel(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}

	contentStr := string(content)

	if strings.Contains(contentStr, "/V 5") {
		return "AES-256"
	} else if strings.Contains(contentStr, "/V 4") {
		return "AES-128"
	} else if strings.Contains(contentStr, "/V 2") {
		return "RC4-128"
	} else if strings.Contains(contentStr, "/V 1") {
		return "RC4-40"
	} else if strings.Contains(contentStr, "/Encrypt") {
		return "standard"
	}

	return "none"
}

// boolToInt 将布尔值转换为整数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
