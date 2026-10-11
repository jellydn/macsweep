package scan_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/jellydn/macsweep/internal/policy"
	"github.com/jellydn/macsweep/internal/scan"
)

func fixture(t *testing.T) (*policy.Policy, string) {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"Library/Caches/fixture/a.cache", "Library/Caches/fixture/b.cache", "Library/Caches/fixture/missing.cache"}
	for i, data := range []string{"abc", "eight123"} {
		path := filepath.Join(home, filepath.FromSlash(paths[i]))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	p, err := policy.New(home, []policy.Rule{{ID: "fixture", Version: "1", Reason: "regenerated test data", Evidence: "synthetic only", Paths: paths, MinAge: time.Hour}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return p, home
}

func TestScanIsFrozenReadOnlyAndSelectionStartsEmpty(t *testing.T) {
	p, _ := fixture(t)
	target := p.Targets()[0]
	before, err := p.Inspect(target, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	s, err := scan.Run(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if s.Checked() != 3 || s.Summary() != (scan.Summary{Count: 2, Bytes: 11}) || len(s.Skipped()) != 1 {
		t.Fatalf("wrong scan totals: checked=%d summary=%+v skipped=%v", s.Checked(), s.Summary(), s.Skipped())
	}
	after, err := p.Inspect(target, time.Now())
	if err != nil || before != after {
		t.Fatalf("scan modified fixture: before=%+v after=%+v err=%v", before, after, err)
	}
	data, err := os.ReadFile(before.Path)
	if err != nil || string(data) != "abc" {
		t.Fatalf("content changed: %q %v", data, err)
	}
	empty, err := s.Select()
	if err != nil || empty.Summary() != (scan.Summary{}) {
		t.Fatalf("default selection is not empty: %v", err)
	}
	selection, err := s.Select(s.Files()[1].Target.ID)
	if err != nil || selection.Summary() != (scan.Summary{Count: 1, Bytes: 8}) {
		t.Fatalf("wrong selection: %+v %v", selection.Summary(), err)
	}
	files := s.Files()
	files[0].Path = "/Documents/private"
	selected := selection.Files()
	selected[0].Identity.Size = 999
	if s.Files()[0].Path == files[0].Path || selection.Summary().Bytes != 8 {
		t.Fatal("caller mutated frozen results")
	}
	for _, ids := range [][]string{{"/Documents/private"}, {target.ID, target.ID}} {
		if _, err := s.Select(ids...); err == nil {
			t.Fatal("invalid IDs accepted")
		}
	}
}

func TestCancelledScanReturnsNoSnapshot(t *testing.T) {
	p, _ := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s, err := scan.Run(ctx, p)
	if s != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled scan returned usable data: %v %v", s, err)
	}
}

func TestProgressCountsCandidatesAndSkips(t *testing.T) {
	p, _ := fixture(t)
	var reports []scan.Progress
	s, err := scan.RunWithProgress(context.Background(), p, func(progress scan.Progress) {
		reports = append(reports, progress)
	})
	want := []scan.Progress{{Total: 3}, {Total: 3, Checked: 3, Candidates: 2, Skipped: 1}}
	if err != nil || s == nil || !slices.Equal(reports, want) {
		t.Fatalf("snapshot=%v reports=%v err=%v", s, reports, err)
	}
}

func TestProgressBatchesAndCancellation(t *testing.T) {
	for _, total := range []int{0, 127, 128, 129, 257} {
		for _, cancelAt := range []int{-1, 0, 128, total} {
			t.Run(fmt.Sprintf("targets=%d/cancel=%d", total, cancelAt), func(t *testing.T) {
				paths := make([]string, total)
				for i := range paths {
					paths[i] = fmt.Sprintf("Library/Caches/fixture/missing-%d", i)
				}
				p, err := policy.New(t.TempDir(), []policy.Rule{{ID: "fixture", Version: "1", Reason: "test", Evidence: "synthetic only", Paths: paths, MinAge: time.Hour}}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var reports []scan.Progress
				cancelled := false
				s, err := scan.RunWithProgress(ctx, p, func(progress scan.Progress) {
					reports = append(reports, progress)
					if progress.Checked == cancelAt {
						cancelled = true
						cancel()
					}
				})
				if cancelled {
					if s != nil || !errors.Is(err, context.Canceled) {
						t.Fatalf("cancelled progress authorized snapshot: %v %v", s, err)
					}
				} else if err != nil || s == nil || s.Checked() != total {
					t.Fatalf("incomplete scan: %v %v", s, err)
				}
				end := total
				if cancelled {
					end = cancelAt
				}
				want := []scan.Progress{{Total: total}}
				for checked := 128; checked < end; checked += 128 {
					want = append(want, scan.Progress{Total: total, Checked: checked, Skipped: checked})
				}
				if end > 0 {
					want = append(want, scan.Progress{Total: total, Checked: end, Skipped: end})
				}
				if !slices.Equal(reports, want) {
					t.Fatalf("reports=%v want=%v", reports, want)
				}
			})
		}
	}
}

func TestRevalidationRejectsChangesExclusionsAndRescan(t *testing.T) {
	for _, change := range []string{"replace", "resize", "exclude", "rescan"} {
		t.Run(change, func(t *testing.T) {
			p, home := fixture(t)
			s, err := scan.Run(context.Background(), p)
			if err != nil {
				t.Fatal(err)
			}
			selection, err := s.Select(s.Files()[0].Target.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := selection.Revalidate(context.Background(), s, p); err != nil {
				t.Fatal(err)
			}
			path := selection.Files()[0].Path
			switch change {
			case "replace":
				// Same size and mtime must not hide a replaced inode.
				old := time.Unix(0, selection.Files()[0].Identity.ModifiedNS)
				if err := os.Rename(path, path+".original"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("xyz"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chtimes(path, old, old); err != nil {
					t.Fatal(err)
				}
			case "resize":
				if err := os.WriteFile(path, []byte("changed bytes"), 0600); err != nil {
					t.Fatal(err)
				}
			case "exclude":
				p, err = policy.New(home, []policy.Rule{{ID: "fixture", Version: "1", Reason: "regenerated test data", Evidence: "synthetic only", Paths: []string{"Library/Caches/fixture/a.cache"}, MinAge: time.Hour}}, []string{"Library/Caches/fixture"})
				if err != nil {
					t.Fatal(err)
				}
			case "rescan":
				s, err = scan.Run(context.Background(), p)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := selection.Revalidate(context.Background(), s, p); err == nil {
				t.Fatalf("accepted %s", change)
			}
		})
	}
}
