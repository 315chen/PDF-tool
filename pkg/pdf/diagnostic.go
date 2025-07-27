package pdf

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// PDFDiagnosticReport PDF文件诊断报告
//
type PDFDiagnosticReport struct {
	FilePath         string                 // 文件路径
	Exists           bool                   // 文件是否存在
	Size             int64                  // 文件大小
	IsPDF            bool                   // 是否为PDF格式
	Encrypted        bool                   // 是否加密
	Permissions      string                 // 权限信息
	PageCount        int                    // 页数
	ValidationError  error                  // 格式/结构校验错误
	Compatibility    string                 // 兼容性建议
	PerformanceTips  []string               // 性能建议
	Extra            map[string]interface{} // 其他信息
	GeneratedAt      time.Time              // 诊断时间
}

// DiagnosePDF 对单个PDF文件进行诊断
func DiagnosePDF(filePath string) *PDFDiagnosticReport {
	report := &PDFDiagnosticReport{
		FilePath:    filePath,
		Exists:      false,
		IsPDF:       false,
		Encrypted:   false,
		Permissions: "",
		PageCount:   0,
		Extra:       make(map[string]interface{}),
		GeneratedAt: time.Now(),
	}

	info, err := os.Stat(filePath)
	if err != nil {
		report.ValidationError = fmt.Errorf("文件不存在: %v", err)
		return report
	}
	report.Exists = true
	report.Size = info.Size()

	// 简单判断PDF格式
	f, err := os.Open(filePath)
	if err != nil {
		report.ValidationError = fmt.Errorf("无法打开文件: %v", err)
		return report
	}
	defer f.Close()
	buf := make([]byte, 5)
	_, err = f.Read(buf)
	if err != nil || string(buf) != "%PDF-" {
		report.ValidationError = fmt.Errorf("不是有效的PDF文件头")
		return report
	}
	report.IsPDF = true

	// 检查加密、页数、权限等（调用现有接口）
	encrypted, _ := false, error(nil)
	if pdfDecryptor, ok := any(NewPDFDecryptor(nil)).(*PDFDecryptor); ok {
		encrypted, _ = pdfDecryptor.IsPDFEncrypted(filePath)
	}
	report.Encrypted = encrypted

	// 页数、权限等可进一步集成PDFInfo等接口
	// 这里只做占位
	report.PageCount = 0
	report.Permissions = "未知"

	// 性能建议
	if report.Size > 10*1024*1024 {
		report.PerformanceTips = append(report.PerformanceTips, "大文件建议分批处理或开启流式模式")
	}
	if encrypted {
		report.PerformanceTips = append(report.PerformanceTips, "加密PDF解密后再合并可提升性能")
	}

	return report
}

// SystemDiagnosticReport 系统环境诊断报告
//
type SystemDiagnosticReport struct {
	GoVersion      string
	OS             string
	Arch           string
	NumCPU         int
	PDFCPUPresent  bool
	PDFCPUVersion  string
	OtherChecks    map[string]interface{}
	GeneratedAt    time.Time
}

// DiagnoseSystem 检查系统环境
func DiagnoseSystem() *SystemDiagnosticReport {
	report := &SystemDiagnosticReport{
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		PDFCPUPresent: true, // 占位，实际可检测CLI或库
		PDFCPUVersion: "v0.11.0", // 可集成真实版本检测
		OtherChecks:   make(map[string]interface{}),
		GeneratedAt:   time.Now(),
	}
	// 可扩展更多依赖/环境检查
	return report
} 