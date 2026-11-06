package fileutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTempAndLockOperations(t *testing.T) {
	// 1. Test WithTempDir
	var createdDir string
	err := WithTempDir("test-scoped-*", func(dir string) error {
		createdDir = dir
		if !IsDir(dir) {
			t.Errorf("Expected temp dir %s to exist inside callback", dir)
		}
		testFile := filepath.Join(dir, "sub.txt")
		return os.WriteFile(testFile, []byte("data"), 0o644)
	})
	if err != nil {
		t.Fatalf("WithTempDir callback failed: %v", err)
	}
	if Exists(createdDir) {
		t.Errorf("Expected temp dir %s to be cleaned up after callback", createdDir)
	}

	// 2. Test WithTempFile
	var createdFilePath string
	err = WithTempFile("test-scoped-file-*", func(f *os.File) error {
		createdFilePath = f.Name()
		if !Exists(createdFilePath) {
			t.Errorf("Expected temp file to exist")
		}
		_, err := f.WriteString("temporary data")
		return err
	})
	if err != nil {
		t.Fatalf("WithTempFile failed: %v", err)
	}
	if Exists(createdFilePath) {
		t.Errorf("Expected temp file %s to be removed after callback", createdFilePath)
	}

	// 3. Test SanitizeFilename
	unsafeName := `foo/bar\baz:test*name?val"one<two>three|done.txt`
	safeName := SanitizeFilename(unsafeName)
	for _, bad := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		for i := 0; i+len(bad) <= len(safeName); i++ {
			if safeName[i:i+len(bad)] == bad {
				t.Errorf("Found forbidden char %s in sanitized string %s", bad, safeName)
			}
		}
	}

	// 4. Test FileLock
	tempDir, err := os.MkdirTemp("", "test-lock-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	resource := filepath.Join(tempDir, "shared.db")

	lock1 := NewFileLock(resource)
	if err := lock1.TryLock(); err != nil {
		t.Fatalf("lock1.TryLock failed: %v", err)
	}

	lock2 := NewFileLock(resource)
	if err := lock2.TryLock(); err != ErrAlreadyLocked {
		t.Errorf("lock2 expected ErrAlreadyLocked, got: %v", err)
	}

	// Lock with timeout should timeout
	timeoutErr := lock2.Lock(100*time.Millisecond, 20*time.Millisecond)
	if timeoutErr != ErrLockAcquireTimeout {
		t.Errorf("lock2 expected ErrLockAcquireTimeout, got: %v", timeoutErr)
	}

	// Unlock lock1 and lock2 should now succeed
	if err := lock1.Unlock(); err != nil {
		t.Fatalf("lock1.Unlock failed: %v", err)
	}

	if err := lock2.TryLock(); err != nil {
		t.Fatalf("lock2.TryLock should succeed after lock1 unlock, got: %v", err)
	}
	_ = lock2.Unlock()

	// 5. Test WithLock
	counter := 0
	err = WithLock(resource, 1*time.Second, func() error {
		counter++
		return nil
	})
	if err != nil || counter != 1 {
		t.Errorf("WithLock failed: %v, counter: %d", err, counter)
	}
}
