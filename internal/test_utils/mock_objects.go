package test_utils

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

// MockPDFService 模拟PDF服务
type MockPDFService struct {
	mutex           sync.RWMutex
	mergeResults    map[string]error
	validateResults map[string]error
	infoResults     map[string]*pdf.PDFInfo
	mergeDelay      time.Duration
	validateDelay   time.Duration
	callCounts      map[string]int
}

// NewMockPDFService 创建新的模拟PDF服务
func NewMockPDFService() *MockPDFService {
	return &MockPDFService{
		mergeResults:    make(map[string]error),
		validateResults: make(map[string]error),
		infoResults:     make(map[string]*pdf.PDFInfo),
		callCounts:      make(map[string]int),
	}
}

// SetMergeResult 设置合并结果
func (m *MockPDFService) SetMergeResult(outputPath string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.mergeResults[outputPath] = err
}

// SetValidateResult 设置验证结果
func (m *MockPDFService) SetValidateResult(filePath string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.validateResults[filePath] = err
}

// SetInfoResult 设置文件信息结果
func (m *MockPDFService) SetInfoResult(filePath string, info *pdf.PDFInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.infoResults[filePath] = info
}

// SetMergeDelay 设置合并延迟
func (m *MockPDFService) SetMergeDelay(delay time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.mergeDelay = delay
}

// SetValidateDelay 设置验证延迟
func (m *MockPDFService) SetValidateDelay(delay time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.validateDelay = delay
}

// GetCallCount 获取调用次数
func (m *MockPDFService) GetCallCount(method string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.callCounts[method]
}

// Merge 模拟合并PDF文件
func (m *MockPDFService) Merge(ctx context.Context, mainFile string, additionalFiles []string, outputPath string, progressCallback func(float64)) error {
	m.mutex.Lock()
	m.callCounts["Merge"]++
	delay := m.mergeDelay
	result := m.mergeResults[outputPath]
	m.mutex.Unlock()

	// 模拟进度更新
	if progressCallback != nil {
		go func() {
			for i := 0; i <= 100; i += 10 {
				select {
				case <-ctx.Done():
					return
				default:
					progressCallback(float64(i))
					time.Sleep(delay / 10)
				}
			}
		}()
	}

	// 模拟处理时间
	if delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return result
}

// Validate 模拟验证PDF文件
func (m *MockPDFService) Validate(filePath string) error {
	m.mutex.Lock()
	m.callCounts["Validate"]++
	delay := m.validateDelay
	result := m.validateResults[filePath]
	m.mutex.Unlock()

	// 模拟处理时间
	if delay > 0 {
		time.Sleep(delay)
	}

	return result
}

// GetInfo 模拟获取PDF文件信息
func (m *MockPDFService) GetInfo(filePath string) (*pdf.PDFInfo, error) {
	m.mutex.Lock()
	m.callCounts["GetInfo"]++
	info := m.infoResults[filePath]
	m.mutex.Unlock()

	if info == nil {
		// 返回默认信息
		return &pdf.PDFInfo{
			PageCount:    10,
			FileSize:     1024 * 1024, // 1MB
			IsEncrypted:  false,
			Title:        "Mock PDF",
			Author:       "Test Author",
			Subject:      "Test Subject",
			Creator:      "Mock Creator",
			Producer:     "Mock Producer",
			CreationDate: time.Now(),
			ModDate:      time.Now(),
		}, nil
	}

	return info, nil
}

// MockFileManager 模拟文件管理器
type MockFileManager struct {
	mutex       sync.RWMutex
	files       map[string][]byte
	directories map[string]bool
	errors      map[string]error
	callCounts  map[string]int
}

// NewMockFileManager 创建新的模拟文件管理器
func NewMockFileManager() *MockFileManager {
	return &MockFileManager{
		files:       make(map[string][]byte),
		directories: make(map[string]bool),
		errors:      make(map[string]error),
		callCounts:  make(map[string]int),
	}
}

// AddFile 添加模拟文件
func (m *MockFileManager) AddFile(path string, content []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.files[path] = content
}

// AddDirectory 添加模拟目录
func (m *MockFileManager) AddDirectory(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.directories[path] = true
}

// SetError 设置操作错误
func (m *MockFileManager) SetError(path string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors[path] = err
}

// GetCallCount 获取调用次数
func (m *MockFileManager) GetCallCount(method string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.callCounts[method]
}

// ValidateFile 模拟验证文件
func (m *MockFileManager) ValidateFile(filePath string) (*model.FileEntry, error) {
	m.mutex.Lock()
	m.callCounts["ValidateFile"]++
	content, exists := m.files[filePath]
	err := m.errors[filePath]
	m.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	return &model.FileEntry{
		Path:        filePath,
		DisplayName: filePath,
		Size:        int64(len(content)),
		PageCount:   10,
		IsValid:     true,
		IsEncrypted: false,
	}, nil
}

