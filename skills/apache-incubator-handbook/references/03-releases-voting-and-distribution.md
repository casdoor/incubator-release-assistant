# Releases, voting, signing, and distribution

Last verified: 2026-08-02.

This file explains policy and process. The separate release-assistant Skill
performs only commit selection, RC preparation, signing, dist-dev upload, and
public RC verification.

## What is official

- The official ASF release is the approved source package. Convenience binaries
  may accompany it, but must be built from that approved source and comply with
  release, licensing, branding, and trademark policy.
- A source package must be sufficient for a user with the appropriate platform
  and tools to build and test it.
- Every distributed package needs an ASCII-armored detached OpenPGP signature.
  Current releases also need a modern checksum such as SHA-256 or SHA-512.
- Podling filenames include `incubating`, and the source package includes
  `LICENSE`, `NOTICE`, and `DISCLAIMER` or `DISCLAIMER-WIP` as applicable.

## Candidate versus release

- An RC is staged for review in the project development community. It is not an
  official release and must not be advertised as one.
- Do not rebuild between vote and publication. The approved bytes are the bytes
  promoted to the official release area.
- If any candidate file changes materially, use a new RC identifier and restart
  review. Re-signing the same source creates different signature bytes and is a
  different candidate record.

## Two-phase podling vote

1. **Podling phase:** open `[VOTE]` on the podling dev list. PPMC votes are
   binding for this phase. Success normally means at least three binding `+1`
   votes and more binding `+1` than binding `-1`.
2. Publish a result message and preserve its public archive URL.
3. **IPMC phase:** open `[VOTE]` on `general@incubator.apache.org`, include the
   podling vote result and archive link, and ask the IPMC to approve the same RC.
   IPMC member votes are binding in this phase.
4. Publish the IPMC result. Only after successful IPMC approval may the release
   be published as an Act of the Foundation.

Both phases normally remain open for at least 72 hours. Only explicit votes
count; the release manager has no automatic `+1`. Release votes use majority
approval, not veto rules. A serious `-1` finding is normally addressed by
cancelling the candidate and preparing a new RC, even though it is not a veto.

Before a binding `+1`, a reviewer downloads the signed source onto their own
machine, verifies policy and legal files, checks the signature and checksum,
builds from the source package, and tests on their platform.

## Signing and KEYS

- Use an OpenPGP-compatible ASCII-armored detached `.asc` signature.
- Keep the private key off ASF machines; signing must not happen on ASF hosts.
- Publish every signing public key in the project's `KEYS` file. Retain keys
  used for older releases so archived artifacts remain verifiable.
- Keys for new artifacts must be RSA with at least 2048 bits; new keys should be
  4096-bit RSA.
- A checksum confirms downloaded bytes; the signature binds those bytes to the
  holder of the private key. Both are needed for different reasons.

## Distribution locations

- `dist/dev`: development-community staging for RC review. It is not the public
  official release channel.
- `dist/release`: authoritative SVN-backed area used to publish approved current
  releases to `downloads.apache.org`.
- `downloads.apache.org`: public current release distribution.
- `archive.apache.org`: automatic long-term archive of releases that previously
  appeared in the release area.

Do not direct end users to `dist.apache.org`. Project download pages link source
downloads through the approved download mechanism and link signatures,
checksums, and `KEYS` using `https://downloads.apache.org/...`.

Docker Hub, GitHub Releases, language registries, and similar channels may host
convenience artifacts only under the applicable ASF policy; they do not replace
the official source distribution.

## Minimal release evidence

Keep these records:

- exact source commit and RC name;
- source archive, checksum, and detached signature;
- public dist-dev URL and re-verification result;
- podling vote email, result, tally, and archive URLs;
- IPMC vote email, result, binding tally, and archive URLs;
- exact-byte promotion/publication record;
- final download-page and announcement links.

Official sources:

- https://www.apache.org/legal/release-policy.html
- https://incubator.apache.org/cookbook/#two-phase-vote-on-podling-releases
- https://www.apache.org/foundation/voting.html
- https://infra.apache.org/release-distribution.html
- https://infra.apache.org/release-signing.html
- https://infra.apache.org/release-download-pages.html
- https://www.apache.org/info/verification
