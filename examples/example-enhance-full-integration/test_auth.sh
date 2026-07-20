#!/bin/bash
# 测试登录验证和授权功能
# 使用方法: ./test_auth.sh
# 脚本会自动启动服务，测试完成后自动关闭

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  登录验证和授权测试${NC}"
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
        echo -e "${RED}FAIL${NC} (HTTP $http_code, 期望 $expected_code)"
        echo "  响应: $body"
        FAIL=$((FAIL + 1))
    fi
}

# 1. 测试登录 - alice (admin角色)
echo -e "\n${BLUE}--- 1. Alice 登录 (admin角色) ---${NC}"
login_response=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"test"}')

echo "登录响应: $login_response"
alice_token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$alice_token" ]; then
    echo -e "${GREEN}PASS${NC} - Alice 登录成功"
    PASS=$((PASS + 1))
else
    echo -e "${RED}FAIL${NC} - Alice 登录失败"
    FAIL=$((FAIL + 1))
fi

# 2. 测试登录 - bob (user角色)
echo -e "\n${BLUE}--- 2. Bob 登录 (user角色) ---${NC}"
login_response=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"bob","password":"test"}')

echo "登录响应: $login_response"
bob_token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$bob_token" ]; then
    echo -e "${GREEN}PASS${NC} - Bob 登录成功"
    PASS=$((PASS + 1))
else
    echo -e "${RED}FAIL${NC} - Bob 登录失败"
    FAIL=$((FAIL + 1))
fi

# 3. 测试未登录访问受保护资源
echo -e "\n${BLUE}--- 3. 未登录访问受保护资源 ---${NC}"
test_api "未登录访问 /api/profile" "GET" "$BASE_URL/api/profile" "" "" "401"

# 5. 测试Alice查看个人资料
echo -e "\n${BLUE}--- 5. Alice 查看个人资料 ---${NC}"
test_api "Alice 查看个人资料" "GET" "$BASE_URL/api/profile" "" "$alice_token" "200"

# 5. 测试Bob查看个人资料
echo -e "\n${BLUE}--- 5. Bob 查看个人资料 ---${NC}"
test_api "Bob 查看个人资料" "GET" "$BASE_URL/api/profile" "" "$bob_token" "200"

# 6. 测试Alice访问管理员接口 (应该有权限)
echo -e "\n${BLUE}--- 6. Alice 访问管理员接口 (admin角色) ---${NC}"
test_api "Alice 访问 /api/admin/users" "GET" "$BASE_URL/api/admin/users" "" "$alice_token" "200"

# 7. 测试Bob访问管理员接口 (应该无权限)
echo -e "\n${BLUE}--- 7. Bob 访问管理员接口 (user角色，应该被拒绝) ---${NC}"
test_api "Bob 访问 /api/admin/users" "GET" "$BASE_URL/api/admin/users" "" "$bob_token" "403"

# 8. 测试Casbin策略管理 - Alice
echo -e "\n${BLUE}--- 8. Alice 查询Casbin策略列表 ---${NC}"
test_api "Alice 查询策略" "GET" "$BASE_URL/api/casbin/policies" "" "$alice_token" "200"

# 10. 测试添加Casbin策略
echo -e "\n${BLUE}--- 10. Alice 添加Casbin策略 ---${NC}"
test_api "Alice 添加策略" "POST" "$BASE_URL/api/casbin/policy" \
    '{"subject":"testuser","object":"/api/test","action":"GET"}' "$alice_token" "200"

# 10. 测试权限检查
echo -e "\n${BLUE}--- 10. Alice 权限检查 ---${NC}"
test_api "权限检查" "POST" "$BASE_URL/api/casbin/check" \
    '{"subject":"admin","object":"/api/admin/users","action":"GET"}' "$alice_token" "200"

# 11. 测试删除Casbin策略
echo -e "\n${BLUE}--- 11. Alice 删除Casbin策略 ---${NC}"
test_api "Alice 删除策略" "DELETE" "$BASE_URL/api/casbin/policy" \
    '{"subject":"testuser","object":"/api/test","action":"GET"}' "$alice_token" "200"

# 总结
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${YELLOW}测试总结${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo -e "总计: $((PASS + FAIL))"

if [ $FAIL -eq 0 ]; then
    echo -e "\n${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "\n${RED}✗ 部分测试失败，请检查日志。${NC}"
    exit 1
fi