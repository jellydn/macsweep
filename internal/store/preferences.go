// Package store persists local preferences, not cleanup intent or recovery data.
package store

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/jellydn/macsweep/internal/policy"
)

const maxPreferencesBytes = 64 << 10

type preferences struct {
	Version    int      `json:"version"`
	Exclusions []string `json:"exclusions"`
}

// LoadExclusions treats absent preferences as first launch, but never ignores
// malformed or inaccessible preferences: doing so could silently remove a ban.
func LoadExclusions(home string) ([]string, error) {
	root, err := preferenceRoot(home, false)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open preferences: %w", err)
	}
	defer root.Close()
	before, err := root.Lstat("preferences.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, errors.New("preferences must not be a link or special file")
	}
	f, err := root.OpenFile("preferences.json", os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("read preferences: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(before, info) || !info.Mode().IsRegular() || !privateOwned(info) || info.Sys().(*syscall.Stat_t).Nlink != 1 {
		return nil, errors.New("preferences must be a private, user-owned regular file with one link")
	}
	data, err := io.ReadAll(io.LimitReader(f, maxPreferencesBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPreferencesBytes {
		return nil, errors.New("preferences exceed 64 KiB")
	}
	var prefs preferences
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&prefs); err != nil {
		return nil, fmt.Errorf("decode preferences: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("preferences contain trailing data")
	}
	if prefs.Version != 1 {
		return nil, errors.New("unsupported preferences version")
	}
	if _, err := policy.New(home, nil, prefs.Exclusions); err != nil {
		return nil, fmt.Errorf("invalid exclusions: %w", err)
	}
	return prefs.Exclusions, nil
}

// SaveExclusions replaces only the app's preference file, never a scanned file.
// An error after rename means persistence is uncertain; callers must block scans
// until preferences have been successfully saved or reloaded.
func SaveExclusions(home string, exclusions []string) error {
	if _, err := policy.New(home, nil, exclusions); err != nil {
		return fmt.Errorf("invalid exclusions: %w", err)
	}
	data, err := json.Marshal(preferences{Version: 1, Exclusions: exclusions})
	if err != nil {
		return err
	}
	if len(data) > maxPreferencesBytes {
		return errors.New("preferences exceed 64 KiB")
	}
	root, err := preferenceRoot(home, true)
	if err != nil {
		return fmt.Errorf("open preferences: %w", err)
	}
	defer root.Close()
	name := ".preferences-" + rand.Text()
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(name)
	_, writeErr := f.Write(data)
	syncErr := f.Sync()
	closeErr := f.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("write preferences: %w", err)
	}
	if err := root.Rename(name, "preferences.json"); err != nil {
		return fmt.Errorf("replace preferences: %w", err)
	}
	dir, err := root.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func privateOwned(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Uid == uint32(os.Getuid()) && info.Mode().Perm()&0077 == 0
}

func preferenceRoot(home string, create bool) (*os.Root, error) {
	root, err := os.OpenRoot(home)
	if err != nil {
		return nil, err
	}
	for _, component := range []string{"Library", "Application Support", "MacSweep"} {
		if create {
			if err := root.Mkdir(component, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				root.Close()
				return nil, err
			}
		}
		info, err := root.Lstat(component)
		if err != nil {
			root.Close()
			return nil, err
		}
		if !info.IsDir() || (component == "MacSweep" && !privateOwned(info)) {
			root.Close()
			return nil, errors.New("unsafe preferences directory; links and shared app state are not supported")
		}
		if create {
			// Persist the child's entry, including on a retry after a failed sync.
			parent, err := root.Open(".")
			if err != nil {
				root.Close()
				return nil, err
			}
			err = errors.Join(parent.Sync(), parent.Close())
			if err != nil {
				root.Close()
				return nil, fmt.Errorf("sync preferences parent: %w", err)
			}
		}
		next, err := root.OpenRoot(component)
		root.Close()
		if err != nil {
			return nil, err
		}
		opened, err := next.Stat(".")
		if err != nil || !os.SameFile(info, opened) {
			next.Close()
			return nil, errors.New("preferences directory changed while opening")
		}
		root = next
	}
	return root, nil
}
