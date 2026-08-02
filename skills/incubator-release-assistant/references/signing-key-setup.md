# Signing-key setup

Read this reference when the configured private key is unavailable, the
fingerprint does not match, or the selected key fails the RSA, UID, expiry, or
signing-capability policy.

## Required result

The selected key must have:

- a full 40-hex primary fingerprint;
- RSA with at least 4096 bits for the Casbin release policy;
- an Apache email UID;
- signing capability;
- no revocation or expiration at release time.

The private key remains in the external GPG home. Only its public key and
public metadata are exported.

## Files and directories

```text
/abc/secretkey/                              GPG-managed private home
/abc/public-key/apache-casbin-release-key.asc  public export
/abc/public-key/key-metadata.json              public facts
```

Example public metadata (`assets/examples/key-metadata.example.json`):

```json
{
  "apacheId": "example-asf-id",
  "primaryUid": "Example User <example-asf-id@apache.org>",
  "primaryFingerprint": "0123456789ABCDEF0123456789ABCDEF01234567",
  "algorithm": "RSA",
  "bits": 4096,
  "canSign": true,
  "expired": false,
  "revoked": false,
  "presentInOfficialKeys": false
}
```

This file is public. It must not contain a passphrase, keygrip, private-key
packet, token, or credential.

## Existing key

If the user already has a dedicated external GPG home, inspect public metadata
with its absolute path. Do not silently fall back to a default home and do not
copy the home into the repository. If it is compliant, pass that directory as
`-SecretDirectory` or `--secret-dir` during signing.

Windows example:

```powershell
gpg --homedir C:\abc\secretkey --batch --with-colons `
  --list-secret-keys <40-hex-fingerprint>
gpg --homedir C:\abc\secretkey --fingerprint <40-hex-fingerprint>
gpg --homedir C:\abc\secretkey `
  --output C:\abc\public-key\apache-casbin-release-key.asc `
  --armor --export <40-hex-fingerprint>
```

Linux/macOS example:

```bash
gpg --homedir /abc/secretkey --batch --with-colons \
  --list-secret-keys <40-hex-fingerprint>
gpg --homedir /abc/secretkey --fingerprint <40-hex-fingerprint>
gpg --homedir /abc/secretkey \
  --output /abc/public-key/apache-casbin-release-key.asc \
  --armor --export <40-hex-fingerprint>
```

## New key

Key generation is a private-key mutation. Explain the target directory and
policy, then obtain approval before starting GPG.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force C:\abc\secretkey | Out-Null
New-Item -ItemType Directory -Force C:\abc\public-key | Out-Null
gpg --homedir C:\abc\secretkey --full-generate-key
```

Linux/macOS:

```bash
mkdir -p /abc/secretkey /abc/public-key
chmod 700 /abc/secretkey
gpg --homedir /abc/secretkey --full-generate-key
```

In the GPG interaction select RSA, 4096 bits, and an Apache email as the primary
UID. GPG obtains the passphrase directly; the Agent must not request, echo,
store, or archive it. After generation, inspect the key, obtain the full
primary fingerprint, export the public key, and update only the public
`signing.fingerprint` field in the local config.

## Agent prompt

```text
The release configuration is valid, but no compliant signing private key is
available.

Private GPG home:
<absolute external secret directory>

Public files I will generate:
<absolute public-key export path>
<absolute public metadata path>

Required policy:
- RSA 4096
- Apache email UID
- signing capability
- not expired or revoked

Choose one path:
1. Inspect an existing dedicated external GPG home.
2. Generate a new key in the directory above.

GPG will ask you for any passphrase directly. I will not receive or store it.
May I prepare the selected path?
```

Once the key is compliant, write only its public fingerprint to the local
config and re-run `doctor`. If it returns `IRA-KEYS-001`, continue with
`asf-keys-publication.md` rather than attempting to sign.
