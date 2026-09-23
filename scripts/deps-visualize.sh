#!/bin/bash
# 包间依赖可视化脚本
# 生成 ASCII 格式的依赖关系图

set -e

echo "=== 包间依赖可视化 ==="
echo ""

# 生成 ASCII 依赖图
generate_ascii_graph() {
    echo "生成 ASCII 依赖图..."
    echo ""
    echo "三层架构依赖关系："
    echo "=================="
    echo ""
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                    Boot Layer                               │"
    echo "│  ┌─────────┐  ┌─────────┐  ┌─────────┐                    │"
    echo "│  │  boot   │  │context  │  │condition│                    │"
    echo "│  └────┬────┘  └────┬────┘  └────┬────┘                    │"
    echo "│       │            │            │                           │"
    echo "└───────┼────────────┼────────────┼───────────────────────────┘"
    echo "        │            │            │"
    echo "        ▼            ▼            ▼"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                    Core Layer                               │"
    echo "│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │"
    echo "│  │  core   │  │  event  │  │ config  │  │lifecycle│     │"
    echo "│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘     │"
    echo "│       │            │            │            │            │"
    echo "└───────┼────────────┼────────────┼────────────┼────────────┘"
    echo "        │            │            │            │"
    echo "        ▼            ▼            ▼            ▼"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                Infrastructure Layer                         │"
    echo "│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │"
    echo "│  │   web   │  │security │  │  cache  │  │ schedule│     │"
    echo "│  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │"
    echo "│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │"
    echo "│  │   log   │  │ metrics │  │ actuator│  │ tracing │     │"
    echo "│  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
}

# 生成 Mermaid 格式依赖图
generate_mermaid_graph() {
    echo "生成 Mermaid 依赖图..."
    echo ""
    echo "graph TD"
    echo "    subgraph Boot Layer"
    echo "        boot[boot]"
    echo "        context[context]"
    echo "        condition[condition]"
    echo "    end"
    echo ""
    echo "    subgraph Core Layer"
    echo "        core[core]"
    echo "        event[event]"
    echo "        config[config]"
    echo "        lifecycle[lifecycle]"
    echo "    end"
    echo ""
    echo "    subgraph Infrastructure Layer"
    echo "        web[web]"
    echo "        security[security]"
    echo "        cache[cache]"
    echo "        schedule[schedule]"
    echo "        log[log]"
    echo "        metrics[metrics]"
    echo "        actuator[actuator]"
    echo "        tracing[tracing]"
    echo "    end"
    echo ""
    echo "    boot --> core"
    echo "    context --> core"
    echo "    condition --> core"
    echo ""
    echo "    core --> web"
    echo "    core --> security"
    echo "    core --> cache"
    echo "    core --> schedule"
    echo "    core --> log"
    echo "    core --> metrics"
    echo "    core --> actuator"
    echo "    core --> tracing"
    echo ""
}

# 生成 JSON 格式依赖数据
generate_json_data() {
    echo "生成 JSON 依赖数据..."
    echo ""
    echo "{"
    echo "  \"layers\": {"
    echo "    \"boot\": [\"boot\", \"context\", \"condition\"],"
    echo "    \"core\": [\"core\", \"event\", \"config\", \"lifecycle\"],"
    echo "    \"infrastructure\": [\"web\", \"security\", \"cache\", \"schedule\", \"log\", \"metrics\", \"actuator\", \"tracing\"]"
    echo "  },"
    echo "  \"dependencies\": ["
    echo "    {\"from\": \"boot\", \"to\": \"core\"},"
    echo "    {\"from\": \"context\", \"to\": \"core\"},"
    echo "    {\"from\": \"condition\", \"to\": \"core\"},"
    echo "    {\"from\": \"core\", \"to\": \"web\"},"
    echo "    {\"from\": \"core\", \"to\": \"security\"},"
    echo "    {\"from\": \"core\", \"to\": \"cache\"},"
    echo "    {\"from\": \"core\", \"to\": \"schedule\"},"
    echo "    {\"from\": \"core\", \"to\": \"log\"},"
    echo "    {\"from\": \"core\", \"to\": \"metrics\"},"
    echo "    {\"from\": \"core\", \"to\": \"actuator\"},"
    echo "    {\"from\": \"core\", \"to\": \"tracing\"}"
    echo "  ]"
    echo "}"
}

# 主函数
main() {
    local format="${1:-ascii}"
    
    case "$format" in
        ascii)
            generate_ascii_graph
            ;;
        mermaid)
            generate_mermaid_graph
            ;;
        json)
            generate_json_data
            ;;
        *)
            echo "用法: $0 [ascii|mermaid|json]"
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"