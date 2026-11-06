package fileutils

import (
	"os"
	"path/filepath"
	"strings"
)

// IsReadable checks if the file or directory exists and can be opened for reading.
func IsReadable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// IsWritable checks if a path can be written to.
// If it's a file, it checks if it can be opened with write permissions.
// If it's a directory, it attempts creating a temporary probe file.
func IsWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// If path doesn't exist, check if its parent directory is writable
		parent := filepath.Dir(path)
		return IsWritable(parent)
	}

	if info.IsDir() {
		probeFile, err := os.CreateTemp(path, ".probe-*")
		if err != nil {
			return false
		}
		probeName := probeFile.Name()
		_ = probeFile.Close()
		_ = os.Remove(probeName)
		return true
	}

	// Regular file check
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// IsExecutable checks whether the path is executable.
// On POSIX, it inspects permission bits (0111).
// On Windows, it verifies file existence and standard executable extensions (.exe, .bat, .cmd, .ps1).
func IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	// POSIX permission check
	if info.Mode()&0o111 != 0 {
		return true
	}

	// Windows executable extensions check
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".exe", ".bat", ".cmd", ".com", ".ps1":
		return true
	default:
		return false
	}
}

// SetReadOnly toggles the read-only attribute of a file.
func SetReadOnly(path string, readOnly bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	currentMode := info.Mode()
	if readOnly {
		// Remove write bits
		return os.Chmod(path, currentMode&^0o222)
	}
	// Restore owner and group write bits
	return os.Chmod(path, currentMode|0o644)
}
