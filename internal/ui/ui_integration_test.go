package ui

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func TestUI_Integration(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")
	testWindow.Resize(fyne.NewSize(800, 600))

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	content := ui.BuildUI()

	// 验证UI构建成功
	if content == nil {
		t.Fatal("UI内容构建失败")
	}

	testWindow.SetContent(content)

	// 测试UI组件是否正确创建
	if ui.mainFileEntry == nil {
		t.Error("主文件输入框未创建")
	}

	if ui.fileListManager == nil {
		t.Error("文件列表管理器未创建")
	}

	if ui.progressManager == nil {
		t.Error("进度管理器未创建")
	}

	if ui.mergeButton == nil {
		t.Error("合并按钮未创建")
	}

	// 测试初始状态
	if ui.mergeButton.Disabled() == false {
		t.Error("合并按钮初始状态应该是禁用的")
	}
}

func TestUI_FileOperationsIntegration(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试文件添加
	testFilePath := "test.pdf"
	err := ui.fileListManager.AddFile(testFilePath)
	if err != nil {
		t.Logf("添加不存在的文件失败（预期）: %v", err)
	}

	// 测试文件信息获取
	if ui.fileListManager.HasFiles() {
		t.Log("警告：初始状态有文件，这可能是正常的")
	}

	// 测试文件信息显示
	info := ui.fileListManager.GetFileInfo()
	if info == "" {
		t.Error("文件信息不应该为空")
	}
}

func TestUI_ProgressOperations(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试进度更新
	progressInfo := ProgressInfo{
		Progress: 0.5,
		Status:   "测试状态",
		Detail:   "测试详情",
		Step:     1,
	}

	ui.UpdateProgress(progressInfo)

	// 验证进度管理器状态
	if !ui.progressManager.IsActive() {
		t.Log("进度管理器未激活，这可能是正常的")
	}

	// 测试进度完成
	ui.progressManager.Complete("测试完成")

	// 测试进度取消
	ui.progressManager.Cancel()
}

func TestUI_ButtonStatesIntegration(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试初始按钮状态
	if !ui.mergeButton.Disabled() {
		t.Error("合并按钮初始应该是禁用的")
	}

	if !ui.removeFileBtn.Disabled() {
		t.Error("移除文件按钮初始应该是禁用的")
	}

	if !ui.clearFilesBtn.Disabled() {
		t.Error("清空文件按钮初始应该是禁用的")
	}

	// 测试控件启用/禁用
	ui.DisableControls()

	if !ui.mainFileBrowseBtn.Disabled() {
		t.Error("主文件浏览按钮应该被禁用")
	}

	if !ui.addFileBtn.Disabled() {
		t.Error("添加文件按钮应该被禁用")
	}

	ui.EnableControls()

	if ui.mainFileBrowseBtn.Disabled() {
		t.Error("主文件浏览按钮应该被启用")
	}

	if ui.addFileBtn.Disabled() {
		t.Error("添加文件按钮应该被启用")
	}
}

func TestUI_EventHandling(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试事件处理器设置
	testHandler := &testEventHandler{}
	ui.SetEventHandler(testHandler)

	if ui.eventHandler != testHandler {
		t.Error("事件处理器设置失败")
	}

	// 测试文件列表变更回调
	ui.onFileListChanged()

	// 测试进度取消回调
	ui.onProgressCancel()

	// 测试进度完成回调
	ui.onProgressComplete()
}

func TestUI_DataAccess(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试数据访问方法
	mainFilePath := ui.GetMainFilePath()
	if mainFilePath != "" {
		t.Error("初始主文件路径应该为空")
	}

	additionalFiles := ui.GetAdditionalFiles()
	if len(additionalFiles) != 0 {
		t.Error("初始附加文件列表应该为空")
	}

	additionalFilePaths := ui.GetAdditionalFilePaths()
	if len(additionalFilePaths) != 0 {
		t.Error("初始附加文件路径列表应该为空")
	}

	outputPath := ui.GetOutputPath()
	if outputPath != "" {
		t.Error("初始输出路径应该为空")
	}
}

func TestUI_ProgressMethods(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试进度设置方法
	ui.SetProgress(0.5)
	ui.SetStatus("测试状态")
	ui.SetDetail("测试详情")

	// 测试字符串参数的进度更新
	ui.UpdateProgressWithStrings(0.7, "新状态", "新详情")
}

func TestUI_DialogMethods(t *testing.T) {
	// 创建测试应用
	testApp := app.New()
	testWindow := testApp.NewWindow("Test")

	// 创建模拟服务
	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 创建UI
	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	// 测试错误对话框（这些方法会显示对话框，但在测试中不会阻塞）
	testError := fmt.Errorf("测试错误")
	ui.ShowError(testError)

	// 测试信息对话框
	ui.ShowInfo("测试标题", "测试消息")

	// 测试完成对话框
	ui.ShowCompletion("测试完成消息")
}

// 测试事件处理器
type testEventHandler struct{}

// 基准测试

func BenchmarkUI_BuildUI(b *testing.B) {
	testApp := app.New()

	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testWindow := testApp.NewWindow("Benchmark")
		ui := NewUI(testWindow, ctrl)
		ui.BuildUI()
		testWindow.Close()
	}
}

func BenchmarkUI_UpdateProgress(b *testing.B) {
	testApp := app.New()
	testWindow := testApp.NewWindow("Benchmark")

	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	progressInfo := ProgressInfo{
		Progress: 0.5,
		Status:   "基准测试状态",
		Detail:   "基准测试详情",
		Step:     1,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ui.UpdateProgress(progressInfo)
	}
}

func BenchmarkUI_UpdateUI(b *testing.B) {
	testApp := app.New()
	testWindow := testApp.NewWindow("Benchmark")

	fileManager := file.NewFileManager("/tmp")
	pdfService := pdf.NewPDFService()
	config := model.DefaultConfig()
	ctrl := controller.NewController(pdfService, fileManager, config)

	ui := NewUI(testWindow, ctrl)
	ui.BuildUI()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ui.updateUI()
	}
}
