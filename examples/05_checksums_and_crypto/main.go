package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

func main() {
	fmt.Println("=== Example 05: Checksums & AES-256-GCM Encryption ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-05-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	secretFile := filepath.Join(tempDir, "credentials.txt")
	_ = os.WriteFile(secretFile, []byte("API_SECRET_KEY=99283741872364"), 0o644)

	// 1. Streaming SHA-256 and MD5 hashes
	sha256Hex, err := fileutils.FileSHA256(secretFile)
	if err != nil {
		panic(err)
	}
	md5Hex, _ := fileutils.FileMD5(secretFile)
	fmt.Printf("1. File Checksums:\n   SHA-256: %s\n   MD5:     %s\n", sha256Hex, md5Hex)

	// 2. Integrity Verification
	match, _ := fileutils.VerifyFileChecksum(secretFile, fileutils.HashSHA256, sha256Hex)
	fmt.Printf("2. Integrity check matches expected: %v\n", match)

	// 3. Encrypt file with AES-256-GCM
	encryptedFile := filepath.Join(tempDir, "credentials.enc")
	passphrase := "master-vault-password-2026"

	if err := fileutils.EncryptFile(secretFile, encryptedFile, passphrase); err != nil {
		panic(err)
	}
	fmt.Printf("3. Encrypted to: %s\n", encryptedFile)

	// 4. Decrypt file and verify content
	decryptedFile := filepath.Join(tempDir, "credentials_decrypted.txt")
	if err := fileutils.DecryptFile(encryptedFile, decryptedFile, passphrase); err != nil {
		panic(err)
	}
	fmt.Printf("4. Decrypted to: %s\n", decryptedFile)

	// 5. Compare files byte-for-byte
	identical, _ := fileutils.CompareFiles(secretFile, decryptedFile)
	fmt.Printf("5. Original matches decrypted file byte-for-byte: %v\n", identical)
}
