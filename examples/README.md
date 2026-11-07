# go-file-utils Interactive Examples & Reference

This directory contains standalone, runnable Go applications demonstrating each capability of the `go-file-utils` library.

Each example is self-contained with its own isolated temporary workspace, allowing you to run, test, and adapt patterns directly in your projects.

---

## Catalog of Examples

| Directory | Topic | Key APIs Demonstrated |
| :--- | :--- | :--- |
| **[`01_basic_fs/`](01_basic_fs/main.go)** | Core Filesystem | `EnsureDir`, `Touch`, `Exists`, `IsFile`, `CopyFile`, `GetInfo`, `GetDirSize` |
| **[`02_atomic_and_streaming/`](02_atomic_and_streaming/main.go)** | Safe I/O & Streaming | `WriteFileAtomic`, `AppendToFile`, `ReadLines`, `CopyWithProgress` |
| **[`03_finder_and_traversal/`](03_finder_and_traversal/main.go)** | File Tree Finder | `Find`, `FindByExtension`, `SearchFilter` (Regex, MinSize, SkipHidden, ModifiedAfter) |
| **[`04_archive_zip_tar/`](04_archive_zip_tar/main.go)** | Secure Compression | `Zip`, `Unzip`, `TarGz`, `UntarGz` (with automatic Zip Slip directory escape defense) |
| **[`05_checksums_and_crypto/`](05_checksums_and_crypto/main.go)** | Hashing & Encryption | `FileSHA256`, `FileMD5`, `VerifyFileChecksum`, `EncryptFile` & `DecryptFile` (AES-256-GCM) |
| **[`06_advanced_manipulation/`](06_advanced_manipulation/main.go)** | Advanced Processing | `DetectMimeType`, `IsText`, `IsBinary`, `ReplaceInFile`, `HeadLines`, `TailLines`, `SplitFile`, `JoinFiles` |
| **[`07_structured_data/`](07_structured_data/main.go)** | Serialization | `WriteJSON`, `ReadJSON`, `WriteCSV`, `ReadCSV` |
| **[`08_locks_and_rotation/`](08_locks_and_rotation/main.go)** | Concurrency & Logs | `WithLock`, `WithTempFile`, `RotateFile`, `RotateBySize` |
| **[`09_sync_and_watch/`](09_sync_and_watch/main.go)** | Sync & Event Watching | `SyncDir` (mirroring & orphan pruning), `WatchPath` (`CREATED`, `MODIFIED`, `DELETED`) |
| **[`10_disk_and_permissions/`](10_disk_and_permissions/main.go)** | Disk & Bulk Actions | `DiskUsage`, `IsReadable`, `IsWritable`, `SetReadOnly`, `BatchRename`, `PruneEmptyDirs` |

---

## How to Run Any Example

From the root of the repository, execute:

```bash
# Example: Run basic filesystem operations
go run ./examples/01_basic_fs

# Example: Run directory sync & watcher
go run ./examples/09_sync_and_watch

# Example: Run disk metrics and permissions
go run ./examples/10_disk_and_permissions
```
