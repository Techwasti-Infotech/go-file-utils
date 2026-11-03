package fileutils

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestFinder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-finder-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test tree:
	// tempDir/
	//   file1.txt (5 bytes)
	//   file2.log (20 bytes)
	//   .hidden.txt (5 bytes)
	//   nested/
	//     file3.txt (100 bytes)
	//     subnested/
	//       file4.go (50 bytes)
	_ = os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("12345"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "file2.log"), []byte("12345678901234567890"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, ".hidden.txt"), []byte("12345"), 0o644)

	nested := filepath.Join(tempDir, "nested")
	subnested := filepath.Join(nested, "subnested")
	_ = os.MkdirAll(subnested, 0o755)

	_ = os.WriteFile(filepath.Join(nested, "file3.txt"), make([]byte, 100), 0o644)
	_ = os.WriteFile(filepath.Join(subnested, "file4.go"), make([]byte, 50), 0o644)

	// 1. Find by extension
	txtFiles, err := FindByExtension(tempDir, ".txt")
	if err != nil {
		t.Fatalf("FindByExtension failed: %v", err)
	}
	// file1.txt, .hidden.txt, nested/file3.txt
	if len(txtFiles) != 3 {
		t.Errorf("Expected 3 txt files, found %d: %+v", len(txtFiles), txtFiles)
	}

	// 2. Find with SkipHidden
	noHidden, err := Find(tempDir, SearchFilter{
		Extensions: []string{".txt"},
		SkipHidden: true,
		OnlyFiles:  true,
	})
	if err != nil {
		t.Fatalf("Find with SkipHidden failed: %v", err)
	}
	if len(noHidden) != 2 {
		t.Errorf("Expected 2 non-hidden txt files, found %d: %+v", len(noHidden), noHidden)
	}

	// 3. Find with MinSize and MaxSize
	sizeFilter, err := Find(tempDir, SearchFilter{
		MinSize:   10,
		MaxSize:   60,
		OnlyFiles: true,
	})
	if err != nil {
		t.Fatalf("Find with size filter failed: %v", err)
	}
	// file2.log (20 bytes), file4.go (50 bytes)
	if len(sizeFilter) != 2 {
		t.Errorf("Expected 2 files with size between 10 and 60, found %d", len(sizeFilter))
	}

	// 4. Find with Regex
	reg := regexp.MustCompile(`file\d+\.(txt|go)`)
	regexMatches, err := Find(tempDir, SearchFilter{
		Regex:     reg,
		OnlyFiles: true,
	})
	if err != nil {
		t.Fatalf("Find with Regex failed: %v", err)
	}
	// file1.txt, file3.txt, file4.go
	if len(regexMatches) != 3 {
		t.Errorf("Expected 3 regex matches, got %d", len(regexMatches))
	}

	// 5. Find with Time Filter
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	timeMatches, err := Find(tempDir, SearchFilter{
		ModifiedAfter:  past,
		ModifiedBefore: future,
		OnlyFiles:      true,
	})
	if err != nil {
		t.Fatalf("Find with Time filter failed: %v", err)
	}
	if len(timeMatches) != 5 {
		t.Errorf("Expected 5 recently created files, got %d", len(timeMatches))
	}
}
