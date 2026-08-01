# RC脚本

此目录用于在Windows PowerShell中完成一次Apache Casbin RC的**制作、签名、**
**ASF dist dev上传**和公开复验。

## 1.准备配置

```powershell
cd release\scripts\release-rc
Copy-Item .\release.example.json .\release.local.json
notepad .\release.local.json
```

只填写仓库、完整commit、版本、RC编号、Apache ID和完整GPG fingerprint。

## 2.完成RC制作和上传

```powershell
.\release-rc.ps1 -Config .\release.local.json -Stage
```

脚本会完成打包、RAT、测试、SHA512、GPG签名、ASF dist dev上传和公开复验。

GPG密码由GPG窗口询问，Apache密码由SVN询问。上传前还需要输入一次：

```text
STAGE RC2
```

如果RC编号不是2，按屏幕显示输入对应编号。

如果只想制作和检查，不上传ASF，去掉 `-Stage`。

如果以前运行失败并且本地 `work`、`output` 目录还存在，确认旧内容无用后增加
`-Clean`：

```powershell
.\release-rc.ps1 -Config .\release.local.json -Stage -Clean
```

每一步的正常结果和失败含义见 `EXPECTED_OUTPUT.md`。
