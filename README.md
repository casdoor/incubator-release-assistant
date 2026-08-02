# Incubator Release Assistant (IRA)

IRA is an unofficial, configuration-driven release assistant and self-contained
Agent Skill for Apache Incubator source releases.

The current implementation deliberately supports **Apache Casbin Go only**.
The engine has an adapter boundary so later languages and repositories can be
added without weakening the shared release gates.

> IRA is a community tool. ASF policy and the required human votes remain the
> authority. Signing, ASF writes, votes, and announcements are never fully
> automated without human approval.

## What is implemented

- strict, unknown-field-rejecting release-ready configuration validation;
- canonical Casbin upstream, ASF URL, legal-file, disclaimer, vote, and naming
  invariants that configuration cannot turn off;
- exact-commit `git archive` packaging with one top-level directory;
- Apache RAT verification and canonical LF-only SHA-512 output;
- `go test ./...` inside Docker or Podman with only a disposable extracted
  source tree mounted—the host artifact directory and credentials are absent;
- separate `prepare`, `sign`, and `stage` trust boundaries;
- signing-key and official `KEYS` verification in an isolated public keyring;
- no-overwrite ASF dist dev staging and public byte-for-byte re-verification;
- resumable state under ignored `.ira/runs/`;
- a self-contained Skill containing the Go engine, schema, and Casbin template.

## Prerequisites

- Go 1.22 or newer;
- Git, tar, Java, GnuPG, SVN;
- Docker or Podman for running repository code without access to host secrets.

The container engine is a security boundary. IRA intentionally has no
"run project tests directly on the signing host" fallback.

## Human workflow

Copy and fill the non-secret template:

```powershell
New-Item -ItemType Directory -Force .\config\local | Out-Null
Copy-Item .\config\examples\casbin-go.json .\config\local\casbin.local.json
notepad .\config\local\casbin.local.json
```

Only the commit, version/RC-derived names, Apache ID, signer fingerprint, and
reviewed runtime choice normally need attention. Passwords, private keys,
tokens, and cookies never belong in JSON.

Then run:

```powershell
.\ira.ps1 validate -Config .\config\local\casbin.local.json
.\ira.ps1 plan -Config .\config\local\casbin.local.json
.\ira.ps1 prepare -Config .\config\local\casbin.local.json
```

`prepare` prints the exact SHA-512 required for the next explicit boundary:

```powershell
.\ira.ps1 sign -Config .\config\local\casbin.local.json `
  -Confirm <exact-128-character-sha512>

.\ira.ps1 stage -Config .\config\local\casbin.local.json `
  -Confirm "STAGE RC2"
```

Staging automatically re-downloads the public candidate and verifies its
archive bytes, checksum, and signature. Re-running `prepare`, `sign`, or
`stage` resumes and revalidates completed work instead of rebuilding it.

On Linux/macOS, use the same subcommands through `bash ./ira.sh`.

## Safety model

The release is split into three trust domains:

1. `prepare` handles untrusted repository code. Go tests run in a container
   which receives only the extracted source tree. It cannot see release
   artifacts, the user's home directory, GPG keyring, or ASF credentials.
2. `sign` executes no repository code. It re-hashes the prepared artifact and
   requires the human to repeat that exact digest before accessing GPG.
3. `stage` revalidates all three release files, requires `STAGE RC<n>`, refuses
   existing remote RC directories, avoids credential caching, and verifies the
   public copy.

The configured Go image is restricted to the reviewed `golang:1.24` image. IRA
records complete command output as private local evidence. Docker/Podman itself
is privileged software and must come from a trusted installation.

## Repository layout

```text
config/                         Human-facing schema and template
docs/                           Architecture, security, and adapter contracts
legacy/casbin-go-rc/            Preserved pre-engine PowerShell baseline
skills/incubator-release-assistant/
  SKILL.md                      Agent workflow
  assets/                       Self-contained schema and template mirrors
  scripts/ira/                  Go CLI, engine, and tests
scripts/sync-skill-assets.ps1   Keeps Skill assets identical to root config
ira.ps1 / ira.sh                Human-friendly entry points
```

## Current boundary and extension path

The only registered adapter is `casbin-go`; other project IDs, repositories,
ASF destinations, and adapters are rejected. To add another ecosystem, first
implement and test a new adapter following [docs/adapter-contract.md](docs/adapter-contract.md).
Do not add arbitrary shell commands or project-name conditionals to config.

The old PowerShell flow remains under `legacy/` as migration evidence, not as
the default engine.

## License

Apache License 2.0. See [LICENSE](LICENSE).
