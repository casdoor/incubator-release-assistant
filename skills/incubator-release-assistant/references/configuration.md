# Configuration reference

Use `config/release.schema.json` as the canonical contract and begin from a
reviewed file in `config/examples/`. Store active configurations under the
ignored `config/local/` directory.

## Sections

- `project`: stable project identity, implementation language, and incubation
  status.
- `source`: canonical repository, exact 40-character commit, and archive root
  name.
- `release`: version, RC number, and artifact base name.
- `checks`: required top-level files, deterministic commands, and RAT settings.
- `signing`: Apache ID, public fingerprint, official `KEYS` URL, UID policy, and
  minimum RSA size.
- `distribution`: ASF dist dev and release URLs.
- `votes`: podling dev list, Incubator general list, and minimum vote duration.

## Rules

- Use only a commit that exists in the canonical upstream repository.
- Keep executable commands reviewable and deterministic.
- Use URLs under the applicable ASF distribution location.
- Put no secret in JSON. A fingerprint and public Apache ID are not secrets;
  passwords, private keys, tokens, and cookies are secrets.
- Treat all-zero commit and fingerprint values as placeholders that must be
  replaced before a release run.
- Increment `release.rc` whenever any candidate artifact byte changes.
