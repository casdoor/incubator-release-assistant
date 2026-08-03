# Repository guidance

When the user asks to set up or perform an Apache Casbin release, read and
follow `skills/incubator-release-assistant/SKILL.md` before acting. Run the
read-only `doctor` command first and use its reported reference and next action.
The remaining guidance in this file also applies to repository maintenance.

## Purpose and current scope

Maintain a safe, configuration-driven Apache Incubator release engine and
self-contained Agent Skill. The current executable scope is Apache Casbin Go
only. Do not claim support for another project/language until its adapter and
contract tests exist.

## Commands

```powershell
# Human entry point from the parent workspace (for example C:\abc)
.\Incubator-release-assistant\ira.ps1 doctor
.\Incubator-release-assistant\ira.ps1 validate `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json
.\Incubator-release-assistant\ira.ps1 plan `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json
.\Incubator-release-assistant\ira.ps1 prepare `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json

# Repository-maintenance commands, still from the parent workspace
$repository = ".\Incubator-release-assistant"
$engine = Join-Path $repository "skills\incubator-release-assistant\scripts\ira"
go -C $engine test ./...
go -C $engine vet ./...

# Repository and Skill contract
& (Join-Path $repository "scripts\validate-repository.ps1")
```

After changing root schema/example files, run
`Incubator-release-assistant/scripts/sync-skill-assets.ps1` and commit both
mirrors.

## Commit and pull-request convention

Use Conventional Commits for every non-merge commit subject and PR title:
`type[(scope)]: summary`. Allowed default types are `feat`, `fix`, `docs`,
`style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, and `revert`.
Examples: `feat: add prepare progress logs` and
`docs: document semantic commit rules`. Before pushing, inspect every subject
in the PR range and make the PR title semantic as well; squash merges use that
title. Do not rely on an empty follow-up commit to repair a non-semantic change.

## Non-negotiable boundaries

- Project tests execute from the disposable extracted source tree with the fixed
  adapter command `go test ./...`; never accept project commands from JSON.
- `prepare` must not invoke GPG or ASF-authenticated commands. Host tests run
  with the current user's permissions, so only use a reviewed upstream commit.
- The Agent runs from the parent workspace. Keep the checkout and secret root as
  siblings (`/abc/Incubator-release-assistant` and `/abc/secretkey`); never
  weaken the wrapper/engine rejection of repository-contained keys.
- Signing and staging remain separate commands with exact human confirmations.
- Configuration contains no arbitrary shell command.
- ASF legal, disclaimer, RAT, checksum, signature, KEYS, no-overwrite,
  and public-download gates are engine policy, not optional project data.
- A staged RC cannot be cleaned or overwritten; changed bytes require a new RC.
- Treat `legacy/casbin-go-rc/` only as historical migration evidence.

## Repository hygiene

Never commit active local configuration, `.ira/`, artifacts, evidence, private
keys, passwords, tokens, cookies, credential stores, or private-list content.
Public Apache IDs and signing fingerprints may appear only where operationally
necessary. Run an available secret scan before publishing changes that touch
release configuration, key guidance, or evidence handling, and inspect staged
changes before committing.

When adding an adapter, follow `docs/adapter-contract.md` and preserve the shared
trust boundaries rather than adding project-name branches to common execution.
