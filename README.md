# Incubator Release Assistant (IRA)

This repository contains two independent Agent Skills:

- **Incubator Release Assistant (IRA):** a configuration-driven, cross-platform
  Apache Casbin Go RC preparation, signing, dist-dev upload, and verification
  tool.
- **Apache Incubator Handbook:** an offline reference for podling governance,
  reporting, roles, voting, release policy, IP/licensing, branding, websites,
  community, infrastructure, security, graduation, and retirement.

Its normal job is deliberately small: select a commit, prepare an RC, sign it,
upload the archive/checksum/signature to ASF dist dev, and verify the public
copy. Voting, final release promotion, and broader compliance programs are
separate follow-up work rather than prerequisites for using the RC workflow.

The current implementation deliberately supports **Apache Casbin Go only**.
The engine has an adapter boundary so later languages and repositories can be
added without weakening the shared release gates.

> IRA is a community tool. ASF policy and the required human votes remain the
> authority. Signing, ASF writes, votes, and announcements are never fully
> automated without human approval.

## What is implemented

- strict, unknown-field-rejecting release-ready configuration validation;
- canonical Casbin upstream, ASF URL, legal-file, disclaimer, and naming
  invariants that configuration cannot turn off;
- exact-commit `git archive` packaging with one top-level directory;
- Apache RAT verification and canonical LF-only SHA-512 output;
- `go test ./...` inside Docker or Podman with only a disposable extracted
  source tree mounted—the host artifact directory and credentials are absent;
- separate `prepare`, `sign`, and `stage` trust boundaries;
- an external signing home under the parent workspace, never inside Git;
- signing-key and official `KEYS` verification in an isolated public keyring;
- knowledge-guided setup for missing config, signing keys, and official `KEYS`;
- no-overwrite ASF dist dev staging and public byte-for-byte re-verification;
- resumable state under ignored `.ira/runs/`;
- a self-contained Skill containing the Go engine, schema, and Casbin template.

## Prerequisites

- Go 1.22 or newer;
- Git, tar, Java, GnuPG, SVN;
- Docker or Podman for running repository code without access to host secrets.

The release Skill includes `scripts/run.ps1` for Windows PowerShell 5.1+ and
`scripts/run.sh` for Bash on Linux and macOS. Both invoke the same Go engine.

The container engine is a security boundary. IRA intentionally has no
"run project tests directly on the signing host" fallback.

## Activate the release Skill

The operational instructions live in
[`skills/incubator-release-assistant/SKILL.md`](skills/incubator-release-assistant/SKILL.md).
Install that Skill directory using the Agent client's Skill mechanism, or
explicitly ask an Agent working with this clone to read and follow that file.
Cloning a repository does not guarantee that every Agent client automatically
discovers a nested Skill.

The Skill uses progressive disclosure. It reads detailed setup knowledge only
when config, signing-key, official-KEYS, or recovery help is needed.

## Expected Claude Code workspace

Run Claude Code from the parent workspace, not from inside the repository. The
parent `/abc` directory itself must not be a Git checkout. The normal
Linux/macOS layout is:

```text
/abc/
├── Incubator-release-assistant/   cloned Git repository
└── secretkey/                     release manager's GPG home; never a Git repository
```

Clone with the intended directory name and prepare the external key directory:

```bash
cd /abc
git clone https://github.com/casdoor/incubator-release-assistant.git Incubator-release-assistant
mkdir -p secretkey
chmod 700 secretkey
```

Import or generate the signing key with `GNUPGHOME=/abc/secretkey`.
Do not copy a private-key export, keyring, passphrase, token, or credential file
into `Incubator-release-assistant/`. The wrappers reject a secret directory
inside the repository. `.gitignore` also excludes a mistakenly created
`secretkey/` as defence in depth, but ignore rules are not the security boundary.
The wrappers and engine also reject `secretkey` when any parent directory owns
Git metadata, preventing accidental capture by a larger wrapper repository.

The ASF `KEYS` file is public verification material. IRA downloads a disposable
copy under ignored run state; it is not the release manager's private keyring.

