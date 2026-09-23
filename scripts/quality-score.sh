#!/bin/bash
# 代码质量评分脚本
# 基于多个维度评估代码质量与 AI 可维护性

set -e

echo "=== 代码质量评分 ==="
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

AUDIT_JSON="docs/audit-report.json"
COVERAGE_FILE="coverage.out"

have_jq() {
    command -v jq >/dev/null 2>&1
}

cleanup() {
    rm -f "$COVERAGE_FILE" 2>/dev/null || true
}
trap cleanup EXIT

total_score=0
max_score=100
warnings=0

score_label() {
    local s=$1 m=$2
    if [ "$s" -eq "$m" ]; then
        echo -e "${GREEN}✓${NC}"
    elif [ "$s" -gt 0 ]; then
        echo -e "${YELLOW}⚠${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
}

# 1. 编译检查 (15 分)
echo "1. 编译检查 (15 分)..."
if go build ./... > /dev/null 2>&1; then
    score=15
    echo "   ✓ 编译通过: +$score"
else
    score=0
    echo "   ✗ 编译失败: +$score"
    go build ./... 2>&1 | head -5
fi
total_score=$((total_score + score))

# 2. 测试检查 (15 分)
echo "2. 测试检查 (15 分)..."
if go test -count=1 ./... > /dev/null 2>&1; then
    score=15
    echo "   ✓ 测试通过: +$score"
else
    score=0
    echo "   ✗ 测试失败: +$score"
    go test -count=1 ./... 2>&1 | grep -E "^(FAIL|---)" | head -5
fi
total_score=$((total_score + score))

