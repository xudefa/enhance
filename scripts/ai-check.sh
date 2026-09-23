#!/bin/bash
# AI 可维护性检查脚本
# 用于检查代码是否符合 enhance 项目的 AI 可维护性规范

set -e

echo "=== AI 可维护性检查 ==="
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

AUDIT_JSON="docs/audit-report.json"
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

pass() {
    echo -e "${GREEN}✓ 通过${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo -e "${RED}✗ 失败${NC}"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
    WARN_COUNT=$((WARN_COUNT + 1))
}

have_jq() {
    command -v jq >/dev/null 2>&1
}

# 1. 编译检查
echo -n "1. 编译... "
if go build ./... > /dev/null 2>&1; then
    pass
else
    fail
    go build ./... 2>&1 | head -3
fi

# 2. 测试检查
echo -n "2. 测试... "
if go test -count=1 ./... > /dev/null 2>&1; then
    pass
else
    fail
    go test -count=1 ./... 2>&1 | grep -E "^(FAIL|---)" | head -3
fi

# 3. 格式检查
echo -n "3. 格式化... "
UNFORMATTED=$(gofmt -l $(find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' -not -path './docs/*') 2>/dev/null || true)
if [ -z "$UNFORMATTED" ]; then
    pass
else
    fail
    echo "$UNFORMATTED" | head -5
fi

# 4. 静态分析
echo -n "4. 静态分析... "
if go vet ./... > /dev/null 2>&1; then
    pass
else
    fail
    go vet ./... 2>&1 | head -3
fi

# 5. 依赖整理检查（只读）
echo -n "5. 依赖整理... "
DIRTY_MODULES=$(git status --porcelain -- 'go.mod' 'go.sum' 'go.work' ':(glob)**/go.mod' ':(glob)**/go.sum' ':(glob)**/go.work' 2>/dev/null || true)
if [ -z "$DIRTY_MODULES" ]; then
    pass
else
    warn "go.mod/go.sum 存在未提交改动，请先 go mod tidy 并提交"
    echo "$DIRTY_MODULES"
fi

# 6. 文件长度检查（生产≤500行，测试≤1000行）
echo -n "6. 文件长度 (生产≤500行/测试≤1000行)... "
LONG_FILES=""
while IFS= read -r -d '' f; do
    line_count=$(wc -l < "$f" | tr -d ' ')
    case "$f" in
        *_test.go) max_lines=1000 ;;
        *) max_lines=500 ;;
    esac
    if [ "$line_count" -gt "$max_lines" ]; then
        LONG_FILES="$LONG_FILES\n$f ($line_count, 限$max_lines)"
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -print0 2>/dev/null)
if [ -z "$LONG_FILES" ]; then
    pass
else
    warn "以下文件超过限制"
    echo -e "$LONG_FILES"
fi

