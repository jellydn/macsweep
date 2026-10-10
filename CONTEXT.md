# MacSweep domain glossary

- **Provider rule:** a reviewed, versioned rule for exact disposable file paths.
  A folder called `Cache` is not a provider rule.
- **Candidate:** a file observed in a scan. Observation does not authorize a move.
- **Eligible file:** a regular file that passes a provider rule and all safety checks.
- **Scan snapshot:** frozen candidates, identities, policy version, and scan status.
- **Selection:** candidate IDs chosen by the user from one completed snapshot.
- **Cleanup:** confirmed movement of revalidated selected files to Trash, not deletion.
- **Staging:** a private, same-volume recovery location used before native Trash.
- **Unknown outcome:** a native operation whose completion cannot be verified.
  Do not label it success or repeat it automatically.
- **Bytes moved:** file-size estimate for acknowledged Trash moves.
- **Observed free-space change:** measured volume capacity delta, not proof that
  MacSweep freed that amount. A Trash move normally frees no disk space.
- **Excluded path:** a location that scan and cleanup must both refuse.
