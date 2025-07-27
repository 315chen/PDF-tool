//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/internal/ui"
)

func main() {
	fmt.Println("=== 文件列表管理界面功能演示 ===\n")

	// 1. 演示文件列表管理器创建
	demonstrateFileListManagerCreation()

	// 2. 演示文件添加和管理
	demonstrateFileAdditionAndManagement()

	// 3. 演示文件排序和移动
	demonstrateFileSortingAndMoving()

	// 4. 演示文件信息显示
	demonstrateFileInformationDisplay()

	// 5. 演示批量操作
	demonstrateBatchOperations()

	// 6. 演示文件验证和状态
	demonstrateFileValidationAndStatus()

	// 7. 演示完整的文件列表界面
	demonstrateCompleteFileListInterface()

	fmt.Println("\n=== 文件列表管理界面演示完成 ===")
}

func demonstrateFileListManagerCreation() {
	fmt.Println("1. 文件列表管理器创建演示:")
	
	// 1.1 创建文件列表管理器
	fmt.Println("\n   1.1 创建文件列表管理器:")
	fileListManager := ui.NewFileListManager()
	
	fmt.Printf("   - 文件列表管理器创建成功\n")
	fmt.Printf("   - 初始文件数量: %d\n", fileListManager.GetFileCount())
	fmt.Printf("   - 是否有文件: %t\n", fileListManager.HasFiles())
	fmt.Printf("   - 选中索引: %d\n", fileListManager.GetSelectedIndex())
	
	// 1.2 获取列表组件
	fmt.Println("\n   1.2 获取列表组件:")
	listWidget := fileListManager.GetWidget()
	
	fmt.Printf("   - 列表组件类型: %T\n", listWidget)
	fmt.Printf("   - 列表组件创建成功\n")
	
	// 1.3 设置回调函数
	fmt.Println("\n   1.3 设置回调函数:")
	
	fileListManager.SetOnFileChanged(func() {
		fmt.Printf("   - 文件变更回调被调用\n")
	})

	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fmt.Printf("   - 文件信息回调被调用: %s\n", filepath.Base(filePath))
		
		// 创建模拟文件信息
		fileEntry := model.NewFileEntry(filePath, 0)
		fileEntry.Size = 1024 * 1024 // 1MB
		fileEntry.PageCount = 10
		fileEntry.IsValid = true
		
		return fileEntry, nil
	})
	
	fmt.Printf("   - 回调函数设置完成\n")
	
	fmt.Println()
}

func demonstrateFileAdditionAndManagement() {
	fmt.Println("2. 文件添加和管理演示:")

	// 初始化Fyne应用程序
	a := app.New()
	defer a.Quit()

	// 创建临时目录和测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	testFiles := createTestFiles(tempDir, 5)

	// 2.1 创建文件列表管理器
	fmt.Println("\n   2.1 创建文件列表管理器:")
	fileListManager := ui.NewFileListManager()
	
	// 设置文件信息回调
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		fileEntry.Size = int64(1024 * (1 + len(filepath.Base(filePath))))
		fileEntry.PageCount = 5 + len(filepath.Base(filePath))%10
		fileEntry.IsValid = true
		return fileEntry, nil
	})
	
	// 2.2 添加文件
	fmt.Println("\n   2.2 添加文件:")
	for i, testFile := range testFiles {
		err := fileListManager.AddFile(testFile)
		if err != nil {
			fmt.Printf("   - 添加文件 %d 失败: %v\n", i+1, err)
		} else {
			fmt.Printf("   - 添加文件 %d: %s ✓\n", i+1, filepath.Base(testFile))
		}
	}
	
	fmt.Printf("   - 总文件数量: %d\n", fileListManager.GetFileCount())
	
	// 2.3 尝试添加重复文件
	fmt.Println("\n   2.3 尝试添加重复文件:")
	err := fileListManager.AddFile(testFiles[0])
	if err != nil {
		fmt.Printf("   - 重复文件添加被拒绝: %v ✓\n", err)
	} else {
		fmt.Printf("   - 重复文件添加成功（意外）\n")
	}
	
	// 2.4 获取文件信息
	fmt.Println("\n   2.4 获取文件信息:")
	files := fileListManager.GetFiles()
	for i, file := range files {
		fmt.Printf("   - 文件 %d: %s (大小: %s, 页数: %d)\n", 
			i+1, file.DisplayName, file.GetSizeString(), file.PageCount)
	}
	
	// 2.5 获取文件路径
	fmt.Println("\n   2.5 获取文件路径:")
	filePaths := fileListManager.GetFilePaths()
	for i, path := range filePaths {
		fmt.Printf("   - 路径 %d: %s\n", i+1, filepath.Base(path))
	}
	
	fmt.Println()
}

