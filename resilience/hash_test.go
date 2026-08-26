package resilience

import (
	"testing"
)

func TestNewConsistentHash(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	if ch == nil {
		t.Fatal("expected non-nil ConsistentHash")
	}
	if ch.replicas != 150 {
		t.Errorf("expected 150 replicas, got %d", ch.replicas)
	}
}

func TestNewConsistentHash_CustomReplicas(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash(200)
	if ch.replicas != 200 {
		t.Errorf("expected 200 replicas, got %d", ch.replicas)
	}
}

func TestConsistentHash_AddNode(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	ch.AddNode("node1", 1)

	node := ch.GetNode("test-key")
	if node == "" {
		t.Error("expected non-empty node ID")
	}
}

func TestConsistentHash_AddNode_ZeroWeight(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	ch.AddNode("node1", 0)

	node := ch.GetNode("test-key")
	if node == "" {
		t.Error("expected non-empty node ID")
	}
}

func TestConsistentHash_RemoveNode(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	ch.AddNode("node1", 1)
	ch.AddNode("node2", 1)

	ch.RemoveNode("node1")
	node := ch.GetNode("test-key")
	if node != "node2" {
		t.Errorf("expected node2, got %s", node)
	}
}

func TestConsistentHash_GetNode_EmptyRing(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	node := ch.GetNode("test-key")
	if node != "" {
		t.Errorf("expected empty node ID, got %s", node)
	}
}

func TestConsistentHash_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	_, err := ch.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestConsistentHash_Next(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result, err := ch.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestConsistentHash_NextByKey(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result, err := ch.NextByKey(backends, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestConsistentHash_NextByKey_EmptyBackends(t *testing.T) {
	t.Parallel()
	ch := NewConsistentHash()
	_, err := ch.NextByKey(nil, "key")
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestHashKey(t *testing.T) {
	t.Parallel()
	hash1 := hashKey("test-key")
	hash2 := hashKey("test-key")
	if hash1 != hash2 {
		t.Error("expected same hash for same key")
	}

	hash3 := hashKey("different-key")
	if hash1 == hash3 {
		t.Error("expected different hash for different key")
	}
}

func TestNewIPHash(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	if ih == nil {
		t.Fatal("expected non-nil IPHash")
	}
}

func TestIPHash_Next(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := ih.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestIPHash_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	_, err := ih.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestIPHash_NextByIP(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend3", ID: "3"},
	}

	result, err := ih.NextByIP(backends, "192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestIPHash_NextByIP_EmptyIP(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := ih.NextByIP(backends, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestIPHash_NextByIP_EmptyBackends(t *testing.T) {
	t.Parallel()
	ih := NewIPHash()
	_, err := ih.NextByIP(nil, "192.168.1.1")
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}
