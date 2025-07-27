package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func TestNewUI(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)

	if ui == nil {
		t.Error("NewUI returned nil")
	}

	if ui.window != window {
		t.Error("UI window not set correctly")
	}

	if ui.controller != ctrl {
		t.Error("UI controller not set correctly")
	}

	if ui.fileListManager == nil {
		t.Error("File list manager not initialized")
	}
}

func TestUI_BuildUI(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)

	// 构建UI
	content := ui.BuildUI()

	if content == nil {
		t.Error("BuildUI returned nil content")
	}

	// 检查UI组件是否已创建
	if ui.mainFileEntry == nil {
		t.Error("Main file entry not created")
	}

	if ui.mainFileBrowseBtn == nil {
		t.Error("Main file browse button not created")
	}

	if ui.fileListManager.GetWidget() == nil {
		t.Error("File list widget not created")
	}

	if ui.addFileBtn == nil {
		t.Error("Add file button not created")
	}

	if ui.removeFileBtn == nil {
		t.Error("Remove file button not created")
	}

	if ui.clearFilesBtn == nil {
		t.Error("Clear files button not created")
	}

	if ui.outputPathEntry == nil {
		t.Error("Output path entry not created")
	}

	if ui.outputBrowseBtn == nil {
		t.Error("Output browse button not created")
	}

	if ui.progressManager == nil {
		t.Error("Progress manager not created")
	}

	if ui.mergeButton == nil {
		t.Error("Merge button not created")
	}

	if ui.cancelButton == nil {
		t.Error("Cancel button not created")
	}
}

func TestUI_UpdateUI(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)
	ui.BuildUI()

	// 测试初始状态
	ui.updateUI()

	// 合并按钮应该被禁用（没有文件）
	if ui.mergeButton.Disabled() == false {
		t.Error("Merge button should be disabled initially")
	}

	// 移除和清空按钮应该被禁用（没有附加文件）
	if ui.removeFileBtn.Disabled() == false {
		t.Error("Remove button should be disabled initially")
	}

	if ui.clearFilesBtn.Disabled() == false {
		t.Error("Clear button should be disabled initially")
	}
}

func TestUI_GettersAndSetters(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)
	ui.BuildUI()

	// 测试主文件路径
	testPath := "/test/main.pdf"
	ui.mainFilePath = testPath
	if ui.GetMainFilePath() != testPath {
		t.Errorf("Expected main file path %s, got %s", testPath, ui.GetMainFilePath())
	}

	// 测试输出路径
	testOutput := "/test/output.pdf"
	ui.outputPath = testOutput
	if ui.GetOutputPath() != testOutput {
		t.Errorf("Expected output path %s, got %s", testOutput, ui.GetOutputPath())
	}

	// 测试附加文件列表
	files := ui.GetAdditionalFiles()
	if files == nil {
		t.Error("GetAdditionalFiles returned nil")
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 additional files, got %d", len(files))
	}

	// 测试进度设置
	ui.SetProgress(0.5)
	if ui.progressManager.GetProgress() != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", ui.progressManager.GetProgress())
	}

	// 测试状态设置
	testStatus := "测试状态"
	ui.SetStatus(testStatus)
	if ui.progressManager.statusLabel.Text != testStatus {
		t.Errorf("Expected status %s, got %s", testStatus, ui.progressManager.statusLabel.Text)
	}
}

func TestUI_FileOperations(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)
	ui.BuildUI()

	// 测试添加文件到列表
	err := ui.fileListManager.AddFile("/test/file1.pdf")
	if err != nil {
		t.Errorf("Failed to add file: %v", err)
	}

	files := ui.GetAdditionalFiles()
	if len(files) != 1 {
		t.Errorf("Expected 1 additional file, got %d", len(files))
	}

	if files[0].Path != "/test/file1.pdf" {
		t.Errorf("Expected file path /test/file1.pdf, got %s", files[0].Path)
	}

	// 测试清空文件列表
	ui.fileListManager.Clear()
	files = ui.GetAdditionalFiles()
	if len(files) != 0 {
		t.Errorf("Expected 0 additional files after clear, got %d", len(files))
	}
}

func TestUI_ButtonStates(t *testing.T) {
	// 创建测试应用和窗口
	app := test.NewApp()
	window := app.NewWindow("Test")

	// 创建模拟服务
	pdfService := pdf.NewPDFService()
	fileManager := file.NewFileManager("/tmp")
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(window, ctrl)
	ui.BuildUI()

	// 测试开始合并状态
	ui.startMerge()

	// 合并按钮应该隐藏
	if ui.mergeButton.Visible() {
		t.Error("Merge button should be hidden during merge")
	}

	// 取消按钮应该显示
	if !ui.cancelButton.Visible() {
		t.Error("Cancel button should be visible during merge")
	}

	// 进度管理器应该活跃
	if !ui.progressManager.IsActive() {
		t.Error("Progress manager should be active during merge")
	}

	// 输入控件应该被禁用
	if !ui.mainFileBrowseBtn.Disabled() {
		t.Error("Main file browse button should be disabled during merge")
	}

	// 测试取消合并状态
	ui.cancelMerge()

	// 合并按钮应该显示
	if !ui.mergeButton.Visible() {
		t.Error("Merge button should be visible after cancel")
	}

	// 取消按钮应该隐藏
	if ui.cancelButton.Visible() {
		t.Error("Cancel button should be hidden after cancel")
	}

	// 进度管理器应该不活跃
	if ui.progressManager.IsActive() {
		t.Error("Progress manager should not be active after cancel")
	}

	// 输入控件应该被启用
	if ui.mainFileBrowseBtn.Disabled() {
		t.Error("Main file browse button should be enabled after cancel")
	}
}
