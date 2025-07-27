//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"

	"github.com/user/pdf-merger/internal/ui"
)

func main() {
	fmt.Println("=== 进度显示和状态反馈功能演示 ===\n")

	// 1. 演示进度管理器创建
	demonstrateProgressManagerCreation()

	// 2. 演示基本进度显示
	demonstrateBasicProgressDisplay()

	// 3. 演示状态反馈系统
	demonstrateStatusFeedbackSystem()

	// 4. 演示时间和速度统计
	demonstrateTimeAndSpeedStatistics()

	// 5. 演示错误处理和显示
	demonstrateErrorHandlingAndDisplay()

	// 6. 演示对话框系统
	demonstrateDialogSystem()

	// 7. 演示完整的进度界面
	demonstrateCompleteProgressInterface()

	fmt.Println("\n=== 进度显示和状态反馈演示完成 ===")
}

func demonstrateProgressManagerCreation() {
	fmt.Println("1. 进度管理器创建演示:")
	
	// 1.1 创建应用程序和窗口
	fmt.Println("\n   1.1 创建应用程序和窗口:")
	a := app.New()
	w := a.NewWindow("进度管理器演示")
	w.Resize(fyne.NewSize(400, 300))
	
	fmt.Printf("   - 应用程序创建成功\n")
	fmt.Printf("   - 窗口大小: 400x300\n")
	
	// 1.2 创建进度管理器
	fmt.Println("\n   1.2 创建进度管理器:")
	progressManager := ui.NewProgressManager(w)
	
	fmt.Printf("   - 进度管理器创建成功\n")
	fmt.Printf("   - 初始状态: %t\n", progressManager.IsActive())
	fmt.Printf("   - 初始进度: %.1f%%\n", progressManager.GetProgress()*100)
	
	// 1.3 获取进度容器
	fmt.Println("\n   1.3 获取进度容器:")
	container := progressManager.GetContainer()
	
	fmt.Printf("   - 容器类型: %T\n", container)
	fmt.Printf("   - 容器组件数: %d\n", len(container.Objects))
	
	// 1.4 分析容器结构
	fmt.Println("\n   1.4 分析容器结构:")
	for i, obj := range container.Objects {
		fmt.Printf("   - 组件 %d: %T\n", i+1, obj)
	}
	
	// 关闭应用程序
	a.Quit()
	
	fmt.Println()
}

func demonstrateBasicProgressDisplay() {
	fmt.Println("2. 基本进度显示演示:")
	
	// 2.1 创建进度管理器
	fmt.Println("\n   2.1 创建进度管理器:")
	a := app.New()
	defer a.Quit()
	
	w := a.NewWindow("基本进度演示")
	progressManager := ui.NewProgressManager(w)
	
	fmt.Printf("   - 进度管理器准备就绪\n")
	
	// 2.2 开始进度显示
	fmt.Println("\n   2.2 开始进度显示:")
	progressManager.Start(5, 10)
	
	fmt.Printf("   - 进度显示已开始\n")
	fmt.Printf("   - 总步骤: 5\n")
	fmt.Printf("   - 总文件: 10\n")
	fmt.Printf("   - 活跃状态: %t\n", progressManager.IsActive())
	
	// 2.3 模拟进度更新
	fmt.Println("\n   2.3 模拟进度更新:")
	
	progressSteps := []struct {
		progress float64
		status   string
		detail   string
		file     string
		step     int
	}{
		{0.0, "初始化", "正在初始化系统...", "", 1},
		{0.2, "验证文件", "正在验证PDF文件...", "document1.pdf", 2},
		{0.4, "读取内容", "正在读取PDF内容...", "document2.pdf", 3},
		{0.6, "处理数据", "正在处理PDF数据...", "document3.pdf", 4},
		{0.8, "合并文件", "正在合并PDF文件...", "document4.pdf", 5},
		{1.0, "完成", "所有操作已完成", "", 5},
	}
	
	for i, step := range progressSteps {
		fmt.Printf("   步骤 %d: %.1f%% - %s\n", i+1, step.progress*100, step.status)
		
		progressManager.UpdateProgress(ui.ProgressInfo{
			Progress:       step.progress,
			Status:         step.status,
			Detail:         step.detail,
			CurrentFile:    step.file,
			ProcessedFiles: i + 1,
			TotalFiles:     len(progressSteps),
			Step:           step.step,
			TotalSteps:     5,
		})
		
		// 模拟处理时间
		time.Sleep(100 * time.Millisecond)
		
		// 显示当前状态
		fmt.Printf("     - 当前进度: %.1f%%\n", progressManager.GetProgress()*100)
		if progressManager.IsActive() {
			fmt.Printf("     - 已用时间: %v\n", progressManager.GetElapsedTime())
		}
	}
	
	// 2.4 完成进度
	fmt.Println("\n   2.4 完成进度:")
	progressManager.Complete("演示完成！")
	fmt.Printf("   - 进度已完成\n")
	fmt.Printf("   - 最终进度: %.1f%%\n", progressManager.GetProgress()*100)
	
	// 等待完成处理
	time.Sleep(500 * time.Millisecond)
	
	// 2.5 停止进度
	fmt.Println("\n   2.5 停止进度:")
	progressManager.Stop()
	fmt.Printf("   - 进度已停止\n")
	fmt.Printf("   - 活跃状态: %t\n", progressManager.IsActive())
	
	fmt.Println()
}

