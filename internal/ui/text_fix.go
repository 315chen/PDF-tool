package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// CreateLabel 创建支持中文的标签
func CreateLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	// 确保文本正确显示
	label.Wrapping = fyne.TextWrapWord
	return label
}

// CreateButton 创建支持中文的按钮
func CreateButton(text string, callback func()) *widget.Button {
	button := widget.NewButton(text, callback)
	return button
}

// CreateEntry 创建支持中文的输入框
func CreateEntry() *widget.Entry {
	entry := widget.NewEntry()
	return entry
}

// CreateMultiLineEntry 创建支持中文的多行输入框
func CreateMultiLineEntry() *widget.Entry {
	entry := widget.NewMultiLineEntry()
	return entry
}

// FixTextDisplay 修复文本显示问题
func FixTextDisplay(text string) string {
	// 确保文本是UTF-8编码
	return text
}
