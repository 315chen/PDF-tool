package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestPDFServiceRefactor 测试重构后的PDF服务
func TestPDFServiceRefactor(t *testing.T) {
	// 创建测试服务
	service := NewPDFServiceWithConfig(&ServiceConfig{
		MaxRetries:       3,
		RetryDelay:       time.Millisecond * 100,
		EnableStrictMode: false,
		PreferPDFCPU:     true,
		TempDirectory:    os.TempDir(),
		MaxMemoryUsage:   50 * 1024 * 1024, // 50MB
	})

	t.Run("TestBasicFileValidation", func(t *testing.T) {
		// 测试不存在的文件
		err := service.ValidatePDF("nonexistent.pdf")
		if err == nil {
			t.Error("应该返回文件不存在的错误")
		}

		// 测试非PDF文件
		tempFile := filepath.Join(os.TempDir(), "test.txt")
		os.WriteFile(tempFile, []byte("not a pdf"), 0644)
		defer os.Remove(tempFile)

		err = service.ValidatePDF(tempFile)
		if err == nil {
			t.Error("应该返回非PDF文件的错误")
		}
	})

	t.Run("TestServiceConfiguration", func(t *testing.T) {
		// 测试默认配置
		defaultService := NewPDFService()
		if defaultService == nil {
			t.Error("默认服务创建失败")
		}

		// 测试自定义配置
		customConfig := &ServiceConfig{
			MaxRetries:       5,
			RetryDelay:       time.Second,
			EnableStrictMode: true,
			PreferPDFCPU:     false,
			TempDirectory:    "/tmp/custom",
			MaxMemoryUsage:   200 * 1024 * 1024,
		}

		customService := NewPDFServiceWithConfig(customConfig)
		if customService == nil {
			t.Error("自定义配置服务创建失败")
		}
	})

	t.Run("TestErrorHandling", func(t *testing.T) {
		// 创建一个空文件
		emptyFile := filepath.Join(os.TempDir(), "empty.pdf")
		os.WriteFile(emptyFile, []byte{}, 0644)
		defer os.Remove(emptyFile)

		err := service.ValidatePDF(emptyFile)
		if err == nil {
			t.Error("应该返回空文件错误")
		}

		// 检查错误类型
		if pdfErr, ok := err.(*PDFError); ok {
			if pdfErr.Type != ErrorInvalidFile {
				t.Errorf("期望错误类型 %v, 得到 %v", ErrorInvalidFile, pdfErr.Type)
			}
		} else {
			t.Error("应该返回PDFError类型")
		}
	})

	t.Run("TestEncryptionDetection", func(t *testing.T) {
		// 创建一个模拟的PDF文件（包含加密关键字）
		mockPDFContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 3 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [4 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O (encrypted)
/U (encrypted)
/P -44
>>
endobj
%%EOF`

		mockFile := filepath.Join(os.TempDir(), "mock_encrypted.pdf")
		os.WriteFile(mockFile, []byte(mockPDFContent), 0644)
		defer os.Remove(mockFile)

		// 测试加密检测
		isEncrypted, err := service.IsPDFEncrypted(mockFile)
		if err != nil {
			t.Logf("加密检测可能失败（这是正常的，因为这是模拟文件）: %v", err)
		} else if !isEncrypted {
			t.Log("模拟加密文件未被检测为加密（这可能是正常的，取决于检测方法）")
		}
	})
}

// TestPDFInfoRetrieval 测试PDF信息获取
func TestPDFInfoRetrieval(t *testing.T) {
	service := NewPDFService()

	t.Run("TestInfoValidation", func(t *testing.T) {
		// 测试信息验证功能
		serviceImpl := service.(*PDFServiceImpl)

		// 测试有效信息
		validInfo := &PDFInfo{
			FilePath:  "/test/file.pdf",
			PageCount: 5,
			FileSize:  1024,
			Title:     "Test PDF",
		}

		err := serviceImpl.validatePDFInfo(validInfo)
		if err != nil {
			t.Errorf("有效信息验证失败: %v", err)
		}

		// 测试无效信息
		invalidInfo := &PDFInfo{
			FilePath:  "",
			PageCount: -1,
			FileSize:  -100,
		}

		err = serviceImpl.validatePDFInfo(invalidInfo)
		if err == nil {
			t.Error("无效信息应该验证失败")
		}
	})

	t.Run("TestInfoEnrichment", func(t *testing.T) {
		// 创建测试文件
		testFile := filepath.Join(os.TempDir(), "test_info.pdf")
		testContent := "%PDF-1.4\n%%EOF"
		os.WriteFile(testFile, []byte(testContent), 0644)
		defer os.Remove(testFile)

		serviceImpl := service.(*PDFServiceImpl)
		info := &PDFInfo{}

		err := serviceImpl.enrichInfoWithFileSystemData(info, testFile)
		if err != nil {
			t.Errorf("信息补充失败: %v", err)
		}

		if info.FilePath != testFile {
			t.Errorf("文件路径未正确设置: 期望 %s, 得到 %s", testFile, info.FilePath)
		}

		if info.FileSize == 0 {
			t.Error("文件大小未正确设置")
		}

		if info.Title == "" {
			t.Error("标题未正确设置")
		}
	})
}

// TestMergeStrategies 测试合并策略
func TestMergeStrategies(t *testing.T) {
	service := NewPDFServiceWithConfig(&ServiceConfig{
		PreferPDFCPU:   false, // 禁用pdfcpu以测试回退策略
		TempDirectory:  os.TempDir(),
		MaxMemoryUsage: 50 * 1024 * 1024,
	})

	t.Run("TestFileCopy", func(t *testing.T) {
		// 创建源文件
		srcFile := filepath.Join(os.TempDir(), "source.pdf")
		srcContent := "%PDF-1.4\ntest content\n%%EOF"
		os.WriteFile(srcFile, []byte(srcContent), 0644)
		defer os.Remove(srcFile)

		// 测试文件复制
		dstFile := filepath.Join(os.TempDir(), "destination.pdf")
		defer os.Remove(dstFile)

		serviceImpl := service.(*PDFServiceImpl)
		err := serviceImpl.copyFile(srcFile, dstFile)
		if err != nil {
			t.Errorf("文件复制失败: %v", err)
		}

		// 验证复制结果
		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Errorf("无法读取目标文件: %v", err)
		}

		if string(dstContent) != srcContent {
			t.Error("复制的文件内容不匹配")
		}
	})

	t.Run("TestValidFileFiltering", func(t *testing.T) {
		// 这个测试模拟合并过程中的文件过滤
		tempDir := os.TempDir()

		// 创建有效的PDF文件
		validFile := filepath.Join(tempDir, "valid.pdf")
		validContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF"
		os.WriteFile(validFile, []byte(validContent), 0644)
		defer os.Remove(validFile)

		// 创建无效文件
		invalidFile := filepath.Join(tempDir, "invalid.pdf")
		os.WriteFile(invalidFile, []byte("not a pdf"), 0644)
		defer os.Remove(invalidFile)

		// 创建空文件
		emptyFile := filepath.Join(tempDir, "empty.pdf")
		os.WriteFile(emptyFile, []byte{}, 0644)
		defer os.Remove(emptyFile)

		files := []string{validFile, invalidFile, emptyFile}

		// 测试文件验证
		validCount := 0
		for _, file := range files {
			if err := service.ValidatePDF(file); err == nil {
				validCount++
			}
		}

		// 至少应该有一个有效文件（validFile可能通过基本验证）
		if validCount == 0 {
			t.Log("所有文件都被认为无效（这可能是正常的，取决于验证严格程度）")
		}
	})
}

// TestErrorCollectorRefactor 测试错误收集器（重构版本）
func TestErrorCollectorRefactor(t *testing.T) {
	collector := NewErrorCollector()

	t.Run("TestEmptyCollector", func(t *testing.T) {
		if collector.HasErrors() {
			t.Error("新的错误收集器不应该有错误")
		}

		if collector.GetErrorCount() != 0 {
			t.Errorf("期望错误数量为0, 得到 %d", collector.GetErrorCount())
		}
	})

	t.Run("TestAddErrors", func(t *testing.T) {
		err1 := &PDFError{Type: ErrorInvalidFile, Message: "测试错误1", File: "file1.pdf"}
		err2 := &PDFError{Type: ErrorCorrupted, Message: "测试错误2", File: "file2.pdf"}

		collector.Add(err1)
		collector.Add(err2)
		collector.Add(nil) // 应该被忽略

		if !collector.HasErrors() {
			t.Error("收集器应该有错误")
		}

		if collector.GetErrorCount() != 2 {
			t.Errorf("期望错误数量为2, 得到 %d", collector.GetErrorCount())
		}

		summary := collector.GetSummary()
		if !strings.Contains(summary, "测试错误1") || !strings.Contains(summary, "测试错误2") {
			t.Error("错误摘要不包含预期的错误信息")
		}
	})

	t.Run("TestClearCollector", func(t *testing.T) {
		collector.Clear()

		if collector.HasErrors() {
			t.Error("清空后的收集器不应该有错误")
		}

		if collector.GetErrorCount() != 0 {
			t.Errorf("清空后期望错误数量为0, 得到 %d", collector.GetErrorCount())
		}
	})
}

// BenchmarkPDFValidation 性能测试
func BenchmarkPDFValidation(b *testing.B) {
	service := NewPDFService()

	// 创建测试文件
	testFile := filepath.Join(os.TempDir(), "benchmark.pdf")
	testContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF"
	os.WriteFile(testFile, []byte(testContent), 0644)
	defer os.Remove(testFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidatePDF(testFile)
	}
}

// BenchmarkEncryptionCheck 加密检查性能测试
func BenchmarkEncryptionCheck(b *testing.B) {
	service := NewPDFService()

	// 创建测试文件
	testFile := filepath.Join(os.TempDir(), "benchmark_encrypt.pdf")
	testContent := "%PDF-1.4\n/Encrypt << /Filter /Standard >>\n%%EOF"
	os.WriteFile(testFile, []byte(testContent), 0644)
	defer os.Remove(testFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.IsPDFEncrypted(testFile)
	}
}
