# Implementation-ready phases

## MVP and exclusions

Ship a local disk dashboard, cancellable reviewed-rule scan, preview with initially
clear selection, explicit confirmation, Trash-only movement, and local result and
recovery history. Exact provider paths need evidence before eligibility. Unknown
large files may be omitted; a generic disk analyzer is not needed for this MVP.

Exclude uninstall, system cleaning, browser profile deletion, duplicate detection,
cloud/external volumes, optimization, empty Trash, permanent deletion, scheduled
cleanup, admin access, FDA requirement, telemetry, network, and runtime updates.

## Architecture

```text
cmd/macsweep → internal/ui → scan coordinator → policy + read-only walker
                          → cleanup coordinator → policy + Darwin file/Trash APIs
                                                → durable local recovery/history
```

`cmd/macsweep` owns the main event loop. `internal/ui` owns presentation only.
Add `internal/policy`, `scan`, `cleanup`, `platform`, and `store` when each is
implemented; do not create empty interfaces now. Pass immutable scan snapshots
and candidate IDs across the UI boundary. Workers use cancellation and bounded
queues, then apply view changes through MyGo `Window.Update`. No shell command
cleanup or network clients. Prefer versioned atomic local files to a database
dependency for the small MVP journal; prove crash recovery before choosing it.

## Milestones and exit gates

| Phase | Work | Exit gate |
| --- | --- | --- |
| 0: platform gate, 2–3 days | Finish setup; exercise native loop/Trash, accessibility, case/Unicode APFS, TCC denial, minimum OS and both CPUs | Record actual Mac evidence. Prove descriptor/staging design or stay scan-only. |
| 1: policy/dashboard, week 1 | Read-only capacity dashboard, exclusion engine, reviewed first provider rule, restrictive state | No writes during scan; protected paths ineligible; moved bytes never called freed bytes. |
| 2: scanner/review, week 2 | Bounded metadata walker, cancellation, partial-access errors, frozen snapshots and preview | No default selection; partial/cancelled scans cannot clean; measurable responsiveness. |
| 3: Trash/recovery, week 3 | Per-file revalidation, descriptor staging, native Trash, journal, results | Race, crash, denied access, occupied restore destination, and unknown-result tests pass. |
| 4: release, week 4 | Full Mac matrix, offline/accessibility/performance checks, signed notarized final artifacts | Owner approves provider list, safety evidence and release gate. No automatic publication. |

Weeks are estimates after Mac access and provider approval, not delivery promises.

## Acceptance tests

Use synthetic homes and disposable fixtures. Keep app-facing assertions; do not
use real user caches or Trash for automated destructive tests.

1. **Scope:** approved regular fixture is eligible; near-prefix root, `..`, NUL,
   symlink ancestor/leaf, hard link, different mount, protected and user-excluded
   paths are not. Case/Unicode aliases cannot bypass exclusions on either APFS mode.
2. **Read-only scan:** content, names, modes, ownership and timestamps do not change.
   TCC/permission denial produces skipped counts, not elevation prompts. Cloud
   placeholders are not read/hydrated. Cancelled snapshots cannot authorize cleanup.
3. **Preview/selection:** initial selection is empty; selected IDs resolve only in
   that snapshot; file count/bytes and excluded selections agree. No cleanup
   before preview and confirmation. A rescan invalidates the old selection.
4. **Revalidation/races:** replace a file, parent, inode, size, mtime/ctime, link
   count, or policy between scan and confirmation and during staging. No
   unselected or new file is sent to Trash. Unknown outcomes are not retried.
5. **Trash/recovery on Mac:** selected unchanged file is in native Trash; permanent
   deletion never occurs. Duplicate Trash names do not overwrite. Crash after each
   transaction boundary preserves recovery intent. Restore collision stops safely.
   Event-loop shutdown does not falsely report success. Test login/session denial.
6. **Space/results:** moving a known-size file increments moved bytes, not claimed
   freed bytes. Failed/cancelled files do not count as moved. Test APFS clones and
   sparse data; unavailable free-space readings are labelled unavailable.
7. **No network/admin:** run signed app offline with traffic/socket observation,
   no FDA, no helper, no sudo; all shipped features still work within allowed scope.
   Keep macOS Gatekeeper traffic distinct. Absence of URL opens alone is not proof.
8. **UI/performance:** light/dark, 520×360 and larger windows, keyboard and VoiceOver;
   no clipped preview/confirmation. On a recorded 100,000-entry fixture, bound
   memory, cancellation latency and UI updates. Initial targets: cancellation
   under one second and idle app RSS under 150 MiB; measure, do not claim now.
9. **Packaging:** minimum OS and latest stable on both CPUs; verify universal
   executable, app version/tag/SHA, Developer ID, hardened runtime, notarization,
   stapled tickets, checksums, quarantined first launch and offline launch.

## Risks and required decisions

- Native Trash is path-based, has no returned destination, and dispatches on the
  UI loop. Prove race/outcome/recovery handling before accepting cleanup scope.
- macOS permission restrictions, APFS identity rules, and minimum OS can invalidate
  a portable implementation. Cross-compilation is not runtime validation.
- Provider rules can destroy data if their regeneration evidence is wrong.
  Start with very few exact rules; unknowns remain ineligible.
- MyGo is evolving. Pin the inspected revision, review upgrades, and audit its
  production inspector/updater/network behavior. Do not claim static review is proof.
- GitHub creation needs owner access; publication needs separate authorization.
- Release requires trusted Mac access, Apple Developer ID/notary credentials, a
  reviewed environment and a named owner. No credentials or shared settings exist yet.

The original reviewed plan is linked from the source Amp thread. This file keeps
its implementation gates with the code and adds the inspected repository practices.
