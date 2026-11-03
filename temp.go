package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// TempDir creates a temporary directory with the given prefix and returns a cleanup function
// that removes the directory and all of its contents.
func TempDir(prefix string) (dir string, cleanup func(), err error) {
	d, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup = func() {
		_ = os.RemoveAll(d)
	}
	return d, cleanup, nil
}

// TempFile creates a temporary file in standard temp directory with prefix and returns the open file
// and a cleanup function that closes and removes the file.
func TempFile(prefix string) (file *os.File, cleanup func(), err error) {
	f, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	filePath := f.Name()
	cleanup = func() {
		_ = f.Close()
		_ = os.Remove(filePath)
	}
	return f, cleanup, nil
}

// WithTempDir executes the given callback within an isolated temporary directory.
// The temporary directory and everything in it is guaranteed to be cleaned up when fn returns.
func WithTempDir(prefix string, fn func(dir string) error) error {
	dir, cleanup, err := TempDir(prefix)
	if err != nil {
		return err
	}
	defer cleanup()
	return fn(dir)
}

// WithTempFile executes the given callback with an isolated temporary file.
// The file is automatically closed and removed when fn returns.
func WithTempFile(prefix string, fn func(f *os.File) error) error {
	file, cleanup, err := TempFile(prefix)
	if err != nil {
		return err
	}
	defer cleanup()
	return fn(file)
}

// SanitizeFilename removes or replaces characters that are dangerous or illegal in file names across Windows and POSIX.
func SanitizeFilename(name string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	res := name
	for _, c := range invalidChars {
		res = replaceAllChars(res, c, "_")
	}
	return filepath.Clean(res)
}

func replaceAllChars(s, old, replacement string) string {
	for {
		idx := -1
		for i := 0; i+len(old) <= len(s); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}
		s = s[:idx] + replacement + s[idx+len(old):]
	}
	return s
}

