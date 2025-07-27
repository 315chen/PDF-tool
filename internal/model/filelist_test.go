package model

import (
	"testing"
)

func TestNewFileList(t *testing.T) {
	fl := NewFileList()

	if fl == nil {
		t.Fatal("Expected non-nil FileList")
	}

	if fl.Count() != 0 {
		t.Errorf("Expected count 0, got %d", fl.Count())
	}

	if !fl.IsEmpty() {
		t.Error("Expected IsEmpty to be true")
	}

	if fl.HasMainFile() {
		t.Error("Expected HasMainFile to be false")
	}
}

func TestFileList_SetMainFile(t *testing.T) {
	fl := NewFileList()
	path := "/path/to/main.pdf"

	entry := fl.SetMainFile(path)

	if entry == nil {
		t.Fatal("Expected non-nil FileEntry")
	}

	if entry.Path != path {
		t.Errorf("Expected Path %s, got %s", path, entry.Path)
	}

	if !fl.HasMainFile() {
		t.Error("Expected HasMainFile to be true")
	}

	mainFile := fl.GetMainFile()
	if mainFile != entry {
		t.Error("Expected GetMainFile to return the same entry")
	}
}

func TestFileList_AddFile(t *testing.T) {
	fl := NewFileList()
	path1 := "/path/to/file1.pdf"
	path2 := "/path/to/file2.pdf"

	// 添加第一个文件
	entry1 := fl.AddFile(path1)
	if entry1 == nil {
		t.Fatal("Expected non-nil FileEntry")
	}

	if fl.Count() != 1 {
		t.Errorf("Expected count 1, got %d", fl.Count())
	}

	if entry1.Order != 1 {
		t.Errorf("Expected Order 1, got %d", entry1.Order)
	}

	// 添加第二个文件
	entry2 := fl.AddFile(path2)
	if fl.Count() != 2 {
		t.Errorf("Expected count 2, got %d", fl.Count())
	}

	if entry2.Order != 2 {
		t.Errorf("Expected Order 2, got %d", entry2.Order)
	}

	// 尝试添加重复文件
	entry3 := fl.AddFile(path1)
	if entry3 != entry1 {
		t.Error("Expected AddFile to return existing entry for duplicate path")
	}

	if fl.Count() != 2 {
		t.Errorf("Expected count to remain 2, got %d", fl.Count())
	}
}

func TestFileList_RemoveFile(t *testing.T) {
	fl := NewFileList()
	path1 := "/path/to/file1.pdf"
	path2 := "/path/to/file2.pdf"
	path3 := "/path/to/file3.pdf"

	fl.AddFile(path1)
	fl.AddFile(path2)
	fl.AddFile(path3)

	// 移除中间的文件
	removed := fl.RemoveFile(path2)
	if !removed {
		t.Error("Expected RemoveFile to return true")
	}

	if fl.Count() != 2 {
		t.Errorf("Expected count 2, got %d", fl.Count())
	}

	// 检查剩余文件的顺序是否正确
	files := fl.GetFiles()
	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	if files[0].Path != path1 || files[0].Order != 1 {
		t.Errorf("Expected first file to be %s with order 1, got %s with order %d", path1, files[0].Path, files[0].Order)
	}

	if files[1].Path != path3 || files[1].Order != 2 {
		t.Errorf("Expected second file to be %s with order 2, got %s with order %d", path3, files[1].Path, files[1].Order)
	}

	// 尝试移除不存在的文件
	removed = fl.RemoveFile("/nonexistent.pdf")
	if removed {
		t.Error("Expected RemoveFile to return false for nonexistent file")
	}
}

