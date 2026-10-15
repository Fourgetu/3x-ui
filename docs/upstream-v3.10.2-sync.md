# Fourgetu v3.10.2 同步记录

同步日期：2026-10-15（Asia/Shanghai）。候选发布版本：`3.10.2-fourgetu.1`。

## 基线与官方优先

- GitHub Release API 核验官方最新稳定版为 `v3.10.2`，发布时间为 2026-10-14 15:19:26 UTC，标签提交为 `608ef4e4`。
- 从 Fourgetu v3.10.0 同步分支与 main 的共同提交 `a47d9c46` 三方合并，保留真实合并历史；工作分支为 `codex/sync-upstream-v3.10.2`。
- 完整采用官方 Shadowsocks 客户端热更新、节点流量同步减少重复写入、共享客户端 flow 和隧道字段归属修复。
- 中转向导继续使用官方链接解析器，直接获得 VMess XHTTP、mKCP 和 finalmask 导入修复；订阅继续使用官方 Hysteria2 Gecko 和客户端路由头隔离修复。
- 完整采用官方 MTProto 单 IP 待握手上限、Xray 完整崩溃日志、赞助页面调整和 Windows OpenSSL 更新。
- 官方此次没有用户级限速、推荐模板或中转向导的等价实现，这些功能继续保留。
- 官方发布说明：[v3.10.2](https://github.com/MHSanaei/3x-ui/releases/tag/v3.10.2)。Go、前端依赖和 Xray 版本与官方标签一致。

## 定制功能兼容

| 功能               | 处理                                                                                                                 |
| ------------------ | -------------------------------------------------------------------------------------------------------------------- |
| 独立上下行限速     | 保留 GOST、数据库表和迁移、恢复与端口回收、订阅和分享链接映射。适配官方客户端解绑函数新的返回值。                    |
| 官方共享客户端状态 | 保留官方有效 flow 重新计算；新增全量同步和增量解绑两种回归，确认只清理解绑入站的限速，保留另一入站的限速及客户端。   |
| 推荐模板与中转向导 | 保留共享订阅 ID、证书和 Reality 密钥读取、复用入口、远程部署与来源标识；链接解析复用官方修复。                       |
| 管理与二维码       | 保留来源过滤、节点名称和端口、同名编号、可选孤儿清理，以及入站多客户端二维码和凭据加载保护。                         |
| 安装更新           | 下载源保持 Fourgetu，稳定更新脚本按发布标签固定，保留数字修订号比较。                                                |
| Windows 发布       | 采用官方 `3x-ui.exe` 文件名、图标和版本资源；验证恢复工作流匹配新文件名，并增加 Windows 启动版本及文件版本资源检查。 |
| GOST 打包          | 延续 Linux 七架构和 Windows amd64 的 GOST SHA-256 校验、文件名映射和打包。                                           |

限速协议范围不变：本机 VLESS/VMess/Trojan 的 TCP/raw、WS、gRPC、HTTPUpgrade、XHTTP，以及 Shadowsocks TCP/UDP。远程入站在实际承载面板配置；MASQUE 和 XDRIVE 尚不支持此限速。

## 本地验证

- Go 全量测试通过：`go test -shuffle=on -count=1 ./internal/... ./tools/... .`，包含新增的全量同步/增量解绑限速兼容回归。
- 前端代码生成、TypeScript、Lint、格式检查和生产构建通过。
- 生产依赖审计：0 vulnerabilities。
- Windows CGO 编译和启动版本通过，输出 `3.10.2-fourgetu.1`；FileVersion、ProductVersion 与版本一致，OriginalFilename 为 `3x-ui.exe`。本机链接器报告多 manifest 合并警告，但构建退出码为 0，程序可启动且上述资源读取正确；云端 Windows 构建另行验证。
- 安装、更新、管理菜单、Docker 初始化脚本 Bash 语法检查通过。
- 前端全量测试、Go Lint 和云端检查结果在完成后补充。

## 发布与升级

本记录初始提交时尚未提升稳定版。预发布安装包完成八平台构建、校验和检查与 CI 后，再提升为 Latest 并同步中文安装器。原有 `3x-ui-cn-installer/` 和 `release-download-v2/` 未跟踪目录不纳入主仓库提交。

Windows 用户需将服务、计划任务、快捷方式或脚本中的 `x-ui.exe` 改为 `3x-ui.exe`。Linux 启动命令不变。节点应随面板升级，才能使用此次节点和 Shadowsocks 热更新修复。

沿用官方 systemd 沙箱：自定义 GOST 配置若位于数据库目录之外，需配置本机 `20-local.conf` 的 `ReadWritePaths=`。没有更新任何生产服务器。
