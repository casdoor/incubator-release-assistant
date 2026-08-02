# Configuration reference

The standalone Skill bundles `assets/release.schema.json` and
`assets/examples/casbin-go.json`. In the full repository, identical mirrors live
under `config/`. Copy the example to an ignored local path; never edit the
reviewed template in place.

## Current contract

- `schemaVersion`: must be `2`.
- `project`: fixed Casbin identity, `casbin-go` adapter, and incubation status.
- `source`: canonical Apache Casbin repository, full upstream commit, and safe
  archive prefix.
- `release`: incubating version, positive RC number, and derived artifact name.
- `checks`: required legal/Go files and reviewed Apache RAT version.
- `signing`: public Apache ID/fingerprint, official KEYS URL, project UID policy,
  and RSA minimum. These identifiers are public; private keys/passphrases are not.
- `distribution`: fixed official Casbin incubator dev/release locations.
- `votes`: fixed podling dev and Incubator general lists, minimum 72 hours.
- `runtime`: ignored state directory and reviewed Docker/Podman Go sandbox.

## Release-ready semantics

JSON Schema describes the template shape. The bundled Go validator additionally
enforces cross-field and release-ready rules:

- all-zero commit/fingerprint and replacement Apache IDs are rejected;
- archive prefix and artifact base must exactly match the version;
- `LICENSE`, `NOTICE`, one disclaimer, `go.mod`, `go.sum`, and `.rat-excludes`
  cannot be removed;
- path traversal, unknown fields, alternate repositories, alternate ASF URLs,
  alternate vote lists, arbitrary commands, and unsupported adapters are
  rejected.

Configuration contains no shell command. `go test ./...` belongs to the trusted
adapter implementation and runs only in the container.

## Values humans normally change

1. `source.commit` — verified full commit from canonical upstream.
2. `release.version` and `release.rc`.
3. `source.archivePrefix` and `release.artifactBaseName`, derived as
   `apache-casbin-<version>-src`.
4. `signing.apacheId` and public full fingerprint.
5. Container engine (`docker` or `podman`) when the reviewed local environment
   requires it.

Never place passwords, private keys, tokens, cookies, SSH material, or private
mail in configuration.
