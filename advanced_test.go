package fileutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMimeAndAdvancedOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-advanced-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Test HumanBytes
	if hb := HumanBytes(500); hb != "500 B" {
		t.Errorf("Expected '500 B', got %s", hb)
	}
	if hb := HumanBytes(1024); hb != "1.00 KB" {
		t.Errorf("Expected '1.00 KB', got %s", hb)
	}
	if hb := HumanBytes(1048576); hb != "1.00 MB" {
		t.Errorf("Expected '1.00 MB', got %s", hb)
	}

	// 2. Test Mime & Text/Binary Detection
	txtPath := filepath.Join(tempDir, "sample.txt")
	_ = os.WriteFile(txtPath, []byte("Hello, this is pure UTF-8 plain text."), 0o644)

	isTxt, err := IsText(txtPath)
	if err != nil || !isTxt {
		t.Errorf("Expected sample.txt to be text, got isTxt=%v, err=%v", isTxt, err)
	}
	isBin, err := IsBinary(txtPath)
	if err != nil || isBin {
		t.Errorf("Expected sample.txt not to be binary")
	}

	binPath := filepath.Join(tempDir, "sample.bin")
	_ = os.WriteFile(binPath, []byte{0x00, 0xFF, 0xFE, 0x12, 0x00}, 0o644)
	isBin, err = IsBinary(binPath)
	if err != nil || !isBin {
		t.Errorf("Expected sample.bin to be binary")
	}

	mimeType, err := DetectMimeType(txtPath)
	if err != nil || mimeType == "" {
		t.Errorf("DetectMimeType failed: %v", err)
	}

	// 3. Test CompareFiles
	txtPath2 := filepath.Join(tempDir, "sample_copy.txt")
	_ = os.WriteFile(txtPath2, []byte("Hello, this is pure UTF-8 plain text."), 0o644)

	identical, err := CompareFiles(txtPath, txtPath2)
	if err != nil || !identical {
		t.Errorf("Expected CompareFiles to return true for identical files")
	}

	diffPath := filepath.Join(tempDir, "sample_diff.txt")
	_ = os.WriteFile(diffPath, []byte("Different content"), 0o644)
	identical, err = CompareFiles(txtPath, diffPath)
	if err != nil || identical {
		t.Errorf("Expected CompareFiles to return false for different files")
	}

	// 4. Test ReplaceInFile
	replacePath := filepath.Join(tempDir, "replace.txt")
	_ = os.WriteFile(replacePath, []byte("Hello cat and cat"), 0o644)
	count, err := ReplaceInFile(replacePath, "cat", "dog")
	if err != nil || count != 2 {
		t.Fatalf("ReplaceInFile failed: count=%d, err=%v", count, err)
	}
	content, _ := os.ReadFile(replacePath)
	if string(content) != "Hello dog and dog" {
		t.Errorf("ReplaceInFile output mismatch: %s", string(content))
	}

	// 5. Test SplitFile and JoinFiles
	largeData := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ") // 36 bytes
	splitSrc := filepath.Join(tempDir, "split_source.dat")
	_ = os.WriteFile(splitSrc, largeData, 0o644)

	chunkDir := filepath.Join(tempDir, "chunks")
	chunks, err := SplitFile(splitSrc, chunkDir, 10) // 10 bytes per chunk -> 4 chunks (10, 10, 10, 6)
	if err != nil {
		t.Fatalf("SplitFile failed: %v", err)
	}
	if len(chunks) != 4 {
		t.Fatalf("Expected 4 chunks, got %d", len(chunks))
	}

	reconstructedFile := filepath.Join(tempDir, "reconstructed.dat")
	if err := JoinFiles(chunks, reconstructedFile); err != nil {
		t.Fatalf("JoinFiles failed: %v", err)
	}

	same, err := CompareFiles(splitSrc, reconstructedFile)
	if err != nil || !same {
		t.Errorf("Reconstructed file does not match original source")
	}
}
