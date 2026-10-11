package platform

import "syscall"

// Volume reports bytes available to this user, not unreserved free blocks.
func Volume(path string) (Capacity, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Capacity{}, err
	}
	return Capacity{stat.Blocks * uint64(stat.Bsize), stat.Bavail * uint64(stat.Bsize)}, nil
}
