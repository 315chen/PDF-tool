package pdf

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ABTestCase A/B测试用例
type ABTestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // "merge", "decrypt", "write", "performance"
	TestFunc    func() error           `json:"-"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    map[string]interface{} `json:"expected"`
}

// ABTestSuite A/B测试套件
type ABTestSuite struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Cases    []ABTestCase `json:"cases"`
	Created  time.Time    `json:"created"`
	Modified time.Time    `json:"modified"`
}

// ABTestStatistics A/B测试统计分析
type ABTestStatistics struct {
	TotalTests     int                    `json:"total_tests"`
	PDFCPUWins     int                    `json:"pdfcpu_wins"`
	UniPDFWins     int                    `json:"unipdf_wins"`
	Ties           int                    `json:"ties"`
	AvgPerformance float64                `json:"avg_performance_gain"`
	AvgMemory      float64                `json:"avg_memory_reduction"`
	CategoryStats  map[string]CategoryStat `json:"category_stats"`
	TimeRange      TimeRange              `json:"time_range"`
}

// CategoryStat 分类统计
type CategoryStat struct {
	TotalTests     int     `json:"total_tests"`
	PDFCPUWins     int     `json:"pdfcpu_wins"`
	UniPDFWins     int     `json:"unipdf_wins"`
	Ties           int     `json:"ties"`
	AvgPerformance float64 `json:"avg_performance_gain"`
	AvgMemory      float64 `json:"avg_memory_reduction"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ABTestManager A/B测试管理器
type ABTestManager struct {
	framework *ABTestFramework
	suites    map[string]*ABTestSuite
	config    *PDFServiceConfig
}

// NewABTestManager 创建A/B测试管理器
func NewABTestManager(config *PDFServiceConfig, framework *ABTestFramework) *ABTestManager {
	return &ABTestManager{
		framework: framework,
		suites:    make(map[string]*ABTestSuite),
		config:    config,
	}
}

// CreateTestSuite 创建测试套件
func (m *ABTestManager) CreateTestSuite(id, name string) *ABTestSuite {
	suite := &ABTestSuite{
		ID:       id,
		Name:     name,
		Cases:    make([]ABTestCase, 0),
		Created:  time.Now(),
		Modified: time.Now(),
	}
	m.suites[id] = suite
	return suite
}

// AddTestCase 添加测试用例
func (m *ABTestManager) AddTestCase(suiteID string, testCase ABTestCase) error {
	suite, exists := m.suites[suiteID]
	if !exists {
		return fmt.Errorf("测试套件不存在: %s", suiteID)
	}
	
	suite.Cases = append(suite.Cases, testCase)
	suite.Modified = time.Now()
	return nil
}

// RunTestSuite 运行测试套件
func (m *ABTestManager) RunTestSuite(suiteID string) ([]*ABTestComparison, error) {
	suite, exists := m.suites[suiteID]
	if !exists {
		return nil, fmt.Errorf("测试套件不存在: %s", suiteID)
	}
	
	var results []*ABTestComparison
	
	for _, testCase := range suite.Cases {
		comparison, err := m.framework.RunABTest(
			fmt.Sprintf("%s_%s", suiteID, testCase.ID),
			testCase.Name,
			testCase.TestFunc,
		)
		if err != nil {
			return results, fmt.Errorf("测试用例 %s 失败: %v", testCase.Name, err)
		}
		results = append(results, comparison)
	}
	
	return results, nil
}

