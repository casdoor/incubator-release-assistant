---
name: incubator-release-assistant
description: Turn a selected Apache Casbin Go commit into an Incubator release candidate, including guiding a new release manager through missing config, signing-key, and official KEYS setup; then prepare, resume, sign, upload, and verify the RC with the bundled IRA engine. Use this skill whenever a user wants to set up, prepare, resume, sign, upload, or verify a Casbin RC, even when they have not configured a key yet. The engine supports only Apache Casbin Go; the skill also guides manual publishing for Casbin adapter repositories via pluggable language templates (currently Rust / crates.io and Java / Maven Central, with more to be added).
---

# Incubator Release Assistant

Use the bundled engine instead of reconstructing release commands manually.
Keep the task focused on the normal release-manager path:

1. select a commit;
2. prepare the RC files;
3. sign the archive;
4. upload the three RC files;
5. verify the public copy.

Do not expand an ordinary RC run into a general compliance, key-management,
cross-platform, or test-coverage project unless the user explicitly asks for
that work. Report engine failures as they occur instead of inventing additional
preconditions.

## Establish the run

1. Treat the Agent's current directory as the parent workspace. In the expected
   deployment Claude Code works from `/abc`, the repository is
   `/abc/Incubator-release-assistant`, and all signing-key material is under
   `/abc/secretkey`.
2. Keep the secret directory outside every Git checkout. The platform wrapper
   defaults to `<current-directory>/secretkey`, creates only that empty directory
   when needed, and rejects a path contained by the release repository
   or any larger Git worktree. `/abc` must therefore be a plain workspace. Only
   `sign` may resolve or access this directory; earlier and later stages do not.
3. Run the read-only `doctor` command first. It returns JSON with a stable code,
   resolved paths, one matching reference, and one next action. It does not
   create a key, sign, or write to ASF.
4. Read only the reference named by `doctor`. If the user already asked to set
   up or release Casbin, create safe local directories and copy the non-secret
   config template without adding another confirmation turn. Ask only for
   public values that cannot be verified: Apache ID and the selected commit.
5. Obtain the fingerprint by inspecting or generating the signing key; never
   ask the user to invent or manually type an unverified fingerprint.
6. Read `references/configuration.md` when changing fields beyond the normal
   Apache ID, commit, key fingerprint, version, or RC inputs.
7. Use the bundled wrapper for the current platform. Both wrappers resolve a
   relative config path from the caller's working directory and run the same Go
   engine.

```powershell
& <skill-directory>\scripts\run.ps1 doctor -Config <config>
& <skill-directory>\scripts\run.ps1 validate -Config <config>
& <skill-directory>\scripts\run.ps1 plan -Config <config>
```

```bash
bash <skill-directory>/scripts/run.sh doctor --config <config>
bash <skill-directory>/scripts/run.sh validate --config <config>
bash <skill-directory>/scripts/run.sh plan --config <config>
```

Stop when the configuration does not validate. Tell the user exactly which
field must be supplied or corrected, then continue from the same step.

## Route incomplete setup through bundled knowledge

Do not leave a new release manager with a terse engine error. Use the code and
reference returned by `doctor`, then explain the next action in the user's
language:

- `IRA-WORKSPACE-001`, missing config, Apache ID, or source commit:
  read `references/workspace-bootstrap.md`;
- missing fingerprint, private key unavailable, wrong algorithm/size, missing
  Apache UID, expired, revoked, or unable to sign: read
  `references/signing-key-setup.md`;
- `IRA-KEYS-001`, configured fingerprint absent from official Casbin `KEYS`:
  read
  `references/asf-keys-publication.md`;
- `IRA-DEPENDENCY-001` or `IRA-PREFLIGHT-001`: read
  `references/prerequisites.md`;
- incomplete or resumable run state, changed bytes, or uncertainty about which
  step to repeat: read `references/release-recovery.md`.

Keep the first response short. Use this shape and stay under about 12 lines
unless the user asks for the detailed commands or an error needs them:

```text
Current gate: <one sentence>
Paths: <only the two to four paths relevant now>
Next: <one action the Agent can perform>
Need from you: <only unverifiable public input, a choice, or an approval>
```

Explain who creates a file and whether it is secret the first time that file is
introduced. Do not repeat the full directory table on later turns. Expand into
the reference's commands and examples only when the user needs that step.
Require explicit approval only for private-key generation/import, signing, ASF
writes, vote sending, and announcements; ordinary read-only checks and the
non-secret local config copy are part of the requested setup.

`secretkey/` is a GPG-managed directory, not a private-key file the user should
copy into the repository. The public `*.asc` export may be reviewed and added to
ASF `KEYS`; it is different from the private GPG home. Never guess an Apache ID
or fingerprint, and never route an active user to `legacy/`.

After each missing item is supplied, rerun `doctor`. Run `validate` and `plan`
only after `doctor` reports `IRA-READY`. Do not restart or rebuild a frozen
candidate merely because setup was incomplete.

## Prepare in the untrusted-code domain

Run:

```powershell
& <skill-directory>\scripts\run.ps1 prepare -Config <config>
```

