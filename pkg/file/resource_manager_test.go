package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResourceManager_AddFile(t *testing.T) {
	tempDir := t.TempDir()
	rm := NewResourceManager()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 添加文件资源
	rm.AddFile(testFile, 1)

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 1 {
		t.Errorf("期望资源数量为1，实际为: %d", count)
	}

	// 清理资源
	errors := rm.Cleanup()
	if len(errors) > 0 {
		t.Errorf("清理资源时发生错误: %v", errors)
	}

	// 验证文件已被删除
	if FileExists(testFile) {
		t.Errorf("文件未被删除: %s", testFile)
	}

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 0 {
		t.Errorf("期望资源数量为0，实际为: %d", count)
	}
}

func TestResourceManager_AddDirectory(t *testing.T) {
	tempDir := t.TempDir()
	rm := NewResourceManager()

	// 创建测试目录
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 在目录中创建文件
	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 添加目录资源
	rm.AddDirectory(testDir, 1)

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 1 {
		t.Errorf("期望资源数量为1，实际为: %d", count)
	}

	// 清理资源
	errors := rm.Cleanup()
	if len(errors) > 0 {
		t.Errorf("清理资源时发生错误: %v", errors)
	}

	// 验证目录已被删除
	if DirExists(testDir) {
		t.Errorf("目录未被删除: %s", testDir)
	}

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 0 {
		t.Errorf("期望资源数量为0，实际为: %d", count)
	}
}

func TestResourceManager_AddCustom(t *testing.T) {
	rm := NewResourceManager()

	// 创建自定义资源
	customCalled := false
	rm.AddCustom(func() error {
		customCalled = true
		return nil
	}, 1)

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 1 {
		t.Errorf("期望资源数量为1，实际为: %d", count)
	}

	// 清理资源
	errors := rm.Cleanup()
	if len(errors) > 0 {
		t.Errorf("清理资源时发生错误: %v", errors)
	}

	// 验证自定义清理函数被调用
	if !customCalled {
		t.Error("自定义清理函数未被调用")
	}

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 0 {
		t.Errorf("期望资源数量为0，实际为: %d", count)
	}
}

func TestResourceManager_CleanupResource(t *testing.T) {
	tempDir := t.TempDir()
	rm := NewResourceManager()

	// 创建多个测试文件
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	if err := os.WriteFile(testFile1, []byte("test content 1"), 0644); err != nil {
		t.Fatalf("创建测试文件1失败: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("创建测试文件2失败: %v", err)
	}

	// 添加文件资源
	rm.AddFile(testFile1, 1)
	rm.AddFile(testFile2, 2)

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 2 {
		t.Errorf("期望资源数量为2，实际为: %d", count)
	}

	// 清理特定资源
	err := rm.CleanupResource(testFile1)
	if err != nil {
		t.Errorf("清理特定资源时发生错误: %v", err)
	}

	// 验证文件1已被删除
	if FileExists(testFile1) {
		t.Errorf("文件1未被删除: %s", testFile1)
	}

	// 验证文件2仍然存在
	if !FileExists(testFile2) {
		t.Errorf("文件2不应被删除: %s", testFile2)
	}

	// 验证资源数量
	if count := rm.GetResourceCount(); count != 1 {
		t.Errorf("期望资源数量为1，实际为: %d", count)
	}

	// 清理剩余资源
	errors := rm.Cleanup()
	if len(errors) > 0 {
		t.Errorf("清理资源时发生错误: %v", errors)
	}

	// 验证文件2已被删除
	if FileExists(testFile2) {
		t.Errorf("文件2未被删除: %s", testFile2)
	}
}

func TestResourceManager_PriorityOrder(t *testing.T) {
	rm := NewResourceManager()

	// 创建有优先级的自定义资源
	executionOrder := make([]int, 0)
	
	rm.AddCustom(func() error {
		executionOrder = append(executionOrder, 1)
		return nil
	}, 1) // 低优先级
	
	rm.AddCustom(func() error {
		executionOrder = append(executionOrder, 3)
		return nil
	}, 3) // 高优先级
	
	rm.AddCustom(func() error {
		executionOrder = append(executionOrder, 2)
		return nil
	}, 2) // 中优先级

	// 清理资源
	rm.Cleanup()

	// 验证执行顺序（应该是按优先级从高到低）
	expectedOrder := []int{3, 2, 1}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("执行顺序长度不匹配，期望: %v, 实际: %v", expectedOrder, executionOrder)
	} else {
		for i, expected := range expectedOrder {
			if executionOrder[i] != expected {
				t.Errorf("执行顺序不匹配，位置 %d 期望: %d, 实际: %d", i, expected, executionOrder[i])
			}
		}
	}
}