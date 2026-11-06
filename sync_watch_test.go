package fileutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-sync-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	_ = EnsureDir(filepath.Join(srcDir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("data 1"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "sub", "file2.txt"), []byte("data 2"), 0o644)

	// Extra destination file to test Delete option
	_ = EnsureDir(dstDir, 0o755)
	_ = os.WriteFile(filepath.Join(dstDir, "extra.txt"), []byte("orphan"), 0o644)

	res, err := SyncDir(srcDir, dstDir, SyncOptions{Delete: true, ByChecksum: true})
	if err != nil {
		t.Fatalf("SyncDir failed: %v", err)
	}

	if len(res.Created) != 3 { // sub dir, file1.txt, file2.txt
		t.Errorf("Expected 3 created items, got %d: %+v", len(res.Created), res.Created)
	}
	if len(res.Deleted) != 1 {
		t.Errorf("Expected 1 deleted item (extra.txt), got %d: %+v", len(res.Deleted), res.Deleted)
	}
	if Exists(filepath.Join(dstDir, "extra.txt")) {
		t.Errorf("Expected extra.txt to be deleted")
	}

	// Update a file and sync again
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("updated data 1"), 0o644)

	res2, err := SyncDir(srcDir, dstDir, SyncOptions{Delete: false, ByChecksum: true})
	if err != nil {
		t.Fatalf("SyncDir update failed: %v", err)
	}
	if len(res2.Updated) != 1 {
		t.Errorf("Expected 1 updated item, got %d", len(res2.Updated))
	}
}

func TestWatchPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-watch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := WatchPath(ctx, tempDir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WatchPath failed: %v", err)
	}

	// Create a new file
	time.Sleep(100 * time.Millisecond)
	testFile := filepath.Join(tempDir, "watched.txt")
	_ = os.WriteFile(testFile, []byte("initial"), 0o644)

	var receivedCreated bool
	timeout := time.After(800 * time.Millisecond)

	for !receivedCreated {
		select {
		case ev, ok := <-events:
			if !ok {
				break
			}
			if ev.Type == EventCreated && ev.Path == testFile {
				receivedCreated = true
			}
		case <-timeout:
			break
		}
	}

	if !receivedCreated {
		t.Errorf("Expected to receive EventCreated for watched.txt")
	}
}
