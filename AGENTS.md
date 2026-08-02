# Repository guidance

## Purpose and current scope

Maintain a safe, configuration-driven Apache Incubator release engine and
self-contained Agent Skill. The current executable scope is Apache Casbin Go
only. Do not claim support for another project/language until its adapter and
contract tests exist.

## Commands

```powershell
# Human entry point
.\ira.ps1 validate -Config .\config\local\casbin.local.json
.\ira.ps1 plan -Config .\config\local\casbin.local.json
.\ira.ps1 prepare -Config .\config\local\casbin.local.json

# Engine tests
$engine = ".\skills\incubator-release-assistant\scripts\ira"
go -C $engine test ./...
go -C $engine vet ./...

# Repository and Skill contract
.\scripts\validate-repository.ps1
```

After changing root schema/example files, run
`scripts/sync-skill-assets.ps1` and commit both mirrors.

## Non-negotiable boundaries

- Project code executes only in the adapter container. Never add a host fallback.
- The sandbox must not mount artifact, home, GPG, SSH, or credential paths.
- Signing and staging remain separate commands with exact human confirmations.
- Configuration contains no arbitrary shell command.
- ASF legal, disclaimer, RAT, checksum, signature, KEYS, vote, no-overwrite,
  and public-download gates are engine policy, not optional project data.
- A staged RC cannot be cleaned or overwritten; changed bytes require a new RC.
- Treat `legacy/casbin-go-rc/` only as historical migration evidence.

## Repository hygiene

Never commit active local configuration, `.ira/`, artifacts, evidence, private
keys, passwords, tokens, cookies, credential stores, or private-list content.
Public Apache IDs and signing fingerprints may appear only where operationally
necessary. Keep the gitleaks workflow enabled.

When adding an adapter, follow `docs/adapter-contract.md` and preserve the shared
trust boundaries rather than adding project-name branches to common execution.
