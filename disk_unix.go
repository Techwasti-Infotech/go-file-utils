//go:build !windows

package fileutils

import (
	"syscall"
)

func getDiskUsage(path string) (DiskStatus, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskStatus{}, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	avail := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	var usedPct float64
	if total > 0 {
		usedPct = (float64(used) / float64(total)) * 100.0
	}

	return DiskStatus{
		Total:     total,
		Used:      used,
		Free:      free,
		Available: avail,
		UsedPct:   usedPct,
	}, nil
}
