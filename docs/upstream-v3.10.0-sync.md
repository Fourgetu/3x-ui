# Fourgetu v3.10.0 同步记录

同步日期：2026-10-13（Asia/Shanghai）。发布版本：`3.10.0-fourgetu.1`。

## 基线与范围

- 官方最新稳定版核验：GitHub Release API 返回 `v3.10.0`，发布时间为 2026-10-12 14:44:53 UTC；标签提交为 `74c3cc38`。
- 从已发布的 Fourgetu `v3.9.0-fourgetu.2`（`c94b01db`）三方合并官方标签，工作分支为 `codex/sync-upstream-v3.10.0`。
- 保留真实合并历史。源码合并完成后，用户授权发布到 GitHub。分支和 main 采用正常快进推送，并创建新版本标签；不更新服务器。
- `3x-ui-cn-installer/` 与 `release-download-v2/` 是原有未跟踪目录，不纳入主仓库提交。中文安装器先生成缓存中的候选脚本并验证，主仓库稳定版发布后再执行其既有同步工作流。

## 官方优先与定制兼容

| 功能                           | 处理                                                                                                                                                                                  |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go 和前端依赖安全更新          | 官方已覆盖上一版定制依赖补丁，`go.mod`、`go.sum`、前端 package/lock 完整使用官方 v3.10.0 版本。                                                                                       |
| 用户级独立上下行限速           | 官方仍无等价功能，保留模型、数据库迁移、API、表单、订阅及分享链接端口映射、GOST 进程恢复与清理。                                                                                      |
| 推荐协议模板、中转向导         | 保留一键模板、共享订阅 ID、密钥/证书读取、远程部署、中转来源标识，继续使用官方 schema、节点和路由服务。                                                                               |
| 节点同步和客户端更新           | 使用官方修复：空节点快照保护、admin-token 节点推送、客户端原位更新、过期流量快照保护。                                                                                                |
| 入站二维码                     | 官方因删除单用户 Shadowsocks 路径而移除了旧弹窗；保留定制的多客户端入口，恢复 action 类型、客户端详情加载和节点地址传递，详情加载失败时不展示错误二维码；不恢复官方废弃的单用户逻辑。 |
| 分享链接、二维码可扫描判断     | 直接使用官方 `inbound-link.ts` 和 `isPostQuantumLink`；保留客户端名称，普通 REALITY 可显示二维码，超长 PQ 密钥隐藏二维码。MASQUE 无分享链接，隐藏该入站快捷入口。                     |
| 批量删除与孤儿清理             | 保留定制包装入口，内部调用官方批量删除；共享客户端、限速记录及官方节点同步逻辑继续受回归测试保护。                                                                                    |
| 客户端来源过滤、节点名称和端口 | 保留来源筛选、入站标签、端口和重复名称编号。                                                                                                                                          |
| 数据库迁移                     | 同时保留限速表和旧节点名称模板迁移，以及官方布尔时代证书校验字段、Shadowsocks 密钥/方法和 finalmask 迁移。                                                                            |
| 安装和更新                     | 合入官方升级前迁移、TLS 保持及 systemd 沙箱；下载源保持 Fourgetu，更新脚本按选定标签固定，版本比较保留数字修订号。                                                                    |

