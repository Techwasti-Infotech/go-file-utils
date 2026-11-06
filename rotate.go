package fileutils

import (
	"fmt"
	"os"
)

// RotateFile rotates the target file by shifting existing backups (.1, .2, ...)
// and moving the current file to .1. Any backup index exceeding maxBackups is removed.
// If maxBackups <= 0, no backups are kept and the target is simply removed.
func RotateFile(path string, maxBackups int) error {
	if !Exists(path) {
		return nil
	}

	if maxBackups <= 0 {
		return os.Remove(path)
	}

	// Delete the oldest backup if it exists
	oldestBackup := fmt.Sprintf("%s.%d", path, maxBackups)
	if Exists(oldestBackup) {
		if err := os.Remove(oldestBackup); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// Shift backups backward: e.g. path.2 -> path.3, path.1 -> path.2
	for i := maxBackups - 1; i >= 1; i-- {
		current := fmt.Sprintf("%s.%d", path, i)
		next := fmt.Sprintf("%s.%d", path, i+1)
		if Exists(current) {
			if Exists(next) {
				_ = os.Remove(next)
			}
			if err := os.Rename(current, next); err != nil {
				return err
			}
		}
	}

	// Rotate the primary file to .1
	firstBackup := fmt.Sprintf("%s.1", path)
	if Exists(firstBackup) {
		_ = os.Remove(firstBackup)
	}

	return os.Rename(path, firstBackup)
}

// RotateBySize inspects the file at path. If its size is greater than or equal to
// maxBytes, it triggers RotateFile.
// Returns true if rotation occurred, false otherwise.
func RotateBySize(path string, maxBytes int64, maxBackups int) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if info.Size() >= maxBytes {
		err := RotateFile(path, maxBackups)
		return err == nil, err
	}

	return false, nil
}