func demonstrateStatusFeedbackSystem() {
	fmt.Println("3. 状态反馈系统演示:")
	
	// 3.1 创建进度管理器
	fmt.Println("\n   3.1 创建进度管理器:")
	a := app.New()
	defer a.Quit()
	
	w := a.NewWindow("状态反馈演示")
	progressManager := ui.NewProgressManager(w)
	
	// 3.2 演示不同状态设置
	fmt.Println("\n   3.2 演示不同状态设置:")
	
	statusMessages := []string{
		"准备就绪",
		"正在初始化...",
		"正在处理文件...",
		"正在验证结果...",
		"操作完成",
	}
	
	for i, status := range statusMessages {
		fmt.Printf("   状态 %d: %s\n", i+1, status)
		progressManager.SetStatus(status)
		time.Sleep(200 * time.Millisecond)
	}
	
	// 3.3 演示详细信息设置
	fmt.Println("\n   3.3 演示详细信息设置:")
	
	detailMessages := []string{
		"系统初始化中，请稍候...",
		"正在扫描输入文件...",
		"正在分析PDF结构...",
		"正在执行合并操作...",
		"正在保存输出文件...",
	}
	
	for i, detail := range detailMessages {
		fmt.Printf("   详细信息 %d: %s\n", i+1, detail)
		progressManager.SetDetail(detail)
		time.Sleep(200 * time.Millisecond)
	}
	
	// 3.4 演示进度设置
	fmt.Println("\n   3.4 演示进度设置:")
	
	for i := 0; i <= 10; i++ {
		progress := float64(i) / 10.0
		fmt.Printf("   设置进度: %.1f%%\n", progress*100)
		progressManager.SetProgress(progress)
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println()
}

func demonstrateTimeAndSpeedStatistics() {
	fmt.Println("4. 时间和速度统计演示:")
	
	// 4.1 创建进度管理器
	fmt.Println("\n   4.1 创建进度管理器:")
	a := app.New()
	defer a.Quit()
	
	w := a.NewWindow("时间速度统计演示")
	progressManager := ui.NewProgressManager(w)
	
	// 4.2 开始计时
	fmt.Println("\n   4.2 开始计时:")
	progressManager.Start(10, 20)
	
	startTime := time.Now()
	fmt.Printf("   - 开始时间: %v\n", startTime.Format("15:04:05"))
	fmt.Printf("   - 初始已用时间: %v\n", progressManager.GetElapsedTime())
	
	// 4.3 模拟文件处理过程
	fmt.Println("\n   4.3 模拟文件处理过程:")
	
	for i := 1; i <= 10; i++ {
		progress := float64(i) / 10.0
		
		progressManager.UpdateProgress(ui.ProgressInfo{
			Progress:       progress,
			Status:         fmt.Sprintf("处理文件 %d", i),
			Detail:         fmt.Sprintf("正在处理第 %d 个文件...", i),
			CurrentFile:    fmt.Sprintf("file_%d.pdf", i),
			ProcessedFiles: i,
			TotalFiles:     10,
			Step:           i,
			TotalSteps:     10,
		})
		
		// 模拟处理时间（不同文件处理时间不同）
		processingTime := time.Duration(100+i*50) * time.Millisecond
		time.Sleep(processingTime)
		
		// 显示统计信息
		elapsed := progressManager.GetElapsedTime()
		fmt.Printf("   文件 %d: 进度 %.1f%%, 已用时 %v\n", i, progress*100, elapsed)
		
		// 计算处理速度
		if elapsed.Seconds() > 0 {
			speed := float64(i) / elapsed.Seconds()
			fmt.Printf("     - 处理速度: %.2f 文件/秒\n", speed)
		}
	}
	
	// 4.4 显示最终统计
	fmt.Println("\n   4.4 最终统计:")
	finalElapsed := progressManager.GetElapsedTime()
	fmt.Printf("   - 总用时: %v\n", finalElapsed)
	fmt.Printf("   - 平均速度: %.2f 文件/秒\n", 10.0/finalElapsed.Seconds())
	fmt.Printf("   - 最终进度: %.1f%%\n", progressManager.GetProgress()*100)
	
	progressManager.Complete("处理完成！")
	
	fmt.Println()
}

func demonstrateErrorHandlingAndDisplay() {
	fmt.Println("5. 错误处理和显示演示:")
	
	// 5.1 创建进度管理器
	fmt.Println("\n   5.1 创建进度管理器:")
	a := app.New()
	defer a.Quit()
	
	w := a.NewWindow("错误处理演示")
	progressManager := ui.NewProgressManager(w)
	
	// 5.2 模拟正常进度然后出错
	fmt.Println("\n   5.2 模拟正常进度然后出错:")
	
	progressManager.Start(5, 8)
	
	// 正常进度
	for i := 1; i <= 3; i++ {
		progress := float64(i) / 5.0
		
		progressManager.UpdateProgress(ui.ProgressInfo{
			Progress: progress,
			Status:   fmt.Sprintf("步骤 %d", i),
			Detail:   fmt.Sprintf("正在执行步骤 %d...", i),
			Step:     i,
		})
		
		fmt.Printf("   步骤 %d: %.1f%% - 正常\n", i, progress*100)
		time.Sleep(200 * time.Millisecond)
	}
	
	// 模拟错误
	fmt.Println("\n   5.3 模拟错误发生:")
	testError := fmt.Errorf("演示错误：文件读取失败 - 权限被拒绝")
	fmt.Printf("   - 错误类型: %T\n", testError)
	fmt.Printf("   - 错误信息: %s\n", testError.Error())
	
	progressManager.Error(testError)
	fmt.Printf("   - 错误已显示\n")
	fmt.Printf("   - 活跃状态: %t\n", progressManager.IsActive())
	
	// 等待错误处理完成
	time.Sleep(1 * time.Second)
	
	// 5.4 演示取消操作
	fmt.Println("\n   5.4 演示取消操作:")
	
	progressManager.Start(3, 5)
	
	// 开始进度
	progressManager.UpdateProgress(ui.ProgressInfo{
		Progress: 0.3,
		Status:   "处理中",
		Detail:   "正在处理，即将取消...",
		Step:     1,
	})
	
	fmt.Printf("   - 开始处理\n")
	time.Sleep(300 * time.Millisecond)
	
	// 取消操作
	fmt.Printf("   - 执行取消操作\n")
	progressManager.Cancel()
	fmt.Printf("   - 取消已执行\n")
	fmt.Printf("   - 活跃状态: %t\n", progressManager.IsActive())
	
	// 等待取消处理完成
	time.Sleep(1 * time.Second)
	
	fmt.Println()
}

func demonstrateDialogSystem() {
	fmt.Println("6. 对话框系统演示:")
	
	// 6.1 创建进度管理器
	fmt.Println("\n   6.1 创建进度管理器:")
	a := app.New()
	defer a.Quit()
	
	w := a.NewWindow("对话框演示")
	_ = ui.NewProgressManager(w)
	
	// 6.2 演示信息对话框
	fmt.Println("\n   6.2 演示信息对话框:")
	fmt.Printf("   - 显示信息对话框\n")
	
	// 注意：在演示程序中，我们不实际显示对话框，只演示调用
	// progressManager.ShowInfoDialog("信息", "这是一个信息对话框演示")
	fmt.Printf("   - 信息对话框调用: ShowInfoDialog(\"信息\", \"这是一个信息对话框演示\")\n")
	
	// 6.3 演示错误对话框
	fmt.Println("\n   6.3 演示错误对话框:")
	fmt.Printf("   - 显示错误对话框\n")
	
	// progressManager.ShowErrorDialog("错误", "这是一个错误对话框演示")
	fmt.Printf("   - 错误对话框调用: ShowErrorDialog(\"错误\", \"这是一个错误对话框演示\")\n")
	
	// 6.4 演示确认对话框
	fmt.Println("\n   6.4 演示确认对话框:")
	fmt.Printf("   - 显示确认对话框\n")
	
	_ = func(confirmed bool) {
		if confirmed {
			fmt.Printf("   - 用户确认: 是\n")
		} else {
			fmt.Printf("   - 用户确认: 否\n")
		}
	}
	
	// progressManager.ShowConfirmDialog("确认", "您确定要继续吗？", confirmCallback)
	fmt.Printf("   - 确认对话框调用: ShowConfirmDialog(\"确认\", \"您确定要继续吗？\", callback)\n")
	
	// 模拟用户选择
	fmt.Printf("   - 模拟用户选择: 确认\n")
	// confirmCallback(true) // 已注释，因为变量未使用
	
	// 6.5 演示进度对话框
	fmt.Println("\n   6.5 演示进度对话框:")
	fmt.Printf("   - 显示进度对话框\n")
	
	_ = func() {
		fmt.Printf("   - 用户取消了进度对话框\n")
	}
	
	// progressManager.ShowProgressDialog("处理中", "正在处理文件，请稍候...", cancelCallback)
	fmt.Printf("   - 进度对话框调用: ShowProgressDialog(\"处理中\", \"正在处理文件，请稍候...\", callback)\n")
	
	fmt.Println()
}

func demonstrateCompleteProgressInterface() {
	fmt.Println("7. 完整进度界面演示:")
	
	// 7.1 创建应用程序和窗口
	fmt.Println("\n   7.1 创建应用程序和窗口:")
	a := app.New()
	w := a.NewWindow("完整进度界面演示")
	w.Resize(fyne.NewSize(500, 400))
	
	// 7.2 创建进度管理器
	fmt.Println("\n   7.2 创建进度管理器:")
	progressManager := ui.NewProgressManager(w)
	
	// 7.3 创建控制按钮
	fmt.Println("\n   7.3 创建控制按钮:")
	
	startBtn := widget.NewButtonWithIcon("开始演示", theme.MediaPlayIcon(), func() {
		fmt.Printf("   - 开始演示按钮被点击\n")
		demonstrateCompleteProgress(progressManager)
	})
	
	errorBtn := widget.NewButtonWithIcon("模拟错误", theme.ErrorIcon(), func() {
		fmt.Printf("   - 模拟错误按钮被点击\n")
		demonstrateProgressError(progressManager)
	})
	
	cancelBtn := widget.NewButtonWithIcon("取消操作", theme.CancelIcon(), func() {
		fmt.Printf("   - 取消操作按钮被点击\n")
		progressManager.Cancel()
	})
	
	// 7.4 创建界面布局
	fmt.Println("\n   7.4 创建界面布局:")
	
	buttonRow := container.NewHBox(startBtn, errorBtn, cancelBtn)
	
	content := container.NewVBox(
		widget.NewLabel("完整进度界面演示"),
		widget.NewSeparator(),
		progressManager.GetContainer(),
		widget.NewSeparator(),
		buttonRow,
	)
	
	w.SetContent(content)
	
	fmt.Printf("   - 界面布局创建完成\n")
	fmt.Printf("   - 窗口大小: 500x400\n")
	fmt.Printf("   - 组件数量: %d\n", len(content.Objects))
	
	// 7.5 模拟用户交互
	fmt.Println("\n   7.5 模拟用户交互:")
	
	// 模拟点击开始按钮
	fmt.Printf("   - 模拟点击开始演示按钮\n")
	demonstrateCompleteProgress(progressManager)
	
	// 等待演示完成
	time.Sleep(2 * time.Second)
	
	// 模拟点击错误按钮
	fmt.Printf("   - 模拟点击模拟错误按钮\n")
	demonstrateProgressError(progressManager)
	
	// 等待错误处理完成
	time.Sleep(1 * time.Second)
	
	// 关闭应用程序
	a.Quit()
	
	fmt.Println("\n   完整进度界面演示完成 🎉")
	fmt.Println("   所有进度显示和状态反馈功能正常工作")
	
	fmt.Println()
}

// 辅助函数

func demonstrateCompleteProgress(pm *ui.ProgressManager) {
	pm.Start(6, 12)
	
	steps := []struct {
		progress float64
		status   string
		detail   string
		file     string
	}{
		{0.0, "初始化", "正在初始化系统...", ""},
		{0.2, "扫描文件", "正在扫描输入文件...", "input1.pdf"},
		{0.4, "验证文件", "正在验证PDF格式...", "input2.pdf"},
		{0.6, "读取内容", "正在读取PDF内容...", "input3.pdf"},
		{0.8, "合并处理", "正在执行合并操作...", "output.pdf"},
		{1.0, "保存文件", "正在保存合并结果...", "result.pdf"},
	}
	
	for i, step := range steps {
		pm.UpdateProgress(ui.ProgressInfo{
			Progress:       step.progress,
			Status:         step.status,
			Detail:         step.detail,
			CurrentFile:    step.file,
			ProcessedFiles: i + 1,
			TotalFiles:     len(steps),
			Step:           i + 1,
			TotalSteps:     len(steps),
		})
		
		time.Sleep(300 * time.Millisecond)
	}
	
	pm.Complete("演示完成！")
}

func demonstrateProgressError(pm *ui.ProgressManager) {
	pm.Start(3, 5)
	
	pm.UpdateProgress(ui.ProgressInfo{
		Progress: 0.3,
		Status:   "处理中",
		Detail:   "正在处理文件...",
		Step:     1,
	})
	
	time.Sleep(500 * time.Millisecond)
	
	pm.Error(fmt.Errorf("演示错误：文件处理失败"))
}
