# Reference review and decisions

Review date: 2026-10-10. This is a static source review at the revisions below,
not execution of the other cleaners. No cleanup source or allowlist was copied.
Only evidence-backed ideas that fit `safety.md` are retained.

## Pinned evidence

- [Oak `6817932`](https://github.com/jellydn/Oak/tree/68179323e1f72770fe3d92f11e8541dcf7684902):
  `LICENSE`, `AGENTS.md`, `CONTEXT.md`, `justfile`, workflows and release builder.
- [MyGo `b7e651a`](https://github.com/egoist/mygo/tree/b7e651a0b43a682703253ebf8ebbfd467d115100):
  `go.mod`, `modules.go`, `loop.go`, `internal/darwin/system.go`,
  native templates, `cmd/mygo/{build,config,dmg}.go`, UI tester and inspector code.
- [Purge `8d95a89`](https://github.com/jithin-sabu/purge-app/tree/8d95a892ae04cd92572045847f0be1bd9fc97377):
  `purge/Services/{DeletionSafetyPolicy,FileDeleter,ScanAccess,ScanExclusions}.swift`,
  `purge/Views/DeletionConfirmSheet.swift`, `.github/workflows/build.yml`.
- [PureMac `0eca648`](https://github.com/momenbasel/PureMac/tree/0eca648d48986b7ea91f56ffe816f817a719c421):
  `PureMac/Services/{ScanEngine,CleaningEngine,CleanupExclusions,UpdateService}.swift`,
  `PureMac/Views/CategoryDetailView.swift`, CLI `Core/Cleaner.swift`, release workflow.
- [Mole `383a037`](https://github.com/tw93/Mole/tree/383a037ff5d6b77f0ebc2a8aec100a46b6d3417c):
  `lib/core/file_ops.sh`, `bin/clean.sh`, `cmd/analyze/delete.go`,
  `tests/core_safe_functions.bats`, `cmd/analyze/delete_test.go`, release workflows.

## Comparison

| Area | Purge | PureMac | Mole | MacSweep decision |
| --- | --- | --- | --- | --- |
| Rules | Broad cache/artifact categories and app-related cleanup | Cache, app support/container and CLI categories | Shell Clean categories; separate Analyze selection | Exact reviewed provider files only; do not import broad rules. |
| Safety | Deletion policy, access/exclusion model, confirmation; some permanent/privileged paths | Exclusions and identity checks; GUI/CLI differ, default selections and permanent paths exist | Shell safe-removal guards; Analyze has separate Trash/identity handling | Fail closed; no default selections, permanent operations or escalation. |
| Path handling | Standardized/canonical containment and ownership checks | Path and stale identity checks, stronger in CLI | Normalization/exclusions, identity/open-handle and destination ownership checks | Component-aware exclusions plus descriptor-relative staging proof; no claim these path guards eliminate races. |
| UX | Limited-access status, confirmation, moved-versus-freed accounting | Category review, exclusion/hidden-selection warnings | Terminal categories, dry run, explicit Analyze actions | Clear preview/reasons/skipped status, separate space accounting; no terminal clean-now flow. |
| Tests | Policy tests exist; inspected build CI does not run them | Swift tests and release verification; no complete native race proof established | Bats synthetic HOME guards and Go Analyze deletion tests | Disposable fixtures, wrong-target/race/crash tests plus real Mac gates. Existing tests in another app are not proof for MacSweep. |
| Packaging | Swift native app; Sparkle integration | GUI/CLI release artifacts | Shell distribution and dual-architecture Go binaries | Native Go universal `.app`, test ZIP, signed release ZIP/DMG. No installer shell scripts or updater. |
| Release | Build/release machinery; no complete Apple notarization chain established here | Tag/SHA/version, sign/notarize/staple/checksum gates | Dual binaries, checksum/provenance; Apple notarization not proved | Adopt PureMac-style artifact gates; explicitly block unverified release. |

### Oak: conventions, not an app skeleton

Oak is Swift/SwiftUI with XcodeGen. Its MIT license names Dung Huynh Duc (2026).
Keep MIT, a domain glossary, readable commands/docs, conventional commits, isolated
fixtures, and short-lived test artifacts. Its earlier production thread confirms
the build → sign → verify → package → upload order and corrected `lipo` usage
(input first, then `-verify_arch arm64 x86_64`).

Do not copy Swift/Xcode tooling, Sparkle, automatic patch releases or workflow
writes to `main`. Oak's ad-hoc signing is suitable for development artifacts,
not evidence of Developer ID/notarized releases. MacSweep separates manual
preparation from publication and requires native safety approval first.

### Adopted ideas and explicit rejections

- **Purge:** keep limited-access scan UX, visible reasons/confirmation, and the
  distinction between moved and freed bytes. Reject broad artifact basenames,
  directory/simulator permanent actions, privileged uninstall and Sparkle.
- **PureMac:** keep hidden/excluded-selection warnings, stale-identity checks and
  exact release identity/signing gates. Reject selected-by-default categories,
  broad Application Support/container deletion, `removeItem`/root `rm`, FDA/admin
  retry, and runtime updates. Identity comparison alone still has a race.
- **Mole:** keep adversarial synthetic-home tests, fail-closed identity/open-handle
  ideas and no-overwrite destination ownership checks. Reject default real shell
  cleanup, `rm`, Trash emptying, sudo, updater and whole-directory operations.
  Analyze's Trash behavior is not the behavior of shell Clean. Its “whitelist”
  means keep/exclude; do not interpret it as a deletion allowlist.

No claim is made that any reference has zero telemetry, perfect race safety, or
complete native acceptance coverage. This review uses only the inspected paths.

## MyGo and macOS evidence

MyGo provides native Go views, Metal rendering on Mac, a main-thread event loop,
window updates and headless CPU-rendered tests; no cgo/web frontend is required.
The pin requires Go 1.27.1. MacSweep uses 1.27.2, a verified official patch archive.
The native config uses `main` and supports universal builds and `minimumSystemVersion`.
The native inspector is an in-window panel, not a network listener. Disable it
for production to reduce debug surface; no runtime updater is configured.

`Shell.TrashItem` calls native `NSFileManager` Trash and returns an error only, not
the resulting URL. `loop.go` can drop dispatch after shutdown without setting an
error. The current dispatch is synchronous on the UI thread. No promise of
nonblocking Trash, exact destination, Put Back, or safe post-shutdown completion.

Apple documents Trash through
[FileManager.trashItem](https://developer.apple.com/documentation/foundation/filemanager/trashitem(at:resultingitemurl:))
and file protections through
[macOS security](https://support.apple.com/guide/security/controlling-app-access-to-files-secddd1d86a6/web).
Permissions can deny access even for the logged-in user. Skip those locations.
[Notarization](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution)
is a distribution gate, not proof that cleanup is safe. Verify the proposed
macOS 13 floor and both architectures on real machines before claiming support.
