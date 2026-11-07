package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 09: Directory Sync & Filesystem Watcher ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-09-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")

	_ = fileutils.EnsureDir(filepath.Join(srcDir, "models"), 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "models", "bert.bin"), []byte("model binary weights"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "manifest.json"), []byte(`{"version": 1}`), 0o644)

	// Extra destination file to test orphan deletion
	_ = fileutils.EnsureDir(dstDir, 0o755)
	_ = os.WriteFile(filepath.Join(dstDir, "orphan.old"), []byte("deprecated"), 0o644)

	// 1. Directory Mirror Synchronization (with checksum matching and orphan pruning)
	res, err := fileutils.SyncDir(srcDir, dstDir, fileutils.SyncOptions{
		Delete:     true,
		ByChecksum: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("1. Synchronized directory:\n   Created: %d items\n   Updated: %d items\n   Deleted: %d items (orphan.old removed)\n",
		len(res.Created), len(res.Updated), len(res.Deleted))

	// 2. Zero-dependency polling watcher (WatchPath)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("2. Launching background WatchPath listener...")
	events, err := fileutils.WatchPath(ctx, dstDir, 100*time.Millisecond)
	if err != nil {
		panic(err)
	}

	go func() {
		// Mutate files after a brief pause
		time.Sleep(200 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(dstDir, "realtime.txt"), []byte("new event file"), 0o644)
	}()

	for ev := range events {
		fmt.Printf("   [EVENT] %s -> %s\n", ev.Type, filepath.Base(ev.Path))
		if ev.Type == fileutils.EventCreated && filepath.Base(ev.Path) == "realtime.txt" {
			fmt.Println("   Received expected change event! Closing watcher.")
			cancel()
		}
	}
}
