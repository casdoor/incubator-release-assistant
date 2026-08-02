# IP clearance, provenance, and licensing

Last verified: 2026-08-02.

This is an operational summary, not legal advice. Fact-specific uncertainty
should be raised through Mentors and the appropriate Incubator or ASF Legal
channel.

## Initial code and provenance

- A podling must establish that the ASF has permission to receive and distribute
  the code it imports.
- Individual contributors normally file ICLAs. Employers may provide CCLAs when
  corporate rights are relevant. A Software Grant Agreement (SGA) is commonly
  required when a person or organization contributes an existing codebase or
  when material rights holders are not joining as initial contributors.
- The proposal acceptance can cover identified initial codebases, but later
  donations follow the applicable IP-clearance process.
- If ownership cannot be established, affected code may need to be removed or
  independently rewritten. A podling cannot release code whose provenance and
  required paperwork have not been established.

Do not infer rights merely from repository ownership, an open-source label, or
a contributor's technical access. Track donor, rights holder, contribution,
paperwork, import commit, and approval evidence separately.

## Third-party licensing categories

- **Category A:** generally compatible “Apache-like” licenses that may be
  included when their terms and notices are followed.
- **Category B:** conditionally acceptable. Many weak-copyleft works may appear
  only in appropriately labelled binary form and must not be bundled in an ASF
  source release.
- **Category X:** must not be distributed in ASF source or convenience binaries.
  A build may sometimes rely on an external optional tool without distributing
  it, but the prohibited component cannot be bundled.

Always verify the exact license and current ASF classification. Similar names,
dual licensing, version differences, generated files, and transitive bundling
can change the answer.

## LICENSE and NOTICE

- Every ASF distribution contains `LICENSE` and `NOTICE` appropriate to the
  actual contents of that artifact.
- `LICENSE` records the licenses applying to the distributed material,
  including bundled third-party components and pointers/copies required by
  their licenses.
- `NOTICE` is not a general credits file. Include legally required notices and
  preserved relevant notices; do not copy every dependency's NOTICE blindly.
- Source and convenience-binary artifacts can contain different material, so
  their licensing files may differ.

## Source headers

- Human-authored ASF source files normally use the standard ASF header.
- Some short, non-creative, generated, test-data, snippet, binary, or media
  files cannot sensibly carry the full header. Record why instead of modifying
  files mechanically.
- A third-party file keeps its applicable original licensing information; do
  not overwrite it with an ASF header merely to silence a scanner.
- Images and media need provenance and any required attribution. A file being
  binary does not make it free of copyright or notice obligations.

## RAT interpretation

Apache RAT helps identify files without recognized license information. It does
not decide ownership, license compatibility, whether an exclusion is justified,
or whether LICENSE/NOTICE accurately describe bundled content.

For each exclusion, record:

1. file or pattern;
2. why a header is impossible or inappropriate;
3. origin and owner;
4. applicable license or grant;
5. where required attribution is preserved;
6. reviewer and date.

A useful release review has zero unexplained RAT findings, not merely a report
that reaches zero after broad exclusions.

## Practical inventory

Review at least:

- source files and generated code;
- vendored dependencies and copied examples;
- test data, fixtures, fonts, images, logos, audio, and datasets;
- archives or binaries committed in the source tree;
- build output accidentally included by packaging;
- package manager metadata and transitive content actually bundled;
- cryptography/export-control declarations when applicable.

Official sources:

- https://incubator.apache.org/guides/ip_clearance.html
- https://www.apache.org/licenses/contributor-agreements.html
- https://www.apache.org/legal/resolved.html
- https://www.apache.org/legal/src-headers.html
- https://www.apache.org/legal/release-policy.html
- https://www.apache.org/legal/
