// This developer-only harness uses a disposable home, never real caches.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/jellydn/macsweep/internal/policy"
	"github.com/jellydn/macsweep/internal/scan"
	appui "github.com/jellydn/macsweep/internal/ui"
)

type measurement struct {
	OS, Arch, GoVersion                         string
	CPUs, GOMAXPROCS, Entries, ProgressReports  int
	FixtureMS, ScanMS, CancelMS, PreviewFrameMS float64
	IdleHeadlessRSSKiB, PreviewRSSKiB, PeakKiB  int64
	Candidates, Skipped                         int
	Bytes                                       int64
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	result := measurement{OS: runtime.GOOS, Arch: runtime.GOARCH, GoVersion: runtime.Version(), CPUs: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0), Entries: scan.MaxTargets}
	m := &appui.Model{}
	tt := ui.NewTester(m.View, 640, 480)
	tt.SetScale(2)
	debug.FreeOSMemory()
	var err error
	result.IdleHeadlessRSSKiB, err = rssKiB()
	if err != nil {
		return err
	}
	home, err := os.MkdirTemp("", "macsweep-perf-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(home) // Only the directory this harness created.
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return err
	}
	paths := make([]string, result.Entries)
	start := time.Now()
	old := start.Add(-48 * time.Hour)
	for i := range paths {
		paths[i] = fmt.Sprintf("Library/Caches/fixture/%03d/%06d.cache", i/1000, i)
		path := filepath.Join(home, paths[i])
		if i%1000 == 0 {
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				return err
			}
		}
		if err := os.WriteFile(path, []byte("ab")[:1+i%2], 0600); err != nil {
			return err
		}
		if err := os.Chtimes(path, old, old); err != nil {
			return err
		}
	}
	p, err := policy.New(home, []policy.Rule{{ID: "fixture", Version: "1", Reason: "synthetic performance fixture", Evidence: "developer harness only; not approved for shipping", Paths: paths, MinAge: time.Hour}}, nil)
	if err != nil {
		return err
	}
	result.FixtureMS = milliseconds(time.Since(start))
	start = time.Now()
	snapshot, err := scan.RunWithProgress(context.Background(), p, func(progress scan.Progress) {
		result.ProgressReports++
		if progress.Checked != progress.Candidates+progress.Skipped || progress.Total != result.Entries {
			panic("inconsistent progress counts")
		}
	})
	result.ScanMS = milliseconds(time.Since(start))
	if err != nil {
		return err
	}
	result.Candidates, result.Bytes = snapshot.Summary().Count, snapshot.Summary().Bytes
	result.Skipped = len(snapshot.Skipped())
	if result.Candidates != 100000 || result.Bytes != 150000 || result.Skipped != 0 || result.ProgressReports != 783 {
		return fmt.Errorf("fixture scan contract failed: %+v", result)
	}
	start = time.Now()
	m.Finish(snapshot, nil)
	tt.Frame()
	result.PreviewFrameMS = milliseconds(time.Since(start))
	if !tt.HasText("Selected files: 0 • Estimated bytes: 0") {
		return errors.New("large preview did not start with empty selection")
	}
	debug.FreeOSMemory()
	result.PreviewRSSKiB, err = rssKiB()
	if err != nil {
		return err
	}

	// Cancel from a different goroutine while the scanner is doing real work,
	// not from its progress callback at a cancellation checkpoint.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 1)
	requested := make(chan time.Time, 1)
	go func() {
		select {
		case <-started:
			time.Sleep(10 * time.Millisecond)
			requested <- time.Now()
			cancel()
		case <-ctx.Done():
		}
	}()
	partial, err := scan.RunWithProgress(ctx, p, func(progress scan.Progress) {
		if progress.Checked == 128 {
			started <- struct{}{}
		}
	})
	finished := time.Now()
	if partial != nil || !errors.Is(err, context.Canceled) {
		return fmt.Errorf("cancellation did not discard partial snapshot: %v", err)
	}
	result.CancelMS = milliseconds(finished.Sub(<-requested))
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return err
	}
	result.PeakKiB = usage.Maxrss
	if runtime.GOOS == "darwin" {
		result.PeakKiB /= 1024
	}
	runtime.KeepAlive(tt) // Include the retained preview in the memory measurements.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func milliseconds(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

func rssKiB() (int64, error) {
	data, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		return 0, fmt.Errorf("measure process RSS with ps: %w", err)
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}
