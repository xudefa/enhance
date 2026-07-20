package metrics

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkSimpleCounter_Inc(b *testing.B) {
	c := NewSimpleCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkSimpleCounter_Add(b *testing.B) {
	c := NewSimpleCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(1.5)
	}
}

func BenchmarkSimpleCounter_Value(b *testing.B) {
	c := NewSimpleCounter()
	c.Add(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Value()
	}
}

func BenchmarkSimpleCounter_ConcurrentInc(b *testing.B) {
	c := NewSimpleCounter()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

func BenchmarkSimpleCounter_ConcurrentAdd(b *testing.B) {
	c := NewSimpleCounter()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1.5)
		}
	})
}

func BenchmarkSimpleGauge_Set(b *testing.B) {
	g := NewSimpleGauge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Set(float64(i))
	}
}

func BenchmarkSimpleGauge_Add(b *testing.B) {
	g := NewSimpleGauge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Add(1.5)
	}
}

func BenchmarkSimpleGauge_Value(b *testing.B) {
	g := NewSimpleGauge()
	g.Set(100.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Value()
	}
}

func BenchmarkSimpleGauge_ConcurrentSet(b *testing.B) {
	g := NewSimpleGauge()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			g.Set(float64(i))
			i++
		}
	})
}

func BenchmarkSimpleGauge_ConcurrentAdd(b *testing.B) {
	g := NewSimpleGauge()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Add(1.5)
		}
	})
}

func BenchmarkSimpleRegistry_Counter(b *testing.B) {
	r := NewSimpleRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Counter(fmt.Sprintf("counter.%d", i%100))
	}
}

func BenchmarkSimpleRegistry_Gauge(b *testing.B) {
	r := NewSimpleRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Gauge(fmt.Sprintf("gauge.%d", i%100))
	}
}

func BenchmarkSimpleRegistry_Histogram(b *testing.B) {
	r := NewSimpleRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Histogram(fmt.Sprintf("histogram.%d", i%100))
	}
}

func BenchmarkSimpleRegistry_CounterWithTags(b *testing.B) {
	r := NewSimpleRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Counter("counter", "host", fmt.Sprintf("host-%d", i%10), "region", fmt.Sprintf("region-%d", i%5))
	}
}

func BenchmarkSimpleRegistry_ConcurrentCounter(b *testing.B) {
	r := NewSimpleRegistry()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r.Counter(fmt.Sprintf("counter.%d", i%100)).Inc()
			i++
		}
	})
}

func BenchmarkSimpleRegistry_ConcurrentGauge(b *testing.B) {
	r := NewSimpleRegistry()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r.Gauge(fmt.Sprintf("gauge.%d", i%100)).Set(float64(i))
			i++
		}
	})
}

func BenchmarkSimpleRegistry_ConcurrentMixed(b *testing.B) {
	r := NewSimpleRegistry()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			op := i % 3
			switch op {
			case 0:
				r.Counter(fmt.Sprintf("counter.%d", i%50)).Inc()
			case 1:
				r.Gauge(fmt.Sprintf("gauge.%d", i%50)).Set(float64(i))
			case 2:
				r.Histogram(fmt.Sprintf("histogram.%d", i%50)).Record(float64(i))
			}
			i++
		}
	})
}

func BenchmarkSimpleRegistry_Collect(b *testing.B) {
	r := NewSimpleRegistry()
	for i := 0; i < 100; i++ {
		r.Counter(fmt.Sprintf("counter.%d", i)).Add(float64(i))
		r.Gauge(fmt.Sprintf("gauge.%d", i)).Set(float64(i))
		r.Histogram(fmt.Sprintf("histogram.%d", i)).Record(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Collect()
	}
}

func TestSimpleRegistry_ConcurrentCounter(t *testing.T) {
	t.Parallel()
	r := NewSimpleRegistry()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				r.Counter(fmt.Sprintf("counter.%d", gid)).Inc()
			}
		}(g)
	}

	wg.Wait()

	metrics := r.Collect()
	if len(metrics) != goroutines {
		t.Errorf("expected %d metrics, got %d", goroutines, len(metrics))
	}

	for _, m := range metrics {
		if m.Type != "counter" {
			t.Errorf("expected counter type, got %s", m.Type)
		}
		if m.Value != 100 {
			t.Errorf("expected counter value 100, got %f", m.Value)
		}
	}
}

func TestSimpleRegistry_ConcurrentGauge(t *testing.T) {
	t.Parallel()
	r := NewSimpleRegistry()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				r.Gauge(fmt.Sprintf("gauge.%d", gid)).Set(float64(i))
			}
		}(g)
	}

	wg.Wait()

	metrics := r.Collect()
	if len(metrics) != goroutines {
		t.Errorf("expected %d metrics, got %d", goroutines, len(metrics))
	}
}

func TestSimpleRegistry_ConcurrentMixed(t *testing.T) {
	t.Parallel()
	r := NewSimpleRegistry()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				op := (gid + i) % 3
				switch op {
				case 0:
					r.Counter(fmt.Sprintf("counter.%d", gid)).Inc()
				case 1:
					r.Gauge(fmt.Sprintf("gauge.%d", gid)).Set(float64(i))
				case 2:
					r.Histogram(fmt.Sprintf("histogram.%d", gid)).Record(float64(i))
				}
			}
		}(g)
	}

	wg.Wait()

	metrics := r.Collect()
	if len(metrics) != goroutines*3 {
		t.Errorf("expected %d metrics, got %d", goroutines*3, len(metrics))
	}
}

func TestSimpleCounter_ConcurrentAdd(t *testing.T) {
	t.Parallel()
	c := NewSimpleCounter()
	var wg sync.WaitGroup
	const goroutines = 100
	const opsPerGoroutine = 1000

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				c.Add(1.5)
			}
		}()
	}

	wg.Wait()

	expected := float64(goroutines * opsPerGoroutine * 1.5)
	if c.Value() != expected {
		t.Errorf("expected %f, got %f", expected, c.Value())
	}
}

func TestSimpleGauge_ConcurrentAdd(t *testing.T) {
	t.Parallel()
	gauge := NewSimpleGauge()
	var wg sync.WaitGroup
	const goroutines = 100
	const opsPerGoroutine = 1000

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				gauge.Add(1.5)
			}
		}(g)
	}

	wg.Wait()

	expected := float64(goroutines * opsPerGoroutine * 1.5)
	if gauge.Value() != expected {
		t.Errorf("expected %f, got %f", expected, gauge.Value())
	}
}
