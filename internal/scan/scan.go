// Package scan builds frozen, metadata-only previews. It has no cleanup API.
package scan

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/jellydn/macsweep/internal/policy"
)

const MaxTargets = 100_000

type Skipped struct{ Path, Reason string }
type Summary struct {
	Count int
	Bytes int64
}

// Progress contains counts only, never selectable partial results.
type Progress struct {
	Total, Checked, Candidates, Skipped int
}

// Snapshot is returned only for a complete scan. Accessors return copies.
type Snapshot struct {
	id      string
	files   []policy.File
	byID    map[string]policy.File
	skipped []Skipped
	checked int
}

// Run visits only exact approved paths, never walks an entire cache tree.
// A cancelled or bounded-out scan returns no usable partial snapshot.
func Run(ctx context.Context, p *policy.Policy) (*Snapshot, error) {
	return RunWithProgress(ctx, p, nil)
}

// RunWithProgress reports before inspection, after each batch of 128 targets,
// and after the last target. report runs synchronously on the caller's goroutine
// and must return promptly. Counts do not imply successful completion: callers
// must wait for the returned snapshot. Cancellation still discards all results.
func RunWithProgress(ctx context.Context, p *policy.Policy, report func(Progress)) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	targets := p.Targets()
	if len(targets) > MaxTargets {
		return nil, errors.New("scan exceeds the reviewed file limit")
	}
	s := &Snapshot{id: rand.Text(), byID: make(map[string]policy.File)}
	publish := func() {
		if report != nil {
			report(Progress{Total: len(targets), Checked: s.checked, Candidates: len(s.files), Skipped: len(s.skipped)})
		}
	}
	publish()
	now := time.Now()
	for _, target := range targets {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		file, err := p.Inspect(target, now)
		s.checked++
		if err != nil {
			s.skipped = append(s.skipped, Skipped{target.RelativePath, err.Error()})
		} else {
			s.files = append(s.files, file)
			s.byID[file.Target.ID] = file
		}
		if s.checked%128 == 0 || s.checked == len(targets) {
			publish()
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := summarize(s.files); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Snapshot) ID() string           { return s.id }
func (s *Snapshot) Files() []policy.File { return slices.Clone(s.files) }
func (s *Snapshot) Skipped() []Skipped   { return slices.Clone(s.skipped) }
func (s *Snapshot) Checked() int         { return s.checked }
func (s *Snapshot) Summary() Summary     { result, _ := summarize(s.files); return result }

type Selection struct {
	snapshotID string
	files      []policy.File
}

// Select binds IDs to this completed snapshot; its default is empty selection.
func (s *Snapshot) Select(ids ...string) (Selection, error) {
	selection := Selection{snapshotID: s.id}
	seen := make(map[string]bool)
	for _, id := range ids {
		file, ok := s.byID[id]
		if !ok || seen[id] {
			return Selection{}, errors.New("selection contains an unknown or duplicate candidate ID")
		}
		seen[id] = true
		selection.files = append(selection.files, file)
	}
	return selection, nil
}

func (s Selection) Files() []policy.File { return slices.Clone(s.files) }
func (s Selection) Summary() Summary     { result, _ := summarize(s.files); return result }

// Revalidate is read-only preflight, not authority to move any file. The caller
// must pass the current snapshot so a rescan invalidates the old selection.
func (s Selection) Revalidate(ctx context.Context, current *Snapshot, p *policy.Policy) error {
	if current == nil || s.snapshotID != current.id {
		return errors.New("selection belongs to a different scan")
	}
	for _, captured := range s.files {
		if err := ctx.Err(); err != nil {
			return err
		}
		file, err := p.Inspect(captured.Target, time.Now())
		if err != nil {
			return fmt.Errorf("selected file no longer passes policy: %w", err)
		}
		if file.Identity != captured.Identity || file.Path != captured.Path {
			return errors.New("selected file changed since preview")
		}
	}
	return ctx.Err()
}

func summarize(files []policy.File) (Summary, error) {
	result := Summary{Count: len(files)}
	for _, file := range files {
		if file.Identity.Size > math.MaxInt64-result.Bytes {
			return Summary{}, errors.New("file byte estimate overflows")
		}
		result.Bytes += file.Identity.Size
	}
	return result, nil
}
