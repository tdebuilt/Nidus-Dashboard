//go:build !linux

package system

import "errors"

// getDiskUsage is not supported on non-Linux platforms.
func getDiskUsage(path string) (total, used uint64, err error) {
	return 0, 0, errors.New("disk usage not supported on this platform")
}
