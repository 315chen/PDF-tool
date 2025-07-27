package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	
	"github.com/user/pdf-merger/internal/model"
)

func TestNewFileListManager(t *testing.T) {
	flm := NewFileListManager()
	
	if flm == nil {
		t.Error("NewFileListManager returned nil")
	}
	
	if flm.files == nil {
		t.Error("Files slice not initialized")
	}
	
	if flm.list == nil {
		t.Error("List widget not created")
	}
	
	if flm.selectedIndex != -1 {
		t.Error("Selected index should be -1 initially")
	}
}

func TestFileListManager_AddFile(t *testing.T) {
	flm := NewFileListManager()
	
	// 测试添加文件
	err := flm.AddFile("/test/file1.pdf")
	if err != nil {
		t.Errorf("AddFile failed: %v", err)
	}
	
	if flm.GetFileCount() != 1 {
		t.Errorf("Expected 1 file, got %d", flm.GetFileCount())
	}
	
	// 测试添加重复文件
	err = flm.AddFile("/test/file1.pdf")
	if err == nil {
		t.Error("Expected error when adding duplicate file")
	}
	
	if flm.GetFileCount() != 1 {
		t.Errorf("Expected 1 file after duplicate add, got %d", flm.GetFileCount())
	}
	
	// 测试添加另一个文件
	err = flm.AddFile("/test/file2.pdf")
	if err != nil {
		t.Errorf("AddFile failed: %v", err)
	}
	
	if flm.GetFileCount() != 2 {
		t.Errorf("Expected 2 files, got %d", flm.GetFileCount())
	}
}

func TestFileListManager_RemoveFile(t *testing.T) {
	flm := NewFileListManager()
	
	// 添加测试文件
	flm.AddFile("/test/file1.pdf")
	flm.AddFile("/test/file2.pdf")
	flm.AddFile("/test/file3.pdf")
	
	if flm.GetFileCount() != 3 {
		t.Errorf("Expected 3 files, got %d", flm.GetFileCount())
	}
	
	// 选择第二个文件并移除
	flm.selectedIndex = 1
	flm.RemoveSelected()
	
	if flm.GetFileCount() != 2 {
		t.Errorf("Expected 2 files after removal, got %d", flm.GetFileCount())
	}
	
	// 检查文件顺序
	files := flm.GetFiles()
	if files[0].Path != "/test/file1.pdf" {
		t.Errorf("Expected first file to be file1.pdf, got %s", files[0].Path)
	}
	if files[1].Path != "/test/file3.pdf" {
		t.Errorf("Expected second file to be file3.pdf, got %s", files[1].Path)
	}
}

func TestFileListManager_MoveFiles(t *testing.T) {
	flm := NewFileListManager()
	
	// 添加测试文件
	flm.AddFile("/test/file1.pdf")
	flm.AddFile("/test/file2.pdf")
	flm.AddFile("/test/file3.pdf")
	
	// 选择第二个文件并上移
	flm.selectedIndex = 1
	flm.MoveSelectedUp()
	
	files := flm.GetFiles()
	if files[0].Path != "/test/file2.pdf" {
		t.Errorf("Expected first file to be file2.pdf after move up, got %s", files[0].Path)
	}
	if files[1].Path != "/test/file1.pdf" {
		t.Errorf("Expected second file to be file1.pdf after move up, got %s", files[1].Path)
	}
	
	// 选择第一个文件并下移
	flm.selectedIndex = 0
	flm.MoveSelectedDown()
	
	files = flm.GetFiles()
	if files[0].Path != "/test/file1.pdf" {
		t.Errorf("Expected first file to be file1.pdf after move down, got %s", files[0].Path)
	}
	if files[1].Path != "/test/file2.pdf" {
		t.Errorf("Expected second file to be file2.pdf after move down, got %s", files[1].Path)
	}
}

