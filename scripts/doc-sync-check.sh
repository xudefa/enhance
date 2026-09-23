#!/bin/bash
# 文档同步检查脚本
# 检查文档中的接口/类型定义与实际代码是否一致

set -e

echo "=== 文档同步检查 ==="
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SYNC_ISSUES=0

# 1. 检查 doc.go 中的接口是否在代码中实现
echo "1. 检查 doc.go 接口实现..."
DOC_GO_FILES=$(find . -name "doc.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./cmd/*" -not -path "./examples/*" -not -path "./starter/*" 2>/dev/null || true)
for doc_file in $DOC_GO_FILES; do
    package_dir=$(dirname "$doc_file")
    package_name=$(basename "$package_dir")

    interfaces=$(grep -o "type [A-Z][a-zA-Z]* interface" "$doc_file" 2>/dev/null | awk '{print $2}' || true)

    for interface in $interfaces; do
        impl_count=$(grep -rn "func.*$interface)" "$package_dir"/*.go 2>/dev/null | grep -v "doc.go" | wc -l | tr -d ' ')
        if [ "$impl_count" -eq 0 ]; then
            method_count=$(grep -c "^[[:space:]]" "$doc_file" 2>/dev/null || echo 0)
            if [ "$method_count" -le 5 ]; then
                echo -e "${YELLOW}⚠ $package_name.$interface 可能没有实现${NC}"
                SYNC_ISSUES=$((SYNC_ISSUES + 1))
            fi
        fi
    done
done
if [ "$SYNC_ISSUES" -eq 0 ]; then
    echo -e "${GREEN}✓ doc.go 接口实现检查通过${NC}"
else
    echo -e "${YELLOW}⚠ $SYNC_ISSUES 个接口可能缺少实现（可能通过隐式实现，需人工确认）${NC}"
fi

# 2. 检查 ARCHITECTURE.md 中的包是否都存在
echo ""
echo "2. 检查 ARCHITECTURE.md 中的包..."
if [ -f "ARCHITECTURE.md" ]; then
    packages=$(grep -oE '`[a-z]+/`' ARCHITECTURE.md | tr -d '`' | sed 's/\/$//')
    missing_pkgs=0
    for pkg in $packages; do
        if [ ! -d "$pkg" ]; then
            echo -e "${YELLOW}⚠ 包 $pkg 在 ARCHITECTURE.md 中提到但不存在${NC}"
            missing_pkgs=$((missing_pkgs + 1))
        fi
    done
    if [ "$missing_pkgs" -eq 0 ]; then
        echo -e "${GREEN}✓ ARCHITECTURE.md 包引用检查通过${NC}"
    fi
else
    echo -e "${YELLOW}⚠ ARCHITECTURE.md 不存在${NC}"
fi

# 3. 检查 AI_NAVIGATION_MAP.md 中的文件路径
echo ""
echo "3. 检查 AI_NAVIGATION_MAP.md 中的文件路径..."
if [ -f "docs/AI_NAVIGATION_MAP.md" ]; then
    paths=$(grep -oE '`[^`]+\.go`' docs/AI_NAVIGATION_MAP.md | tr -d '`' | sort -u)
    missing_files=0
    for path in $paths; do
        case "$path" in
            */*)  ;;
            *)    continue ;;
        esac
        if [ ! -f "$path" ]; then
            echo -e "${YELLOW}⚠ 文件 $path 在 AI_NAVIGATION_MAP.md 中提到但不存在${NC}"
            missing_files=$((missing_files + 1))
        fi
    done
    if [ "$missing_files" -eq 0 ]; then
        echo -e "${GREEN}✓ AI_NAVIGATION_MAP.md 文件引用检查通过${NC}"
    else
        echo -e "${YELLOW}⚠ $missing_files 个文件路径不存在${NC}"
    fi
else
    echo -e "${YELLOW}⚠ docs/AI_NAVIGATION_MAP.md 不存在${NC}"
fi

