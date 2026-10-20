# SQLx Adapter 1.9.0：一次完整发版及收尾

记录日期：2026-10-20。对象为 Apache Casbin SQLx Adapter (Incubating) 1.9.0
RC2。正式源码、crate、官网和预定三个列表公告已完成核验。本文保存公开可查的
结论和有限范围的验证说明，不包含本机运行证据、凭据或私人讨论。
下次发布使用[可复用流程](../release-rust-crates.md)，不能重放本文的 tag/run。

## 发布对象

| 对象 | 本次值 / 公开入口 |
| --- | --- |
| 仓库与准备修复 | [apache/casbin-sqlx-adapter，PR #116](https://github.com/apache/casbin-sqlx-adapter/pull/116) |
| 批准 commit | `29ff5a3632be9f649a68f2f7c6241a9704e2ab4e` |
| 候选 / 最终标签 | `v1.9.0-rc2` / [v1.9.0](https://github.com/apache/casbin-sqlx-adapter/releases/tag/v1.9.0)，均指向上述 commit |
| 正式源码 | [apache-casbin-sqlx-adapter-1.9.0-incubating-src.tar.gz](https://downloads.apache.org/incubator/casbin/sqlx-adapter-1.9.0-incubating/apache-casbin-sqlx-adapter-1.9.0-incubating-src.tar.gz) |
| 完整性文件 | [PGP](https://downloads.apache.org/incubator/casbin/sqlx-adapter-1.9.0-incubating/apache-casbin-sqlx-adapter-1.9.0-incubating-src.tar.gz.asc)、[SHA-512](https://downloads.apache.org/incubator/casbin/sqlx-adapter-1.9.0-incubating/apache-casbin-sqlx-adapter-1.9.0-incubating-src.tar.gz.sha512)、[KEYS](https://downloads.apache.org/incubator/casbin/KEYS) |
| 注册表包 | [casbin-sqlx-adapter 1.9.0](https://crates.io/crates/casbin-sqlx-adapter/1.9.0)，[版本及 checksum API](https://crates.io/api/v1/crates/casbin-sqlx-adapter/1.9.0) |
| Rust 导入身份 | 库名 `sqlx_adapter`，类型 `SqlxAdapter`；与 Cargo 包名分开核验 |
| 项目下载页 | [Casbin Downloads](https://casbin.apache.org/docs/downloads/) |

源码归档大小为 24,218 字节，SHA-512 为：

```text
2e4ef1097e7f4f4a6877ea427c60ed8b0153474f0da266be3a4e81002a9adb69d00002c707041f05aeec4492e16c57fefee39944da7fe711aaaaa529641956b3
```

注册表实际 `.crate` 的 SHA-256 为：

```text
442ed51b853b872a44e9a0144a588acaf65785f1f997c10fc1152d0cd0721ce3
```

这些数值标识本次历史对象，不能用来证明未来同名 URL 的内容或当前镜像状态。
旧版本离开 downloads 后，应到 ASF archive 按同样文件名和哈希查验。

## 从投票到收尾的公开记录

| 阶段 | 已核实结果与来源 |
| --- | --- |
| PPMC 阶段 | [dev VOTE](https://lists.apache.org/thread/d5nmn4r0r5554n5xfzoc3xj7zdxks5ol)；该线程明确限于 SQLx RC2，成功后另交 IPMC 审核，与 JCasbin 分开 |
| IPMC 表决 | [general VOTE](https://lists.apache.org/thread/g756s1mn0qg3l6ooc1ccqthc0zztvwj8)；RC2 有三个 IPMC binding +1，无 0 / -1；不继承旧 RC 票 |
| 表决结果 | [RESULT](https://lists.apache.org/thread/0cqwtqkr9xvny3gsbd1rtx629gs9mow3)；结果与后续实际分发分开记录 |
| ASF 正式源码 | SVN 晋升 r88535，2026-10-13 15:10:41 UTC；[正式目录](https://dist.apache.org/repos/dist/release/incubator/casbin/sqlx-adapter-1.9.0-incubating/)与下载副本复验一致 |
| GitHub / crate | [发布工作流](https://github.com/apache/casbin-sqlx-adapter/actions/runs/38420853592)：GitHub Release 于 10-13 完成；crate 首次上传失败，邮箱验证后 attempt 2 于 10-14 成功 |
| 网站 | [PR #493](https://github.com/apache/casbin-website/pull/493) 于 10-19 合并；[合并后的部署](https://github.com/apache/casbin-website/actions/runs/39279872647)成功，线上下载页已核对 |
| 公告 | 10-20 独立 ANNOUNCE 新线程；[dev](https://lists.apache.org/thread/frrxkz6tdtcqlyqggjnvl46m3q2c3dks)、[general](https://lists.apache.org/thread/4xqvh028y8dtxkbln851rfbn8hvz0o42)、[announce](https://lists.apache.org/thread/j61o5zqb2cpt2ysc24jso6lmdhy33ymh) 三处公开存档均已核对 |

以上时间均为 UTC。每个渠道分别确认，网站合并前与公告归档前的状态曾是
“仍待收尾”；这不影响已经发布的 crate 可用，也不能把当时状态写成全部完成。

## 问题、处理与可复用教训

### 源码身份与打包转换

RC2 选票披露了 `Cargo.toml` 的版本由标签源码中的 `0.0.0` 替换为发行包中的
`1.9.0`；其余跟踪文件与批准 commit 一致。本次包包含 26 个源码文件。
此转换是已披露的历史处理，不是 IRA 对新候选的默认规则；今后应优先在准备
PR 中提交正确版本，任何转换都需事先说明和审查。

本地预打包曾得到 20 文件清单，但注册表真实包包含 29 个文件。不能把较早的
清单当成上传结果：逐项下载对账后，26 个非生成文件与批准源码逐字节对应，
另外 3 个为 Cargo 规范化 `Cargo.toml`、`Cargo.lock` 和 `.cargo_vcs_info.json`。
规范化 manifest 的包名/版本另行检查；VCS 元数据的 commit 一致，dirty 为
true，与上述版本替换相符。记录差异的原因，不能为了“干净”而隐去它。

### 权限与部分成功

SVN 最初因缺少认证失败，用户在本机交互认证后完成晋升。GitHub 中已配置
发布 token，却不保证 crates.io 邮箱已验证。首次上传在构建后收到
`A verified email address is required`，此时 GitHub Release 和附件已经存在。

处理方式是验证**创建 token 的 crates.io 账号**邮箱，查询注册表确认目标版本
尚未存在，然后仅重试失败的 Publish job。保留成功的 Release、tag 和获批
源码；没有重新打包、移动标签或重复投票。这个案例只证明本次权限链路成功，
不证明 PPMC 已共同管理全部平台账号。

### 官网检查与实际部署

最初 crates.io 页面在普通检查请求中返回 404，而 API 已显示版本存在；对照
已知 crate 后发现 HTML Accept 请求返回 200。曾验证 URL 范围的 header 修复，
最终网站采用更小的变更：共享插件区中仅一行名称/版本与 Source / PGP /
SHA-512，因此未保留无用的 crates 链接检查补丁。没有忽略 404 或将坏链当成功。

网站外部发布 Action 固定到已审核 SHA，并补自动 Actions 依赖更新。最终 PR
涉及下载页及这两项工作流维护，没有为单个插件增加独立宣传页。具体策略见
PR；这是一次范围明确的修复，不是全站或全组织合规证明。

PR 的检查成功之后，合并后的部署仍曾排队。等实际部署成功，再检查线上内容。
下载镜像选择器的 HTML 200 只证明选择页可访问，必须继续下载归档验证。本次
下载文章范围内 9 个 HTTP 链接与 2 个锚点有效，并重新校验了其中源码包的
SHA-512、签名及 SQLx 三个文件与获批原件的一致性；未做全站爬取。

### RESULT、ANNOUNCE 与完成边界

RESULT 用于回复对应投票并报告计票，ANNOUNCE 用于告知用户已可下载的发行，
本次作为新线程发送。发送成功后还检查了 Sent MIME 与三个列表的公开正文、
Message-ID 和根线程位置。公告含官方源码下载页、项目简介及孵化免责声明，
不是只发 crates.io 链接。没有因归档延迟重复发送。

## 验证边界和下一次执行

- 本次源码验证曾覆盖 SQLite / MySQL / PostgreSQL 与四种 runtime/TLS
  组合，每组合 3 个测试，共 36 项通过；不是所有平台矩阵的证明。
- 最终干净 consumer 使用 Windows GNU Rust 1.94，从注册表解析固定 1.9.0，
  核对锁文件来源和 `sqlx_adapter::SqlxAdapter`，验证 SQLite 策略写入、重载、
  删除以及允许/拒绝行为。这个 consumer 不覆盖全部数据库后端。
- 技术结果不代替来源、法律文件和人工投票结论；公开入口不承诺永久保持在
  当前托管位置。本案例不声称 IRA 已经自动执行 Rust 发布。

下一次复用时，重新完成准备审查、权限检查、候选验证和两阶段表决，将所有
commit、包哈希、run、部署与邮件链接换成本次真实对象。保留上述失败分支的
恢复检查，不复制旧票、旧 token、历史包或历史执行状态。