func TestFileListManager_Clear(t *testing.T) {
	flm := NewFileListManager()
	
	// 添加测试文件
	flm.AddFile("/test/file1.pdf")
	flm.AddFile("/test/file2.pdf")
	
	if flm.GetFileCount() != 2 {
		t.Errorf("Expected 2 files before clear, got %d", flm.GetFileCount())
	}
	
	// 清空列表
	flm.Clear()
	
	if flm.GetFileCount() != 0 {
		t.Errorf("Expected 0 files after clear, got %d", flm.GetFileCount())
	}
	
	if flm.selectedIndex != -1 {
		t.Errorf("Expected selected index to be -1 after clear, got %d", flm.selectedIndex)
	}
}

func TestFileListManager_GetFilePaths(t *testing.T) {
	flm := NewFileListManager()
	
	// 添加测试文件
	testPaths := []string{"/test/file1.pdf", "/test/file2.pdf", "/test/file3.pdf"}
	for _, path := range testPaths {
		flm.AddFile(path)
	}
	
	paths := flm.GetFilePaths()
	
	if len(paths) != len(testPaths) {
		t.Errorf("Expected %d paths, got %d", len(testPaths), len(paths))
	}
	
	for i, path := range paths {
		if path != testPaths[i] {
			t.Errorf("Expected path %s at index %d, got %s", testPaths[i], i, path)
		}
	}
}

func TestFileListManager_HasFiles(t *testing.T) {
	flm := NewFileListManager()
	
	// 初始状态应该没有文件
	if flm.HasFiles() {
		t.Error("Expected HasFiles to be false initially")
	}
	
	// 添加文件后应该有文件
	flm.AddFile("/test/file1.pdf")
	if !flm.HasFiles() {
		t.Error("Expected HasFiles to be true after adding file")
	}
	
	// 清空后应该没有文件
	flm.Clear()
	if flm.HasFiles() {
		t.Error("Expected HasFiles to be false after clear")
	}
}

func TestFileListManager_GetFileInfo(t *testing.T) {
	flm := NewFileListManager()
	
	// 没有文件时的信息
	info := flm.GetFileInfo()
	if info != "没有文件" {
		t.Errorf("Expected '没有文件', got '%s'", info)
	}
	
	// 添加文件后的信息
	flm.AddFile("/test/file1.pdf")
	info = flm.GetFileInfo()
	
	if !contains(info, "文件: 1个") {
		t.Errorf("Expected file count in info, got: %s", info)
	}
}

func TestFileListManager_Callbacks(t *testing.T) {
	flm := NewFileListManager()
	
	// 测试文件变更回调
	changeCallbackCalled := false
	flm.SetOnFileChanged(func() {
		changeCallbackCalled = true
	})
	
	flm.AddFile("/test/file1.pdf")
	if !changeCallbackCalled {
		t.Error("Expected file changed callback to be called")
	}
	
	// 测试文件信息回调
	infoCallbackCalled := false
	flm.SetOnFileInfo(func(path string) (*model.FileEntry, error) {
		infoCallbackCalled = true
		return &model.FileEntry{
			Path:        path,
			DisplayName: "test.pdf",
			Size:        1024,
			PageCount:   5,
			IsValid:     true,
		}, nil
	})
	
	flm.AddFile("/test/file2.pdf")
	if !infoCallbackCalled {
		t.Error("Expected file info callback to be called")
	}
}

func TestFileListManager_Widget(t *testing.T) {
	flm := NewFileListManager()
	
	widget := flm.GetWidget()
	if widget == nil {
		t.Error("GetWidget returned nil")
	}
	
	// 测试widget是否可以正常使用
	app := test.NewApp()
	window := app.NewWindow("Test")
	window.SetContent(widget)
	
	// 添加文件并检查列表长度
	flm.AddFile("/test/file1.pdf")
	
	// 这里我们无法直接测试widget的内容，但可以确保没有崩溃
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}