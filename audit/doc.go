// Package audit 提供审计日志功能，用于 enhance 框架。
//
// 该模块提供操作审计日志记录功能，支持记录用户操作、数据变更、安全事件等。
// 参考 Spring Boot 的 Spring Audit 设计。
//
// # 架构设计
//
//   - Auditor: 审计日志器接口，负责记录审计事件
//   - Event: 审计事件，包含事件类型、操作人、资源等信息
//   - EventWriter: 事件写入器接口，支持控制台、文件等多种写入方式
//   - AuditInterceptor: 审计拦截器接口，用于自动审计
//   - AuditLogger: 审计日志助手接口，简化审计日志记录
//   - EventType/EventSeverity: 事件类型和严重程度枚举
//
// # 核心功能
//
//   - 事件类型: 支持 CREATE/UPDATE/DELETE/LOGIN/SECURITY 等多种事件类型
//   - 异步处理: 支持异步事件写入，提高性能
//   - 多写入器: 支持控制台、文件等多种写入方式
//   - 拦截器: 提供 AuditInterceptor 用于自动审计
//   - 日志助手: 提供 AuditLogger 简化审计日志记录
//
// # 使用方式
//
// 创建审计日志器：
//
//	auditor := audit.NewAuditor(audit.WithWriter(consoleWriter))
//
// 记录操作日志：
//
//	auditor.Log(audit.Event{
//	    Actor:    "user123",
//	    Action:   audit.EventCreate,
//	    Resource: "user",
//	    Target:   "user:456",
//	    Details:  map[string]any{"name": "John"},
//	})
//
// 使用审计日志助手：
//
//	logger := audit.NewAuditLogger(auditor, "user123", "web-app")
//	logger.Create("user", "user:456", map[string]any{"name": "John"})
//
// # 异步模式
//
//	auditor := audit.NewAuditor(
//	    audit.WithWriter(fileWriter),
//	    audit.WithAsync(),
//	    audit.WithBufferSize(1000),
//	)
//	defer auditor.Close()
//
// # 设计原则
//
// 核心框架零外部依赖，仅使用 Go 标准库。
package audit

import (
	"errors"
	"time"
)

// EventType 事件类型。
type EventType string

// 内置事件类型。
const (
	// EventCreate 创建事件。
	EventCreate EventType = "CREATE"
	// EventUpdate 更新事件。
	EventUpdate EventType = "UPDATE"
	// EventDelete 删除事件。
	EventDelete EventType = "DELETE"
	// EventRead 读取事件。
	EventRead EventType = "READ"
	// EventLogin 登录事件。
	EventLogin EventType = "LOGIN"
	// EventLogout 登出事件。
	EventLogout EventType = "LOGOUT"
	// EventAccess 访问事件。
	EventAccess EventType = "ACCESS"
	// EventPermission 权限事件。
	EventPermission EventType = "PERMISSION"
	// EventSecurity 安全事件。
	EventSecurity EventType = "SECURITY"
	// EventCustom 自定义事件。
	EventCustom EventType = "CUSTOM"
)

// EventSeverity 事件严重程度。
type EventSeverity string

// 内置严重程度。
const (
	// SeverityInfo 信息级别。
	SeverityInfo EventSeverity = "INFO"
	// SeverityWarning 警告级别。
	SeverityWarning EventSeverity = "WARNING"
	// SeverityError 错误级别。
	SeverityError EventSeverity = "ERROR"
	// SeverityCritical 严重级别。
	SeverityCritical EventSeverity = "CRITICAL"
)

// Event 审计事件。
type Event struct {
	// ID 事件 ID。
	ID string `json:"id"`
	// Timestamp 事件时间戳。
	Timestamp time.Time `json:"timestamp"`
	// Actor 操作者（用户 ID 或系统）。
	Actor string `json:"actor"`
	// Action 操作类型。
	Action EventType `json:"action"`
	// Resource 资源类型。
	Resource string `json:"resource"`
	// Target 操作目标。
	Target string `json:"target,omitempty"`
	// Details 详细信息。
	Details map[string]any `json:"details,omitempty"`
	// Severity 严重程度。
	Severity EventSeverity `json:"severity"`
	// Source 事件来源（IP 地址、服务名等）。
	Source string `json:"source,omitempty"`
	// Result 操作结果（success/failure）。
	Result string `json:"result,omitempty"`
	// ErrorMessage 错误信息。
	ErrorMessage string `json:"errorMessage,omitempty"`
	// Duration 操作耗时。
	Duration time.Duration `json:"duration,omitempty"`
	// Tags 标签。
	Tags []string `json:"tags,omitempty"`
}

// EventWriter 事件写入器接口。
//
// 用于将审计事件写入到不同的目标（控制台、文件等）。
// 实现应保证线程安全和写入的原子性。
type EventWriter interface {
	// Write 写入事件。
	Write(event Event) error

	// Close 关闭写入器，释放资源。
	Close() error
}

// Auditor 审计日志器接口。
//
// 负责记录审计事件，支持同步和异步两种写入模式。
// 异步模式下，事件会先写入缓冲区，由后台 goroutine 处理。
type Auditor interface {
	// Log 记录审计事件。
	Log(event Event)

	// Close 关闭审计日志器，释放资源。
	Close() error

	// IsClosed 检查审计日志器是否已关闭。
	IsClosed() bool
}

// AuditInterceptor 审计拦截器接口。
//
// 用于拦截方法调用并自动记录审计日志。
// 通常与 AOP 框架配合使用，实现声明式审计。
type AuditInterceptor interface {
	// Intercept 拦截方法调用并记录审计日志。
	Intercept(methodName string, args []any, result any, err error)
}

// AuditLogger 审计日志助手接口。
//
// 提供便捷的审计日志记录方法，封装常用的审计场景。
// 通过预设操作者和来源信息，简化日志记录调用。
type AuditLogger interface {
	// Create 记录创建事件。
	Create(resource string, target string, details map[string]any)

	// Update 记录更新事件。
	Update(resource string, target string, details map[string]any)

	// Delete 记录删除事件。
	Delete(resource string, target string)

	// Login 记录登录事件。
	Login(target string, details map[string]any)

	// Security 记录安全事件。
	Severity(resource string, target string, severity EventSeverity, details map[string]any)
}

// AuditorOption 审计器选项函数。
type AuditorOption func(auditor Auditor)

// 审计模块错误变量。
var (
	// ErrWriterClosed 表示写入器已关闭。
	ErrWriterClosed = errors.New("audit writer is closed")
	// ErrChannelFull 表示事件通道已满。
	ErrChannelFull = errors.New("audit event channel is full")
)
