# Fourgetu v3.11.0 同步记录

同步日期：2026-10-27（Asia/Shanghai）。发布版本：`3.11.0-fourgetu.1`。

## 官方基线与优先策略

- 三方合并官方稳定标签 `v3.11.0`，官方发布时间为 2026-10-25 19:09:16 UTC。没有合入标签之后的 main 开发提交。
- Go 1.27.3、Xray v26.10.21、前端及文档依赖采用官方版本；官方已包含此前的依赖安全修补，不另行保留重复版本覆盖。
- 采用官方客户端 Host 绑定、Host 全局排序、SNI 池、CIDR 地址抽样、订阅设置缓存、共享客户端统计及 PostgreSQL 并发修复。
- 采用官方随机初始凭据、WireGuard/AmneziaWG 配置行校验、代理 IP 校验、内核 AmneziaWG、keepalive 区间、Telegram 重构与 XDRIVE 配置保护。
- 保留官方表单提示、无障碍校验和协议链接规则；定制控件补齐可访问标签。
- 官方没有等价的用户级独立上下行限速、推荐模板、中转向导和自然月快捷续期实现，继续保留这些功能。

## 定制功能兼容

| 功能               | 本次处理                                                                                                                                                                    |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 用户级限速         | 保留 GOST、限速表迁移、端口分配、恢复、解绑和删除清理。先应用官方客户端 Host 规则，再改写专属限速端口。                                                                     |
| 订阅与分享链接     | raw、JSON、Clash、单客户端链接及入站导出统一兼容 Host 绑定与限速；已禁用、删除或排除的绑定不回退到其他 Host。JSON 使用独立入站副本，避免客户端间端口污染。                  |
| 入站多客户端二维码 | 保留快捷入口和分客户端展示；协议 URL 改用后端官方链接生成器，正确应用 Host 绑定和限速。服务端失败不回退生成旧链接，空列表保持为空。保留隧道配置下载及官方超长 PQ 密钥规则。 |
| 推荐模板与中转向导 | 保留共享 subID、证书和 Reality 密钥读取、复用入口、远程部署及来源标识。                                                                                                     |
| 客户端管理         | 保留来源筛选、节点名和端口、重名编号、可选孤儿清理；共享客户端及其他入站限速不误删。                                                                                        |
| 快捷续期           | 保留 +1、+3、+6、+12 个自然月；从当前有效到期日顺延，已过期或未设置时从当前时间起算。保存才生效。                                                                           |
| 安装更新           | 保留 Fourgetu 下载源、固定标签更新脚本及数字修订号比较。                                                                                                                    |
| Docker/GOST        | 采用官方 ARM/386 架构映射；保留 GOST 3.3.0 校验和验证及正确运行文件名。                                                                                                     |
| 翻译               | 为官方新增 sk-SK 补齐定制功能键；全语言键集合检查保持启用。                                                                                                                 |

限速范围保持为本机 VLESS/VMess/Trojan 的 TCP/raw、WS、gRPC、HTTPUpgrade、XHTTP，以及 Shadowsocks TCP/UDP。远程入站在实际承载面板设置；MASQUE、XDRIVE 不支持此限速。

## 验证记录

- 后端全量 `go test -shuffle=on -count=1 ./internal/... ./tools/... .` 通过。
- 新增两种排序模式 × 六种 Host 状态 × 五种链接输出的兼容测试，验证限速与不限速客户端相互隔离，并重复请求覆盖设置缓存。
- 新增单入站链接的归属校验测试；二维码服务端链接、空绑定、失败保护与快捷续期专项共 16 项通过。
- 前端代码生成、TypeScript、Lint、格式检查、生产构建、Storybook 构建通过；生产依赖审计为 0 vulnerabilities。
- 前端全量单元、组件与 Storybook 浏览器测试：221 个文件、2,083 项全部通过。
- Go Lint（golangci-lint 2.15.0）为 0 issues；actionlint 和安装、更新、管理菜单、Docker 初始化脚本语法检查通过。
- 文档类型、Lint、格式检查及 12 个文件 / 112 项测试通过。
- 本机 Windows Go 测试链接器报告已有的多 manifest 合并警告，测试退出码为 0；正式 Windows 包另经云端构建和版本资源校验通过。

## 发布边界

本记录初始提交为候选状态，现已通过云端验证和八平台产物校验，并提升为稳定版 / Latest。没有更新生产服务器。

Docs Deploy 按原有设置保持手动禁用，Docs CI 保持启用。Docker 镜像工作流仍使用上游镜像仓库，独立于本次 GitHub 安装包发布。

本地 `3x-ui-cn-installer/` 与 `release-download-v2/` 不纳入主仓库提交。

## 云端发布核验

- 发布标签 `v3.11.0-fourgetu.1` 固定于三方合并提交 `d9d7ee4c0cddb38c0446d3ff3924873f3f032002`。后续文档提交不移动此标签。
- [发布源码 CI](https://github.com/Fourgetu/3x-ui/actions/runs/40527589825) 全部通过，涵盖 Go 全量、PostgreSQL、race、fuzz、节点 E2E、代码生成、Lint、漏洞扫描及前端全套验证。
- [发布源码 Docs CI](https://github.com/Fourgetu/3x-ui/actions/runs/40527596991) 通过；主分支 CodeQL 也通过。
- [八平台构建](https://github.com/Fourgetu/3x-ui/actions/runs/40529396309) 全部通过，包含 Windows 程序启动与版本资源检查。
- [产物验证与上传](https://github.com/Fourgetu/3x-ui/actions/runs/40530404409) 通过：校验来源 SHA、八包及八个 SHA-256 文件、GOST 文件名，并实际启动 Linux amd64 面板、GOST 和 Xray。
- 下载全部八个 SHA-256 文件，与 GitHub 附件摘要逐一核对一致。稳定版共 16 个附件。
- [主分支 dev 构建](https://github.com/Fourgetu/3x-ui/actions/runs/40529397572) 通过。
- 附带 Docker 工作流因未配置 Docker Hub 用户名和密码而失败，没有发布 Docker 镜像；不影响上述安装包。

- [中文安装器同步](https://github.com/Fourgetu/3x-ui-cn-installer/actions/runs/40530764617) 成功，提交 `afc73a0`；本地与公开 raw 版本记录均为 `v3.11.0-fourgetu.1`。中文脚本语法与翻译回归通过。
