//go:build linux

package system

import "syscall"

// getDiskUsage returns total and used bytes for a mount point.
func getDiskUsage(path string) (total, used uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used = total - free
	return total, used, nil
}
