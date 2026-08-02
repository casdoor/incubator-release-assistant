# Workspace bootstrap

Read this reference when `doctor` reports an unsafe workspace, a missing
configuration, or missing public inputs such as Apache ID or source commit.
When only the fingerprint is missing, read `signing-key-setup.md` instead.

## Supported layout

Run the Agent from a plain parent workspace. Keep the cloned tool, ignored
local config, external GPG home, and optional public-key export separate:

```text
/abc/
|-- Incubator-release-assistant/          cloned Git repository
|   |-- config/local/casbin.local.json    non-secret local config
|   `-- .ira/runs/...                     generated state and evidence
|-- secretkey/                            GPG-managed private home
`-- public-key/
    |-- apache-casbin-release-key.asc     generated public export
    `-- key-metadata.json                 generated public metadata
```

`/abc` must not be a Git checkout. `secretkey/` must be outside every Git
worktree. The wrapper creates that directory only when the user reaches the
separate signing phase.

## File ownership

| Path | Owner | Contents | Secret |
| --- | --- | --- | --- |
| `config/local/casbin.local.json` | user and Agent | commit, RC, Apache ID, public fingerprint | no |
| `secretkey/` | GPG | private keyring and trust database | yes |
| `public-key/apache-casbin-release-key.asc` | Agent via GPG | ASCII-armored public key | no |
| `public-key/key-metadata.json` | Agent | fingerprint, UID, algorithm, size, KEYS status | no |
| `.ira/runs/.../state.json` | IRA | resumable release state | no credentials |
| `.ira/runs/.../evidence/` | IRA | reports, logs, manifests, public downloads | no credentials |

Never ask the user to create individual files inside `secretkey/`. GPG owns its
internal `private-keys-v1.d/`, `pubring.kbx`, and `trustdb.gpg` files.

## Create the non-secret config

Linux/macOS:

```bash
mkdir -p /abc/Incubator-release-assistant/config/local
cp /abc/Incubator-release-assistant/config/examples/casbin-go.json \
  /abc/Incubator-release-assistant/config/local/casbin.local.json
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force `
  C:\abc\Incubator-release-assistant\config\local | Out-Null
Copy-Item `
  C:\abc\Incubator-release-assistant\config\examples\casbin-go.json `
  C:\abc\Incubator-release-assistant\config\local\casbin.local.json
```

The complete example is bundled at `../assets/examples/casbin-go.json`. Values
that normally need attention are:

```json
{
  "source": {
    "commit": "08dab401f7e78a3af923239fff1fcef20ab78464"
  },
  "release": {
    "version": "3.11.0-incubating",
    "rc": 2
  },
  "signing": {
    "apacheId": "replace-with-your-asf-id",
    "fingerprint": "REPLACE_WITH_40_HEX_FINGERPRINT"
  }
}
```

Do not guess the Apache ID or commit. Ask for public values that cannot be
verified. Obtain the fingerprint from the selected key in the next setup gate;
do not ask the user to invent it. Never add a password, passphrase, private key,
token, cookie, or SVN credential to JSON.

## Agent prompt

Resolve real absolute paths first, then use this shape:

```text
The release workspace is incomplete.

Config file:
<absolute config path>

Private GPG home:
<absolute external secret directory>

Public files I may generate later:
<absolute public-key export path>
<absolute public metadata path>

I will create the safe local config directory and copy the non-secret Casbin
template as part of this setup. I still need your ASF ID and selected source
commit. I will obtain the fingerprint from the signing key rather than asking
you to invent it.

Need from you: <ASF ID and source commit, if they cannot be verified>
```

Create only the safe directories and config copy. Do not generate or import a
private key as part of workspace bootstrap. Re-run `doctor`; after the public
inputs are complete it will route the missing fingerprint to
`signing-key-setup.md`. Run `validate` only after `doctor` reports ready.
