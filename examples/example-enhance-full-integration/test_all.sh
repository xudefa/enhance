#!/bin/bash
# test_all.sh - 完整测试所有接口
# 使用方法: chmod +x test_all.sh && ./test_all.sh

BASE_URL="http://localhost:8080"

echo "=========================================="
echo "  Enhance 全功能示例项目 - 接口测试"
echo "=========================================="

# 1. 登录
echo -e "\n=== 1. alice 登录 ==="
ALICE_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${ALICE_TOKEN:0:30}..."

echo -e "\n=== 2. bob 登录 ==="
BOB_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${BOB_TOKEN:0:30}..."

# 2. 基础接口
echo -e "\n=== 3. alice 查看个人资料 ==="
curl -s $BASE_URL/api/profile -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 4. alice 访问管理员接口 ==="
curl -s $BASE_URL/api/admin/users -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 5. bob 访问管理员接口 (应该 403) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/admin/users -H "Authorization: Bearer $BOB_TOKEN"

echo -e "\n=== 6. 未登录访问 (应该 401) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/profile

# 3. 用户管理
echo -e "\n=== 7. 获取用户列表 ==="
curl -s $BASE_URL/api/users -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 8. 创建用户 ==="
curl -s -X POST $BASE_URL/api/users \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试用户","email":"test@example.com","age":25}' 

echo -e "\n=== 9. 获取用户详情 ==="
curl -s $BASE_URL/api/users/1 -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 10. 更新用户 ==="
curl -s -X PUT $BASE_URL/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"更新用户","email":"updated@example.com","age":26}' 

echo -e "\n=== 11. 删除用户 ==="
curl -s -X DELETE $BASE_URL/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" 

# 4. Casbin 策略管理
echo -e "\n=== 12. 查看策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 13. 添加策略 ==="
curl -s -X POST $BASE_URL/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' 

echo -e "\n=== 14. 再次查看策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=== 15. 权限检查 ==="
curl -s -X POST $BASE_URL/api/casbin/check \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' 

echo -e "\n=== 16. 删除策略 ==="
curl -s -X DELETE $BASE_URL/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' 

echo -e "\n=== 17. 最终策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" 

echo -e "\n=========================================="
echo "  测试完成！"
echo "=========================================="