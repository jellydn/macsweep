# 100,000-file measurement

This records Linux and Apple Silicon headless fixtures plus limited native Mac
observations. It is not Intel acceptance or cleanup proof. Native cancellation
under one second remains unverified. Cleanup and release remain blocked.

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

Run the harness on Intel and record hardware, macOS, APFS mode,
fixture location and raw output. Separately measure the production app's idle RSS,
Metal frame/input response and cancel-button-to-result latency, including denied
paths, active apps and cloud placeholders. Do not use this synthetic rule in a
shipping build. Test the real empty-provider app now; approved provider and native
safety evidence are still needed before claiming real scan or cleanup acceptance.

## Apple Silicon results — 2026-10-13

Runner `itman-mac27-app`: Mac14,10, Apple M2 Pro, 12 CPUs, macOS 27.0.1
(26A434), Go 1.27.2, GOMAXPROCS=12. `uname -m` returned `arm64` and
`sysctl -n hw.optional.arm64` returned `1`. This is actual arm64 execution,
not an inference from a universal binary. The Data volume is APFS; a disposable
case probe resolved differently cased names to the same inode (case-insensitive).
The harness created its own homes under the macOS temporary directory.

Three separate processes tested the complete stack with the exclusion fixes,
MyGo v0.4.2 and x/text v0.42.0. No real caches or Trash were changed.

| Measurement | Run 1 | Run 2 | Run 3 |
| --- | ---: | ---: | ---: |
| Fixture + policy (ms) | 9268.704 | 8997.486 | 9144.022 |
| Metadata scan (ms) | 7837.873 | 7869.287 | 7922.722 |
| Cancellation return (ms) | 0.041125 | 0.051459 | 0.030834 |
| Preview handoff + headless frame (ms) | 15.227 | 10.764 | 11.453 |
| Idle headless RSS (KiB) | 33520 | 33904 | 33472 |
| Retained preview RSS (KiB) | 261936 | 261440 | 261424 |
| Process peak RSS (KiB) | 262480 | 261936 | 261472 |
| Progress reports | 783 | 783 | 783 |

All three runs passed the 100,000-candidate, 150,000-byte, zero-skip, empty-selection
and cancelled-no-snapshot assertions. Retained preview memory is about 255–256 MiB;
do not hide this cost or compare it with the separate native idle-app target.
These warm-fixture samples do not cover blocked syscalls or native cancel input.

The no-cgo native app launched, rendered its dark preview and exclusion editor,
and an executed Scan Mac action ended with no approved rules, zero selection and
disabled Trash. One ordinary native idle sample after the empty scan was 101856
KiB (about 99.5 MiB), without forced GC. `lsof -nP -a -p <pid> -i` returned no
Internet sockets at that instant; this is not continuous offline/network proof.
Darwin race tests passed policy aliases, links, sparse files, immutable exclusions,
snapshot cancellation and selection, private preferences and disabled UI controls.
Both no-cgo Darwin architectures built and universal ad-hoc packaging passed;
the amd64 binary was not executed. The combined sequence with x/text v0.43.0
also passed the Mac race suite. These results approve no real provider rule.

## Obtain Intel evidence

Intel has **not** been tested in this session. Use a graphical Intel Mac for
AppKit/Metal, keyboard/VoiceOver, TCC, APFS variants, native cancellation and
staging/Trash/recovery acceptance. Record `uname -m`, `sw_vers`, hardware and
`go env GOARCH`; execute the same check script and three harness runs above at
the exact reviewed commit. Rosetta on this arm64 Mac is not an Intel Mac test.

For the unit/race suite and headless fixture only, GitHub provides
[`macos-15-intel` and `macos-26-intel` runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners).
A follow-up reviewed CI change can select an explicit Intel label, assert
`uname -m` is `x86_64` and `go env GOARCH` is `amd64`, run `scripts/check.sh`
and three harness processes, then retain raw JSON and machine metadata as
artifacts. Current `macos-latest` plus cross-builds does not supply that evidence.
No Intel job was added or run here. Minimum-macOS and the other native safety
gates remain open on both architectures.
