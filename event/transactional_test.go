package event

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestTransactionalEvent_Wrapping(t *testing.T) {
	t.Parallel()
	inner := &BaseEvent{EventType: "test.event"}
	te := NewTransactionalEvent(inner, PhaseAfterCommit)

	if te.Type() != "test.event" {
		t.Errorf("expected type 'test.event', got %s", te.Type())
	}

	if te.Phase() != PhaseAfterCommit {
		t.Errorf("expected phase after_commit, got %s", te.Phase())
	}

	if te.Event() != inner {
		t.Error("expected Event() to return inner event")
	}
}

func TestTransactionContext_Commit(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var afterCommitCount int32
	var beforeCommitCount int32

	bus.Subscribe("before.event", func(e ApplicationEvent) {
		atomic.AddInt32(&beforeCommitCount, 1)
	})
	bus.Subscribe("after.event", func(e ApplicationEvent) {
		atomic.AddInt32(&afterCommitCount, 1)
	})

	tc.PublishBeforeCommit(&BaseEvent{EventType: "before.event"})
	tc.PublishAfterCommit(&BaseEvent{EventType: "after.event"})

	// 提交前不应该有事件被发布
	if atomic.LoadInt32(&beforeCommitCount) != 0 {
		t.Error("before commit event should not be published before commit")
	}
	if atomic.LoadInt32(&afterCommitCount) != 0 {
		t.Error("after commit event should not be published before commit")
	}

	tc.Commit(bus)

	if atomic.LoadInt32(&beforeCommitCount) != 1 {
		t.Error("before commit event should be published on commit")
	}
	if atomic.LoadInt32(&afterCommitCount) != 1 {
		t.Error("after commit event should be published on commit")
	}

	if !tc.IsCommitted() {
		t.Error("transaction should be marked as committed")
	}
}

func TestTransactionContext_Rollback(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var afterRollbackCount int32
	var afterCommitCount int32

	bus.Subscribe("rollback.event", func(e ApplicationEvent) {
		atomic.AddInt32(&afterRollbackCount, 1)
	})
	bus.Subscribe("commit.event", func(e ApplicationEvent) {
		atomic.AddInt32(&afterCommitCount, 1)
	})

	tc.PublishAfterRollback(&BaseEvent{EventType: "rollback.event"})
	tc.PublishAfterCommit(&BaseEvent{EventType: "commit.event"})

	tc.Rollback(bus)

	if atomic.LoadInt32(&afterRollbackCount) != 1 {
		t.Error("after rollback event should be published on rollback")
	}
	if atomic.LoadInt32(&afterCommitCount) != 0 {
		t.Error("after commit event should NOT be published on rollback")
	}

	if !tc.IsRolledBack() {
		t.Error("transaction should be marked as rolled back")
	}
}

func TestTransactionContext_DoubleCommit(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var count int32
	bus.Subscribe("test.event", func(e ApplicationEvent) {
		atomic.AddInt32(&count, 1)
	})

	tc.PublishAfterCommit(&BaseEvent{EventType: "test.event"})

	tc.Commit(bus)
	tc.Commit(bus) // 第二次提交应该被忽略

	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected event published only once, got %d", count)
	}
}

func TestTransactionContext_CommitThenRollback(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var commitCount int32
	var rollbackCount int32

	bus.Subscribe("commit.event", func(e ApplicationEvent) {
		atomic.AddInt32(&commitCount, 1)
	})
	bus.Subscribe("rollback.event", func(e ApplicationEvent) {
		atomic.AddInt32(&rollbackCount, 1)
	})

	tc.PublishAfterCommit(&BaseEvent{EventType: "commit.event"})
	tc.PublishAfterRollback(&BaseEvent{EventType: "rollback.event"})

	tc.Commit(bus)
	tc.Rollback(bus) // 提交后再回滚应该被忽略

	if atomic.LoadInt32(&commitCount) != 1 {
		t.Error("commit event should be published")
	}
	if atomic.LoadInt32(&rollbackCount) != 0 {
		t.Error("rollback event should NOT be published after commit")
	}
}

func TestTransactionContext_RollbackThenCommit(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var commitCount int32
	var rollbackCount int32

	bus.Subscribe("commit.event", func(e ApplicationEvent) {
		atomic.AddInt32(&commitCount, 1)
	})
	bus.Subscribe("rollback.event", func(e ApplicationEvent) {
		atomic.AddInt32(&rollbackCount, 1)
	})

	tc.PublishAfterCommit(&BaseEvent{EventType: "commit.event"})
	tc.PublishAfterRollback(&BaseEvent{EventType: "rollback.event"})

	tc.Rollback(bus)
	tc.Commit(bus) // 回滚后再提交应该被忽略

	if atomic.LoadInt32(&rollbackCount) != 1 {
		t.Error("rollback event should be published")
	}
	if atomic.LoadInt32(&commitCount) != 0 {
		t.Error("commit event should NOT be published after rollback")
	}
}

func TestTransactionalEventPublisher(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewTransactionalEventPublisher(bus)

	var afterCommitCount int32
	var afterRollbackCount int32

	bus.Subscribe("commit.event", func(e ApplicationEvent) {
		atomic.AddInt32(&afterCommitCount, 1)
	})
	bus.Subscribe("rollback.event", func(e ApplicationEvent) {
		atomic.AddInt32(&afterRollbackCount, 1)
	})

	tx := publisher.BeginTransaction()
	tx.PublishAfterCommit(&BaseEvent{EventType: "commit.event"})
	tx.PublishAfterRollback(&BaseEvent{EventType: "rollback.event"})

	tx.Commit(bus)

	if atomic.LoadInt32(&afterCommitCount) != 1 {
		t.Error("after commit event should be published")
	}
	if atomic.LoadInt32(&afterRollbackCount) != 0 {
		t.Error("after rollback event should NOT be published on commit")
	}
}

func TestTransactionContext_ConcurrentRegistration(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tc.PublishAfterCommit(&BaseEvent{EventType: "concurrent.event"})
		}(i)
	}
	wg.Wait()

	var count int32
	bus.Subscribe("concurrent.event", func(e ApplicationEvent) {
		atomic.AddInt32(&count, 1)
	})

	tc.Commit(bus)

	if atomic.LoadInt32(&count) != 10 {
		t.Errorf("expected 10 events, got %d", count)
	}
}

func TestTransactionContext_DefaultPhase(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	tc := NewTransactionContext()

	var count int32
	bus.Subscribe("default.event", func(e ApplicationEvent) {
		atomic.AddInt32(&count, 1)
	})

	// 直接注册非 TransactionalEvent，应该默认为 AfterCommit
	tc.RegisterEvent(&BaseEvent{EventType: "default.event"})

	tc.Commit(bus)

	if atomic.LoadInt32(&count) != 1 {
		t.Error("non-transactional event should default to after_commit phase")
	}
}
