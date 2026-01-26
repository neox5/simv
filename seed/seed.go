package seed

import (
	"math/rand/v2"
	"sync"
)

var (
	globalRegistry *registry
	registryOnce   sync.Once
)

// registry provides deterministic seed sequences.
type registry struct {
	mu         sync.Mutex
	masterSeed uint64
	nextStream uint64
}

// Init initializes the global seed registry with a master seed.
// MUST be called before any simv sources are created.
// Panics if called multiple times.
//
// For deterministic simulations, provide an explicit seed:
//
//	seed.Init(12345)
//
// For non-repeatable behavior, use a time-based seed:
//
//	seed.Init(uint64(time.Now().UnixNano()))
func Init(masterSeed uint64) {
	initialized := false
	registryOnce.Do(func() {
		globalRegistry = &registry{
			masterSeed: masterSeed,
			nextStream: 0,
		}
		initialized = true
	})

	if !initialized {
		panic("seed.Init called multiple times")
	}
}

// NewRand returns a new independent random number generator.
// Each call returns an RNG with seeds (masterSeed, streamN) where N increments.
// Panics if Init() was not called.
func NewRand() *rand.Rand {
	if globalRegistry == nil {
		panic("seed.NewRand called before seed.Init - call seed.Init() at program start")
	}
	return globalRegistry.newRand()
}

// Current returns the active seed state for logging and reproducibility.
// Returns (masterSeed, streamCounter) where:
// - masterSeed: The seed value provided to Init()
// - streamCounter: Current stream counter (number of NewRand() calls made)
//
// Panics if Init() was not called.
//
// For reproducibility, call Init(masterSeed) before creating any sources.
// Each source will receive RNGs with seeds (masterSeed, 0), (masterSeed, 1), etc.
func Current() (masterSeed, streamCounter uint64) {
	if globalRegistry == nil {
		panic("seed.Current called before seed.Init - call seed.Init() at program start")
	}

	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	return globalRegistry.masterSeed, globalRegistry.nextStream
}

func (r *registry) newRand() *rand.Rand {
	r.mu.Lock()
	defer r.mu.Unlock()

	seed1 := r.masterSeed
	seed2 := r.nextStream
	r.nextStream++

	return rand.New(rand.NewPCG(seed1, seed2))
}
