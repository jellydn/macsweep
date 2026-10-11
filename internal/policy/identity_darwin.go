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
	return Identity{uint64(s.Dev), s.Ino, uint64(s.Nlink), s.Uid, s.Size, s.Mtimespec.Nano(), s.Ctimespec.Nano()}, true
}

func unsafeStorage(info fs.FileInfo) bool {
	s, ok := info.Sys().(*syscall.Stat_t)
	// Darwin flags cover compressed, immutable, append-only and dataless files.
	// APFS clone ownership cannot be established here; cleanup remains disabled.
	return !ok || s.Flags != 0 || (s.Size > 0 && s.Blocks*512 < s.Size)
}
