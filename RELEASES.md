# Release preparation

**Blocked for this setup build.** No signed or notarized release has been tested.
The workflow does not publish, create tags, bump versions, or push commits.

## Before enabling the release gate

1. Implement and review the safety contract and all acceptance groups in
   `docs/implementation.md`. Approve exact shipping provider rules.
2. Test minimum macOS, latest stable macOS, Intel, and Apple Silicon. Include
   real Trash/recovery, TCC denial, accessibility, zero app networking, and performance.
3. Verify license ownership, third-party notices, and final icon/resources.
4. Name a release owner and provision a dedicated trusted Mac runner with the
   `macsweep-release` label. Do not attach this runner to PR jobs.
5. Configure the `release` environment with required human reviewers. Grant
   access only to release maintainers. None of these settings are configured yet.

The release Mac needs Go, `jq`, Xcode notarization tools, an unlocked keychain
with a Developer ID Application certificate, and a `notarytool` credential
profile. Provision credentials through the owner's secure workflow; do not
commit, print, or pass private keys through chat. Oak's ad-hoc release signing
and Sparkle payload signatures are not substitutes for Developer ID/notarization.

After the checklist is approved, set environment variables:

- `MACSWEEP_RELEASE_READY=true`
- `MACSWEEP_SIGNING_IDENTITY`: Developer ID Application certificate identity
- `MACSWEEP_NOTARY_PROFILE`: local keychain profile name

Do not set the readiness variable for a setup-only app. Keep credentials on the
trusted release account's keychain, not on untrusted PR runners.

## Manual preparation

Commit `mygo.json` version first. Review that commit, then create an approved
`vX.Y.Z` tag for it. Dispatch **Prepare signed release artifacts** with that
existing tag and its full commit SHA. The script checks exact checkout/tag
identity, main ancestry, clean source, version agreement, tests, production
build, both architectures, Developer ID, hardened runtime, notarization and
stapled tickets. Checksums are made from the final artifacts.

The workflow uses refs fetched at checkout. Before publication, the owner must
recheck that the remote tag still resolves to the approved SHA; no remote tag
mutation or overwrite is permitted. For a private repository, the release runner
also needs normal authenticated checkout access.

Review the notarization log for warnings and assess the downloaded quarantined
artifacts. Test first launch offline on fresh Mac accounts; keep macOS Gatekeeper
traffic separate from app traffic. Verify `SHA256SUMS` against the final downloads.
Publish only after a separate explicit approval. There is no runtime updater.
