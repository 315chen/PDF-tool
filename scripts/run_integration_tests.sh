#!/bin/bash

# PDF合并工具集成测试运行脚本

set -e

echo "PDF合并工具 - 集成测试套件"
echo "============================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试配置
TEST_TIMEOUT="30s"
TEST_VERBOSE="-v"
SHORT_MODE=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --short)
            SHORT_MODE="-short"
            echo "启用短测试模式"
            shift
            ;;
        --quiet)
            TEST_VERBOSE=""
            shift
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        *)
            echo "未知参数: $1"
            echo "用法: $0 [--short] [--quiet] [--timeout <duration>]"
            exit 1
            ;;
    esac
done

echo "测试配置:"
echo "  超时时间: $TEST_TIMEOUT"
echo "  详细模式: $([ -n "$TEST_VERBOSE" ] && echo "启用" || echo "禁用")"
echo "  短测试模式: $([ -n "$SHORT_MODE" ] && echo "启用" || echo "禁用")"
echo ""

# 测试函数
run_test_suite() {
    local suite_name="$1"
    local test_pattern="$2"
    local description="$3"
    
    echo -e "${BLUE}运行测试套件: $suite_name${NC}"
    echo "描述: $description"
    echo "模式: $test_pattern"
    echo ""
    
    if go test ./test $TEST_VERBOSE -run "$test_pattern" -timeout "$TEST_TIMEOUT" $SHORT_MODE; then
        echo -e "${GREEN}✅ $suite_name 测试通过${NC}"
        return 0
    else
        echo -e "${RED}❌ $suite_name 测试失败${NC}"
        return 1
    fi
}

# 运行基准测试
run_benchmark() {
    local bench_name="$1"
    local bench_pattern="$2"
    local description="$3"
    
    echo -e "${BLUE}运行基准测试: $bench_name${NC}"
    echo "描述: $description"
    echo ""
    
    if go test ./test -bench "$bench_pattern" -benchmem -timeout "$TEST_TIMEOUT" $SHORT_MODE; then
        echo -e "${GREEN}✅ $bench_name 基准测试完成${NC}"
        return 0
    else
        echo -e "${RED}❌ $bench_name 基准测试失败${NC}"
        return 1
    fi
}

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 运行测试套件
echo "开始运行集成测试..."
echo "==================="

# 1. 数据模型集成测试
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "数据模型集成" "TestIntegration_DataModelOperations" "测试数据模型的各种操作"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 2. 错误处理集成测试
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "错误处理集成" "TestIntegration_ErrorHandling" "测试各种错误情况的处理"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 3. 配置管理集成测试
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "配置管理集成" "TestIntegration_ConfigurationManagement" "测试配置管理功能"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 4. 文件不存在错误场景
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "文件不存在场景" "TestErrorScenarios_FileNotFound" "测试文件不存在的错误处理"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 5. 无效文件错误场景
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "无效文件场景" "TestErrorScenarios_InvalidFiles" "测试无效文件的错误处理"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 6. 无效参数错误场景
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "无效参数场景" "TestErrorScenarios_InvalidParameters" "测试无效参数的错误处理"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 7. 错误恢复测试
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_suite "错误恢复" "TestErrorScenarios_Recovery" "测试系统错误恢复能力"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# 如果不是短测试模式，运行更多测试
if [ -z "$SHORT_MODE" ]; then
    echo "运行完整测试套件..."
    echo "=================="
    
    # 8. 完整工作流程测试
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "完整工作流程" "TestIntegration_FullWorkflow" "测试完整的PDF合并工作流程"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    # 9. 取消工作流程测试
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "取消工作流程" "TestIntegration_CancellationWorkflow" "测试任务取消功能"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    # 10. 内存监控测试
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "内存监控" "TestIntegration_MemoryMonitoring" "测试内存监控功能"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    # 11. 并发操作测试
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "并发操作" "TestErrorScenarios_ConcurrentAccess" "测试并发访问处理"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    # 12. 性能测试
    echo "运行性能测试..."
    echo "=============="
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "文件验证性能" "TestPerformance_FileValidation" "测试文件验证性能"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "并发操作性能" "TestPerformance_ConcurrentOperations" "测试并发操作性能"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test_suite "内存泄漏检测" "TestPerformance_MemoryLeaks" "检测内存泄漏"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
    
    # 运行基准测试
    echo "运行基准测试..."
    echo "=============="
    
    run_benchmark "文件验证基准" "BenchmarkPerformance_FileValidation" "文件验证操作的性能基准"
    echo ""
    
    run_benchmark "事件处理基准" "BenchmarkPerformance_EventHandling" "事件处理的性能基准"
    echo ""
    
    run_benchmark "错误处理基准" "BenchmarkErrorScenarios_FileNotFound" "错误处理的性能基准"
    echo ""
fi

# 生成测试报告
echo "测试报告"
echo "========"
echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有集成测试通过！${NC}"
    SUCCESS_RATE="100%"
else
    SUCCESS_RATE=$(echo "scale=1; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)
    echo -e "${YELLOW}⚠️  成功率: $SUCCESS_RATE%${NC}"
fi

echo ""
echo "测试环境信息:"
echo "============"
echo "Go版本: $(go version)"
echo "操作系统: $(uname -s)"
echo "架构: $(uname -m)"
echo "测试时间: $(date)"

# 生成覆盖率报告（如果需要）
if [ -z "$SHORT_MODE" ]; then
    echo ""
    echo "生成覆盖率报告..."
    echo "================"
    
    if go test ./test -coverprofile=integration_coverage.out -timeout "$TEST_TIMEOUT"; then
        echo "覆盖率报告已生成: integration_coverage.out"
        
        # 显示覆盖率统计
        COVERAGE=$(go tool cover -func=integration_coverage.out | tail -1 | awk '{print $3}')
        echo -e "集成测试覆盖率: ${BLUE}$COVERAGE${NC}"
        
        # 生成HTML报告
        go tool cover -html=integration_coverage.out -o integration_coverage.html
        echo "HTML覆盖率报告: integration_coverage.html"
    else
        echo -e "${YELLOW}覆盖率报告生成失败${NC}"
    fi
fi

echo ""
echo "集成测试完成！"

# 退出码
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi