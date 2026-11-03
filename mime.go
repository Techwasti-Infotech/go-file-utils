package fileutils

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// DetectMimeType sniffs the MIME type of a file based on initial bytes and fallback to its extension.
func DetectMimeType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Sniff first 512 bytes per HTTP/MIME spec
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	// If detected as generic octet-stream or plain text, check extension map for more specific mime types
	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		ext := strings.ToLower(filepath.Ext(path))
		if extMime := mime.TypeByExtension(ext); extMime != "" {
			return extMime, nil
		}
	}

	return contentType, nil
}

// IsBinary detects whether a file contains non-printable binary content.
// It checks up to the first 8000 bytes for null bytes or invalid UTF-8 sequences.
func IsBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 8000)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	chunk := buf[:n]

	// Look for null bytes
	if bytes.IndexByte(chunk, 0) != -1 {
		return true, nil
	}

	// Look for invalid UTF-8 sequences
	if !utf8.Valid(chunk) {
		return true, nil
	}

	return false, nil
}

// IsText returns true if the file is a valid text file (not binary).
func IsText(path string) (bool, error) {
	isBin, err := IsBinary(path)
	if err != nil {
		return false, err
	}
	return !isBin, nil
}

