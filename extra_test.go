package fileutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type sampleConfig struct {
	Name    string   `json:"name"`
	Version int      `json:"version"`
	Tags    []string `json:"tags"`
}

func TestLinesCryptoStructured(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-lines-crypto-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Test HeadLines, TailLines, CountLines
	linesFilePath := filepath.Join(tempDir, "multiline.txt")
	var buf bytes.Buffer
	for i := 1; i <= 100; i++ {
		buf.WriteString(fmt.Sprintf("Line %d\n", i))
	}
	if err := os.WriteFile(linesFilePath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	count, err := CountLines(linesFilePath)
	if err != nil || count != 100 {
		t.Errorf("CountLines expected 100, got %d (err: %v)", count, err)
	}

	head, err := HeadLines(linesFilePath, 3)
	if err != nil || len(head) != 3 || head[0] != "Line 1" || head[2] != "Line 3" {
		t.Errorf("HeadLines failed: %+v", head)
	}

	tail, err := TailLines(linesFilePath, 3)
	if err != nil || len(tail) != 3 || tail[0] != "Line 98" || tail[2] != "Line 100" {
		t.Errorf("TailLines failed: %+v", tail)
	}

	// 2. Test EncryptFile & DecryptFile (AES-256-GCM)
	plainTextFile := filepath.Join(tempDir, "secret.txt")
	secretContent := []byte("Top secret credentials and private token!")
	_ = os.WriteFile(plainTextFile, secretContent, 0o644)

	encryptedFile := filepath.Join(tempDir, "secret.enc")
	passphrase := "super-secure-master-password-123"

	if err := EncryptFile(plainTextFile, encryptedFile, passphrase); err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Encrypted file should not match plaintext
	encBytes, _ := os.ReadFile(encryptedFile)
	if bytes.Equal(encBytes, secretContent) {
		t.Fatalf("Encrypted file should not match plain text")
	}

	decryptedFile := filepath.Join(tempDir, "secret_decrypted.txt")
	if err := DecryptFile(encryptedFile, decryptedFile, passphrase); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	decBytes, _ := os.ReadFile(decryptedFile)
	if !bytes.Equal(decBytes, secretContent) {
		t.Errorf("Decrypted content mismatch. Got: %s", string(decBytes))
	}

	// Test DecryptFile with invalid passphrase
	badDecFile := filepath.Join(tempDir, "secret_bad.txt")
	if err := DecryptFile(encryptedFile, badDecFile, "wrong-password"); err == nil {
		t.Errorf("Expected DecryptFile to fail with incorrect passphrase")
	}

	// 3. Test ReadJSON & WriteJSON
	cfg := sampleConfig{
		Name:    "go-file-utils",
		Version: 2,
		Tags:    []string{"file", "crypto", "io"},
	}
	jsonPath := filepath.Join(tempDir, "config.json")
	if err := WriteJSON(jsonPath, cfg, 0o644); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	var loadedCfg sampleConfig
	if err := ReadJSON(jsonPath, &loadedCfg); err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}
	if loadedCfg.Name != cfg.Name || loadedCfg.Version != cfg.Version || len(loadedCfg.Tags) != 3 {
		t.Errorf("Loaded JSON mismatch: %+v", loadedCfg)
	}

	// 4. Test ReadCSV & WriteCSV
	csvRecords := [][]string{
		{"ID", "Username", "Role"},
		{"101", "alice", "admin"},
		{"102", "bob", "developer"},
	}
	csvPath := filepath.Join(tempDir, "users.csv")
	if err := WriteCSV(csvPath, csvRecords, 0o644); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	loadedCSV, err := ReadCSV(csvPath)
	if err != nil || len(loadedCSV) != 3 || loadedCSV[1][1] != "alice" {
		t.Errorf("ReadCSV failed: %+v", loadedCSV)
	}
}
