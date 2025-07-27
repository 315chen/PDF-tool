package pdf

import (
	"fmt"
	"log"
)

// PDFCPUAvailability 检查pdfcpu库的可用性
type PDFCPUAvailability struct {
	isAvailable bool
	version     string
	error       error
}

// CheckPDFCPUAvailability 检查pdfcpu是否可用
func CheckPDFCPUAvailability() *PDFCPUAvailability {
	availability := &PDFCPUAvailability{
		isAvailable: false,
		version:     "unknown",
	}

	// 首先检查CLI版本是否可用
	if cliAdapter, err := NewPDFCPUCLIAdapter(); err == nil {
		if cliAdapter.IsAvailable() {
			if version, err := cliAdapter.GetVersion(); err == nil {
				availability.isAvailable = true
				availability.version = version + " (CLI)"
				cliAdapter.Close()
				return availability
			}
		}
		cliAdapter.Close()
	}

	// 尝试导入pdfcpu包
	defer func() {
		if r := recover(); r != nil {
			availability.error = fmt.Errorf("pdfcpu import failed: %v", r)
			availability.isAvailable = false
		}
	}()

	// TODO: 当网络恢复时，取消注释以下代码来检查pdfcpu Go库
	/*
		// 尝试导入pdfcpu
		import (
			"github.com/pdfcpu/pdfcpu/pkg/api"
			"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
		)

		// 如果能成功导入，设置为可用
		availability.isAvailable = true
		availability.version = pdfcpu.VersionStr + " (Go Library)"
	*/

	// 如果CLI和Go库都不可用
	if !availability.isAvailable {
		availability.error = fmt.Errorf("pdfcpu not available: CLI not found and Go library not installed")
	}

	return availability
}

// IsAvailable 返回pdfcpu是否可用
func (a *PDFCPUAvailability) IsAvailable() bool {
	return a.isAvailable
}

// GetVersion 返回pdfcpu版本
func (a *PDFCPUAvailability) GetVersion() string {
	return a.version
}

// GetError 返回错误信息
func (a *PDFCPUAvailability) GetError() error {
	return a.error
}

// LogStatus 记录pdfcpu状态
func (a *PDFCPUAvailability) LogStatus(logger *log.Logger) {
	if a.isAvailable {
		logger.Printf("pdfcpu is available (version: %s)", a.version)
	} else {
		logger.Printf("pdfcpu is not available: %v", a.error)
	}
}

// GetFallbackMessage 获取回退消息
func (a *PDFCPUAvailability) GetFallbackMessage() string {
	if a.isAvailable {
		return ""
	}
	return "Using placeholder implementation. Install pdfcpu for full functionality."
}

// ShouldUseFallback 是否应该使用回退实现
func (a *PDFCPUAvailability) ShouldUseFallback() bool {
	return !a.isAvailable
}
