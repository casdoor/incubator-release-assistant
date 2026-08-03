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
邮箱：hsluoyz@apache.org
备注：CODE SIGNING KEY
```

生成后查看完整fingerprint：

```powershell
gpg --list-secret-keys --keyid-format LONG
gpg --fingerprint hsluoyz@apache.org
```

运行结果：
```powershell
PS C:\Users\luo> gpg --fingerprint hsluoyz@apache.org
pub   rsa4096 2026-08-03 [SC]
D783 932F B7B9 DCBF 7CCE  E2EE 49FF B438 CD21 C87B
uid           [ultimate] Yang Luo (CODE SIGNING KEY) <hsluoyz@apache.org>
sub   rsa4096 2026-08-03 [E]
D886 8357 FB24 7FF7 1C5E  129B 7400 B397 7C7E 4EFC
```

## 2.更新官方KEYS

```
svn checkout --depth=files https://dist.apache.org/repos/dist/release/incubator/casbin casbin-release-keys
cd casbin-release-keys

// 上一步的fingerprint贴过来
Add-Content -LiteralPath .\KEYS -Value "" -Encoding ascii
gpg --armor --export D783932FB7B9DCBF7CCEE2EE49FFB438CD21C87B | Out-File -FilePath .\KEYS -Append -Encoding ascii

// 上传
svn diff .\KEYS
svn commit .\KEYS --username hsluoyz --no-auth-cache -m "[casbin] Add Apache release signing key"
```

## 3.查看公开结果

```powershell
cd ..
svn export --force https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS .\KEYS-public
gpg --show-keys .\KEYS-public
```

确认新KEY显示完整fingerprint和Apache邮箱，然后把fingerprint填写到
`release-rc\release.local.json`。
