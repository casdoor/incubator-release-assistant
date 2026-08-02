# Manual package release

Read this reference when the user asks to publish a new version of a Casbin
adapter repository whose semantic-release was removed. Publishing to a public
package registry is irreversible and always requires explicit human
authorization.

## Supported repositories

| Repository | Language | Registry | Publish command |
| --- | --- | --- | --- |
| `casbin-sqlx-adapter` | Rust | crates.io | `cargo publish` |
| `casbin-jcasbin-jdbc-adapter` | Java | Maven Central | `mvn deploy` |

## Pre-flight checks

1. The release tag exists upstream and matches the version declared in
   `Cargo.toml` or `pom.xml`.
2. Upstream CI (`.github/workflows/ci.yml` or `maven-ci.yml`) passed on
   `master`.
3. The local working tree is clean and checked out at the tagged commit.

## Rust: publish to crates.io

Credentials come from the `CARGO_REGISTRY_TOKEN` environment variable or an
external secret store. Never put the token in configuration or the repository.

```powershell
$env:CARGO_REGISTRY_TOKEN = "<from-secret-store>"
cargo publish --dry-run
cargo publish
```

```bash
export CARGO_REGISTRY_TOKEN="<from-secret-store>"
cargo publish --dry-run
cargo publish
```

Verify the published version at:
`https://crates.io/crates/sqlx-adapter/<version>`.

## Java: deploy to Maven Central

Credentials (OSSRH token and GPG signing key) come from environment variables
or external secret storage. `maven-settings.xml` must exist and point to the
`ossrh` server.

```powershell
mvn deploy -s maven-settings.xml
```

```bash
mvn deploy -s maven-settings.xml
```

Verify the published version at:
`https://central.sonatype.com/artifact/org.casbin/jdbc-adapter/<version>`.

## Failure and rollback

- crates.io and Maven Central versions are immutable once published. Do not
  blindly retry a failed publish; report the error first.
- For a mistaken crates.io release, use `cargo yank` to unlist the version.

## Agent response

Keep the response short: the version, the registry URL, where the credentials
came from, and the one next action. Never print credentials or private keys.