// GenerateStatistics 生成统计分析
func (m *ABTestManager) GenerateStatistics() *ABTestStatistics {
	results := m.framework.GetResults()
	
	stats := &ABTestStatistics{
		TotalTests:    len(results),
		CategoryStats: make(map[string]CategoryStat),
		TimeRange: TimeRange{
			Start: time.Now(),
			End:   time.Now(),
		},
	}
	
	var totalPerformance float64
	var totalMemory float64
	var validComparisons int
	
	for _, comparison := range results {
		// 统计获胜者
		switch comparison.Winner {
		case "pdfcpu":
			stats.PDFCPUWins++
		case "unipdf":
			stats.UniPDFWins++
		case "tie":
			stats.Ties++
		}
		
		// 累计性能数据
		if comparison.PerformanceGain != 0 {
			totalPerformance += comparison.PerformanceGain
			validComparisons++
		}
		if comparison.MemoryReduction != 0 {
			totalMemory += comparison.MemoryReduction
		}
		
		// 更新时间范围
		if comparison.GeneratedAt.Before(stats.TimeRange.Start) {
			stats.TimeRange.Start = comparison.GeneratedAt
		}
		if comparison.GeneratedAt.After(stats.TimeRange.End) {
			stats.TimeRange.End = comparison.GeneratedAt
		}
		
		// 按分类统计
		category := m.getCategoryFromTestName(comparison.TestName)
		if categoryStat, exists := stats.CategoryStats[category]; exists {
			categoryStat.TotalTests++
			switch comparison.Winner {
			case "pdfcpu":
				categoryStat.PDFCPUWins++
			case "unipdf":
				categoryStat.UniPDFWins++
			case "tie":
				categoryStat.Ties++
			}
			categoryStat.AvgPerformance = (categoryStat.AvgPerformance*float64(categoryStat.TotalTests-1) + comparison.PerformanceGain) / float64(categoryStat.TotalTests)
			categoryStat.AvgMemory = (categoryStat.AvgMemory*float64(categoryStat.TotalTests-1) + comparison.MemoryReduction) / float64(categoryStat.TotalTests)
			stats.CategoryStats[category] = categoryStat
		} else {
			pdfcpuWins := 0
			unipdfWins := 0
			ties := 0
			if comparison.Winner == "pdfcpu" {
				pdfcpuWins = 1
			} else if comparison.Winner == "unipdf" {
				unipdfWins = 1
			} else if comparison.Winner == "tie" {
				ties = 1
			}
			stats.CategoryStats[category] = CategoryStat{
				TotalTests:     1,
				PDFCPUWins:     pdfcpuWins,
				UniPDFWins:     unipdfWins,
				Ties:           ties,
				AvgPerformance: comparison.PerformanceGain,
				AvgMemory:      comparison.MemoryReduction,
			}
		}
	}
	
	// 计算平均值
	if validComparisons > 0 {
		stats.AvgPerformance = totalPerformance / float64(validComparisons)
		stats.AvgMemory = totalMemory / float64(validComparisons)
	}
	
	return stats
}

// getCategoryFromTestName 从测试名称推断分类
func (m *ABTestManager) getCategoryFromTestName(testName string) string {
	testName = strings.ToLower(testName)
	
	switch {
	case strings.Contains(testName, "merge"):
		return "merge"
	case strings.Contains(testName, "decrypt"):
		return "decrypt"
	case strings.Contains(testName, "write"):
		return "write"
	case strings.Contains(testName, "performance"):
		return "performance"
	default:
		return "other"
	}
}

// GetTopPerformers 获取性能最佳的测试
func (m *ABTestManager) GetTopPerformers(limit int) []*ABTestComparison {
	results := m.framework.GetResults()
	
	var comparisons []*ABTestComparison
	for _, comparison := range results {
		comparisons = append(comparisons, comparison)
	}
	
	// 按性能提升排序
	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].PerformanceGain > comparisons[j].PerformanceGain
	})
	
	if limit > len(comparisons) {
		limit = len(comparisons)
	}
	
	return comparisons[:limit]
}

// GetWorstPerformers 获取性能最差的测试
func (m *ABTestManager) GetWorstPerformers(limit int) []*ABTestComparison {
	results := m.framework.GetResults()
	
	var comparisons []*ABTestComparison
	for _, comparison := range results {
		comparisons = append(comparisons, comparison)
	}
	
	// 按性能提升排序（升序）
	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].PerformanceGain < comparisons[j].PerformanceGain
	})
	
	if limit > len(comparisons) {
		limit = len(comparisons)
	}
	
	return comparisons[:limit]
}

