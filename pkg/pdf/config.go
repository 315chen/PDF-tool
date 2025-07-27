package pdf

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

// PDFServiceConfig 全局功能开关与配置
//
type PDFServiceConfig struct {
	UsePDFCPU      bool   `json:"use_pdfcpu"`      // 是否启用pdfcpu（否则用UniPDF）
	EnableLogging  bool   `json:"enable_logging"`  // 是否启用详细日志
	EnableMetrics  bool   `json:"enable_metrics"`  // 是否启用指标采集
	EnableBackup   bool   `json:"enable_backup"`   // 是否启用写入备份
	ConfigFilePath string `json:"-"`               // 配置文件路径（不序列化）
	mutex          sync.RWMutex
}

// 默认配置
func DefaultPDFServiceConfig() *PDFServiceConfig {
	return &PDFServiceConfig{
		UsePDFCPU:      true,
		EnableLogging:  true,
		EnableMetrics:  true,
		EnableBackup:   true,
		ConfigFilePath: "pdf_service_config.json",
	}
}

// LoadConfig 从文件加载配置
func (c *PDFServiceConfig) LoadConfig(path string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

// SaveConfig 保存配置到文件
func (c *PDFServiceConfig) SaveConfig(path string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// SetUsePDFCPU 切换pdfcpu/UniPDF
func (c *PDFServiceConfig) SetUsePDFCPU(use bool) {
	c.mutex.Lock()
	c.UsePDFCPU = use
	c.mutex.Unlock()
}

// IsPDFCPUEnabled 查询当前是否启用pdfcpu
func (c *PDFServiceConfig) IsPDFCPUEnabled() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.UsePDFCPU
}

// 热重载配置（简单实现）
func (c *PDFServiceConfig) Reload() error {
	if c.ConfigFilePath == "" {
		return nil
	}
	return c.LoadConfig(c.ConfigFilePath)
}

// 全局单例（可选）
var GlobalPDFServiceConfig = DefaultPDFServiceConfig()

// InitConfigFromFile 初始化全局配置
func InitConfigFromFile(path string) error {
	return GlobalPDFServiceConfig.LoadConfig(path)
} 