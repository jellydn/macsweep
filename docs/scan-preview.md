# Read-only scan and selection

Related work: [MVP umbrella #2](https://github.com/jellydn/macsweep/issues/2).

`scan.Run` reads metadata for exact reviewed rule paths. It does not enumerate
whole cache trees, read file contents, discover unknown large files, or hydrate
cloud placeholders. The shipping policy is empty until provider approval.

Work is serial and capped at 100,000 configured targets. Cancellation is checked
between metadata inspections and before completion. A cancelled or over-limit
scan returns no usable snapshot. A slow native filesystem call is not interruptible
by this check; measure native cancellation latency before claiming the one-second gate.

`scan.RunWithProgress` adds optional synchronous count reports: before inspection,
after each 128 targets, and after the last target (no duplicate final report).
Reports contain total targets, checked targets, candidates and skips, not paths or
selectable files. There are at most 783 reports for 100,000 targets. Consumers must
return promptly and marshal UI work to the UI thread. Counts can reach the total
before cancellation or final validation fails; only a returned snapshot means
completion. A slow metadata call can delay progress as well as cancellation.

Completed snapshots have unique IDs and copied immutable candidate/skip metadata.
Selections begin empty, accept only IDs in one snapshot, reject duplicates, and
report count and estimated file bytes. A rescan invalidates older selections.
Read-only revalidation checks current policy and observed identity; it is not
cleanup authority or a substitute for native descriptor/staging race proof.

Errors and skips are local only. No logs or file paths are uploaded. There is no
cleanup service, automatic retry or arbitrary-path action. Trash remains disabled.

Still pending: approved rules, native APFS/TCC/clone/cloud/active-app tests,
100,000-file real-fixture responsiveness/RSS measurements, progress for slow scans,
durable recovery, native Trash proof and separate cleanup-enable review.