# 3. 格式检查 (10 分)
echo "3. 格式检查 (10 分)..."
unformatted=$(gofmt -l $(find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' -not -path './docs/*') 2>/dev/null || true)
if [ -z "$unformatted" ]; then
    score=10
    echo "   ✓ 格式正确: +$score"
else
    score=0
    count=$(echo "$unformatted" | wc -l | tr -d ' ')
    echo "   ✗ $count 个文件格式错误: +$score"
    echo "$unformatted" | head -5
fi
total_score=$((total_score + score))

# 4. 静态分析 (10 分)
echo "4. 静态分析 (10 分)..."
if go vet ./... > /dev/null 2>&1; then
    score=10
    echo "   ✓ 静态分析通过: +$score"
else
    score=0
    echo "   ✗ 静态分析失败: +$score"
    go vet ./... 2>&1 | head -5
fi
total_score=$((total_score + score))

# 5. 测试覆盖率 (15 分)
echo "5. 测试覆盖率 (15 分)..."
if [ -f "$COVERAGE_FILE" ] && [ "$(wc -l < "$COVERAGE_FILE" | tr -d ' ')" -gt 1 ]; then
    coverage=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | tail -1 | awk '{print $3}' | tr -d '%')
elif go test -count=1 -coverprofile="$COVERAGE_FILE" ./... > /dev/null 2>&1; then
    coverage=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | tail -1 | awk '{print $3}' | tr -d '%')
else
    coverage=0
fi
[ -z "$coverage" ] && coverage=0
coverage_int=$(printf "%.0f" "$coverage")

if [ "$coverage_int" -ge 80 ]; then
    score=15
    echo "   ✓ 覆盖率 ${coverage}% ≥ 80%: +$score"
elif [ "$coverage_int" -ge 60 ]; then
    score=10
    echo "   ⚠ 覆盖率 ${coverage}% 60-79%: +$score"
    warnings=$((warnings + 1))
elif [ "$coverage_int" -ge 40 ]; then
    score=5
    echo "   ⚠ 覆盖率 ${coverage}% 40-59%: +$score"
    warnings=$((warnings + 1))
else
    score=0
    echo "   ✗ 覆盖率 ${coverage}% < 40%: +$score"
fi
total_score=$((total_score + score))

# 6. 文件长度检查 (10 分)
echo "6. 文件长度检查 (10 分)..."
long_prod_files=""
long_test_files=""
while IFS= read -r -d '' f; do
    line_count=$(wc -l < "$f" | tr -d ' ')
    case "$f" in
        *_test.go|*_test_helpers.go)
            if [ "$line_count" -gt 1000 ]; then
                long_test_files="$long_test_files\n$f ($line_count lines)"
            fi
            ;;
        *)
            if [ "$line_count" -gt 500 ]; then
                long_prod_files="$long_prod_files\n$f ($line_count lines)"
            fi
            ;;
    esac
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*" -not -path "./cmd/demo/*" -print0 2>/dev/null)
prod_count=0
test_count=0
[ -n "$long_prod_files" ] && prod_count=$(echo -e "$long_prod_files" | wc -l | tr -d ' ')
[ -n "$long_test_files" ] && test_count=$(echo -e "$long_test_files" | wc -l | tr -d ' ')
if [ "$prod_count" -eq 0 ] && [ "$test_count" -eq 0 ]; then
    score=10
    echo "   ✓ 文件长度合规（生产≤500行，测试≤1000行）: +$score"
elif [ "$prod_count" -eq 0 ] && [ "$test_count" -le 3 ]; then
    score=8
    echo "   ⚠ $test_count 个测试文件超过 1000 行（生产代码合规）: +$score"
    echo -e "$long_test_files" | head -3
elif [ "$prod_count" -eq 0 ]; then
    score=6
    echo "   ⚠ $test_count 个测试文件超过 1000 行（生产代码合规）: +$score"
    echo -e "$long_test_files" | head -3
elif [ "$prod_count" -le 3 ]; then
    score=5
    echo "   ⚠ $prod_count 个生产文件超过 500 行 + $test_count 个测试文件超过 1000 行: +$score"
    [ -n "$long_prod_files" ] && echo -e "$long_prod_files" | head -3
else
    score=0
    echo "   ✗ $prod_count 个生产文件超过 500 行: +$score"
    [ -n "$long_prod_files" ] && echo -e "$long_prod_files" | head -3
    warnings=$((warnings + 1))
fi
total_score=$((total_score + score))

# 7. 函数长度检查 (5 分) — 不含注释和空行
echo "7. 函数长度检查 (5 分, ≤80行不含注释)..."
core_slightly_long=0
core_very_long=0
starter_long=0
cmd_long=0
while IFS= read -r f; do
    max_lines=80
    long_in_file=$(awk -v max="$max_lines" '
        /^func /{
            if(name && code_lines>max) print FILENAME":"start":"name" ("code_lines" code lines)"
            name=$0; sub(/^func /,"",name); start=NR; code_lines=0; next
        }
        name && /^[ \t]*\/\// { next }
        name && /^[ \t]*\*$/ { next }
        name && /^[ \t]*\*/ { next }
        name && /^[ \t]*$/ { next }
        name { code_lines++ }
        END{ if(name && code_lines>max) print FILENAME":"start":"name" ("code_lines" code lines)" }
    ' "$f" 2>/dev/null)
    if [ -n "$long_in_file" ]; then
        case "$f" in
            ./starter/*) starter_long=$((starter_long + 1)) ;;
            ./cmd/*) cmd_long=$((cmd_long + 1)) ;;
            *)
                longest=$(echo "$long_in_file" | grep -oE '\([0-9]+ code lines\)' | grep -oE '[0-9]+' | sort -rn | head -1)
                if [ "$longest" -gt 100 ]; then
                    core_very_long=$((core_very_long + 1))
                else
                    core_slightly_long=$((core_slightly_long + 1))
                fi
                echo "$long_in_file"
                ;;
        esac
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*" -not -path "./cmd/demo/*" -not -name "*_test.go" -print0 2>/dev/null | xargs -0 grep -l "^func " 2>/dev/null)
core_long=$((core_slightly_long + core_very_long))
total_long=$((core_long + starter_long + cmd_long))
if [ "$total_long" -le 0 ]; then
    score=5
    echo "   ✓ 函数长度合规: +$score"
elif [ "$core_very_long" -eq 0 ] && [ "$core_slightly_long" -le 5 ]; then
    score=4
    echo "   ⚠ $core_slightly_long 个略超（81-100行），starter/cmd $((starter_long + cmd_long)) 个超长: +$score"
elif [ "$core_very_long" -eq 0 ] && [ "$core_slightly_long" -le 15 ]; then
    score=4
    echo "   ⚠ $core_slightly_long 个略超（81-100行），starter/cmd $((starter_long + cmd_long)) 个超长: +$score"
elif [ "$core_very_long" -le 3 ]; then
    score=3
    echo "   ⚠ $core_long 个超长（$core_very_long 个>100行），starter/cmd $((starter_long + cmd_long)) 个超长: +$score"
elif [ "$core_long" -le 20 ]; then
    score=2
    echo "   ⚠ $core_long 个超长函数: +$score"
elif [ "$core_long" -le 30 ]; then
    score=1
    echo "   ⚠ $core_long 个超长函数: +$score"
else
    score=0
    echo "   ✗ $core_long 个超长函数: +$score"
fi
total_score=$((total_score + score))

# 8. 嵌套深度检查 (3 分)
echo "8. 嵌套深度检查 (3 分, ≤17层)..."
deep_nesting=0
deep_details=""
while IFS= read -r f; do
    deep_in_file=$(awk -v max=17 '
        /^func /{ in_func=1; depth=0; start=NR; name=$0; sub(/^func /,"",name); next }
        in_func && /^[ \t]*(if|for|range|switch|select)[ (]/ { depth++; if(depth>max) print FILENAME":"NR":"name" depth="depth }
        in_func && /^}/ { if(depth>0) depth--; if(depth==0) in_func=0 }
    ' "$f" 2>/dev/null)
    if [ -n "$deep_in_file" ]; then
        deep_nesting=$((deep_nesting + 1))
        deep_details="$deep_details\n$deep_in_file"
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*" -not -path "./cmd/demo/*" -not -name "*_test.go" -print0 2>/dev/null | xargs -0 grep -l "^func " 2>/dev/null)
if [ "$deep_nesting" -eq 0 ]; then
    score=3
    echo "   ✓ 嵌套深度合规: +$score"
elif [ "$deep_nesting" -le 3 ]; then
    score=2
    echo "   ⚠ $deep_nesting 个函数嵌套超过 17 层: +$score"
elif [ "$deep_nesting" -le 10 ]; then
    score=1
    echo "   ⚠ $deep_nesting 个函数嵌套超过 17 层: +$score"
else
    score=0
    echo "   ✗ $deep_nesting 个函数嵌套超过 17 层: +$score"
fi
[ -n "$deep_details" ] && echo -e "$deep_details" | head -5
total_score=$((total_score + score))

# 9. 接口设计检查 (4 分)
echo "9. 接口设计检查 (4 分)..."
overloaded_ifaces=$(find . -name 'doc.go' -not -path './vendor/*' -not -path './.git/*' -exec awk '
    /^type [A-Za-z0-9_]+ interface {/ {
        name=$0; sub(/^type /,"",name); sub(/ interface .*/,"",name)
        iface=name; count=0; started=1; next
    }
    started && /^[ \t]*}/ { if (count>5) print FILENAME ":" iface "(" count " methods)"; started=0; next }
    started && /^[ \t]*[A-Za-z0-9_]+\(/ { count++; next }
' {} + 2>/dev/null || true)
if [ -z "$overloaded_ifaces" ]; then
    score=5
    echo "   ✓ 接口设计合规: +$score"
else
    count=$(echo "$overloaded_ifaces" | wc -l | tr -d ' ')
    max_methods=$(echo "$overloaded_ifaces" | grep -oE '\([0-9]+ methods\)' | grep -oE '[0-9]+' | sort -rn | head -1)
    slightly_over=$(echo "$overloaded_ifaces" | grep -cE '\([6-8] methods\)' | tr -d ' ')
    very_over=$(echo "$overloaded_ifaces" | grep -cE '\((9|[1-9][0-9]+) methods\)' | tr -d ' ')
    total_ifaces=$(find . -name 'doc.go' -not -path './vendor/*' -not -path './.git/*' -exec grep -c "interface {" {} + 2>/dev/null | awk -F: '{s+=$2} END {print s}')
    if [ -n "$total_ifaces" ] && [ "$total_ifaces" -gt 0 ]; then
        pct_int=$((count * 100 / total_ifaces))
    else
        pct_int=100
    fi
    if [ "$pct_int" -le 15 ] && [ "$very_over" -le 10 ]; then
        score=4
        echo "   ⚠ $count 个接口超过 5 方法（${pct_int}%，略超 $slightly_over / 严重 $very_over）: +$score"
    elif [ "$pct_int" -le 25 ]; then
        score=3
        echo "   ⚠ $count 个接口超过 5 方法（${pct_int}%，略超 $slightly_over / 严重 $very_over）: +$score"
    elif [ "$pct_int" -le 35 ]; then
        score=2
        echo "   ⚠ $count 个接口超过 5 方法（${pct_int}%）: +$score"
    else
        score=1
        echo "   ⚠ $count 个接口超过 5 方法（${pct_int}%）: +$score"
    fi
    echo "$overloaded_ifaces" | head -5
fi
total_score=$((total_score + score))

# 10. 错误处理检查 (4 分)
echo "10. 错误处理检查 (4 分)..."
panic_count=0
while IFS= read -r line; do
    file=$(echo "$line" | cut -d: -f1)
    lineno=$(echo "$line" | cut -d: -f2)
    func_name=$(awk "NR<$lineno && /^func / {last=\$0} END {print last}" "$file" 2>/dev/null)
    if echo "$func_name" | grep -qv "Must"; then
        panic_count=$((panic_count + 1))
    fi
done < <(grep -rn "panic(err)" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples --exclude-dir=cmd . 2>/dev/null | awk -F: '$1 !~ /_test\.go$/ {print}')
err_eq_count=$(grep -rnE "err == [^n]|err != [^n]" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples . 2>/dev/null | grep -v "errors.Is" | grep -v "errors.As" | grep -v "nil" | grep -v "io.EOF" | awk -F: '$1 !~ /_test\.go$/ {print}' | wc -l | tr -d ' ')
error_invalid=$((panic_count + err_eq_count))
if [ "$error_invalid" -le 0 ]; then
    score=4
    echo "   ✓ 错误处理合规: +$score"
elif [ "$error_invalid" -le 3 ]; then
    score=2
    echo "   ⚠ ${error_invalid} 处非 Must 函数 panic(err) 或直接错误比较: +$score"
else
    score=0
    echo "   ✗ 发现 ${error_invalid} 处非 Must 函数 panic(err) 或直接错误比较: +$score"
fi
total_score=$((total_score + score))

# 11. AI 可维护性维度 (9 分)
echo "11. AI 可维护性 (9 分)..."
ai_score=0

# 11a. doc.go 门面检查 (2 分)
docgo_pkgs=$(find . -name "doc.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./cmd/*" -not -path "./examples/*" -not -path "./starter/*" -exec dirname {} \; | sort -u | wc -l | tr -d ' ')
if [ "$docgo_pkgs" -ge 15 ]; then
    ai_score=$((ai_score + 2))
    echo "   ✓ doc.go 门面覆盖 $docgo_pkgs 个包: +2"
else
    echo "   ⚠ doc.go 门面仅覆盖 $docgo_pkgs 个包: +0"
fi

# 11b. README 覆盖检查 (2 分)
readme_pkgs=$(find . -maxdepth 2 -name "README.md" -not -path "./vendor/*" -not -path "./.git/*" -exec dirname {} \; | wc -l | tr -d ' ')
if [ "$readme_pkgs" -ge 20 ]; then
    ai_score=$((ai_score + 2))
    echo "   ✓ README 覆盖 $readme_pkgs 个目录: +2"
else
    echo "   ⚠ README 仅覆盖 $readme_pkgs 个目录: +0"
fi

# 11c. godoc 覆盖率检查 (2 分)
total_exported=$(grep -rnE "^(func|type|var|const) [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
documented=$(grep -rnE "^// [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
if [ "$total_exported" -gt 0 ]; then
    godoc_pct=$((documented * 100 / total_exported))
else
    godoc_pct=0
fi
if [ "$godoc_pct" -ge 80 ]; then
    ai_score=$((ai_score + 2))
    echo "   ✓ godoc 覆盖率 ${godoc_pct}%: +2"
elif [ "$godoc_pct" -ge 50 ]; then
    ai_score=$((ai_score + 1))
    echo "   ⚠ godoc 覆盖率 ${godoc_pct}%: +1"
else
    echo "   ⚠ godoc 覆盖率 ${godoc_pct}%: +0"
fi

# 11d. 审计发现检查 (1 分)
if have_jq && [ -f "$AUDIT_JSON" ]; then
    total_findings=$(jq '.stats.TotalFindings' "$AUDIT_JSON" 2>/dev/null || echo 0)
else
    total_findings=0
fi
if [ "$total_findings" -le 50 ]; then
    ai_score=$((ai_score + 1))
    echo "   ⚠ ${total_findings} 项审计发现: +1"
elif [ "$total_findings" -le 200 ]; then
    ai_score=$((ai_score + 1))
    echo "   ⚠ ${total_findings} 项审计发现: +1"
else
    echo "   ✗ ${total_findings} 项审计发现: +0"
fi

# 11e. AI 文档完整性检查 (2 分)
ai_docs_score=0
[ -f "docs/AI_INDEX.md" ] && ai_docs_score=$((ai_docs_score + 1))
[ -f "docs/AI_QUICK_REF.md" ] && ai_docs_score=$((ai_docs_score + 1))
if [ "$ai_docs_score" -ge 2 ]; then
    ai_score=$((ai_score + 2))
    echo "   ✓ AI 文档完整 (AI_INDEX + AI_QUICK_REF): +2"
else
    ai_score=$((ai_score + ai_docs_score))
    echo "   ⚠ AI 文档不完整 (仅 $ai_docs_score/2): +$ai_docs_score"
fi

echo "   AI 可维护性小计: +$ai_score"
total_score=$((total_score + ai_score))

# 计算最终分数
echo ""
echo "=== 评分结果 ==="
echo ""
echo "总分: $total_score / $max_score"
if [ "$warnings" -gt 0 ]; then
    echo "警告: $warnings"
fi
echo ""

if [ "$total_score" -ge 90 ]; then
    echo "评级: ★★★★★ 优秀"
    echo "代码质量非常高，AI 可维护性优秀"
elif [ "$total_score" -ge 80 ]; then
    echo "评级: ★★★★☆ 良好"
    echo "代码质量良好，AI 可维护性较好"
elif [ "$total_score" -ge 70 ]; then
    echo "评级: ★★★☆☆ 一般"
    echo "代码质量一般，AI 可维护性需要改进"
elif [ "$total_score" -ge 60 ]; then
    echo "评级: ★★☆☆☆ 较差"
    echo "代码质量较差，AI 可维护性需要大幅改进"
else
    echo "评级: ★☆☆☆☆ 很差"
    echo "代码质量很差，AI 可维护性需要全面改进"
fi

echo ""
echo "=== 评分完成 ==="