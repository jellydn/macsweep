# Security and data safety

This setup build has no scanner or cleanup service. Report any path validation,
selection, recovery, or misleading space-accounting issue as a safety issue.

If GitHub private vulnerability reporting is enabled, use the repository's
Security tab. Otherwise contact the maintainer privately at `dunghd.it@gmail.com`,
the address in Oak's maintainer history. Do not put private paths, file contents,
credentials, or exploit details in a public issue before coordination.

Only the initial development line is maintained; no stable cleanup release is
supported yet. There is no telemetry or automatic report upload. Share minimal
redacted reproduction steps and disposable fixtures voluntarily.

Safety covers ordinary filesystem replacement races. It cannot isolate a fully
compromised process running as the same user. Read `docs/safety.md` for the limits.
