package fileutils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// HashAlgorithm specifies supported cryptographic hashing functions.
type HashAlgorithm string

const (
	HashMD5    HashAlgorithm = "md5"
	HashSHA1   HashAlgorithm = "sha1"
	HashSHA256 HashAlgorithm = "sha256"
	HashSHA512 HashAlgorithm = "sha512"
)

// FileHash computes the cryptographic checksum of a file given the specified algorithm.
// Computations are streamed in chunks, making it safe for arbitrarily large gigabyte files without high memory usage.
func FileHash(path string, algo HashAlgorithm) (string, error) {
	var h hash.Hash
	switch strings.ToLower(string(algo)) {
	case string(HashMD5):
		h = md5.New()
	case string(HashSHA1):
		h = sha1.New()
	case string(HashSHA256):
		h = sha256.New()
	case string(HashSHA512):
		h = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 64 KB streaming buffer
	buf := make([]byte, 64*1024)
	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileMD5 returns the hexadecimal MD5 checksum of the file.
func FileMD5(path string) (string, error) {
	return FileHash(path, HashMD5)
}

// FileSHA256 returns the hexadecimal SHA-256 checksum of the file.
func FileSHA256(path string) (string, error) {
	return FileHash(path, HashSHA256)
}

// VerifyFileChecksum verifies if the file at path matches the expected hexadecimal checksum.
func VerifyFileChecksum(path string, algo HashAlgorithm, expectedChecksum string) (bool, error) {
	actual, err := FileHash(path, algo)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(actual, strings.TrimSpace(expectedChecksum)), nil
}

