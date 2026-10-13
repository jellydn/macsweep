# User exclusions

Exclusions remove paths from consideration; they cannot approve provider rules or
override built-in exclusions. Paths are home-relative (for example
`Library/Caches/example`) or absolute within home. `.` excludes all of home.
Traversal, globs, line breaks and paths outside home are rejected. Line breaks
cannot round-trip through the one-path-per-line editor; stored paths with them
block scanning rather than silently changing exclusions. Existing component-aware,
case-folded and Unicode-normalized policy matching also applies to preferences.

Settings live in `~/Library/Application Support/MacSweep/preferences.json`, with
schema version 1. The app directory is private (0700), and the file is private
(0600). A save syncs each parent directory so newly created state directories
are persisted, including on retries. It validates exclusions before writing a fresh file, syncing it,
renaming it over the old preferences and syncing the directory. No scanned file
is written. This is preference storage, not a cleanup journal or crash-recovery
proof. Concurrent app instances use last successful save wins.

Absent settings mean no user exclusions. Invalid, unsupported, inaccessible,
linked or shared settings block scanning instead of silently dropping exclusions.
The file size is limited to 64 KiB. A failed save also blocks scanning, since an
error after replacement can leave persistence uncertain. An explicit successful
save or reload is required before scanning again. Directory links are refused;
the app does not change existing permissions to bypass an error.

Native APFS aliases, cloud storage, TCC and malicious same-user replacement are
not proved by these portable preference tests. No preference enables cleanup.
