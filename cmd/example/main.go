package main

import (
	"fmt"
	"time"

	"github.com/neox5/simval/internal/clock"
	"github.com/neox5/simval/internal/source"
	"github.com/neox5/simval/internal/transform"
	"github.com/neox5/simval/internal/value"
)

func main() {
	// Create clock
	clk := clock.NewPeriodicClock(100 * time.Millisecond)

	// Create random source
	randomSrc := source.NewRandomIntSource(clk, 1, 10)

	// Create two values using the same random source:
	// 1. Direct random value (no transform)
	randomValue := value.New(randomSrc)

	// 2. Accumulated random value
	accumulatedValue := value.New(
		randomSrc,
		transform.NewAccumulateTransform[int](),
	)

	// Start clock
	clk.Start()
	defer clk.Stop()

	// Read and print every 500ms
	for range 10 {
		fmt.Printf("Random: %d, Accumulated: %d\n",
			randomValue.Value(),
			accumulatedValue.Value(),
		)
		time.Sleep(500 * time.Millisecond)
	}
}
