package fileutils

// DiskStatus holds storage metrics for a filesystem mount or drive.
type DiskStatus struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	UsedPct   float64 `json:"used_pct"`
}

// DiskUsage returns the disk space metrics for the filesystem containing the specified path.
func DiskUsage(path string) (DiskStatus, error) {
	return getDiskUsage(path)
}
