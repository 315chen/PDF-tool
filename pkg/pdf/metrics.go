package pdf

import (
	"runtime"
	"sync"
	"time"
)

// MigrationMetrics 迁移/处理指标
//
type MigrationMetrics struct {
	StartTime      time.Time     // 任务开始时间
	EndTime        time.Time     // 任务结束时间
	FilesProcessed int           // 处理文件数
	PagesProcessed int           // 处理页数
	BytesProcessed int64         // 处理字节数
	ErrorCount     int           // 错误次数
	LastError      error         // 最后一次错误
	MemoryUsage    uint64        // 峰值内存占用（字节）
	Custom         map[string]interface{} // 其他自定义指标
	mutex          sync.Mutex
}

// NewMigrationMetrics 创建新指标对象
func NewMigrationMetrics() *MigrationMetrics {
	return &MigrationMetrics{
		StartTime: time.Now(),
		Custom:    make(map[string]interface{}),
	}
}

// MarkEnd 记录结束时间
func (m *MigrationMetrics) MarkEnd() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.EndTime = time.Now()
}

// AddFile 增加处理文件数
func (m *MigrationMetrics) AddFile() {
	m.mutex.Lock()
	m.FilesProcessed++
	m.mutex.Unlock()
}

// AddPage 增加处理页数
func (m *MigrationMetrics) AddPage(n int) {
	m.mutex.Lock()
	m.PagesProcessed += n
	m.mutex.Unlock()
}

// AddBytes 增加处理字节数
func (m *MigrationMetrics) AddBytes(n int64) {
	m.mutex.Lock()
	m.BytesProcessed += n
	m.mutex.Unlock()
}

// AddError 记录错误
func (m *MigrationMetrics) AddError(err error) {
	m.mutex.Lock()
	m.ErrorCount++
	m.LastError = err
	m.mutex.Unlock()
}

// UpdateMemoryUsage 记录当前内存占用
func (m *MigrationMetrics) UpdateMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.mutex.Lock()
	if memStats.Alloc > m.MemoryUsage {
		m.MemoryUsage = memStats.Alloc
	}
	m.mutex.Unlock()
}

// SetCustom 设置自定义指标
func (m *MigrationMetrics) SetCustom(key string, value interface{}) {
	m.mutex.Lock()
	m.Custom[key] = value
	m.mutex.Unlock()
}

// GetDuration 获取处理总时长
func (m *MigrationMetrics) GetDuration() time.Duration {
	if m.EndTime.IsZero() {
		return time.Since(m.StartTime)
	}
	return m.EndTime.Sub(m.StartTime)
}

// Report 生成指标报告
func (m *MigrationMetrics) Report() map[string]interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return map[string]interface{}{
		"start_time":      m.StartTime,
		"end_time":        m.EndTime,
		"duration":        m.GetDuration(),
		"files_processed": m.FilesProcessed,
		"pages_processed": m.PagesProcessed,
		"bytes_processed": m.BytesProcessed,
		"error_count":     m.ErrorCount,
		"last_error":      m.LastError,
		"memory_usage":    m.MemoryUsage,
		"custom":          m.Custom,
	}
} 