# Manual package release

Read this reference when the user asks to publish a new version of a Casbin
adapter repository whose semantic-release was removed.

---

## When this reference is used

This reference is activated only when the user explicitly asks to publish a
Casbin adapter repository.  The Agent does not monitor tags or scan
repositories on its own.

The flow is **semi-automatic**: the Agent runs pre-flight checks
unattended, then stops at a confirmation preview.  The user must review
every detail and explicitly authorise the publish step.  No publish
command runs without human approval.

---

## Common pre-flight (all languages)

Execute these checks in order.  Stop and report to the user if any check
fails.

### 1. Clone the upstream canonical repository

Publishing must start from the **upstream Apache canonical repository**.
Local forks, personal clones, or working directories with uncommitted
changes are not accepted.

```bash
git clone --origin upstream https://github.com/apache/<repo>.git /tmp/ira-publish/<repo>
```

After cloning, verify the remote URL:

```bash
git -C /tmp/ira-publish/<repo> remote get-url upstream
```

The URL must match `https://github.com/apache/casbin-*.git` or
`https://github.com/apache/casbin.git`.  If it points to a personal
account or fork, **warn the user** that this is not the canonical
upstream and strongly recommend switching.  Proceed only if the user
explicitly confirms they want to continue with the non-canonical URL.

### 2. Checkout the release tag

```bash
cd /tmp/ira-publish/<repo>
git fetch upstream --tags
git checkout <tag>
```

The tag must already exist on upstream.  Do not create tags locally.

### 3. Verify the working tree is clean

```bash
git status --porcelain   # must produce NO output
```

### 4. Detect the project language

Look for one of the following files at the repository root.  The first
match determines which template to follow.  **Jump directly to that
template section and do not read the other templates.**

| Detection file | Language | Registry | Jump to |
| --- | --- | --- | --- |
| `Cargo.toml` | Rust | crates.io | `## Template: Rust (crates.io)` |
| `pom.xml` | Java | Maven Central | `## Template: Java (Maven Central)` |

If none is found, stop and report which manifest patterns are supported.

### 5. Verify the tag matches the manifest version

Read the version from the language-specific manifest (see template).
Strip common prefixes (`v`, `V`, `ver-`) from the tag and compare.
Stop on mismatch.

### 6. Verify upstream CI passed

Check the GitHub Actions tab for the upstream repository.  The most
recent workflow run on the default branch must have a `success`
conclusion.

---

## Template: Rust (crates.io)

**Detection**: `Cargo.toml` at the repository root.

**Manifest fields**:

| Field | Source |
| --- | --- |
| crate name | `Cargo.toml` → `[package] name` |
| version | `Cargo.toml` → `[package] version` |

### Step 1 — Compile and test

```bash
cargo check
cargo test
```

### Step 2 — Generate local package

```bash
cargo package
# Produces: target/package/<crate>-<version>.crate
```

Show the generated `.crate` file path and size.

### Step 3 — Dry-run publish

```bash
cargo publish --dry-run
```

Validates registry metadata and eligibility without uploading.  If it
fails because of a network timeout to crates.io, report the error and
ask the user whether to skip or retry.

### Step 4 — Confirmation preview with edit opportunity

Display the full release summary and ask the user to review.  They may
change any field before confirming:

```
=== RELEASE PREVIEW ===
  Repository:  <repo>
  Package:     <crate-name>
  Version:     <version>
  Tag:         <tag>
  Registry:    crates.io
  Artifact:    target/package/<...>.crate
  Dry-run:     passed / skipped

  ⚠️  IRREVERSIBLE.  Once published, this version
  cannot be changed.  A wrong version can only be yank-ed;
  that version number is consumed forever.

Do any fields need to change?  (version / tag / package-name)
If everything is correct, reply "CONFIRM".
```

Wait for the user.  If they want changes, apply them and re-show the
preview.  Do not proceed until they say "CONFIRM".

### Step 5 — Obtain credentials

After the user confirms, ask:

> Where can I read the `CARGO_REGISTRY_TOKEN`?
> (environment variable name, file path, or secret store reference)

Read the token **only from the location the user specifies**.  Never
hardcode it, read it from a repository file, or log it.

### Step 6 — Publish

```powershell
$env:CARGO_REGISTRY_TOKEN = "<from-user-specified-location>"
cargo publish
```

```bash
export CARGO_REGISTRY_TOKEN="<from-user-specified-location>"
cargo publish
```

### Step 7 — Report result

After a successful publish, output:

```
=== PUBLISH COMPLETE ===
Package:   <crate-name> @ <version>
Registry:  crates.io
Verify:    https://crates.io/crates/<crate-name>/<version>
```

If the publish fails, report the full error.  Do not retry
automatically — the user must decide whether to fix and re-run
(which consumes nothing) or yank a partially published artifact.

---

## Template: Java (Maven Central)

**Detection**: `pom.xml` at the repository root.

**Manifest fields**:

| Field | Source |
| --- | --- |
| group ID | `pom.xml` → `<groupId>` |
| artifact ID | `pom.xml` → `<artifactId>` |
| version | `pom.xml` → `<version>` |

### Step 1 — Compile and test

```bash
mvn clean verify
```

### Step 2 — Generate local package

```bash
mvn clean package -DskipTests
# Produces: target/<artifact>-<version>.jar
```

> Maven has no true dry-run for `mvn deploy`.  `mvn verify` + `mvn
> package` together provide the closest safe pre-flight.

### Step 3 — Confirmation preview with edit opportunity

Display the full release summary and ask the user to review:

```
=== RELEASE PREVIEW ===
  Repository:   <repo>
  Group ID:     <group-id>
  Artifact ID:  <artifact-id>
  Version:      <version>
  Tag:          <tag>
  Registry:     Maven Central
  Build:        passed

  ⚠️  IRREVERSIBLE.  No programmatic rollback.
  Once deployed, the artifact cannot be deleted or overwritten.

Do any fields need to change?  (version / tag / group-id / artifact-id)
If everything is correct, reply "CONFIRM".
```

Wait for the user.  If they want changes, apply them and re-show the
preview.  Do not proceed until they say "CONFIRM".

### Step 4 — Obtain credentials

After the user confirms, ask:

> Where can I read the OSSRH token and GPG signing key?
> (environment variable names, file paths, or secret store reference)

Read credentials **only from the location the user specifies**.

### Step 5 — Deploy

```powershell
mvn deploy -s maven-settings.xml
```

```bash
mvn deploy -s maven-settings.xml
```

### Step 6 — Report result

After a successful deploy, output:

```
=== PUBLISH COMPLETE ===
Artifact:  <group-id>:<artifact-id> @ <version>
Registry:  Maven Central
Verify:    https://central.sonatype.com/artifact/<group-id>/<artifact-id>/<version>
```

If the deploy fails, report the full error.  Do not retry automatically.

---

## Cleanup

After a publish completes (success or failure), clean up the temporary
workspace:

```bash
rm -rf /tmp/ira-publish/<repo>
```
