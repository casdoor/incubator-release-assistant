# RC脚本每步预期输出

脚本按固定顺序显示 `[1/15]` 到 `[15/15]`。任何一步出现 `[FAILED]` 都表示
已经停止，后面的步骤没有执行。

没有看到最后的“RC已完成制作、签名、dist dev上传和公开复验”，就不能发起
投票。

如果PowerShell在显示 `[1/15]` 之前就提示禁止运行脚本，可在当前窗口使用：

```powershell
Set-ExecutionPolicy -Scope Process Bypass
```

它只对当前PowerShell窗口生效。

## 1.读取并检查配置

正常结果：

```text
[1/15] 读取并检查配置
[OK] 配置有效：3.11.0-incubating-rc2 / <完整commit>
```

如果失败：

- `commit必须是40位完整Git commit`：不能填短commit。
- `gpgFingerprint必须是40位`：复制的是短key ID，不是完整fingerprint。

## 2.检查本机工具

正常结果：

```text
[2/15] 检查本机工具
[OK] 找到git
[OK] 找到tar
[OK] 找到go
[OK] 找到java
[OK] 找到gpg
[OK] 找到svn
```

缺少哪个命令，就先安装对应工具。不要跳过。

## 3.准备相对工作目录

正常结果：

```text
[3/15] 准备相对工作目录
[OK] 工作目录：.\work\3.11.0-incubating-rc2
[OK] 输出目录：.\output\3.11.0-incubating-rc2
```

如果提示目录已经存在，说明以前运行过。确认旧内容不再需要后，用 `-Clean`
重新运行。`-Clean`只清理当前版本和RC编号对应的 `work`、`output` 目录。

## 4.取得配置指定的源码commit

正常结果最后是：

```text
[OK] 取得commit：<与配置完全相同的40位commit>
```

如果失败：

- 检查仓库地址和网络。
- 确认该仓库中确实存在配置填写的commit。
- 配置可以填写HTTPS仓库地址，也可以填写相对于 `release\scripts` 的本地
  仓库路径。
- 脚本只确认commit存在，不判断它是否已经合并或是否适合发版。

## 5.生成源码包

正常结果：

```text
[5/15] 生成源码包
[OK] 已生成：.\output\3.11.0-incubating-rc2\apache-casbin-3.11.0-incubating-src.tar.gz
```

源码包由配置中的commit直接通过 `git archive` 生成，不是GitHub自动生成的
Source code压缩包。

## 6.解压并检查源码包内容

正常结果包括：

```text
[OK] 存在LICENSE
[OK] 存在NOTICE
[OK] 存在DISCLAIMER
[OK] 存在go.mod
[OK] 存在go.sum
[OK] 存在.rat-excludes
[OK] 按配置约定，前置许可证、RAT排除和logo检查已经完成
```

缺少任何一个文件都停止。最后一行表示脚本按约定直接采用配置指定的commit，
不再判断它的合并状态；它也不表示脚本能够自行判断logo权利或第三方许可证。

## 7.下载并运行Apache RAT

正常结果：

```text
[OK] Apache RAT下载校验通过
[OK] RAT：Unapproved 0 / Unknown 0
```

脚本通过ASF公开SVN下载并自动重试。如果下载校验失败，删除 `tools` 中的
RAT ZIP后重新运行。

如果RAT失败，查看最新的：

```text
.\evidence\<本次目录>\rat-report.txt
```

不能仅因为文件可以排除就忽略结果；需要先修复或确认 `.rat-excludes` 的依据。

## 8.运行Go测试

正常结果：

```text
[8/15] 运行Go测试
[OK] go test ./...通过
```

完整输出保存在：

```text
.\evidence\<本次目录>\go-test.txt
```

Go构建缓存和依赖缓存临时放在本次 `work` 目录，不使用或修改用户原有的全局
缓存。

任何包失败都不能继续签名。

## 9.生成并验证LF-only SHA512

正常结果：

```text
[OK] SHA512匹配，格式为无BOM、无CR、结尾一个LF
```

这一步同时检查：

1. SHA512数值正确。
2. hash后是两个空格和纯文件名。
3. 文件没有UTF-8 BOM。
4. 文件没有Windows CR。
5. 文件最后只有一个LF。

这就是为避免RC1在macOS执行 `shasum -c` 失败而增加的检查。

## 10.检查GPG私钥、Apache UID和官方KEYS

正常结果：

```text
[OK] GPG私钥存在，primary UID使用apache.org邮箱
[OK] 签名key符合ASF要求：RSA 4096 bit
[OK] 官方KEYS包含完整fingerprint
[OK] 官方KEYS包含本机primary UID
```

常见失败：

- `本机找不到...GPG私钥`：只有public key，没有用于签名的private key。
- `primary UID不是apache.org邮箱`：不能使用该key签署本次RC。
- `ASF新发布签名key必须是至少2048位RSA`：当前key类型或长度不符合ASF发布
  政策；新key建议使用4096位RSA。
- `官方KEYS中找不到...fingerprint`：先由有权限的人更新官方KEYS。
- `官方KEYS尚未包含本机key的primary UID`：本机虽然已经添加Apache UID，
  但官方KEYS还是旧内容，需要重新导出并更新官方KEYS。

不要把私钥或密码放进配置文件。

## 11.签名并使用官方KEYS验证

正常过程：

1. GPG可能弹出密码窗口。
2. 私钥持有人本人输入密码。
3. 最后显示：

```text
[OK] GPG签名有效，并且签名key来自官方KEYS
```

如果取消密码窗口、密码错误或签名key不匹配，脚本停止。

## 12.确认三个RC文件

正常结果只列出：

```text
apache-casbin-3.11.0-incubating-src.tar.gz
apache-casbin-3.11.0-incubating-src.tar.gz.asc
apache-casbin-3.11.0-incubating-src.tar.gz.sha512
```

每行还会显示文件大小。任何文件为空都会停止。

没有使用 `-Stage` 时，脚本在这里正常结束，并明确显示“尚未上传ASF dist
dev”。

## 13.检查ASF dist dev目标

只有使用 `-Stage` 才会执行。

正常结果：

```text
[OK] 远程RC目录不存在，可以创建
```

如果远程已经存在同名RC，脚本必须停止。不能覆盖已经上传并可能被投票引用的
RC目录；需要使用新的RC编号。

## 14.准备并提交ASF dist dev

提交前会显示 `svn status`，应当只有本次RC目录和三个文件，状态都是 `A`。

然后要求输入：

```text
STAGE RC2
```

只有完全一致才会继续。随后SVN询问Apache账号密码。密码只在SVN提示中输入，
不要写到JSON或PowerShell命令中。

正常结果最后包括：

```text
Committed revision <数字>.
[OK] ASF dist dev提交完成
```

鉴权失败通常表示：

- Apache ID或LDAP密码错误。
- 当前账号不是Casbin committer，或者没有对应dist dev权限。
- 本地缓存了错误的旧密码。

脚本不会为了测试权限而提前提交文件。

## 15.从公开地址重新下载并验证

正常结果：

```text
[OK] 公开下载的SHA512、GPG签名和源码包字节全部验证通过

RC已完成制作、签名、dist dev上传和公开复验。
投票地址：https://dist.apache.org/repos/dist/dev/incubator/casbin/3.11.0-incubating-rc2/
```

这一步验证的是SVN提交后公开服务器实际提供的文件，而不是本地原文件。

如果提交刚完成但下载暂时失败，先确认公开目录是否已经出现；不要重新提交或
覆盖同名RC。
