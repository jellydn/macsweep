package store_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/jellydn/macsweep/internal/store"
)

func TestExclusionsRoundTripAndInvalidSave(t *testing.T) {
	home := t.TempDir()
	paths, err := store.LoadExclusions(home)
	if err != nil || len(paths) != 0 {
		t.Fatalf("first launch: %v %v", paths, err)
	}
	want := []string{"Library/Caches/fixture", "Documents"}
	if err := store.SaveExclusions(home, want); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"../outside", home + "/Library/../Documents"} {
		if err := store.SaveExclusions(home, []string{invalid}); err == nil {
			t.Fatalf("invalid preference accepted: %q", invalid)
		}
	}
	paths, err = store.LoadExclusions(home)
	if err != nil || !slices.Equal(paths, want) {
		t.Fatalf("saved exclusions lost: %v %v", paths, err)
	}
	dir := filepath.Join(home, "Library/Application Support/MacSweep")
	for path, mode := range map[string]os.FileMode{dir: 0700, filepath.Join(dir, "preferences.json"): 0600} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != mode {
			t.Fatalf("unsafe permissions: %s %v %v", path, info, err)
		}
	}
	if err := store.SaveExclusions(home, nil); err != nil {
		t.Fatal(err)
	}
	paths, err = store.LoadExclusions(home)
	if err != nil || len(paths) != 0 {
		t.Fatalf("could not clear preferences: %v %v", paths, err)
	}
}

func TestUnsafePreferencesBlockLoad(t *testing.T) {
	for _, state := range []string{"corrupt", "version", "unknown", "traversal", "symlink", "hardlink", "permissions", "directory-link"} {
		t.Run(state, func(t *testing.T) {
			home := t.TempDir()
			if err := store.SaveExclusions(home, []string{"Documents"}); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(home, "Library/Application Support/MacSweep")
			path := filepath.Join(dir, "preferences.json")
			var err error
			switch state {
			case "corrupt":
				err = os.WriteFile(path, []byte("{"), 0600)
			case "version":
				err = os.WriteFile(path, []byte(`{"version":2,"exclusions":[]}`), 0600)
			case "unknown":
				err = os.WriteFile(path, []byte(`{"version":1,"allow":["Documents"]}`), 0600)
			case "traversal":
				err = os.WriteFile(path, []byte(`{"version":1,"exclusions":["../outside"]}`), 0600)
			case "symlink":
				if err = os.Rename(path, path+".original"); err == nil {
					err = os.Symlink("preferences.json.original", path)
				}
			case "hardlink":
				err = os.Link(path, path+".link")
			case "permissions":
				err = os.Chmod(path, 0644)
			case "directory-link":
				if err = os.Rename(dir, dir+".original"); err == nil {
					err = os.Symlink("MacSweep.original", dir)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.LoadExclusions(home); err == nil {
				t.Fatalf("unsafe %s loaded as empty/default preferences", state)
			}
		})
	}
}
