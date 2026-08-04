# Configuration reference

The standalone Skill bundles `assets/release.schema.json` and
`assets/examples/casbin-go.json`. In the full repository, identical mirrors live
under `config/`. Copy the example to an ignored local path; never edit the
reviewed template in place.

## Current contract

- `schemaVersion`: must be `4`. Version 4 removes the Docker/Podman runtime
  configuration and uses the installed host Go toolchain.
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

Configuration contains no shell command. `go test ./...` belongs to the trusted
adapter implementation and runs from the disposable extracted source tree with
the installed host Go toolchain.

## Values humans normally change

1. `source.commit` — verified full commit from canonical upstream.
2. `release.version` and `release.rc`.
3. `source.archivePrefix` and `release.artifactBaseName`, derived as
   `apache-casbin-<version>-src`.
4. `signing.apacheId` and public full fingerprint.

Never place passwords, private keys, tokens, cookies, SSH material, or private
mail in configuration.

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
