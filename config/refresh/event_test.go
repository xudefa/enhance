package refresh

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestNewConfigChangeEvent(t *testing.T) {
	t.Parallel()
	keys := []string{"app.name", "server.port"}
	oldVals := map[string]any{"app.name": "old"}
	newVals := map[string]any{"app.name": "new"}
	evt := NewConfigChangeEvent(
		"modify",
		environment.WithEventKeys(keys),
		environment.WithEventValues(oldVals, newVals),
		environment.WithEventSource("file"),
	)

	if evt.EventType != "modify" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "modify")
	}
	if len(evt.Keys) != 2 {
		t.Errorf("Keys len = %d, want 2", len(evt.Keys))
	}
	if evt.Source != "file" {
		t.Errorf("Source = %q, want %q", evt.Source, "file")
	}
}
