package fileutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// HumanBytes converts an integer byte count into a human-readable string (e.g. "1.50 MB", "500 KB").
func HumanBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// DiffFile represents line-level differences between two files.
type DiffResult struct {
	Identical bool
	OnlyInA   []string
	OnlyInB   []string
}

// CompareFiles performs a fast byte-by-byte comparison of two files.
// Returns true if both files exist and are identical in content.
func CompareFiles(pathA, pathB string) (bool, error) {
	infoA, err := os.Stat(pathA)
	if err != nil {
		return false, err
	}
	infoB, err := os.Stat(pathB)
	if err != nil {
		return false, err
	}

	// Size mismatch means immediate non-identical
	if infoA.Size() != infoB.Size() {
		return false, nil
	}

	fA, err := os.Open(pathA)
	if err != nil {
		return false, err
	}
	defer fA.Close()

	fB, err := os.Open(pathB)
	if err != nil {
		return false, err
	}
	defer fB.Close()

	bufA := make([]byte, 64*1024)
	bufB := make([]byte, 64*1024)

	for {
		nA, errA := fA.Read(bufA)
		nB, errB := fB.Read(bufB)

		if nA != nB || !bytes.Equal(bufA[:nA], bufB[:nB]) {
			return false, nil
		}

		if errA != nil || errB != nil {
			if errA == errB {
				break
			}
			return false, nil
		}
	}

	return true, nil
}

// ReplaceInFile replaces occurrences of oldStr with newStr across a file safely and atomically.
func ReplaceInFile(path string, oldStr, newStr string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	count := bytes.Count(data, []byte(oldStr))
	if count == 0 {
		return 0, nil
	}

	replaced := bytes.ReplaceAll(data, []byte(oldStr), []byte(newStr))

	info, err := os.Stat(path)
	perm := DefaultFilePerm
	if err == nil {
		perm = info.Mode().Perm()
	}

	if err := WriteFileAtomic(path, replaced, perm); err != nil {
		return 0, err
	}

	return count, nil
}

// SplitFile splits a large file into smaller chunks of chunkSize bytes each inside outDir.
// It returns the list of generated chunk file paths.
func SplitFile(srcPath, outDir string, chunkSize int64) ([]string, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be greater than 0")
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := EnsureDir(outDir, DefaultDirPerm); err != nil {
		return nil, err
	}

	baseName := filepath.Base(srcPath)
	var chunkPaths []string
	buf := make([]byte, 64*1024)
	chunkIndex := 0

	for {
		chunkFileName := filepath.Join(outDir, fmt.Sprintf("%s.part%04d", baseName, chunkIndex))
		chunkFile, err := os.Create(chunkFileName)
		if err != nil {
			return chunkPaths, err
		}

		var bytesInCurrentChunk int64
		var hitEOF bool

		for bytesInCurrentChunk < chunkSize {
			toRead := int64(len(buf))
			if remaining := chunkSize - bytesInCurrentChunk; remaining < toRead {
				toRead = remaining
			}

			n, rErr := f.Read(buf[:toRead])
			if n > 0 {
				if _, wErr := chunkFile.Write(buf[:n]); wErr != nil {
					chunkFile.Close()
					return chunkPaths, wErr
				}
				bytesInCurrentChunk += int64(n)
			}
			if rErr != nil {
				hitEOF = true
				break
			}
		}

		chunkFile.Close()

		if bytesInCurrentChunk == 0 {
			_ = os.Remove(chunkFileName)
			break
		}

		chunkPaths = append(chunkPaths, chunkFileName)
		chunkIndex++

		if hitEOF {
			break
		}
	}

	return chunkPaths, nil
}

// JoinFiles reconstructs a single file at dstPath by combining the provided chunk files in order.
func JoinFiles(chunkPaths []string, dstPath string) error {
	if err := EnsureParentDir(dstPath, DefaultDirPerm); err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 64*1024)
	for _, chunkPath := range chunkPaths {
		cf, err := os.Open(chunkPath)
		if err != nil {
			return err
		}
		for {
			n, rErr := cf.Read(buf)
			if n > 0 {
				if _, wErr := dstFile.Write(buf[:n]); wErr != nil {
					cf.Close()
					return wErr
				}
			}
			if rErr != nil {
				break
			}
		}
		cf.Close()
	}

	return dstFile.Sync()
}