// CreateTempFile 模拟创建临时文件
func (m *MockFileManager) CreateTempFile(prefix string) (string, error) {
	m.mutex.Lock()
	m.callCounts["CreateTempFile"]++
	m.mutex.Unlock()

	tempPath := fmt.Sprintf("/tmp/%s_%d", prefix, time.Now().UnixNano())
	m.AddFile(tempPath, []byte("temp content"))
	return tempPath, nil
}

// CreateTempFileWithContent 模拟创建带内容的临时文件
func (m *MockFileManager) CreateTempFileWithContent(prefix string, content []byte) (string, error) {
	m.mutex.Lock()
	m.callCounts["CreateTempFileWithContent"]++
	m.mutex.Unlock()

	tempPath := fmt.Sprintf("/tmp/%s_%d", prefix, time.Now().UnixNano())
	m.AddFile(tempPath, content)
	return tempPath, nil
}

// CopyToTempFile 模拟复制到临时文件
func (m *MockFileManager) CopyToTempFile(srcPath, prefix string) (string, error) {
	m.mutex.Lock()
	m.callCounts["CopyToTempFile"]++
	content, exists := m.files[srcPath]
	m.mutex.Unlock()

	if !exists {
		return "", fmt.Errorf("源文件不存在: %s", srcPath)
	}

	return m.CreateTempFileWithContent(prefix, content)
}

// CleanupTempFiles 模拟清理临时文件
func (m *MockFileManager) CleanupTempFiles() error {
	m.mutex.Lock()
	m.callCounts["CleanupTempFiles"]++
	// 清理所有/tmp/开头的文件
	for path := range m.files {
		if len(path) > 5 && path[:5] == "/tmp/" {
			delete(m.files, path)
		}
	}
	m.mutex.Unlock()
	return nil
}

// RemoveTempFile 模拟删除临时文件
func (m *MockFileManager) RemoveTempFile(filePath string) error {
	m.mutex.Lock()
	m.callCounts["RemoveTempFile"]++
	delete(m.files, filePath)
	m.mutex.Unlock()
	return nil
}

// GetFileInfo 模拟获取文件信息
func (m *MockFileManager) GetFileInfo(filePath string) (*file.FileInfo, error) {
	m.mutex.Lock()
	m.callCounts["GetFileInfo"]++
	content, exists := m.files[filePath]
	m.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	return &file.FileInfo{
		Path:    filePath,
		Size:    int64(len(content)),
		Name:    filePath,
		IsValid: true,
	}, nil
}

// EnsureDirectoryExists 模拟确保目录存在
func (m *MockFileManager) EnsureDirectoryExists(dirPath string) error {
	m.mutex.Lock()
	m.callCounts["EnsureDirectoryExists"]++
	m.directories[dirPath] = true
	m.mutex.Unlock()
	return nil
}

// GetTempDir 模拟获取临时目录
func (m *MockFileManager) GetTempDir() string {
	return "/tmp"
}

// WriteFile 模拟写文件
func (m *MockFileManager) WriteFile(filePath string, data []byte) error {
	m.mutex.Lock()
	m.callCounts["WriteFile"]++
	m.files[filePath] = data
	m.mutex.Unlock()
	return nil
}

// ReadFile 模拟读文件
func (m *MockFileManager) ReadFile(filePath string) ([]byte, error) {
	m.mutex.Lock()
	m.callCounts["ReadFile"]++
	content, exists := m.files[filePath]
	m.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	return content, nil
}

// CopyFile 模拟复制文件
func (m *MockFileManager) CopyFile(srcPath, dstPath string) error {
	content, err := m.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return m.WriteFile(dstPath, content)
}

// MockProgressCallback 模拟进度回调
type MockProgressCallback struct {
	mutex     sync.RWMutex
	updates   []float64
	statuses  []string
	details   []string
	callCount int
}

// NewMockProgressCallback 创建新的模拟进度回调
func NewMockProgressCallback() *MockProgressCallback {
	return &MockProgressCallback{
		updates:  make([]float64, 0),
		statuses: make([]string, 0),
		details:  make([]string, 0),
	}
}

// OnProgress 进度更新回调
func (m *MockProgressCallback) OnProgress(progress float64, status, detail string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.updates = append(m.updates, progress)
	m.statuses = append(m.statuses, status)
	m.details = append(m.details, detail)
	m.callCount++
}

// GetUpdates 获取所有进度更新
func (m *MockProgressCallback) GetUpdates() []float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]float64, len(m.updates))
	copy(result, m.updates)
	return result
}

// GetStatuses 获取所有状态更新
func (m *MockProgressCallback) GetStatuses() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]string, len(m.statuses))
	copy(result, m.statuses)
	return result
}

// GetDetails 获取所有详情更新
func (m *MockProgressCallback) GetDetails() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]string, len(m.details))
	copy(result, m.details)
	return result
}

// GetCallCount 获取调用次数
func (m *MockProgressCallback) GetCallCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.callCount
}

// Reset 重置回调数据
func (m *MockProgressCallback) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.updates = m.updates[:0]
	m.statuses = m.statuses[:0]
	m.details = m.details[:0]
	m.callCount = 0
}

// MockErrorHandler 模拟错误处理器
type MockErrorHandler struct {
	mutex  sync.RWMutex
	errors []error
	count  int
}

