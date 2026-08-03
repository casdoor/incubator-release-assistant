---
name: incubator-release-assistant
description: Turn a selected Apache Casbin Go commit into an Incubator release candidate, including guiding a new release manager through missing config, signing-key, and official KEYS setup; then prepare, resume, sign, upload, and verify the RC with the bundled IRA engine. Use this skill whenever a user wants to set up, prepare, resume, sign, upload, or verify a Casbin RC, even when they have not configured a key yet. The current implementation supports only Apache Casbin Go.
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
with the host Go toolchain from the disposable extracted source tree. Docker
and Podman are not required. A successful run prints the artifact SHA-512 and
records resumable state under `.ira/runs/`.

Every potentially long-running external command prints an `[IRA] START` record
with its command, working directory, evidence-log path, and child-process PID.
While it is still running, the engine prints an `[IRA] RUNNING` heartbeat every
30 seconds; completion prints `[IRA] DONE` or `[IRA] FAILED` with elapsed time.
The same records and live command output are written immediately to the named
`evidence/*.log` file.

If the user reports that `prepare` appears stuck, do not immediately restart it
or use `--clean`. Identify the most recent `[IRA] START`/`[IRA] RUNNING` record,
read the named evidence log, and inspect that exact child process or its network
access. Report the active phase, elapsed time, last log output, and safest next
action. Preserve the run directory so another Agent can continue the diagnosis.

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
- Consult `legacy/casbin-go-rc/` only as repository history; the bundled Go
  engine is the operational implementation.
