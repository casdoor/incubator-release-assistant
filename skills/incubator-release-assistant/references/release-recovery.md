# Release setup and recovery routing

Read this reference when the engine or wrapper stops on incomplete setup or
resumable state. Match the observed condition, load the referenced knowledge,
and give the user one concrete next action.

| Observed condition | Knowledge to read | Next action |
| --- | --- | --- |
| config missing, Apache ID placeholder, fingerprint placeholder | `workspace-bootstrap.md` | create/fill non-secret config |
| private key unavailable or not found | `signing-key-setup.md` | inspect existing key or generate one |
| key is not RSA 4096, lacks Apache UID, is expired/revoked, or cannot sign | `signing-key-setup.md` | select or create a compliant key |
| fingerprint absent from official KEYS | `asf-keys-publication.md` | prepare public-key publication |
| secret directory is inside a Git worktree | `workspace-bootstrap.md` | move to a plain parent workspace |
| incomplete prepare state exists | this page | inspect evidence, then approve `--clean` only if unstaged data is disposable |
| prepared/signed/staged state exists | this page | re-run the same command and let IRA revalidate frozen bytes |
| frozen candidate bytes differ | this page | stop and use a new RC number |

Use the actual resolved paths and observed values. Do not show abstract
placeholders when the filesystem already supplies the answer.

## Response shape

```text
Current gate
<what failed and why it matters>

Expected paths
<config path>
<secret directory if relevant>
<public export or evidence path if relevant>

Files
<who creates each file and whether it is secret>

Next action
<one safe action the Agent can perform>

Approval
<the exact mutation that still needs user approval>
```

Never route active users to `legacy/casbin-go-rc/`. Those files are migration
history and do not match the current trust boundaries or resumable engine.
