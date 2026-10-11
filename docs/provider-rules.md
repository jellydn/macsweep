# Provider approval and current policy

Related work: [MVP umbrella #2](https://github.com/jellydn/macsweep/issues/2).

**There are no shipping provider rules.** `policy.Default` returns an empty set.
Tests inject disposable synthetic rules; they are not evidence that a real app's
cache is safe. No config import or user allowlist can widen policy.

The first rule needs owner-app/version, exact home-relative file paths,
regeneration evidence, minimum age, file/type/storage constraints, mandatory
exclusions, restoration notes, and approval in a reviewed PR. A reason string
or age alone does not prove disposability. Do not add globs or recursive directory
rules. Start with exact files below `Library/Caches/<owner>/...` only.

The current read-only inspector rejects invalid/traversal/ambiguous paths,
case/Unicode exclusion aliases, symlink components, hard links, other owners,
different devices, nonregular, young/future and sparse files. Darwin also skips
flagged files. Directory descriptors pin the descent; observed identity swaps
stop the inspection. No policy method mutates files or authorizes Trash.

Remaining native gates: APFS case/Unicode behavior, same-user replacement races,
mounted aliases, clone detection, dataless/cloud/active-app handling, TCC denial,
staging and recovery. Passing this inspector is not proof for a later move; the
cleanup path must revalidate and meet `docs/safety.md`. Cleanup stays disabled.
