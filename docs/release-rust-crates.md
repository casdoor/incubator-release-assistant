# Rust / crates.io 发版流程：从准备到公告

适用范围：采用 ASF 源码投票、GitHub Release 和 crates.io 分发的孵化中
Rust 组件。以 [SQLx Adapter 1.9.0 案例](cases/sqlx-adapter-1.9.0.md) 为依据，
政策核验日期为 2026-10-20；下一次执行前重新核验政策、权限及目标仓库的
`RELEASING.md`。这是人工与 Agent 协作的操作指南，**不是已实现的 Rust 适配器**。
IRA 当前可执行的发布引擎仍只支持 Casbin Go；不能把 Rust 项目塞入 Go 配置，
也不能让配置携带任意 shell 命令。新增引擎能力须遵守
[adapter contract](adapter-contract.md) 并补齐契约测试。

## 1. 固定对象，先完成源码准备

建立同一份本地进度记录，至少包含仓库、组件、版本、RC、完整 commit、
源码包文件名及 SHA-512、签名指纹、平台包名、工作流 run/attempt、投票和
公告公开链接。每个阶段分别记录完成、待办、阻塞和下一动作；不能用一个
“发布成功”状态覆盖尚未完成的渠道。记录和原始日志留在忽略的本地数据目录，
公共贡献仅保留脱敏结论及公开出处。

对每个仓库/版本最多使用一个准备 PR。提交前集中完成：

- 版本、候选源码和贡献来源/历史审查；ASF 源码头、第三方文件及例外逐类说明。
- 依照实际分发内容核对 `LICENSE`、`NOTICE`、适用的孵化免责声明；检查嵌套
  许可证及 NOTICE 指针，逐项解释 RAT 排除和未覆盖内容。
- 依赖安全公告及许可证、构建/测试环境与失败边界；RAT 或构建通过不等于
  法律审查完成，也不代替投票者从签名源码包构建和验证。
- Cargo `package.name/version`、`lib.name`、功能开关、MSRV、仓库/主页、
  许可证、README、打包 include/exclude、示例和所需构建材料。
  包名、Rust 导入名和 ASF 归档名是三个独立身份。是否改名应有项目结论，
  不能凭 404 推断名称可用，也不能发明所有生态都必须采用同一前缀的规则。
- 全部发布入口：push/tag、release created/published、手工触发、重试、环境
  审批及 secret 引用。查明哪一步会写 GitHub Release 或上传 crates.io，避免
  未批准源码经移动标签或自动任务提前发布；第三方 Actions 审查见第 7 节。
- 回查历次公开评审并核实修复的最终源码。旧检查、旧承诺及旧 RC 的票不能
  自动成为当前候选的通过证据。

将源码/分发内容可能受影响的决定在 PR 前解决。PR 合并后重新绑定实际
commit、归档内容和校验结果；若出现新的源码阻塞，先报告，不静默追加
另一轮准备 PR。正式 RC 签名、上传和票据属于后续阶段，不能作为提交准备
PR 的前置材料。采用 IRA readiness 标准的工作区还须逐项完成 PREP-01～08
与 AUDIT-01～07，而非只读清单。

## 2. 权限一次查清，秘密由用户在本机配置

| 系统 | 实际需要 | 可验证内容与边界 |
| --- | --- | --- |
| GPG / ASF KEYS | 外置密钥目录、签名能力、官方 KEYS 中的公钥 | 核验真实指纹和独立公钥环；公钥存在不证明能签名 |
| ASF dist SVN | 目标目录写权限和本机交互认证 | 认证失败与授权失败分别排查；PPMC 身份不自动代表 SVN 已可写 |
| GitHub | 仓库/环境权限、必要的 Actions 审批和发布 secret | secret 名存在不证明值有效，PR 通过不等于环境已批准 |
| crates.io | 有发布权的账号/团队、已验证邮箱、适当范围且未过期的 token | 核查实际 crate owner；GitHub 身份或 PPMC 身份不自动授予 crate 发布权 |

Cargo 是 Rust 的构建和包管理工具，crates.io 是包注册表。token 在创建它的
crates.io 账号中管理，GitHub Actions 只引用已配置的 secret。本案例使用
`CARGO_REGISTRY_TOKEN`；其他项目以其审核后的工作流为准。用户在外部浏览器
登录 crates.io 的账号设置，确认邮箱显示已验证，并将 token 直接写入 GitHub
secret。需要 SVN 密码时使用本机交互提示；不把秘密传到聊天、命令参数、日志
或仓库，也不要求用户把现有 secret 读回来。

