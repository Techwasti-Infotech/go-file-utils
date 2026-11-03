package fileutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCoreFileOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-fileutils-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "sample.txt")

	// 1. Test Exists, IsFile, IsDir on nonexistent
	if Exists(testFilePath) {
		t.Errorf("Expected %s to not exist", testFilePath)
	}
	if IsFile(testFilePath) {
		t.Errorf("Expected %s to not be a file", testFilePath)
	}

	// 2. Test Touch
	if err := Touch(testFilePath); err != nil {
		t.Fatalf("Touch failed: %v", err)
	}
	if !Exists(testFilePath) {
		t.Errorf("Expected %s to exist after Touch", testFilePath)
	}
	if !IsFile(testFilePath) {
		t.Errorf("Expected %s to be recognized as regular file", testFilePath)
	}

	// Touch existing file to update mod time
	time.Sleep(10 * time.Millisecond)
	if err := Touch(testFilePath); err != nil {
		t.Errorf("Second Touch failed: %v", err)
	}

	// 3. Test EnsureDir and EnsureParentDir
	subDir := filepath.Join(tempDir, "a", "b", "c")
	if err := EnsureDir(subDir, DefaultDirPerm); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	if !IsDir(subDir) {
		t.Errorf("Expected %s to be directory", subDir)
	}

	deepFile := filepath.Join(tempDir, "x", "y", "nested.txt")
	if err := EnsureParentDir(deepFile, DefaultDirPerm); err != nil {
		t.Fatalf("EnsureParentDir failed: %v", err)
	}
	if !IsDir(filepath.Dir(deepFile)) {
		t.Errorf("Expected parent of nested.txt to exist")
	}

	// 4. Test CopyFile
	content := []byte("Hello, Go File Utils!")
	if err := os.WriteFile(testFilePath, content, DefaultFilePerm); err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	destFile := filepath.Join(tempDir, "copy.txt")
	if err := CopyFile(testFilePath, destFile, CopyFileOptions{Overwrite: true, PreserveTimestamps: true}); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}
	readBytes, err := os.ReadFile(destFile)
	if err != nil || string(readBytes) != string(content) {
		t.Errorf("CopyFile content mismatch. Got: %s, Want: %s", string(readBytes), string(content))
	}

	// Overwrite guard
	if err := CopyFile(testFilePath, destFile, CopyFileOptions{Overwrite: false}); err == nil {
		t.Errorf("Expected error when copying to existing file without Overwrite: true")
	}

	// 5. Test CopyDir
	srcDir := filepath.Join(tempDir, "src_dir")
	if err := EnsureDir(srcDir, DefaultDirPerm); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("data1"), DefaultFilePerm)
	_ = EnsureDir(filepath.Join(srcDir, "nested"), DefaultDirPerm)
	_ = os.WriteFile(filepath.Join(srcDir, "nested", "file2.txt"), []byte("data2"), DefaultFilePerm)

	destDir := filepath.Join(tempDir, "dst_dir")
	if err := CopyDir(srcDir, destDir, true); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}
	if !Exists(filepath.Join(destDir, "file1.txt")) || !Exists(filepath.Join(destDir, "nested", "file2.txt")) {
		t.Errorf("CopyDir files missing in destination")
	}

	// 6. Test Move
	movedDir := filepath.Join(tempDir, "moved_dir")
	if err := Move(destDir, movedDir); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if Exists(destDir) || !Exists(filepath.Join(movedDir, "file1.txt")) {
		t.Errorf("Move failed directory state verification")
	}

	// 7. Test EmptyDir & IsEmptyDir
	emptyTestDir := filepath.Join(tempDir, "empty_test")
	_ = EnsureDir(emptyTestDir, DefaultDirPerm)
	isEmpty, err := IsEmptyDir(emptyTestDir)
	if err != nil || !isEmpty {
		t.Errorf("Expected emptyTestDir to be empty: %v", err)
	}
	_ = os.WriteFile(filepath.Join(emptyTestDir, "test.txt"), []byte("test"), DefaultFilePerm)
	isEmpty, _ = IsEmptyDir(emptyTestDir)
	if isEmpty {
		t.Errorf("Expected emptyTestDir not to be empty")
	}
	if err := EmptyDir(emptyTestDir); err != nil {
		t.Fatalf("EmptyDir failed: %v", err)
	}
	isEmpty, _ = IsEmptyDir(emptyTestDir)
	if !isEmpty {
		t.Errorf("Expected emptyTestDir to be empty after EmptyDir")
	}

	// 8. Test GetInfo & GetDirSize
	info, err := GetInfo(testFilePath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}
	if info.Size != int64(len(content)) || info.Extension != ".txt" || info.IsDir {
		t.Errorf("GetInfo mismatch: %+v", info)
	}

	totalSize, err := GetDirSize(srcDir)
	if err != nil || totalSize != 10 { // "data1" (5) + "data2" (5)
		t.Errorf("GetDirSize got %d, expected 10, err: %v", totalSize, err)
	}

	// 9. Test RemoveIfExists
	if err := RemoveIfExists(testFilePath); err != nil {
		t.Fatalf("RemoveIfExists failed: %v", err)
	}
	if Exists(testFilePath) {
		t.Errorf("File should have been removed")
	}
	// Calling RemoveIfExists on nonexistent should succeed without error
	if err := RemoveIfExists(testFilePath); err != nil {
		t.Errorf("RemoveIfExists on absent path failed: %v", err)
	}
}

