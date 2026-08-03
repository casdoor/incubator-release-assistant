# Release setup and recovery routing

Read this reference when the engine or wrapper stops on incomplete setup or
resumable state. Match the observed condition, load the referenced knowledge,
and give the user one concrete next action.

| Code or condition | Knowledge to read | Next action |
| --- | --- | --- |
| `IRA-CONFIG-001` / config missing | `workspace-bootstrap.md` | copy non-secret config template |
| `IRA-CONFIG-003` / Apache ID or commit missing | `workspace-bootstrap.md` | collect only the missing public inputs |
| `IRA-KEY-001` / fingerprint or private key missing | `signing-key-setup.md` | inspect an existing key or approve generation |
| key is not RSA 4096, lacks Apache UID, is expired/revoked, or cannot sign | `signing-key-setup.md` | select or create a compliant key |
| `IRA-KEYS-001` / fingerprint absent from official KEYS | `asf-keys-publication.md` | prepare public-key publication |
| `IRA-WORKSPACE-001` / secret directory is inside Git | `workspace-bootstrap.md` | move to a plain external directory |
| `IRA-DEPENDENCY-001` or `IRA-PREFLIGHT-001` | `prerequisites.md` | install the one missing command or retry the public check |
| prepare is still running without recent output | this page | read the latest `evidence/*.log`, match its PID and command, and preserve the run |
| incomplete prepare state exists | this page | inspect evidence, then approve `--clean` only if unstaged data is disposable |
| prepared/signed/staged state exists | this page | re-run the same command and let IRA revalidate frozen bytes |
| frozen candidate bytes differ | this page | stop and use a new RC number |

Use the actual resolved paths and observed values. Do not show abstract
placeholders when the filesystem already supplies the answer.

## Response shape

```text
Current gate: <what failed and why it matters>
Paths: <only paths relevant to this gate>
Next: <one safe action the Agent can perform>
Need from you: <missing public input, choice, or approval; omit if none>
```

Keep this first response under about 12 lines. Explain file ownership when the
file first appears, and show longer commands only when executing that step.

Never route active users to `legacy/casbin-go-rc/`. Those files are migration
history and do not match the current trust boundaries or resumable engine.
