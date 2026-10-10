# Development

## Requirements

- Go 1.27.2, pinned in `.go-version` and `go.mod`.
- macOS with a graphical login session for the app; proposed floor macOS 13.
- Xcode 14+ notarization tools and Developer ID only for release preparation.
- No Bun, Node.js, SQLite, server, database, or app runtime secrets.
- `just` is optional. The shell/Go commands are the source of truth.

## Commands

```sh
go mod download
./scripts/check.sh          # formatting, module verification, vet, race tests
go run ./cmd/macsweep --version
go tool mygo doctor
go tool mygo dev            # Mac graphical session only
go tool mygo build -platform darwin/universal -skip-notarize
```

Production compilation uses `CGO_ENABLED=0`. The Go race detector needs a host C
compiler; this does not add cgo to the shipped app. Core/headless view tests run
on Linux and macOS. Neither cross-builds nor headless tests prove macOS runtime
support. Use both Intel and Apple Silicon Macs before making compatibility claims.

## Orb setup

`.agents/setup` installs a checksum-verified Linux Go archive under the user's
home, adds `go`/`gofmt` to `~/.local/bin`, and downloads pinned modules. It is
idempotent and needs no credentials. It does not start the native app or services.

## PR test artifacts

CI packages a universal, ad-hoc-signed app on a macOS runner. Artifacts include
the tested Git SHA and expire after seven days. They are not notarized releases.
Download and unzip the artifact, then extract `MacSweep-test.zip`. Verify its
signature and architecture before opening it. Gatekeeper can require explicit
user approval. Never disable Gatekeeper or remove quarantine as a test shortcut.

The current window must say that scanning and cleanup are unavailable; both
buttons must remain disabled. Check light/dark mode, keyboard focus, window
resizing, and VoiceOver on a Mac. Do not grant Full Disk Access.
