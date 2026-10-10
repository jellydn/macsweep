# Setup validation — 2026-10-10

Executed in a Linux x64 Amp orb with Go 1.27.2. This is not macOS runtime proof.

| Check | Result |
| --- | --- |
| `.agents/setup`, twice | Both runs passed; a clean login shell resolves Go 1.27.2. |
| `./scripts/check.sh` | Modules verified; `gofmt` gate, `go vet`, race tests passed. |
| Disabled-control regression | Passed; temporarily enabling Move to Trash failed with “setup control … is active”. Reverted the mutation. |
| Headless render | Inspected 2× light, dark, and 520×360 renders; readable text, disabled buttons, no clipping. Temporary render test removed. |
| `actionlint` v1.7.12 | Both workflows passed, including the declared custom release-runner label. |
| `shellcheck` and `bash -n` | Setup, check, and release scripts passed. |
| `govulncheck` v1.8.0 | “No vulnerabilities found.” This is not a general security certification. |
| `go run ./cmd/macsweep --version` | “MacSweep development”; no app window started. |
| MyGo production universal build, cgo disabled | Succeeded; `file` reports arm64 and x86_64 slices. Inspector build tag confirmed in Go build metadata. |
| Bundle inspection | Version 0.1.0, identifier, declared macOS 13 floor, license/notices/resources verified. |
| Release script on Linux | Refused with “Release preparation requires macOS.” No signing or Apple submission attempted. |

Not executed: native Mac launch, Metal rendering, VoiceOver, TCC, actual file
scan/Trash/recovery, real socket observation, Developer ID signing, notarization,
stapling, quarantined/offline launch, or GitHub Actions. Scanner and cleanup do
not exist yet. Native target compatibility and the release pipeline remain gates.

GitHub repository creation failed because the current token cannot create
repositories. Setup remains local; origin is configured for
`https://github.com/jellydn/macsweep.git`. No push, release, environment setting,
runner provisioning, or Slack message was made.
