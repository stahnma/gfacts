package gfacts

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/stahnma/gfacts/external"
	"github.com/stahnma/gfacts/facts"
)

var globalRegistry = &programmaticFacts{
	static: make(map[string]any),
	funcs:  make(map[string]FactFunc),
}

type programmaticFacts struct {
	mu     sync.Mutex
	static map[string]any
	funcs  map[string]FactFunc
}

func (p *programmaticFacts) registerStatic(key string, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.static[key] = value
}

func (p *programmaticFacts) registerFunc(key string, fn FactFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.funcs[key] = fn
}

func (p *programmaticFacts) resolve() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make(map[string]any, len(p.static)+len(p.funcs))
	for k, v := range p.static {
		result[k] = v
	}
	for k, fn := range p.funcs {
		v, err := fn()
		if err != nil {
			slog.Warn("programmatic fact failed", "key", k, "error", err)
			continue
		}
		result[k] = v
	}
	return result
}

type registry struct {
	opts Options
}

func newRegistry(opts Options) *registry {
	return &registry{opts: opts}
}

func (r *registry) resolve() (map[string]any, error) {
	result := make(map[string]any)

	// Layer 1: Core facts (lowest precedence)
	coreFacts := r.collectCore()
	for k, v := range coreFacts {
		result[k] = v
	}

	// Layer 2: External facts
	if !r.opts.NoExternal {
		extFacts := r.collectExternal()
		for k, v := range extFacts {
			result[k] = v
		}
	}

	// Layer 3: Programmatic facts (highest precedence)
	progFacts := globalRegistry.resolve()
	for k, v := range progFacts {
		result[k] = v
	}

	return result, nil
}

func (r *registry) collectCore() map[string]any {
	collectors := facts.All()
	results := make([]map[string]any, len(collectors))

	var wg sync.WaitGroup
	for i, c := range collectors {
		wg.Add(1)
		go func(idx int, col facts.Collector) {
			defer wg.Done()
			m, err := col.Collect()
			if err != nil {
				slog.Debug("collector failed",
					"collector", fmt.Sprintf("%T", col),
					"error", err)
				return
			}
			results[idx] = m
		}(i, c)
	}
	wg.Wait()

	merged := make(map[string]any)
	for _, m := range results {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func (r *registry) collectExternal() map[string]any {
	merged := make(map[string]any)
	for _, dir := range r.opts.ExternalDirs {
		extFacts, err := external.Load(dir, external.LoadOptions{
			NoExec:      r.opts.NoCustomExec,
			ExecTimeout: r.opts.ExecTimeout,
		})
		if err != nil {
			slog.Debug("external fact dir failed", "dir", dir, "error", err)
			continue
		}
		for k, v := range extFacts {
			merged[k] = v
		}
	}
	return merged
}

// lookup finds a value by dotted key. If the key matches a prefix,
// it returns a sub-map of all matching facts.
func lookup(facts map[string]any, key string) any {
	// Exact match first.
	if v, ok := facts[key]; ok {
		return v
	}

	// Prefix match: collect all keys under this prefix.
	prefix := key + "."
	sub := make(map[string]any)
	for k, v := range facts {
		if strings.HasPrefix(k, prefix) {
			sub[k] = v
		}
	}
	if len(sub) > 0 {
		return sub
	}
	return nil
}
