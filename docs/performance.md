# 100,000-file measurement

This is Linux/headless evidence, not Intel or Apple Silicon Mac acceptance.
The native idle-app RSS target of 150 MiB and native cancellation target of one
second remain unverified. Cleanup and release remain blocked.

## Reproduce

Run as a non-root user with Go and `ps` available:

```sh
CGO_ENABLED=0 go build -tags mygo_noinspector -o /tmp/macsweep-perf ./scripts/perf
for n in 1 2 3; do /tmp/macsweep-perf > "/tmp/macsweep-perf-$n.json"; done
```

The developer-only harness creates and removes its own temporary home. It writes
100,000 regular files in 100 directories, alternating one and two bytes, with
mtime 48 hours old. The synthetic rule is not a shipping provider. Fixture setup
and policy construction are timed separately from the scan. Metadata is likely
warm after fixture creation; no caches are flushed and no privileges are needed.

Each run asserts 100,000 candidates, 150,000 bytes, zero skips, 783 progress
reports, empty initial selection and no snapshot after cancellation. Cancellation
is requested by a separate goroutine 10 ms after the scanner passes 128 targets,
not inside the progress callback. Reported latency starts immediately before the
context cancellation request and ends when the scan returns. A blocked native
metadata syscall is not covered by this fixture.

The 640×480 logical headless view uses 2× scale. Preview time includes copying the
snapshot into the model and settling its first frame. Scanner timing does not
include native UI dispatch or progress rendering. The report count bounds calls
to `Window.Update`, not native frames or event-loop latency.

RSS samples use `ps -o rss=` in KiB after explicit Go GC/scavenging, with the view
kept live. Idle is sampled before fixture setup; preview RSS retains policy,
snapshot and rendered view. Peak RSS comes from `getrusage` and includes fixture
creation, scan, preview and cancellation (before fixture removal). Forced GC makes
these retained-memory samples different from ordinary app idle behavior.

## Recorded results — 2026-10-11

E2B orb: Linux 6.1.158+, amd64, ext4 `/dev/root`, virtual Intel Xeon @ 2.60 GHz,
8 visible CPUs, GOMAXPROCS=8, Go 1.27.2. Three separate process invocations, using
the implementation above #10/#11 plus exclusion preferences and selection UI.

| Measurement | Run 1 | Run 2 | Run 3 |
| --- | ---: | ---: | ---: |
| Fixture + policy (ms) | 3629.55 | 3837.49 | 3668.46 |
| Metadata scan (ms) | 3782.52 | 4330.18 | 4111.60 |
| Cancellation return (ms) | 0.0422 | 0.0185 | 0.0169 |
| Preview handoff + frame (ms) | 179.96 | 11.04 | 22.34 |
| Idle headless RSS (KiB) | 25336 | 24460 | 25064 |
| Retained preview RSS (KiB) | 131904 | 131712 | 131712 |
| Process peak RSS (KiB) | 239240 | 228228 | 227520 |
| Progress reports | 783 | 783 | 783 |

All runs returned 100,000 candidates, 150,000 bytes and zero skips. The preview
started with zero selected files. Cancellation returned `context.Canceled` and
no snapshot. No timing threshold is enforced by this harness: these are samples,
not a worst-case bound. The 180 ms first-run preview cost and 222–234 MiB process
peak remain visible performance costs, not hidden by averaging.

## Remaining native measurements

Run the harness on both Mac architectures and record hardware, macOS, APFS mode,
fixture location and raw output. Separately measure the production app's idle RSS,
Metal frame/input response and cancel-button-to-result latency, including denied
paths, active apps and cloud placeholders. Do not use this synthetic rule in a
shipping build. Test the real empty-provider app now; approved provider and native
safety evidence are still needed before claiming real scan or cleanup acceptance.
