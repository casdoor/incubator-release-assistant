# Apache Incubator source release gates

## Candidate identity and legal content

- Use the exact canonical upstream commit and a clean archive operation.
- The filename includes `incubating` and the archive has one top-level directory.
- The source root includes `LICENSE`, `NOTICE`, and `DISCLAIMER` or
  `DISCLAIMER-WIP`; legal files describe the actual bundled content.
- Run Apache RAT on the final archive and explain exclusions. Never add ASF
  headers mechanically to third-party, generated, binary, or test-data files.

## Build and test

- Test the extracted source archive rather than only a developer checkout.
- Treat repository code as untrusted: run it in the adapter sandbox without
  artifact, home-directory, keyring, SSH, or credential mounts.
- Preserve complete private evidence and stop on every failing command.

## Checksum and signature

- Generate SHA-512 as 128 lowercase hex characters, two spaces, the plain
  filename, and exactly one LF; reject BOM, CR, extra lines, and path prefixes.
- Re-hash the prepared artifact immediately before signing.
- Create an ASCII-armored detached signature only after human authorization.
- Verify RSA policy, configured UID policy, full fingerprint, and the official
  project KEYS file in an isolated public keyring.

## Distribution and votes

- Stage only to ASF dist dev before votes pass; never overwrite an RC directory.
- Re-download public files and verify archive bytes, checksum, and signature.
- Run the podling dev vote first. A successful Incubator general vote is then
  required for an official podling release.
- Keep each vote open at least 72 hours and distinguish binding IPMC votes.
- Changed bytes or a substantive problem cancel the candidate: increment RC and
  restart both votes.
- After approval, promote the exact voted bytes; never rebuild them.
