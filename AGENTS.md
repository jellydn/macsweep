# MacSweep agent guidance

Read `CONTEXT.md` and `docs/safety.md` before changing scan or cleanup behavior.
This is a read-only build. Do not enable cleanup before the native safety gates pass.

## Boundaries

- Go + pinned MyGo native Go UI; no cgo in production builds.
- `cmd/macsweep` owns lifecycle; `internal/ui` owns presentation.
- Add policy, scan, cleanup, platform, and store packages only when implemented.
- Keep filesystem services outside view callbacks. Use `Window.Update` for UI state.
- No updater, sockets, telemetry, sudo, admin helper, or permanent delete API.
- Do not infer safety from age, folder name, classification, or user confirmation.
- No borrowed cleanup rules become eligible without evidence and tests.

## Checks and conventions

Use `gofmt`, idiomatic Go, errors with context, and early returns. Use conventional
commits: `type(scope): description`. Keep tests next to source. Use `t.TempDir()`
for filesystem fixtures; never operate on the developer's real caches or Trash.

Run `./scripts/check.sh`. Run Darwin builds with `CGO_ENABLED=0` for both
architectures after shared changes. Run and inspect representative light/dark UI
renders after appearance changes. Headless views do not prove AppKit, Metal,
TCC, VoiceOver, or native Trash behavior; test those on real Macs.

MyGo `Shell.TrashItem` synchronously dispatches to the UI thread and can return a
nil error for dropped post-shutdown work. Future cleanup must keep the app loop
active, verify outcomes, and handle unknown results without automatic retry.

Do not trigger releases, publish, push, change shared settings, or modify signing
credentials without explicit authorization. No automatic release on `main`.
