package fileutils

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// TailLines reads the last n lines from a file efficiently without loading the whole file into RAM.
// It scans backwards from the end in chunks.
func TailLines(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	if fileSize == 0 {
		return nil, nil
	}

	chunkSize := int64(4096)
	offset := fileSize
	var lines []string
	var leftover []byte

	for offset > 0 && len(lines) < n {
		toRead := chunkSize
		if offset < toRead {
			toRead = offset
		}
		offset -= toRead

		_, err := f.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, toRead)
		if _, err := io.ReadFull(f, buf); err != nil {
			return nil, err
		}

		if len(leftover) > 0 {
			buf = append(buf, leftover...)
			leftover = nil
		}

		// Split by newline
		parts := bytes.Split(buf, []byte("\n"))

		if offset > 0 {
			// First part might be incomplete across chunk boundaries
			leftover = parts[0]
			parts = parts[1:]
		}

		for i := len(parts) - 1; i >= 0; i-- {
			line := string(bytes.TrimRight(parts[i], "\r"))
			// Skip trailing newline if empty at EOF
			if line == "" && offset+toRead == fileSize && len(lines) == 0 {
				continue
			}
			lines = append([]string{line}, lines...)
			if len(lines) >= n {
				break
			}
		}
	}

	if len(leftover) > 0 && len(lines) < n {
		lines = append([]string{string(bytes.TrimRight(leftover, "\r"))}, lines...)
	}

	return lines, nil
}

// HeadLines reads the first n lines from a file.
func HeadLines(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= n {
			break
		}
	}
	return lines, scanner.Err()
}

// CountLines returns the total count of lines in a file.
func CountLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	buf := make([]byte, 32*1024)
	var lastByte byte = '\n'

	for {
		n, err := f.Read(buf)
		if n > 0 {
			count += bytes.Count(buf[:n], []byte{'\n'})
			lastByte = buf[n-1]
		}
		if err == io.EOF {
			// If file didn't end with a newline, count the last line
			if lastByte != '\n' && n > 0 {
				count++
			}
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
