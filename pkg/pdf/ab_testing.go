package pdf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"runtime"
	"sync"
	"time"
)

// ABTestResult A/B测试结果
type ABTestResult struct {
	TestID         string        `json:"test_id"`
	TestName       string        `json:"test_name"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	FilesProcessed int           `json:"files_processed"`
	BytesProcessed int64         `json:"bytes_processed"`
	ErrorCount     int           `json:"error_count"`
	MemoryUsage    uint64        `json:"memory_usage"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

// ABTestComparison A/B测试对比结果
type ABTestComparison struct {
	TestID          string        `json:"test_id"`
	TestName        string        `json:"test_name"`
	PDFCPUResult    *ABTestResult `json:"pdfcpu_result"`
	UniPDFResult    *ABTestResult `json:"unipdf_result"`
	PerformanceGain float64       `json:"performance_gain"` // 性能提升百分比
	MemoryReduction float64       `json:"memory_reduction"` // 内存减少百分比
	Winner          string        `json:"winner"`           // "pdfcpu", "unipdf", "tie"
	Recommendation  string        `json:"recommendation"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

// ABTestFramework A/B测试框架
type ABTestFramework struct {
	config     *PDFServiceConfig
	results    map[string]*ABTestComparison
	mutex      sync.RWMutex
	outputPath string
}

// NewABTestFramework 创建A/B测试框架
func NewABTestFramework(config *PDFServiceConfig, outputPath string) *ABTestFramework {
	return &ABTestFramework{
		config:     config,
		results:    make(map[string]*ABTestComparison),
		outputPath: outputPath,
	}
}

// RunABTest 运行A/B测试
func (f *ABTestFramework) RunABTest(testID, testName string, testFunc func() error) (*ABTestComparison, error) {
	comparison := &ABTestComparison{
		TestID:      testID,
		TestName:    testName,
		GeneratedAt: time.Now(),
	}

	// 测试pdfcpu
	f.config.SetUsePDFCPU(true)
	pdfcpuResult, err := f.runSingleTest("pdfcpu", testFunc)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu测试失败: %v", err)
	}
	comparison.PDFCPUResult = pdfcpuResult

	// 测试UniPDF
	f.config.SetUsePDFCPU(false)
	unipdfResult, err := f.runSingleTest("unipdf", testFunc)
	if err != nil {
		return nil, fmt.Errorf("unipdf测试失败: %v", err)
	}
	comparison.UniPDFResult = unipdfResult

	// 分析结果
	f.analyzeComparison(comparison)

	// 保存结果
	f.mutex.Lock()
	f.results[testID] = comparison
	f.mutex.Unlock()

	// 保存到文件
	if f.outputPath != "" {
		f.saveResults()
	}

	return comparison, nil
}

// runSingleTest 运行单个测试
func (f *ABTestFramework) runSingleTest(engine string, testFunc func() error) (*ABTestResult, error) {
	result := &ABTestResult{
		TestID:    fmt.Sprintf("%s_%d", engine, time.Now().Unix()),
		TestName:  engine,
		StartTime: time.Now(),
	}

	// 记录开始时的内存
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialMemory := memStats.Alloc

	// 执行测试
	err := testFunc()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = err == nil
	if err != nil {
		result.Error = err.Error()
		result.ErrorCount = 1
	}

	// 记录结束时的内存
	runtime.ReadMemStats(&memStats)
	result.MemoryUsage = memStats.Alloc - initialMemory

	return result, nil
}

// analyzeComparison 分析对比结果
func (f *ABTestFramework) analyzeComparison(comparison *ABTestComparison) {
	pdfcpu := comparison.PDFCPUResult
	unipdf := comparison.UniPDFResult

	if pdfcpu == nil || unipdf == nil {
		comparison.Winner = "unknown"
		comparison.Recommendation = "测试数据不完整"
		return
	}

	// 计算性能提升（基于处理时间）
	if unipdf.Duration > 0 {
		comparison.PerformanceGain = float64(unipdf.Duration-pdfcpu.Duration) / float64(unipdf.Duration) * 100
	}

	// 计算内存减少
	if unipdf.MemoryUsage > 0 {
		comparison.MemoryReduction = float64(unipdf.MemoryUsage-pdfcpu.MemoryUsage) / float64(unipdf.MemoryUsage) * 100
	}

	// 确定获胜者
	pdfcpuScore := 0
	unipdfScore := 0

	// 成功率评分
	if pdfcpu.Success {
		pdfcpuScore += 10
	}
	if unipdf.Success {
		unipdfScore += 10
	}

	// 性能评分
	if comparison.PerformanceGain > 0 {
		pdfcpuScore += int(math.Min(comparison.PerformanceGain, 20))
	} else {
		unipdfScore += int(math.Min(-comparison.PerformanceGain, 20))
	}

	// 内存评分
	if comparison.MemoryReduction > 0 {
		pdfcpuScore += int(math.Min(comparison.MemoryReduction/10, 10))
	} else {
		unipdfScore += int(math.Min(-comparison.MemoryReduction/10, 10))
	}

	// 确定获胜者
	if pdfcpuScore > unipdfScore {
		comparison.Winner = "pdfcpu"
		comparison.Recommendation = "建议使用pdfcpu"
	} else if unipdfScore > pdfcpuScore {
		comparison.Winner = "unipdf"
		comparison.Recommendation = "建议使用UniPDF"
	} else {
		comparison.Winner = "tie"
		comparison.Recommendation = "两者性能相当"
	}
}

// saveResults 保存结果到文件
func (f *ABTestFramework) saveResults() error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	data, err := json.MarshalIndent(f.results, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f.outputPath, data, 0644)
}

// LoadResults 从文件加载结果
func (f *ABTestFramework) LoadResults() error {
	data, err := ioutil.ReadFile(f.outputPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &f.results)
}

// GetResults 获取所有测试结果
func (f *ABTestFramework) GetResults() map[string]*ABTestComparison {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	results := make(map[string]*ABTestComparison)
	for k, v := range f.results {
		results[k] = v
	}
	return results
}

// GenerateReport 生成测试报告
func (f *ABTestFramework) GenerateReport() string {
	results := f.GetResults()

	report := "# A/B测试报告\n\n"
	report += fmt.Sprintf("生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("测试总数: %d\n\n", len(results))

	totalTests := len(results)
	pdfcpuWins := 0
	unipdfWins := 0
	ties := 0

	for _, comparison := range results {
		report += fmt.Sprintf("## 测试: %s\n", comparison.TestName)
		report += fmt.Sprintf("- 测试ID: %s\n", comparison.TestID)
		report += fmt.Sprintf("- 获胜者: %s\n", comparison.Winner)
		report += fmt.Sprintf("- 建议: %s\n", comparison.Recommendation)

		if comparison.PDFCPUResult != nil {
			report += fmt.Sprintf("- pdfcpu耗时: %v\n", comparison.PDFCPUResult.Duration)
		}
		if comparison.UniPDFResult != nil {
			report += fmt.Sprintf("- UniPDF耗时: %v\n", comparison.UniPDFResult.Duration)
		}
		report += fmt.Sprintf("- 性能提升: %.2f%%\n", comparison.PerformanceGain)
		report += fmt.Sprintf("- 内存减少: %.2f%%\n\n", comparison.MemoryReduction)

		switch comparison.Winner {
		case "pdfcpu":
			pdfcpuWins++
		case "unipdf":
			unipdfWins++
		case "tie":
			ties++
		}
	}

	report += "## 统计摘要\n"
	report += fmt.Sprintf("- pdfcpu获胜: %d (%.1f%%)\n", pdfcpuWins, float64(pdfcpuWins)/float64(totalTests)*100)
	report += fmt.Sprintf("- UniPDF获胜: %d (%.1f%%)\n", unipdfWins, float64(unipdfWins)/float64(totalTests)*100)
	report += fmt.Sprintf("- 平局: %d (%.1f%%)\n", ties, float64(ties)/float64(totalTests)*100)

	return report
}
