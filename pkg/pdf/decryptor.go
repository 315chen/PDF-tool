package pdf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PDFDecryptor 提供PDF文件解密功能
type PDFDecryptor struct {
	tempDir           string
	commonPasswords   []string
	maxAttempts       int
	attemptDelay      time.Duration
	tempFiles         []string
	mutex             sync.Mutex
	progressCallback  func(current, total int, password string)
	adapter           *PDFCPUAdapter // 新增pdfcpu适配器
}

// DecryptorOptions 解密器选项
type DecryptorOptions struct {
	TempDirectory     string        // 临时文件目录
	CommonPasswords   []string      // 常用密码列表
	MaxAttempts       int           // 最大尝试次数
	AttemptDelay      time.Duration // 尝试间隔
	ProgressCallback  func(current, total int, password string) // 进度回调
}

// DecryptResult 解密结果
type DecryptResult struct {
	Success         bool
	DecryptedPath   string
	UsedPassword    string
	AttemptCount    int
	ProcessingTime  time.Duration
	IsOriginalFile  bool // 是否为原始文件（未加密）
}

// NewPDFDecryptor 创建一个新的PDF解密器
func NewPDFDecryptor(options *DecryptorOptions) *PDFDecryptor {
	if options == nil {
		options = &DecryptorOptions{
			TempDirectory: os.TempDir(),
			MaxAttempts:   100,
			AttemptDelay:  time.Millisecond * 100,
		}
	}

	adapter, _ := NewPDFCPUAdapter(&PDFCPUConfig{
		TempDirectory: options.TempDirectory,
	})

	decryptor := &PDFDecryptor{
		tempDir:          options.TempDirectory,
		commonPasswords:  options.CommonPasswords,
		maxAttempts:      options.MaxAttempts,
		attemptDelay:     options.AttemptDelay,
		progressCallback: options.ProgressCallback,
		tempFiles:        make([]string, 0),
		adapter:          adapter,
	}

	// 如果没有提供常用密码，使用默认列表
	if len(decryptor.commonPasswords) == 0 {
		decryptor.commonPasswords = getDefaultCommonPasswords()
	}

	return decryptor
}

// getDefaultCommonPasswords 获取默认常用密码列表
func getDefaultCommonPasswords() []string {
	return []string{
		"", // 空密码
		"123456",
		"password",
		"123456789",
		"12345678",
		"12345",
		"1234567",
		"1234567890",
		"qwerty",
		"abc123",
		"111111",
		"123123",
		"admin",
		"letmein",
		"welcome",
		"monkey",
		"1234",
		"dragon",
		"pass",
		"master",
		"hello",
		"freedom",
		"whatever",
		"qazwsx",
		"trustno1",
		"jordan23",
		"harley",
		"robert",
		"matthew",
		"jordan",
		"michelle",
		"daniel",
		"andrew",
		"martin",
		"joshua",
		"franklin",
		"hannah",
		"camila",
		"amanda",
		"jeremy",
		"justin",
		"melissa",
		"sarah",
		"heather",
		"nicole",
		"ginger",
		"stephanie",
		"thomas",
		"anthony",
		"charles",
		"patricia",
		"jennifer",
		"linda",
		"helen",
		"margaret",
		"ruth",
		"sharon",
		"michelle",
		"laura",
		"sarah",
		"kimberly",
		"deborah",
		"jessica",
		"shirley",
		"cynthia",
		"angela",
		"melissa",
		"brenda",
		"emma",
		"olivia",
		"ava",
		"isabella",
		"sophia",
		"charlotte",
		"mia",
		"amelia",
		"harper",
		"evelyn",
		"abigail",
		"emily",
		"elizabeth",
		"mila",
		"ella",
		"avery",
		"sofia",
		"camila",
		"aria",
		"scarlett",
		"victoria",
		"madison",
		"luna",
		"grace",
		"chloe",
		"penelope",
		"layla",
		"riley",
		"zoey",
		"nora",
		"lily",
		"eleanor",
		"hannah",
		"lillian",
		"addison",
		"aubrey",
		"ellie",
		"stella",
		"natalie",
		"zoe",
	}
}

// DecryptPDF 使用指定密码解密PDF文件
func (d *PDFDecryptor) DecryptPDF(filePath, password string) (string, error) {
	// 生成临时输出文件路径
	outputPath := d.generateTempFilePath(filePath)

	// 调用pdfcpu适配器进行解密
	err := d.adapter.DecryptFile(filePath, outputPath, password)
	if err != nil {
		return "", &PDFError{
			Type:    ErrorEncrypted,
			Message: "pdfcpu解密失败",
			File:    filePath,
			Cause:   err,
		}
	}

	return outputPath, nil
}

