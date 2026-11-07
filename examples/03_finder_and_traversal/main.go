package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 03: File Finder & Directory Traversal ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-03-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock folder structure
	_ = os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# Title"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, ".secret.env"), []byte("KEY=VAL"), 0o644)

	srcDir := filepath.Join(tempDir, "pkg", "utils")
	_ = fileutils.EnsureDir(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "math.go"), make([]byte, 2048), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "math_test.go"), make([]byte, 4096), 0o644)

	// 1. Find by extension
	goFiles, err := fileutils.FindByExtension(tempDir, ".go")
	if err != nil {
		panic(err)
	}
	fmt.Printf("1. Found %d .go files:\n", len(goFiles))
	for _, f := range goFiles {
		rel, _ := filepath.Rel(tempDir, f)
		fmt.Printf("   - %s\n", rel)
	}

	// 2. Find with advanced SearchFilter (MinSize, SkipHidden, Regex)
	filter := fileutils.SearchFilter{
		Regex:      regexp.MustCompile(`.*_test\.go$`),
		MinSize:    1024,
		SkipHidden: true,
		OnlyFiles:  true,
	}
	testFiles, err := fileutils.Find(tempDir, filter)
	if err != nil {
		panic(err)
	}
	fmt.Printf("2. Found %d test files >= 1KB:\n", len(testFiles))
	for _, f := range testFiles {
		rel, _ := filepath.Rel(tempDir, f)
		fmt.Printf("   - %s\n", rel)
	}

	// 3. Find files modified in the last 10 minutes
	recentFiles, _ := fileutils.Find(tempDir, fileutils.SearchFilter{
		ModifiedAfter: time.Now().Add(-10 * time.Minute),
		OnlyFiles:     true,
		SkipHidden:    true,
	})
	fmt.Printf("3. Found %d non-hidden files modified recently.\n", len(recentFiles))
}
