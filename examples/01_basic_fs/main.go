package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 01: Core Filesystem Operations ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-01-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Ensure directory exists
	nestedDir := filepath.Join(tempDir, "workspace", "projects")
	if err := fileutils.EnsureDir(nestedDir, 0o755); err != nil {
		panic(err)
	}
	fmt.Printf("1. EnsureDir created: %s (IsDir: %v)\n", nestedDir, fileutils.IsDir(nestedDir))

	// 2. Touch a file
	logFile := filepath.Join(nestedDir, "app.log")
	if err := fileutils.Touch(logFile); err != nil {
		panic(err)
	}
	fmt.Printf("2. Touched file: %s (Exists: %v, IsFile: %v)\n", logFile, fileutils.Exists(logFile), fileutils.IsFile(logFile))

	// 3. Write data & CopyFile with timestamps preserved
	_ = os.WriteFile(logFile, []byte("Hello from go-file-utils!"), 0o644)
	backupFile := filepath.Join(nestedDir, "app.log.bak")
	err = fileutils.CopyFile(logFile, backupFile, fileutils.CopyFileOptions{
		Overwrite:          true,
		PreserveTimestamps: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("3. Copied file to: %s\n", backupFile)

	// 4. Retrieve structured metadata
	info, err := fileutils.GetInfo(backupFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("4. FileInfo: Name=%s, Size=%d bytes, Ext=%s, ModTime=%s\n",
		info.Name, info.Size, info.Extension, info.ModTime.Format("2006-01-02 15:04:05"))

	// 5. Calculate total directory size
	dirSize, _ := fileutils.GetDirSize(nestedDir)
	fmt.Printf("5. Total directory size: %d bytes (%s)\n", dirSize, fileutils.HumanBytes(dirSize))
}