// AutoDecrypt 自动解密PDF文件
func (d *PDFDecryptor) AutoDecrypt(filePath string) (*DecryptResult, error) {
	startTime := time.Now()
	result := &DecryptResult{
		Success:        false,
		ProcessingTime: 0,
	}

	// 检查文件是否加密
	isEncrypted, err := d.IsPDFEncrypted(filePath)
	if err != nil {
		return result, err
	}

	if !isEncrypted {
		// 文件未加密，直接返回原文件路径
		result.Success = true
		result.DecryptedPath = filePath
		result.IsOriginalFile = true
		result.ProcessingTime = time.Since(startTime)
		return result, nil
	}

	// 使用常用密码列表进行自动解密
	totalPasswords := len(d.commonPasswords)
	if totalPasswords > d.maxAttempts {
		totalPasswords = d.maxAttempts
	}

	for i, password := range d.commonPasswords {
		if i >= d.maxAttempts {
			break
		}

		// 调用进度回调
		if d.progressCallback != nil {
			d.progressCallback(i+1, totalPasswords, password)
		}

		// 尝试解密
		decryptedPath, err := d.DecryptPDF(filePath, password)
		result.AttemptCount = i + 1

		if err == nil {
			// 解密成功
			result.Success = true
			result.DecryptedPath = decryptedPath
			result.UsedPassword = password
			result.ProcessingTime = time.Since(startTime)
			
			// 记录临时文件以便后续清理
			d.addTempFile(decryptedPath)
			return result, nil
		}

		// 检查是否是密码错误（可以继续尝试）还是其他错误（应该停止）
		if pdfErr, ok := err.(*PDFError); ok {
			if pdfErr.Type != ErrorEncrypted {
				// 非密码错误，停止尝试
				result.ProcessingTime = time.Since(startTime)
				return result, err
			}
		}

		// 添加延迟避免过快尝试
		if d.attemptDelay > 0 {
			time.Sleep(d.attemptDelay)
		}
	}

	// 所有密码都失败
	result.ProcessingTime = time.Since(startTime)
	return result, &PDFError{
		Type:    ErrorEncrypted,
		Message: fmt.Sprintf("无法使用 %d 个常用密码解密文件", result.AttemptCount),
		File:    filePath,
	}
}

// TryDecryptWithPasswords 尝试使用指定密码列表解密PDF文件
func (d *PDFDecryptor) TryDecryptWithPasswords(filePath string, passwords []string) (*DecryptResult, error) {
	startTime := time.Now()
	result := &DecryptResult{
		Success:        false,
		ProcessingTime: 0,
	}

	// 检查文件是否加密
	isEncrypted, err := d.IsPDFEncrypted(filePath)
	if err != nil {
		return result, err
	}

	if !isEncrypted {
		// 文件未加密，直接返回原文件路径
		result.Success = true
		result.DecryptedPath = filePath
		result.IsOriginalFile = true
		result.ProcessingTime = time.Since(startTime)
		return result, nil
	}

	// 尝试使用每个密码解密
	totalPasswords := len(passwords)
	for i, password := range passwords {
		// 调用进度回调
		if d.progressCallback != nil {
			d.progressCallback(i+1, totalPasswords, password)
		}

		decryptedPath, err := d.DecryptPDF(filePath, password)
		result.AttemptCount = i + 1

		if err == nil {
			// 解密成功
			result.Success = true
			result.DecryptedPath = decryptedPath
			result.UsedPassword = password
			result.ProcessingTime = time.Since(startTime)
			
			// 记录临时文件以便后续清理
			d.addTempFile(decryptedPath)
			return result, nil
		}

		// 检查是否是密码错误
		if pdfErr, ok := err.(*PDFError); ok {
			if pdfErr.Type != ErrorEncrypted {
				// 非密码错误，停止尝试
				result.ProcessingTime = time.Since(startTime)
				return result, err
			}
		}

		// 添加延迟
		if d.attemptDelay > 0 {
			time.Sleep(d.attemptDelay)
		}
	}

	// 所有密码都失败
	result.ProcessingTime = time.Since(startTime)
	return result, &PDFError{
		Type:    ErrorEncrypted,
		Message: fmt.Sprintf("无法使用提供的 %d 个密码解密文件", len(passwords)),
		File:    filePath,
	}
}

// TryDecryptPDF 尝试使用多个密码解密PDF文件（保持向后兼容）
func (d *PDFDecryptor) TryDecryptPDF(filePath string, passwords []string) (string, string, error) {
	result, err := d.TryDecryptWithPasswords(filePath, passwords)
	if err != nil {
		return "", "", err
	}
	
	if result.Success {
		return result.DecryptedPath, result.UsedPassword, nil
	}
	
	return "", "", &PDFError{
		Type:    ErrorEncrypted,
		Message: "解密失败",
		File:    filePath,
	}
}

