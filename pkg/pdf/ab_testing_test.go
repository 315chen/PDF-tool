package pdf

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestABTestFramework_Basic(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	
	// 测试基本A/B测试
	comparison, err := framework.RunABTest("test_001", "基础性能测试", func() error {
		// 模拟一些工作
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	
	if err != nil {
		t.Fatalf("A/B测试失败: %v", err)
	}
	
	if comparison == nil {
		t.Fatal("比较结果为空")
	}
	
	if comparison.TestID != "test_001" {
		t.Errorf("期望测试ID为 'test_001', 实际为 '%s'", comparison.TestID)
	}
	
	if comparison.PDFCPUResult == nil || comparison.UniPDFResult == nil {
		t.Fatal("pdfcpu或UniPDF结果为空")
	}
	
	// 验证结果结构
	if comparison.PDFCPUResult.TestName != "pdfcpu" {
		t.Errorf("期望pdfcpu测试名称为 'pdfcpu', 实际为 '%s'", comparison.PDFCPUResult.TestName)
	}
	
	if comparison.UniPDFResult.TestName != "unipdf" {
		t.Errorf("期望UniPDF测试名称为 'unipdf', 实际为 '%s'", comparison.UniPDFResult.TestName)
	}
}

func TestABTestFramework_ErrorHandling(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	
	// 测试错误处理
	comparison, err := framework.RunABTest("test_002", "错误测试", func() error {
		return fmt.Errorf("模拟错误")
	})
	
	if err != nil {
		t.Fatalf("A/B测试框架应该处理错误: %v", err)
	}
	
	if comparison == nil {
		t.Fatal("比较结果为空")
	}
	
	// 验证错误被正确记录
	if comparison.PDFCPUResult.Success {
		t.Error("pdfcpu测试应该失败")
	}
	
	if comparison.UniPDFResult.Success {
		t.Error("UniPDF测试应该失败")
	}
	
	if comparison.PDFCPUResult.Error == "" {
		t.Error("pdfcpu测试应该有错误信息")
	}
	
	if comparison.UniPDFResult.Error == "" {
		t.Error("UniPDF测试应该有错误信息")
	}
}

func TestABTestFramework_ResultsManagement(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	
	// 运行多个测试
	testCases := []struct {
		id   string
		name string
	}{
		{"test_001", "测试1"},
		{"test_002", "测试2"},
		{"test_003", "测试3"},
	}
	
	for _, tc := range testCases {
		_, err := framework.RunABTest(tc.id, tc.name, func() error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("测试 %s 失败: %v", tc.id, err)
		}
	}
	
	// 验证结果管理
	results := framework.GetResults()
	if len(results) != 3 {
		t.Errorf("期望3个测试结果, 实际为 %d", len(results))
	}
	
	for _, tc := range testCases {
		if _, exists := results[tc.id]; !exists {
			t.Errorf("测试结果 %s 不存在", tc.id)
		}
	}
}

func TestABTestManager_Basic(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 创建测试套件
	suite := manager.CreateTestSuite("test_suite", "测试套件")
	if suite == nil {
		t.Fatal("无法创建测试套件")
	}
	
	if suite.ID != "test_suite" {
		t.Errorf("期望套件ID为 'test_suite', 实际为 '%s'", suite.ID)
	}
	
	if suite.Name != "测试套件" {
		t.Errorf("期望套件名称为 '测试套件', 实际为 '%s'", suite.Name)
	}
}

func TestABTestManager_AddTestCase(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 创建测试套件
	suite := manager.CreateTestSuite("test_suite", "测试套件")
	
	// 添加测试用例
	testCase := ABTestCase{
		ID:          "test_case_001",
		Name:        "测试用例1",
		Description: "这是一个测试用例",
		Category:    "merge",
		TestFunc:    func() error { return nil },
		Parameters:  map[string]interface{}{"param1": "value1"},
		Expected:    map[string]interface{}{"success": true},
	}
	
	err := manager.AddTestCase("test_suite", testCase)
	if err != nil {
		t.Fatalf("添加测试用例失败: %v", err)
	}
	
	// 验证测试用例被添加
	if len(suite.Cases) != 1 {
		t.Errorf("期望1个测试用例, 实际为 %d", len(suite.Cases))
	}
	
	if suite.Cases[0].ID != "test_case_001" {
		t.Errorf("期望测试用例ID为 'test_case_001', 实际为 '%s'", suite.Cases[0].ID)
	}
}

func TestABTestManager_RunTestSuite(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 创建测试套件并添加测试用例
	_ = manager.CreateTestSuite("test_suite", "测试套件")
	
	testCase := ABTestCase{
		ID:          "test_case_001",
		Name:        "测试用例1",
		Description: "这是一个测试用例",
		Category:    "merge",
		TestFunc:    func() error { 
			time.Sleep(5 * time.Millisecond)
			return nil 
		},
		Parameters: map[string]interface{}{"param1": "value1"},
		Expected:   map[string]interface{}{"success": true},
	}
	
	manager.AddTestCase("test_suite", testCase)
	
	// 运行测试套件
	results, err := manager.RunTestSuite("test_suite")
	if err != nil {
		t.Fatalf("运行测试套件失败: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("期望1个测试结果, 实际为 %d", len(results))
	}
	
	if results[0].TestName != "测试用例1" {
		t.Errorf("期望测试名称为 '测试用例1', 实际为 '%s'", results[0].TestName)
	}
}

func TestABTestManager_Statistics(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 运行一些测试来生成数据
	_ = manager.CreateTestSuite("stats_suite", "统计测试套件")
	testCases := []ABTestCase{
		{
			ID:       "merge_test",
			Name:     "PDF合并测试",
			Category: "merge",
			TestFunc: func() error { 
				time.Sleep(10 * time.Millisecond)
				return nil 
			},
		},
		{
			ID:       "decrypt_test",
			Name:     "PDF解密测试",
			Category: "decrypt",
			TestFunc: func() error { 
				time.Sleep(15 * time.Millisecond)
				return nil 
			},
		},
		{
			ID:       "write_test",
			Name:     "PDF写入测试",
			Category: "write",
			TestFunc: func() error { 
				time.Sleep(20 * time.Millisecond)
				return nil 
			},
		},
	}
	
	_ = manager.CreateTestSuite("stats_suite", "统计测试套件")
	for _, tc := range testCases {
		manager.AddTestCase("stats_suite", tc)
	}
	
	// 运行测试套件
	_, err := manager.RunTestSuite("stats_suite")
	if err != nil {
		t.Fatalf("运行测试套件失败: %v", err)
	}
	
	// 生成统计
	stats := manager.GenerateStatistics()
	if stats == nil {
		t.Fatal("统计结果为空")
	}
	
	if stats.TotalTests != 3 {
		t.Errorf("期望3个测试, 实际为 %d", stats.TotalTests)
	}
	
	// 验证分类统计
	if len(stats.CategoryStats) == 0 {
		t.Error("分类统计为空")
	}
	
	// 验证时间范围
	if stats.TimeRange.Start.IsZero() || stats.TimeRange.End.IsZero() {
		t.Error("时间范围未正确设置")
	}
}

func TestABTestManager_TopPerformers(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 运行一些测试
	_ = manager.CreateTestSuite("perf_suite", "性能测试套件")
	
	for i := 1; i <= 5; i++ {
		testCase := ABTestCase{
			ID:       fmt.Sprintf("test_%d", i),
			Name:     fmt.Sprintf("性能测试%d", i),
			Category: "performance",
			TestFunc: func() error { 
				time.Sleep(time.Duration(i*10) * time.Millisecond)
				return nil 
			},
		}
		manager.AddTestCase("perf_suite", testCase)
	}
	
	// 运行测试套件
	_, err := manager.RunTestSuite("perf_suite")
	if err != nil {
		t.Fatalf("运行测试套件失败: %v", err)
	}
	
	// 获取最佳性能测试
	topPerformers := manager.GetTopPerformers(3)
	if len(topPerformers) != 3 {
		t.Errorf("期望3个最佳性能测试, 实际为 %d", len(topPerformers))
	}
	
	// 验证排序（性能提升应该递减）
	for i := 1; i < len(topPerformers); i++ {
		if topPerformers[i-1].PerformanceGain < topPerformers[i].PerformanceGain {
			t.Error("最佳性能测试未正确排序")
		}
	}
	
	// 获取最差性能测试
	worstPerformers := manager.GetWorstPerformers(3)
	if len(worstPerformers) != 3 {
		t.Errorf("期望3个最差性能测试, 实际为 %d", len(worstPerformers))
	}
	
	// 验证排序（性能提升应该递增）
	for i := 1; i < len(worstPerformers); i++ {
		if worstPerformers[i-1].PerformanceGain > worstPerformers[i].PerformanceGain {
			t.Error("最差性能测试未正确排序")
		}
	}
}

func TestABTestManager_DetailedReport(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 运行一些测试
	_ = manager.CreateTestSuite("report_suite", "报告测试套件")
	
	testCases := []ABTestCase{
		{
			ID:       "merge_test",
			Name:     "PDF合并测试",
			Category: "merge",
			TestFunc: func() error { 
				time.Sleep(10 * time.Millisecond)
				return nil 
			},
		},
		{
			ID:       "decrypt_test",
			Name:     "PDF解密测试",
			Category: "decrypt",
			TestFunc: func() error { 
				time.Sleep(15 * time.Millisecond)
				return nil 
			},
		},
	}
	
	for _, tc := range testCases {
		manager.AddTestCase("report_suite", tc)
	}
	
	// 运行测试套件
	_, err := manager.RunTestSuite("report_suite")
	if err != nil {
		t.Fatalf("运行测试套件失败: %v", err)
	}
	
	// 生成详细报告
	report := manager.GenerateDetailedReport()
	if report == "" {
		t.Fatal("生成的报告为空")
	}
	
	// 验证报告包含必要内容
	expectedSections := []string{
		"# A/B测试详细报告",
		"## 总体统计",
		"## 分类统计",
		"## 详细测试结果",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(report, section) {
			t.Errorf("报告缺少必要章节: %s", section)
		}
	}
}

func TestABTestManager_CreatePredefinedTestCases(t *testing.T) {
	config := DefaultPDFServiceConfig()
	framework := NewABTestFramework(config, "test_results.json")
	manager := NewABTestManager(config, framework)
	
	// 创建预定义测试用例
	manager.CreatePredefinedTestCases()
	
	// 验证基础测试套件被创建
	basicSuite, exists := manager.suites["basic"]
	if !exists {
		t.Fatal("基础测试套件未创建")
	}
	
	if basicSuite.Name != "基础功能测试" {
		t.Errorf("期望套件名称为 '基础功能测试', 实际为 '%s'", basicSuite.Name)
	}
	
	// 验证预定义测试用例
	expectedCases := []string{
		"merge_small_files",
		"merge_large_files", 
		"decrypt_encrypted_files",
		"write_performance",
	}
	
	if len(basicSuite.Cases) != len(expectedCases) {
		t.Errorf("期望 %d 个预定义测试用例, 实际为 %d", len(expectedCases), len(basicSuite.Cases))
	}
	
	for _, expectedCase := range expectedCases {
		found := false
		for _, testCase := range basicSuite.Cases {
			if testCase.ID == expectedCase {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("缺少预定义测试用例: %s", expectedCase)
		}
	}
} 