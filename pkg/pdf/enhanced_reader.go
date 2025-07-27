package pdf

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationMode 定义验证模式
type ValidationMode int

const (
	// ValidationStrict 严格验证模式
	ValidationStrict ValidationMode = iota
	// ValidationRelaxed 宽松验证模式
	ValidationRelaxed
	// ValidationBasic 基本验证模式
	ValidationBasic
)

// EnhancedPDFReader 增强的PDF读取器
type EnhancedPDFReader struct {
	filePath       string
	info           *PDFInfo
	isOpen         bool
	validationMode ValidationMode
	cliAdapter     *PDFCPUCLIAdapter
	useCLI         bool
}

// NewEnhancedPDFReader 创建增强的PDF读取器
func NewEnhancedPDFReader(filePath string, mode ValidationMode) (*EnhancedPDFReader, error) {
	reader := &EnhancedPDFReader{
		filePath:       filePath,
		isOpen:         false,
		validationMode: mode,
		useCLI:         false,
	}

	// 只在严格模式下使用CLI适配器
	if mode == ValidationStrict {
		if cliAdapter, err := NewPDFCPUCLIAdapter(); err == nil && cliAdapter.IsAvailable() {
			reader.cliAdapter = cliAdapter
			reader.useCLI = true
		}
	}

	if err := reader.Open(); err != nil {
		return nil, err
	}

	return reader, nil
}

// Open 打开PDF文件
func (r *EnhancedPDFReader) Open() error {
	if r.isOpen {
		return nil
	}

	// 验证文件是否存在
	if _, err := os.Stat(r.filePath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法访问PDF文件",
			File:    r.filePath,
			Cause:   err,
		}
	}

	// 根据验证模式进行不同级别的验证
	switch r.validationMode {
	case ValidationStrict:
		if err := r.strictValidation(); err != nil {
			return err
		}
	case ValidationRelaxed:
		if err := r.relaxedValidation(); err != nil {
			return err
		}
	case ValidationBasic:
		if err := r.basicValidation(); err != nil {
			return err
		}
	}

	r.isOpen = true
	return nil
}

// strictValidation 严格验证
func (r *EnhancedPDFReader) strictValidation() error {
	if r.useCLI && r.cliAdapter != nil {
		if err := r.cliAdapter.ValidateFile(r.filePath); err != nil {
			return &PDFError{
				Type:    ErrorInvalidFile,
				Message: "PDF文件严格验证失败",
				File:    r.filePath,
				Cause:   err,
			}
		}
	}
	return r.basicValidation()
}

// relaxedValidation 宽松验证
func (r *EnhancedPDFReader) relaxedValidation() error {
	// 基本PDF头部检查
	if err := r.basicValidation(); err != nil {
		return err
	}

	// 检查是否有基本的PDF结构
	if err := r.checkBasicStructure(); err != nil {
		return err
	}

	return nil
}

// basicValidation 基本验证
func (r *EnhancedPDFReader) basicValidation() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法打开PDF文件",
			File:    r.filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 检查PDF头部
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "无法读取文件头部",
			File:    r.filePath,
			Cause:   err,
		}
	}

	if !strings.HasPrefix(string(header), "%PDF-") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "不是有效的PDF文件",
			File:    r.filePath,
		}
	}

	return nil
}

// checkBasicStructure 检查基本PDF结构
func (r *EnhancedPDFReader) checkBasicStructure() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hasObjects := false
	hasXref := false
	hasTrailer := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 检查是否有对象定义
		if strings.Contains(line, "obj") {
			hasObjects = true
		}
		
		// 检查是否有交叉引用表
		if line == "xref" {
			hasXref = true
		}
		
		// 检查是否有trailer
		if line == "trailer" {
			hasTrailer = true
		}
	}

	if !hasObjects {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件缺少对象定义",
			File:    r.filePath,
		}
	}

	// 宽松模式下，xref和trailer不是必需的
	if r.validationMode == ValidationStrict && (!hasXref || !hasTrailer) {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件结构不完整",
			File:    r.filePath,
		}
	}

	return nil
}

// GetInfo 获取PDF信息
func (r *EnhancedPDFReader) GetInfo() (*PDFInfo, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	if r.info != nil {
		return r.info, nil
	}

	// 如果使用CLI且为严格模式，从CLI获取信息
	if r.useCLI && r.cliAdapter != nil && r.validationMode == ValidationStrict {
		info, err := r.cliAdapter.GetFileInfo(r.filePath)
		if err == nil {
			r.info = info
			return r.info, nil
		}
	}

	// 回退到手动解析
	info, err := r.parseBasicInfo()
	if err != nil {
		return nil, err
	}

	r.info = info
	return r.info, nil
}

