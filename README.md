# Incubator Release Assistant

An unofficial, configuration-driven release assistant and Agent Skill for
Apache Incubator source releases.

The project separates reusable release policy from repository-specific facts.
Humans provide a reviewed JSON configuration; the Skill and scripts use that
configuration to prepare, verify, sign, stage, and document a release
candidate without hard-coding a project name or programming language.

> This is a community tool. It is not an official Apache Software Foundation
> service and does not replace ASF release policy or a release vote.

## Repository layout

```text
config/
  release.schema.json       Shared configuration contract
  examples/                 Reviewed, non-secret project examples
docs/
  architecture.md           Boundaries and migration design
legacy/
  casbin-go-rc/              Proven Casbin Go PowerShell baseline
skills/
  incubator-release-assistant/
    SKILL.md                 Agent workflow
    agents/openai.yaml       Skill UI metadata
    scripts/                 Deterministic validation helpers
    references/              Configuration and ASF gate references
```

## Current scope

The first version provides:

- a reusable Skill contract;
- a repository-neutral JSON configuration schema;
- a Casbin Go example configuration;
- a deterministic PowerShell configuration validator;
- the original Casbin Go RC script as an explicitly labelled migration
  baseline.

The legacy script is not yet the generic engine. It remains available so each
proven gate can be migrated without losing the RC1/RC2 lessons encoded in it.

## Quick start

1. Copy `config/examples/casbin-go.json` to
   `config/local/<project>.local.json`.
2. Replace the example repository, commit, version, signing identity, and ASF
   distribution values.
3. Validate it:

```powershell
.\skills\incubator-release-assistant\scripts\validate-release-config.ps1 `
  -Config .\config\local\project.local.json
```

4. Invoke the Skill and provide the validated configuration path.

Never put passwords, private keys, tokens, authentication cookies, or private
mailing-list content in a configuration file.

## License

Apache License 2.0. See [LICENSE](LICENSE).
