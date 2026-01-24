package main

import (
	"fmt"
	"time"

	"github.com/neox5/simv/clock"
	"github.com/neox5/simv/seed"
	"github.com/neox5/simv/source"
	"github.com/neox5/simv/transform"
	"github.com/neox5/simv/value"
)

func main() {
	// Initialize seed for repeatable simulations
	// Comment out this line for non-repeatable (time-based) behavior
	seed.Init(12345)

	// Create clock
	clk := clock.NewPeriodicClock(100 * time.Millisecond)

	// Create random source
	randomSrc := source.NewRandomIntSource(clk, 1, 10)

	// Create accumulated value
	accumulated := value.New(randomSrc).
		AddTransform(transform.NewAccumulate[int]())

	// Create reset-on-read value (same source, independent state)
	resetOnRead := value.New(randomSrc).
		AddTransform(transform.NewAccumulate[int]()).
		EnableResetOnRead(0).
		SetUpdateHook(value.NewDefaultTraceHook[int]())

	// Start values (configuration locked after this)
	accumulated.Start()
	resetOnRead.Start()
	defer accumulated.Stop()
	defer resetOnRead.Stop()

	// Start clock
	clk.Start()
	defer clk.Stop()

	// Read and print every 500ms
	for range 10 {
		fmt.Printf(">>> ResetOnRead Value: %d\n",
			resetOnRead.Value(),
		)

		time.Sleep(500 * time.Millisecond)
	}

	// Print stats after execution
	fmt.Println("\n=== Final Stats ===")

	clockStats := clk.Stats()
	fmt.Printf("Clock: ticks=%d running=%v interval=%v\n",
		clockStats.TickCount,
		clockStats.IsRunning,
		clockStats.Interval,
	)

	sourceStats := randomSrc.Stats()
	fmt.Printf("Source: generations=%d subscribers=%d\n",
		sourceStats.GenerationCount,
		sourceStats.SubscriberCount,
	)

	accumulatedStats := accumulated.Stats()
	fmt.Printf("Accumulated: updates=%d current=%d transforms=%d\n",
		accumulatedStats.UpdateCount,
		accumulatedStats.CurrentValue,
		accumulatedStats.TransformCount,
	)

	resetStats := resetOnRead.Stats()
	fmt.Printf("ResetOnRead: updates=%d current=%d transforms=%d\n",
		resetStats.UpdateCount,
		resetStats.CurrentValue,
		resetStats.TransformCount,
	)
}
