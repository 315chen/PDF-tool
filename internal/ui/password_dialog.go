package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// PasswordDialog 密码输入对话框
type PasswordDialog struct {
	window        fyne.Window
	dialog        *dialog.CustomDialog
	passwordEntry *widget.Entry
	rememberCheck *widget.Check
	result        chan PasswordDialogResult
	filePath      string
	attempt       int
	lastError     error
}

// PasswordDialogResult 密码对话框结果
type PasswordDialogResult struct {
	Password     string
	Remember     bool
	UserCanceled bool
}

// PasswordDialogOptions 密码对话框选项
type PasswordDialogOptions struct {
	Title        string
	Message      string
	ShowRemember bool
	ShowAttempt  bool
	ShowError    bool
}

// NewPasswordDialog 创建新的密码输入对话框
func NewPasswordDialog(window fyne.Window, filePath string, attempt int, lastError error, options *PasswordDialogOptions) *PasswordDialog {
	if options == nil {
		options = &PasswordDialogOptions{
			Title:        "需要密码",
			ShowRemember: true,
			ShowAttempt:  true,
			ShowError:    true,
		}
	}

	pd := &PasswordDialog{
		window:    window,
		filePath:  filePath,
		attempt:   attempt,
		lastError: lastError,
		result:    make(chan PasswordDialogResult, 1),
	}

	pd.createDialog(options)
	return pd
}

// createDialog 创建对话框界面
func (pd *PasswordDialog) createDialog(options *PasswordDialogOptions) {
	// 创建密码输入框
	pd.passwordEntry = widget.NewPasswordEntry()
	pd.passwordEntry.SetPlaceHolder("请输入PDF文件密码")

	// 创建记住密码复选框
	pd.rememberCheck = widget.NewCheck("记住此文件的密码", nil)
	pd.rememberCheck.SetChecked(true)

	// 创建内容容器
	content := container.NewVBox()

	// 添加文件信息
	fileName := filepath.Base(pd.filePath)
	fileLabel := widget.NewLabel(fmt.Sprintf("文件: %s", fileName))
	fileLabel.Wrapping = fyne.TextWrapWord
	content.Add(fileLabel)

	// 添加尝试次数信息
	if options.ShowAttempt && pd.attempt > 1 {
		attemptLabel := widget.NewLabel(fmt.Sprintf("尝试次数: %d/3", pd.attempt))
		attemptLabel.Importance = widget.MediumImportance
		content.Add(attemptLabel)
	}

	// 添加错误信息
	if options.ShowError && pd.lastError != nil {
		errorLabel := widget.NewLabel(fmt.Sprintf("错误: %v", pd.lastError))
		errorLabel.Importance = widget.DangerImportance
		content.Add(errorLabel)
	}

	// 添加消息
	if options.Message != "" {
		messageLabel := widget.NewLabel(options.Message)
		messageLabel.Wrapping = fyne.TextWrapWord
		content.Add(messageLabel)
	}

	// 添加密码输入框
	content.Add(widget.NewLabel("密码:"))
	content.Add(pd.passwordEntry)

	// 添加记住密码选项
	if options.ShowRemember {
		content.Add(pd.rememberCheck)
	}

	// 创建按钮
	confirmButton := widget.NewButton("确定", pd.onConfirm)
	confirmButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("取消", pd.onCancel)
	skipButton := widget.NewButton("跳过此文件", pd.onSkip)

	buttonContainer := container.NewHBox(
		confirmButton,
		cancelButton,
		skipButton,
	)

	content.Add(buttonContainer)

	// 创建对话框
	pd.dialog = dialog.NewCustom(options.Title, "", content, pd.window)
	pd.dialog.Resize(fyne.NewSize(400, 300))

	// 设置回车键确认
	pd.passwordEntry.OnSubmitted = func(string) {
		pd.onConfirm()
	}

	// 聚焦到密码输入框
	pd.window.Canvas().Focus(pd.passwordEntry)
}

