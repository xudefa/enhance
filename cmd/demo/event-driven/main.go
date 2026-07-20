// Package main 演示 enhance 框架的事件驱动架构
//
// 该示例展示：
// - 自定义事件类型定义
// - 事件订阅与发布
// - 多监听器注册
// - 事件监听器取消订阅
package main

import (
	"fmt"
	"time"

	"github.com/xudefa/enhance/event"
)

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	eventType string
	timestamp time.Time
	UserID    int
	Username  string
}

// NewUserCreatedEvent 创建用户创建事件
func NewUserCreatedEvent(userID int, username string) *UserCreatedEvent {
	return &UserCreatedEvent{
		eventType: "user.created",
		timestamp: time.Now(),
		UserID:    userID,
		Username:  username,
	}
}

// Type 返回事件类型
func (e *UserCreatedEvent) Type() string {
	return e.eventType
}

// Timestamp 返回事件时间戳
func (e *UserCreatedEvent) Timestamp() time.Time {
	return e.timestamp
}

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	eventType string
	timestamp time.Time
	OrderID   string
	UserID    int
	Amount    float64
}

// NewOrderCreatedEvent 创建订单创建事件
func NewOrderCreatedEvent(orderID string, userID int, amount float64) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		eventType: "order.created",
		timestamp: time.Now(),
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
	}
}

// Type 返回事件类型
func (e *OrderCreatedEvent) Type() string {
	return e.eventType
}

// Timestamp 返回事件时间戳
func (e *OrderCreatedEvent) Timestamp() time.Time {
	return e.timestamp
}

func main() {
	fmt.Println("=== 事件驱动架构示例 ===")
	fmt.Println()

	// 创建事件总线
	bus := event.NewEventBus()

	// 订阅用户创建事件
	userCreatedHandler1 := func(e event.ApplicationEvent) {
		evt := e.(*UserCreatedEvent)
		fmt.Printf("[监听器 1] 用户创建事件: ID=%d, 用户名=%s\n", evt.UserID, evt.Username)
	}

	userCreatedHandler2 := func(e event.ApplicationEvent) {
		evt := e.(*UserCreatedEvent)
		fmt.Printf("[监听器 2] 发送欢迎邮件给用户: %s\n", evt.Username)
	}

	bus.Subscribe("user.created", userCreatedHandler1)
	bus.Subscribe("user.created", userCreatedHandler2)

	// 订阅订单创建事件
	orderCreatedHandler := func(e event.ApplicationEvent) {
		evt := e.(*OrderCreatedEvent)
		fmt.Printf("[订单监听器] 订单创建: 订单号=%s, 用户ID=%d, 金额=%.2f\n",
			evt.OrderID, evt.UserID, evt.Amount)
	}

	bus.Subscribe("order.created", orderCreatedHandler)

	// 发布用户创建事件
	fmt.Println("1. 发布用户创建事件:")
	bus.Publish(NewUserCreatedEvent(1001, "张三"))
	fmt.Println()

	// 发布订单创建事件
	fmt.Println("2. 发布订单创建事件:")
	bus.Publish(NewOrderCreatedEvent("ORD-20260709-001", 1001, 299.99))
	fmt.Println()

	// 取消订阅
	fmt.Println("3. 取消监听器 2 的订阅:")
	bus.Unsubscribe("user.created", userCreatedHandler2)
	fmt.Println()

	// 再次发布用户创建事件（只有监听器 1 会收到）
	fmt.Println("4. 再次发布用户创建事件（只有监听器 1 会收到）:")
	bus.Publish(NewUserCreatedEvent(1002, "李四"))
	fmt.Println()

	fmt.Println("=== 示例完成 ===")
}
