package pdf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PDFCPUCLIAdapter 使用pdfcpu命令行工具的适配器
type PDFCPUCLIAdapter struct {
	cliPath string
	tempDir string
	logger  SimpleLogger
}

// SimpleLogger 简单的日志接口
type SimpleLogger interface {
	Printf(format string, v ...interface{})
}

// NewPDFCPUCLIAdapter 创建新的CLI适配器
func NewPDFCPUCLIAdapter() (*PDFCPUCLIAdapter, error) {
	// 检查pdfcpu命令是否可用
	cliPath, err := exec.LookPath("pdfcpu")
	if err != nil {
		return nil, fmt.Errorf("pdfcpu command not found: %w", err)
	}

	tempDir := filepath.Join(os.TempDir(), "pdfcpu-cli-adapter")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &PDFCPUCLIAdapter{
		cliPath: cliPath,
		tempDir: tempDir,
		logger:  &defaultLogger{},
	}, nil
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	fmt.Printf("[PDFCPU-CLI] "+format+"\n", v...)
}

// IsAvailable 检查pdfcpu CLI是否可用
func (a *PDFCPUCLIAdapter) IsAvailable() bool {
	cmd := exec.Command(a.cliPath, "version")
	return cmd.Run() == nil
}

// GetVersion 获取pdfcpu版本
func (a *PDFCPUCLIAdapter) GetVersion() (string, error) {
	cmd := exec.Command(a.cliPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 解析版本信息
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "pdfcpu:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "unknown", nil
}

// ValidateFile 验证PDF文件
func (a *PDFCPUCLIAdapter) ValidateFile(filePath string) error {
	a.logger.Printf("Validating PDF file using CLI: %s", filePath)

	// 使用宽松模式验证，允许修复一些常见问题
	cmd := exec.Command(a.cliPath, "validate", "-mode=relaxed", filePath)

	// 添加超时机制，避免进程卡住
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, a.cliPath, "validate", "-mode=relaxed", filePath)

	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("validation timeout after 30 seconds")
		}
		return fmt.Errorf("validation failed: %s", string(output))
	}

	a.logger.Printf("Validation successful: %s", filePath)
	return nil
}

// GetFileInfo 获取PDF文件信息
func (a *PDFCPUCLIAdapter) GetFileInfo(filePath string) (*PDFInfo, error) {
	a.logger.Printf("Getting PDF info using CLI: %s", filePath)

	// 使用pdfcpu info命令，添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.cliPath, "info", filePath)
	output, err := cmd.Output()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("info command timeout after 30 seconds")
		}
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	// 使用新的映射函数解析输出
	info := mapPDFCPUInfo(filePath, string(output))

	// 获取文件大小和时间信息
	if fileInfo, err := os.Stat(filePath); err == nil {
		info.FileSize = fileInfo.Size()
		info.CreationDate = fileInfo.ModTime()
		info.ModDate = fileInfo.ModTime()
	}

	// 设置pdfcpu版本信息
	if version, err := a.GetVersion(); err == nil {
		info.PDFCPUVersion = version
	}

	return info, nil
}

// MergeFiles 合并PDF文件
func (a *PDFCPUCLIAdapter) MergeFiles(inputFiles []string, outputFile string) error {
	a.logger.Printf("Merging %d PDF files using CLI to: %s", len(inputFiles), outputFile)

	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	// 构建命令参数: pdfcpu merge outFile inFile1 inFile2 ...
	args := []string{"merge", outputFile}
	args = append(args, inputFiles...)

	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 合并操作需要更长时间
	defer cancel()
	cmd := exec.CommandContext(ctx, a.cliPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("merge command timeout after 60 seconds")
		}
		return fmt.Errorf("merge failed: %s", string(output))
	}

	a.logger.Printf("Merge successful: %s", outputFile)
	return nil
}

// DecryptFile 解密PDF文件
func (a *PDFCPUCLIAdapter) DecryptFile(inputFile, outputFile, password string) error {
	a.logger.Printf("Decrypting PDF file using CLI: %s -> %s", inputFile, outputFile)

	cmd := exec.Command(a.cliPath, "decrypt", "-upw", password, inputFile, outputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("decryption failed: %s", string(output))
	}

	a.logger.Printf("Decryption successful: %s", outputFile)
	return nil
}

// OptimizeFile 优化PDF文件
func (a *PDFCPUCLIAdapter) OptimizeFile(inputFile, outputFile string) error {
	a.logger.Printf("Optimizing PDF file using CLI: %s -> %s", inputFile, outputFile)

	cmd := exec.Command(a.cliPath, "optimize", inputFile, outputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("optimization failed: %s", string(output))
	}

	a.logger.Printf("Optimization successful: %s", outputFile)
	return nil
}

