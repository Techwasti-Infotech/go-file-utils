package fileutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-hash-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sampleFile := filepath.Join(tempDir, "hash_sample.txt")
	// "hello world\n"
	content := []byte("hello world\n")
	if err := os.WriteFile(sampleFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. MD5 of "hello world\n" is 6f5902ac237024bdd0c176cb93063dc4
	expectedMD5 := "6f5902ac237024bdd0c176cb93063dc4"
	md5Val, err := FileMD5(sampleFile)
	if err != nil {
		t.Fatalf("FileMD5 failed: %v", err)
	}
	if md5Val != expectedMD5 {
		t.Errorf("FileMD5 mismatch. Got %s, expected %s", md5Val, expectedMD5)
	}

	// 2. SHA256 of "hello world\n" is a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
	expectedSHA256 := "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"
	sha256Val, err := FileSHA256(sampleFile)
	if err != nil {
		t.Fatalf("FileSHA256 failed: %v", err)
	}
	if sha256Val != expectedSHA256 {
		t.Errorf("FileSHA256 mismatch. Got %s, expected %s", sha256Val, expectedSHA256)
	}

	// 3. SHA1 and SHA512
	sha1Val, err := FileHash(sampleFile, HashSHA1)
	if err != nil || len(sha1Val) != 40 {
		t.Errorf("FileHash SHA1 failed: %v, val: %s", err, sha1Val)
	}

	sha512Val, err := FileHash(sampleFile, HashSHA512)
	if err != nil || len(sha512Val) != 128 {
		t.Errorf("FileHash SHA512 failed: %v, val: %s", err, sha512Val)
	}

	// 4. VerifyFileChecksum
	valid, err := VerifyFileChecksum(sampleFile, HashSHA256, expectedSHA256)
	if err != nil || !valid {
		t.Errorf("VerifyFileChecksum expected true, got %v (err: %v)", valid, err)
	}

	invalid, err := VerifyFileChecksum(sampleFile, HashSHA256, "wrong-hash")
	if err != nil || invalid {
		t.Errorf("VerifyFileChecksum expected false for incorrect hash, got %v", invalid)
	}
}
