// Package audit 提供审计日志功能，用于 enhance 框架。
package audit

import "time"

// auditInterceptorImpl AuditInterceptor 接口的默认实现。
type auditInterceptorImpl struct {
	auditor    Auditor
	actorFunc  func() string
	sourceFunc func() string
}

// NewAuditInterceptor 创建审计拦截器
//
// 默认操作者为 "system",来源为 "unknown"。
func NewAuditInterceptor(auditor Auditor) AuditInterceptor {
	return &auditInterceptorImpl{
		auditor: auditor,
		actorFunc: func() string {
			return "system"
		},
		sourceFunc: func() string {
			return "unknown"
		},
	}
}

func (i *auditInterceptorImpl) Intercept(methodName string, args []any, result any, err error) {
	startTime := time.Now()

	event := Event{
		Actor:     i.actorFunc(),
		Action:    EventAccess,
		Resource:  methodName,
		Severity:  SeverityInfo,
		Source:    i.sourceFunc(),
		Timestamp: startTime,
		Details: map[string]any{
			"args": args,
		},
	}

	event.Duration = time.Since(startTime)

	if err != nil {
		event.Severity = SeverityError
		event.Result = "failure"
		event.ErrorMessage = err.Error()
		i.auditor.Log(event)
		return
	}
	event.Severity = SeverityInfo
	event.Result = "success"
	i.auditor.Log(event)
}
