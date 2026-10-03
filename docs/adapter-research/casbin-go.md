# Casbin Go adapter research

Preparation date: 2026-08-06
Status: Active design evidence for the `casbin-go` adapter

## What the upstream repository actually does

The authoritative upstream workflow is
[`apache/casbin/.github/workflows/release.yml`](https://github.com/apache/casbin/blob/master/.github/workflows/release.yml).
It is triggered only by a manually pushed tag; merging a pull request does not
publish a release.

- `vX.Y.Z-rcN` creates a GitHub **pre-release**.
- `vX.Y.Z` creates the GitHub release and requests publication of the matching
  Go module version through `proxy.golang.org`.
- The workflow creates source `.tar.gz` and `.zip` archives from the exact tag,
  writes SHA-512 files, verifies those checksums, and attaches the output to the
  GitHub release.
- The normal CI workflow runs `make test` on pushes and pull requests. IRA must
  check the selected commit's already-run CI; it must not execute target code
  merely to duplicate it.

The Apache source-release portion is governed separately: a source artifact is
reviewed, signed, staged for voting, and only then becomes an official release.
See [Apache Release Policy](https://apache.org/dev/release),
[Apache release publishing guidance](https://infra.apache.org/release-publishing.html),
and [ASF signing guidance](https://infra.apache.org/release-signing.html).

## Adapter lifecycle

The Go Casbin adapter must represent both release surfaces without claiming that
a tag alone is an Apache approval:

```text
selected commit
  -> verify upstream CI evidence
  -> prepare signed Apache source RC
  -> vote and approve (human/community work)
  -> push final vX.Y.Z tag (explicit human authority)
  -> upstream GitHub workflow publishes GitHub Release and Go module proxy
  -> verify tag, GitHub assets, checksums, and proxy availability
```

RC tags and final tags are deliberately different. The final tag must not be
created by a queue traversal or by a JSON command string. A release manager
must explicitly authorize it after the required Apache process has completed.

## What IRA owns in this phase

1. An adapter registry: only a registered adapter can run from the queue.
2. Project-specific validation inside the adapter, rather than in the generic
   engine entrypoint.
3. Ordered queue state: the current repository, its next action, and the next
   incomplete repository.
4. Existing immutable source-candidate, signing, staging, and public-byte
   verification controls for the Casbin Apache source release.

## Compatibility with the existing RC workflow

The adapter extraction does **not** change the existing Casbin source-RC
configuration schema, `RunID`, artifact name, state-file digest, or the
`validate -> prepare -> sign -> stage -> verify-public` command path. Existing
prepared candidates therefore continue to use their matching `.ira/runs/`
state. The queue and `adapters` commands are additive.

The current upstream GitHub workflow uses the archive basename
`apache-casbin-incubating-<version>-src`, while the existing ASF source-RC
workflow preserves its already-tried configuration-derived archive name. This
is a pre-existing difference between the two release surfaces, not a change
introduced by adapter extraction. Do not rename an existing RC to match the
GitHub workflow: changed bytes require a new RC. The future final-publication
verifier must model the GitHub artifact naming separately.

## Deliberately deferred work

- A new generic configuration schema that can express npm, Maven, PyPI, and
  GitHub-only release targets without Casbin fields.
- Read-only verification of the final tag's GitHub Release assets and Go module
  proxy availability.
- An explicit, human-confirmed final-tag action.
- Additional adapters. Each must begin with the target repository's actual
  release workflow and package-registry rules, not a guessed generic command.

## Design implication

The generic engine must never interpret a tracker row as executable shell.
It accepts structured adapter data, then the registered adapter supplies the
reviewed checks, artifact names, and publication-verification rules. This keeps
the queue useful for every Casbin repository while making unsupported releases
visible rather than pretending they were published.