// onConfirm 确认按钮处理
func (pd *PasswordDialog) onConfirm() {
	password := pd.passwordEntry.Text
	remember := pd.rememberCheck.Checked

	result := PasswordDialogResult{
		Password:     password,
		Remember:     remember,
		UserCanceled: false,
	}

	pd.result <- result
	pd.dialog.Hide()
}

// onCancel 取消按钮处理
func (pd *PasswordDialog) onCancel() {
	result := PasswordDialogResult{
		Password:     "",
		Remember:     false,
		UserCanceled: true,
	}

	pd.result <- result
	pd.dialog.Hide()
}

// onSkip 跳过按钮处理
func (pd *PasswordDialog) onSkip() {
	result := PasswordDialogResult{
		Password:     "",
		Remember:     false,
		UserCanceled: true,
	}

	pd.result <- result
	pd.dialog.Hide()
}

// Show 显示对话框
func (pd *PasswordDialog) Show() {
	pd.dialog.Show()
}

// GetResult 获取对话框结果
func (pd *PasswordDialog) GetResult() PasswordDialogResult {
	return <-pd.result
}

// ShowAndWait 显示对话框并等待结果
func (pd *PasswordDialog) ShowAndWait() PasswordDialogResult {
	pd.Show()
	return pd.GetResult()
}

// CreateGUIPasswordPrompt 创建GUI密码输入提示函数
func CreateGUIPasswordPrompt(window fyne.Window) func(filePath string, attempt int, lastError error) (string, bool) {
	return func(filePath string, attempt int, lastError error) (string, bool) {
		options := &PasswordDialogOptions{
			Title:        "PDF文件需要密码",
			ShowRemember: true,
			ShowAttempt:  true,
			ShowError:    attempt > 1,
		}

		if attempt == 1 {
			options.Message = "此PDF文件已加密，请输入密码以继续。"
		} else {
			options.Message = "密码错误，请重新输入。"
		}

		dialog := NewPasswordDialog(window, filePath, attempt, lastError, options)
		result := dialog.ShowAndWait()

		if result.UserCanceled {
			return "", false
		}

		return result.Password, true
	}
}

// PasswordBatchDialog 批量密码输入对话框
type PasswordBatchDialog struct {
	window       fyne.Window
	dialog       *dialog.CustomDialog
	passwordList *widget.List
	passwords    []string
	result       chan []string
}

// NewPasswordBatchDialog 创建批量密码输入对话框
func NewPasswordBatchDialog(window fyne.Window, initialPasswords []string) *PasswordBatchDialog {
	pbd := &PasswordBatchDialog{
		window:    window,
		passwords: make([]string, len(initialPasswords)),
		result:    make(chan []string, 1),
	}

	copy(pbd.passwords, initialPasswords)
	pbd.createDialog()
	return pbd
}

// createDialog 创建批量密码对话框界面
func (pbd *PasswordBatchDialog) createDialog() {
	// 创建密码列表
	pbd.passwordList = widget.NewList(
		func() int {
			return len(pbd.passwords)
		},
		func() fyne.CanvasObject {
			entry := widget.NewEntry()
			entry.SetPlaceHolder("输入密码")
			return entry
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			entry := obj.(*widget.Entry)
			if id < len(pbd.passwords) {
				entry.SetText(pbd.passwords[id])
				entry.OnChanged = func(text string) {
					if id < len(pbd.passwords) {
						pbd.passwords[id] = text
					}
				}
			}
		},
	)

	// 创建按钮
	addButton := widget.NewButton("添加密码", func() {
		pbd.passwords = append(pbd.passwords, "")
		pbd.passwordList.Refresh()
	})

	removeButton := widget.NewButton("删除最后一个", func() {
		if len(pbd.passwords) > 0 {
			pbd.passwords = pbd.passwords[:len(pbd.passwords)-1]
			pbd.passwordList.Refresh()
		}
	})

	confirmButton := widget.NewButton("确定", func() {
		// 过滤空密码
		validPasswords := make([]string, 0)
		for _, password := range pbd.passwords {
			if password != "" {
				validPasswords = append(validPasswords, password)
			}
		}
		pbd.result <- validPasswords
		pbd.dialog.Hide()
	})

	cancelButton := widget.NewButton("取消", func() {
		pbd.result <- nil
		pbd.dialog.Hide()
	})

	// 创建布局
	buttonContainer := container.NewHBox(addButton, removeButton)
	actionContainer := container.NewHBox(confirmButton, cancelButton)

	content := container.NewVBox(
		widget.NewLabel("批量密码设置"),
		widget.NewLabel("为多个加密PDF文件设置常用密码:"),
		buttonContainer,
		pbd.passwordList,
		actionContainer,
	)

	pbd.dialog = dialog.NewCustom("批量密码设置", "", content, pbd.window)
	pbd.dialog.Resize(fyne.NewSize(400, 500))
}

