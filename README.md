# Incubator Release Assistant (IRA)

[English](README.md) | [简体中文](README.zh-CN.md)

IRA helps an Agent prepare, sign, upload, and verify an Apache Casbin Go
release candidate (RC). It guides the user through missing configuration,
signing keys, and the official ASF `KEYS` file one item at a time.

> IRA currently supports Apache Casbin Go only. Signing, ASF writes, voting,
> and announcements still require human review and approval.

## Quick start

Use a plain parent directory that is not a Git repository:

```text
/abc/
├── Incubator-release-assistant/   this repository
└── secretkey/                     external GPG home; never commit it
```

Clone the repository:

```bash
cd /abc
git clone https://github.com/casdoor/incubator-release-assistant.git Incubator-release-assistant
```

Start your Agent from `/abc` and send it this prompt:

```text
Read and follow ./Incubator-release-assistant/skills/incubator-release-assistant/SKILL.md.
Run its read-only doctor first, resolve one reported gate at a time, and help me
prepare the Apache Casbin Go RC. Do not generate a private key, sign, or write to
ASF without asking me at that boundary.
```

The Agent will first run one of these commands:

```powershell
.\Incubator-release-assistant\ira.ps1 doctor
```

```bash
./Incubator-release-assistant/ira.sh doctor
```

`doctor` is read-only. It reports one missing item, the exact path involved,
the relevant guide, and one next action. Follow the prompt and re-run it until
the result is `IRA-READY`.

## What you need to provide

- a non-secret release configuration: commit, version/RC, Apache ID, and the
  real signing-key fingerprint;
- an external GPG home such as `/abc/secretkey`;
- a public key that is present in the official ASF `KEYS` file.

Do not put private keys, passphrases, tokens, or credentials in this repository
or in its JSON configuration.

Once setup is ready, the Agent validates the plan, prepares the source archive,
and asks for confirmation at the signing and ASF upload boundaries. Docker and
Podman are not required: `prepare` runs `go test ./...` with the installed host
Go toolchain from the disposable extracted source tree. Host tests run with the
current user's permissions, so use a reviewed upstream commit. Voting and final
release promotion remain separate human-led steps.

## Guides

- [Agent workflow](skills/incubator-release-assistant/SKILL.md)
- [Prerequisites](skills/incubator-release-assistant/references/prerequisites.md)
- [Workspace setup](skills/incubator-release-assistant/references/workspace-bootstrap.md)
- [Configuration](skills/incubator-release-assistant/references/configuration.md)
- [Signing-key setup](skills/incubator-release-assistant/references/signing-key-setup.md)
- [Publishing the ASF KEYS file](skills/incubator-release-assistant/references/asf-keys-publication.md)
- [Recovery](skills/incubator-release-assistant/references/release-recovery.md)
- [Architecture and security](docs/architecture.md)

## License

Apache License 2.0. See [LICENSE](LICENSE).
