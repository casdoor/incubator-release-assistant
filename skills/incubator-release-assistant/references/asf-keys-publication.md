# ASF KEYS publication

Read this reference when a compliant local signing key exists but its full
primary fingerprint is absent from the official Casbin `KEYS` file.

## Inputs and target

```text
Local public export:
/abc/public-key/apache-casbin-release-key.asc

Official public file:
https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS
```

Only the ASCII-armored public key is appended. The private GPG home is never
copied into the SVN checkout, diff, evidence, or repository.

## Prepare a reviewable update

Windows PowerShell example:

```powershell
svn checkout --depth files `
  https://dist.apache.org/repos/dist/release/incubator/casbin `
  C:\abc\keys-svn

$keys = 'C:\abc\keys-svn\KEYS'
$public = 'C:\abc\public-key\apache-casbin-release-key.asc'
$publicText = [IO.File]::ReadAllText($public, [Text.Encoding]::ASCII)
[IO.File]::AppendAllText($keys, "`n" + $publicText, [Text.Encoding]::ASCII)

svn diff C:\abc\keys-svn\KEYS
svn status C:\abc\keys-svn\KEYS
```

Linux/macOS example:

```bash
svn checkout --depth files \
  https://dist.apache.org/repos/dist/release/incubator/casbin \
  /abc/keys-svn
printf '\n' >> /abc/keys-svn/KEYS
cat /abc/public-key/apache-casbin-release-key.asc >> /abc/keys-svn/KEYS
svn diff /abc/keys-svn/KEYS
svn status /abc/keys-svn/KEYS
```

Show the complete diff, target URL, Apache ID, primary fingerprint, primary UID,
and public export path. Require exact approval such as:

```text
PUBLISH KEY 0123456789ABCDEF0123456789ABCDEF01234567
```

Only after that approval, commit with the user's Apache ID and no auth cache:

```text
svn commit <KEYS-path> --username <apache-id> --no-auth-cache \
  -m "Add release signing key for Apache Casbin"
```

SVN may obtain credentials interactively. Never put them in JSON or evidence.

## Public read-back

After commit, export the official file again into a fresh short temporary GPG
home. Import it and require the full primary fingerprint and Apache UID before
marking key setup complete. Record only public metadata, the SVN revision,
official URL, and verification time.

## Agent prompt

```text
The local signing key is compliant, but its fingerprint is not present in the
official Casbin KEYS file.

Fingerprint:
<full primary fingerprint>

Public export:
<absolute public-key export path>

Official target:
https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS

I can prepare an SVN checkout and append-only diff for review. I will request a
separate exact confirmation before committing, then re-download KEYS into a
fresh temporary keyring and verify the fingerprint and Apache UID.

May I prepare the KEYS update for review?
```

When public read-back succeeds, update public metadata, re-run `doctor`, and
continue from the same release config when it reports `IRA-READY`. Do not
regenerate the key or candidate.