// Show 显示批量密码对话框
func (pbd *PasswordBatchDialog) Show() {
	pbd.dialog.Show()
}

// GetResult 获取批量密码结果
func (pbd *PasswordBatchDialog) GetResult() []string {
	return <-pbd.result
}

// ShowAndWait 显示对话框并等待结果
func (pbd *PasswordBatchDialog) ShowAndWait() []string {
	pbd.Show()
	return pbd.GetResult()
}

// PasswordCacheDialog 密码缓存管理对话框
type PasswordCacheDialog struct {
	window     fyne.Window
	dialog     *dialog.CustomDialog
	cacheList  *widget.List
	cacheFiles []string
	onClear    func(filePath string)
}

// NewPasswordCacheDialog 创建密码缓存管理对话框
func NewPasswordCacheDialog(window fyne.Window, cacheFiles []string, onClear func(string)) *PasswordCacheDialog {
	pcd := &PasswordCacheDialog{
		window:     window,
		cacheFiles: cacheFiles,
		onClear:    onClear,
	}

	pcd.createDialog()
	return pcd
}

// createDialog 创建缓存管理对话框界面
func (pcd *PasswordCacheDialog) createDialog() {
	// 创建缓存文件列表
	pcd.cacheList = widget.NewList(
		func() int {
			return len(pcd.cacheFiles)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewButton("清除", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			button := container.Objects[1].(*widget.Button)

			if id < len(pcd.cacheFiles) {
				filePath := pcd.cacheFiles[id]
				fileName := filepath.Base(filePath)
				label.SetText(fileName)

				button.OnTapped = func() {
					if pcd.onClear != nil {
						pcd.onClear(filePath)
					}
					// 从列表中移除
					pcd.cacheFiles = append(pcd.cacheFiles[:id], pcd.cacheFiles[id+1:]...)
					pcd.cacheList.Refresh()
				}
			}
		},
	)

	// 创建按钮
	clearAllButton := widget.NewButton("清除全部", func() {
		if pcd.onClear != nil {
			for _, filePath := range pcd.cacheFiles {
				pcd.onClear(filePath)
			}
		}
		pcd.cacheFiles = pcd.cacheFiles[:0]
		pcd.cacheList.Refresh()
	})

	closeButton := widget.NewButton("关闭", func() {
		pcd.dialog.Hide()
	})

	// 创建布局
	content := container.NewVBox(
		widget.NewLabel("密码缓存管理"),
		widget.NewLabel(fmt.Sprintf("当前缓存了 %d 个文件的密码:", len(pcd.cacheFiles))),
		pcd.cacheList,
		container.NewHBox(clearAllButton, closeButton),
	)

	pcd.dialog = dialog.NewCustom("密码缓存管理", "", content, pcd.window)
	pcd.dialog.Resize(fyne.NewSize(500, 400))
}

// Show 显示缓存管理对话框
func (pcd *PasswordCacheDialog) Show() {
	pcd.dialog.Show()
}
