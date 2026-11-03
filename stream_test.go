package fileutils

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamAndAtomicOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-stream-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "atomic_test.txt")

	// 1. Test WriteFileAtomic
	initialData := []byte("Initial safe atomic content")
	if err := WriteFileAtomic(targetFile, initialData, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	read, err := os.ReadFile(targetFile)
	if err != nil || !bytes.Equal(read, initialData) {
		t.Fatalf("WriteFileAtomic content mismatch. Got: %s", string(read))
	}

	// Overwrite atomically
	updatedData := []byte("Second safe atomic revision")
	if err := WriteFileAtomic(targetFile, updatedData, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic overwrite failed: %v", err)
	}
	read, _ = os.ReadFile(targetFile)
	if !bytes.Equal(read, updatedData) {
		t.Fatalf("Updated content mismatch")
	}

	// 2. Test AppendToFile
	appendFile := filepath.Join(tempDir, "append_test.txt")
	if err := AppendToFile(appendFile, []byte("line 1\n"), 0o644); err != nil {
		t.Fatalf("AppendToFile failed: %v", err)
	}
	if err := AppendToFile(appendFile, []byte("line 2\n"), 0o644); err != nil {
		t.Fatalf("AppendToFile second append failed: %v", err)
	}

	lines, err := ReadLinesAll(appendFile)
	if err != nil {
		t.Fatalf("ReadLinesAll failed: %v", err)
	}
	if len(lines) != 2 || lines[0] != "line 1" || lines[1] != "line 2" {
		t.Errorf("Unexpected lines from appendFile: %+v", lines)
	}

	// 3. Test ReadLines with early stop
	var scanned []string
	err = ReadLines(appendFile, func(line string) (bool, error) {
		scanned = append(scanned, line)
		return false, nil // Stop after first line
	})
	if err != nil || len(scanned) != 1 || scanned[0] != "line 1" {
		t.Errorf("ReadLines early stop failed: %v, scanned: %+v", err, scanned)
	}

	// 4. Test CopyWithProgress
	srcData := strings.Repeat("ABCDEFGHIJ", 1000) // 10,000 bytes
	srcBuf := bytes.NewBufferString(srcData)
	var dstBuf bytes.Buffer

	var lastProgress int64
	totalSize := int64(len(srcData))

	written, err := CopyWithProgress(&dstBuf, srcBuf, totalSize, 1024, func(bytesCopied int64, total int64) {
		if total != totalSize {
			t.Errorf("Expected total size %d, got %d", totalSize, total)
		}
		lastProgress = bytesCopied
	})
	if err != nil {
		t.Fatalf("CopyWithProgress failed: %v", err)
	}
	if written != totalSize || lastProgress != totalSize || dstBuf.String() != srcData {
		t.Errorf("CopyWithProgress verification failed: written=%d, lastProgress=%d", written, lastProgress)
	}
}