// SplitFile 分割PDF文件
func (a *PDFCPUCLIAdapter) SplitFile(inputFile, outputDir string, pageRange string) error {
	a.logger.Printf("Splitting PDF file using CLI: %s", inputFile)

	args := []string{"split", inputFile, outputDir}
	if pageRange != "" {
		args = append(args, pageRange)
	}

	cmd := exec.Command(a.cliPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("split failed: %s", string(output))
	}

	a.logger.Printf("Split successful to: %s", outputDir)
	return nil
}

// ExtractPages 提取页面
func (a *PDFCPUCLIAdapter) ExtractPages(inputFile, outputFile string, pages string) error {
	a.logger.Printf("Extracting pages from PDF using CLI: %s", inputFile)

	cmd := exec.Command(a.cliPath, "trim", "-pages", pages, inputFile, outputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("page extraction failed: %s", string(output))
	}

	a.logger.Printf("Page extraction successful: %s", outputFile)
	return nil
}

// Close 清理资源
func (a *PDFCPUCLIAdapter) Close() error {
	a.logger.Printf("Closing PDFCPUCLIAdapter")

	// 清理临时目录
	if err := os.RemoveAll(a.tempDir); err != nil {
		a.logger.Printf("Warning: failed to clean temp directory: %v", err)
	}

	return nil
}

// SetLogger 设置日志记录器
func (a *PDFCPUCLIAdapter) SetLogger(logger SimpleLogger) {
	a.logger = logger
}

// GetTempDir 获取临时目录
func (a *PDFCPUCLIAdapter) GetTempDir() string {
	return a.tempDir
}

// CreateTestPDF 创建测试PDF文件（用于测试）
func (a *PDFCPUCLIAdapter) CreateTestPDF(outputFile string, pageCount int) error {
	a.logger.Printf("Creating test PDF with %d pages: %s", pageCount, outputFile)

	// 创建JSON配置文件
	jsonFile := outputFile + ".json"
	jsonContent := fmt.Sprintf(`{
   "pages": {`)

	for i := 1; i <= pageCount; i++ {
		if i > 1 {
			jsonContent += ","
		}
		jsonContent += fmt.Sprintf(`
      "%d": {
         "content": {
            "text": [
               {
                  "value": "Test Page %d",
                  "anchor": "center",
                  "font": {
                     "name": "Helvetica",
                     "size": 12
                   }
               }
            ]
         }
      }`, i, i)
	}

	jsonContent += `
   }
}`

	// 写入JSON文件
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer os.Remove(jsonFile) // 清理JSON文件

	// 使用pdfcpu create命令
	cmd := exec.Command(a.cliPath, "create", jsonFile, outputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("test PDF creation failed: %s", string(output))
	}

	return nil
}

// GetPermissions 获取PDF文件的详细权限信息
func (a *PDFCPUCLIAdapter) GetPermissions(filePath string) (map[string]interface{}, error) {
	a.logger.Printf("Getting PDF permissions using CLI: %s", filePath)

	// 使用pdfcpu info命令获取详细信息
	cmd := exec.Command(a.cliPath, "info", "-verbose", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return a.parsePermissionsFromOutput(string(output)), nil
}

// parsePermissionsFromOutput 解析pdfcpu输出中的权限信息
func (a *PDFCPUCLIAdapter) parsePermissionsFromOutput(output string) map[string]interface{} {
	permissions := make(map[string]interface{})
	lines := strings.Split(output, "\n")

	// 默认值
	permissions["encrypted"] = false
	permissions["permissions"] = []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"}
	permissions["has_user_password"] = false
	permissions["has_owner_password"] = false
	permissions["encryption_method"] = ""
	permissions["key_length"] = 0
	permissions["security_handler"] = ""
	permissions["filter"] = ""
	permissions["version"] = 0
	permissions["revision"] = 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Encrypted:") {
			encrypted := strings.Contains(strings.ToLower(line), "true") ||
				strings.Contains(strings.ToLower(line), "yes")
			permissions["encrypted"] = encrypted

			if encrypted {
				// 如果加密，设置默认受限权限
				permissions["permissions"] = []string{"print", "copy"}
			}
		} else if strings.HasPrefix(line, "Security handler:") {
			permissions["security_handler"] = extractStringValue(line)
		} else if strings.HasPrefix(line, "Filter:") {
			permissions["filter"] = extractStringValue(line)
		} else if strings.HasPrefix(line, "V:") {
			if version, err := extractIntValue(line); err == nil {
				permissions["version"] = version
			}
		} else if strings.HasPrefix(line, "R:") {
			if revision, err := extractIntValue(line); err == nil {
				permissions["revision"] = revision
			}
		} else if strings.HasPrefix(line, "Length:") {
			if keyLen, err := extractIntValue(line); err == nil {
				permissions["key_length"] = keyLen
			}
		} else if strings.HasPrefix(line, "P:") {
			// 解析权限标志位
			if permValue, err := extractIntValue(line); err == nil {
				permissions["permission_flags"] = permValue
				permissions["permissions"] = a.parsePermissionFlags(permValue)
			}
		} else if strings.HasPrefix(line, "User password:") {
			permissions["has_user_password"] = strings.Contains(strings.ToLower(line), "true") ||
				strings.Contains(strings.ToLower(line), "yes")
		} else if strings.HasPrefix(line, "Owner password:") {
			permissions["has_owner_password"] = strings.Contains(strings.ToLower(line), "true") ||
				strings.Contains(strings.ToLower(line), "yes")
		}
	}

	return permissions
}

