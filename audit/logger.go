// Package audit 提供审计日志功能，用于 enhance 框架。
package audit

// auditLoggerImpl AuditLogger 接口的默认实现。
type auditLoggerImpl struct {
	auditor Auditor
	actor   string
	source  string
}

// NewAuditLogger 创建审计日志助手。
func NewAuditLogger(auditor Auditor, actor string, source string) AuditLogger {
	return &auditLoggerImpl{
		auditor: auditor,
		actor:   actor,
		source:  source,
	}
}

func (l *auditLoggerImpl) Create(resource string, target string, details map[string]any) {
	l.auditor.Log(Event{
		Actor:    l.actor,
		Action:   EventCreate,
		Resource: resource,
		Target:   target,
		Details:  details,
		Severity: SeverityInfo,
		Source:   l.source,
		Result:   "success",
	})
}

func (l *auditLoggerImpl) Update(resource string, target string, details map[string]any) {
	l.auditor.Log(Event{
		Actor:    l.actor,
		Action:   EventUpdate,
		Resource: resource,
		Target:   target,
		Details:  details,
		Severity: SeverityInfo,
		Source:   l.source,
		Result:   "success",
	})
}

func (l *auditLoggerImpl) Delete(resource string, target string) {
	l.auditor.Log(Event{
		Actor:    l.actor,
		Action:   EventDelete,
		Resource: resource,
		Target:   target,
		Severity: SeverityInfo,
		Source:   l.source,
		Result:   "success",
	})
}

func (l *auditLoggerImpl) Login(target string, details map[string]any) {
	l.auditor.Log(Event{
		Actor:    l.actor,
		Action:   EventLogin,
		Severity: SeverityInfo,
		Source:   l.source,
		Target:   target,
		Details:  details,
		Result:   "success",
	})
}

func (l *auditLoggerImpl) Severity(resource string, target string, severity EventSeverity, details map[string]any) {
	l.auditor.Log(Event{
		Actor:    l.actor,
		Action:   EventSecurity,
		Resource: resource,
		Target:   target,
		Severity: severity,
		Source:   l.source,
		Details:  details,
	})
}
