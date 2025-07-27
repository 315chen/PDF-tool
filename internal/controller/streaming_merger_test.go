package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pdf-merger/internal/model"
)

func TestStreamingMerger_shouldUseStreaming(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 测试内存使用判断
	shouldUse := streamingMerger.shouldUseStreaming()

	// 这个测试结果取决于当前系统内存使用情况
	t.Logf("是否应该使用流式处理: %v", shouldUse)
}

func TestStreamingMerger_isMemoryHigh(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 测试内存高使用判断
	isHigh := streamingMerger.isMemoryHigh()

	// 这个测试结果取决于当前系统内存使用情况
	t.Logf("内存使用是否过高: %v", isHigh)
}

func TestStreamingMerger_createTempFile(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "streaming-merger-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &extendedMockFileManager{
		mockFileManager: &mockFileManager{},
		tempDir:         tempDir,
	}
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 创建临时文件
	tempFile, err := streamingMerger.createTempFile("test_", ".pdf")
	if err != nil {
		t.Errorf("创建临时文件失败: %v", err)
	}

	// 验证文件被添加到临时文件列表
	if len(streamingMerger.tempFiles) != 1 {
		t.Errorf("期望临时文件列表长度为1，实际为%d", len(streamingMerger.tempFiles))
	}

	if streamingMerger.tempFiles[0] != tempFile {
		t.Errorf("临时文件路径不匹配")
	}
}

func TestStreamingMerger_cleanup(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "streaming-merger-cleanup-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &extendedMockFileManager{
		mockFileManager: &mockFileManager{},
		tempDir:         tempDir,
	}
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 添加一些临时文件到列表
	streamingMerger.tempFiles = []string{"temp1.pdf", "temp2.pdf"}

	// 执行清理
	streamingMerger.cleanup()

	// 验证临时文件列表被清空
	if len(streamingMerger.tempFiles) != 0 {
		t.Errorf("期望临时文件列表为空，实际长度为%d", len(streamingMerger.tempFiles))
	}
}

func TestStreamingMerger_preprocessFile(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "streaming-merger-preprocess-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	testContent := []byte("test pdf content")
	err = os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &extendedMockFileManager{
		mockFileManager: &mockFileManager{},
		tempDir:         tempDir,
	}
	config := model.DefaultConfig()
	config.TempDirectory = tempDir

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 预处理文件
	ctx := context.Background()
	processedFile, err := streamingMerger.preprocessFile(ctx, testFile)
	if err != nil {
		t.Errorf("预处理文件失败: %v", err)
	}

	// 对于小文件，应该返回原文件路径
	if processedFile != testFile {
		t.Errorf("期望返回原文件路径，实际返回: %s", processedFile)
	}
}

func TestStreamingMerger_writePDFHeader(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "streaming-merger-header-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试输出文件
	outputFile := filepath.Join(tempDir, "output.pdf")
	file, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 写入PDF头部
	err = streamingMerger.writePDFHeader(file)
	if err != nil {
		t.Errorf("写入PDF头部失败: %v", err)
	}

	// 验证文件内容
	file.Close()
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}

	expected := "%PDF-1.4\n"
	if string(content) != expected {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", expected, string(content))
	}
}

func TestStreamingMerger_writePDFFooter(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "streaming-merger-footer-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试输出文件
	outputFile := filepath.Join(tempDir, "output.pdf")
	file, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 写入PDF尾部
	err = streamingMerger.writePDFFooter(file)
	if err != nil {
		t.Errorf("写入PDF尾部失败: %v", err)
	}

	// 验证文件内容
	file.Close()
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}

	expected := "%%EOF\n"
	if string(content) != expected {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", expected, string(content))
	}
}

func TestBatchProcessor_createBatches(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 创建批处理器
	batchProcessor := NewBatchProcessor(streamingMerger)
	batchProcessor.batchSize = 3

	// 测试文件列表
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf", "file4.pdf", "file5.pdf", "file6.pdf", "file7.pdf"}

	// 创建批次
	batches := batchProcessor.createBatches(files)

	// 验证批次数量
	expectedBatches := 3 // 3, 3, 1
	if len(batches) != expectedBatches {
		t.Errorf("期望批次数量为%d，实际为%d", expectedBatches, len(batches))
	}

	// 验证第一个批次
	if len(batches[0]) != 3 {
		t.Errorf("期望第一个批次大小为3，实际为%d", len(batches[0]))
	}

	// 验证第二个批次
	if len(batches[1]) != 3 {
		t.Errorf("期望第二个批次大小为3，实际为%d", len(batches[1]))
	}

	// 验证第三个批次
	if len(batches[2]) != 1 {
		t.Errorf("期望第三个批次大小为1，实际为%d", len(batches[2]))
	}

	// 验证文件内容
	expectedFiles := [][]string{
		{"file1.pdf", "file2.pdf", "file3.pdf"},
		{"file4.pdf", "file5.pdf", "file6.pdf"},
		{"file7.pdf"},
	}

	for i, batch := range batches {
		for j, file := range batch {
			if file != expectedFiles[i][j] {
				t.Errorf("批次%d文件%d期望为%s，实际为%s", i, j, expectedFiles[i][j], file)
			}
		}
	}
}

func TestBatchProcessor_ProcessBatch_SmallBatch(t *testing.T) {
	// 创建模拟服务
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	// 创建控制器和流式合并器
	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	// 创建批处理器
	batchProcessor := NewBatchProcessor(streamingMerger)
	batchProcessor.batchSize = 10 // 设置较大的批次大小

	// 测试小批次文件列表
	files := []string{"file1.pdf", "file2.pdf", "file3.pdf"}
	outputPath := "small_batch_output.pdf"

	// 执行批处理
	ctx := context.Background()
	err := batchProcessor.ProcessBatch(ctx, files, outputPath, nil)

	if err != nil {
		t.Errorf("小批次处理失败: %v", err)
	}
}

// 扩展mockFileManager以支持更多功能
type extendedMockFileManager struct {
	*mockFileManager
	tempDir string
}

func (m *extendedMockFileManager) CreateTempFileWithPrefix(prefix string, suffix string) (string, *os.File, error) {
	if m.tempDir == "" {
		m.tempDir = os.TempDir()
	}

	tempFile := filepath.Join(m.tempDir, prefix+"test"+suffix)
	file, err := os.Create(tempFile)
	if err != nil {
		return "", nil, err
	}

	return tempFile, file, nil
}

func (m *extendedMockFileManager) CopyFile(sourcePath, destPath string) error {
	// 简单的文件复制模拟
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, sourceContent, 0644)
}

// 基准测试

func BenchmarkStreamingMerger_shouldUseStreaming(b *testing.B) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		streamingMerger.shouldUseStreaming()
	}
}

func BenchmarkStreamingMerger_isMemoryHigh(b *testing.B) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		streamingMerger.isMemoryHigh()
	}
}

func BenchmarkBatchProcessor_createBatches(b *testing.B) {
	mockPDF := &mockPDFService{}
	mockFile := &mockFileManager{}
	config := model.DefaultConfig()

	controller := NewController(mockPDF, mockFile, config)
	streamingMerger := NewStreamingMerger(controller)
	batchProcessor := NewBatchProcessor(streamingMerger)

	// 创建大量文件用于测试
	files := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		files[i] = fmt.Sprintf("file%d.pdf", i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		batchProcessor.createBatches(files)
	}
}