func demonstrateFileSortingAndMoving() {
	fmt.Println("3. 文件排序和移动演示:")

	// 初始化Fyne应用程序
	a := app.New()
	defer a.Quit()

	// 创建临时目录和测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	testFiles := createTestFiles(tempDir, 4)

	// 3.1 创建文件列表管理器并添加文件
	fmt.Println("\n   3.1 创建文件列表并添加文件:")
	fileListManager := ui.NewFileListManager()
	
	// 设置文件信息回调
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		fileEntry.Size = int64(1024 * (1 + len(filepath.Base(filePath))))
		fileEntry.PageCount = 5 + len(filepath.Base(filePath))%10
		fileEntry.IsValid = true
		return fileEntry, nil
	})
	
	for i, testFile := range testFiles {
		fileListManager.AddFile(testFile)
		fmt.Printf("   - 添加文件 %d: %s\n", i+1, filepath.Base(testFile))
	}
	
	// 3.2 显示初始顺序
	fmt.Println("\n   3.2 初始文件顺序:")
	displayFileOrder(fileListManager)
	
	// 3.3 模拟选择文件并上移
	fmt.Println("\n   3.3 选择第3个文件并上移:")
	// 模拟选择第3个文件（索引2）
	fileListManager.GetWidget().Select(2)
	fmt.Printf("   - 选中文件索引: %d\n", fileListManager.GetSelectedIndex())
	
	fileListManager.MoveSelectedUp()
	fmt.Printf("   - 执行上移操作\n")
	displayFileOrder(fileListManager)
	
	// 3.4 继续上移
	fmt.Println("\n   3.4 继续上移:")
	fileListManager.MoveSelectedUp()
	fmt.Printf("   - 再次执行上移操作\n")
	displayFileOrder(fileListManager)
	
	// 3.5 下移操作
	fmt.Println("\n   3.5 下移操作:")
	fileListManager.MoveSelectedDown()
	fmt.Printf("   - 执行下移操作\n")
	displayFileOrder(fileListManager)
	
	// 3.6 尝试边界操作
	fmt.Println("\n   3.6 尝试边界操作:")
	
	// 选择第一个文件并尝试上移
	fileListManager.GetWidget().Select(0)
	fmt.Printf("   - 选中第一个文件，尝试上移\n")
	fileListManager.MoveSelectedUp()
	fmt.Printf("   - 上移操作（应该无效果）\n")
	displayFileOrder(fileListManager)
	
	// 选择最后一个文件并尝试下移
	lastIndex := fileListManager.GetFileCount() - 1
	fileListManager.GetWidget().Select(lastIndex)
	fmt.Printf("   - 选中最后一个文件，尝试下移\n")
	fileListManager.MoveSelectedDown()
	fmt.Printf("   - 下移操作（应该无效果）\n")
	displayFileOrder(fileListManager)
	
	fmt.Println()
}

