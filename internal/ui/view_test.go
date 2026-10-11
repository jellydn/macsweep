package ui_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/jellydn/macsweep/internal/platform"
	"github.com/jellydn/macsweep/internal/policy"
	"github.com/jellydn/macsweep/internal/scan"
	appui "github.com/jellydn/macsweep/internal/ui"
)

func preview(t *testing.T) *scan.Snapshot {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"Library/Caches/fixture/small.cache", "Library/Caches/fixture/large.cache"}
	for i, content := range []string{"abc", "eight123"} {
		path := filepath.Join(home, filepath.FromSlash(paths[i]))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	p, err := policy.New(home, []policy.Rule{{ID: "fixture", Version: "1", Reason: "synthetic regenerated data", Evidence: "test only, not a shipping rule", Paths: paths, MinAge: time.Hour}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	s, err := scan.Run(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReadOnlyPreviewSelectionAndReset(t *testing.T) {
	m := &appui.Model{Capacity: platform.Capacity{Total: 400 << 30, Available: 137 << 30}}
	m.OnScan = m.Begin
	tt := ui.NewTester(m.View, 640, 480)
	if !tt.HasText("Total: 400.0 GiB • Available: 137.0 GiB") {
		t.Fatal("wrong capacity labels or units")
	}
	s := preview(t)
	m.Finish(s, nil)
	tt.Frame()
	if !tt.HasText("Selected files: 0 • Estimated bytes: 0") {
		t.Fatal("selection did not start empty")
	}
	if err := tt.Click("Library/Caches/fixture/large.cache"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Selected files: 1 • Estimated bytes: 8") {
		t.Fatalf("wrong selected count/estimate: %v", tt.Texts())
	}
	if err := tt.Click("Move to Trash"); err != nil {
		t.Fatal(err)
	}
	if tt.Focused("Move to Trash") {
		t.Fatal("selected preview enabled cleanup")
	}
	tt.SetSize(520, 360)
	tt.Scroll(250, 215, 0, 200)
	if !tt.HasText("synthetic regenerated data") || !tt.HasText("Selected files: 1 • Estimated bytes: 8") {
		t.Fatal("minimum-size scrolling hid the reason or lost the selection")
	}
	if err := tt.Click("Scan Mac"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Selected files: 0 • Estimated bytes: 0") || tt.HasText("large.cache") {
		t.Fatal("rescan kept stale selection or candidates")
	}
	if urls := tt.OpenedURLs(); len(urls) != 0 {
		t.Fatalf("preview opened external URLs: %v", urls)
	}
}

func TestCleanupIsDisabledInAllReadOnlyStates(t *testing.T) {
	for _, state := range []string{"idle", "scanning", "empty", "cancelled", "denied"} {
		t.Run(state, func(t *testing.T) {
			m := &appui.Model{}
			switch state {
			case "scanning":
				m.Begin()
			case "empty":
				p, err := policy.Default(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				s, err := scan.Run(context.Background(), p)
				if err != nil {
					t.Fatal(err)
				}
				m.Finish(s, nil)
			case "cancelled":
				m.Finish(nil, context.Canceled)
			case "denied":
				m.Finish(nil, os.ErrPermission)
				m.CapacityError = "permission denied"
			}
			tt := ui.NewTester(m.View, 520, 360)
			if err := tt.Click("Move to Trash"); err != nil {
				t.Fatal(err)
			}
			if tt.Focused("Move to Trash") {
				t.Fatal("cleanup control is active")
			}
			if !tt.HasText("native safety gates pending") || !tt.HasText("normally free no disk space") {
				t.Fatal("safety limits not explained")
			}
			if state == "empty" && !tt.HasText("No cache files were inspected") {
				t.Fatal("empty rules produced misleading results")
			}
			if state == "denied" && !tt.HasText("Disk capacity unavailable") {
				t.Fatal("unavailable capacity was presented as zero")
			}
		})
	}
}

func TestCancelScanClearsThePreview(t *testing.T) {
	m := &appui.Model{}
	m.OnScan = m.Begin
	m.OnCancel = func() { m.Finish(nil, context.Canceled) }
	tt := ui.NewTester(m.View, 640, 480)
	if err := tt.Click("Scan Mac"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Cancel scan"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Scan cancelled. No preview is available.") {
		t.Fatal("cancellation left a usable preview")
	}
	m.Finish(nil, errors.New("fixture unavailable"))
	tt.Frame()
	if !tt.HasText("fixture unavailable") {
		t.Fatal("scan error hidden")
	}
}
