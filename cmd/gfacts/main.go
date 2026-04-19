package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/stahnma/gfacts"
)

var version = "dev"

func main() {
	plain := flag.Bool("plain", false, "Output in key=value format")
	debug := flag.Bool("debug", false, "Enable debug logging")
	noExternal := flag.Bool("no-external", false, "Skip external facts")
	externalDir := flag.String("external-dir", "", "Additional external fact directory")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	opts := gfacts.DefaultOptions()
	opts.NoExternal = *noExternal
	if *externalDir != "" {
		opts.ExternalDirs = append(opts.ExternalDirs, *externalDir)
	}

	facts, err := gfacts.GatherWithOptions(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	queries := flag.Args()

	if len(queries) == 0 {
		outputAll(facts, *plain)
		return
	}

	// Single exact key -> bare value
	if len(queries) == 1 {
		val := lookup(facts, queries[0])
		if val == nil {
			fmt.Fprintf(os.Stderr, "fact not found: %s\n", queries[0])
			os.Exit(2)
		}
		// If it's a sub-map, output as JSON or plain
		if m, ok := val.(map[string]any); ok {
			outputAll(m, *plain)
		} else {
			fmt.Println(val)
		}
		return
	}

	// Multiple queries -> filtered output
	filtered := make(map[string]any)
	for _, q := range queries {
		val := lookup(facts, q)
		if val == nil {
			continue
		}
		if m, ok := val.(map[string]any); ok {
			for k, v := range m {
				filtered[k] = v
			}
		} else {
			filtered[q] = val
		}
	}
	if len(filtered) == 0 {
		fmt.Fprintln(os.Stderr, "no matching facts found")
		os.Exit(2)
	}
	outputAll(filtered, *plain)
}

func outputAll(facts map[string]any, plain bool) {
	if plain {
		keys := make([]string, 0, len(facts))
		for k := range facts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%v\n", k, facts[k])
		}
	} else {
		// Build nested JSON from dotted keys
		nested := buildNested(facts)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(nested)
	}
}

// buildNested converts a flat dotted-key map to a nested structure for JSON output.
func buildNested(flat map[string]any) map[string]any {
	root := make(map[string]any)
	for key, val := range flat {
		parts := strings.Split(key, ".")
		current := root
		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = val
			} else {
				if next, ok := current[part]; ok {
					if m, ok := next.(map[string]any); ok {
						current = m
					} else {
						// Collision: leaf value exists where we need a map.
						// Preserve the leaf under a "_value" key.
						m := map[string]any{"_value": next}
						current[part] = m
						current = m
					}
				} else {
					m := make(map[string]any)
					current[part] = m
					current = m
				}
			}
		}
	}
	return root
}

func lookup(facts map[string]any, key string) any {
	if v, ok := facts[key]; ok {
		return v
	}
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
