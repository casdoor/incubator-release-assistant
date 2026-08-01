# Architecture

## Goal

Support many Apache Incubator repositories without maintaining one release
script per repository. Repository facts are data; release gates are reusable
policy; language-specific commands are adapters.

## Layers

```text
Human-reviewed JSON configuration
             |
             v
Agent Skill: plan, explain, stop, and request confirmation
             |
             v
Release engine: archive, evidence, checksum, signing, dist
             |
             v
Language/project adapters: build, test, and package-specific checks
```

### Configuration

The configuration records the exact repository, commit, release identity,
required files, validation commands, signing identity, ASF URLs, and vote
destinations. It contains no secret values.

### Skill

The Skill resolves the configuration, selects the applicable checks, explains
gates in plain language, records evidence, and stops when required information
or human authority is missing.

### Engine

The future generic engine owns operations that should not vary by language:

- creating a clean source archive from an exact commit;
- enforcing one top-level archive directory;
- running declared checks and preserving evidence;
- producing LF-only SHA-512 files;
- verifying GPG signatures against official `KEYS`;
- preventing RC directory overwrite;
- re-downloading public ASF dist files and comparing their bytes.

### Adapters

Adapters contribute build and test commands plus repository-specific
classifications. Initial targets are Go, Java, Node.js, Python, Rust, and .NET.
An adapter must not weaken the shared legal, signature, checksum, or vote gates.

## Migration strategy

`legacy/casbin-go-rc/` is the executable baseline. Migrate it gate by gate:

1. move hard-coded Casbin names and URLs into configuration;
2. move `go.mod`, `go.sum`, and `go test ./...` into a Go adapter;
3. retain the existing checksum, GPG, `KEYS`, SVN, and public-download checks;
4. add resumable state so prepared artifacts can be staged without rebuilding;
5. add adapter contract tests before adding more languages.

Until this migration is complete, do not describe the legacy script as a
repository-neutral release engine.
