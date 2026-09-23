package devtools

import (
	"sync"
	"testing"
	"time"
)

// hotReloaderDetectEventArgs 热重载检测事件的测试参数。
type hotReloaderDetectEventArgs struct {
	t          *testing.T
	wg         *sync.WaitGroup
	events     *[]ReloadEvent
	mu         *sync.Mutex
	wantType   ReloadType
	wantFile   string
	timeoutMsg string
}

func testHotReloaderDetectEvent(args hotReloaderDetectEventArgs) {
	done := make(chan struct{})
	go func() {
		args.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功检测到变化
	case <-time.After(2 * time.Second):
		args.t.Fatal(args.timeoutMsg)
	}

	args.mu.Lock()
	defer args.mu.Unlock()
	if len(*args.events) == 0 {
		args.t.Fatal("expected at least one reload event")
	}

	event := (*args.events)[0]
	if event.Type != args.wantType {
		args.t.Errorf("expected event type %v, got %s", args.wantType, event.Type)
	}

	if args.wantFile != "" && event.File != args.wantFile {
		args.t.Errorf("expected file %s, got %s", args.wantFile, event.File)
	}
}

func testStopWaitsForCallbacks(t *testing.T, done, stopDone <-chan struct{}, release chan struct{}) {
	// Stop 必须等待正在执行的回调完成，而不是立即返回
	select {
	case <-stopDone:
		t.Fatal("Stop returned while callbacks were still running")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Callback did not complete after release")
	}

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return after callbacks completed")
	}
}

func testHotReloaderExpectCallbacks(t *testing.T, mu *sync.Mutex, callback1Count, callback2Count *int) {
	mu.Lock()
	defer mu.Unlock()
	if *callback1Count == 0 {
		t.Error("expected callback1 to be called")
	}

	if *callback2Count == 0 {
		t.Error("expected callback2 to be called")
	}
}
