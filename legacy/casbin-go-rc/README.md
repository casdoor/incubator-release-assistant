# Legacy Casbin Go RC baseline

This directory preserves the PowerShell release flow developed for Apache
Casbin `3.11.0-incubating`. It is retained because it already encodes important
checks discovered during RC1 and RC2 work:

- clean `git archive` packaging;
- Apache RAT and Go tests on extracted source;
- LF-only SHA-512 validation;
- GPG private-key, Apache UID, RSA-size, and official `KEYS` checks;
- ASF dist dev collision prevention and explicit staging confirmation;
- public re-download and byte-for-byte verification.

It is intentionally labelled `legacy` because it still hard-codes Casbin,
Go-specific files and commands, ASF URLs, and Windows PowerShell behavior.
Do not add more project-name branches to it. Migrate reusable gates into the
generic engine and language-specific behavior into adapters.

- `release-rc.ps1`: executable baseline.
- `release.example.legacy.json`: original script configuration format.
- `LEGACY_USAGE.md`: original usage instructions.
- `EXPECTED_OUTPUT.md`: original 15-step result guide.
- `keyupdate.md`: one-time release-signing key guide.
