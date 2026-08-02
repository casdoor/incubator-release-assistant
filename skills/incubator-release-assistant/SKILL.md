---
name: incubator-release-assistant
description: Prepare, verify, sign, stage, resume, and document Apache Casbin Go Incubator source release candidates with the bundled IRA engine. Use this skill whenever a user asks about making or validating a Casbin RC, source archive, RAT/legal checks, SHA-512, GPG/KEYS, ASF dist dev staging, or dev/general release votes. The current implementation supports only Apache Casbin Go; do not imply that other repositories or languages are executable yet.
---

# Incubator Release Assistant

Use the bundled engine instead of reconstructing release commands manually. It
encodes exact-byte state, sandboxed Go tests, signing and staging confirmations,
official endpoint checks, and public re-verification.

## Establish the run

1. Resolve this Skill directory and the user's workspace.
2. Read `references/configuration.md` when creating or changing config.
3. Copy `assets/examples/casbin-go.json` to an ignored project-local path if no
   active configuration exists. Never fill Apache ID, commit, or fingerprint by
   guessing; obtain them from the user or verified public state.
4. Read `references/apache-release-gates.md` before declaring a candidate ready,
   preparing vote text, or interpreting a failure.
5. Resolve the configuration to an absolute path, then run the engine from
   `scripts/ira`:

```powershell
go -C <skill-directory>\scripts\ira run .\cmd\ira validate --config <config>
go -C <skill-directory>\scripts\ira run .\cmd\ira plan --config <config>
```

Stop on placeholders, unsupported adapters, noncanonical endpoints, missing
policy files, unresolved provenance, or config that does not validate.

## Prepare in the untrusted-code domain

Run:

```powershell
go -C <skill-directory>\scripts\ira run .\cmd\ira prepare --config <config>
```

The engine builds the exact-commit archive, runs RAT, and executes `go test
./...` in Docker/Podman with only the disposable extracted source mounted. Do
not reproduce project tests on the host or add credential mounts. A successful
run prints the artifact SHA-512 and records resumable state under `.ira/runs/`.

If preparation fails, report the failed gate and evidence path. Use `--clean`
only for an unstaged matching run after confirming its local work is disposable.
Changed candidate bytes require a new RC number.

## Sign in the trusted domain

Signing is a separate authority boundary. Show the prepared artifact path and
SHA-512 to the user. Proceed only when the user explicitly authorizes signing,
then pass the exact digest printed by `prepare`:

```powershell
go -C <skill-directory>\scripts\ira run .\cmd\ira sign `
  --config <config> --confirm <exact-sha512>
```

The engine re-hashes the artifact, checks the configured private key, project
UID policy, RSA strength, and official KEYS, then creates and independently
verifies the detached signature. Never request or store a passphrase.

## Stage and publicly verify

ASF dist mutation requires a second explicit user authorization. Display the
target URL and exact confirmation text, then run only after approval:

```powershell
go -C <skill-directory>\scripts\ira run .\cmd\ira stage `
  --config <config> --confirm "STAGE RC<number>"
```

The engine refuses an existing RC directory, disables SVN credential caching,
stages only the archive/signature/checksum, and re-downloads public files for
byte, checksum, and signature verification. Do not prepare vote text until
`publicVerified` is true.

## Votes and records

Apache Incubator releases require both phases:

1. podling dev vote;
2. Incubator general vote after a successful dev vote.

Each vote stays open at least the configured 72 hours. Archive exact draft and
sent text in the user's project record, not in this tool repository. Separate
verified facts, binding vote counts, assumptions, and pending human decisions.

## Boundaries

- Current executable adapter: `casbin-go` only.
- Do not execute arbitrary commands from JSON or introduce a host-test fallback.
- Do not store passwords, tokens, cookies, private keys, or private-list text.
- Treat legal, provenance, RAT, test, checksum, signature, KEYS, no-overwrite,
  and public-download failures as hard stops.
- Signing, ASF writes, vote sending, and announcements always require explicit
  human authorization.
- Consult `legacy/casbin-go-rc/` only as repository history; the bundled Go
  engine is the operational implementation.
