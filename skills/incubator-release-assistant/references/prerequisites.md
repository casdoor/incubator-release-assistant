# Release prerequisites

Read this reference when `doctor` returns `IRA-DEPENDENCY-001` or
`IRA-PREFLIGHT-001`.

## Required commands

| Command | Used for | Required before |
| --- | --- | --- |
| `go` | run the bundled IRA engine | every command |
| `git` | fetch and archive the selected commit | `prepare` |
| `tar` | inspect the source archive | `prepare` |
| `java` | run Apache RAT | `prepare` |
| `svn` | download RAT/KEYS and write ASF dist dev | `doctor`, `prepare`, `stage` |
| `gpg` | inspect keys, sign, and verify | `doctor`, `sign` |
| `docker` or `podman` | isolate project tests | `prepare` |

Check what the current shell can see:

```powershell
go version
git --version
tar --version
java -version
svn --version --quiet
gpg --version
docker version
```

```bash
go version
git --version
tar --version
java -version
svn --version --quiet
gpg --version
docker version
```

Use `podman version` instead of `docker version` when the config selects
Podman. Install a missing command from the operating system's trusted package
source, then open a new terminal if the installer changed `PATH`.

`IRA-PREFLIGHT-001` can also mean that the public ASF `KEYS` check could not
reach its HTTPS/SVN endpoint. Show the underlying error, preserve the config
and key, and retry `doctor`; do not generate another key or RC for a transient
network failure.

## Agent response

Name only the missing command or failed public check, explain which release
step needs it, and give one platform-appropriate verification command. Do not
dump installation instructions for tools that already passed.
