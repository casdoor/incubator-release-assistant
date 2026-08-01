# Apache source release gates

## Candidate identity

- Record the version, RC number, canonical upstream repository, exact commit,
  artifact name, and signer fingerprint.
- Build from the exact upstream commit with a clean archive operation.
- Use a single top-level source directory and exclude VCS metadata and secrets.

## Legal and provenance

- Require the configured `LICENSE`, `NOTICE`, and incubation disclaimer files.
- Ensure legal files describe content actually bundled in the source archive.
- Run Apache RAT on the final archive and explain every exclusion.
- Preserve durable provenance or grant evidence for binary and non-source
  assets.

## Build and test

- Run configured build and test commands with complete evidence.
- Test the extracted source archive, not only a developer checkout.
- Stop if tests mutate tracked fixtures or make the release non-reproducible.

## Checksum and signature

- Generate SHA-512 as lowercase hexadecimal, two spaces, plain filename, and
  exactly one LF; reject BOM and CR bytes.
- Create a detached ASCII-armored signature with the configured private key.
- Verify the fingerprint, UID and key-strength policy, and official `KEYS` in an
  isolated keyring.

## Distribution and votes

- Stage only to ASF dist dev before votes pass.
- Never overwrite an existing RC directory.
- Verify files again after downloading them from the public dist URL.
- Run the podling dev vote first, then the Incubator general vote when required.
- Keep each vote open for the configured minimum duration and record binding
  status accurately.
- If candidate bytes change or a substantive policy issue is found, cancel the
  RC, increment the RC number, and restart the full workflow and votes.
- Promote the exact voted bytes to dist release; never rebuild after approval.
