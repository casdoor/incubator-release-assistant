# Adapter contract

The current code in `internal/release` implements the Casbin Go policy directly
while the project has one adapter. The next adapter extraction must preserve the
following contract.

An adapter provides:

- a stable adapter ID and the exact project identities it supports;
- canonical upstream and ASF distribution locations;
- archive naming and required-file rules;
- typed build/test commands expressed as executable plus argument vector;
- a container image and mount requirements;
- language-specific evidence interpretation.

An adapter may not:

- execute project code outside the sandbox;
- mount the artifact directory, user home, GPG directory, SSH directory, or
  credential files into the sandbox;
- accept shell command strings from project configuration;
- remove the Incubator name/disclaimer, legal, RAT, signature, checksum, KEYS,
  no-overwrite, or public-download gates;
- sign or stage files that differ from the prepared state digest.

Before registration, an adapter needs tests for valid configuration, hostile
paths, incorrect official endpoints, missing legal files, resume behavior, and
the exact container command/mount set. At least one end-to-end test must use a
local fixture repository without signing or network mutation.