// parsePermissionFlags 解析PDF权限标志位
func (a *PDFCPUCLIAdapter) parsePermissionFlags(flags int) []string {
	var permissions []string

	// PDF权限标志位定义（基于PDF规范）
	if flags&(1<<2) != 0 { // 位3：打印
		permissions = append(permissions, "print")
	}
	if flags&(1<<3) != 0 { // 位4：修改文档
		permissions = append(permissions, "modify")
	}
	if flags&(1<<4) != 0 { // 位5：复制或提取文本和图形
		permissions = append(permissions, "copy")
	}
	if flags&(1<<5) != 0 { // 位6：添加或修改注释，填写表单字段
		permissions = append(permissions, "annotate")
	}
	if flags&(1<<8) != 0 { // 位9：填写表单字段
		permissions = append(permissions, "fill")
	}
	if flags&(1<<9) != 0 { // 位10：提取文本和图形（辅助功能）
		permissions = append(permissions, "extract")
	}
	if flags&(1<<10) != 0 { // 位11：组装文档
		permissions = append(permissions, "assemble")
	}
	if flags&(1<<11) != 0 { // 位12：高质量打印
		permissions = append(permissions, "print_high")
	}

	return permissions
}

// GetSecurityDetails 获取PDF安全详细信息
func (a *PDFCPUCLIAdapter) GetSecurityDetails(filePath string) (map[string]interface{}, error) {
	a.logger.Printf("Getting PDF security details using CLI: %s", filePath)

	// 首先获取基本权限信息
	permissions, err := a.GetPermissions(filePath)
	if err != nil {
		return nil, err
	}

	// 如果文件加密，尝试获取更多安全信息
	if encrypted, ok := permissions["encrypted"].(bool); ok && encrypted {
		// 可以在这里添加更多安全信息提取逻辑
		// 例如证书信息、数字签名等
		permissions["certificate_info"] = a.getCertificateInfo(filePath)
		permissions["signature_info"] = a.getSignatureInfo(filePath)
	}

	return permissions, nil
}

// getCertificateInfo 获取证书信息（占位符实现）
func (a *PDFCPUCLIAdapter) getCertificateInfo(filePath string) map[string]interface{} {
	// TODO: 实现证书信息提取
	return map[string]interface{}{
		"has_certificates":  false,
		"certificate_count": 0,
		"certificates":      []interface{}{},
	}
}

// getSignatureInfo 获取数字签名信息（占位符实现）
func (a *PDFCPUCLIAdapter) getSignatureInfo(filePath string) map[string]interface{} {
	// TODO: 实现数字签名信息提取
	return map[string]interface{}{
		"has_signatures":  false,
		"signature_count": 0,
		"signatures":      []interface{}{},
	}
}

// GetCapabilities 获取支持的功能列表
func (a *PDFCPUCLIAdapter) GetCapabilities() []string {
	return []string{
		"validate",
		"info",
		"merge",
		"decrypt",
		"optimize",
		"split",
		"extract",
		"create",
		"permissions",
		"security",
	}
}

// ExecuteCommand 执行自定义pdfcpu命令
func (a *PDFCPUCLIAdapter) ExecuteCommand(args ...string) (string, error) {
	cmd := exec.Command(a.cliPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// IsEncrypted 检查PDF文件是否加密
func (a *PDFCPUCLIAdapter) IsEncrypted(filePath string) (bool, error) {
	a.logger.Printf("Checking encryption status using CLI: %s", filePath)

	// 使用pdfcpu info命令检查加密状态
	cmd := exec.Command(a.cliPath, "info", filePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// 如果命令失败，可能文件不存在或格式错误
		return false, fmt.Errorf("failed to check encryption: %s", string(output))
	}

	// 检查输出中是否包含加密相关信息
	outputStr := string(output)
	encryptionKeywords := []string{
		"encrypted",
		"Encrypted",
		"ENCRYPTED",
		"password",
		"Password",
		"PASSWORD",
	}

	for _, keyword := range encryptionKeywords {
		if strings.Contains(outputStr, keyword) {
			return true, nil
		}
	}

	return false, nil
}
