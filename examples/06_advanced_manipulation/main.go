package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 06: Advanced Manipulation (MIME, Split/Join, Replace, Head/Tail) ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-06-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. MIME and Binary/Text detection
	txtPath := filepath.Join(tempDir, "script.py")
	_ = os.WriteFile(txtPath, []byte("print('Hello Python')\n"), 0o644)

	mimeType, _ := fileutils.DetectMimeType(txtPath)
	isTxt, _ := fileutils.IsText(txtPath)
	isBin, _ := fileutils.IsBinary(txtPath)
	fmt.Printf("1. MIME: %s | IsText: %v | IsBinary: %v\n", mimeType, isTxt, isBin)

	// 2. Atomic Find-and-Replace in file
	count, err := fileutils.ReplaceInFile(txtPath, "Python", "Go File Utils")
	if err != nil {
		panic(err)
	}
	content, _ := os.ReadFile(txtPath)
	fmt.Printf("2. Replaced %d occurrences -> %s", count, string(content))

	// 3. HeadLines, TailLines, and CountLines
	multilineFile := filepath.Join(tempDir, "numbers.txt")
	var data []byte
	for i := 1; i <= 50; i++ {
		data = append(data, []byte(fmt.Sprintf("Row %02d\n", i))...)
	}
	_ = os.WriteFile(multilineFile, data, 0o644)

	totalLines, _ := fileutils.CountLines(multilineFile)
	head, _ := fileutils.HeadLines(multilineFile, 3)
	tail, _ := fileutils.TailLines(multilineFile, 3)
	fmt.Printf("3. Total Lines: %d\n   Head (first 3): %v\n   Tail (last 3):  %v\n", totalLines, head, tail)

	// 4. Split large files into chunks and join back
	chunksDir := filepath.Join(tempDir, "chunks")
	chunks, err := fileutils.SplitFile(multilineFile, chunksDir, 100) // 100 byte chunks
	if err != nil {
		panic(err)
	}
	fmt.Printf("4. Split %d bytes file into %d chunks:\n", len(data), len(chunks))
	for _, c := range chunks {
		fmt.Printf("   - %s\n", filepath.Base(c))
	}

	reconstructedFile := filepath.Join(tempDir, "numbers_reconstructed.txt")
	if err := fileutils.JoinFiles(chunks, reconstructedFile); err != nil {
		panic(err)
	}
	identical, _ := fileutils.CompareFiles(multilineFile, reconstructedFile)
	fmt.Printf("   Rejoined successfully! Matches original: %v\n", identical)
}
