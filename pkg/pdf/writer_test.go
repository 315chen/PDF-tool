package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDFWriterBasic 测试PDF写入器基本功能
func TestPDFWriterBasic(t *testing.T) {
	// 创建测试目录
	testDir := filepath.Join(os.TempDir(), "writer_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "test_output.pdf")

	// 创建写入器选项
	options := &WriterOptions{
		MaxRetries:        3,
		RetryDelay:        time.Second * 1,
		BackupEnabled:     true,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
		EncryptUsingAES:   true,
		EncryptKeyLength:  256,
	}

	// 创建PDF写入器
	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()

	// 测试打开写入器
	t.Run("TestOpen", func(t *testing.T) {
		err := writer.Open()
		assert.NoError(t, err)
		assert.True(t, writer.IsOpen())
	})

	// 测试添加内容
	t.Run("TestAddContent", func(t *testing.T) {
		// 使用有效的PDF内容
		pdfContent := []byte(`%PDF-1.4
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
/Length 44
>>
stream
BT
/F1 12 Tf
72 720 Td
(Test PDF content) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`)
		err := writer.AddContent(pdfContent)
		assert.NoError(t, err)
	})

	// 测试写入文件
	t.Run("TestWrite", func(t *testing.T) {
		result, err := writer.Write(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Greater(t, result.FileSize, int64(0))
		assert.Greater(t, result.WriteTime, time.Duration(0))
		assert.Equal(t, 0, result.RetryCount)

		// 验证输出文件存在
		assert.FileExists(t, outputPath)
	})

	// 测试获取器方法
	t.Run("TestGetters", func(t *testing.T) {
		assert.Equal(t, outputPath, writer.GetOutputPath())
		assert.NotEmpty(t, writer.GetTempPath())
		assert.NotNil(t, writer.GetAdapter())
		assert.NotNil(t, writer.GetConfig())
	})
}

// TestPDFWriterOptions 测试PDF写入器选项配置
func TestPDFWriterOptions(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_options_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	t.Run("TestDefaultOptions", func(t *testing.T) {
		outputPath := filepath.Join(testDir, "default_options.pdf")
		writer, err := NewPDFWriter(outputPath, nil)
		require.NoError(t, err)
		defer writer.Close()

		config := writer.GetConfig()
		assert.Equal(t, "relaxed", config.ValidationMode)
		assert.True(t, config.WriteObjectStream)
		assert.True(t, config.WriteXRefStream)
		assert.True(t, config.EncryptUsingAES)
		assert.Equal(t, 256, config.EncryptKeyLength)
	})

	t.Run("TestCustomOptions", func(t *testing.T) {
		outputPath := filepath.Join(testDir, "custom_options.pdf")
		options := &WriterOptions{
			MaxRetries:        5,
			RetryDelay:        time.Second * 3,
			BackupEnabled:     false,
			TempDirectory:     testDir,
			ValidationMode:    "strict",
			WriteObjectStream: false,
			WriteXRefStream:   false,
			EncryptUsingAES:   false,
			EncryptKeyLength:  128,
		}

		writer, err := NewPDFWriter(outputPath, options)
		require.NoError(t, err)
		defer writer.Close()

		config := writer.GetConfig()
		assert.Equal(t, "strict", config.ValidationMode)
		assert.False(t, config.WriteObjectStream)
		assert.False(t, config.WriteXRefStream)
		assert.False(t, config.EncryptUsingAES)
		assert.Equal(t, 128, config.EncryptKeyLength)
	})
}

// TestPDFWriterErrorHandling 测试PDF写入器错误处理
func TestPDFWriterErrorHandling(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_error_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	t.Run("TestInvalidOutputPath", func(t *testing.T) {
		// 测试无效的输出路径
		_, err := NewPDFWriter("", nil)
		assert.Error(t, err)
		assert.IsType(t, &PDFError{}, err)

		_, err = NewPDFWriter("test.txt", nil)
		assert.Error(t, err)
		assert.IsType(t, &PDFError{}, err)
	})

	t.Run("TestWriteWithoutOpen", func(t *testing.T) {
		outputPath := filepath.Join(testDir, "write_without_open.pdf")
		writer, err := NewPDFWriter(outputPath, nil)
		require.NoError(t, err)
		defer writer.Close()

		// 尝试在未打开状态下写入
		_, err = writer.Write(context.Background(), nil)
		assert.Error(t, err)
		assert.IsType(t, &PDFError{}, err)
	})

	t.Run("TestAddContentWithoutOpen", func(t *testing.T) {
		outputPath := filepath.Join(testDir, "add_content_without_open.pdf")
		writer, err := NewPDFWriter(outputPath, nil)
		require.NoError(t, err)
		defer writer.Close()

		// 尝试在未打开状态下添加内容
		err = writer.AddContent([]byte("test"))
		assert.Error(t, err)
		assert.IsType(t, &PDFError{}, err)
	})
}

// 创建有效的PDF内容函数
func createWriterTestPDFContent(text string) []byte {
	return []byte(fmt.Sprintf(`%%PDF-1.4
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
BT
/F1 12 Tf
72 720 Td
(%s) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`, len(text)+44, text))
}

// TestPDFWriterBackup 测试PDF写入器备份功能
func TestPDFWriterBackup(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_backup_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "backup_test.pdf")

	// 创建初始文件
	initialContent := createWriterTestPDFContent("Initial content")
	err := os.WriteFile(outputPath, initialContent, 0644)
	require.NoError(t, err)

	t.Run("TestBackupEnabled", func(t *testing.T) {
		options := &WriterOptions{
			BackupEnabled: true,
			TempDirectory: testDir,
		}

		writer, err := NewPDFWriter(outputPath, options)
		require.NoError(t, err)
		defer writer.Close()

		err = writer.Open()
		require.NoError(t, err)

		err = writer.AddContent(createWriterTestPDFContent("New content"))
		require.NoError(t, err)

		result, err := writer.Write(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.BackupPath)

		// 验证备份文件存在
		assert.FileExists(t, result.BackupPath)
	})

	t.Run("TestBackupDisabled", func(t *testing.T) {
		options := &WriterOptions{
			BackupEnabled: false,
			TempDirectory: testDir,
		}

		writer, err := NewPDFWriter(outputPath, options)
		require.NoError(t, err)
		defer writer.Close()

		err = writer.Open()
		require.NoError(t, err)

		err = writer.AddContent(createWriterTestPDFContent("Another content"))
		require.NoError(t, err)

		result, err := writer.Write(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Empty(t, result.BackupPath)
	})
}

// TestPDFWriterRetry 测试PDF写入器重试机制
func TestPDFWriterRetry(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_retry_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "retry_test.pdf")

	options := &WriterOptions{
		MaxRetries:    2,
		RetryDelay:    time.Millisecond * 100,
		BackupEnabled: false,
		TempDirectory: testDir,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()

	err = writer.Open()
	require.NoError(t, err)

	err = writer.AddContent(createWriterTestPDFContent("Retry test content"))
	require.NoError(t, err)

	result, err := writer.Write(context.Background(), nil)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.RetryCount) // 正常情况下不需要重试
}

// TestPDFWriterValidation 测试PDF写入器验证功能
func TestPDFWriterValidation(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_validation_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "validation_test.pdf")

	options := &WriterOptions{
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
		TempDirectory:     testDir,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()

	err = writer.Open()
	require.NoError(t, err)

	// 测试创建基本PDF
	t.Run("TestCreateBasicPDF", func(t *testing.T) {
		result, err := writer.Write(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Greater(t, result.FileSize, int64(0))

		// 验证生成的PDF文件
		assert.FileExists(t, outputPath)

		// 验证PDF格式
		validator := NewPDFValidator()
		err = validator.ValidatePDFFile(outputPath)
		assert.NoError(t, err)
	})
}

// TestPDFWriterConcurrent 测试PDF写入器并发安全性
func TestPDFWriterConcurrent(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_concurrent_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	const numWriters = 5
	var wg sync.WaitGroup
	errors := make(chan error, numWriters)

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			outputPath := filepath.Join(testDir, fmt.Sprintf("concurrent_%d.pdf", index))
			writer, err := NewPDFWriter(outputPath, nil)
			if err != nil {
				errors <- err
				return
			}
			defer writer.Close()

			err = writer.Open()
			if err != nil {
				errors <- err
				return
			}

			err = writer.AddContent(createWriterTestPDFContent(fmt.Sprintf("Content from writer %d", index)))
			if err != nil {
				errors <- err
				return
			}

			result, err := writer.Write(context.Background(), nil)
			if err != nil {
				errors <- err
				return
			}

			if !result.Success {
				errors <- fmt.Errorf("writer %d failed", index)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		assert.NoError(t, err)
	}
}

// TestPDFWriterPerformance 测试PDF写入器性能
func TestPDFWriterPerformance(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_performance_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	// 允许通过环境变量控制测试规模
	pageCount := 100
	if v := os.Getenv("PDFWRITER_PERF_PAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageCount = n
		}
	}

	t.Run("TestSmallFile", func(t *testing.T) {
		testPerformanceWithSize(t, testDir, 1, "小文件性能测试")
	})

	t.Run("TestMediumFile", func(t *testing.T) {
		testPerformanceWithSize(t, testDir, 10, "中等文件性能测试")
	})

	t.Run("TestLargeFile", func(t *testing.T) {
		testPerformanceWithSize(t, testDir, pageCount, "大文件性能测试")
	})

	t.Run("TestConcurrentWrites", func(t *testing.T) {
		testConcurrentPerformance(t, testDir, 5, "并发写入性能测试")
	})
}

// testPerformanceWithSize 测试指定大小的文件写入性能
func testPerformanceWithSize(t *testing.T, testDir string, pageCount int, testName string) {
	outputPath := filepath.Join(testDir, fmt.Sprintf("perf_%s.pdf", testName))

	options := &WriterOptions{
		MaxRetries:        1,
		RetryDelay:        time.Millisecond * 50,
		BackupEnabled:     false,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()

	err = writer.Open()
	require.NoError(t, err)

	// 创建指定页数的PDF内容
	content := createMultiPagePDFContent(pageCount, fmt.Sprintf("%s content", testName))

	// 记录内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialAlloc := m.Alloc

	err = writer.AddContent(content)
	require.NoError(t, err)

	start := time.Now()
	result, err := writer.Write(context.Background(), nil)
	duration := time.Since(start)

	// 再次记录内存使用情况
	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Greater(t, result.FileSize, int64(0))

	t.Logf("%s: 页数=%d, 文件大小=%d bytes, 写入时间=%v, 验证时间=%v",
		testName, pageCount, result.FileSize, result.WriteTime, result.ValidationTime)
	t.Logf("总耗时: %v, 内存使用: %d bytes", duration, finalAlloc-initialAlloc)
}

// testConcurrentPerformance 测试并发写入性能
func testConcurrentPerformance(t *testing.T, testDir string, numWriters int, testName string) {
	var wg sync.WaitGroup
	errors := make(chan error, numWriters)
	results := make(chan *WriteResult, numWriters)

	start := time.Now()

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			outputPath := filepath.Join(testDir, fmt.Sprintf("concurrent_perf_%d.pdf", index))
			options := &WriterOptions{
				MaxRetries:        1,
				RetryDelay:        time.Millisecond * 10,
				BackupEnabled:     false,
				TempDirectory:     testDir,
				ValidationMode:    "relaxed",
				WriteObjectStream: true,
				WriteXRefStream:   true,
			}

			writer, err := NewPDFWriter(outputPath, options)
			if err != nil {
				errors <- err
				return
			}
			defer writer.Close()

			err = writer.Open()
			if err != nil {
				errors <- err
				return
			}

			content := createWriterTestPDFContent(fmt.Sprintf("Concurrent content %d", index))
			err = writer.AddContent(content)
			if err != nil {
				errors <- err
				return
			}

			result, err := writer.Write(context.Background(), nil)
			if err != nil {
				errors <- err
				return
			}

			if result.Success {
				results <- result
			} else {
				errors <- fmt.Errorf("writer %d failed", index)
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	duration := time.Since(start)

	// 检查错误
	for err := range errors {
		assert.NoError(t, err)
	}

	// 统计结果
	successCount := 0
	totalSize := int64(0)
	for result := range results {
		successCount++
		totalSize += result.FileSize
	}

	t.Logf("%s: 并发数=%d, 成功数=%d, 总文件大小=%d bytes, 总耗时=%v",
		testName, numWriters, successCount, totalSize, duration)
	assert.Equal(t, numWriters, successCount)
}

// createMultiPagePDFContent 创建多页PDF内容
func createMultiPagePDFContent(pageCount int, text string) []byte {
	if pageCount <= 0 {
		pageCount = 1
	}

	// 创建基本的单页PDF内容
	baseContent := createWriterTestPDFContent(text)

	// 对于多页，我们创建包含多个页面对象的PDF
	if pageCount == 1 {
		return baseContent
	}

	// 创建多页PDF内容（简化版本）
	multiPageContent := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [%s]
/Count %d
>>
endobj`, generatePageRefs(pageCount), pageCount)

	// 添加页面对象
	for i := 1; i <= pageCount; i++ {
		multiPageContent += fmt.Sprintf(`
%d 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents %d 0 R
>>
endobj`, i+2, i+pageCount+2)
	}

	// 添加内容流
	for i := 1; i <= pageCount; i++ {
		content := fmt.Sprintf("BT\n/F1 12 Tf\n72 %d Td\n(Page %d: %s) Tj\nET", 720-i*50, i, text)
		multiPageContent += fmt.Sprintf(`
%d 0 obj
<<
/Length %d
>>
stream
%s
endstream
endobj`, i+pageCount+2, len(content), content)
	}

	// 添加交叉引用表和尾部
	multiPageContent += fmt.Sprintf(`
xref
0 %d
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
%s
trailer
<<
/Size %d
/Root 1 0 R
>>
startxref
%d
%%EOF`, pageCount*2+3, generateXrefEntries(pageCount), pageCount*2+3, len(multiPageContent)-6)

	return []byte(multiPageContent)
}

// generatePageRefs 生成页面引用
func generatePageRefs(pageCount int) string {
	refs := make([]string, pageCount)
	for i := 0; i < pageCount; i++ {
		refs[i] = fmt.Sprintf("%d 0 R", i+3)
	}
	return strings.Join(refs, " ")
}

// generateXrefEntries 生成交叉引用条目
func generateXrefEntries(pageCount int) string {
	var entries []string
	entries = append(entries, "0000000010 00000 n ")
	entries = append(entries, "0000000079 00000 n ")

	// 添加页面对象引用
	for i := 1; i <= pageCount; i++ {
		entries = append(entries, fmt.Sprintf("000000%04d 00000 n ", 1000+i*100))
	}

	// 添加内容流引用
	for i := 1; i <= pageCount; i++ {
		entries = append(entries, fmt.Sprintf("000000%04d 00000 n ", 2000+i*100))
	}

	return strings.Join(entries, "\n")
}

// BenchmarkPDFWriter 基准测试PDF写入器性能
func BenchmarkPDFWriter(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "writer_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	options := &WriterOptions{
		MaxRetries:        1,
		RetryDelay:        time.Millisecond * 10,
		BackupEnabled:     false,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
	}

	content := createWriterTestPDFContent("Benchmark content")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(testDir, fmt.Sprintf("benchmark_%d.pdf", i))

		writer, err := NewPDFWriter(outputPath, options)
		if err != nil {
			b.Fatalf("Failed to create writer: %v", err)
		}

		err = writer.Open()
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to open writer: %v", err)
		}

		err = writer.AddContent(content)
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to add content: %v", err)
		}

		_, err = writer.Write(context.Background(), nil)
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to write: %v", err)
		}

		writer.Close()
	}
}

// BenchmarkPDFWriterConcurrent 并发写入基准测试
func BenchmarkPDFWriterConcurrent(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "writer_concurrent_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// 允许通过环境变量控制并发数
	goroutines := 4
	if v := os.Getenv("PDFWRITER_BENCH_CONCURRENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			goroutines = n
		}
	}
	b.Logf("Benchmarking with %d goroutines", goroutines)

	options := &WriterOptions{
		MaxRetries:        1,
		RetryDelay:        time.Millisecond * 10,
		BackupEnabled:     false,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
	}

	content := createWriterTestPDFContent("Concurrent benchmark content")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		errors := make(chan error, goroutines)

		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				outputPath := filepath.Join(testDir, fmt.Sprintf("concurrent_bench_%d_%d.pdf", i, index))
				writer, err := NewPDFWriter(outputPath, options)
				if err != nil {
					errors <- err
					return
				}
				defer writer.Close()

				err = writer.Open()
				if err != nil {
					errors <- err
					return
				}

				err = writer.AddContent(content)
				if err != nil {
					errors <- err
					return
				}

				_, err = writer.Write(context.Background(), nil)
				if err != nil {
					errors <- err
					return
				}
			}(g)
		}

		wg.Wait()
		close(errors)

		// 检查错误
		for err := range errors {
			b.Fatalf("Concurrent write failed: %v", err)
		}
	}
}

// BenchmarkPDFWriterMemory 内存使用基准测试
func BenchmarkPDFWriterMemory(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "writer_memory_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	options := &WriterOptions{
		MaxRetries:        1,
		RetryDelay:        time.Millisecond * 10,
		BackupEnabled:     false,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
	}

	// 测试不同大小的内容
	sizes := []int{1, 10, 50, 100}

	for _, pageCount := range sizes {
		b.Run(fmt.Sprintf("Pages_%d", pageCount), func(b *testing.B) {
			content := createMultiPagePDFContent(pageCount, "Memory benchmark content")

			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			initialAlloc := m.Alloc

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				outputPath := filepath.Join(testDir, fmt.Sprintf("memory_bench_%d_%d.pdf", pageCount, i))

				writer, err := NewPDFWriter(outputPath, options)
				if err != nil {
					b.Fatalf("Failed to create writer: %v", err)
				}

				err = writer.Open()
				if err != nil {
					writer.Close()
					b.Fatalf("Failed to open writer: %v", err)
				}

				err = writer.AddContent(content)
				if err != nil {
					writer.Close()
					b.Fatalf("Failed to add content: %v", err)
				}

				_, err = writer.Write(context.Background(), nil)
				if err != nil {
					writer.Close()
					b.Fatalf("Failed to write: %v", err)
				}

				writer.Close()
			}

			runtime.ReadMemStats(&m)
			finalAlloc := m.Alloc
			b.ReportMetric(float64(finalAlloc-initialAlloc), "bytes/op")
		})
	}
}

// BenchmarkPDFWriterLarge 大文件写入基准测试
func BenchmarkPDFWriterLarge(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "writer_large_benchmark")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// 允许通过环境变量控制测试规模
	pageCount := 100
	if v := os.Getenv("PDFWRITER_BENCH_PAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageCount = n
		}
	}
	b.Logf("Benchmarking with %d pages", pageCount)

	options := &WriterOptions{
		MaxRetries:        1,
		RetryDelay:        time.Millisecond * 10,
		BackupEnabled:     false,
		TempDirectory:     testDir,
		ValidationMode:    "relaxed",
		WriteObjectStream: true,
		WriteXRefStream:   true,
	}

	content := createMultiPagePDFContent(pageCount, "Large benchmark content")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(testDir, fmt.Sprintf("large_benchmark_%d.pdf", i))

		writer, err := NewPDFWriter(outputPath, options)
		if err != nil {
			b.Fatalf("Failed to create writer: %v", err)
		}

		err = writer.Open()
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to open writer: %v", err)
		}

		err = writer.AddContent(content)
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to add large content: %v", err)
		}

		_, err = writer.Write(context.Background(), nil)
		if err != nil {
			writer.Close()
			b.Fatalf("Failed to write large file: %v", err)
		}

		writer.Close()
	}
}

// TestPDFWriterRetry_ExponentialBackoff 测试指数退避重试机制
func TestPDFWriterRetry_ExponentialBackoff(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_retry_backoff_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "backoff_test.pdf")

	failCount := 2
	var mu sync.Mutex

	// hook: 模拟前两次写入失败
	origWriteToTempFile := writeToTempFile
	writeToTempFile = func(w *PDFWriter) error {
		mu.Lock()
		defer mu.Unlock()
		if failCount > 0 {
			failCount--
			return &PDFError{Type: ErrorIO, Message: "模拟IO错误"}
		}
		return origWriteToTempFile(w)
	}
	defer func() { writeToTempFile = origWriteToTempFile }()

	options := &WriterOptions{
		MaxRetries:        3,
		InitialRetryDelay: time.Millisecond * 100,
		MaxRetryDelay:     time.Millisecond * 500,
		BackoffFactor:     2.0,
		BackupEnabled:     false,
		TempDirectory:     testDir,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()
	require.NoError(t, writer.Open())
	require.NoError(t, writer.AddContent(createWriterTestPDFContent("backoff test")))

	ctx := context.Background()
	start := time.Now()
	result, err := writer.Write(ctx, nil)
	dur := time.Since(start)
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.RetryCount)
	assert.GreaterOrEqual(t, dur, time.Millisecond*100+time.Millisecond*200) // 至少两次递增延迟
}

// TestPDFWriterRetry_ContextCancel 测试写入过程中取消
func TestPDFWriterRetry_ContextCancel(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_retry_cancel_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "cancel_test.pdf")

	failCount := 5
	var mu sync.Mutex
	origWriteToTempFile := writeToTempFile
	writeToTempFile = func(w *PDFWriter) error {
		mu.Lock()
		defer mu.Unlock()
		if failCount > 0 {
			failCount--
			return &PDFError{Type: ErrorIO, Message: "模拟IO错误"}
		}
		return origWriteToTempFile(w)
	}
	defer func() { writeToTempFile = origWriteToTempFile }()

	options := &WriterOptions{
		MaxRetries:        10,
		InitialRetryDelay: time.Millisecond * 100,
		MaxRetryDelay:     time.Millisecond * 500,
		BackoffFactor:     2.0,
		BackupEnabled:     false,
		TempDirectory:     testDir,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()
	require.NoError(t, writer.Open())
	require.NoError(t, writer.AddContent(createWriterTestPDFContent("cancel test")))

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*250)
	defer cancel()
	start := time.Now()
	result, err := writer.Write(ctx, nil)
	dur := time.Since(start)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline")
	assert.False(t, result.Success)
	assert.GreaterOrEqual(t, dur, time.Millisecond*100) // 至少有一次重试
}

// TestPDFWriterRetry_ErrorType 测试不可恢复错误不重试
func TestPDFWriterRetry_ErrorType(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "writer_retry_errtype_test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	defer os.RemoveAll(testDir)

	outputPath := filepath.Join(testDir, "errtype_test.pdf")

	// hook: 第一次返回权限错误
	origWriteToTempFile := writeToTempFile
	writeToTempFile = func(w *PDFWriter) error {
		return &PDFError{Type: ErrorPermission, Message: "模拟权限错误"}
	}
	defer func() { writeToTempFile = origWriteToTempFile }()

	options := &WriterOptions{
		MaxRetries:        3,
		InitialRetryDelay: time.Millisecond * 100,
		MaxRetryDelay:     time.Millisecond * 500,
		BackoffFactor:     2.0,
		BackupEnabled:     false,
		TempDirectory:     testDir,
	}

	writer, err := NewPDFWriter(outputPath, options)
	require.NoError(t, err)
	defer writer.Close()
	require.NoError(t, writer.Open())
	defer writer.Close()
	_ = writer.AddContent(createWriterTestPDFContent("errtype test"))

	ctx := context.Background()
	result, err := writer.Write(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "权限错误")
	assert.False(t, result.Success)
	assert.Equal(t, 0, result.RetryCount) // 不应重试
}