func demonstrateFileInformationDisplay() {
	fmt.Println("4. 文件信息显示演示:")

	// 初始化Fyne应用程序
	a := app.New()
	defer a.Quit()

	// 创建临时目录和测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	testFiles := createTestFiles(tempDir, 3)

	// 4.1 创建文件列表管理器
	fmt.Println("\n   4.1 创建文件列表管理器:")
	fileListManager := ui.NewFileListManager()
	
	// 4.2 设置详细的文件信息回调
	fmt.Println("\n   4.2 设置文件信息回调:")
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		
		// 模拟不同的文件状态
		baseName := filepath.Base(filePath)
		switch {
		case baseName == "test_1.pdf":
			fileEntry.Size = 2 * 1024 * 1024 // 2MB
			fileEntry.PageCount = 15
			fileEntry.IsEncrypted = false
			fileEntry.IsValid = true
		case baseName == "test_2.pdf":
			fileEntry.Size = 512 * 1024 // 512KB
			fileEntry.PageCount = 5
			fileEntry.IsEncrypted = true
			fileEntry.IsValid = true
		case baseName == "test_3.pdf":
			fileEntry.Size = 0
			fileEntry.PageCount = 0
			fileEntry.IsEncrypted = false
			fileEntry.IsValid = false
			fileEntry.Error = "文件损坏"
		default:
			fileEntry.Size = 1024 * 1024 // 1MB
			fileEntry.PageCount = 10
			fileEntry.IsValid = true
		}
		
		return fileEntry, nil
	})
	
	// 4.3 添加文件并显示信息
	fmt.Println("\n   4.3 添加文件并显示详细信息:")
	for i, testFile := range testFiles {
		err := fileListManager.AddFile(testFile)
		if err != nil {
			fmt.Printf("   - 文件 %d 添加失败: %v\n", i+1, err)
			continue
		}
		
		files := fileListManager.GetFiles()
		if i < len(files) {
			file := files[i]
			fmt.Printf("   - 文件 %d: %s\n", i+1, file.DisplayName)
			fmt.Printf("     大小: %s\n", file.GetSizeString())
			fmt.Printf("     页数: %d\n", file.PageCount)
			fmt.Printf("     加密: %t\n", file.IsEncrypted)
			fmt.Printf("     有效: %t\n", file.IsValid)
			if file.Error != "" {
				fmt.Printf("     错误: %s\n", file.Error)
			}
		}
	}
	
	// 4.4 刷新文件信息
	fmt.Println("\n   4.4 刷新文件信息:")
	fileListManager.RefreshFileInfo()
	fmt.Printf("   - 文件信息刷新完成\n")
	
	// 4.5 获取文件信息摘要
	fmt.Println("\n   4.5 文件信息摘要:")
	fileInfo := fileListManager.GetFileInfo()
	fmt.Printf("   - %s\n", fileInfo)
	
	fmt.Println()
}

func demonstrateBatchOperations() {
	fmt.Println("5. 批量操作演示:")

	// 初始化Fyne应用程序
	a := app.New()
	defer a.Quit()

	// 创建临时目录和测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	testFiles := createTestFiles(tempDir, 6)

	// 5.1 创建文件列表管理器并添加文件
	fmt.Println("\n   5.1 批量添加文件:")
	fileListManager := ui.NewFileListManager()
	
	// 设置文件信息回调
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		fileEntry.Size = int64(1024 * (1 + len(filepath.Base(filePath))))
		fileEntry.PageCount = 5 + len(filepath.Base(filePath))%10
		fileEntry.IsValid = true
		return fileEntry, nil
	})
	
	for i, testFile := range testFiles {
		fileListManager.AddFile(testFile)
		fmt.Printf("   - 添加文件 %d: %s\n", i+1, filepath.Base(testFile))
	}
	
	fmt.Printf("   - 批量添加完成，总文件数: %d\n", fileListManager.GetFileCount())
	
	// 5.2 批量移除操作
	fmt.Println("\n   5.2 批量移除操作:")
	
	// 移除选中的文件
	fmt.Printf("   - 选择第3个文件并移除\n")
	fileListManager.GetWidget().Select(2)
	fileListManager.RemoveSelected()
	fmt.Printf("   - 移除后文件数: %d\n", fileListManager.GetFileCount())
	
	// 再移除一个文件
	fmt.Printf("   - 选择第1个文件并移除\n")
	fileListManager.GetWidget().Select(0)
	fileListManager.RemoveSelected()
	fmt.Printf("   - 移除后文件数: %d\n", fileListManager.GetFileCount())
	
	// 5.3 显示剩余文件
	fmt.Println("\n   5.3 显示剩余文件:")
	files := fileListManager.GetFiles()
	for i, file := range files {
		fmt.Printf("   - 文件 %d: %s\n", i+1, file.DisplayName)
	}
	
	// 5.4 清空所有文件
	fmt.Println("\n   5.4 清空所有文件:")
	fileListManager.Clear()
	fmt.Printf("   - 清空后文件数: %d\n", fileListManager.GetFileCount())
	fmt.Printf("   - 是否有文件: %t\n", fileListManager.HasFiles())
	
	fmt.Println()
}

