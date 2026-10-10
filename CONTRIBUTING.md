# Contributing

Read `CONTEXT.md`, `docs/safety.md`, and `DEVELOPMENT.md`. Keep changes small and
describe user-visible behavior and the tests run in each PR.

Use conventional commit messages. Run `./scripts/check.sh` before review.
For new safety behavior, add a public-interface regression test with an input
that a plausible unsafe implementation would mishandle. Isolate file fixtures
with `t.TempDir()` and record native Mac evidence separately from headless tests.

Provider-rule PRs must identify the owner app, exact relative paths, regeneration
evidence, age/type constraints, exclusions, restore instructions, and tests.
Do not import another cleaner's allowlist. Age and app classification are not
proof of disposable content. User exclusions may narrow rules, never widen them.

Do not add an updater, telemetry, privilege escalation, permanent deletion, or
automatic release as convenience features. See `RELEASES.md` for the release gate.
