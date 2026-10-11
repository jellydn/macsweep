# MacSweep

A safety-first, offline macOS cleaner in Go and [MyGo](https://github.com/egoist/mygo).

**Status: read-only MVP foundation.** The app shows local disk capacity and scans
only exact reviewed provider files. No provider rules are approved or shipped yet,
so a normal scan inspects no cache files. Cleanup is always disabled. There is no
Trash service, history, recovery, or usable cleaner yet. Do not use this build to clean a Mac.

Follow the [MVP umbrella issue](https://github.com/jellydn/macsweep/issues/2),
[provider approval](docs/provider-rules.md), and [scan/preview contract](docs/scan-preview.md).

## Safety contract

- Preview before cleanup. Start with no selected files.
- Move confirmed, approved, unchanged regular files to Trash only.
- Reject symlinks, hard links, unsafe paths, changed identity, and unknown rules.
- Built-in exclusions cannot be disabled. User exclusions only narrow scope.
- No admin helper, sudo, Full Disk Access requirement, network, or telemetry.
- No permanent deletion, Trash emptying, automatic cleanup, or runtime updater.
- Bytes moved to Trash are not bytes freed. Report those values separately.

The app targets macOS 13+, Intel and Apple Silicon. These are targets, not tested
compatibility claims. See [the safety contract](docs/safety.md) and
[the implementation phases](docs/implementation.md).

## Development

Install Go **1.27.2**. MyGo is pinned to inspected revision
[`b7e651a`](https://github.com/egoist/mygo/commit/b7e651a0b43a682703253ebf8ebbfd467d115100).
It requires at least Go 1.27.1; this setup uses the current 1.27.2 patch.

```sh
go mod download
./scripts/check.sh
go run ./cmd/macsweep --version
# On a Mac with a graphical login session:
go tool mygo dev
go tool mygo build -platform darwin/universal -skip-notarize
```

`just` is optional; [justfile](justfile) exposes the same commands. Native Go UI
needs no Bun, frontend server, or webview. Source downloads and release
notarization use network during development, not from the shipped app.

Read [DEVELOPMENT.md](DEVELOPMENT.md), [CONTRIBUTING.md](CONTRIBUTING.md), and
[RELEASES.md](RELEASES.md). The release workflow prepares artifacts only after an
explicit safety gate. It does not create tags or publish releases.

## Reference projects

[Oak](https://github.com/jellydn/Oak) supplied MIT licensing, domain glossary,
developer documentation, command organization, and test-artifact ideas. It is a
Swift app, not a Go starter. Purge, PureMac, and Mole supplied reviewed safety,
UX, test, and packaging ideas, not imported cleanup code. See the
[source-backed comparison](docs/reference-review.md) for adopted and rejected patterns.

## License

[MIT](LICENSE), matching Oak's license. Dependency notices ship in app resources.
