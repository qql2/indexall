package vault

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/construct/indexall/internal/db/gen"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors other devices' log files and applies new entries to SQLite.
type Watcher struct {
	vault     *Vault
	db        *sql.DB
	q         *gen.Queries
	positions map[string]int64 // file path → bytes already processed
	mu        sync.Mutex
}

// NewWatcher creates a Watcher for the given vault.
func NewWatcher(v *Vault, db *sql.DB, q *gen.Queries) *Watcher {
	return &Watcher{
		vault:     v,
		db:        db,
		q:         q,
		positions: make(map[string]int64),
	}
}

// Start begins watching the vault log directory and processing incoming changes.
// It runs until ctx is cancelled. Should be called in a goroutine.
func (w *Watcher) Start(ctx context.Context) error {
	logDir := filepath.Join(w.vault.dir, "log")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("vault watcher: %w", err)
	}
	if err := watcher.Add(logDir); err != nil {
		watcher.Close()
		return fmt.Errorf("vault watcher: failed to watch %s: %w", logDir, err)
	}

	// Process any files that arrived while we were offline
	if err := w.processAllFiles(ctx); err != nil {
		fmt.Printf("vault watcher: initial scan error: %v\n", err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !strings.HasSuffix(event.Name, ".jsonl") {
					continue
				}
				if w.isOwnFile(event.Name) {
					continue
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if err := w.processFile(ctx, event.Name); err != nil {
						fmt.Printf("vault watcher: error processing %s: %v\n", filepath.Base(event.Name), err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("vault watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

func (w *Watcher) processAllFiles(ctx context.Context) error {
	logDir := filepath.Join(w.vault.dir, "log")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		fullPath := filepath.Join(logDir, entry.Name())
		if w.isOwnFile(fullPath) {
			continue
		}
		if err := w.processFile(ctx, fullPath); err != nil {
			fmt.Printf("vault watcher: error processing %s: %v\n", entry.Name(), err)
		}
	}
	return nil
}

func (w *Watcher) processFile(ctx context.Context, path string) error {
	w.mu.Lock()
	offset := w.positions[path]
	w.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return err
		}
	}

	scanner := bufio.NewScanner(f)
	newOffset := offset

	for scanner.Scan() {
		line := scanner.Bytes()
		newOffset += int64(len(line)) + 1 // +1 for newline
		if len(line) == 0 {
			continue
		}

		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Printf("vault watcher: skipping malformed entry: %v\n", err)
			continue
		}

		if err := ApplyEntry(ctx, w.db, w.q, entry); err != nil {
			fmt.Printf("vault watcher: failed to apply entry %s (%s %s): %v\n",
				entry.ID, entry.Op, entry.Type, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	w.mu.Lock()
	w.positions[path] = newOffset
	w.mu.Unlock()

	return nil
}

// isOwnFile returns true if the given path is this device's log file.
func (w *Watcher) isOwnFile(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, w.vault.deviceID+"-")
}