// parseBasicInfo 手动解析基本PDF信息
func (r *EnhancedPDFReader) parseBasicInfo() (*PDFInfo, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法打开文件",
			File:    r.filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    r.filePath,
			Cause:   err,
		}
	}

	info := &PDFInfo{
		FilePath:      r.filePath,
		FileSize:      fileInfo.Size(),
		CreationDate:  fileInfo.ModTime(),
		ModDate:       fileInfo.ModTime(),
		Title:         r.extractTitle(),
		Version:       "1.4", // 默认版本
		Author:        "",
		Subject:       "",
		Creator:       "",
		Producer:      "",
		PDFCPUVersion: "",
		Permissions:   []string{},
	}

	// 解析PDF版本
	if version, err := r.extractVersion(); err == nil {
		info.Version = version
	}

	// 计算页数
	if pageCount, err := r.countPages(); err == nil {
		info.PageCount = pageCount
	} else {
		info.PageCount = 1 // 默认至少1页
	}

	// 检查加密状态
	if encrypted, err := r.checkEncryption(); err == nil {
		info.IsEncrypted = encrypted
	}

	// 提取元数据
	if metadata, err := r.extractMetadata(); err == nil {
		if title, ok := metadata["Title"]; ok {
			info.Title = title
		}
		if author, ok := metadata["Author"]; ok {
			info.Author = author
		}
		if subject, ok := metadata["Subject"]; ok {
			info.Subject = subject
		}
		if creator, ok := metadata["Creator"]; ok {
			info.Creator = creator
		}
		if producer, ok := metadata["Producer"]; ok {
			info.Producer = producer
		}
	}

	return info, nil
}

// extractVersion 提取PDF版本
func (r *EnhancedPDFReader) extractVersion() (string, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 16)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	headerStr := string(header)
	re := regexp.MustCompile(`%PDF-(\d+\.\d+)`)
	matches := re.FindStringSubmatch(headerStr)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "1.4", nil // 默认版本
}

// countPages 计算页数
func (r *EnhancedPDFReader) countPages() (int, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pageCount := 0

	// 简单的页面计数方法：查找 /Type /Page
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/Type") && strings.Contains(line, "/Page") && !strings.Contains(line, "/Pages") {
			pageCount++
		}
	}

	// 如果没有找到页面，尝试查找 /Count
	if pageCount == 0 {
		file.Seek(0, 0)
		scanner = bufio.NewScanner(file)
		
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "/Count") {
				re := regexp.MustCompile(`/Count\s+(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if count, err := strconv.Atoi(matches[1]); err == nil {
						pageCount = count
						break
					}
				}
			}
		}
	}

	if pageCount == 0 {
		pageCount = 1 // 默认至少1页
	}

	return pageCount, nil
}

// checkEncryption 检查加密状态
func (r *EnhancedPDFReader) checkEncryption() (bool, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/Encrypt") || strings.Contains(line, "/Filter") {
			return true, nil
		}
	}

	return false, nil
}

// extractMetadata 提取元数据
func (r *EnhancedPDFReader) extractMetadata() (map[string]string, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metadata := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		
		// 查找元数据字段
		fields := []string{"Title", "Author", "Subject", "Creator", "Producer"}
		for _, field := range fields {
			pattern := fmt.Sprintf(`/%s\s*\(([^)]*)\)`, field)
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				metadata[field] = matches[1]
			}
		}
	}

	return metadata, nil
}

// extractTitle 提取标题
func (r *EnhancedPDFReader) extractTitle() string {
	fileName := filepath.Base(r.filePath)
	if ext := filepath.Ext(fileName); ext != "" {
		fileName = fileName[:len(fileName)-len(ext)]
	}
	return fileName
}

// Close 关闭读取器
func (r *EnhancedPDFReader) Close() error {
	if !r.isOpen {
		return nil
	}

	if r.cliAdapter != nil {
		r.cliAdapter.Close()
	}

	r.info = nil
	r.isOpen = false
	return nil
}

// GetValidationMode 获取验证模式
func (r *EnhancedPDFReader) GetValidationMode() ValidationMode {
	return r.validationMode
}

// SetValidationMode 设置验证模式
func (r *EnhancedPDFReader) SetValidationMode(mode ValidationMode) {
	r.validationMode = mode
}

// IsOpen 检查是否已打开
func (r *EnhancedPDFReader) IsOpen() bool {
	return r.isOpen
}

// GetFilePath 获取文件路径
func (r *EnhancedPDFReader) GetFilePath() string {
	return r.filePath
}

// ValidateWithMode 使用指定模式验证文件
func (r *EnhancedPDFReader) ValidateWithMode(mode ValidationMode) error {
	oldMode := r.validationMode
	r.validationMode = mode
	
	defer func() {
		r.validationMode = oldMode
	}()

	switch mode {
	case ValidationStrict:
		return r.strictValidation()
	case ValidationRelaxed:
		return r.relaxedValidation()
	case ValidationBasic:
		return r.basicValidation()
	default:
		return r.basicValidation()
	}
}
