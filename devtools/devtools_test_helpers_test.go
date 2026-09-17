package devtools

import (
	"sync"
	"testing"
	"time"
)

func testHotReloaderDetectEvent(t *testing.T, wg *sync.WaitGroup, events *[]ReloadEvent, mu *sync.Mutex, wantType ReloadType, wantFile, timeoutMsg string) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功检测到变化
	case <-time.After(2 * time.Second):
		t.Fatal(timeoutMsg)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(*events) == 0 {
		t.Fatal("expected at least one reload event")
	}

	event := (*events)[0]
	if event.Type != wantType {
		t.Errorf("expected event type %v, got %s", wantType, event.Type)
	}

	if wantFile != "" && event.File != wantFile {
		t.Errorf("expected file %s, got %s", wantFile, event.File)
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
