package fileutils

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// SyncOptions controls directory synchronization behavior.
type SyncOptions struct {
	// Delete removes destination files/folders that do not exist in source.
	Delete bool
	// ByChecksum uses SHA-256 hash comparison instead of modtime/size.
	ByChecksum bool
}

// SyncResult records the actions taken during directory synchronization.
type SyncResult struct {
	Created []string `json:"created"`
	Updated []string `json:"updated"`
	Deleted []string `json:"deleted"`
}

// SyncDir synchronizes the contents of srcDir to dstDir.
// It creates missing files/folders, updates modified files, and optionally deletes extra files in dst.
func SyncDir(srcDir, dstDir string, opts SyncOptions) (*SyncResult, error) {
	if !IsDir(srcDir) {
		return nil, ErrNotADirectory
	}

	if err := EnsureDir(dstDir, DefaultDirPerm); err != nil {
		return nil, err
	}

	result := &SyncResult{}
	srcMap := make(map[string]fs.DirEntry)

	cleanSrc := filepath.Clean(srcDir)
	cleanDst := filepath.Clean(dstDir)

	// 1. Walk source directory and sync to destination
	err := filepath.WalkDir(cleanSrc, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cleanSrc {
			return nil
		}

		rel, err := filepath.Rel(cleanSrc, path)
		if err != nil {
			return err
		}
		srcMap[rel] = d

		dstPath := filepath.Join(cleanDst, rel)

		if d.IsDir() {
			if !Exists(dstPath) {
				if err := EnsureDir(dstPath, DefaultDirPerm); err != nil {
					return err
				}
				result.Created = append(result.Created, dstPath)
			}
			return nil
		}

		// File synchronization check
		needsCopy := false
		isNew := false

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			if os.IsNotExist(err) {
				needsCopy = true
				isNew = true
			} else {
				return err
			}
		} else {
			srcInfo, err := d.Info()
			if err != nil {
				return err
			}

			if opts.ByChecksum {
				srcHash, err := FileSHA256(path)
				if err != nil {
					return err
				}
				dstHash, err := FileSHA256(dstPath)
				if err != nil {
					return err
				}
				if srcHash != dstHash {
					needsCopy = true
				}
			} else {
				// Fast size and modtime comparison
				if srcInfo.Size() != dstInfo.Size() || srcInfo.ModTime().After(dstInfo.ModTime().Add(time.Second)) {
					needsCopy = true
				}
			}
		}

		if needsCopy {
			if err := CopyFile(path, dstPath, CopyFileOptions{Overwrite: true, PreserveTimestamps: true}); err != nil {
				return err
			}
			if isNew {
				result.Created = append(result.Created, dstPath)
			} else {
				result.Updated = append(result.Updated, dstPath)
			}
		}

		return nil
	})
	if err != nil {
		return result, err
	}

	// 2. If Delete option is enabled, prune files and folders in dst that don't exist in src
	if opts.Delete {
		var toDelete []string

		err := filepath.WalkDir(cleanDst, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == cleanDst {
				return nil
			}

			rel, err := filepath.Rel(cleanDst, path)
			if err != nil {
				return err
			}

			if _, exists := srcMap[rel]; !exists {
				toDelete = append(toDelete, path)
				if d.IsDir() {
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err != nil {
			return result, err
		}

		for _, p := range toDelete {
			if err := os.RemoveAll(p); err == nil {
				result.Deleted = append(result.Deleted, p)
			}
		}
	}

	return result, nil
}
