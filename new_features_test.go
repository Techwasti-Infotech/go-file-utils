package fileutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-perm-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("permissions check"), 0o644)

	// IsReadable
	if !IsReadable(testFile) {
		t.Errorf("Expected testFile to be readable")
	}
	if IsReadable(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Errorf("Expected nonexistent file not to be readable")
	}

	// IsWritable
	if !IsWritable(testFile) {
		t.Errorf("Expected testFile to be writable")
	}
	if !IsWritable(tempDir) {
		t.Errorf("Expected tempDir to be writable")
	}

	// IsExecutable
	exeFile := filepath.Join(tempDir, "run.exe")
	_ = os.WriteFile(exeFile, []byte("echo hi"), 0o755)
	if !IsExecutable(exeFile) {
		t.Errorf("Expected run.exe to be detected as executable")
	}

	// SetReadOnly
	if err := SetReadOnly(testFile, true); err != nil {
		t.Fatalf("SetReadOnly(true) failed: %v", err)
	}
	// Restore writable
	_ = SetReadOnly(testFile, false)
}

func TestRotate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-rotate-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "app.log")
	_ = os.WriteFile(logFile, []byte("log version 1"), 0o644)

	// Rotate 1: app.log -> app.log.1
	if err := RotateFile(logFile, 3); err != nil {
		t.Fatalf("RotateFile failed: %v", err)
	}
	if Exists(logFile) {
		t.Errorf("Expected app.log to be moved")
	}
	if !Exists(logFile + ".1") {
		t.Errorf("Expected app.log.1 to exist")
	}

	// Write new log version 2
	_ = os.WriteFile(logFile, []byte("log version 2"), 0o644)

	// Rotate 2: app.log -> app.log.1, app.log.1 -> app.log.2
	if err := RotateFile(logFile, 3); err != nil {
		t.Fatalf("RotateFile 2 failed: %v", err)
	}
	if !Exists(logFile+".1") || !Exists(logFile+".2") {
		t.Errorf("Expected app.log.1 and app.log.2 to exist")
	}

	// RotateBySize test
	_ = os.WriteFile(logFile, []byte("1234567890"), 0o644) // 10 bytes
	rotated, err := RotateBySize(logFile, 20, 3)
	if err != nil || rotated {
		t.Errorf("Expected no rotation when size < maxBytes")
	}

	rotated, err = RotateBySize(logFile, 5, 3)
	if err != nil || !rotated {
		t.Errorf("Expected rotation when size >= maxBytes")
	}
}

func TestDiskUsage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-disk-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	status, err := DiskUsage(tempDir)
	if err != nil {
		t.Fatalf("DiskUsage failed: %v", err)
	}

	if status.Total == 0 {
		t.Errorf("Expected total disk size > 0, got %d", status.Total)
	}
	if status.Free == 0 {
		t.Errorf("Expected free disk size > 0, got %d", status.Free)
	}
}

func TestBatchAndPrune(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-batch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create files: img_1.png, img_2.png, doc.txt
	_ = os.WriteFile(filepath.Join(tempDir, "img_1.png"), []byte("1"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "img_2.png"), []byte("2"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "doc.txt"), []byte("d"), 0o644)

	// Batch rename img_*.png -> photo_*.png
	results, err := BatchRename(tempDir, `^img_(\d+)\.png$`, `photo_$1.png`)
	if err != nil {
		t.Fatalf("BatchRename failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 renamed files, got %d", len(results))
	}
	if !Exists(filepath.Join(tempDir, "photo_1.png")) || !Exists(filepath.Join(tempDir, "photo_2.png")) {
		t.Errorf("Expected renamed photo files to exist")
	}

	// Test PruneEmptyDirs
	nestedEmpty := filepath.Join(tempDir, "empty_a", "empty_b", "empty_c")
	_ = EnsureDir(nestedEmpty, 0o755)

	nestedWithFile := filepath.Join(tempDir, "non_empty", "child")
	_ = EnsureDir(nestedWithFile, 0o755)
	_ = os.WriteFile(filepath.Join(nestedWithFile, "keep.txt"), []byte("keep"), 0o644)

	pruned, err := PruneEmptyDirs(tempDir)
	if err != nil {
		t.Fatalf("PruneEmptyDirs failed: %v", err)
	}
	if len(pruned) != 3 {
		t.Errorf("Expected 3 pruned directories, got %d: %+v", len(pruned), pruned)
	}
	if Exists(nestedEmpty) {
		t.Errorf("Expected empty dir hierarchy to be pruned")
	}
	if !Exists(nestedWithFile) {
		t.Errorf("Expected non_empty dir to be retained")
	}
}
