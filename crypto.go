package fileutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// EncryptFile encrypts the file at srcPath using AES-256-GCM authenticated encryption
// derived from passphrase and writes the result to dstPath.
func EncryptFile(srcPath, dstPath, passphrase string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return WriteFileAtomic(dstPath, ciphertext, 0o600)
}

// DecryptFile decrypts an AES-256-GCM encrypted file produced by EncryptFile.
// Returns an error if the passphrase is wrong or the ciphertext has been altered.
func DecryptFile(srcPath, dstPath, passphrase string) error {
	ciphertext, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("malformed ciphertext: length smaller than nonce size")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (invalid passphrase or corrupted data): %w", err)
	}

	return WriteFileAtomic(dstPath, plaintext, 0o644)
}
