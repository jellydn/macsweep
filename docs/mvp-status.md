# Read-only MVP stack status

[Umbrella issue #2](https://github.com/jellydn/macsweep/issues/2) owns all remaining
work. The stack is `main ← mvp/policy ← mvp/scan ← mvp/preview`. Each PR depends
on the layer below it; review bottom first. These PRs do not complete the MVP.

## Implemented for review

- Source-reviewed exact-file rule contract and immutable path policy. Shipping
  rule set stays empty; no user allowlist/import can widen it.
- Pinned directory descent, conservative case/Unicode exclusions, file identity,
  age/type/link/owner/device/sparse checks; root scanning rejected.
- Metadata-only bounded scans, frozen snapshots, skipped reasons, cancellation,
  initially empty snapshot-bound selection and stale-data preflight.
- Local capacity dashboard, cancellable read-only UI, virtualized preview rows,
  selection estimates and visible blocked-cleanup/space-accounting notices.
- Tests for asymmetric estimates, exclusions/aliases/links, identity changes,
  rescan invalidation, cancellation and disabled Trash across UI states.

## What is deliberately absent

No approved real provider rules, cleanup service, Trash calls, staging, durable
recovery/history, confirmation-to-mutation flow, automatic retries, or release.
Read-only preflight is not cleanup authority. The disabled Trash button has no
handler. No runtime updater, app network client, telemetry or elevation was added.

## Gates still requiring evidence or owner action

1. Approve the first exact provider rule with regeneration/restore evidence.
2. Test proposed minimum macOS and latest stable on Intel and Apple Silicon.
3. Prove APFS case/Unicode, mount/clone/cloud/active-app and TCC behavior.
4. Prove native descriptor staging, replacement races, crash recovery, occupied
   restore destinations, Trash outcomes and event-loop shutdown before cleanup work.
5. Record native cancellation/RSS performance, Metal/VoiceOver/keyboard behavior
   and real app socket/offline evidence. Linux/headless tests are not substitutes.
6. Provision trusted signing/notary access and reviewed release environment only
   after safety approval. Creating tags, preparing/publishing a release and merging
   PRs require separate approval; no such actions were performed.

Follow the umbrella checklists, [safety contract](safety.md),
[provider approval](provider-rules.md) and [scan contract](scan-preview.md).
