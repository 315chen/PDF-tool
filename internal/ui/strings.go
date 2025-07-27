package ui

// 界面文本常量 - 使用英文避免字体问题
const (
	// 窗口标题
	WindowTitle = "PDF Merger Tool"

	// 按钮文本
	BrowseButton     = "Browse..."
	AddFileButton    = "Add Files"
	RemoveFileButton = "Remove Selected"
	ClearFilesButton = "Clear All"
	MoveUpButton     = "Move Up"
	MoveDownButton   = "Move Down"
	RefreshButton    = "Refresh"
	StartMergeButton = "Start Merge"
	CancelButton     = "Cancel"

	// 标签文本
	MainFileLabel        = "Main PDF File:"
	AdditionalFilesLabel = "Additional PDF Files:"
	OutputPathLabel      = "Output Path:"
	NoFilesLabel         = "No files"
	ProgressLabel        = "Progress:"
	StatusLabel          = "Status:"

	// 状态消息
	StatusReadyText     = "Ready"
	StatusMerging       = "Merging..."
	StatusCompletedText = "Completed"
	StatusCancelledText = "Cancelled"
	StatusErrorText     = "Error"

	// 对话框文本
	SelectMainFileTitle = "Select Main PDF File"
	SelectFilesTitle    = "Select PDF Files"
	SelectOutputTitle   = "Select Output Location"
	ErrorDialogTitle    = "Error"
	InfoDialogTitle     = "Information"
	SuccessDialogTitle  = "Success"

	// 文件过滤器
	PDFFileFilter = "PDF Files (*.pdf)"

	// 错误消息
	ErrorNoMainFile   = "Please select a main PDF file first"
	ErrorNoFiles      = "Please add at least one PDF file"
	ErrorInvalidFile  = "Invalid PDF file"
	ErrorMergeFailed  = "Merge failed"
	ErrorFileNotFound = "File not found"

	// 成功消息
	SuccessMergeComplete = "PDF files merged successfully!"

	// 提示消息
	HintDropFiles      = "Drag PDF files here or click Add Files button"
	HintSelectMainFile = "Please select a main PDF file as the base for merging"
	HintSelectOutput   = "Please select the output file location"
)