If setup is incomplete, the release Skill does not stop at the raw engine
error. It loads the matching bundled reference and explains:

- the resolved config, external GPG home, public-key export, and evidence paths;
- which files are user-provided, Agent-generated, GPG-managed, or official ASF
  state;
- the exact safe next action;
- the separate confirmation required before key generation or official KEYS
  publication.

The normal layout may also include public, non-secret key material:

```text
/abc/
├── Incubator-release-assistant/
│   ├── config/local/casbin.local.json   ignored non-secret config
│   └── .ira/runs/                       generated state and evidence
├── secretkey/                           GPG-managed private home
└── public-key/
    ├── apache-casbin-release-key.asc    generated public export
    └── key-metadata.json                generated public metadata
```

Do not manually create individual files inside `secretkey/`; GPG owns that
directory. See the Skill references for Windows, Linux, and macOS examples.

## Human workflow

From `/abc`, copy and fill the non-secret template:

```bash
mkdir -p ./Incubator-release-assistant/config/local
cp ./Incubator-release-assistant/config/examples/casbin-go.json \
  ./Incubator-release-assistant/config/local/casbin.local.json
```

The corresponding PowerShell setup from a parent workspace is:

```powershell
New-Item -ItemType Directory -Force .\Incubator-release-assistant\config\local | Out-Null
Copy-Item .\Incubator-release-assistant\config\examples\casbin-go.json `
  .\Incubator-release-assistant\config\local\casbin.local.json
notepad .\Incubator-release-assistant\config\local\casbin.local.json
```

Only the commit, version/RC-derived names, Apache ID, signer fingerprint, and
reviewed runtime choice normally need attention. Passwords, private keys,
tokens, and cookies never belong in JSON.

Then run:

```bash
./Incubator-release-assistant/ira.sh validate \
  --config ./Incubator-release-assistant/config/local/casbin.local.json
./Incubator-release-assistant/ira.sh plan \
  --config ./Incubator-release-assistant/config/local/casbin.local.json
./Incubator-release-assistant/ira.sh prepare \
  --config ./Incubator-release-assistant/config/local/casbin.local.json
```

During the later `sign` command, the Bash wrapper defaults to
`/abc/secretkey` because the caller is `/abc`. Override it only with another
absolute external directory by passing `--secret-dir`. Validation, planning,
preparation, staging, and public verification do not access the private key
directory.

On Windows, use the same parent-workspace layout:

```powershell
.\Incubator-release-assistant\ira.ps1 validate `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json
.\Incubator-release-assistant\ira.ps1 plan `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json
.\Incubator-release-assistant\ira.ps1 prepare `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json
```

`prepare` prints the exact SHA-512 required for the next explicit boundary:

```powershell
.\Incubator-release-assistant\ira.ps1 sign `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json `
  -Confirm <exact-128-character-sha512>

.\Incubator-release-assistant\ira.ps1 stage `
  -Config .\Incubator-release-assistant\config\local\casbin.local.json `
  -Confirm "STAGE RC2"
```

On Linux/macOS, remain in `/abc` and use:

```bash
./Incubator-release-assistant/ira.sh sign \
  --config ./Incubator-release-assistant/config/local/casbin.local.json \
  --confirm <exact-128-character-sha512>
./Incubator-release-assistant/ira.sh stage \
  --config ./Incubator-release-assistant/config/local/casbin.local.json \
  --confirm "STAGE RC2"
```

Staging automatically re-downloads the public candidate and verifies its
archive, checksum file, and signature file byte-for-byte, then also verifies
the checksum meaning and GPG signature. Re-running `prepare`, `sign`, or
`stage` resumes and revalidates the frozen files instead of rebuilding them.

After the run succeeds, the project author must still personally inspect the
RAT report, every `.rat-excludes` entry, and the actual `LICENSE`, `NOTICE`, and
`DISCLAIMER` content. RAT reporting zero unapproved files does not by itself
approve the legal reason for an exclusion or bundled third-party material.

