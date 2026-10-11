# Safety contract

This is an implementation contract, not a claim that cleanup exists. The current
build has read-only policy/scan/preview services but no cleanup mutation code.
Explicit preference saves write only MacSweep's local exclusion settings.
No provider rules are approved or shipped. Cleanup controls remain disabled.

## Scan and preview

Scan read-only within reviewed roots for the logged-in, non-root user. Never
request sudo, a privileged helper, or Full Disk Access. Report inaccessible
locations as skipped; do not ask the user to weaken macOS protections.

Return a frozen snapshot. Each candidate records rule/version, absolute path,
relative path, volume/device, inode, link count, type, size, nanosecond modification
and change times, owner, and reason. Do not read file contents to classify files.
Cancellation leaves no selectable partial snapshot. No selection is checked by
default. Selection uses candidate IDs, never arbitrary paths supplied by UI.

## Eligibility and paths

- Approve exact provider-relative paths and regeneration evidence before shipping.
  Unknown rules and broad cache/artifact basenames remain ineligible.
- Accept regular files only. Reject directory moves, symlinks in any component,
  hard links, sockets, devices, FIFOs, mount boundaries, and non-user-owned files.
- Use component-aware containment, not string prefixes or `filepath.Clean` alone.
  Reject traversal, NUL bytes, invalid encodings, aliases, and ambiguous resolution.
- Resolve with filesystem identity. Account for case-sensitive and insensitive
  APFS plus Unicode normalization. Display text is not authority to resolve a path.
- Built-in exclusions cannot be overridden. Never target `/` or the home root
  themselves. Exclude whole trees under `/System`, `/Library`, `/Applications`,
  `/Volumes`, documents/downloads/desktop/media, Keychains, Mail, Messages,
  browser profiles, app databases, containers, cloud storage, FileProvider data,
  backup data, and MacSweep's state/staging/Trash. Allowed children of home are
  still subject to exact provider rules; excluding home itself is not a ban on
  every descendant.
- A user exclusion can only remove eligibility. Do not traverse protected roots
  to discover candidates. No policy turns an entire `~/Library/Caches` tree safe.
- Initial eligibility is local user-volume files only. External, network, cloud,
  compressed/sparse/clone-dependent, active-app, and unverifiable files are skipped.
  Detecting open files is best-effort and is not proof of disposability.

The exact provider list is not approved yet. The first rule can ship only after
its evidence and tests pass. A scan may truthfully find no eligible files.

## Cleanup transaction

1. Require a completed snapshot, explicit selection, preview, and confirmation
   showing file count, paths, reasons, estimated bytes, and Trash limitations.
2. Revalidate policy, exclusions, ownership, every ancestor, mount, and captured
   identity. A replaced, resized, relinked, or newly excluded file is skipped.
3. Pin trusted ancestors with directory descriptors and no-follow operations.
   Move one selected regular file to a private 0700 staging location on the same
   volume with a descriptor-relative, no-overwrite rename. Verify the moved
   identity before calling native Trash. Never recursively remove a directory.
4. Write recovery intent durably before mutation; record each observed stage.
   Use restrictive local state permissions. Do not store file contents or upload paths.
5. Call native Trash only while the MyGo event loop is alive. Its call currently
   blocks the UI thread; do not hold locks needed by that thread. Prevent app exit
   during an active transaction and verify completion independently of a nil error.
6. Stop on cancellation between files. On a failure, retain recoverable staging.
   Restore only if the original parent identity is unchanged and the destination
   is vacant. Never overwrite or copy-delete as fallback. Unknown results require
   manual review, not automatic retry.

This staging design is a proposal that must pass native race/crash tests before
cleanup is enabled. Go portability and lexical checks alone do not prove it.
If the Darwin APIs cannot meet the contract, keep cleanup disabled and implement
a minimal native adapter in a separate reviewed change, or retain scan-only mode.

MyGo's Trash API returns no destination URL. Do not promise Finder **Put Back**,
one-click undo, or exact recovery from Trash, especially after staging changes
the source path. The MVP must explain recovery limits and retain local history.

## Results and boundaries

Report moved, skipped, failed, cancelled, and unknown items separately. Report
estimated bytes moved and observed free-space delta separately. APFS clones,
compression, snapshots, and other processes make reclaimed-space attribution
unreliable. A Trash move normally does not release space; never suggest emptying
Trash as the next automatic action.

No permanent deletion, background cleanup, Trash emptying, network, telemetry,
updater, or privilege escalation. Native signing/notarization and developer
dependency downloads are build-time operations, not app behavior. Disable MyGo
DevTools and the production inspector. Verify no app sockets on real Macs.

This contract covers accidental changes and ordinary replacement races. It does
not claim isolation from malicious processes already running as the same user.
