package value_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neox5/simv/clock"
	"github.com/neox5/simv/source"
	"github.com/neox5/simv/transform"
	"github.com/neox5/simv/value"
)

// ============================================================================
// BENCHMARK TESTS
// Measure performance characteristics
// ============================================================================

// BenchmarkResetOnRead_SingleReader measures single-threaded read performance.
func BenchmarkResetOnRead_SingleReader(b *testing.B) {
	clk := clock.NewPeriodicClock(1 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	val := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer val.Stop()

	clk.Start()
	defer clk.Stop()

	b.ResetTimer()

	for b.Loop() {
		_ = val.Value()
	}
}

// BenchmarkResetOnRead_ConcurrentReads measures concurrent read performance.
func BenchmarkResetOnRead_ConcurrentReads(b *testing.B) {
	clk := clock.NewPeriodicClock(1 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	val := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer val.Stop()

	clk.Start()
	defer clk.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = val.Value()
		}
	})
}

// BenchmarkResetOnRead_Stats measures Stats() performance.
func BenchmarkResetOnRead_Stats(b *testing.B) {
	clk := clock.NewPeriodicClock(1 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	val := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer val.Stop()

	clk.Start()
	defer clk.Stop()

	b.ResetTimer()

	for b.Loop() {
		_ = val.Stats()
	}
}

// BenchmarkValue_WithoutReset measures baseline performance without reset-on-read.
func BenchmarkValue_WithoutReset(b *testing.B) {
	clk := clock.NewPeriodicClock(1 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	val := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		Start()
	defer val.Stop()

	clk.Start()
	defer clk.Stop()

	b.ResetTimer()

	for b.Loop() {
		_ = val.Value()
	}
}

// ============================================================================
// STRESS TESTS
// Extreme scenarios to expose race conditions and verify robustness
// Run with: go test -run=Stress ./value
// Skip with: go test -short ./value
// ============================================================================

// TestResetOnRead_Stress_RaceCondition targets the specific race from the bug report.
//
// Original bug scenario (from report):
// - Clock: 200ms, Export: 100ms
// - Race: Export reads between source update and reset
// - Symptom: Occasional missing datapoints (~every 2 minutes)
//
// This test maximizes race exposure:
// - Fast clock (10ms) → Many source updates
// - Slow export (100ms) → Large window between reads
// - Single reader → Isolates the race (no reader-vs-reader conflicts)
// - Duration: 30 seconds → ~3000 clock ticks, 300 export reads
//
// Mathematical validation: Sum of reset reads MUST equal accumulated total.
// Any discrepancy indicates the race condition exists.
func TestResetOnRead_Stress_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	clk := clock.NewPeriodicClock(10 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	accumulated := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		Start()
	defer accumulated.Stop()

	resetOnRead := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer resetOnRead.Stop()

	clk.Start()
	defer clk.Stop()

	// Single reader - isolates the race condition
	var exportSum atomic.Int64
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(30 * time.Second)

		for {
			select {
			case <-ticker.C:
				readValue := resetOnRead.Value()
				exportSum.Add(int64(readValue))
			case <-timeout:
				close(done)
				return
			}
		}
	}()

	<-done
	time.Sleep(50 * time.Millisecond)

	finalAccumulated := accumulated.Value()
	finalExportSum := int(exportSum.Load())

	if finalExportSum != finalAccumulated {
		t.Errorf("RACE CONDITION DETECTED: export sum = %d, accumulated = %d, lost = %d",
			finalExportSum, finalAccumulated, finalAccumulated-finalExportSum)
	}

	if finalAccumulated == 0 {
		t.Error("Accumulated value is zero - clock/source not working")
	}
}

// TestResetOnRead_Stress_HighFrequencyConcurrent is the worst-case scenario.
// Very fast clock (5ms), very fast reads (2ms), multiple concurrent readers.
// If there's a race condition, this will find it.
func TestResetOnRead_Stress_HighFrequencyConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	clk := clock.NewPeriodicClock(5 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	accumulated := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		Start()
	defer accumulated.Stop()

	resetOnRead := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer resetOnRead.Stop()

	clk.Start()
	defer clk.Stop()

	// 5 concurrent readers, each reading 500 times
	var exportSum atomic.Int64
	var wg sync.WaitGroup

	for range 5 {
		wg.Go(func() {
			ticker := time.NewTicker(2 * time.Millisecond) // Faster than clock
			defer ticker.Stop()

			for range 500 {
				<-ticker.C
				readValue := resetOnRead.Value()
				exportSum.Add(int64(readValue))
			}
		})
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	finalAccumulated := accumulated.Value()
	finalExportSum := int(exportSum.Load())

	if finalExportSum != finalAccumulated {
		t.Errorf("DATA LOSS under extreme stress: export sum = %d, accumulated = %d, lost = %d",
			finalExportSum, finalAccumulated, finalAccumulated-finalExportSum)
	}

	if finalAccumulated == 0 {
		t.Error("Accumulated value is zero - clock/source not working")
	}
}

// TestResetOnRead_Stress_BurstPattern simulates real-world bursty load.
// 10 bursts of 20 concurrent readers, followed by quiet periods.
// This mirrors actual production patterns (periodic exports, batch processing).
func TestResetOnRead_Stress_BurstPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	clk := clock.NewPeriodicClock(20 * time.Millisecond)
	src := source.NewConstSource(clk, 1)

	accumulated := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		Start()
	defer accumulated.Stop()

	resetOnRead := value.New(src).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		Start()
	defer resetOnRead.Stop()

	clk.Start()
	defer clk.Stop()

	var exportSum atomic.Int64
	var wg sync.WaitGroup

	// 10 bursts with quiet periods
	for range 10 {
		// Burst: 20 concurrent readers
		for range 20 {
			wg.Go(func() {
				readValue := resetOnRead.Value()
				exportSum.Add(int64(readValue))
			})
		}
		wg.Wait()

		// Quiet period (allows accumulation)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	finalAccumulated := accumulated.Value()
	finalExportSum := int(exportSum.Load())

	if finalExportSum != finalAccumulated {
		t.Errorf("DATA LOSS during burst pattern: export sum = %d, accumulated = %d, lost = %d",
			finalExportSum, finalAccumulated, finalAccumulated-finalExportSum)
	}

	if finalAccumulated == 0 {
		t.Error("Accumulated value is zero - clock/source not working")
	}
}
