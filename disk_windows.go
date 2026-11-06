//go:build windows

package fileutils

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func getDiskUsage(path string) (DiskStatus, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	ptr, err := syscall.UTF16PtrFromString(absPath)
	if err != nil {
		return DiskStatus{}, err
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	r1, _, errSys := procGetDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if r1 == 0 {
		return DiskStatus{}, errSys
	}

	used := totalNumberOfBytes - totalNumberOfFreeBytes
	var usedPct float64
	if totalNumberOfBytes > 0 {
		usedPct = (float64(used) / float64(totalNumberOfBytes)) * 100.0
	}

	return DiskStatus{
		Total:     totalNumberOfBytes,
		Used:      used,
		Free:      totalNumberOfFreeBytes,
		Available: freeBytesAvailable,
		UsedPct:   usedPct,
	}, nil
}
