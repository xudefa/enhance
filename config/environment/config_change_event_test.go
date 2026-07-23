package environment

import (
	"testing"
	"time"
)

func TestConfigChangeEvent_BasicFields(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent(
		"modify",
		[]string{"server.port", "server.host"},
		map[string]any{"server.port": "8080"},
		map[string]any{"server.port": "9090"},
		"nacos",
	)

	if event.EventType != "modify" {
		t.Errorf("EventType = %v, want modify", event.EventType)
	}
	if len(event.Keys) != 2 {
		t.Errorf("Keys length = %v, want 2", len(event.Keys))
	}
	if event.Keys[0] != "server.port" {
		t.Errorf("Keys[0] = %v, want server.port", event.Keys[0])
	}
	if event.Source != "nacos" {
		t.Errorf("Source = %v, want nacos", event.Source)
	}
}

func TestConfigChangeEvent_Type(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent("modify", nil, nil, nil, "test")

	if event.Type() != "ConfigChange" {
		t.Errorf("Type() = %v, want ConfigChange", event.Type())
	}
}

func TestConfigChangeEvent_Timestamp(t *testing.T) {
	t.Parallel()

	before := time.Now()
	event := NewConfigChangeEvent("modify", nil, nil, nil, "test")
	after := time.Now()

	ts := event.Timestamp()
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp() = %v, want between %v and %v", ts, before, after)
	}
}

func TestConfigChangeEvent_Values(t *testing.T) {
	t.Parallel()

	oldVals := map[string]any{"a": 1}
	newVals := map[string]any{"a": 2}
	event := NewConfigChangeEvent("modify", []string{"a"}, oldVals, newVals, "test")

	if event.OldValues["a"] != 1 {
		t.Errorf("OldValues[a] = %v, want 1", event.OldValues["a"])
	}
	if event.NewValues["a"] != 2 {
		t.Errorf("NewValues[a] = %v, want 2", event.NewValues["a"])
	}
}

func TestConfigChangeEvent_Metadata(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent("modify", nil, nil, nil, "test")

	if event.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}
	if len(event.Metadata) != 0 {
		t.Errorf("Metadata length = %v, want 0", len(event.Metadata))
	}
}
