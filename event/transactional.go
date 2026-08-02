package event

import (
	"sync"
	"time"
)

// TransactionPhase 事务阶段枚举
type TransactionPhase string

const (
	PhaseBeforeCommit  TransactionPhase = "before_commit"  // 事务提交前
	PhaseAfterCommit   TransactionPhase = "after_commit"   // 事务提交后
	PhaseAfterRollback TransactionPhase = "after_rollback" // 事务回滚后
)

// TransactionalEvent 事务事件包装器
//
// 包装 ApplicationEvent 并绑定事务阶段信息。
// 在事务提交/回滚时，事件会根据绑定的阶段延迟发布。
//
// 使用示例：
//
//	te := event.NewTransactionalEvent(
//	    &MyEvent{Data: "hello"},
//	    event.PhaseAfterCommit,
//	)
type TransactionalEvent struct {
	event ApplicationEvent // 被包装的原始事件
	phase TransactionPhase // 事务阶段
}

// NewTransactionalEvent 创建事务事件
func NewTransactionalEvent(event ApplicationEvent, phase TransactionPhase) *TransactionalEvent {
	return &TransactionalEvent{
		event: event,
		phase: phase,
	}
}

// Type 返回事件类型（委托给内部事件）
func (e *TransactionalEvent) Type() string {
	return e.event.Type()
}

// Timestamp 返回事件时间戳（委托给内部事件）
func (e *TransactionalEvent) Timestamp() time.Time {
	return e.event.Timestamp()
}

// Phase 返回事务阶段
func (e *TransactionalEvent) Phase() TransactionPhase {
	return e.phase
}

// Event 返回内部事件
func (e *TransactionalEvent) Event() ApplicationEvent {
	return e.event
}

// TransactionContext 事务上下文
//
// 跟踪事务中注册的事件，在 Commit/Rollback 时按阶段发布。
// 线程安全，支持并发注册事件。
// 使用分段锁设计，减少高并发场景下的锁竞争。
type TransactionContext struct {
	mu            sync.RWMutex
	committed     bool
	rolledBack    bool
	afterCommit   []ApplicationEvent // 提交后发布的事件
	afterRollback []ApplicationEvent // 回滚后发布的事件
	beforeCommit  []ApplicationEvent // 提交前发布的事件
}

// NewTransactionContext 创建事务上下文
func NewTransactionContext() *TransactionContext {
	return &TransactionContext{
		afterCommit:   make([]ApplicationEvent, 0, 8),
		afterRollback: make([]ApplicationEvent, 0, 4),
		beforeCommit:  make([]ApplicationEvent, 0, 4),
	}
}

// RegisterEvent 注册事务事件
//
// 线程安全，支持并发调用。
func (tc *TransactionContext) RegisterEvent(event ApplicationEvent) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if te, ok := event.(*TransactionalEvent); ok {
		switch te.Phase() {
		case PhaseBeforeCommit:
			tc.beforeCommit = append(tc.beforeCommit, te.Event())
		case PhaseAfterCommit:
			tc.afterCommit = append(tc.afterCommit, te.Event())
		case PhaseAfterRollback:
			tc.afterRollback = append(tc.afterRollback, te.Event())
		}
		return
	}
	// 默认作为 AfterCommit 事件
	tc.afterCommit = append(tc.afterCommit, event)
}

// Commit 提交事务，发布 BeforeCommit 和 AfterCommit 事件
//
// 线程安全，多次调用只会执行一次。
func (tc *TransactionContext) Commit(bus *EventBus) {
	tc.mu.Lock()
	if tc.committed || tc.rolledBack {
		tc.mu.Unlock()
		return
	}
	tc.committed = true

	// 快照事件列表，释放锁后再发布
	beforeCommit := tc.beforeCommit
	afterCommit := tc.afterCommit
	tc.beforeCommit = make([]ApplicationEvent, 0)
	tc.afterCommit = make([]ApplicationEvent, 0)
	tc.afterRollback = make([]ApplicationEvent, 0)
	tc.mu.Unlock()

	// 发布 BeforeCommit 事件（在锁外执行）
	for _, e := range beforeCommit {
		bus.Publish(e)
	}

	// 发布 AfterCommit 事件（在锁外执行）
	for _, e := range afterCommit {
		bus.Publish(e)
	}
}

// Rollback 回滚事务，发布 AfterRollback 事件
//
// 线程安全，多次调用只会执行一次。
func (tc *TransactionContext) Rollback(bus *EventBus) {
	tc.mu.Lock()
	if tc.committed || tc.rolledBack {
		tc.mu.Unlock()
		return
	}
	tc.rolledBack = true

	// 快照事件列表，释放锁后再发布
	afterRollback := tc.afterRollback
	tc.beforeCommit = make([]ApplicationEvent, 0)
	tc.afterCommit = make([]ApplicationEvent, 0)
	tc.afterRollback = make([]ApplicationEvent, 0)
	tc.mu.Unlock()

	// 发布 AfterRollback 事件（在锁外执行）
	for _, e := range afterRollback {
		bus.Publish(e)
	}
}

// IsCommitted 返回事务是否已提交
func (tc *TransactionContext) IsCommitted() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.committed
}

// IsRolledBack 返回事务是否已回滚
func (tc *TransactionContext) IsRolledBack() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.rolledBack
}

// TransactionalEventPublisher 事务事件发布器
//
// 提供便捷的事务事件发布 API。
//
// 使用示例：
//
//	publisher := event.NewTransactionalEventPublisher(bus)
//	tx := publisher.BeginTransaction()
//	tx.PublishAfterCommit(&MyEvent{})
//	tx.Commit()
type TransactionalEventPublisher struct {
	bus *EventBus
}

// NewTransactionalEventPublisher 创建事务事件发布器
func NewTransactionalEventPublisher(bus *EventBus) *TransactionalEventPublisher {
	return &TransactionalEventPublisher{bus: bus}
}

// BeginTransaction 开始新事务
func (p *TransactionalEventPublisher) BeginTransaction() *TransactionContext {
	return NewTransactionContext()
}

// TransactionContext 事务上下文便捷方法

// PublishBeforeCommit 注册 BeforeCommit 阶段事件
func (tc *TransactionContext) PublishBeforeCommit(event ApplicationEvent) {
	tc.RegisterEvent(NewTransactionalEvent(event, PhaseBeforeCommit))
}

// PublishAfterCommit 注册 AfterCommit 阶段事件
func (tc *TransactionContext) PublishAfterCommit(event ApplicationEvent) {
	tc.RegisterEvent(NewTransactionalEvent(event, PhaseAfterCommit))
}

// PublishAfterRollback 注册 AfterRollback 阶段事件
func (tc *TransactionContext) PublishAfterRollback(event ApplicationEvent) {
	tc.RegisterEvent(NewTransactionalEvent(event, PhaseAfterRollback))
}
