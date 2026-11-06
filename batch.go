package fileutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// RenameResult records an individual file rename operation during a batch process.
type RenameResult struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

// BatchRename renames all files matching regexPattern in the specified directory using replacement string repl.
// Supports standard regex capture groups ($1, ${name}, etc.).
func BatchRename(dir string, regexPattern string, repl string) ([]RenameResult, error) {
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var results []RenameResult
	for _, entry := range entries {
		name := entry.Name()
		if !re.MatchString(name) {
			continue
		}

		newName := re.ReplaceAllString(name, repl)
		if newName == name {
			continue
		}

		oldPath := filepath.Join(dir, name)
		newPath := filepath.Join(dir, newName)

		if Exists(newPath) {
			return results, fmt.Errorf("%w: %s", ErrDestinationExists, newPath)
		}

		if err := os.Rename(oldPath, newPath); err != nil {
			return results, err
		}

		results = append(results, RenameResult{
			OldPath: oldPath,
			NewPath: newPath,
		})
	}

	return results, nil
}

// PruneEmptyDirs recursively searches and removes all empty subdirectories under root bottom-up.
// Returns the list of pruned directory paths. The root directory itself is not pruned.
func PruneEmptyDirs(root string) ([]string, error) {
	var dirList []string

	cleanRoot := filepath.Clean(root)
	err := filepath.WalkDir(cleanRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && filepath.Clean(path) != cleanRoot {
			dirList = append(dirList, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort deepest directories first so we delete from the bottom up
	sort.Slice(dirList, func(i, j int) bool {
		return len(dirList[i]) > len(dirList[j])
	})

	var pruned []string
	for _, dir := range dirList {
		empty, err := IsEmptyDir(dir)
		if err == nil && empty {
			if err := os.Remove(dir); err == nil {
				pruned = append(pruned, dir)
			}
		}
	}

	return pruned, nil
}
