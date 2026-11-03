package fileutils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-archive-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Build sample tree to zip and tar
	srcDir := filepath.Join(tempDir, "sample_data")
	if err := EnsureDir(filepath.Join(srcDir, "sub"), DefaultDirPerm); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("Zip Content 1"), DefaultFilePerm)
	_ = os.WriteFile(filepath.Join(srcDir, "sub", "file2.txt"), []byte("Nested Tar Content 2"), DefaultFilePerm)

	// 1. Test Zip and Unzip
	zipTarget := filepath.Join(tempDir, "archive.zip")
	if err := Zip(srcDir, zipTarget); err != nil {
		t.Fatalf("Zip failed: %v", err)
	}
	if !Exists(zipTarget) {
		t.Fatalf("Zip file was not created")
	}

	unzipDest := filepath.Join(tempDir, "unzipped")
	if err := Unzip(zipTarget, unzipDest); err != nil {
		t.Fatalf("Unzip failed: %v", err)
	}

	content1, err := os.ReadFile(filepath.Join(unzipDest, "sample_data", "file1.txt"))
	if err != nil || string(content1) != "Zip Content 1" {
		t.Errorf("Unzip content 1 mismatch: %v, got %s", err, string(content1))
	}

	// 2. Test TarGz and UntarGz
	tarTarget := filepath.Join(tempDir, "archive.tar.gz")
	if err := TarGz(srcDir, tarTarget); err != nil {
		t.Fatalf("TarGz failed: %v", err)
	}
	if !Exists(tarTarget) {
		t.Fatalf("Tar.gz file was not created")
	}

	untarDest := filepath.Join(tempDir, "untarred")
	if err := UntarGz(tarTarget, untarDest); err != nil {
		t.Fatalf("UntarGz failed: %v", err)
	}

	content2, err := os.ReadFile(filepath.Join(untarDest, "sample_data", "sub", "file2.txt"))
	if err != nil || string(content2) != "Nested Tar Content 2" {
		t.Errorf("UntarGz content 2 mismatch: %v, got %s", err, string(content2))
	}

	// 3. Test Zip Slip Security Guard
	maliciousZipPath := filepath.Join(tempDir, "malicious.zip")
	f, err := os.Create(maliciousZipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// Inject relative path escape "../evil.txt"
	header := &zip.FileHeader{
		Name:   "../../evil.txt",
		Method: zip.Store,
	}
	writer, _ := zw.CreateHeader(header)
	_, _ = writer.Write([]byte("malicious payload"))
	_ = zw.Close()
	_ = f.Close()

	maliciousExtractDir := filepath.Join(tempDir, "safe_target")
	err = Unzip(maliciousZipPath, maliciousExtractDir)
	if err == nil {
		t.Fatalf("Expected Unzip to reject path traversal (Zip Slip), but it succeeded")
	}
}

