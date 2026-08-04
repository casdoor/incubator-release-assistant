# Repository map

本页说明仓库中每个受维护目录和文件的职责。`.git/`、`.gocache/` 和
`.ira/` 属于 Git 或本机运行数据，不是发布助手源码。

## Execution path

```text
ira.ps1 / ira.sh
  -> skills/incubator-release-assistant/scripts/run.ps1 / run.sh
     -> scripts/ira/cmd/ira/main.go
     -> internal/release/doctor.go
     -> internal/release/config.go
     -> internal/release/engine.go
     -> internal/release/state.go
     -> internal/release/runner.go
```

## Root files

- `README.md`: 项目入口，说明范围、依赖、基本命令、签名和原始字节。
- `AGENTS.md`: 给维护该仓库的编码 Agent 使用的约束和常用命令。
- `LICENSE`: 仓库本身使用的 Apache License 2.0。
- `.gitattributes`: 固定 Markdown、JSON、YAML、Go、Shell 和 SHA-512 文件为
  LF，PowerShell 文件为 CRLF。
- `.gitignore`: 排除本地配置、凭据、私钥、发布产物、运行证据和工具缓存。
- `ira.ps1`: 完整仓库的 Windows 入口，转交给 Skill 内 `scripts/run.ps1`。
- `ira.sh`: 完整仓库的 Linux/macOS 入口，转交给 Skill 内 `scripts/run.sh`。

这两个入口预期由仓库父目录调用。例如 Agent 在 `/abc` 工作，仓库位于
`/abc/Incubator-release-assistant`；包装脚本会把 `/abc/secretkey`
设置为外置 GPG home，并拒绝仓库内部或任何更大 Git 工作树内的密钥目录。
因此 `/abc` 本身应为普通目录，而不是 Git 仓库。

## `.github/`

- `.github/workflows/ci.yml`: 本仓库自身的轻量 CI（10 分钟上限），运行 IRA
  引擎的 `go test`、`go vet`、`gofmt` 检查、Bash 包装脚本语法检查和
  `scripts/validate-repository.ps1`；不构建也不测试目标发布项目。

## `config/`

这是供直接克隆完整仓库的用户使用的配置入口。

- `config/release.schema.json`: 配置文件的 JSON Schema，限定 Casbin、RC
  命名、RAT、签名、ASF 地址和本地运行状态目录。
- `config/examples/casbin-go.json`: 无秘密信息的示例。用户应复制到被忽略的
  `config/local/` 后填写 commit、RC、Apache ID 和签名 fingerprint。

这两个文件是仓库侧的主副本。修改后用 `scripts/sync-skill-assets.ps1`
同步到 Skill 的 `assets/`。

## `docs/`

- `docs/repository-map.md`: 本目录地图。
- `docs/architecture.md`: 说明 prepare、sign、stage 三段流程、状态恢复和
  当前 `casbin-go` 适配边界。
- `docs/security-model.md`: 罗列被保护资产、不可信输入、已有控制和剩余风险。
- `docs/adapter-contract.md`: 将来支持其他项目或语言时，新适配器必须遵守的
  接口和隔离规则；当前运行 Casbin RC 不需要阅读它。

## `scripts/`

- `scripts/sync-skill-assets.ps1`: 将根目录的 schema 和示例复制到自包含
  Skill，避免两套配置漂移。
- `scripts/validate-repository.ps1`: 比较两套配置副本的 SHA-256，检查 JSON
  可解析、Skill frontmatter 和必需资源是否存在，并确认引擎没有重新引入目标
  项目测试执行、轻量 CI 工作流仍包含必需检查。

## `skills/apache-incubator-handbook/`

独立的离线知识 Skill，不执行发布命令，也不会被发版 Skill 自动读取。

- `SKILL.md`: 根据问题选择一份相关参考资料，并规定哪些实时事实仍需核验。
- `agents/openai.yaml`: 客户端显示名称和默认提示。
- `evals/evals.json`: 月报、两阶段投票、第三方许可三个离线回答样例。
- `references/README.md`: 知识路由表和核验日期。
- `references/01-lifecycle-and-roles.md`: 孵化生命周期及 Board、IPMC、Sponsor、
  Champion、Mentor、PPMC、committer、contributor 角色。
- `references/02-governance-and-reporting.md`: 邮件列表、共识、投票术语、月报、
  status page、roster 和人员决策。
- `references/03-releases-voting-and-distribution.md`: 正式发布、两阶段投票、
  签名、KEYS、dist-dev/dist-release/downloads/archive。
- `references/04-ip-and-licensing.md`: CLA/CCLA/SGA、来源、许可证分类、
  LICENSE/NOTICE、源码头和 RAT 解读。
- `references/05-branding-and-websites.md`: 名称、免责声明、商标、网站、下载页
  和宣传。
- `references/06-community-graduation-and-retirement.md`: 社区健康、毕业条件和
  流程、退休。
- `references/07-infrastructure-accounts-and-security.md`: ASF 基础设施、账号、
  邮件列表、公开记录、秘密和漏洞处理。
- `references/official-sources.md`: ASF 官方来源索引及必须查询实时状态的边界。

## `skills/incubator-release-assistant/`

这是可以独立安装和执行的 Agent Skill。Go 引擎放在 Skill 内部，因此安装
Skill 后不依赖仓库根目录源码。

- `SKILL.md`: Agent 的正常工作流：选 commit、prepare、sign、stage、公开
  复验，最后提醒作者亲自检查 RAT 和法律文件。
