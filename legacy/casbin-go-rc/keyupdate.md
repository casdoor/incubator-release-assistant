# 更新发布签名KEY

这一步只执行一次

原因：查的时候发现apache官方推荐用RSA 4096，之前用的不是；所以更新一下

KEYS不能覆盖，需要保留所有使用过key的记录，所以需要重新上传，且只需要做这一次

## 1.生成新KEY

```powershell
gpg --full-generate-key
```

按提示填写：

```text
密钥类型：RSA and RSA
密钥长度：4096
真实姓名：Yang Luo
邮箱：<Apache ID>@apache.org
备注：CODE SIGNING KEY
```

生成后查看完整fingerprint：

```powershell
gpg --list-secret-keys --keyid-format LONG
gpg --fingerprint <Apache邮箱>
```

## 2.更新官方KEYS

```
svn checkout --depth=files https://dist.apache.org/repos/dist/release/incubator/casbin casbin-release-keys
cd casbin-release-keys

// 上一步的fingerprint贴过来
Add-Content -LiteralPath .\KEYS -Value "" -Encoding ascii
gpg --armor --export <完整fingerprint> | Out-File -FilePath .\KEYS -Append -Encoding ascii

// 上传
svn diff .\KEYS
svn commit .\KEYS --username <Apache ID> --no-auth-cache -m "[casbin] Add Apache release signing key"
```

## 3.查看公开结果

```powershell
cd ..
svn export --force https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS .\KEYS-public
gpg --show-keys .\KEYS-public
```

确认新KEY显示完整fingerprint和Apache邮箱，然后把fingerprint填写到
`release-rc\release.local.json`。
