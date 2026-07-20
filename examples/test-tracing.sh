#!/bin/bash

# 分布式链路追踪示例测试脚本
# 测试所有 Web 框架（Gin, Fiber, Echo, Chi）的 tracing 集成

set -e

echo "========================================="
echo "  分布式链路追踪集成测试"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=$3
    
    echo -e "${YELLOW}测试: ${name}${NC}"
    echo "URL: ${url}"
    
    response=$(curl -s -w "\n%{http_code}" ${url} 2>/dev/null || echo -e "\n000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ 状态码: ${http_code} (预期: ${expected_status})${NC}"
    else
        echo -e "${RED}✗ 状态码: ${http_code} (预期: ${expected_status})${NC}"
        return 1
    fi
    
    echo "响应: ${body}"
    echo ""
}

# 等待服务启动
wait_for_service() {
    local name=$1
    local url=$2
    local max_attempts=10
    local attempt=1
    
    echo -e "${YELLOW}等待 ${name} 启动...${NC}"
    while [ $attempt -le $max_attempts ]; do
        if curl -s ${url} > /dev/null 2>&1; then
            echo -e "${GREEN}✓ ${name} 已就绪${NC}"
            return 0
        fi
        echo "  尝试 ${attempt}/${max_attempts}..."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}✗ ${name} 启动超时${NC}"
    return 1
}

# 清理函数
cleanup() {
    echo ""
    echo -e "${YELLOW}清理进程...${NC}"
    pkill -f "gin-tracing-example" 2>/dev/null || true
    pkill -f "fiber-tracing-example" 2>/dev/null || true
    pkill -f "echo-tracing-example" 2>/dev/null || true
    pkill -f "chi-tracing-example" 2>/dev/null || true
    echo -e "${GREEN}✓ 所有进程已清理${NC}"
}

trap cleanup EXIT

# 启动所有示例服务
echo -e "${YELLOW}启动示例服务...${NC}"
echo ""

cd "$(dirname "$0")"

echo "启动 Gin Tracing 示例 (端口 8081)..."
cd example-gin-tracing
go run main.go &
GIN_PID=$!
cd ..

echo "启动 Fiber Tracing 示例 (端口 8082)..."
cd example-fiber-tracing
go run main.go &
FIBER_PID=$!
cd ..

echo "启动 Echo Tracing 示例 (端口 8083)..."
cd example-echo-tracing
go run main.go &
ECHO_PID=$!
cd ..

echo "启动 Chi Tracing 示例 (端口 8084)..."
cd example-chi-tracing
go run main.go &
CHI_PID=$!
cd ..

echo ""
echo "等待服务启动..."
sleep 5

echo ""
echo "========================================="
echo "  开始测试"
echo "========================================="
echo ""

# 测试 Gin
echo "========================================="
echo "  Gin Tracing (端口 8081)"
echo "========================================="
wait_for_service "Gin" "http://localhost:8081/api/hello"
test_endpoint "Gin - Hello" "http://localhost:8081/api/hello" "200"
test_endpoint "Gin - Error" "http://localhost:8081/api/error" "500"
test_endpoint "Gin - Spans" "http://localhost:8081/api/spans" "200"
echo ""

# 测试 Fiber
echo "========================================="
echo "  Fiber Tracing (端口 8082)"
echo "========================================="
wait_for_service "Fiber" "http://localhost:8082/api/hello"
test_endpoint "Fiber - Hello" "http://localhost:8082/api/hello" "200"
test_endpoint "Fiber - Error" "http://localhost:8082/api/error" "500"
test_endpoint "Fiber - Spans" "http://localhost:8082/api/spans" "200"
echo ""

# 测试 Echo
echo "========================================="
echo "  Echo Tracing (端口 8083)"
echo "========================================="
wait_for_service "Echo" "http://localhost:8083/api/hello"
test_endpoint "Echo - Hello" "http://localhost:8083/api/hello" "200"
test_endpoint "Echo - Error" "http://localhost:8083/api/error" "500"
test_endpoint "Echo - Spans" "http://localhost:8083/api/spans" "200"
echo ""

# 测试 Chi
echo "========================================="
echo "  Chi Tracing (端口 8084)"
echo "========================================="
wait_for_service "Chi" "http://localhost:8084/api/hello"
test_endpoint "Chi - Hello" "http://localhost:8084/api/hello" "200"
test_endpoint "Chi - Error" "http://localhost:8084/api/error" "500"
test_endpoint "Chi - Spans" "http://localhost:8084/api/spans" "200"
echo ""

echo "========================================="
echo -e "${GREEN}  所有测试完成！${NC}"
echo "========================================="
echo ""
echo "提示："
echo "- 访问 /api/spans 查看收集的链路追踪数据"
echo "- 每个请求都会自动生成 Span 记录"
echo "- 错误请求（>=400）会被标记为 ERROR 状态"
echo ""