- `agents/openai.yaml`: Skill 在客户端显示的名称、短描述和默认提示语。
- `evals/evals.json`: 行为样例，覆盖 prepare、恢复后签名、平台入口、外置
  密钥布局，以及缺配置、缺私钥、未发布 KEYS 和不安全目录的引导。
- `references/configuration.md`: 仅在创建或修改配置时读取的字段说明。
- `references/workspace-bootstrap.md`: 缺配置或目录不安全时，说明标准工作区、
  文件职责、示例和安全创建步骤。
- `references/signing-key-setup.md`: 缺少或不合规签名密钥时，说明外置 GPG
  home、RSA 4096、Apache UID、公钥导出和用户提示。
- `references/asf-keys-publication.md`: 本地密钥合规但官方 `KEYS` 缺失时，
  说明公开导出、SVN diff、确认、提交和公网回读。
- `references/prerequisites.md`: `doctor` 发现工具或公网检查缺失时，只说明
  当前缺项、用途和复验命令。
- `references/release-recovery.md`: 将配置、密钥、KEYS 和续跑错误路由到唯一
  下一步。
- `assets/release.schema.json`: `config/release.schema.json` 的 Skill 内镜像。
- `assets/examples/casbin-go.json`: 根目录 Casbin 示例的 Skill 内镜像。
- `assets/examples/key-metadata.example.json`: 不含秘密的公钥元数据示例。
- `assets/examples/doctor-report.example.json`: 结构化缺项、路径和下一步示例。
- `scripts/run.ps1`: Windows PowerShell 5.1+ 入口，解析外置
  `-SecretDirectory`。
- `scripts/run.sh`: Linux/macOS Bash 入口，解析外置 `--secret-dir`。

### Bundled Go engine

- `scripts/ira/go.mod`: Go 模块身份和最低 Go 版本；当前无第三方 Go 依赖。
- `scripts/ira/cmd/ira/main.go`: CLI 命令分发，提供 `validate`、`plan`、
  `doctor`、`prepare`、`sign`、`stage`、`verify-public` 和 `version`，失败时
  输出稳定错误码、知识页和下一步。
- `scripts/ira/internal/release/doctor.go`: 只读检查工作区、配置、依赖、本地
  签名密钥和官方 `KEYS`，输出真实路径和唯一下一步。
- `scripts/ira/internal/release/config.go`: 读取、严格校验配置，并派生运行 ID、
  产物名和 `.ira/runs/...` 状态目录，同时验证外置密钥目录不在仓库内。
- `scripts/ira/internal/release/engine.go`: 核心流程；克隆 commit、打包、解压、
  RAT、SHA-512、GPG 签名、SVN 上传和公开复验都在这里编排；`prepare` 不执行
  目标项目代码，项目测试由所选 commit 自身的 GitHub CI 负责。
- `scripts/ira/internal/release/state.go`: 保存 prepare/sign/stage/publicVerified
  状态以及三个候选文件的摘要，用于安全续跑。
- `scripts/ira/internal/release/runner.go`: 统一启动 Git、tar、Java、GPG、
  SVN 等外部命令，实时记录命令、工作目录、日志路径、进程 PID、30 秒心跳和
  耗时，并严格解析 LF-only SHA-512 文件。
- `scripts/ira/internal/release/config_test.go`: 当前测试集合，覆盖配置约束、
  CRLF/BOM 校验、状态保存、`prepare` 不重跑目标项目测试的边界、错误确认和
  文件字节变化等。
- `scripts/ira/internal/release/doctor_test.go`: 覆盖缺配置、公开输入、缺指纹、
  不安全工作区和错误码路由。

## `legacy/casbin-go-rc/`

旧版 Windows PowerShell 单脚本流程，只保留为迁移依据，不被新 CLI 或 Skill
调用。

- `README.md`: 解释为什么保留旧版以及各文件角色。
- `release-rc.ps1`: 旧的 15 步 RC 制作、签名、上传和公开复验实现。
- `release.example.legacy.json`: 旧脚本使用的简化配置样例。
- `LEGACY_USAGE.md`: 旧脚本的简要人工操作方法。
- `EXPECTED_OUTPUT.md`: 旧脚本每一步的正常输出和失败解释。
- `keyupdate.md`: 一次性生成 RSA 发布密钥并追加到官方 `KEYS` 的旧操作指南。

这些中文文件是 UTF-8；旧版 Windows PowerShell 默认读取无 BOM UTF-8 时可能
显示乱码，文件内容本身并未损坏。

## Local generated directories

- `.git/`: Git 元数据。
- `.gocache/`: 本地 Go 编译缓存，由测试产生且已忽略。
- `.ira/`: 本地运行状态、证据、工作目录和 Go 缓存；已忽略，不能提交。

`/abc/secretkey/` 不在本仓库目录中，因此没有出现在上面的文件树。它只保存
发布者的密钥材料；Git、prepare 和 ASF 上传步骤都不应主动读取或复制它。

实际 RC 运行后，`.ira/runs/<project>-<version>-rc<n>/` 会进一步包含
`state.json`、`artifacts/`、`work/` 和 `evidence/`。

## Scope separation

发版 Skill 的配置和引擎只覆盖 dist-dev RC：选 commit、prepare、sign、stage
和公开复验。投票、正式发布、治理、IP、品牌、毕业等知识只在独立 handbook
Skill 中出现。旧 Windows 单脚本仍放在 `legacy/`，不会进入新执行链。
