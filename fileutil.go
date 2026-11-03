// Package fileutils provides a comprehensive, production-ready, zero-dependency file and filesystem
// utility library for Go applications. It covers basic and advanced filesystem operations, atomic writes,
// recursive traversal, chunked/streamed I/O, archiving, checksums, and temporary file lifecycle management.
package fileutils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrSourceNotFound is returned when the source file or directory does not exist.
	ErrSourceNotFound = errors.New("fileutils: source path does not exist")
	// ErrDestinationExists is returned when target destination already exists and overwrite is false.
	ErrDestinationExists = errors.New("fileutils: destination already exists")
	// ErrInvalidPath indicates an untrusted, unsafe, or malformed path (e.g. Zip Slip vulnerability).
	ErrInvalidPath = errors.New("fileutils: invalid or unsafe path")
	// ErrNotADirectory is returned when a path is expected to be a directory.
	ErrNotADirectory = errors.New("fileutils: path is not a directory")
	// ErrNotAFile is returned when a path is expected to be a regular file.
	ErrNotAFile = errors.New("fileutils: path is not a regular file")
)

// DefaultDirPerm represents standard directory permissions (rwxr-xr-x).
const DefaultDirPerm os.FileMode = 0o755

// DefaultFilePerm represents standard file permissions (rw-r--r--).
const DefaultFilePerm os.FileMode = 0o644

// Exists returns true if the specified file or directory path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// IsFile returns true if path exists and is a regular file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// IsDir returns true if path exists and is a directory.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsSymlink returns true if path is a symbolic link.
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// IsEmptyDir checks if a directory exists and has no files or subdirectories.
func IsEmptyDir(path string) (bool, error) {
	if !IsDir(path) {
		return false, ErrNotADirectory
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	names, err := f.Readdirnames(1)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	return len(names) == 0, nil
}

// EnsureDir checks whether the directory path exists, and creates it (including parents) if it does not.
func EnsureDir(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultDirPerm
	}
	return os.MkdirAll(path, perm)
}

// EnsureParentDir ensures that the parent directory of the given file path exists.
func EnsureParentDir(filePath string, perm os.FileMode) error {
	parent := filepath.Dir(filePath)
	return EnsureDir(parent, perm)
}

// Touch creates an empty file if it doesn't exist, or updates its access and modification times to now.
func Touch(path string) error {
	if !Exists(path) {
		if err := EnsureParentDir(path, DefaultDirPerm); err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, DefaultFilePerm)
		if err != nil {
			return err
		}
		return f.Close()
	}
	now := time.Now()
	return os.Chtimes(path, now, now)
}

// RemoveIfExists deletes the file or directory at path if it exists. Returns nil if file doesn't exist.
func RemoveIfExists(path string) error {
	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// EmptyDir removes all contents inside the given directory without removing the directory itself.
func EmptyDir(dirPath string) error {
	if !IsDir(dirPath) {
		return ErrNotADirectory
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		targetPath := filepath.Join(dirPath, entry.Name())
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	return nil
}

// CopyFileOptions configures file copying behavior.
type CopyFileOptions struct {
	Overwrite          bool
	PreserveTimestamps bool
	BufferSize         int
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string, opts ...CopyFileOptions) error {
	var opt CopyFileOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.BufferSize <= 0 {
		opt.BufferSize = 64 * 1024 // 64 KB buffer default
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrSourceNotFound, src)
		}
		return err
	}
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("%w: %s", ErrNotAFile, src)
	}

	if Exists(dst) {
		if !opt.Overwrite {
			return fmt.Errorf("%w: %s", ErrDestinationExists, dst)
		}
	}

	if err := EnsureParentDir(dst, DefaultDirPerm); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, opt.BufferSize)
	if _, err := io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		return err
	}

	if err := dstFile.Sync(); err != nil {
		return err
	}

	if opt.PreserveTimestamps {
		_ = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}

	return nil
}

// CopyDir recursively copies directory src to dst.
func CopyDir(src, dst string, overwrite bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return ErrNotADirectory
	}

	if err := EnsureDir(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath, overwrite); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath, CopyFileOptions{Overwrite: overwrite, PreserveTimestamps: true}); err != nil {
				return err
			}
		}
	}
	return nil
}

// Move moves or renames a file or directory from src to dst.
func Move(src, dst string) error {
	if !Exists(src) {
		return fmt.Errorf("%w: %s", ErrSourceNotFound, src)
	}
	if err := EnsureParentDir(dst, DefaultDirPerm); err != nil {
		return err
	}
	// Try fast atomic rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Cross-device fallback
	if IsDir(src) {
		if err := CopyDir(src, dst, true); err != nil {
			return err
		}
		return os.RemoveAll(src)
	}

	if err := CopyFile(src, dst, CopyFileOptions{Overwrite: true, PreserveTimestamps: true}); err != nil {
		return err
	}
	return os.Remove(src)
}

// FileInfo holds clean, cross-platform file metadata.
type FileInfo struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Size      int64       `json:"size"`
	Mode      os.FileMode `json:"mode"`
	ModTime   time.Time   `json:"mod_time"`
	IsDir     bool        `json:"is_dir"`
	IsSymlink bool        `json:"is_symlink"`
	Extension string      `json:"extension"`
}

// GetInfo retrieves structured metadata for a file or directory.
func GetInfo(path string) (*FileInfo, error) {
	lst, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	isSym := lst.Mode()&os.ModeSymlink != 0
	info := lst
	if isSym {
		targetInfo, err := os.Stat(path)
		if err == nil {
			info = targetInfo
		}
	}

	absPath, _ := filepath.Abs(path)

	return &FileInfo{
		Name:      lst.Name(),
		Path:      absPath,
		Size:      info.Size(),
		Mode:      lst.Mode(),
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		IsSymlink: isSym,
		Extension: strings.ToLower(filepath.Ext(path)),
	}, nil
}

// GetDirSize calculates the total byte size of all files in a directory recursively.
func GetDirSize(dirPath string) (int64, error) {
	if !IsDir(dirPath) {
		return 0, ErrNotADirectory
	}
	var totalSize int64
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}
