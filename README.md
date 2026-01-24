# simV

A Go library for simulating time-varying values in testing environments.

## Overview

simV provides thread-safe entities representing different types of time-varying values, enabling developers to create realistic simulated metrics and time-series data for system validation.

**Use cases:**

- Testing OTEL/Prometheus metric pipelines
- Validating observability systems
- Simulating realistic load patterns
- Mocking external metric sources during development
- Creating observable test fixtures with temporal behavior

## Installation

```bash
go get github.com/neox5/simv@latest
```

**Requirements:** Go 1.25+

## Quick Start

```go
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
    // Optional: Initialize seed for repeatable simulations
    seed.Init(12345)

    // Create clock that ticks every 100ms
    clk := clock.NewPeriodicClock(100 * time.Millisecond)

    // Create source that generates random integers [1, 10]
    src := source.NewRandomIntSource(clk, 1, 10)

    // Create value with accumulate transform and reset-on-read
    val := value.New(src).
        AddTransform(transform.NewAccumulate[int]()).
        EnableResetOnRead(0).
        SetUpdateHook(value.NewDefaultTraceHook[int]())

    // Start the value (locks configuration)
    val.Start()
    defer val.Stop()

    // Start the clock
    clk.Start()
    defer clk.Stop()

    // Read current value (resets to 0 after read)
    current := val.Value()
    fmt.Printf("Current value: %d\n", current)

    // Access metrics without side effects
    stats := val.Stats()
    fmt.Printf("Updates: %d, Current: %d\n",
        stats.UpdateCount,
        stats.CurrentValue,
    )
}
```

## Architecture

simV uses a pipeline architecture:

**Clock** → **Source** → **Transform** → **Value**

- **Clock**: Generates timing signals at fixed intervals
- **Source**: Produces values on each clock tick
- **Transform**: Modifies values (accumulate, average, etc.)
- **Value**: Manages state and provides thread-safe access

## Core Concepts

### Seed

Control repeatability of random value generation.

```go
// Repeatable simulations - same seed produces identical sequences
seed.Init(12345)

// Non-repeatable (default) - auto-initializes with time-based seed
// No need to call Init if repeatability is not required
```

### Clock

Provides timing signals for value generation.

```go
clk := clock.NewPeriodicClock(100 * time.Millisecond)
clk.Start()
defer clk.Stop()

// Access metrics
stats := clk.Stats()
fmt.Printf("Ticks: %d, Running: %v\n", stats.TickCount, stats.IsRunning)
```

### Source

Generates values driven by clock ticks.

```go
// Constant value
constSrc := source.NewConstSource(clk, 42)

// Random integers
randomSrc := source.NewRandomIntSource(clk, 1, 100)

// Access metrics
stats := randomSrc.Stats()
fmt.Printf("Generated: %d, Subscribers: %d\n",
    stats.GenerationCount,
    stats.SubscriberCount,
)
```

### Transform

Applies operations to incoming values.

```go
// Running total
val.AddTransform(transform.NewAccumulate[int]())
```

### Value

Thread-safe value management with configurable behaviors.

```go
// Create value (configuration phase)
val := value.New(src).
    AddTransform(transform.NewAccumulate[int]()).
    EnableResetOnRead(0).
    SetUpdateHook(value.NewDefaultTraceHook[int]())

// Start value (locks configuration, begins updates)
val.Start()
defer val.Stop()

// Read value
current := val.Value()

// Access metrics without side effects
stats := val.Stats()
fmt.Printf("Updates: %d, Current: %d, Transforms: %d\n",
    stats.UpdateCount,
    stats.CurrentValue,
    stats.TransformCount,
)
```

**Important:** Configuration methods (AddTransform, EnableResetOnRead) panic if called after Start().

### Multiple Values from Same Source

Create independent values that receive the same source stream:

```go
src := source.NewRandomIntSource(clk, 1, 10)

// Value A: Running total (never resets)
valueA := value.New(src).
    AddTransform(transform.NewAccumulate[int]()).
    Start()

// Value B: Delta counter (resets on each read)
valueB := value.New(src).
    AddTransform(transform.NewAccumulate[int]()).
    EnableResetOnRead(0).
    Start()
```

Both values maintain independent state while receiving the same random integers.

## Observability

### Metrics

All components expose metrics via `Stats()` methods for integration with Prometheus/OTEL:

```go
// Clock metrics
clockStats := clk.Stats()
// - TickCount: total ticks generated
// - IsRunning: current operational state
// - Interval: tick rate

// Source metrics
sourceStats := src.Stats()
// - GenerationCount: total values produced
// - SubscriberCount: active subscriptions

// Value metrics
valueStats := val.Stats()
// - UpdateCount: total updates received
// - CurrentValue: current value without side effects
// - TransformCount: number of transforms in chain
```

**Prometheus example:**

```go
tickCounter := prometheus.NewCounter(...)
stats := clk.Stats()
tickCounter.Add(float64(stats.TickCount))
```

### Tracing

Enable trace output to observe value flow through the pipeline:

```go
val.SetUpdateHook(value.NewDefaultTraceHook[int]())
// Output: [15:04:05.000] 7 | Accumulate(s:42) | 49
```

## Features

- Generic type support
- Thread-safe concurrent access
- Explicit lifecycle management (construct → configure → start)
- Atomic reset-on-read (no data loss)
- Subscription-based value distribution
- Composable transform pipeline
- Observable update cycles via hooks
- Metrics exposure for monitoring systems
- Repeatable simulations via seed control
- Zero external dependencies

## Examples

See `cmd/example/main.go` for complete working example.

## License

MIT License - see LICENSE file for details.
