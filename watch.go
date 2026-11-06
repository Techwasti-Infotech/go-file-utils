package fileutils

import (
	"context"
	"io/fs"
	"path/filepath"
	"time"
)

// FileEventType defines the kind of filesystem mutation observed.
type FileEventType string

const (
	EventCreated  FileEventType = "CREATED"
	EventModified FileEventType = "MODIFIED"
	EventDeleted  FileEventType = "DELETED"
)

// FileEvent represents a filesystem change notification.
type FileEvent struct {
	Type FileEventType `json:"type"`
	Path string        `json:"path"`
}

type fileSnapshot struct {
	size    int64
	modTime time.Time
	isDir   bool
}

// WatchPath polls a file or directory tree for modifications at pollingInterval.
// It sends FileEvent instances to the returned channel until ctx is cancelled.
// Operates with 0 external dependencies (pure Go standard library).
func WatchPath(ctx context.Context, targetPath string, pollingInterval time.Duration) (<-chan FileEvent, error) {
	if !Exists(targetPath) {
		return nil, ErrSourceNotFound
	}

	if pollingInterval <= 0 {
		pollingInterval = 200 * time.Millisecond
	}

	events := make(chan FileEvent, 32)

	takeSnapshot := func() map[string]fileSnapshot {
		snap := make(map[string]fileSnapshot)
		_ = filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			snap[path] = fileSnapshot{
				size:    info.Size(),
				modTime: info.ModTime(),
				isDir:   d.IsDir(),
			}
			return nil
		})
		return snap
	}

	previous := takeSnapshot()

	go func() {
		defer close(events)
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current := takeSnapshot()

				// Detect created and modified
				for path, currSnap := range current {
					prevSnap, exists := previous[path]
					if !exists {
						events <- FileEvent{Type: EventCreated, Path: path}
					} else if !currSnap.isDir && (currSnap.size != prevSnap.size || currSnap.modTime.After(prevSnap.modTime)) {
						events <- FileEvent{Type: EventModified, Path: path}
					}
				}

				// Detect deleted
				for path := range previous {
					if _, exists := current[path]; !exists {
						events <- FileEvent{Type: EventDeleted, Path: path}
					}
				}

				previous = current
			}
		}
	}()

	return events, nil
}