# 7. 函数长度检查（不含注释和空行）
echo -n "7. 函数长度 (≤80行，不含注释)... "
core_long=0
starter_long=0
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
            ./starter/*|./cmd/*) starter_long=$((starter_long + 1)) ;;
            *) core_long=$((core_long + 1)) ;;
        esac
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*" -not -path "./cmd/demo/*" -not -name "*_test.go" -print0 2>/dev/null | xargs -0 grep -l "^func " 2>/dev/null)
if [ "$core_long" -eq 0 ] && [ "$starter_long" -eq 0 ]; then
    pass
else
    warn "核心 $core_long 个 + starter/cmd $starter_long 个超长函数"
fi

# 7b. 嵌套深度检查
echo -n "7b. 嵌套深度 (≤17层)... "
deep_nesting=0
while IFS= read -r f; do
    deep_in_file=$(awk -v max=17 '
        /^func /{ in_func=1; depth=0; start=NR; name=$0; sub(/^func /,"",name); next }
        in_func && /^[ \t]*(if|for|range|switch|select)[ (]/ { depth++; if(depth>max) print FILENAME":"NR":"name" depth="depth }
        in_func && /^}/ { if(depth>0) depth--; if(depth==0) in_func=0 }
    ' "$f" 2>/dev/null)
    if [ -n "$deep_in_file" ]; then
        deep_nesting=$((deep_nesting + 1))
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*" -not -path "./cmd/demo/*" -not -name "*_test.go" -print0 2>/dev/null | xargs -0 grep -l "^func " 2>/dev/null)
if [ "$deep_nesting" -eq 0 ]; then
    pass
else
    warn "$deep_nesting 个函数嵌套深度超过 17 层"
fi

# 8. 接口方法数量检查
echo -n "8. 接口方法数 (≤5)... "
IFACE_EXEMPT="ApplicationContext|ContainerExt|BeanRegistry|LifecycleManager|Config|Configurer|ExpressionInterceptUrlRegistry|SecurityRequest|UserDetails|UrlAuthorizationRuleBuilder|MeterRegistry|Scheduler|Starter|TenantManager|Queue|TestContext|Router|Context|HTTPClient|TestServer|ResponseVerifier"
OVERLOADED_IFACES=$(find . -name 'doc.go' -not -path './vendor/*' -not -path './.git/*' -exec awk -v exempt="$IFACE_EXEMPT" '
    BEGIN { split(exempt, e, "|"); for (i in e) exempt_map[e[i]]=1 }
    /^type [A-Za-z0-9_]+ interface {/ {
        name=$0; sub(/^type /,"",name); sub(/ interface .*/,"",name)
        iface=name; count=0; started=1; next
    }
    started && /^[ \t]*}/ { if (count>5 && !(iface in exempt_map)) print FILENAME ":" iface "(" count ")"; started=0; next }
    started && /^[ \t]*[A-Za-z0-9_]+\(/ { count++; next }
' {} + 2>/dev/null || true)
if [ -z "$OVERLOADED_IFACES" ]; then
    pass
else
    IFACE_TOTAL=$(echo "$OVERLOADED_IFACES" | wc -l | tr -d ' ')
    warn "发现 ${IFACE_TOTAL} 个接口方法数超过 5"
    echo "$OVERLOADED_IFACES" | head -5
fi

# 9. 错误处理检查
echo -n "9. 错误处理... "
PANIC_ERR=""
while IFS= read -r line; do
    file=$(echo "$line" | cut -d: -f1)
    lineno=$(echo "$line" | cut -d: -f2)
    func_name=$(awk "NR<$lineno && /^func / {last=\$0} END {print last}" "$file" 2>/dev/null)
    if echo "$func_name" | grep -qv "Must"; then
        PANIC_ERR="$PANIC_ERR\n$line"
    fi
done < <(grep -rn "panic(err)" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples --exclude-dir=cmd . 2>/dev/null | awk -F: '$1 !~ /_test\.go$/ {print}')
ERR_EQ=$(grep -rnE "err == [^n]|err != [^n]" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples . 2>/dev/null | grep -v "errors.Is" | grep -v "errors.As" | grep -v "nil" | grep -v "io.EOF" | awk -F: '$1 !~ /_test\.go$/ {print}' | head -5 || true)
if [ -z "$PANIC_ERR" ] && [ -z "$ERR_EQ" ]; then
    pass
else
    warn "发现非 Must 函数中的 panic(err) 或直接错误比较"
    [ -n "$PANIC_ERR" ] && echo -e "$PANIC_ERR" | head -5
    [ -n "$ERR_EQ" ] && echo "$ERR_EQ"
fi

# 10. 裸 goroutine 检查
echo -n "10. 裸 goroutine... "
NAKED_GOROUTINES=$(grep -rn "go func()" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples --exclude-dir=cmd . 2>/dev/null | awk -F: '$1 !~ /_test\.go$/ {print}' | grep -v "errgroup" | grep -v "WaitGroup" | head -5 || true)
if [ -z "$NAKED_GOROUTINES" ]; then
    pass
else
    warn "发现裸 goroutine"
    echo "$NAKED_GOROUTINES"
fi

# 11. doc.go 门面检查（核心包必须有 doc.go）
echo -n "11. doc.go 门面... "
CORE_PKGS="core boot context condition config event cache web security log metrics actuator schedule resilience validation exception lifecycle async audit retry i18n spel testing webtest openapi metadata devtools email mq tenant"
MISSING_DOCGO=""
for pkg in $CORE_PKGS; do
    if [ -d "$pkg" ] && [ ! -f "$pkg/doc.go" ]; then
        MISSING_DOCGO="$MISSING_DOCGO $pkg"
    fi
done
if [ -z "$MISSING_DOCGO" ]; then
    pass
else
    fail
    echo "  缺少 doc.go 的核心包:$MISSING_DOCGO"
fi

# 12. 核心包第三方依赖检查
echo -n "12. 核心包零外部依赖... "
MODULE_PATH=$(head -1 go.mod | awk '{print $2}')
THIRD_PARTY_IMPORTS=""
for pkg in $CORE_PKGS; do
    if [ -d "$pkg" ]; then
        imports=$(grep -rnE '"(github\.com|golang\.org|google\.golang\.org|go\.uber\.org|go\.opentelemetry\.io)/' "$pkg"/*.go 2>/dev/null | grep -v "_test.go" | grep -v "doc.go" | grep -v "$MODULE_PATH" | head -1 || true)
        if [ -n "$imports" ]; then
            THIRD_PARTY_IMPORTS="$THIRD_PARTY_IMPORTS\n$imports"
        fi
    fi
done
if [ -z "$THIRD_PARTY_IMPORTS" ]; then
    pass
else
    fail
    echo -e "$THIRD_PARTY_IMPORTS" | head -5
fi

# 13. 构造函数返回接口检查
echo -n "13. 构造函数返回接口... "
CONCRETE_CTOR=""
while IFS= read -r line; do
    file=$(echo "$line" | cut -d: -f1)
    pkg=$(dirname "$file")
    ctor_return=$(echo "$line" | grep -oE '\*+[A-Z][A-Za-z0-9]*' | head -1)
    if [ -z "$ctor_return" ]; then
        continue
    fi
    type_name=$(echo "$ctor_return" | tr -d '*')
    if [ -f "$pkg/doc.go" ] && grep -qE "type ${type_name} interface" "$pkg/doc.go" 2>/dev/null; then
        CONCRETE_CTOR="$CONCRETE_CTOR\n$line"
    fi
done < <(grep -rnE "func New[A-Z].+\*+[A-Z][a-z]" --include="*.go" --exclude-dir=docs --exclude-dir=vendor --exclude-dir=.git --exclude-dir=examples --exclude-dir=starter . 2>/dev/null | grep -v "_test.go" | grep -v "doc.go" | head -10 || true)
if [ -z "$CONCRETE_CTOR" ]; then
    pass
else
    warn "构造函数返回具体类型但 doc.go 中有对应接口"
    echo -e "$CONCRETE_CTOR"
fi

# 14. godoc 覆盖率检查
echo -n "14. godoc 覆盖率... "
total_exported=$(grep -rnE "^(func|type|var|const) [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
documented=$(grep -rnE "^// [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
if [ "$total_exported" -gt 0 ]; then
    godoc_pct=$((documented * 100 / total_exported))
else
    godoc_pct=0
fi
if [ "$godoc_pct" -ge 80 ]; then
    pass
    echo "  覆盖率 ${godoc_pct}%"
elif [ "$godoc_pct" -ge 50 ]; then
    warn "覆盖率 ${godoc_pct}%（建议 ≥80%）"
else
    warn "覆盖率 ${godoc_pct}%（过低）"
fi

# 15. AI 文档完整性检查
echo -n "15. AI 文档完整性... "
AI_DOCS_OK=true
[ ! -f "docs/AI_INDEX.md" ] && AI_DOCS_OK=false
[ ! -f "docs/AI_QUICK_REF.md" ] && AI_DOCS_OK=false
[ ! -f "docs/AI_NAVIGATION_MAP.md" ] && AI_DOCS_OK=false
[ ! -f "docs/AI_CHANGE_IMPACT.md" ] && AI_DOCS_OK=false
[ ! -f "docs/AI_PITFALLS.md" ] && AI_DOCS_OK=false
[ ! -f "docs/AI_TASK_HANDBOOK.md" ] && AI_DOCS_OK=false
if [ "$AI_DOCS_OK" = true ]; then
    pass
else
    warn "缺少 AI 文档"
    [ ! -f "docs/AI_INDEX.md" ] && echo "  缺少 docs/AI_INDEX.md"
    [ ! -f "docs/AI_QUICK_REF.md" ] && echo "  缺少 docs/AI_QUICK_REF.md"
    [ ! -f "docs/AI_NAVIGATION_MAP.md" ] && echo "  缺少 docs/AI_NAVIGATION_MAP.md"
    [ ! -f "docs/AI_CHANGE_IMPACT.md" ] && echo "  缺少 docs/AI_CHANGE_IMPACT.md"
    [ ! -f "docs/AI_PITFALLS.md" ] && echo "  缺少 docs/AI_PITFALLS.md"
    [ ! -f "docs/AI_TASK_HANDBOOK.md" ] && echo "  缺少 docs/AI_TASK_HANDBOOK.md"
fi

# 16. 刷新审计数据
echo ""
echo -n "刷新审计数据... "
if [ -d "./cmd/audit" ]; then
    if go run ./cmd/audit -dir . -out docs/AI_READABILITY_AUDIT.md -json "$AUDIT_JSON" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 已刷新${NC}"
    else
        echo -e "${YELLOW}⚠ 审计刷新失败${NC}"
    fi
else
    echo -e "${YELLOW}⚠ 跳过（cmd/audit 不存在）${NC}"
fi

# 汇总
echo ""
echo "=== 检查汇总 ==="
echo -e "  通过: ${GREEN}$PASS_COUNT${NC}  警告: ${YELLOW}$WARN_COUNT${NC}  失败: ${RED}$FAIL_COUNT${NC}"
if [ "$FAIL_COUNT" -gt 0 ]; then
    echo -e "  ${RED}存在硬性违规，请修复后再提交${NC}"
    exit 1
elif [ "$WARN_COUNT" -gt 0 ]; then
    echo -e "  ${YELLOW}存在警告项，建议修复${NC}"
else
    echo -e "  ${GREEN}全部通过 ✓${NC}"
fi
echo ""
echo "=== 检查完成 ==="