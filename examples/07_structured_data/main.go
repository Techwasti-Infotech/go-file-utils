package main

import (
	"fmt"
	"os"
	"path/filepath"

	fileutils "github.com/Techwasti-Infotech/go-file-utils"
)

type AppSettings struct {
	AppName    string   `json:"app_name"`
	Port       int      `json:"port"`
	Features   []string `json:"features"`
	Production bool     `json:"production"`
}

func main() {
	fmt.Println("=== Example 07: Structured Data (JSON & CSV) ===")

	tempDir, err := os.MkdirTemp("", "fileutils-example-07-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Atomic JSON Write & Read
	jsonFile := filepath.Join(tempDir, "settings.json")
	settings := AppSettings{
		AppName:    "FileProcessingEngine",
		Port:       8080,
		Features:   []string{"clustering", "metrics", "encryption"},
		Production: true,
	}

	if err := fileutils.WriteJSON(jsonFile, settings, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("1. Written indented JSON to: %s\n", jsonFile)

	var loadedSettings AppSettings
	if err := fileutils.ReadJSON(jsonFile, &loadedSettings); err != nil {
		panic(err)
	}
	fmt.Printf("   Parsed JSON -> AppName: %s, Port: %d, Features: %v\n",
		loadedSettings.AppName, loadedSettings.Port, loadedSettings.Features)

	// 2. CSV Write & Read
	csvFile := filepath.Join(tempDir, "metrics.csv")
	records := [][]string{
		{"Timestamp", "Metric", "Value"},
		{"2026-11-07T00:00:00Z", "CPU_LOAD", "42.5"},
		{"2026-11-07T00:01:00Z", "MEM_USAGE", "65.1"},
		{"2026-11-07T00:02:00Z", "DISK_IOPS", "1200"},
	}

	if err := fileutils.WriteCSV(csvFile, records, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("2. Written CSV data to: %s\n", csvFile)

	readRecords, err := fileutils.ReadCSV(csvFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Read %d rows from CSV successfully.\n", len(readRecords))
}