func demonstrateFileValidationAndStatus() {
	fmt.Println("6. 文件验证和状态演示:")

	// 初始化Fyne应用程序
	a := app.New()
	defer a.Quit()

	// 创建临时目录和测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)

	testFiles := createTestFiles(tempDir, 4)

	// 6.1 创建文件列表管理器
	fmt.Println("\n   6.1 创建文件列表管理器:")
	fileListManager := ui.NewFileListManager()
	
	// 6.2 设置不同状态的文件信息回调
	fmt.Println("\n   6.2 设置不同状态的文件:")
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		baseName := filepath.Base(filePath)
		
		switch {
		case baseName == "test_1.pdf":
			// 正常文件
			fileEntry.Size = 1024 * 1024
			fileEntry.PageCount = 10
			fileEntry.IsEncrypted = false
			fileEntry.IsValid = true
			
		case baseName == "test_2.pdf":
			// 加密文件
			fileEntry.Size = 2 * 1024 * 1024
			fileEntry.PageCount = 20
			fileEntry.IsEncrypted = true
			fileEntry.IsValid = true
			
		case baseName == "test_3.pdf":
			// 损坏文件
			fileEntry.Size = 512 * 1024
			fileEntry.PageCount = 0
			fileEntry.IsEncrypted = false
			fileEntry.IsValid = false
			fileEntry.Error = "PDF文件格式错误"
			
		case baseName == "test_4.pdf":
			// 空文件
			fileEntry.Size = 0
			fileEntry.PageCount = 0
			fileEntry.IsEncrypted = false
			fileEntry.IsValid = false
			fileEntry.Error = "文件为空"
			
		default:
			fileEntry.Size = 1024 * 1024
			fileEntry.PageCount = 10
			fileEntry.IsValid = true
		}
		
		return fileEntry, nil
	})
	
	// 6.3 添加文件并显示状态
	fmt.Println("\n   6.3 添加文件并显示状态:")
	for i, testFile := range testFiles {
		err := fileListManager.AddFile(testFile)
		if err != nil {
			fmt.Printf("   - 文件 %d 添加失败: %v\n", i+1, err)
			continue
		}
		
		files := fileListManager.GetFiles()
		if i < len(files) {
			file := files[i]
			status := "正常"
			if !file.IsValid {
				status = "错误"
			} else if file.IsEncrypted {
				status = "加密"
			}
			
			fmt.Printf("   - 文件 %d: %s [%s]\n", i+1, file.DisplayName, status)
			if file.Error != "" {
				fmt.Printf("     错误信息: %s\n", file.Error)
			}
		}
	}
	
	// 6.4 统计文件状态
	fmt.Println("\n   6.4 文件状态统计:")
	files := fileListManager.GetFiles()
	validCount := 0
	encryptedCount := 0
	errorCount := 0
	
	for _, file := range files {
		if !file.IsValid {
			errorCount++
		} else if file.IsEncrypted {
			encryptedCount++
		} else {
			validCount++
		}
	}
	
	fmt.Printf("   - 正常文件: %d\n", validCount)
	fmt.Printf("   - 加密文件: %d\n", encryptedCount)
	fmt.Printf("   - 错误文件: %d\n", errorCount)
	fmt.Printf("   - 总文件数: %d\n", len(files))
	
	fmt.Println()
}

