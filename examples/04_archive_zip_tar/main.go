package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 04: Secure Zip & Tar.gz Archiving ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-04-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Prepare files to package
	dataDir := filepath.Join(tempDir, "dataset")
	_ = fileutils.EnsureDir(filepath.Join(dataDir, "images"), 0o755)
	_ = os.WriteFile(filepath.Join(dataDir, "index.txt"), []byte("Dataset Index v1"), 0o644)
	_ = os.WriteFile(filepath.Join(dataDir, "images", "photo.raw"), []byte("raw image binary data"), 0o644)

	// 1. Create ZIP archive
	zipFile := filepath.Join(tempDir, "archive.zip")
	if err := fileutils.Zip(dataDir, zipFile); err != nil {
		panic(err)
	}
	info, _ := fileutils.GetInfo(zipFile)
	fmt.Printf("1. Created ZIP: %s (%s)\n", zipFile, fileutils.HumanBytes(info.Size))

	// 2. Unzip with built-in Zip Slip protection
	extractedZipDir := filepath.Join(tempDir, "unzipped_output")
	if err := fileutils.Unzip(zipFile, extractedZipDir); err != nil {
		panic(err)
	}
	fmt.Printf("2. Safely extracted ZIP to: %s\n", extractedZipDir)

	// 3. Create Tar.gz archive
	targzFile := filepath.Join(tempDir, "archive.tar.gz")
	if err := fileutils.TarGz(dataDir, targzFile); err != nil {
		panic(err)
	}
	tarInfo, _ := fileutils.GetInfo(targzFile)
	fmt.Printf("3. Created Tar.gz: %s (%s)\n", targzFile, fileutils.HumanBytes(tarInfo.Size))

	// 4. Untar with traversal guard
	extractedTarDir := filepath.Join(tempDir, "untarred_output")
	if err := fileutils.UntarGz(targzFile, extractedTarDir); err != nil {
		panic(err)
	}
	fmt.Printf("4. Safely extracted Tar.gz to: %s\n", extractedTarDir)
}
