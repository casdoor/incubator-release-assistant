# Architecture

## Design rule

Repository facts are reviewed data, language behavior is an adapter, and the RC
preparation/signing/staging gates belong to the engine. Configuration cannot
disable the legal-file, RAT, signature, checksum, no-overwrite, or public-copy
checks used by this workflow.

```text
reviewed JSON
    |
    v
strict config + project policy
    |
    +--> prepare domain --> adapter sandbox --> artifact + SHA-512 + evidence
    |
    +--> sign domain ----> exact digest confirmation --> external GPG home + official KEYS
    |
    +--> stage domain ---> RC confirmation --> ASF dist --> public verification
```

## Current adapter

`casbin-go` is the only accepted adapter. Its policy fixes:

- canonical upstream `https://github.com/apache/casbin.git`;
- the Apache Casbin incubator KEYS and dist locations;
- `LICENSE`, `NOTICE`, a disclaimer, `go.mod`, `go.sum`, and `.rat-excludes`;
- Apache RAT 0.18;
- `go test ./...` in a reviewed Go container;

This intentional narrowness makes the first implementation executable without
pretending that unimplemented repositories are supported.

## State and resumption

Each candidate has one ignored `.ira/runs/<project>-<version>-rc<n>/` directory:

- `state.json` records config digest, exact commit, separate digests for the
  archive/checksum/signature files, signer, and completed stages;
- `artifacts/` contains only the source archive, signature, and checksum;
- `work/` contains disposable repositories, extracts, keyrings, and SVN data;
- `evidence/` contains complete local command output.

A resumed step first verifies the config digest and every candidate-file
digest. State schema 1 predates this protection and is rejected rather than
silently resumed. A staged candidate cannot be cleaned; changed bytes in any
of the three files require a new RC number.

## Trust boundaries

Repository tests execute only in Docker/Podman. The container mounts the
disposable extracted source, not the host home, artifact directory, GPG keyring,
or SVN credentials. Signing is a later process which executes no source code.
Staging is a third process with exact confirmation and no credential cache.

The caller workspace and repository are separate security roots. In the normal
layout Claude Code runs from `/abc`, the checkout is
`/abc/Incubator-release-assistant`, and the wrappers export
`IRA_SECRET_DIR=/abc/secretkey` with `GNUPGHOME=/abc/secretkey`. The Go
engine resolves symlinks, rejects a secret directory inside the checkout, and
rejects one captured by any larger Git worktree, then passes the external home
explicitly to secret-key inspection and signing. `/abc` is therefore a plain
workspace rather than another Git checkout.

This separation is more important than the implementation language: it prevents
a compromised test from modifying the signed archive or reading release keys.

## Extension sequence

1. Add an adapter implementation and project policy as described in
   `adapter-contract.md`.
2. Add a reviewed example with no credentials.
3. Add contract tests proving legal files, archive identity, sandbox behavior,
   and the dist-dev endpoint cannot be weakened.
4. Add a CI matrix for the adapter's supported host platforms.
5. Only then expose the adapter in the schema and Skill.

Do not introduce arbitrary shell strings into configuration. If a behavior is
reusable, make it typed adapter code with an argument vector and tests.
