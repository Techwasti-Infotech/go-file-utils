package fileutils

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SearchFilter configures criteria for finding files in a filesystem tree.
type SearchFilter struct {
	// Pattern is a shell glob pattern (e.g. "*.go", "doc*")
	Pattern string
	// Regex matches against filename or path
	Regex *regexp.Regexp
	// Extensions filter files by specific extension (e.g. [".go", ".md"], case-insensitive)
	Extensions []string
	// MinSize filters files >= MinSize (bytes). 0 ignores.
	MinSize int64
	// MaxSize filters files <= MaxSize (bytes). 0 ignores.
	MaxSize int64
	// ModifiedAfter filters files modified strictly after this timestamp.
	ModifiedAfter time.Time
	// ModifiedBefore filters files modified strictly before this timestamp.
	ModifiedBefore time.Time
	// OnlyFiles returns regular files only.
	OnlyFiles bool
	// OnlyDirs returns directories only.
	OnlyDirs bool
	// MaxDepth limits recursion depth (0 means no limit, 1 is top-level only).
	MaxDepth int
	// SkipHidden skips dotfiles and hidden directories (starting with '.').
	SkipHidden bool
}

// Find walks root directory and returns a slice of matching absolute or relative file paths based on filter.
func Find(root string, filter SearchFilter) ([]string, error) {
	var matches []string

	// Normalize extensions
	extMap := make(map[string]struct{}, len(filter.Extensions))
	for _, ext := range filter.Extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[strings.ToLower(ext)] = struct{}{}
	}

	cleanRoot := filepath.Clean(root)
	rootDepth := strings.Count(cleanRoot, string(filepath.Separator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		cleanPath := filepath.Clean(path)

		// Check MaxDepth
		if filter.MaxDepth > 0 && cleanPath != cleanRoot {
			currentDepth := strings.Count(cleanPath, string(filepath.Separator)) - rootDepth
			if currentDepth > filter.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		name := d.Name()

		// Skip hidden
		if filter.SkipHidden && strings.HasPrefix(name, ".") && cleanPath != cleanRoot {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		isDir := d.IsDir()

		// Directory / File exclusive filters
		if filter.OnlyFiles && isDir {
			return nil
		}
		if filter.OnlyDirs && !isDir {
			return nil
		}

		// Don't include the root folder itself if matching
		if cleanPath == cleanRoot && (filter.OnlyFiles || filter.Pattern != "" || filter.Regex != nil) {
			return nil
		}

		// Extension filter
		if len(extMap) > 0 {
			ext := strings.ToLower(filepath.Ext(name))
			if _, ok := extMap[ext]; !ok {
				return nil
			}
		}

		// Glob pattern filter
		if filter.Pattern != "" {
			matched, err := filepath.Match(filter.Pattern, name)
			if err != nil || !matched {
				return nil
			}
		}

		// Regex filter
		if filter.Regex != nil {
			if !filter.Regex.MatchString(name) && !filter.Regex.MatchString(path) {
				return nil
			}
		}

		// Size & Time checks (require file info)
		if filter.MinSize > 0 || filter.MaxSize > 0 || !filter.ModifiedAfter.IsZero() || !filter.ModifiedBefore.IsZero() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if filter.MinSize > 0 && info.Size() < filter.MinSize {
				return nil
			}
			if filter.MaxSize > 0 && info.Size() > filter.MaxSize {
				return nil
			}
			if !filter.ModifiedAfter.IsZero() && !info.ModTime().After(filter.ModifiedAfter) {
				return nil
			}
			if !filter.ModifiedBefore.IsZero() && !info.ModTime().Before(filter.ModifiedBefore) {
				return nil
			}
		}

		matches = append(matches, path)
		return nil
	})

	return matches, err
}

// FindByExtension is a convenience method that returns all files matching a specific extension (e.g. ".json").
func FindByExtension(root, ext string) ([]string, error) {
	return Find(root, SearchFilter{
		Extensions: []string{ext},
		OnlyFiles:  true,
	})
}

// FindByName is a convenience method that searches files by name pattern.
func FindByName(root, pattern string) ([]string, error) {
	return Find(root, SearchFilter{
		Pattern: pattern,
	})
}