// NewMockErrorHandler 创建新的模拟错误处理器
func NewMockErrorHandler() *MockErrorHandler {
	return &MockErrorHandler{
		errors: make([]error, 0),
	}
}

// OnError 错误回调
func (m *MockErrorHandler) OnError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = append(m.errors, err)
	m.count++
}

// GetErrors 获取所有错误
func (m *MockErrorHandler) GetErrors() []error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]error, len(m.errors))
	copy(result, m.errors)
	return result
}

// GetErrorCount 获取错误数量
func (m *MockErrorHandler) GetErrorCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.count
}

// HasError 检查是否有错误
func (m *MockErrorHandler) HasError() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.count > 0
}

// GetLastError 获取最后一个错误
func (m *MockErrorHandler) GetLastError() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.errors) == 0 {
		return nil
	}
	return m.errors[len(m.errors)-1]
}

// Reset 重置错误处理器
func (m *MockErrorHandler) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = m.errors[:0]
	m.count = 0
}

// MockUIStateHandler 模拟UI状态处理器
type MockUIStateHandler struct {
	mutex        sync.RWMutex
	states       []bool
	callCount    int
	currentState bool
}

// NewMockUIStateHandler 创建新的模拟UI状态处理器
func NewMockUIStateHandler() *MockUIStateHandler {
	return &MockUIStateHandler{
		states: make([]bool, 0),
	}
}

// OnUIStateChange UI状态变更回调
func (m *MockUIStateHandler) OnUIStateChange(enabled bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.states = append(m.states, enabled)
	m.currentState = enabled
	m.callCount++
}

// GetStates 获取所有状态变更
func (m *MockUIStateHandler) GetStates() []bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]bool, len(m.states))
	copy(result, m.states)
	return result
}

// GetCurrentState 获取当前状态
func (m *MockUIStateHandler) GetCurrentState() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentState
}

// GetCallCount 获取调用次数
func (m *MockUIStateHandler) GetCallCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.callCount
}

// Reset 重置状态处理器
func (m *MockUIStateHandler) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.states = m.states[:0]
	m.callCount = 0
	m.currentState = false
}

// MockCompletionHandler 模拟完成处理器
type MockCompletionHandler struct {
	mutex     sync.RWMutex
	messages  []string
	callCount int
	completed bool
}

// NewMockCompletionHandler 创建新的模拟完成处理器
func NewMockCompletionHandler() *MockCompletionHandler {
	return &MockCompletionHandler{
		messages: make([]string, 0),
	}
}

// OnCompletion 完成回调
func (m *MockCompletionHandler) OnCompletion(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, message)
	m.callCount++
	m.completed = true
}

// GetMessages 获取所有完成消息
func (m *MockCompletionHandler) GetMessages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

// GetCallCount 获取调用次数
func (m *MockCompletionHandler) GetCallCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.callCount
}

// IsCompleted 检查是否已完成
func (m *MockCompletionHandler) IsCompleted() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.completed
}

// GetLastMessage 获取最后一条消息
func (m *MockCompletionHandler) GetLastMessage() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.messages) == 0 {
		return ""
	}
	return m.messages[len(m.messages)-1]
}

// Reset 重置完成处理器
func (m *MockCompletionHandler) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = m.messages[:0]
	m.callCount = 0
	m.completed = false
}

// MockReader 模拟读取器
type MockReader struct {
	data   []byte
	pos    int
	closed bool
	err    error
}

// NewMockReader 创建新的模拟读取器
func NewMockReader(data []byte) *MockReader {
	return &MockReader{
		data: data,
		pos:  0,
	}
}

// SetError 设置读取错误
func (m *MockReader) SetError(err error) {
	m.err = err
}

// Read 读取数据
func (m *MockReader) Read(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}

	if m.closed {
		return 0, fmt.Errorf("reader is closed")
	}

	if m.pos >= len(m.data) {
		return 0, io.EOF
	}

	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// Close 关闭读取器
func (m *MockReader) Close() error {
	m.closed = true
	return nil
}

// MockWriter 模拟写入器
type MockWriter struct {
	data   []byte
	closed bool
	err    error
}

// NewMockWriter 创建新的模拟写入器
func NewMockWriter() *MockWriter {
	return &MockWriter{
		data: make([]byte, 0),
	}
}

// SetError 设置写入错误
func (m *MockWriter) SetError(err error) {
	m.err = err
}

// Write 写入数据
func (m *MockWriter) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}

	if m.closed {
		return 0, fmt.Errorf("writer is closed")
	}

	m.data = append(m.data, p...)
	return len(p), nil
}

// Close 关闭写入器
func (m *MockWriter) Close() error {
	m.closed = true
	return nil
}

// GetData 获取写入的数据
func (m *MockWriter) GetData() []byte {
	result := make([]byte, len(m.data))
	copy(result, m.data)
	return result
}

// Reset 重置写入器
func (m *MockWriter) Reset() {
	m.data = m.data[:0]
	m.closed = false
	m.err = nil
}
