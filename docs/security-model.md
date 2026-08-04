# Security model

## Protected assets

- release-signing private keys and passphrases;
- ASF credentials and authenticated SVN state;
- the exact source artifact approved for voting;
- private mailing-list content and local evidence.

## Untrusted inputs

- the source repository and its archive contents;
- release configuration received through a pull request;
- downloaded tools until verified/reviewed;
- existing local run directories until their recorded hashes are checked.

## Controls

- strict JSON decoding rejects unknown fields;
- Casbin upstream and ASF endpoints are allowlisted;
- names and paths are normalized and traversal is rejected;
- arbitrary commands are absent from configuration;
- RAT downloads are obtained from ASF and verified with the published SHA-512;
- target-project code is not executed during `prepare`; project test status is
  checked separately in the selected commit's existing GitHub CI;
- platform wrappers keep the signing keyring in an external sibling
  `secretkey` directory, and the engine rejects repository-contained or
  relative secret paths;
- signing and staging require exact, stage-specific human confirmations;
- state freezes separate archive, checksum-file, and signature-file digests;
- every frozen file is re-hashed on resume, before staging, and after public
  download;
- existing remote RC directories are never overwritten;
- public files are downloaded and verified after staging;
- local configs, state, artifacts, evidence, credentials, and common key formats
  are ignored; staged changes must be inspected for secrets before committing.

The expected checkout `/abc/Incubator-release-assistant` and key root
`/abc/secretkey` are siblings under a non-Git `/abc` workspace. Git never
traverses the sibling directory, and
staging uploads only the source archive, LF-only checksum, and detached public
signature. Private key files are not copied into release artifacts.

## Residual risks

- GitHub CI status is external evidence and must be checked for the exact
  selected upstream commit;
- GPG pinentry and SVN authentication are external programs;
- a release manager must still review legal/provenance findings;
- ignored files are not a security boundary, so secret scanning remains
  necessary.
