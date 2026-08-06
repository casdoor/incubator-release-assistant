# Adapter contract

`internal/release` has an explicit adapter registry. The current `casbin-go`
adapter owns Casbin-specific validation; the engine owns shared state,
evidence, queue, signing, staging, and public-byte verification. The next
adapter must preserve the following contract.

An adapter provides:

- a stable adapter ID and the exact project identities it supports;
- a short operator-facing description exposed by `ira adapters`;
- canonical upstream and ASF distribution locations;
- archive naming and required-file rules;
- typed build/test commands expressed as executable plus argument vector;
- language-specific evidence interpretation.

An adapter may not:

- execute project code during signing or staging;
- accept shell command strings from project configuration;
- remove the Incubator name/disclaimer, legal, RAT, signature, checksum, KEYS,
  no-overwrite, or public-download gates;
- sign or stage files that differ from the prepared state digest.

Before registration, an adapter needs tests for valid configuration, hostile
paths, incorrect official endpoints, missing legal files, resume behavior, and
the exact host command and working directory. At least one end-to-end test must
use a local fixture repository without signing or network mutation.

The current release JSON schema remains a Casbin Go schema. Registering a
second adapter also requires the next version of the reviewed configuration
schema; do not weaken the Casbin rules merely to make an adapter name parse.
