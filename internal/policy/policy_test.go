package policy_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jellydn/macsweep/internal/policy"
)

func rule(paths ...string) policy.Rule {
	return policy.Rule{ID: "fixture", Version: "1", Reason: "regenerated test data", Evidence: "disposable synthetic fixture only", Paths: paths, MinAge: time.Hour}
}

func fixture(t *testing.T) (string, string) {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "Library/Caches/fixture/old.cache")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("disposable"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	return home, path
}

func TestDefaultHasNoApprovedRules(t *testing.T) {
	home, _ := fixture(t)
	p, err := policy.Default(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Targets()) != 0 {
		t.Fatal("shipping rules must require explicit approval")
	}
}

func TestExactRuleAndExclusions(t *testing.T) {
	home, path := fixture(t)
	rules := []policy.Rule{rule("Library/Caches/fixture/old.cache")}
	p, err := policy.New(home, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	target := p.Targets()[0]
	file, err := p.Inspect(target, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if file.Path != path || file.Identity.Size != 10 || file.Identity.Links != 1 {
		t.Fatalf("incorrect metadata: %+v", file)
	}
	// Constructor and accessor data cannot mutate the policy.
	rules[0].Paths[0] = "Documents/private"
	targets := p.Targets()
	targets[0].RelativePath = "Documents/private"
	if _, err := p.Inspect(targets[0], time.Now()); err == nil {
		t.Fatal("forged target accepted")
	}
	for _, excluded := range []string{"Library/Caches/fixture", "library/caches/FIXTURE", path, "."} {
		p, err := policy.New(home, []policy.Rule{rule(target.RelativePath)}, []string{excluded})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Inspect(target, time.Now()); err == nil {
			t.Fatalf("exclusion %q bypassed", excluded)
		}
	}
	p, err = policy.New(home, []policy.Rule{rule(target.RelativePath)}, []string{"Library/Caches/fixture2"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Inspect(target, time.Now()); err != nil {
		t.Fatalf("near-prefix exclusion matched: %v", err)
	}
}

func TestUnsafeProviderPaths(t *testing.T) {
	home, _ := fixture(t)
	for _, path := range []string{
		".", "/Library/Caches/app/file", "Documents/file", "Library/CachesExtra/app/file",
		"Library/Caches", "Library/Caches/file", "Library/Caches/app/../file", "Library/Caches/app/*",
		"Library/Caches/app/\x00file", "Library/Caches/Keychains/file", "Library/Caches/MacSweep/file",
		"Library/Caches/app/bad\xffname", "Library/Caches/app/Photos.photoslibrary/database",
	} {
		if _, err := policy.New(home, []policy.Rule{rule(path)}, nil); err == nil {
			t.Errorf("unsafe path %q accepted", path)
		}
	}
}

func TestUnicodeExclusionAndDuplicateAliases(t *testing.T) {
	home, original := fixture(t)
	path := "Library/Caches/fixture/caf\u00e9.cache"
	if err := os.Rename(original, filepath.Join(home, filepath.FromSlash(path))); err != nil {
		t.Fatal(err)
	}
	r := rule(path)
	p, err := policy.New(home, []policy.Rule{r}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Inspect(p.Targets()[0], time.Now()); err != nil {
		t.Fatalf("unexcluded Unicode fixture not accepted: %v", err)
	}
	p, err = policy.New(home, []policy.Rule{r}, []string{"Library/Caches/fixture/cafe\u0301.cache"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Inspect(p.Targets()[0], time.Now()); err == nil {
		t.Fatal("Unicode exclusion bypassed")
	}
	if _, err := policy.New(home, []policy.Rule{rule(path, "Library/Caches/fixture/cafe\u0301.cache")}, nil); err == nil {
		t.Fatal("alias rules accepted")
	}
}

func TestLinksYoungAndSparseFilesAreRejected(t *testing.T) {
	for _, mutation := range []string{"leaf symlink", "ancestor symlink", "hard link", "directory", "young", "sparse"} {
		t.Run(mutation, func(t *testing.T) {
			home, path := fixture(t)
			p, err := policy.New(home, []policy.Rule{rule("Library/Caches/fixture/old.cache")}, nil)
			if err != nil {
				t.Fatal(err)
			}
			switch mutation {
			case "leaf symlink":
				if err := os.Rename(path, path+".real"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path+".real", path); err != nil {
					t.Fatal(err)
				}
			case "ancestor symlink":
				dir := filepath.Dir(path)
				if err := os.Rename(dir, dir+".real"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(dir+".real", dir); err != nil {
					t.Fatal(err)
				}
			case "hard link":
				if err := os.Link(path, path+".link"); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "young":
				if err := os.Chtimes(path, time.Now(), time.Now()); err != nil {
					t.Fatal(err)
				}
			case "sparse":
				// Ext4 can preallocate an extended written file. Start from an empty inode.
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.Truncate(1 << 20); err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
				old := time.Now().Add(-48 * time.Hour)
				if err := os.Chtimes(path, old, old); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := p.Inspect(p.Targets()[0], time.Now()); err == nil {
				t.Fatalf("accepted %s", mutation)
			}
		})
	}
}
