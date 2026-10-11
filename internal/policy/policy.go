// Package policy validates metadata for exact reviewed provider files.
// Passing this read-only policy never authorizes cleanup.
package policy

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// Rule must be reviewed in source. There is no user-supplied allowlist loader.
type Rule struct {
	ID, Version, Reason, Evidence string
	Paths                         []string // Exact home-relative files, not directory or glob rules.
	MinAge                        time.Duration
}

type Target struct {
	ID, RuleID, Version, RelativePath, Reason string
	MinAge                                    time.Duration
}

// Identity captures metadata for later stale-snapshot detection, not race proof.
type Identity struct {
	Device, Inode, Links        uint64
	Owner                       uint32
	Size, ModifiedNS, ChangedNS int64
}

type File struct {
	Target   Target
	Path     string
	Identity Identity
}

type Policy struct {
	home       string
	homeInfo   fs.FileInfo
	targets    []Target
	byID       map[string]Target
	exclusions []string
}

var ErrIneligible = errors.New("file is not eligible")
var ErrExcluded = errors.New("excluded by path policy")

// Default ships no rules until exact provider paths have been approved.
func Default(home string) (*Policy, error) { return New(home, nil, nil) }

// WithExclusions replaces user exclusions without changing reviewed rules.
func (p *Policy) WithExclusions(paths []string) (*Policy, error) {
	next, err := New(p.home, nil, paths)
	if err != nil {
		return nil, err
	}
	next.targets, next.byID = p.targets, p.byID
	return next, nil
}

func New(home string, rules []Rule, exclusions []string) (*Policy, error) {
	if os.Getuid() == 0 {
		return nil, errors.New("scanning as root is not supported")
	}
	if home == "" {
		return nil, errors.New("home is required")
	}
	home, err := filepath.Abs(home)
	if err != nil {
		return nil, err
	}
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return nil, fmt.Errorf("resolve home: %w", err)
	}
	info, err := os.Lstat(home)
	if err != nil || !info.IsDir() || home == string(filepath.Separator) {
		return nil, errors.New("home must be a directory other than the filesystem root")
	}
	p := &Policy{home: home, homeInfo: info, byID: make(map[string]Target)}
	seen := make(map[string]bool)
	for _, rule := range rules {
		if rule.ID == "" || strings.ContainsAny(rule.ID, ":/\\") || rule.Version == "" || rule.Reason == "" || rule.Evidence == "" || rule.MinAge <= 0 {
			return nil, errors.New("rule requires ID, version, reason, evidence and positive minimum age")
		}
		for _, path := range rule.Paths {
			if !validRelative(path) || !within("Library/Caches", path) || len(strings.Split(path, "/")) < 4 || p.excluded(path) {
				return nil, fmt.Errorf("unsafe provider path %q", path)
			}
			key := folded(path)
			if seen[key] {
				return nil, fmt.Errorf("duplicate provider path %q", path)
			}
			seen[key] = true
			target := Target{rule.ID + ":" + path, rule.ID, rule.Version, path, rule.Reason, rule.MinAge}
			p.targets = append(p.targets, target)
			p.byID[target.ID] = target
		}
	}
	for _, path := range exclusions {
		if filepath.IsAbs(path) {
			// Validate before Rel can erase traversal components.
			if !validRelative(strings.TrimPrefix(filepath.ToSlash(path), "/")) {
				return nil, fmt.Errorf("invalid exclusion %q", path)
			}
			path, err = filepath.Rel(home, path)
			if err != nil {
				return nil, err
			}
		}
		path = filepath.ToSlash(path)
		if path != "." && !validRelative(path) {
			return nil, fmt.Errorf("invalid exclusion %q", path)
		}
		p.exclusions = append(p.exclusions, path)
	}
	return p, nil
}

func (p *Policy) Home() string      { return p.home }
func (p *Policy) Targets() []Target { return slices.Clone(p.targets) }

func validRelative(path string) bool {
	if !utf8.ValidString(path) || !fs.ValidPath(path) || strings.ContainsAny(path, "\x00\\:*?[]") || path == "." {
		return false
	}
	return !filepath.IsAbs(path)
}

func folded(path string) string { return cases.Fold().String(norm.NFD.String(path)) }

func within(root, path string) bool {
	root, path = folded(root), folded(path)
	return path == root || strings.HasPrefix(path, root+"/")
}

func (p *Policy) excluded(path string) bool {
	for _, root := range []string{
		"Library/Caches/MacSweep", "Library/Caches/com.apple", "Library/Caches/CloudKit",
		"Library/Caches/com.apple.bird", "Library/Caches/com.apple.FileProvider",
	} {
		if within(root, path) {
			return true
		}
	}
	for _, part := range strings.Split(folded(path), "/") {
		if slices.Contains([]string{"keychains", "mail", "messages", "containers", "group containers", "cloudstorage", "mobile documents", ".trash"}, part) ||
			strings.HasSuffix(part, ".app") || strings.HasSuffix(part, ".photoslibrary") {
			return true
		}
	}
	for _, root := range p.exclusions {
		if root == "." || within(root, path) {
			return true
		}
	}
	return false
}

// Inspect uses pinned directory roots and metadata only. It rejects links and
// identity changes while descending. No filesystem-mutation API is exposed.
func (p *Policy) Inspect(target Target, now time.Time) (File, error) {
	known, ok := p.byID[target.ID]
	if !ok || known != target {
		return File{}, ErrIneligible
	}
	if p.excluded(target.RelativePath) {
		return File{}, ErrExcluded
	}
	root, err := os.OpenRoot(p.home)
	if err != nil {
		return File{}, err
	}
	defer root.Close()
	homeInfo, err := root.Stat(".")
	if err != nil {
		return File{}, err
	}
	if !os.SameFile(p.homeInfo, homeInfo) {
		return File{}, ErrIneligible
	}
	homeID, ok := identity(homeInfo)
	if !ok {
		return File{}, errors.New("filesystem identity is unavailable")
	}
	current := root
	parts := strings.Split(target.RelativePath, "/")
	for i, part := range parts {
		info, err := current.Lstat(part)
		if err != nil {
			return File{}, err
		}
		id, ok := identity(info)
		if !ok || info.Mode()&os.ModeSymlink != 0 || id.Device != homeID.Device || id.Owner != uint32(os.Getuid()) {
			return File{}, ErrIneligible
		}
		if i == len(parts)-1 {
			if !info.Mode().IsRegular() || id.Links != 1 || id.Size < 0 || info.ModTime().After(now) || now.Sub(info.ModTime()) < target.MinAge || unsafeStorage(info) {
				return File{}, ErrIneligible
			}
			return File{target, filepath.Join(p.home, filepath.FromSlash(target.RelativePath)), id}, nil
		}
		if !info.IsDir() {
			return File{}, ErrIneligible
		}
		next, err := current.OpenRoot(part)
		if err != nil {
			return File{}, err
		}
		pinned, err := next.Stat(".")
		if err != nil || !os.SameFile(info, pinned) {
			next.Close()
			return File{}, ErrIneligible
		}
		after, err := current.Lstat(part)
		if err != nil || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(info, after) {
			next.Close()
			return File{}, ErrIneligible
		}
		defer next.Close()
		current = next
	}
	return File{}, ErrIneligible
}