## 3. 生成、签名、暂存，验证候选包

先完成不带秘密的源码准备；目标项目的构建/测试在单独、明确授权的验证环境
运行，不能塞进 IRA `prepare`、`sign` 或 `stage`。签名与 ASF 写入维持独立的
授权边界，沿用本次任务已明确的授权，不为每项检查反复询问。

1. 从审核后的确定 commit 生成单一顶层目录的源码包，检查路径、法律文件、
   内容清单和源码可构建性。优先在准备阶段将正确版本提交到源码；任何确有
   必要的生成/替换都要预先审核、列明差异并在选票中披露。
2. 对最终字节运行 RAT 并完成其覆盖/排除审查、人工法律审查，以及从解压源码
   构建和测试的验证。分别记录工具输出和人工结论，不把二者互相冒充。
3. 使用仓库外的 GPG home 为归档签名，生成 SHA-512，独立验证签名及官方
   KEYS。暂存到 ASF dist-dev 后重新下载归档、`.asc`、`.sha512` 验证。
4. 固定 RC 的 commit、文件名和哈希。不覆盖或重建已暂存候选；内容变化使用
   新 RC，并重新走受影响的审查和投票。

## 4. PPMC / IPMC 投票与 RESULT

按项目约定安排 dev 上的 PPMC 阶段及 general 上的 IPMC 阶段，逐组件、逐
候选分别计票。通常保留至少 72 小时审查时间；仍须检查实际有效的 binding
票、反对意见和对应候选，不能仅凭时间、绿色 CI 或别人已发布就认定批准。
IPMC 阶段至少需要三个有效 binding +1，且正票多于负票；只有 IPMC 成员的
票在此阶段具有约束力。PPMC 阶段按项目流程独立核实，不能合并两阶段票数。
换 RC 不继承旧票。具体孵化规则见第 7 节的政策与指南。

结束时在**相应 VOTE 线程回复 RESULT**，列清候选、投票链接、有效票和结论。
保持邮件 `In-Reply-To` / `References`，读回 Sent MIME，并在公开存档确认真实
父子关系。回复线程是此流程的可追溯做法，不宣称为所有邮件客户端的技术限制。
RESULT 表示表决结果，尚不代表源码晋升、注册表上传或网站部署已经完成。

## 5. 获批后晋升与生态发布

1. 将**获批的同一组字节**从 dist-dev 晋升到 dist-release；记录 SVN revision。
   若最终目录已存在，先核对，不能盲目覆盖或重新打包。
2. 等待官方 downloads 可用，重新获取归档、签名、校验和及 KEYS，比较获批
   字节。网站下载入口和公告至少在上传后等待一小时；保守地从确认正式下载
   可用时计时，并保留时间证据。只看到 SVN commit 或 HTTP 200 不够。
3. 检查最终 tag 指向批准 commit；先检查 tag 会触发的所有发布动作，再在授权
   范围内创建。已有 tag 先比较，不能移动。GitHub Release 复用获批源码附件，
   检查版本、非 RC 展示、三个附件的名称及哈希；不可用自动生成源码 zip 替代
   投票包。本案例先晋升，再由最终 tag 工作流创建 Release 并发布 crate。
4. 在隔离的验证环境，使用本项目审核后的功能参数运行
   `cargo package --list` 及 `cargo publish --dry-run`。它们可能执行构建脚本，
   不是 IRA 安全归档操作，也不证明 token、邮箱或上传权限已正常。
   对实际待上传 `.crate` 检查源文件和必需法律文件；解释 Cargo 生成/规范化
   的 manifest、lockfile 和 VCS 元数据，不能只比较文件数量。
5. 通过审核后的发布工作流上传一次。下载注册表实际 `.crate`，核对注册表
   checksum、名称/版本、未 yank 状态、内容清单、法律文件和源码对应关系。
   `.crate` 与 ASF `.tar.gz` 是不同格式，不要求归档哈希相同；逐文件比较
   非生成源码，并审查所有生成/转换差异。
6. 建立干净 consumer，显式固定注册表版本，不使用本地 path/patch 替代依赖；
   查看锁文件来源，验证文档导入方式与最小业务行为。记录工具链/平台/数据库
   范围，不用一次 SQLite 成功代表全部后端或操作系统通过。

