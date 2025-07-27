package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStreamingMerger(t *testing.T) {
	// 测试默认选项
	merger := NewStreamingMerger(nil)
	if merger == nil {
		t.Fatal("NewStreamingMerger 返回 nil")
	}

	if merger.maxMemoryUsage != 100*1024*1024 {
		t.Errorf("期望默认内存限制为 100MB，实际为 %d", merger.maxMemoryUsage)
	}

	// 测试自定义选项
	options := &MergeOptions{
		MaxMemoryUsage: 50 * 1024 * 1024,
		TempDirectory:  "/tmp/test",
		EnableGC:       false,
		ChunkSize:      5,
	}

	merger = NewStreamingMerger(options)
	if merger.maxMemoryUsage != 50*1024*1024 {
		t.Errorf("期望内存限制为 50MB，实际为 %d", merger.maxMemoryUsage)
	}

	if merger.tempDir != "/tmp/test" {
		t.Errorf("期望临时目录为 /tmp/test，实际为 %s", merger.tempDir)
	}
}

func TestStreamingMerger_MemoryManagement(t *testing.T) {
	merger := NewStreamingMerger(&MergeOptions{
		MaxMemoryUsage: 1024, // 很小的内存限制用于测试
		TempDirectory:  os.TempDir(),
		EnableGC:       true,
		ChunkSize:      1,
	})

	// 测试内存使用量获取
	memUsage := merger.getCurrentMemoryUsage()
	if memUsage <= 0 {
		t.Error("内存使用量应该大于0")
	}

	t.Logf("当前内存使用量: %d bytes", memUsage)

	// 测试垃圾回收
	initialMem := merger.getCurrentMemoryUsage()

	// 分配一些内存
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	merger.forceGC()

	finalMem := merger.getCurrentMemoryUsage()
	t.Logf("垃圾回收前: %d bytes, 垃圾回收后: %d bytes", initialMem, finalMem)
}

func TestStreamingMerger_ProgressTracking(t *testing.T) {
	merger := NewStreamingMerger(nil)

	// 测试进度跟踪器创建
	if merger.GetProgressTracker() != nil {
		t.Error("初始状态下进度跟踪器应该为 nil")
	}

	// 模拟创建进度跟踪器（通常在 MergeFiles 中创建）
	testDir := filepath.Join(os.TempDir(), "progress_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建空的测试文件
	mainFile := filepath.Join(testDir, "main.pdf")
	outputFile := filepath.Join(testDir, "output.pdf")

	f, err := os.Create(mainFile)
	if err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}
	f.Close()

	// 测试取消功能
	go func() {
		time.Sleep(100 * time.Millisecond)
		merger.Cancel()
	}()

	var progressBuffer bytes.Buffer
	_, err = merger.MergeFilesLegacy(mainFile, []string{}, outputFile, &progressBuffer)

	// 检查是否有进度跟踪器
	if merger.GetProgressTracker() == nil {
		t.Error("合并后应该有进度跟踪器")
	} else {
		progress := merger.GetProgressTracker().GetProgress()
		t.Logf("最终进度状态: 完成=%t, 取消=%t", progress.IsCompleted, progress.IsCancelled)
	}
}

func TestMergeOptions_Validation(t *testing.T) {
	// 测试各种选项组合
	testCases := []struct {
		name    string
		options *MergeOptions
		valid   bool
	}{
		{
			name:    "nil选项",
			options: nil,
			valid:   true, // 应该使用默认值
		},
		{
			name: "正常选项",
			options: &MergeOptions{
				MaxMemoryUsage: 50 * 1024 * 1024,
				TempDirectory:  "/tmp",
				EnableGC:       true,
				ChunkSize:      10,
			},
			valid: true,
		},
		{
			name: "最小内存限制",
			options: &MergeOptions{
				MaxMemoryUsage: 1024, // 1KB
				TempDirectory:  "/tmp",
				EnableGC:       true,
				ChunkSize:      1,
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merger := NewStreamingMerger(tc.options)
			if merger == nil && tc.valid {
				t.Error("期望创建成功但返回了 nil")
			}
			if merger != nil && !tc.valid {
				t.Error("期望创建失败但返回了有效对象")
			}
		})
	}
}

func TestMergeResult_Structure(t *testing.T) {
	// 测试合并结果结构
	result := &MergeResult{
		OutputPath:     "/path/to/output.pdf",
		TotalPages:     100,
		ProcessedFiles: 5,
		SkippedFiles:   []string{"bad1.pdf", "bad2.pdf"},
		ProcessingTime: time.Second * 30,
		MemoryUsage:    50 * 1024 * 1024,
	}

	if result.OutputPath != "/path/to/output.pdf" {
		t.Error("输出路径不匹配")
	}

	if result.TotalPages != 100 {
		t.Error("总页数不匹配")
	}

	if result.ProcessedFiles != 5 {
		t.Error("处理文件数不匹配")
	}

	if len(result.SkippedFiles) != 2 {
		t.Error("跳过文件数不匹配")
	}

	if result.ProcessingTime != time.Second*30 {
		t.Error("处理时间不匹配")
	}

	if result.MemoryUsage != 50*1024*1024 {
		t.Error("内存使用量不匹配")
	}
}

func TestStreamingMerger_ErrorHandling(t *testing.T) {
	merger := NewStreamingMerger(nil)

	// 测试不存在的文件
	testDir := filepath.Join(os.TempDir(), "error_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	nonExistentFile := filepath.Join(testDir, "nonexistent.pdf")
	outputFile := filepath.Join(testDir, "output.pdf")

	var progressBuffer bytes.Buffer
	result, err := merger.MergeFilesLegacy(nonExistentFile, []string{}, outputFile, &progressBuffer)

	// 应该有错误或者文件被跳过
	if err == nil && result != nil {
		if len(result.SkippedFiles) == 0 {
			t.Error("期望文件被跳过或返回错误")
		} else {
			t.Logf("文件被正确跳过: %v", result.SkippedFiles)
		}
	} else if err != nil {
		t.Logf("正确返回错误: %v", err)
	}
}