On Linux/macOS use `bash <skill-directory>/scripts/run.sh prepare --config
<config>`.

The engine builds the selected commit, runs RAT, and executes `go test ./...`
in Docker/Podman. A successful run prints the artifact SHA-512 and records
resumable state under `.ira/runs/`.

If preparation fails, report the failed gate and evidence path. Use `--clean`
only for an unstaged matching run after confirming its local work is disposable.
Changed candidate bytes require a new RC number.

## Sign in the trusted domain

Signing is a separate authority boundary. Show the prepared artifact path and
SHA-512 to the user. Proceed only when the user explicitly authorizes signing,
then pass the exact digest printed by `prepare`:

```powershell
& <skill-directory>\scripts\run.ps1 sign `
  -Config <config> -Confirm <exact-sha512>
```

On Linux/macOS run:

```bash
bash <skill-directory>/scripts/run.sh sign \
  --config <config> --confirm <exact-sha512>
```

The engine re-hashes the artifact, uses the configured signing key, creates the
detached signature, and verifies it. It records the archive, checksum, and
signature digests in `evidence/candidate-manifest.txt`. Never request or store
a passphrase. The signing key must come from
`<parent-workspace>/secretkey`; do not fall back to `~/.gnupg` and do not
copy private-key material into the repository.

## Stage and publicly verify

ASF dist mutation requires a second explicit user authorization. Display the
target URL and exact confirmation text, then run only after approval:

```powershell
& <skill-directory>\scripts\run.ps1 stage `
  -Config <config> -Confirm "STAGE RC<number>"
```

On Linux/macOS run:

```bash
bash <skill-directory>/scripts/run.sh stage \
  --config <config> --confirm "STAGE RC<number>"
```

The engine uploads only the archive, checksum, and signature, then downloads
the public files and verifies them again. The RC run is complete when
`publicVerified` is true.

## Finish with the author's RAT reminder

After public verification succeeds, give the user a short completion summary:

- selected commit and RC name;
- public dist-dev URL;
- archive, checksum, and signature filenames;
- confirmation that the public copy was verified.

End with this explicit reminder: the project author must personally review the
RAT report, every `.rat-excludes` entry, and the release's `LICENSE`, `NOTICE`,
and `DISCLAIMER` content. An automated RAT result of zero unapproved files does
not decide whether an exclusion or bundled third-party material is legally
appropriate.

Voting and post-vote release work are separate follow-up tasks. Help with them
only when the user asks; they are not part of this Skill's normal RC run.

## Publish adapter packages manually

Package registries are a separate authority boundary from the RC workflow. The
engine steps above never touch crates.io, Maven Central, or any other package
registry. When the user asks to publish a Casbin adapter repository whose
semantic-release was removed, follow the reference instead of the engine
commands:

1. read `references/manual-package-release.md`.  This reference is activated
   only on an explicit user request; do not monitor tags or scan repositories
   on your own;
2. ask the user for the **repository name** (e.g. `casbin-sqlx-adapter`)
   and the **release tag**.  Do not guess or create tags;
3. run the **Common pre-flight** checks from the reference: clone from the
   Apache canonical upstream (warn if the URL points to a personal fork,
   do not hard-stop), checkout the tag, verify the tree is clean, detect
   the project language from the manifest file, verify the tag-version
   match, and confirm upstream CI passed.  Stop if any check fails;
4. jump to the **language-specific template** (currently Rust and Java)
   in the reference.  Do not read the other template.  Follow it through
   compile, package, and dry-run;
5. present the **confirmation preview** from the template.  Let the user
   edit any field (version, tag, package name) before confirming.
   Wait for an explicit "CONFIRM" — do not proceed on a simple "yes";
6. ask where the registry credentials are (environment variable, file path,
   or secret store).  Read them only from the location the user specifies;
7. run the publish command.  Report the result and the verification URL.
   Clean up `/tmp/ira-publish/<repo>` when done.  Do not retry a failed
   publish automatically.

## Boundaries

- Current executable adapter: `casbin-go` only.
- Do not execute arbitrary commands from JSON or introduce a host-test fallback.
- Do not store passwords, tokens, cookies, private keys, or private-list text.
- Never stage, commit, upload, or mount `<parent-workspace>/secretkey`; only GPG
  may access that directory during the separate sign step.
- Follow the engine's validation and stop on an actual command failure.
- Treat any byte change in the archive, checksum file, or signature file as a
  different candidate that requires a new RC and fresh votes.
- Signing, ASF writes, vote sending, and announcements always require explicit
  human authorization.
- Do not turn optional hardening ideas into blockers for the ordinary
  commit-to-RC workflow.
- Package-registry credentials (`CARGO_REGISTRY_TOKEN`, OSSRH, GPG) must come
  from environment variables or external secret storage; never from JSON
  configuration or repository files.
- Publishing to a public package registry is irreversible and always requires
  explicit human authorization.
- Never auto-create or auto-increment version numbers or tags; the version and
  tag are decided by the user or community vote.
- Consult `legacy/casbin-go-rc/` only as repository history; the bundled Go
  engine is the operational implementation.