func TestFileList_MoveFile(t *testing.T) {
	fl := NewFileList()
	path1 := "/path/to/file1.pdf"
	path2 := "/path/to/file2.pdf"
	path3 := "/path/to/file3.pdf"

	fl.AddFile(path1)
	fl.AddFile(path2)
	fl.AddFile(path3)

	// 将第三个文件移动到第一位
	moved := fl.MoveFile(path3, 1)
	if !moved {
		t.Error("Expected MoveFile to return true")
	}

	files := fl.GetFiles()
	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(files))
	}

	// 检查新的顺序
	expectedOrder := []string{path3, path1, path2}
	for i, expectedPath := range expectedOrder {
		if files[i].Path != expectedPath {
			t.Errorf("Expected file at position %d to be %s, got %s", i, expectedPath, files[i].Path)
		}
		if files[i].Order != i+1 {
			t.Errorf("Expected file at position %d to have order %d, got %d", i, i+1, files[i].Order)
		}
	}

	// 尝试移动到无效位置
	moved = fl.MoveFile(path1, 0)
	if moved {
		t.Error("Expected MoveFile to return false for invalid position")
	}

	moved = fl.MoveFile(path1, 5)
	if moved {
		t.Error("Expected MoveFile to return false for position > count")
	}
}

func TestFileList_GetFilePaths(t *testing.T) {
	fl := NewFileList()
	paths := []string{"/file1.pdf", "/file2.pdf", "/file3.pdf"}

	for _, path := range paths {
		fl.AddFile(path)
	}

	filePaths := fl.GetFilePaths()
	if len(filePaths) != len(paths) {
		t.Errorf("Expected %d paths, got %d", len(paths), len(filePaths))
	}

	for i, expectedPath := range paths {
		if filePaths[i] != expectedPath {
			t.Errorf("Expected path at position %d to be %s, got %s", i, expectedPath, filePaths[i])
		}
	}
}

func TestFileList_GetAllFilePaths(t *testing.T) {
	fl := NewFileList()
	mainPath := "/main.pdf"
	additionalPaths := []string{"/file1.pdf", "/file2.pdf"}

	fl.SetMainFile(mainPath)
	for _, path := range additionalPaths {
		fl.AddFile(path)
	}

	allPaths := fl.GetAllFilePaths()
	expectedTotal := 1 + len(additionalPaths)
	if len(allPaths) != expectedTotal {
		t.Errorf("Expected %d total paths, got %d", expectedTotal, len(allPaths))
	}

	if allPaths[0] != mainPath {
		t.Errorf("Expected first path to be main file %s, got %s", mainPath, allPaths[0])
	}

	for i, expectedPath := range additionalPaths {
		if allPaths[i+1] != expectedPath {
			t.Errorf("Expected path at position %d to be %s, got %s", i+1, expectedPath, allPaths[i+1])
		}
	}
}

func TestFileList_Clear(t *testing.T) {
	fl := NewFileList()
	fl.SetMainFile("/main.pdf")
	fl.AddFile("/file1.pdf")
	fl.AddFile("/file2.pdf")

	fl.Clear()

	if !fl.IsEmpty() {
		t.Error("Expected IsEmpty to be true after Clear")
	}

	if fl.HasMainFile() {
		t.Error("Expected HasMainFile to be false after Clear")
	}

	if fl.Count() != 0 {
		t.Errorf("Expected count 0 after Clear, got %d", fl.Count())
	}
}

func TestFileList_GetValidFiles(t *testing.T) {
	fl := NewFileList()

	// 设置主文件
	mainFile := fl.SetMainFile("/main.pdf")

	// 添加有效和无效的文件
	validFile := fl.AddFile("/valid.pdf")
	invalidFile := fl.AddFile("/invalid.pdf")
	invalidFile.SetError("Test error")

	validFiles := fl.GetValidFiles()

	// 应该包含主文件和有效的附加文件
	expectedCount := 2
	if len(validFiles) != expectedCount {
		t.Errorf("Expected %d valid files, got %d", expectedCount, len(validFiles))
	}

	// 检查返回的文件是否正确
	foundMain := false
	foundValid := false

	for _, file := range validFiles {
		if file == mainFile {
			foundMain = true
		} else if file == validFile {
			foundValid = true
		}
	}

	if !foundMain {
		t.Error("Expected main file to be in valid files")
	}

	if !foundValid {
		t.Error("Expected valid additional file to be in valid files")
	}
}
