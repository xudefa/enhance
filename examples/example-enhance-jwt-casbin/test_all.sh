#!/bin/bash
# test_all.sh - 快速测试所有接口
# 使用方法: chmod +x test_all.sh && ./test_all.sh

BASE_URL="http://localhost:8080"

echo "=========================================="
echo "  JWT + Casbin 示例项目 - 接口测试"
echo "=========================================="

# 1. 登录
echo -e "\n=== 1. Alice 登录 (admin 角色) ==="
ALICE_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${ALICE_TOKEN:0:30}..."

echo -e "\n=== 2. Bob 登录 (user 角色) ==="
BOB_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${BOB_TOKEN:0:30}..."

# 2. 测试接口
echo -e "\n=== 3. Alice 查看个人资料 ==="
curl -s $BASE_URL/api/profile -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 4. Bob 查看个人资料 ==="
curl -s $BASE_URL/api/profile -H "Authorization: Bearer $BOB_TOKEN" | python3 -m json.tool

echo -e "\n=== 5. Alice 访问管理员接口 (应该有权限) ==="
curl -s $BASE_URL/api/admin/users -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 6. Bob 访问管理员接口 (应该 403) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/admin/users -H "Authorization: Bearer $BOB_TOKEN"

echo -e "\n=== 7. 未登录访问 (应该 401) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/profile

echo -e "\n=========================================="
echo "  测试完成！"
echo "=========================================="