// GenerateDetailedReport 生成详细报告
func (m *ABTestManager) GenerateDetailedReport() string {
	stats := m.GenerateStatistics()
	results := m.framework.GetResults()
	
	report := "# A/B测试详细报告\n\n"
	report += fmt.Sprintf("生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("测试时间范围: %s - %s\n", 
		stats.TimeRange.Start.Format("2006-01-02 15:04:05"),
		stats.TimeRange.End.Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("总测试数: %d\n\n", stats.TotalTests)
	
	// 总体统计
	report += "## 总体统计\n"
	report += fmt.Sprintf("- pdfcpu获胜: %d (%.1f%%)\n", 
		stats.PDFCPUWins, float64(stats.PDFCPUWins)/float64(stats.TotalTests)*100)
	report += fmt.Sprintf("- UniPDF获胜: %d (%.1f%%)\n", 
		stats.UniPDFWins, float64(stats.UniPDFWins)/float64(stats.TotalTests)*100)
	report += fmt.Sprintf("- 平局: %d (%.1f%%)\n", 
		stats.Ties, float64(stats.Ties)/float64(stats.TotalTests)*100)
	report += fmt.Sprintf("- 平均性能提升: %.2f%%\n", stats.AvgPerformance)
	report += fmt.Sprintf("- 平均内存减少: %.2f%%\n\n", stats.AvgMemory)
	
	// 分类统计
	report += "## 分类统计\n"
	for category, categoryStat := range stats.CategoryStats {
		report += fmt.Sprintf("### %s\n", category)
		report += fmt.Sprintf("- 测试数: %d\n", categoryStat.TotalTests)
		report += fmt.Sprintf("- pdfcpu获胜: %d (%.1f%%)\n", 
			categoryStat.PDFCPUWins, float64(categoryStat.PDFCPUWins)/float64(categoryStat.TotalTests)*100)
		report += fmt.Sprintf("- UniPDF获胜: %d (%.1f%%)\n", 
			categoryStat.UniPDFWins, float64(categoryStat.UniPDFWins)/float64(categoryStat.TotalTests)*100)
		report += fmt.Sprintf("- 平局: %d (%.1f%%)\n", 
			categoryStat.Ties, float64(categoryStat.Ties)/float64(categoryStat.TotalTests)*100)
		report += fmt.Sprintf("- 平均性能提升: %.2f%%\n", categoryStat.AvgPerformance)
		report += fmt.Sprintf("- 平均内存减少: %.2f%%\n\n", categoryStat.AvgMemory)
	}
	
	// 最佳性能测试
	topPerformers := m.GetTopPerformers(5)
	if len(topPerformers) > 0 {
		report += "## 最佳性能测试\n"
		for i, comparison := range topPerformers {
			report += fmt.Sprintf("%d. %s (性能提升: %.2f%%)\n", 
				i+1, comparison.TestName, comparison.PerformanceGain)
		}
		report += "\n"
	}
	
	// 最差性能测试
	worstPerformers := m.GetWorstPerformers(5)
	if len(worstPerformers) > 0 {
		report += "## 最差性能测试\n"
		for i, comparison := range worstPerformers {
			report += fmt.Sprintf("%d. %s (性能提升: %.2f%%)\n", 
				i+1, comparison.TestName, comparison.PerformanceGain)
		}
		report += "\n"
	}
	
	// 详细测试结果
	report += "## 详细测试结果\n"
	for _, comparison := range results {
		report += fmt.Sprintf("### %s\n", comparison.TestName)
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
	}
	
	return report
}

// CreatePredefinedTestCases 创建预定义测试用例
func (m *ABTestManager) CreatePredefinedTestCases() {
	// 创建基础测试套件
	basicSuite := m.CreateTestSuite("basic", "基础功能测试")
	
	// 添加PDF合并测试用例
	basicSuite.Cases = append(basicSuite.Cases, ABTestCase{
		ID:          "merge_small_files",
		Name:        "小文件合并测试",
		Description: "测试合并多个小PDF文件的性能",
		Category:    "merge",
		TestFunc:    func() error { return m.createSmallFilesMergeTest() },
		Parameters: map[string]interface{}{
			"file_count": 5,
			"file_size":  "1MB",
		},
		Expected: map[string]interface{}{
			"success": true,
		},
	})
	
	basicSuite.Cases = append(basicSuite.Cases, ABTestCase{
		ID:          "merge_large_files",
		Name:        "大文件合并测试",
		Description: "测试合并大PDF文件的性能",
		Category:    "merge",
		TestFunc:    func() error { return m.createLargeFilesMergeTest() },
		Parameters: map[string]interface{}{
			"file_count": 2,
			"file_size":  "50MB",
		},
		Expected: map[string]interface{}{
			"success": true,
		},
	})
	
	basicSuite.Cases = append(basicSuite.Cases, ABTestCase{
		ID:          "decrypt_encrypted_files",
		Name:        "加密文件解密测试",
		Description: "测试解密加密PDF文件的性能",
		Category:    "decrypt",
		TestFunc:    func() error { return m.createDecryptTest() },
		Parameters: map[string]interface{}{
			"password": "test123",
		},
		Expected: map[string]interface{}{
			"success": true,
		},
	})
	
	basicSuite.Cases = append(basicSuite.Cases, ABTestCase{
		ID:          "write_performance",
		Name:        "写入性能测试",
		Description: "测试PDF写入性能",
		Category:    "write",
		TestFunc:    func() error { return m.createWriteTest() },
		Parameters: map[string]interface{}{
			"page_count": 100,
		},
		Expected: map[string]interface{}{
			"success": true,
		},
	})
}

// 预定义测试用例的具体实现（占位符）
func (m *ABTestManager) createSmallFilesMergeTest() error {
	// 这里应该实现具体的小文件合并测试
	// 暂时返回nil作为占位符
	return nil
}

func (m *ABTestManager) createLargeFilesMergeTest() error {
	// 这里应该实现具体的大文件合并测试
	return nil
}

func (m *ABTestManager) createDecryptTest() error {
	// 这里应该实现具体的解密测试
	return nil
}

func (m *ABTestManager) createWriteTest() error {
	// 这里应该实现具体的写入测试
	return nil
} 