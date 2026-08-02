# Security model

## Protected assets

- release-signing private keys and passphrases;
- ASF credentials and authenticated SVN state;
- the exact source artifact approved for voting;
- private mailing-list content and local evidence.

## Untrusted inputs

- the source repository and every test/build action in it;
- release configuration received through a pull request;
- downloaded tools and container images until verified/reviewed;
- existing local run directories until their recorded hashes are checked.

## Controls

- strict JSON decoding rejects unknown fields;
- Casbin upstream and ASF endpoints are allowlisted;
- names and paths are normalized and traversal is rejected;
- arbitrary commands are absent from configuration;
- RAT downloads are obtained from ASF and verified with the published SHA-512;
- project code runs in a container without artifact or credential mounts;
- signing and staging require exact, stage-specific human confirmations;
- state and release files are re-hashed on resume;
- existing remote RC directories are never overwritten;
- public files are downloaded and verified after staging;
- local configs, state, artifacts, evidence, credentials, and common key formats
  are ignored, while CI scans committed history for secrets.

## Residual risks

- Docker/Podman controls a privileged host service and must be trusted;
- the reviewed container tag can change upstream; command evidence records the
  runtime, and a future release should support immutable image digests;
- GPG pinentry and SVN authentication are external programs;
- a release manager must still review legal/provenance findings and vote text;
- ignored files are not a security boundary, so secret scanning remains
  necessary.
