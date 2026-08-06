# Configuration reference

The standalone Skill bundles `assets/release.schema.json` and
`assets/examples/casbin-go.json`. In the full repository, identical mirrors live
under `config/`. Copy the example to an ignored local path; never edit the
reviewed template in place.

## Current contract

- `schemaVersion`: must be `4`. Version 4 removes the Docker/Podman runtime
  configuration.
- `project`: fixed Casbin identity, `casbin-go` adapter, and incubation status.
- `source`: canonical Apache Casbin repository, full upstream commit, and safe
  archive prefix.
- `release`: incubating version, positive RC number, and derived artifact name.
- `checks`: required legal/Go files and reviewed Apache RAT version.
- `signing`: public Apache ID/fingerprint, official KEYS URL, project UID policy,
  and RSA minimum. These identifiers are public; private keys/passphrases are not.
- `distribution`: fixed official Casbin incubator dist-dev staging location.
- `runtime`: ignored state directory.

## Release-ready semantics

JSON Schema describes the template shape. The bundled Go validator additionally
enforces cross-field and release-ready rules:

- all-zero commit/fingerprint and replacement Apache IDs are rejected;
- archive prefix and artifact base must exactly match the version;
- `LICENSE`, `NOTICE`, one disclaimer, `go.mod`, `go.sum`, and `.rat-excludes`
  cannot be removed;
- path traversal, unknown fields, alternate repositories, alternate ASF URLs,
  arbitrary commands, and unsupported adapters are rejected.

Configuration contains no shell command. IRA does not execute target-project
tests during `prepare`; confirm the exact selected commit's GitHub CI separately.

## Values humans normally change

1. `source.commit` — verified full commit from canonical upstream.
2. `release.version` and `release.rc`.
3. `source.archivePrefix` and `release.artifactBaseName`, derived as
   `apache-casbin-<version>-src`.
4. `signing.apacheId` and public full fingerprint.

Never place passwords, private keys, tokens, cookies, SSH material, or private
mail in configuration.

## Ordered release queue

`assets/release-queue.schema.json` and
`assets/examples/casbin-release-queue.json` define an ordered worklist for
several repositories. Copy both the queue and each referenced release config to
the same ignored local directory. `releaseConfig` is a safe relative JSON path
resolved from the queue file's directory.

Each item records its repository identity, adapter, and one of four states:

- `queued`: has a release config and can be assessed against local IRA state;
- `blocked`: requires a non-empty explanation before the queue can continue;
- `manual`: records work IRA must not perform and requires a non-empty note;
- `complete`: a reviewed external completion record with a non-empty note.

`queue-status` reports the first incomplete item as current. It derives the
next action from the matching release state: `prepare`, `sign`, `stage`, or
complete. `queue-prepare` only prepares that current item; signing and staging
remain individual, explicitly confirmed commands. The queue does not accept
shell commands and does not make an unsupported adapter executable.

## External secret directory

The secret location is deliberately not a JSON field. Start the Agent from the
parent workspace and let the platform wrapper use
`<current-directory>/secretkey`, or pass `-SecretDirectory` on PowerShell and
`--secret-dir` on Bash. During `sign`, the wrapper exports an absolute
`IRA_SECRET_DIR` and sets `GNUPGHOME` to the same directory before invoking the
engine. Other stages do not inspect or create the secret directory.

For the standard deployment this resolves to `/abc/secretkey` while the
repository remains `/abc/Incubator-release-assistant`. The engine rejects a
missing, relative, repository-contained, or larger-Git-worktree-contained
secret directory. The release JSON contains only the public fingerprint and
Apache ID; no key bytes or passphrase.
