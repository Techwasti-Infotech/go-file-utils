# go-file-utils

[![Go Report Card](https://goreportcard.com/badge/github.com/Techwasti-Infotech/go-file-utils)](https://goreportcard.com/report/github.com/Techwasti-Infotech/go-file-utils)
[![CI](https://github.com/Techwasti-Infotech/go-file-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/Techwasti-Infotech/go-file-utils/actions/workflows/ci.yml)
[![Documentation](https://img.shields.io/badge/docs-GitHub%20Pages-blue.svg)](https://techwasti-infotech.github.io/go-file-utils/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A modern, high-performance, **zero-dependency** Go filesystem utility library. Designed for production services, CLI applications, data pipelines, and distributed tools requiring reliable, secure, and cross-platform file operations.

---

## 🌟 Key Features

- ⚡ **Zero External Dependencies**: 100% pure standard library.
- 🛡️ **Zip Slip & Traversal Protection**: Archive extractors actively sanitize paths to prevent directory traversal exploits.
- 💾 **Atomic File Writes**: Prevent corrupted or partially-written files during sudden crashes or power failures.
- 🔄 **Chunked & Streamed I/O**: Constant, low memory usage (default 64 KB buffers) with real-time transfer progress callbacks.
- 🔍 **Powerful File Finder**: Filter files by Glob patterns, compiled Regex, file extensions, file sizes, modification timestamps, and directory depth.
- 📦 **Archive Helpers**: Simple single-call Zip and Tar.gz compression and extraction.
- 🔐 **Cryptographic Checksums**: Streamed MD5, SHA-1, SHA-256, and SHA-512 hashes.
- 🔒 **File Locking & Scoped Temp**: Atomic inter-process file locking (`FileLock`) and auto-cleaning temporary directories (`WithTempDir`).
- 🔎 **MIME & Content Detection**: Sniff MIME types via `DetectMimeType` and determine text vs binary files (`IsBinary`, `IsText`).
- ⚡ **File Comparison & Substitution**: Byte-by-byte file comparison (`CompareFiles`) and atomic string replacement (`ReplaceInFile`).
- ✂️ **Chunk Splitting & Joining**: Split gigabyte files into chunks (`SplitFile`) and reassemble without data loss (`JoinFiles`).
- 📜 **Head, Tail & Line Counting**: Read top N lines (`HeadLines`), bottom N lines (`TailLines`) via reverse chunk seek, and rapid line counts (`CountLines`).
- 🛡️ **AES-256-GCM File Encryption**: Encrypt (`EncryptFile`) and decrypt (`DecryptFile`) files using authenticated AES-GCM cryptography.
- 📑 **Structured Data Serialization**: Atomic JSON (`ReadJSON`, `WriteJSON`) and CSV (`ReadCSV`, `WriteCSV`) persistence.
- 💽 **Disk Space & Filesystem Metrics**: Retrieve volume space (`DiskUsage`) with total, used, free, available, and percentage.
- 🔄 **Log Rotation**: Keep files bounded with index shifting (`RotateFile`) or size thresholds (`RotateBySize`).
- 🏷️ **Batch Operations**: Bulk regex file renaming (`BatchRename`) and recursive empty directory pruning (`PruneEmptyDirs`).
- 📁 **Directory Synchronization**: High-performance mirror sync (`SyncDir`) with checksum-based or timestamp-based comparison, and orphan pruning (`SyncOptions.Delete`).
- 👀 **Filesystem Watching**: Polling-based file and directory watcher (`WatchPath`) with real-time `CREATED`, `MODIFIED`, and `DELETED` change events.
- 🔑 **Permissions & Access Checks**: Fast access checks (`IsReadable`, `IsWritable`, `IsExecutable`, `SetReadOnly`).
- 📏 **Human-Readable Formatting**: Format byte values to readable units with `HumanBytes`.

---

## 📦 Installation

```bash
go get github.com/Techwasti-Infotech/go-file-utils
```

---

## 🚀 Quick Start & Usage Examples

### 1. Basic Filesystem Operations

```go
package main

import (
	"fmt"
	"github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	// Check existence
	if fileutils.Exists("configs/settings.json") {
		fmt.Println("Config file exists!")
	}

	// Ensure parent directory exists and touch a file
	_ = fileutils.EnsureDir("logs/2026/03", 0o755)
	_ = fileutils.Touch("logs/2026/03/app.log")

	// Safe copy with timestamps preserved
	_ = fileutils.CopyFile("logs/2026/03/app.log", "logs/backup.log", fileutils.CopyFileOptions{
		Overwrite:          true,
		PreserveTimestamps: true,
	})

	// Retrieve clean file metadata
	info, _ := fileutils.GetInfo("logs/backup.log")
	fmt.Printf("File: %s, Size: %d bytes\n", info.Name, info.Size)
}
```

### 2. Atomic Writes & Streaming Line Readers

```go
// Atomically write data without risking partial writes during crashes
data := []byte(`{"status": "ok"}`)
err := fileutils.WriteFileAtomic("configs/app.json", data, 0o644)

// Stream large log files line by line
err = fileutils.ReadLines("logs/large.log", func(line string) (bool, error) {
    if strings.Contains(line, "FATAL") {
        fmt.Println("Alert:", line)
        return false, nil // Stop reading early
    }
    return true, nil // Continue reading
})
```

### 3. Advanced File Finder

```go
// Find all Go files larger than 1KB modified in the last 24 hours
files, err := fileutils.Find(".", fileutils.SearchFilter{
    Extensions:    []string{".go"},
    MinSize:       1024,
    ModifiedAfter: time.Now().Add(-24 * time.Hour),
    SkipHidden:    true,
    OnlyFiles:     true,
})
```

### 4. Secure Archiving (Zip & Tar.gz)

```go
// Compress an entire directory tree
_ = fileutils.Zip("data/reports", "data/reports.zip")

// Extract with automatic Zip Slip security defense
_ = fileutils.Unzip("data/reports.zip", "data/extracted")
```

### 5. Checksums & Verification

```go
// Calculate streaming SHA256 checksum
hash, _ := fileutils.FileSHA256("data/reports.zip")

// Verify integrity against expected checksum
match, _ := fileutils.VerifyFileChecksum("data/reports.zip", fileutils.HashSHA256, hash)
```

### 6. Process File Locking

```go
// Hold exclusive atomic lock during critical section
err := fileutils.WithLock("database.sqlite", 5*time.Second, func() error {
    // Perform isolated write
    return nil
})
```

---

## 💡 Runnable Examples & Cookbook

We provide a complete suite of runnable standalone example programs inside the [`examples/`](examples/) directory:

- [**01_basic_fs**](examples/01_basic_fs/main.go): Directory creation, file touch, copying with timestamps, file info.
- [**02_atomic_and_streaming**](examples/02_atomic_and_streaming/main.go): Atomic file writes, line-by-line reading, and streaming with progress callbacks.
- [**03_finder_and_traversal**](examples/03_finder_and_traversal/main.go): Recursive search with regex, size thresholds, and time filters.
- [**04_archive_zip_tar**](examples/04_archive_zip_tar/main.go): Zip & Tar.gz creation and extraction with Zip Slip defense.
- [**05_checksums_and_crypto**](examples/05_checksums_and_crypto/main.go): SHA-256, MD5, and AES-256-GCM file encryption/decryption.
- [**06_advanced_manipulation**](examples/06_advanced_manipulation/main.go): MIME sniffing, string substitution, head/tail, and file chunk splitting/joining.
- [**07_structured_data**](examples/07_structured_data/main.go): Atomic JSON and CSV reading and writing.
- [**08_locks_and_rotation**](examples/08_locks_and_rotation/main.go): Process file locking, scoped temp files, and log file rotation.
- [**09_sync_and_watch**](examples/09_sync_and_watch/main.go): Directory synchronization (with orphan pruning) and zero-dependency filesystem watcher.
- [**10_disk_and_permissions**](examples/10_disk_and_permissions/main.go): Drive metrics (`DiskUsage`), permissions inspection, regex batch rename, and empty folder pruning.

Run any example directly with:
```bash
go run ./examples/01_basic_fs
```

---

## 📖 API Documentation & Website

Browse the interactive online documentation hosted on GitHub Pages:
👉 **[https://techwasti-infotech.github.io/go-file-utils/](https://techwasti-infotech.github.io/go-file-utils/)**

Additional reference documents:
- [Architecture Guide](ARCHITECTURE.md)
- [Contributing Guide](CONTRIBUTING.md)

---

## 🧪 Testing

Run all unit tests with data race detector and coverage report:

```bash
go test -v -race -cover ./...
```

---

## 📄 License

This library is distributed under the terms of the MIT License. See [LICENSE](LICENSE) for details.

