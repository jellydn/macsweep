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

func preview(t *testing.T, exclusions ...string) *scan.Snapshot {
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
	p, err := policy.New(home, []policy.Rule{{ID: "fixture", Version: "1", Reason: "synthetic regenerated data", Evidence: "test only, not a shipping rule", Paths: paths, MinAge: time.Hour}}, exclusions)
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
	m.Progress(scan.Progress{Total: 257, Checked: 128, Candidates: 3, Skipped: 125})
	tt.Frame()
	if !tt.HasText("Scanning: 128 / 257 checked • 3 candidates • 125 skipped. Preview pending.") ||
		!tt.HasText("Selected files: 0 • Estimated bytes: 0") {
		t.Fatal("progress counts are hidden or stale selection survived")
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
				m.Progress(scan.Progress{Total: 100000, Checked: 100000, Candidates: 87654, Skipped: 12346})
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
	m.Progress(scan.Progress{Total: 3, Checked: 3, Candidates: 2, Skipped: 1})
	tt.Frame()
	if !tt.HasText("Preview pending.") || !tt.HasText("Selected files: 0 • Estimated bytes: 0") {
		t.Fatal("final progress was treated as a completed preview")
	}
	if err := tt.Click("Cancel scan"); err != nil {
		t.Fatal(err)
	}
	m.Progress(scan.Progress{Total: 3, Checked: 3, Candidates: 2, Skipped: 1})
	tt.Frame()
	if !tt.HasText("Scan cancelled. No preview is available.") {
		t.Fatal("cancellation left a usable preview")
	}
	if err := tt.Click("Scan Mac"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("3 / 3") || !tt.HasText("Scanning reviewed file metadata") {
		t.Fatal("new scan kept old progress")
	}
	m.Finish(nil, errors.New("fixture unavailable"))
	tt.Frame()
	if !tt.HasText("fixture unavailable") {
		t.Fatal("scan error hidden")
	}
}

func TestExclusionSaveClearsSelectionAndBlocksOnFailure(t *testing.T) {
	m := &appui.Model{}
	m.OnScan = m.Begin
	var saved []string
	m.OnSaveExclusions = func(paths []string) { saved = paths }
	m.ExclusionsLoaded([]string{"Documents"}, nil)
	m.Finish(preview(t), nil)
	tt := ui.NewTester(m.View, 640, 480)
	if err := tt.Click("Library/Caches/fixture/large.cache"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Exclusions"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Excluded paths"); err != nil {
		t.Fatal(err)
	}
	tt.Command("selectAll")
	tt.Type("Library/Caches/fixture\nDocuments")
	if err := tt.Click("Save exclusions"); err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 || saved[0] != "Library/Caches/fixture" || !m.PreferencesBusy {
		t.Fatalf("wrong save request: %v", saved)
	}
	m.ExclusionsLoaded(nil, errors.New("fixture save denied"))
	tt.Frame()
	if !tt.HasText("fixture save denied") {
		t.Fatal("save failure hidden")
	}
	if err := tt.Click("Back to preview"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Selected files: 0 • Estimated bytes: 0") || tt.HasText("large.cache") {
		t.Fatal("changed preferences kept stale preview")
	}
	if err := tt.Click("Scan Mac"); err != nil {
		t.Fatal(err)
	}
	if m.Scanning {
		t.Fatal("failed preferences enabled scan")
	}
	m.ExclusionsLoaded(saved, nil)
	tt.Frame()
	if err := tt.Click("Scan Mac"); err != nil || !m.Scanning {
		t.Fatalf("successful save did not unblock scan: %v", err)
	}
	m.Finish(preview(t, saved...), nil)
	tt.Frame()
	if !tt.HasText("Preview: 0 reviewed files; 2 skipped.") || !tt.HasText("excluded by path policy") ||
		!tt.HasText("Selected files: 0 • Estimated bytes: 0") {
		t.Fatal("excluded results were selectable or unexplained")
	}
}

func TestOffscreenSelectionCanBeCleared(t *testing.T) {
	m := &appui.Model{}
	m.Finish(preview(t), nil)
	tt := ui.NewTester(m.View, 640, 480)
	if err := tt.Click("Library/Caches/fixture/small.cache"); err != nil {
		t.Fatal(err)
	}
	tt.SetSize(520, 360)
	tt.Scroll(250, 200, 0, 1000)
	if !tt.HasText("Selected outside view: 1") || !tt.HasText("Selected files: 1 • Estimated bytes: 3") {
		t.Fatalf("hidden selection not explained: %v", tt.Texts())
	}
	if err := tt.Click("Clear selection"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Selected files: 0 • Estimated bytes: 0") || !tt.HasText("Selected outside view: 0") {
		t.Fatal("hidden selection not cleared")
	}
}