## What signing and "exact bytes" mean

Signing does not change the source archive. GPG reads the exact
`*.tar.gz` bytes and uses the release manager's private key to create a
separate ASCII-armored `*.asc` file. A reviewer imports the public key from
the official `KEYS` file and verifies the signature. Verification proves both
that the configured key signed the archive and that the archive has not
changed since signing.

An RC consists of three immutable files:

1. the source archive (`*.tar.gz`);
2. its canonical checksum line (`*.sha512`);
3. its detached signature (`*.asc`).

"Exact bytes" means every byte in all three files must remain identical to
the files proposed for the vote. CRLF instead of LF, recompressing the same
source, or signing the same archive again produces different bytes. Any such
change creates a different candidate and requires a new RC and fresh votes.

After signing, IRA writes `evidence/candidate-manifest.txt` with a SHA-512 for
each of the three files. Resume, staging, remote-directory recovery, and public
verification must all match that manifest. The checksum file itself is
accepted only as exactly 128 lowercase hexadecimal characters, two spaces, the
plain archive filename, and one final LF byte.

## Safety model

The release is split into three trust domains:

1. `prepare` handles untrusted repository code. Go tests run in a container
   which receives only the extracted source tree. It cannot see release
   artifacts, the user's home directory, GPG keyring, or ASF credentials.
2. `sign` executes no repository code. It re-hashes the prepared artifact and
   requires the human to repeat that exact digest before accessing GPG. It
   explicitly uses the external `<workspace>/secretkey` home, then freezes
   the archive, checksum, and signature file digests.
3. `stage` revalidates all three frozen files and the GPG signature, requires
   `STAGE RC<n>`, refuses existing remote RC directories, avoids credential
   caching, and verifies every public file byte-for-byte.

The configured Go image is restricted to the reviewed `golang:1.24` image. IRA
records complete command output as private local evidence. Docker/Podman itself
is privileged software and must come from a trusted installation.

## Repository layout

```text
config/                         Human-facing schema and template
docs/                           Architecture, security, adapter contract, and repository map
legacy/casbin-go-rc/            Preserved pre-engine PowerShell baseline
skills/apache-incubator-handbook/
  SKILL.md                      Offline incubation-knowledge router
  references/                   Roles, governance, release, IP, branding, lifecycle, and security
skills/incubator-release-assistant/
  SKILL.md                      Agent workflow
  assets/                       Self-contained schema and template mirrors
  references/                   Config, workspace, key, KEYS, and recovery guidance
  scripts/run.ps1               Windows Skill entry point
  scripts/run.sh                Linux/macOS Skill entry point
  scripts/ira/                  Go CLI, engine, and tests
scripts/sync-skill-assets.ps1   Keeps Skill assets identical to root config
ira.ps1 / ira.sh                Human-friendly entry points
```

`secretkey/` is intentionally absent from this tree because it is a sibling of
the repository, not repository content.

For a file-by-file explanation, see
[`docs/repository-map.md`](docs/repository-map.md).

## Offline incubation knowledge

Broader Apache Incubator knowledge lives in the separate
[`apache-incubator-handbook`](skills/apache-incubator-handbook/SKILL.md) Skill.
It uses progressive disclosure: an Agent reads only the reference matching the
question. The release Skill neither loads nor depends on the handbook.

The bundled summaries were verified against official ASF sources on 2026-08-02.
Ordinary questions should be answered offline; live rosters, current report
templates/deadlines, vote state, distribution contents, and later policy changes
still require current official evidence.

## Current boundary and extension path

The only registered adapter is `casbin-go`; other project IDs, repositories,
ASF destinations, and adapters are rejected. To add another ecosystem, first
implement and test a new adapter following [docs/adapter-contract.md](docs/adapter-contract.md).
Do not add arbitrary shell commands or project-name conditionals to config.

The old PowerShell flow remains under `legacy/` as migration evidence, not as
the default engine.

## License

Apache License 2.0. See [LICENSE](LICENSE).
