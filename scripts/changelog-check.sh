#!/bin/bash
# API 变更检查脚本
# 检查代码变更是否记录在 CHANGELOG.md 中

set -e

echo "=== API 变更检查 ==="
echo ""

# 检查是否有未记录的接口变更
check_unrecorded_changes() {
    echo "检查未记录的接口变更..."

    # 获取最近的提交
    local recent_commits=$(git log --oneline -10)

    echo "最近 10 次提交:"
    echo "$recent_commits"
    echo ""

    # 检查是否修改了 doc.go 文件
    local doc_changes=$(git diff --name-only HEAD~10 HEAD | grep "doc.go" || true)

    if [ -n "$doc_changes" ]; then
        echo "⚠ 发现 doc.go 文件变更:"
        echo "$doc_changes"
        echo ""
        echo "请确保在 CHANGELOG.md 中记录这些 API 变更"
    else
        echo "✓ 未发现 doc.go 文件变更"
    fi
}

# 检查版本号一致性
check_version_consistency() {
    echo "检查版本号一致性..."

    # 从 CHANGELOG.md 提取最新版本号
    local latest_version=$(sed -nE 's/.*## \[([0-9]+\.[0-9]+\.[0-9]+).*/\1/p' CHANGELOG.md | head -1)

    if [ -z "$latest_version" ]; then
        echo "⚠ 无法从 CHANGELOG.md 提取版本号"
        return
    fi

    echo "最新版本号: $latest_version"

    # 检查 git tags
    local latest_tag=$(git tag -l | sort -V | tail -1)

    if [ -n "$latest_tag" ]; then
        echo "最新 git tag: $latest_tag"

        if [ "$latest_version" != "$latest_tag" ]; then
            echo "⚠ 版本号不一致"
            echo "  CHANGELOG.md: $latest_version"
            echo "  git tag: $latest_tag"
        else
            echo "✓ 版本号一致"
        fi
    else
        echo "⚠ 未找到 git tags"
    fi
}

# 主检查函数
main() {
    echo "开始 API 变更检查..."
    echo ""

    check_unrecorded_changes
    echo ""

    check_version_consistency
    echo ""

    echo "=== API 变更检查完成 ==="
}

# 运行主函数
main
