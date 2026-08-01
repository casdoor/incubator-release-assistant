---
name: incubator-release-assistant
description: Prepare, verify, stage, and document configuration-driven Apache Incubator source release candidates. Use when Codex must work on RC packaging, clean source archives, legal files, Apache RAT, build and test gates, LF-only SHA-512, GPG and KEYS verification, ASF dist dev staging, dev or general vote handoff, or repeatable release automation across repositories and languages.
---

# Incubator Release Assistant

Use a reviewed JSON configuration as the source of repository-specific facts.
Keep shared ASF release gates independent of project name and language.

## Start

1. Resolve the workspace and configuration path.
2. Run `scripts/validate-release-config.ps1 -Config <path>`.
3. Read `references/configuration.md` when authoring or changing configuration.
4. Read `references/apache-release-gates.md` before packaging, signing,
   staging, voting, or declaring a candidate ready.
5. Record the exact upstream commit, RC identifier, signer fingerprint, and
   evidence directory before mutating external state.

Stop if the configuration contains placeholders, a non-upstream commit, a
missing required file, or an unresolved legal or provenance decision.

## Execute the release workflow

1. Fetch the configured repository and resolve the exact 40-character commit.
2. Create a clean source archive with the configured prefix. Do not use a
   GitHub-generated source archive.
3. Extract the archive and enforce one top-level directory.
4. Check every configured required file.
5. Run declared project commands in both source and extracted contexts when
   applicable, preserving complete output as evidence.
6. Run Apache RAT on the final archive. Classify findings manually; never add
   ASF headers mechanically to third-party, generated, binary, or test-data
   files.
7. Produce an ASCII, LF-only SHA-512 file with two spaces before the plain
   artifact filename. Verify its bytes and digest independently.
8. Verify the signing key, Apache UID requirement, key strength, and presence
   in the configured official `KEYS`; then create and verify a detached ASCII
   signature in an isolated keyring.
9. Confirm the candidate directory contains only the source archive,
   signature, and checksum.
10. Before ASF dist mutation, show the exact target and request explicit human
    confirmation. Never overwrite an existing RC directory.
11. Re-download public dist files and compare checksum, signature, and archive
    bytes before preparing vote text.
12. Archive evidence and exact communication text. Changed artifacts require a
    new RC number and fresh votes.

## Handle language and project differences

Use `checks.commands`, `checks.requiredFiles`, and RAT configuration for the
current repository. Add a reusable adapter when several repositories share the
same language behavior. Do not add project-name conditionals to the shared
workflow.

The repository root contains `legacy/casbin-go-rc/`, the proven Casbin Go
PowerShell baseline. Consult it only while migrating a gate or operating that
specific legacy flow; do not present it as generic automation.

## Preserve safety and evidence

- Never store passwords, tokens, private keys, cookies, or private-list text.
- Keep local configuration, artifacts, downloaded tools, worktrees, and
  evidence in ignored paths.
- Treat signature, checksum, license, provenance, RAT, test, and public-download
  failures as hard stops.
- Treat signing, ASF dist writes, vote messages, and announcements as explicit
  human-confirmation boundaries.
- Report verified facts separately from assumptions and pending human choices.
