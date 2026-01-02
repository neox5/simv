package main

import (
	"fmt"
	"time"

	"github.com/neox5/synthetic/internal/counter"
)

func main() {
	// Counter increments every 100ms
	c := counter.New(100 * time.Millisecond)
	defer c.Stop()

	// Read and print every 500ms (different interval!)
	for range 10 {
		fmt.Printf("Counter value: %d\n", c.Value())
		time.Sleep(500 * time.Millisecond)
	}
}
