// Package ui owns MacSweep's presentation. Filesystem operations belong to
// services outside the view and are not implemented in this setup build.
package ui

import "github.com/egoist/mygo/ui"

// View shows the setup state without exposing a working cleanup action.
func View(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(16).Padding(32).Children(func() {
		ui.Text(c, "MacSweep").FontSize(30).Bold()
		ui.Text(c, "Safety before cleanup").FontSize(20)
		ui.Text(c, "Setup build. Scanning and cleanup are not available yet.").TextColor(t.TextMuted)
		ui.Row(c).Gap(12).Children(func() {
			ui.Button(c, "Scan Mac").Disabled(true)
			ui.Button(c, "Move to Trash").Disabled(true)
		})
		ui.Text(c, "Preview first. Select files. Confirm before Trash.").TextColor(t.TextMuted)
		ui.Text(c, "No admin access. No network. No telemetry.").TextColor(t.TextMuted)
	})
}
