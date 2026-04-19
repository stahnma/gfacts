// Package external loads facts from external directories.
package external

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LoadOptions controls external fact loading behavior.
type LoadOptions struct {
	NoExec      bool
	ExecTimeout time.Duration
}

// Load scans a directory for external fact files and executables.
func Load(dir string, opts LoadOptions) (map[string]any, error) {
	result := make(map[string]any)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		ext := filepath.Ext(entry.Name())

		switch ext {
		case ".json":
			facts, err := loadJSON(path)
			if err != nil {
				slog.Warn("failed to load JSON fact file", "path", path, "error", err)
				continue
			}
			for k, v := range facts {
				result[k] = v
			}
		case ".txt":
			facts, err := loadText(path)
			if err != nil {
				slog.Warn("failed to load text fact file", "path", path, "error", err)
				continue
			}
			for k, v := range facts {
				result[k] = v
			}
		default:
			if opts.NoExec {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 != 0 {
				facts, err := loadExec(path, opts.ExecTimeout)
				if err != nil {
					slog.Warn("executable fact failed", "path", path, "error", err)
					continue
				}
				for k, v := range facts {
					result[k] = v
				}
			}
		}
	}

	return result, nil
}