# 4. 检查 AI_CHANGE_IMPACT.md 中的包引用
echo ""
echo "4. 检查 AI_CHANGE_IMPACT.md 中的包引用..."
if [ -f "docs/AI_CHANGE_IMPACT.md" ]; then
    impact_pkgs=$(grep -oE '`[a-z]+`' docs/AI_CHANGE_IMPACT.md | tr -d '`' | sort -u)
    missing_impact=0
    for pkg in $impact_pkgs; do
        if [ -n "$pkg" ] && [ ${#pkg} -ge 3 ] && [ ! -d "$pkg" ] && [ "$pkg" != "cmd" ] && [ "$pkg" != "enhance" ]; then
            if echo "$pkg" | grep -qE '^[a-z]+$'; then
                echo -e "${YELLOW}⚠ 包 $pkg 在 AI_CHANGE_IMPACT.md 中提到但不存在${NC}"
                missing_impact=$((missing_impact + 1))
            fi
        fi
    done
    if [ "$missing_impact" -eq 0 ]; then
        echo -e "${GREEN}✓ AI_CHANGE_IMPACT.md 包引用检查通过${NC}"
    fi
else
    echo -e "${YELLOW}⚠ docs/AI_CHANGE_IMPACT.md 不存在${NC}"
fi

# 5. 检查核心包 doc.go 存在性
echo ""
echo "5. 检查核心包 doc.go 存在性..."
CORE_PKGS="core boot context condition config event cache web security log metrics actuator schedule resilience validation exception lifecycle async audit retry i18n spel testing webtest openapi metadata devtools email mq tenant"
missing_docgo=0
for pkg in $CORE_PKGS; do
    if [ -d "$pkg" ] && [ ! -f "$pkg/doc.go" ]; then
        echo -e "${RED}✗ 核心包 $pkg 缺少 doc.go${NC}"
        missing_docgo=$((missing_docgo + 1))
    fi
done
if [ "$missing_docgo" -eq 0 ]; then
    echo -e "${GREEN}✓ 核心包 doc.go 存在性检查通过${NC}"
fi

# 6. 检查核心包 README.md 存在性
echo ""
echo "6. 检查核心包 README.md 存在性..."
missing_readme=0
for pkg in $CORE_PKGS; do
    if [ -d "$pkg" ] && [ ! -f "$pkg/README.md" ]; then
        echo -e "${YELLOW}⚠ 核心包 $pkg 缺少 README.md${NC}"
        missing_readme=$((missing_readme + 1))
    fi
done
if [ "$missing_readme" -eq 0 ]; then
    echo -e "${GREEN}✓ 核心包 README.md 存在性检查通过${NC}"
fi

# 7. AI 文档完整性检查
echo ""
echo "7. 检查 AI 文档完整性..."
AI_DOCS="docs/AI_INDEX.md docs/AI_QUICK_REF.md docs/AI_NAVIGATION_MAP.md docs/AI_CHANGE_IMPACT.md docs/AI_PITFALLS.md docs/AI_TASK_HANDBOOK.md"
missing_ai_docs=0
for doc in $AI_DOCS; do
    if [ ! -f "$doc" ]; then
        echo -e "${YELLOW}⚠ 缺少 $doc${NC}"
        missing_ai_docs=$((missing_ai_docs + 1))
    fi
done
if [ "$missing_ai_docs" -eq 0 ]; then
    echo -e "${GREEN}✓ AI 文档完整性检查通过${NC}"
fi

# 8. godoc 覆盖率检查
echo ""
echo "8. 检查 godoc 覆盖率..."
total_exported=$(grep -rnE "^(func|type|var|const) [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
documented=$(grep -rnE "^// [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ')
if [ "$total_exported" -gt 0 ]; then
    godoc_pct=$((documented * 100 / total_exported))
else
    godoc_pct=0
fi
if [ "$godoc_pct" -ge 80 ]; then
    echo -e "${GREEN}✓ godoc 覆盖率 ${godoc_pct}%${NC}"
elif [ "$godoc_pct" -ge 50 ]; then
    echo -e "${YELLOW}⚠ godoc 覆盖率 ${godoc_pct}%（建议 ≥80%）${NC}"
else
    echo -e "${RED}✗ godoc 覆盖率 ${godoc_pct}%（过低）${NC}"
fi

echo ""
echo "=== 文档同步检查完成 ==="