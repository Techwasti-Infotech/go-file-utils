package fileutils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFileAtomic safely writes data to a temporary file in the same directory as path,
// flushes and syncs to disk, and then renames it atomically to path.
// This prevents data corruption during unexpected crashes or power failures.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir, DefaultDirPerm); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Ensure cleanup in case of failure before atomic rename
	var success bool
	defer func() {
		if !success {
			_ = tmpFile.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed writing to temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed syncing temp file: %w", err)
	}

	if err := tmpFile.Chmod(perm); err != nil {
		return fmt.Errorf("failed chmod on temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		// On Windows, os.Rename fails if target exists. Try removing target first.
		if Exists(path) {
			if rmErr := os.Remove(path); rmErr == nil {
				if renameErr := os.Rename(tmpName, path); renameErr == nil {
					success = true
					return nil
				}
			}
		}
		return fmt.Errorf("failed atomic rename: %w", err)
	}

	success = true
	return nil
}

// AppendToFile appends data to the specified file. Creates the file if it does not exist.
func AppendToFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureParentDir(path, DefaultDirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}

// ReadLines reads a text file line by line and passes each line to the handler callback.
// Returning false or an error from the callback stops scanning early.
func ReadLines(path string, onLine func(line string) (bool, error)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Support long lines up to 10MB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		keepGoing, err := onLine(line)
		if err != nil {
			return err
		}
		if !keepGoing {
			break
		}
	}
	return scanner.Err()
}

// ReadLinesAll reads all lines of a text file into a slice of strings.
func ReadLinesAll(path string) ([]string, error) {
	var lines []string
	err := ReadLines(path, func(line string) (bool, error) {
		lines = append(lines, line)
		return true, nil
	})
	return lines, err
}

// CopyProgressFunc is a callback invoked during chunked stream copying.
// It receives the total bytes copied so far and the total size (-1 if unknown).
type CopyProgressFunc func(bytesCopied int64, totalBytes int64)

// CopyWithProgress copies data from reader to writer with real-time progress reporting.
func CopyWithProgress(dst io.Writer, src io.Reader, totalSize int64, bufferSize int, onProgress CopyProgressFunc) (int64, error) {
	if bufferSize <= 0 {
		bufferSize = 64 * 1024 // 64 KB default
	}
	buf := make([]byte, bufferSize)
	var written int64

	for {
		nr, rErr := src.Read(buf)
		if nr > 0 {
			nw, wErr := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if wErr == nil {
					wErr = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if onProgress != nil {
				onProgress(written, totalSize)
			}
			if wErr != nil {
				return written, wErr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rErr != nil {
			if errors.Is(rErr, io.EOF) {
				break
			}
			return written, rErr
		}
	}
	return written, nil
}
