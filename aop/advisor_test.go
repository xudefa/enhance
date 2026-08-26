package aop

import (
	"reflect"
	"testing"
)

func TestNewAdvisor(t *testing.T) {
	t.Parallel()

	advice := Before(func(jp JoinPoint) {})
	pointCut := MatchByName("Do*")
	advisor := NewAdvisor(advice, pointCut, 5)

	if advisor.Advice() != advice {
		t.Error("Advice() returned wrong advice")
	}
	if advisor.PointCut() != pointCut {
		t.Error("PointCut() returned wrong pointcut")
	}
	if advisor.Order() != 5 {
		t.Errorf("Order() = %d, want 5", advisor.Order())
	}
}

func TestNewAdvisor_DifferentOrders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		order int
	}{
		{"zero order", 0},
		{"negative order", -10},
		{"positive order", 100},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			advisor := NewAdvisor(nil, nil, tt.order)
			if advisor.Order() != tt.order {
				t.Errorf("Order() = %d, want %d", advisor.Order(), tt.order)
			}
		})
	}
}

func TestDefaultAdvisor_ReturnsAdvisorInterface(t *testing.T) {
	t.Parallel()

	advisor := NewAdvisor(nil, nil, 0)
	var _ Advisor = advisor
	if reflect.TypeOf(advisor).String() != "*aop.defaultAdvisor" {
		t.Errorf("unexpected type: %v", reflect.TypeOf(advisor))
	}
}
