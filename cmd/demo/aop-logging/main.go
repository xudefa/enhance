// Package main 演示如何使用 AOP 实现日志切面
//
// 该示例展示：
// - AOP 切面定义
// - 方法拦截
// - 日志记录
// - 性能监控
package main

import (
	"fmt"
	"time"

	"github.com/xudefa/enhance/aop"
)

// UserService 用户服务
type UserService struct{}

// GetUser 获取用户
func (s *UserService) GetUser(id int) (string, error) {
	time.Sleep(10 * time.Millisecond) // 模拟数据库查询
	return fmt.Sprintf("User-%d", id), nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(name string) (int, error) {
	time.Sleep(20 * time.Millisecond) // 模拟数据库插入
	return 1, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int) error {
	time.Sleep(15 * time.Millisecond) // 模拟数据库删除
	return nil
}

func main() {
	fmt.Println("=== AOP 日志示例 ===")
	fmt.Println()

	// 创建服务实例
	userService := &UserService{}

	// 创建 AOP 代理工厂
	factory := aop.NewProxyFactory(userService)

	// 创建日志切面
	loggingAspect := func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
		method := jp.Method()
		fmt.Printf("[LOG] 调用方法: %v\n", method)
		fmt.Printf("[LOG] 参数: %v\n", jp.Args())

		start := time.Now()
		result := proceed()
		elapsed := time.Since(start)

		fmt.Printf("[LOG] 方法执行完成, 返回值: %v, 耗时: %v\n", result, elapsed)
		return result
	}

	// 创建性能监控切面
	threshold := 50 * time.Millisecond
	perfAspect := func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
		start := time.Now()
		result := proceed()
		elapsed := time.Since(start)

		if elapsed > threshold {
			fmt.Printf("[SLOW] 方法执行耗时 %v (阈值: %v)\n", elapsed, threshold)
		}
		return result
	}

	// 设置切面
	factory.SetAspects([]*aop.AspectMeta{
		{
			PointCut: aop.MatchByName("GetUser"),
			Advice:   aop.Around(loggingAspect),
			Order:    1,
		},
		{
			PointCut: aop.MatchByName("CreateUser"),
			Advice:   aop.Around(loggingAspect),
			Order:    1,
		},
		{
			PointCut: aop.MatchByName("DeleteUser"),
			Advice:   aop.Around(loggingAspect),
			Order:    1,
		},
		{
			PointCut: aop.MatchByName("GetUser"),
			Advice:   aop.Around(perfAspect),
			Order:    2,
		},
		{
			PointCut: aop.MatchByName("CreateUser"),
			Advice:   aop.Around(perfAspect),
			Order:    2,
		},
		{
			PointCut: aop.MatchByName("DeleteUser"),
			Advice:   aop.Around(perfAspect),
			Order:    2,
		},
	})

	// 获取代理对象
	// 注意：当设置了切面后，GetProxy() 返回的是 *aop.ReflectiveAopProxy
	// 需要通过 Call/CallContext 方法调用目标方法
	proxy := factory.GetProxy()

	// 调用方法（会被切面拦截）
	fmt.Println("1. 调用 GetUser:")
	result, err := proxy.(*aop.ReflectiveAopProxy).Call("GetUser", 1)
	if err != nil {
		fmt.Printf("   错误: %v\n\n", err)
	} else {
		// GetUser 返回 (string, error)，多返回值时返回 []any
		if results, ok := result.([]any); ok && len(results) > 0 {
			fmt.Printf("   结果: %s\n\n", results[0])
		}
	}

	fmt.Println("2. 调用 CreateUser:")
	result, err = proxy.(*aop.ReflectiveAopProxy).Call("CreateUser", "张三")
	if err != nil {
		fmt.Printf("   错误: %v\n\n", err)
	} else {
		if results, ok := result.([]any); ok && len(results) > 0 {
			fmt.Printf("   结果: ID=%v\n\n", results[0])
		}
	}

	fmt.Println("3. 调用 DeleteUser:")
	_, err = proxy.(*aop.ReflectiveAopProxy).Call("DeleteUser", 1)
	if err != nil {
		fmt.Printf("   错误: %v\n\n", err)
	} else {
		fmt.Println("   结果: 删除成功")
	}

	fmt.Println("=== 示例完成 ===")
}
