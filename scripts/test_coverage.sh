#!/bin/bash

# PDF合并工具测试覆盖率脚本

set -e

echo "PDF合并工具 - 测试覆盖率报告"
echo "=============================="

# 创建覆盖率输出目录
mkdir -p coverage

# 运行测试并生成覆盖率报告
echo "正在运行测试..."
go test ./internal/... ./pkg/... -coverprofile=coverage/coverage.out -covermode=atomic

# 生成HTML报告
echo "生成HTML覆盖率报告..."
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# 显示覆盖率统计
echo ""
echo "覆盖率统计:"
echo "==========="
go tool cover -func=coverage/coverage.out | tail -1

echo ""
echo "各模块覆盖率:"
echo "============"

# 按模块显示覆盖率
echo "控制器模块:"
go tool cover -func=coverage/coverage.out | grep "internal/controller" | awk '{sum+=$3; count++} END {if(count>0) printf "平均覆盖率: %.1f%%\n", sum/count}'

echo "模型模块:"
go tool cover -func=coverage/coverage.out | grep "internal/model" | awk '{sum+=$3; count++} END {if(count>0) printf "平均覆盖率: %.1f%%\n", sum/count}'

echo "UI模块:"
go tool cover -func=coverage/coverage.out | grep "internal/ui" | awk '{sum+=$3; count++} END {if(count>0) printf "平均覆盖率: %.1f%%\n", sum/count}'

echo "文件管理模块:"
go tool cover -func=coverage/coverage.out | grep "pkg/file" | awk '{sum+=$3; count++} END {if(count>0) printf "平均覆盖率: %.1f%%\n", sum/count}'

echo "PDF处理模块:"
go tool cover -func=coverage/coverage.out | grep "pkg/pdf" | awk '{sum+=$3; count++} END {if(count>0) printf "平均覆盖率: %.1f%%\n", sum/count}'

echo ""
echo "低覆盖率函数 (<50%):"
echo "=================="
go tool cover -func=coverage/coverage.out | awk '$3 < 50.0 && $3 != "0.0%" {print $1 ": " $3}' | head -10

echo ""
echo "未覆盖函数 (0%):"
echo "=============="
go tool cover -func=coverage/coverage.out | awk '$3 == "0.0%" {print $1}' | head -10

echo ""
echo "覆盖率报告已生成: coverage/coverage.html"
echo "可以在浏览器中打开查看详细报告"

# 检查覆盖率目标
TOTAL_COVERAGE=$(go tool cover -func=coverage/coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
TARGET_COVERAGE=70

echo ""
if (( $(echo "$TOTAL_COVERAGE >= $TARGET_COVERAGE" | bc -l) )); then
    echo "✅ 覆盖率目标达成: $TOTAL_COVERAGE% >= $TARGET_COVERAGE%"
else
    echo "❌ 覆盖率未达目标: $TOTAL_COVERAGE% < $TARGET_COVERAGE%"
    echo "需要为以下模块添加更多测试:"
    
    # 显示需要改进的模块
    go tool cover -func=coverage/coverage.out | awk -v target=$TARGET_COVERAGE '$3 < target && $3 != "0.0%" {print "  - " $1 ": " $3}' | head -5
fi

echo ""
echo "测试覆盖率分析完成！"