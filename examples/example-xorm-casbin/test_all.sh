#!/bin/bash

# XORM + Casbin-XORM 集成测试脚本
# 该脚本验证所有 API 端点是否正常工作

BASE_URL="http://localhost:8081"
PASS=0
FAIL=0

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}XORM + Casbin-XORM 集成测试${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    local token=$5
    local expected_code=$6

    echo -n "测试: $name ... "

    if [ -n "$token" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            ${data:+-d "$data"})
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            ${data:+-d "$data"})
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC} (HTTP $http_code)"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL${NC} (HTTP $http_code, expected $expected_code)"
        echo "  Response: $body"
        FAIL=$((FAIL + 1))
    fi
}

# 1. 测试健康检查
echo -e "\n${YELLOW}--- 健康检查 ---${NC}"
test_api "健康检查" "GET" "$BASE_URL/api/health" "" "" "200"

# 2. 测试登录
echo -e "\n${YELLOW}--- 认证测试 ---${NC}"
login_response=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"test"}')

token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$token" ]; then
    echo -e "${GREEN}PASS${NC} - 登录成功，获取 Token"
    PASS=$((PASS + 1))
else
    echo -e "${RED}FAIL${NC} - 登录失败，未获取到 Token"
    echo "  Response: $login_response"
    FAIL=$((FAIL + 1))
fi

# 3. 测试需要认证的接口
echo -e "\n${YELLOW}--- 授权测试 ---${NC}"
test_api "用户资料（带 Token）" "GET" "$BASE_URL/api/profile" "" "$token" "200"
test_api "用户资料（无 Token）" "GET" "$BASE_URL/api/profile" "" "" "401"

# 4. 测试管理员接口
echo -e "\n${YELLOW}--- 角色权限测试 ---${NC}"
test_api "管理员接口（admin 角色）" "GET" "$BASE_URL/api/admin/users" "" "$token" "200"

# 5. 测试用户 CRUD
echo -e "\n${YELLOW}--- 用户 CRUD 测试 ---${NC}"
test_api "获取用户列表" "GET" "$BASE_URL/api/users" "" "$token" "200"

# 创建用户（使用时间戳确保邮箱唯一）
TEST_EMAIL="test-$(date +%s)@example.com"
create_response=$(curl -s -X POST "$BASE_URL/api/users" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d "{\"name\":\"Test User\",\"email\":\"$TEST_EMAIL\",\"age\":30}")

create_id=$(echo "$create_response" | grep -o '"id":[0-9]*' | cut -d':' -f2)

if [ -n "$create_id" ]; then
    echo -e "${GREEN}PASS${NC} - 创建用户成功 (ID: $create_id)"
    PASS=$((PASS + 1))
else
    echo -e "${RED}FAIL${NC} - 创建用户失败"
    echo "  Response: $create_response"
    FAIL=$((FAIL + 1))
fi

# 获取单个用户
if [ -n "$create_id" ]; then
    test_api "获取单个用户" "GET" "$BASE_URL/api/users/$create_id" "" "$token" "200"
    
    # 更新用户
    test_api "更新用户" "PUT" "$BASE_URL/api/users/$create_id" \
        '{"name":"Updated User","email":"updated@example.com","age":31}' "$token" "200"
    
    # 删除用户
    test_api "删除用户" "DELETE" "$BASE_URL/api/users/$create_id" "" "$token" "200"
fi

# 6. 测试 Casbin 策略管理
echo -e "\n${YELLOW}--- Casbin 策略管理测试 ---${NC}"
test_api "获取策略列表" "GET" "$BASE_URL/api/casbin/policies" "" "$token" "200"

# 添加策略
test_api "添加策略" "POST" "$BASE_URL/api/casbin/policy" \
    '{"subject":"testuser","object":"/api/test","action":"GET"}' "$token" "200"

# 删除策略
test_api "删除策略" "DELETE" "$BASE_URL/api/casbin/policy" \
    '{"subject":"testuser","object":"/api/test","action":"GET"}' "$token" "200"

# 总结
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${YELLOW}测试总结${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo -e "总计: $((PASS + FAIL))"

if [ $FAIL -eq 0 ]; then
    echo -e "\n${GREEN}所有测试通过！${NC}"
    exit 0
else
    echo -e "\n${RED}部分测试失败，请检查日志。${NC}"
    exit 1
fi