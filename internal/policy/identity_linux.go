package policy

import (
	"io/fs"
	"syscall"
)

func identity(info fs.FileInfo) (Identity, bool) {
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return Identity{}, false
	}
	return Identity{s.Dev, s.Ino, s.Nlink, s.Uid, s.Size, s.Mtim.Nano(), s.Ctim.Nano()}, true
}

func unsafeStorage(info fs.FileInfo) bool {
	s, ok := info.Sys().(*syscall.Stat_t)
	return !ok || (s.Size > 0 && s.Blocks*512 < s.Size)
}
