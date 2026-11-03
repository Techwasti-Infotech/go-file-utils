package fileutils

import (
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	// ErrLockAcquireTimeout is returned when a file lock cannot be acquired within the timeout.
	ErrLockAcquireTimeout = errors.New("fileutils: timeout acquiring file lock")
	// ErrAlreadyLocked is returned when a non-blocking lock attempt encounters an existing lock.
	ErrAlreadyLocked = errors.New("fileutils: file is already locked")
)

// FileLock represents an exclusive lock on a file resource via an atomic lock file sentinel.
type FileLock struct {
	targetPath string
	lockPath   string
	lockFile   *os.File
}

// NewFileLock initializes a FileLock instance for targetPath.
func NewFileLock(targetPath string) *FileLock {
	return &FileLock{
		targetPath: targetPath,
		lockPath:   targetPath + ".lock",
	}
}

// TryLock attempts to acquire the lock immediately without waiting.
// Returns ErrAlreadyLocked if the lock is held by another process or goroutine.
func (fl *FileLock) TryLock() error {
	if err := EnsureParentDir(fl.lockPath, DefaultDirPerm); err != nil {
		return err
	}

	// O_CREATE | O_EXCL ensures atomic creation
	f, err := os.OpenFile(fl.lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return ErrAlreadyLocked
		}
		return err
	}

	fl.lockFile = f
	_, _ = fmt.Fprintf(f, "pid: %d\ntime: %s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	_ = f.Sync()
	return nil
}

// Lock waits up to timeout duration to acquire the lock, retrying at retryInterval intervals.
// A timeout of 0 waits indefinitely until acquired.
func (fl *FileLock) Lock(timeout time.Duration, retryInterval time.Duration) error {
	if retryInterval <= 0 {
		retryInterval = 25 * time.Millisecond
	}

	start := time.Now()
	for {
		err := fl.TryLock()
		if err == nil {
			return nil
		}
		if !errors.Is(err, ErrAlreadyLocked) {
			return err
		}

		if timeout > 0 && time.Since(start) >= timeout {
			return ErrLockAcquireTimeout
		}

		time.Sleep(retryInterval)
	}
}

// Unlock releases the lock and removes the lock file.
func (fl *FileLock) Unlock() error {
	if fl.lockFile != nil {
		_ = fl.lockFile.Close()
		fl.lockFile = nil
	}
	err := os.Remove(fl.lockPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// WithLock executes fn while holding an exclusive lock on targetPath.
func WithLock(targetPath string, timeout time.Duration, fn func() error) error {
	lock := NewFileLock(targetPath)
	if err := lock.Lock(timeout, 20*time.Millisecond); err != nil {
		return err
	}
	defer lock.Unlock()
	return fn()
}
