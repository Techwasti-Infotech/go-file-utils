# Architecture & Design Guide - go-file-utils

`go-file-utils` is an idiomatic, high-performance Go filesystem library designed to bridge the gap between low-level `os`/`io` primitives and daily production development tasks.

---

## 1. Core Design Principles

1. **Zero External Dependencies**:
   - Written exclusively using the Go Standard Library (`os`, `io`, `path/filepath`, `archive/zip`, `archive/tar`, `crypto/*`).
   - Guarantees fast builds, low binary footprint, and 0 supply-chain security vulnerabilities.

2. **Cross-Platform Consistency**:
   - Transparently handles filesystem differences between Windows (`\`) and POSIX/Unix (`/`).
   - Normalizes path separators in archives and regex searches.
   - Provides safe fallback for file moves across different mount points / volumes.

3. **Security First**:
   - **Zip Slip Defense**: Prevents directory traversal attacks (`../../`) inside archives by sanitizing target extraction paths.
   - **Atomic File Writing**: Writes via temporary staging files and performs atomic replacement to guarantee zero partial-write corruption on sudden process termination or power loss.
   - **Filename Sanitization**: Cleans user-supplied strings of illegal and dangerous control characters (`:`, `*`, `?`, `"`, `<`, `>`, `|`, `/`, `\`).

4. **Resource Safety**:
   - All streams use bounded buffers (e.g. 64 KB default) preventing out-of-memory errors even on gigabyte or terabyte files.
   - High-level scoped callbacks (`WithTempDir`, `WithTempFile`, `WithLock`) guarantee automatic resource release via Go `defer` mechanics.

---

## 2. Architecture Diagram

```
+---------------------------------------------------------------------------------------+
|                                    Client Application                                 |
+---------------------------------------------------------------------------------------+
                                           |
                                           v
+---------------------------------------------------------------------------------------+
|                                    go-file-utils API                                  |
+---------------------------------------------------------------------------------------+
   |                  |                    |                  |                    |
   v                  v                    v                  v                    v
[ Core FS ]     [ Atomic/Stream ]      [ Finder ]       [ Archive ]        [ Lock & Temp ]
 - Exists        - WriteFileAtomic      - Find           - Zip (ZipSlip)    - WithLock
 - CopyFile      - AppendToFile         - FindByName     - Unzip            - FileLock
 - CopyDir       - ReadLines            - FindByExt      - TarGz            - WithTempDir
 - Move          - CopyWithProgress     - Size/Date/Reg  - UntarGz          - Sanitize
 - Touch         - Line-by-Line         - Depth control                     - WithTempFile
   |                  |                    |                  |                    |
+---------------------------------------------------------------------------------------+
|                                  Go Standard Library                                  |
|         (os, io, path/filepath, bufio, crypto/*, archive/zip, archive/tar)            |
+---------------------------------------------------------------------------------------+
```

---

## 3. Module Breakdown

### 3.1 Core Filesystem (`fileutil.go`)
- **Inspection**: `Exists`, `IsFile`, `IsDir`, `IsSymlink`, `IsEmptyDir`, `GetInfo`, `GetDirSize`.
- **Directory Operations**: `EnsureDir` and `EnsureParentDir` (recursive directory creation), `EmptyDir`.
- **File Management**: `Touch`, `RemoveIfExists`, `CopyFile` (with timestamp preservation and buffer tuning), `CopyDir`, `Move`.

### 3.2 Atomic & Streaming I/O (`stream.go`)
- **Atomic Writes**: `WriteFileAtomic` writes into a sibling hidden temporary file, flushes dirty pages with `Sync()`, chmods, and atomically renames.
- **Append**: `AppendToFile` safely ensures parent directories and writes with `O_APPEND`.
- **Line Readers**: `ReadLines` processes large files line by line using an early-exit callback pattern to avoid storing the entire file in RAM; `ReadLinesAll` offers convenience.
- **Progress Streaming**: `CopyWithProgress` streams bytes between `io.Reader` and `io.Writer` while invoking a progress callback with `(bytesCopied, totalBytes)`.

### 3.3 Advanced Finder (`finder.go`)
- `Find` supports a rich declarative `SearchFilter`:
  - Shell glob `Pattern`
  - Compiled `Regex`
  - File `Extensions` (case-insensitive)
  - `MinSize` and `MaxSize` filters
  - `ModifiedAfter` and `ModifiedBefore` time windows
  - `MaxDepth` restriction
  - `SkipHidden` flag for dotfiles

### 3.4 Secure Archiving (`archive.go`)
- Supports both `.zip` and `.tar.gz` formats.
- Built-in `sanitizeArchivePath` enforces strict directory root boundaries, catching path traversal attacks (`../../`) before writing to disk.

### 3.5 Checksums & Hashes (`hash.go`)
- Streams `MD5`, `SHA1`, `SHA256`, and `SHA512` without reading entire files into memory.
- `VerifyFileChecksum` provides quick validation against known hashes.

### 3.6 Scoped Temp & Concurrency Locks (`temp.go`, `lock.go`)
- `WithTempDir` and `WithTempFile` manage temporary lifecycles reliably.
- `FileLock` provides atomic mutual exclusion across processes or workers using `O_CREATE|O_EXCL` lock files with customizable timeouts.