// IsPDFEncrypted 检查PDF文件是否加密
func (d *PDFDecryptor) IsPDFEncrypted(filePath string) (bool, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return false, &PDFError{
			Type:    ErrorIO,
			Message: "无法打开文件",
			File:    filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 使用pdfcpu适配器检查加密状态
	isEncrypted, err := d.adapter.IsEncrypted(filePath)
	if err != nil {
		return false, &PDFError{
			Type:    ErrorCorrupted,
			Message: "无法确定加密状态",
			File:    filePath,
			Cause:   err,
		}
	}

	return isEncrypted, nil
}

// generateTempFilePath 生成临时文件路径
func (d *PDFDecryptor) generateTempFilePath(originalPath string) string {
	// 确保临时目录存在
	if d.tempDir == "" {
		d.tempDir = os.TempDir()
	}
	
	if _, err := os.Stat(d.tempDir); os.IsNotExist(err) {
		os.MkdirAll(d.tempDir, 0755)
	}

	// 获取原始文件名
	fileName := filepath.Base(originalPath)
	
	// 生成临时文件路径
	return filepath.Join(d.tempDir, "decrypted_"+fileName)
}

// addTempFile 添加临时文件到清理列表
func (d *PDFDecryptor) addTempFile(filePath string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// 避免重复添加
	for _, existing := range d.tempFiles {
		if existing == filePath {
			return
		}
	}
	
	d.tempFiles = append(d.tempFiles, filePath)
}

// CleanupTempFiles 清理所有临时文件
func (d *PDFDecryptor) CleanupTempFiles() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	var errors []string
	
	for _, filePath := range d.tempFiles {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("无法删除临时文件 %s: %v", filePath, err))
		}
	}
	
	// 清空临时文件列表
	d.tempFiles = d.tempFiles[:0]
	
	if len(errors) > 0 {
		return &PDFError{
			Type:    ErrorIO,
			Message: fmt.Sprintf("清理临时文件时出现错误: %s", strings.Join(errors, "; ")),
		}
	}
	
	return nil
}

// GetTempFiles 获取当前临时文件列表
func (d *PDFDecryptor) GetTempFiles() []string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// 返回副本避免并发问题
	result := make([]string, len(d.tempFiles))
	copy(result, d.tempFiles)
	return result
}

// SetProgressCallback 设置进度回调函数
func (d *PDFDecryptor) SetProgressCallback(callback func(current, total int, password string)) {
	d.progressCallback = callback
}

// GetCommonPasswords 获取常用密码列表
func (d *PDFDecryptor) GetCommonPasswords() []string {
	// 返回副本避免外部修改
	result := make([]string, len(d.commonPasswords))
	copy(result, d.commonPasswords)
	return result
}

// SetCommonPasswords 设置常用密码列表
func (d *PDFDecryptor) SetCommonPasswords(passwords []string) {
	d.commonPasswords = make([]string, len(passwords))
	copy(d.commonPasswords, passwords)
}

// AddCommonPassword 添加常用密码
func (d *PDFDecryptor) AddCommonPassword(password string) {
	// 检查是否已存在
	for _, existing := range d.commonPasswords {
		if existing == password {
			return
		}
	}
	
	d.commonPasswords = append(d.commonPasswords, password)
}

// RemoveCommonPassword 移除常用密码
func (d *PDFDecryptor) RemoveCommonPassword(password string) {
	for i, existing := range d.commonPasswords {
		if existing == password {
			d.commonPasswords = append(d.commonPasswords[:i], d.commonPasswords[i+1:]...)
			return
		}
	}
}

// GetMaxAttempts 获取最大尝试次数
func (d *PDFDecryptor) GetMaxAttempts() int {
	return d.maxAttempts
}

// SetMaxAttempts 设置最大尝试次数
func (d *PDFDecryptor) SetMaxAttempts(maxAttempts int) {
	if maxAttempts > 0 {
		d.maxAttempts = maxAttempts
	}
}

// GetAttemptDelay 获取尝试延迟
func (d *PDFDecryptor) GetAttemptDelay() time.Duration {
	return d.attemptDelay
}

// SetAttemptDelay 设置尝试延迟
func (d *PDFDecryptor) SetAttemptDelay(delay time.Duration) {
	d.attemptDelay = delay
}

// DecryptWithProgress 带进度显示的解密
func (d *PDFDecryptor) DecryptWithProgress(filePath string, progressWriter io.Writer) (*DecryptResult, error) {
	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "开始自动解密文件: %s\n", filepath.Base(filePath))
	}
	
	// 设置临时进度回调
	originalCallback := d.progressCallback
	d.progressCallback = func(current, total int, password string) {
		if progressWriter != nil {
			if password == "" {
				fmt.Fprintf(progressWriter, "尝试空密码 (%d/%d)\n", current, total)
			} else {
				fmt.Fprintf(progressWriter, "尝试密码: %s (%d/%d)\n", password, current, total)
			}
		}
		
		// 调用原始回调
		if originalCallback != nil {
			originalCallback(current, total, password)
		}
	}
	
	// 执行自动解密
	result, err := d.AutoDecrypt(filePath)
	
	// 恢复原始回调
	d.progressCallback = originalCallback
	
	if progressWriter != nil {
		if result.Success {
			if result.IsOriginalFile {
				fmt.Fprintf(progressWriter, "文件未加密，无需解密\n")
			} else {
				fmt.Fprintf(progressWriter, "解密成功！使用密码: %s，尝试次数: %d，用时: %v\n", 
					result.UsedPassword, result.AttemptCount, result.ProcessingTime)
			}
		} else {
			fmt.Fprintf(progressWriter, "解密失败，尝试次数: %d，用时: %v\n", 
				result.AttemptCount, result.ProcessingTime)
		}
	}
	
	return result, err
}