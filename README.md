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
go get github.com/neox5/simv@v0.1.0
```

**Requirements:** Go 1.25+

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/neox5/simv/clock"
    "github.com/neox5/simv/source"
    "github.com/neox5/simv/transform"
    "github.com/neox5/simv/value"
)

func main() {
    // Create clock that ticks every 100ms
    clk := clock.NewPeriodicClock(100 * time.Millisecond)

    // Create source that generates random integers [1, 10]
    src := source.NewRandomIntSource(clk, 1, 10)

    // Create value that accumulates incoming values
    val := value.New(src, transform.NewAccumulate[int]())

    // Wrap with reset-on-read behavior
    resetVal := value.NewResetOnRead(val.Clone(), 0)

    // Enable trace output
    resetVal.SetUpdateHook(value.NewDefaultTraceHook[int]())

    // Start the clock
    clk.Start()
    defer clk.Stop()

    // Read current value
    current := resetVal.Value()
    fmt.Printf("Current value: %d\n", current)
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

### Clock

Provides timing signals for value generation.

```go
clk := clock.NewPeriodicClock(100 * time.Millisecond)
clk.Start()
defer clk.Stop()
```

### Source

Generates values driven by clock ticks.

```go
// Constant value
constSrc := source.NewConstSource(clk, 42)

// Random integers
randomSrc := source.NewRandomIntSource(clk, 1, 100)
```

### Transform

Applies operations to incoming values.

```go
// Running total
accumulated := value.New(src, transform.NewAccumulate[int]())
```

### Value

Thread-safe value management with optional behaviors.

```go
// Standard value
val := value.New(src)

// Reset on each read
resetVal := value.NewResetOnRead(val.Clone(), 0)

// Enable tracing
val.SetUpdateHook(value.NewDefaultTraceHook[int]())
```

## Features

- Generic type support
- Thread-safe concurrent access
- Subscription-based value distribution
- Composable transform pipeline
- Observable update cycles via hooks
- Zero external dependencies

## Examples

See `cmd/example/main.go` for complete working example.

## License

MIT License - see LICENSE file for details.
