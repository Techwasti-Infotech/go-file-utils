package fileutils

import (
	"encoding/csv"
	"encoding/json"
	"os"
)

// ReadJSON reads and unmarshals JSON file content into v.
func ReadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals v into indented JSON and writes it atomically to path.
func WriteJSON(path string, v any, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, append(data, '\n'), perm)
}

// ReadCSV reads all records from a CSV file.
func ReadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	return reader.ReadAll()
}

// WriteCSV writes records to a CSV file.
func WriteCSV(path string, records [][]string, perm os.FileMode) error {
	if err := EnsureParentDir(path, DefaultDirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	if err := writer.WriteAll(records); err != nil {
		return err
	}
	return f.Sync()
}

