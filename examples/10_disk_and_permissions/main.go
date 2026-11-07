package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 10: Disk Metrics, Permissions & Batch Operations ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-10-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. DiskUsage metrics for current partition
	disk, err := fileutils.DiskUsage(tempDir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("1. Filesystem Metrics for %s:\n   Total:     %s\n   Used:      %s (%.1f%%)\n   Available: %s\n",
		tempDir, fileutils.HumanBytes(int64(disk.Total)), fileutils.HumanBytes(int64(disk.Used)),
		disk.UsedPct, fileutils.HumanBytes(int64(disk.Available)))

	// 2. Permission and Access Checks
	sampleFile := filepath.Join(tempDir, "document.txt")
	_ = os.WriteFile(sampleFile, []byte("Sensitive contract details"), 0o644)

	fmt.Printf("2. Permissions:\n   IsReadable:   %v\n   IsWritable:   %v\n   IsExecutable: %v\n",
		fileutils.IsReadable(sampleFile), fileutils.IsWritable(sampleFile), fileutils.IsExecutable(sampleFile))

	// Set Read-Only
	_ = fileutils.SetReadOnly(sampleFile, true)
	fmt.Printf("   After SetReadOnly(true) -> IsWritable: %v\n", fileutils.IsWritable(sampleFile))
	_ = fileutils.SetReadOnly(sampleFile, false)

	// 3. Batch Rename with Regex
	_ = os.WriteFile(filepath.Join(tempDir, "IMG_2026_01.jpg"), []byte("img1"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "IMG_2026_02.jpg"), []byte("img2"), 0o644)

	results, err := fileutils.BatchRename(tempDir, `^IMG_2026_(\d+)\.jpg$`, `Photo_$1.jpg`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("3. Batch Renamed %d files:\n", len(results))
	for _, r := range results {
		fmt.Printf("   %s -> %s\n", filepath.Base(r.OldPath), filepath.Base(r.NewPath))
	}

	// 4. Prune empty directory trees
	emptyTree := filepath.Join(tempDir, "deep", "nested", "vacant", "leaf")
	_ = fileutils.EnsureDir(emptyTree, 0o755)
	pruned, err := fileutils.PruneEmptyDirs(tempDir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("4. Pruned %d empty folders bottom-up recursively.\n", len(pruned))
}
