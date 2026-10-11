// Package ui owns presentation. Filesystem work runs outside view callbacks.
package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/jellydn/macsweep/internal/platform"
	"github.com/jellydn/macsweep/internal/policy"
	"github.com/jellydn/macsweep/internal/scan"
)

// Model is used only on the UI thread. Worker results arrive through Window.Update.
type Model struct {
	Capacity                   platform.Capacity
	CapacityError, PolicyError string
	OnScan, OnCancel           func()
	Scanning                   bool
	snapshot                   *scan.Snapshot
	files                      []policy.File
	skipped                    []scan.Skipped
	selected                   map[string]bool
	status                     string
	list                       ui.ListState
}

func (m *Model) Begin() {
	m.snapshot, m.files, m.skipped, m.selected = nil, nil, nil, nil
	m.list = ui.ListState{}
	m.Scanning = true
	m.status = "Scanning reviewed file metadata…"
}

func (m *Model) Progress(progress scan.Progress) {
	if !m.Scanning {
		return
	}
	m.status = fmt.Sprintf("Scanning: %d / %d checked • %d candidates • %d skipped. Preview pending.",
		progress.Checked, progress.Total, progress.Candidates, progress.Skipped)
}

func (m *Model) Finish(snapshot *scan.Snapshot, err error) {
	m.Scanning = false
	m.snapshot, m.files, m.skipped, m.selected = nil, nil, nil, nil
	if err != nil {
		m.status = "Scan unavailable: " + err.Error()
		if errors.Is(err, context.Canceled) {
			m.status = "Scan cancelled. No preview is available."
		}
		return
	}
	m.snapshot = snapshot
	m.files, m.skipped = snapshot.Files(), snapshot.Skipped()
	m.status = fmt.Sprintf("Preview: %d reviewed files; %d skipped.", len(m.files), len(m.skipped))
	if snapshot.Checked() == 0 {
		m.status = "No approved provider rules. No cache files were inspected."
	}
}

func (m *Model) View(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Gap(10).Padding(20).Children(func() {
		ui.Text(c, "MacSweep").FontSize(26).Bold()
		ui.Text(c, "Read-only preview • No admin access, network or telemetry").FontSize(12).TextColor(t.TextMuted)
		ui.Row(c).Gap(10).Children(func() {
			ui.Button(c, "Scan Mac").Disabled(m.Scanning || m.OnScan == nil).OnClick(m.OnScan)
			if m.Scanning {
				ui.Button(c, "Cancel scan").Disabled(m.OnCancel == nil).OnClick(m.OnCancel)
			}
		})
		if m.CapacityError != "" {
			ui.Text(c, "Disk capacity unavailable: "+m.CapacityError).FillWidth().MaxLines(2)
		} else if m.Capacity.Total != 0 {
			ui.Text(c, fmt.Sprintf("Total: %.1f GiB • Available: %.1f GiB", float64(m.Capacity.Total)/(1<<30), float64(m.Capacity.Available)/(1<<30)))
		}
		status := m.status
		if status == "" {
			status = "Preview reviewed files. All selections start clear."
		}
		if m.PolicyError != "" {
			status = "Scan unavailable: " + m.PolicyError
		}
		ui.Text(c, status).FillWidth().MaxLines(3)
		m.list.Key = func(i int) any {
			if i < len(m.files) {
				return m.files[i].Target.ID
			}
			return "skip:" + m.skipped[i-len(m.files)].Path
		}
		ui.List(c, &m.list, len(m.files)+len(m.skipped), func(i int) {
			ui.Column(c).FillWidth().Gap(4).PaddingY(8).Children(func() {
				if i >= len(m.files) {
					skip := m.skipped[i-len(m.files)]
					ui.Text(c, "Skipped: "+skip.Path).FillWidth()
					ui.Text(c, skip.Reason).FillWidth().FontSize(12).TextColor(t.TextMuted)
					return
				}
				file := m.files[i]
				checked := m.selected[file.Target.ID]
				ui.Checkbox(c, &checked, file.Target.RelativePath).OnChange(func() {
					if m.selected == nil {
						m.selected = make(map[string]bool)
					}
					m.selected[file.Target.ID] = checked
				})
				ui.Text(c, file.Path).FillWidth().FontSize(12).TextColor(t.TextMuted)
				ui.Text(c, fmt.Sprintf("%d bytes • %s", file.Identity.Size, file.Target.Reason)).FillWidth().FontSize(12)
			})
		}).Grow(1).FillWidth()
		var summary scan.Summary
		if m.snapshot != nil {
			var ids []string
			for id, checked := range m.selected {
				if checked {
					ids = append(ids, id)
				}
			}
			selection, err := m.snapshot.Select(ids...)
			if err == nil {
				summary = selection.Summary()
			}
		}
		ui.Text(c, fmt.Sprintf("Selected files: %d • Estimated bytes: %d", summary.Count, summary.Bytes))
		ui.Row(c).Gap(10).Children(func() {
			// No handler exists: preview and selection never authorize mutation.
			ui.Button(c, "Move to Trash").Disabled(true)
			ui.Text(c, "Cleanup blocked: native safety gates pending.").FontSize(12).Grow(1)
		})
		ui.Text(c, "Trash moves normally free no disk space.").FontSize(12).TextColor(t.TextMuted)
	})
}
