package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 08: Process File Locks & Log Rotation ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-08-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	sharedDb := filepath.Join(tempDir, "shared.db")

	// 1. Process File Lock (WithLock ensures mutual exclusion with timeout)
	err = fileutils.WithLock(sharedDb, 3*time.Second, func() error {
		fmt.Println("1. Acquired exclusive file lock successfully.")
		_ = os.WriteFile(sharedDb, []byte("database state updated under lock"), 0o644)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// 2. Safe scoped temporary file helper
	_ = fileutils.WithTempFile("task-worker-*", func(f *os.File) error {
		fmt.Printf("2. Operating in isolated temporary file: %s\n", f.Name())
		_, _ = f.WriteString("temporary scratch work")
		return nil
	})
	fmt.Println("   Scoped temporary file was automatically deleted upon callback completion.")

	// 3. Log Rotation: shift backups (.1, .2, .3)
	logFile := filepath.Join(tempDir, "service.log")
	for i := 1; i <= 5; i++ {
		_ = os.WriteFile(logFile, []byte(fmt.Sprintf("log generation %d\n", i)), 0o644)
		_ = fileutils.RotateFile(logFile, 3) // Keep max 3 backups
	}
	fmt.Println("3. Performed log rotations. Existing files in directory:")
	entries, _ := os.ReadDir(tempDir)
	for _, e := range entries {
		fmt.Printf("   - %s\n", e.Name())
	}

	// 4. RotateBySize: conditional rotation on size trigger
	bigFile := filepath.Join(tempDir, "big.log")
	_ = os.WriteFile(bigFile, make([]byte, 5000), 0o644)
	rotated, _ := fileutils.RotateBySize(bigFile, 4000, 3)
	fmt.Printf("4. Triggered RotateBySize (5000 bytes >= 4000 limit): %v\n", rotated)
}
