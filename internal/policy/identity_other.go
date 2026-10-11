//go:build !linux && !darwin

package policy

import "io/fs"

func identity(fs.FileInfo) (Identity, bool) { return Identity{}, false }
func unsafeStorage(fs.FileInfo) bool        { return true }
