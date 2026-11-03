package fileutils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// sanitizeArchivePath verifies that the archive file path does not attempt path traversal (Zip Slip vulnerability).
// It returns the clean, absolute destination path inside targetDir.
func sanitizeArchivePath(targetDir, archivePath string) (string, error) {
	cleanTarget := filepath.Clean(targetDir)
	cleanDest := filepath.Clean(filepath.Join(cleanTarget, archivePath))

	// Ensure cleanDest starts with cleanTarget + separator (or equals cleanTarget)
	if cleanDest != cleanTarget && !strings.HasPrefix(cleanDest, cleanTarget+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: path '%s' escapes destination directory '%s'", ErrInvalidPath, archivePath, targetDir)
	}
	return cleanDest, nil
}

// Zip creates a ZIP archive at zipFilePath containing all files and folders in srcPath.
func Zip(srcPath, zipFilePath string) error {
	if !Exists(srcPath) {
		return fmt.Errorf("%w: %s", ErrSourceNotFound, srcPath)
	}

	if err := EnsureParentDir(zipFilePath, DefaultDirPerm); err != nil {
		return err
	}

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	srcClean := filepath.Clean(srcPath)
	baseDir := filepath.Dir(srcClean)

	return filepath.Walk(srcClean, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		// In zip archives, slashes must be forward slashes
		headerName := filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = headerName
		header.Method = zip.Deflate

		if info.IsDir() {
			header.Name += "/"
			_, err = archive.CreateHeader(header)
			return err
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// Unzip extracts a ZIP archive to destDir, guarding against Zip Slip vulnerabilities.
func Unzip(zipFilePath, destDir string) error {
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := EnsureDir(destDir, DefaultDirPerm); err != nil {
		return err
	}

	for _, file := range reader.File {
		destPath, err := sanitizeArchivePath(destDir, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := EnsureDir(destPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := EnsureParentDir(destPath, DefaultDirPerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

// TarGz creates a .tar.gz archive at archivePath containing srcPath.
func TarGz(srcPath, archivePath string) error {
	if !Exists(srcPath) {
		return fmt.Errorf("%w: %s", ErrSourceNotFound, srcPath)
	}

	if err := EnsureParentDir(archivePath, DefaultDirPerm); err != nil {
		return err
	}

	outFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcClean := filepath.Clean(srcPath)
	baseDir := filepath.Dir(srcClean)

	return filepath.Walk(srcClean, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}

// UntarGz extracts a .tar.gz archive into destDir with Zip Slip protection.
func UntarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	if err := EnsureDir(destDir, DefaultDirPerm); err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		destPath, err := sanitizeArchivePath(destDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := EnsureDir(destPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := EnsureParentDir(destPath, DefaultDirPerm); err != nil {
				return err
			}
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
