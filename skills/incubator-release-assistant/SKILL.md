---
name: incubator-release-assistant
description: Turn a selected Apache Casbin Go commit into an Incubator release candidate, sign it, upload it to ASF dist dev, and verify the public files with the bundled IRA engine. Use this skill whenever a user wants to prepare, resume, sign, upload, or verify a Casbin RC. The current implementation supports only Apache Casbin Go.
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

1. Resolve this Skill directory and the user's workspace.
2. Read `references/configuration.md` when creating or changing config.
3. Copy `assets/examples/casbin-go.json` to an ignored project-local path if no
   active configuration exists. Never fill Apache ID, commit, or fingerprint by
   guessing; obtain them from the user or verified public state.
4. Use the bundled wrapper for the current platform. Both wrappers resolve a
   relative config path from the caller's working directory and run the same Go
   engine.

```powershell
& <skill-directory>\scripts\run.ps1 validate -Config <config>
& <skill-directory>\scripts\run.ps1 plan -Config <config>
```

```bash
bash <skill-directory>/scripts/run.sh validate --config <config>
bash <skill-directory>/scripts/run.sh plan --config <config>
```

Stop when the configuration does not validate. Tell the user exactly which
field must be supplied or corrected, then continue from the same step.

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
a passphrase.

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
- Follow the engine's validation and stop on an actual command failure.
- Treat any byte change in the archive, checksum file, or signature file as a
  different candidate that requires a new RC and fresh votes.
- Signing, ASF writes, vote sending, and announcements always require explicit
  human authorization.
- Do not turn optional hardening ideas into blockers for the ordinary
  commit-to-RC workflow.
- Consult `legacy/casbin-go-rc/` only as repository history; the bundled Go
  engine is the operational implementation.
