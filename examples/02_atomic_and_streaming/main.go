package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 02: Atomic Writes & Streaming I/O ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-02-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "config.json")

	// 1. Safe atomic file write (avoids corrupt writes on crashes)
	data := []byte(`{"service": "auth-api", "version": "v1.2.0"}`)
	if err := fileutils.WriteFileAtomic(targetFile, data, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("1. Atomically written to: %s\n", targetFile)

	// 2. Append lines safely
	logPath := filepath.Join(tempDir, "server.log")
	_ = fileutils.AppendToFile(logPath, []byte("2026-11-07 INFO Server started\n"), 0o644)
	_ = fileutils.AppendToFile(logPath, []byte("2026-11-07 ERROR Database connection timeout\n"), 0o644)
	_ = fileutils.AppendToFile(logPath, []byte("2026-11-07 INFO Retrying connection...\n"), 0o644)
	fmt.Println("2. Appended log entries.")

	// 3. Streaming line-by-line read with early exit
	fmt.Println("3. Scanning log line-by-line (stopping on ERROR):")
	err = fileutils.ReadLines(logPath, func(line string) (bool, error) {
		fmt.Printf("   -> %s\n", line)
		if strings.Contains(line, "ERROR") {
			fmt.Println("   [!] Triggered alert, terminating scan early.")
			return false, nil // Stop reading
		}
		return true, nil // Continue
	})
	if err != nil {
		panic(err)
	}

	// 4. Chunked stream copy with live progress callbacks
	payload := bytes.Repeat([]byte("Go File Utils High Throughput Streaming\n"), 1000)
	srcStream := bytes.NewReader(payload)
	var dstBuf bytes.Buffer

	fmt.Println("4. Streaming data with progress reporting:")
	totalBytes, err := fileutils.CopyWithProgress(&dstBuf, srcStream, int64(len(payload)), 8192, func(copied, total int64) {
		pct := (float64(copied) / float64(total)) * 100.0
		fmt.Printf("   Copied: %d / %d bytes (%.1f%%)\n", copied, total, pct)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Successfully streamed %d bytes.\n", totalBytes)
}
