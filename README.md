# Incubator Release Assistant（IRA）

IRA 帮助 Agent 准备、签名、上传并验证 Apache Casbin Go 候选版本（RC）。
如果配置、签名密钥或 ASF 官方 `KEYS` 文件尚未准备好，它会一次提示一个缺项，
并告诉用户下一步该做什么。

> IRA 目前只支持 Apache Casbin Go。签名、写入 ASF、投票和发布公告仍需人工
> 检查与确认。

## 快速开始

准备一个本身不是 Git 仓库的父目录：

```text
/abc/
├── Incubator-release-assistant/   本仓库
└── secretkey/                     独立的 GPG 主目录，绝不能提交
```

克隆仓库：

```bash
cd /abc
git clone https://github.com/casdoor/incubator-release-assistant.git Incubator-release-assistant
```

从 `/abc` 启动 Agent，然后发送下面这段提示词：

```text
请读取并遵循 ./Incubator-release-assistant/skills/incubator-release-assistant/SKILL.md。
先运行只读的 doctor，每次只解决它报告的一个检查项，帮助我准备 Apache Casbin Go RC。
生成私钥、签名或写入 ASF 前必须单独征得我的确认。
```

Agent 首先会运行以下命令之一：

```powershell
.\Incubator-release-assistant\ira.ps1 doctor
```

```bash
./Incubator-release-assistant/ira.sh doctor
```

`doctor` 是只读检查。它每次只报告一个缺项，同时给出实际路径、对应说明和
下一步操作。按照提示补充后重新运行，直到结果变为 `IRA-READY`。

## 需要准备什么

- 非敏感的发版配置：commit、版本/RC、Apache ID 和真实的签名密钥指纹；
- 仓库外部的 GPG 主目录，例如 `/abc/secretkey`；
- 已加入 ASF 官方 `KEYS` 文件的公钥。

不要把私钥、口令、令牌或其他凭据放进本仓库或 JSON 配置文件。

准备完成后，Agent 会检查计划、生成源码归档，并在签名和上传 ASF 前请求确认。
投票和最终发布仍由人来完成。

## 详细说明

- [Agent 工作流程](skills/incubator-release-assistant/SKILL.md)
- [依赖工具](skills/incubator-release-assistant/references/prerequisites.md)
- [工作区准备](skills/incubator-release-assistant/references/workspace-bootstrap.md)
- [配置文件](skills/incubator-release-assistant/references/configuration.md)
- [签名密钥准备](skills/incubator-release-assistant/references/signing-key-setup.md)
- [发布 ASF KEYS 文件](skills/incubator-release-assistant/references/asf-keys-publication.md)
- [中断与恢复](skills/incubator-release-assistant/references/release-recovery.md)
- [架构与安全模型](docs/architecture.md)

## 许可证

Apache License 2.0，参见 [LICENSE](LICENSE)。