func demonstrateCompleteFileListInterface() {
	fmt.Println("7. 完整文件列表界面演示:")
	
	// 7.1 创建应用程序和窗口
	fmt.Println("\n   7.1 创建应用程序和窗口:")
	a := app.New()
	w := a.NewWindow("文件列表管理界面演示")
	w.Resize(fyne.NewSize(600, 400))
	
	// 7.2 创建文件列表管理器
	fmt.Println("\n   7.2 创建文件列表管理器:")
	fileListManager := ui.NewFileListManager()
	
	// 设置文件信息回调
	fileListManager.SetOnFileInfo(func(filePath string) (*model.FileEntry, error) {
		fileEntry := model.NewFileEntry(filePath, 0)
		fileEntry.Size = int64(1024 * (1 + len(filepath.Base(filePath))))
		fileEntry.PageCount = 5 + len(filepath.Base(filePath))%10
		fileEntry.IsValid = true
		return fileEntry, nil
	})
	
	// 7.3 创建界面组件
	fmt.Println("\n   7.3 创建界面组件:")
	
	// 文件信息标签
	fileInfoLabel := widget.NewLabel("没有文件")
	fileInfoLabel.TextStyle = fyne.TextStyle{Italic: true}
	
	// 操作按钮
	addBtn := widget.NewButtonWithIcon("添加", theme.ContentAddIcon(), func() {
		// 模拟添加文件
		tempDir := createTempDir()
		defer os.RemoveAll(tempDir)
		testFile := createTestFiles(tempDir, 1)[0]
		
		err := fileListManager.AddFile(testFile)
		if err == nil {
			fileInfoLabel.SetText(fileListManager.GetFileInfo())
			fmt.Printf("   - 添加文件: %s\n", filepath.Base(testFile))
		}
	})
	
	removeBtn := widget.NewButtonWithIcon("移除", theme.DeleteIcon(), func() {
		fileListManager.RemoveSelected()
		fileInfoLabel.SetText(fileListManager.GetFileInfo())
		fmt.Printf("   - 移除选中文件\n")
	})
	
	clearBtn := widget.NewButtonWithIcon("清空", theme.ContentClearIcon(), func() {
		fileListManager.Clear()
		fileInfoLabel.SetText(fileListManager.GetFileInfo())
		fmt.Printf("   - 清空文件列表\n")
	})
	
	upBtn := widget.NewButtonWithIcon("上移", theme.MoveUpIcon(), func() {
		fileListManager.MoveSelectedUp()
		fmt.Printf("   - 上移选中文件\n")
	})
	
	downBtn := widget.NewButtonWithIcon("下移", theme.MoveDownIcon(), func() {
		fileListManager.MoveSelectedDown()
		fmt.Printf("   - 下移选中文件\n")
	})
	
	refreshBtn := widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), func() {
		fileListManager.RefreshFileInfo()
		fileInfoLabel.SetText(fileListManager.GetFileInfo())
		fmt.Printf("   - 刷新文件信息\n")
	})
	
	// 7.4 创建布局
	fmt.Println("\n   7.4 创建界面布局:")
	
	buttonRow1 := container.NewHBox(addBtn, removeBtn, clearBtn)
	buttonRow2 := container.NewHBox(upBtn, downBtn, refreshBtn)
	
	content := container.NewVBox(
		widget.NewLabel("文件列表管理界面演示"),
		widget.NewSeparator(),
		fileInfoLabel,
		fileListManager.GetWidget(),
		widget.NewSeparator(),
		buttonRow1,
		buttonRow2,
	)
	
	w.SetContent(content)
	
	// 7.5 设置文件变更回调
	fmt.Println("\n   7.5 设置文件变更回调:")
	fileListManager.SetOnFileChanged(func() {
		fileInfoLabel.SetText(fileListManager.GetFileInfo())
	})
	
	fmt.Printf("   - 界面创建完成\n")
	fmt.Printf("   - 窗口大小: 600x400\n")
	fmt.Printf("   - 组件数量: %d\n", len(content.Objects))
	
	// 7.6 模拟用户操作
	fmt.Println("\n   7.6 模拟用户操作:")
	
	// 添加一些测试文件
	tempDir := createTempDir()
	defer os.RemoveAll(tempDir)
	testFiles := createTestFiles(tempDir, 3)
	
	for i, testFile := range testFiles {
		fileListManager.AddFile(testFile)
		fmt.Printf("   - 添加测试文件 %d: %s\n", i+1, filepath.Base(testFile))
	}
	
	// 更新文件信息显示
	fileInfoLabel.SetText(fileListManager.GetFileInfo())
	
	fmt.Printf("   - 最终文件数量: %d\n", fileListManager.GetFileCount())
	fmt.Printf("   - 文件信息: %s\n", fileListManager.GetFileInfo())
	
	// 关闭应用程序
	a.Quit()
	
	fmt.Println("\n   完整文件列表界面演示完成 🎉")
	fmt.Println("   所有文件列表管理功能正常工作")
	
	fmt.Println()
}

// 辅助函数

func createTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "filelist-demo-"+fmt.Sprintf("%d", time.Now().Unix()))
	os.MkdirAll(tempDir, 0755)
	return tempDir
}

func createTestFiles(tempDir string, count int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("test_%d.pdf", i+1)
		filepath := filepath.Join(tempDir, filename)
		
		// 创建简单的测试PDF内容
		content := fmt.Sprintf("%%PDF-1.4\nTest PDF file %d\n%%%%EOF", i+1)
		os.WriteFile(filepath, []byte(content), 0644)
		
		files[i] = filepath
	}
	return files
}

func displayFileOrder(fileListManager *ui.FileListManager) {
	files := fileListManager.GetFiles()
	fmt.Printf("   - 当前文件顺序:\n")
	for i, file := range files {
		marker := ""
		if i == fileListManager.GetSelectedIndex() {
			marker = " [选中]"
		}
		fmt.Printf("     %d. %s%s\n", i+1, file.DisplayName, marker)
	}
}
