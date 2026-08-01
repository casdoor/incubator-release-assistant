# Repository guidance

## Purpose

Maintain a repository-neutral Agent Skill and deterministic tooling for Apache
Incubator source releases. Project and language differences belong in reviewed
configuration or adapters, not in the core workflow.

## Boundaries

- Treat `config/release.schema.json` as the public configuration contract.
- Keep reusable Agent instructions under
  `skills/incubator-release-assistant/`.
- Keep project examples free of credentials and private information.
- Treat `legacy/casbin-go-rc/` as a migration baseline, not the generic engine.
- Do not silently weaken checksum, signature, legal-file, RAT, test, vote, or
  public-download verification gates.
- Require explicit human confirmation immediately before signing, ASF dist
  mutation, vote sending, or any other irreversible external action.

## Repository hygiene

- Never commit local configuration, release artifacts, private keys, passwords,
  tokens, cookies, generated evidence, downloaded tools, or working trees.
- Preserve materially different project-specific behavior in an adapter or
  example rather than adding project-name conditionals to shared code.
- Run the Skill validator and configuration validator before publishing
  changes.