| 发生的问题 | 恢复方法 |
| --- | --- |
| SVN 缺认证或无权限 | 本机交互认证或请有权人员执行；先读回远端状态，避免重复晋升 |
| Release 成功而 crate 上传失败 | 分别记录两个渠道的状态；检查注册表实际版本，保留成功的 tag 和附件 |
| 上传报邮箱未验证 | 验证创建 token 的 crates.io 账号邮箱；修复后只重试失败的上传 job |
| 上传超时 / 版本已存在 | 先查 API 并下载现有包比对；匹配则记录已完成，不再次上传；不匹配立即阻塞并报告 |
| crate 网页 404 但 API 存在 | 对照 HTML Accept 请求与 API/实际下载结果；不能直接判定版本不存在或忽略坏链 |

crates.io 的同名同版本不可覆盖，GitHub 成功也不自动回滚注册表失败。恢复前
必须核对当前 run/attempt、commit、远端副作用和批准包；不把历史 run ID 当成
下次发布命令。这里的恢复只适用于原源码与原工作流不变的情况，新的源码
问题须重新评估候选及授权。

## 6. 官网、公告和关闭检查

插件可以集中在共享的 **Adapters and plugins** 小节：一个简短条目给出
名称、版本、Source / PGP / SHA-512，并复用 KEYS 入口。这是本次项目采用的
简洁展示方式，其他项目按其下载页布局处理。ASF 源码入口必须保留，不能
因为用户通常只用 Cargo 就省略。

网站 PR 的检查通过、合并、部署完成是三个状态。核对合并 SHA 对应的部署
run，再获取线上页面。验证正文的 URL 和锚点，镜像选择器返回 HTML 200
不等于源码已下载：继续获取实际归档，核对文件格式、哈希和签名。只检查
下载文章时，不报告整个网站已检查。

源码、注册表、官网均核验后发送独立的 `[ANNOUNCE]` 新线程。邮件应指向
项目源码下载页，包含简短项目说明和适用的孵化免责声明；发送到 ASF-wide
announce 列表时使用 `apache.org` 发件地址。按项目约定发送至适当列表，
Reply-To 指向开发列表。发送前检查 From/To/Cc、主题、正文及新线程头；发送
后读回 Sent MIME，再逐列表核对公开存档的内容和线程位置。邮件签名推荐与
发布物强制签名分开理解；不要因存档延迟重复发送。

完成条件逐项确认：批准结果公开、ASF 源码及签名可下载、最终 tag/附件对应
批准源码、注册表包与 consumer 通过、官网部署和实际链接有效、预定公告列表
存档已核对。若只剩外部审批/部署/投递，报告具体对象和后续核验方式；不得
虚构后台监控。用户要求暂停时保存检查点，等待明确恢复。

## 7. 规则来源与适用边界

- [ASF Release Policy](https://www.apache.org/legal/release-policy.html)：源码、
  签名、批准、正式分发、公告和一小时等待要求。
- [Incubation Policy](https://incubator.apache.org/policy/incubation.html) 与
  [Release Management Guide](https://incubator.apache.org/guides/releasemanagement.html)：
  孵化项目发布与审核；实际票数和候选以公开线程为证。
- [Distribution Guidelines](https://incubator.apache.org/guides/distribution.html)：
  生态平台身份和管理的指导。命名指导的 SHOULD、具体项目决定和强制政策
  分开记录，不能由一次 SQLx 处理推导所有项目必须改名。
- [Release Download Pages](https://infra.apache.org/release-download-pages.html)：
  项目下载页的源码、校验和、签名和验证指引。
- [GitHub Actions Policy](https://infra.apache.org/github-actions-policy.html)：
  外部 Actions 固定经审查的 commit SHA，并配置自动依赖更新；保留政策列明
  的第一方例外。PR 检查通过不是对组织设置或其他仓库合规的全面证明。
- [Cargo Publishing](https://doc.rust-lang.org/cargo/reference/publishing.html)：
  账号、邮箱验证、token、打包及不可覆盖版本。具体 GitHub job/secret 名称由
  项目工作流决定。

上述官方来源与 IRA 的工作流约束共同构成本指南。案例中的包名、构建参数、
网站布局和邮件线程做法须按其适用范围使用，不能升级为 ASF 的通用规定。