官方 MASQUE、XDRIVE、WARP over MASQUE、Lua DNS/路由、Xray v26.10.10、订阅掩码、ALPN 和主题修复完整合入。官方功能细节以 [v3.10.0 发布说明](https://github.com/MHSanaei/3x-ui/releases/tag/v3.10.0) 为准。

限速仍沿用原协议范围：本机 VLESS/VMess/Trojan 的 TCP、WS、gRPC、HTTPUpgrade、XHTTP，以及 Shadowsocks TCP/UDP；远程入站需在承载面板配置。没有把新 MASQUE 或 XDRIVE 标记为支持限速。

## 安装包兼容修复

检查发现原 Release 工作流没有打包限速必需的 GOST，而 Docker 中 386/ARM 的 GOST 名称与 `runtime.GOARCH` 不一致：

- Linux 七架构与 Windows amd64 发布流程增加 GOST 3.3.0，并在解压前核验官方 SHA-256 清单；缺失清单条目或校验不符即失败。
- ARMv5/v6/v7 安装包统一使用运行时查找的 `gost-linux-arm` 名称；386 使用 `gost-linux-386`。Docker 同步修正。
- 只提取 GOST 可执行文件，避免覆盖 Xray 包中的 LICENSE/README。
- 官方 systemd 单元保持原样。默认配置位于数据库目录 `gost/`，属于官方允许写入的路径。

## 验证

环境：Windows amd64，PowerShell 7，Go 1.27.2、Node 26.11.1、WinLibs GCC，`CGO_ENABLED=1`。

- Go 全量：`go test -shuffle=on -count=1 ./internal/... ./tools/... .` 通过，包括限速、订阅、数据库迁移、节点及核心业务。
- `golangci-lint run`：0 issues。
- `govulncheck ./...`：当前调用代码受影响漏洞为 0，导入包漏洞为 0；一个未调用的模块漏洞由工具报告，未强制偏离官方依赖版本。
- `npm run gen`、TypeScript、Lint、全量前端格式检查、生产构建通过。
- `npm audit --omit=dev --audit-level=moderate`：0 vulnerabilities。
- 前端全量单元/组件/Storybook 浏览器测试：193 个文件、1,893 项全部通过。随后补回二维码凭据详情加载，两个文件的 5 项针对性回归全部通过（其中 2 项新增，合计覆盖 1,895 项）；覆盖多客户端入口、节点地址/名称、REALITY 和 PQ 隐藏规则、详情加载成功及失败行为。
- CGO 主程序编译通过，`-v` 输出 `3.10.0-fourgetu.1`。
- 安装、更新、菜单、Docker 脚本 Bash 语法检查通过；中文安装器现有 URL 回归、候选中文脚本生成和语法检查通过。
- Release YAML 可解析；GOST Linux amd64 压缩包通过官方 SHA-256 核验。Windows Defender 拦截了测试压缩包，Linux 二进制本机重命名也遭拒绝，因此本机未完成 GOST 解压/运行及跨平台实际发布打包验证，未改变防护设置。

未在本机执行 PostgreSQL 专用用例、Linux systemd 沙箱实际运行、真实节点 E2E 或生产流量限速压测。后续云端构建与发布结果见下节。

## 升级注意事项

- 节点也应升级至 v3.10.0，才能使用官方此次节点端修复和 MASQUE。
- 出站使用私有 CA 时，启用官方 TLS 的 `Use system CA`。
- xdns/端口跳跃客户端应使用 Xray v26.10.10 或更新版本。
- 自定义 `XUI_GOST_CONFIG_PATH` 若位于官方可写目录之外，需要将其父目录加入管理员的 systemd drop-in `20-local.conf` 的 `ReadWritePaths=`；修改 `XUI_*_FOLDER` 后按官方流程重新运行更新脚本。
- README 的固定版本安装示例指向已发布的稳定版 v3.10.0-fourgetu.1。

## GitHub 发布流程

发布标签：`v3.10.0-fourgetu.1`。先构建 Linux 七架构和 Windows amd64，确认 8 个安装包和对应的 8 个 SHA-256 文件齐全、检查云端 CI，再将预发布提升为稳定版并设为 Latest。最后同步 Fourgetu/3x-ui-cn-installer，并核对版本记录、下载源及中文菜单。

## 云端完成核验（2026-10-13）

- GitHub 暂停了 fork 的 Actions，首次推送没有触发构建；用户恢复 Actions 后，从原始标签手动构建，不移动标签。
- 增加 CI/Docs CI 手动入口及 `Publish verified release artifacts` 恢复流程。上传前要求构建成功、工作流路径正确、构建提交与标签一致，校验 8 个平台的文件、SHA-256、GOST 名称及 Linux amd64 启动。主仓库后续工作流提交未改变标签内的产品代码。
- [八平台构建](https://github.com/Fourgetu/3x-ui/actions/runs/38403432039)：Linux amd64、386、arm64、armv5、armv6、armv7、s390x 和 Windows amd64 全部成功。
- [云端 CI](https://github.com/Fourgetu/3x-ui/actions/runs/38403615658)：Go 全量、PostgreSQL、node-e2e、race、fuzz、golangci、govulncheck、代码生成及前端全部成功。前端 194 个文件、1,895 项测试全部通过，生产/Storybook 构建及审计通过。
- [Docs CI](https://github.com/Fourgetu/3x-ui/actions/runs/38403615423) 通过。
- [产物校验与上传](https://github.com/Fourgetu/3x-ui/actions/runs/38404224854) 通过，包含 GOST 文件和 Linux amd64 面板/GOST 启动验证。
- 已从 Release 下载全部 8 个 SHA-256 文件，并与 GitHub API 公布的每个安装包 digest 逐一比对一致。
- Release `v3.10.0-fourgetu.1` 已提升为稳定版 / Latest，共 16 个附件。未更新任何生产服务器。

- [中文安装器同步](https://github.com/Fourgetu/3x-ui-cn-installer/actions/runs/38404634682) 成功，提交 a057224；版本记录和公开 raw 地址均为 v3.10.0-fourgetu.1，安装/菜单脚本 Bash 语法及 Fourgetu 下载源验证通过。
