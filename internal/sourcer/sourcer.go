// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fileReader "github.com/driscollos/config/internal/sourcer/file-reader"
	terminalReader "github.com/driscollos/config/internal/sourcer/terminal-reader"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -destination=../mocks/mock-data-sourcer.go -package=mocks . Sourcer

// Sourcer provides values to the populator by looking through multiple inputs
// (CLI/terminal, env vars, and one or more files).
type Sourcer interface {
	Get(path string) string
	HotReload(ctx context.Context, onChange func())
	Source(path string)
}

// sourcer implements Sourcer.
type sourcer struct {
	readers struct {
		file     fileReader.FileReader
		terminal terminalReader.TerminalReader
	}
	sources struct {
		files          []string
		hotReload      map[string]time.Time
		useCommandLine bool
		useEnvironment bool
	}

	mu      sync.RWMutex
	isSetup bool
	values  []map[string]interface{}
}

// Source sets the single file to read from and resets setup state.
func (s *sourcer) Source(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sources.files = []string{path}
	if s.sources.hotReload == nil {
		s.sources.hotReload = make(map[string]time.Time)
	}
	s.isSetup = false
}

// HotReload watches for file changes and replaces the in-memory values
// atomically when any source file changes. onChange is invoked after a swap.
func (s *sourcer) HotReload(ctx context.Context, onChange func()) {
	ticker := time.NewTicker(time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Snapshot current files and mtimes without blocking lookups for long.
				s.mu.RLock()
				files := append([]string(nil), s.sources.files...)
				last := make(map[string]time.Time, len(s.sources.hotReload))
				for k, v := range s.sources.hotReload {
					last[k] = v
				}
				s.mu.RUnlock()

				changed := false
				newValues := make([]map[string]interface{}, 0, len(files))
				newMtimes := make(map[string]time.Time, len(files))

				for _, file := range files {
					info, err := os.Stat(file)
					if err != nil {
						continue
					}
					// Detect change
					if info.ModTime().After(last[file]) {
						changed = true
					}

					bytes, err := os.ReadFile(file)
					if err != nil {
						continue
					}

					data := make(map[string]interface{})
					switch ext := fileExt(file); ext {
					case "yml", "yaml":
						if err := yaml.Unmarshal(bytes, &data); err != nil {
							continue
						}
					case "json":
						if err := json.Unmarshal(bytes, &data); err != nil {
							continue
						}
					default:
						continue
					}

					newValues = append(newValues, data)
					newMtimes[file] = info.ModTime()
				}

				if changed {
					s.mu.Lock()
					s.values = newValues // replace, don't append
					// update mtimes
					if s.sources.hotReload == nil {
						s.sources.hotReload = make(map[string]time.Time, len(newMtimes))
					}
					for k, v := range newMtimes {
						s.sources.hotReload[k] = v
					}
					s.mu.Unlock()

					onChange()
				}
			}
		}
	}()
}

// Get retrieves the value associated with a "path" from the available sources.
// Precedence: CLI/terminal > environment variables > files (last non-nil wins).
// For composite values (map/slice), it returns full JSON for predictability.
func (s *sourcer) Get(path string) string {
	if err := s.setup(); err != nil {
		return ""
	}

	// 1) CLI/terminal
	if s.sources.useCommandLine {
		if argVal, err := s.readers.terminal.Get(path); err == nil && argVal != "" {
			return argVal
		}
	}

	// 2) Environment
	if s.sources.useEnvironment {
		if val, ok := os.LookupEnv(path); ok && val != "" {
			return val
		}
	}

	// 3) Files (later non-nil overrides earlier)
	var retVal interface{}
	s.mu.RLock()
	for _, source := range s.values {
		if v := s.get(source, path); v != nil {
			retVal = v
		}
	}
	s.mu.RUnlock()

	if retVal == nil {
		return ""
	}

	switch v := retVal.(type) {
	case string:
		return strings.TrimSpace(v)
	case float32, float64, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	case []interface{}, map[string]interface{}:
		b, _ := json.Marshal(v) // lossless; populator can parse
		return string(b)
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

// --- internals ---

func (s *sourcer) setup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isSetup {
		return nil
	}

	if s.sources.hotReload == nil {
		s.sources.hotReload = make(map[string]time.Time)
	}

	s.values = s.values[:0] // clear

	var firstErr error
	for _, file := range s.sources.files {
		bytes, err := s.readers.file.Read(file)
		if err != nil {
			if len(s.sources.files) == 1 {
				return fmt.Errorf("could not read from source file : %s", file)
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err = s.loadFromSource(file, bytes); err != nil {
			return fmt.Errorf("error parsing source file : %s : %v", file, err)
		}
		if info, err := os.Stat(file); err == nil {
			s.sources.hotReload[file] = info.ModTime()
		} else {
			s.sources.hotReload[file] = time.Now()
		}
	}

	s.isSetup = true
	return nil
}

func (s *sourcer) loadFromSource(filename string, source []byte) error {
	data := make(map[string]interface{})
	ext := fileExt(filename)
	if ext == "" {
		return errors.New(ErrorUnknownFileFormat)
	}

	switch ext {
	case "yml", "yaml":
		if err := yaml.Unmarshal(source, &data); err != nil {
			return err
		}
	case "json":
		if err := json.Unmarshal(source, &data); err != nil {
			return err
		}
	default:
		return errors.New(ErrorUnknownFileFormat)
	}

	s.values = append(s.values, data)
	return nil
}

// get traverses a nested map[string]interface{} using either "_" or "." separators.
// It supports slice indexing with numeric segments. Returns nil if not found.
func (s *sourcer) get(source map[string]interface{}, path string) interface{} {
	if source == nil || path == "" {
		return nil
	}

	var parts []string
	if strings.Contains(path, ".") {
		parts = strings.Split(path, ".")
	} else {
		parts = strings.Split(path, "_")
	}

	// fast path for single segment
	if len(parts) == 1 {
		return source[parts[0]]
	}

	var cur any = source
	for i, seg := range parts {
		if cur == nil {
			return nil
		}
		last := i == len(parts)-1

		switch node := cur.(type) {
		case map[string]interface{}:
			cur = node[seg]
			if last {
				return cur
			}

		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil {
				return nil
			}
			if idx < 0 || idx >= len(node) { // ✅ correct slice bounds check
				return nil
			}
			cur = node[idx]
			if last {
				return cur
			}

		default:
			return nil
		}
	}
	return nil
}

// fileExt returns the lowercased extension (without dot) or "" if none.
func fileExt(filename string) string {
	bits := strings.Split(filename, ".")
	if len(bits) < 2 {
		return ""
	}
	return strings.ToLower(bits[len(bits)-1])
}

// --- notes ---
//
// * Precedence is CLI > ENV > files (later non-nil value from files wins).
// * Composite values returned as JSON strings are predictable and lossless.
//
