<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/3x-ui-dark.png">
    <img alt="3x-ui" src="./media/3x-ui-light.png">
  </picture>
</p>

<p align="center">
  <a href="https://github.com/Fourgetu/3x-ui/releases"><img src="https://img.shields.io/github/v/release/Fourgetu/3x-ui" alt="版本"></a>
  <a href="https://github.com/Fourgetu/3x-ui/actions"><img src="https://img.shields.io/github/actions/workflow/status/Fourgetu/3x-ui/release.yml.svg" alt="构建"></a>
  <a href="https://github.com/Fourgetu/3x-ui/releases/latest"><img src="https://img.shields.io/github/downloads/Fourgetu/3x-ui/total.svg" alt="下载量"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0.en.html"><img src="https://img.shields.io/badge/license-GPL%20V3-blue.svg?longCache=true" alt="许可证"></a>
  <a href="https://pkg.go.dev/github.com/mhsanaei/3x-ui/v3"><img src="https://pkg.go.dev/badge/github.com/mhsanaei/3x-ui/v3.svg" alt="Go Reference"></a>
  <a href="https://docs.sanaei.dev"><img src="https://img.shields.io/badge/docs-docs.sanaei.dev-22d3ee" alt="Documentation"></a>
</p>

# Fourgetu 3X-UI

这是基于 [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) 持续同步的中文增强版 3X-UI，用于管理 [Xray-core](https://github.com/XTLS/Xray-core) 服务器。项目保留官方核心功能，并加入了推荐协议模板、中转配置、多服务器部署、节点识别和二维码等实用功能。

> [!IMPORTANT]
> 本项目仅供个人学习和使用。请遵守当地法律法规，不要将其用于违法用途或未经授权的网络访问。

## 本分支新增功能

### 一键推荐协议模板

在“添加入站”窗口中可以直接选择推荐协议模板，自动填充一套完整配置：

- **VLESS + Reality（Vision）**：免证书，适合作为主力节点。
- **VLESS + Reality（gRPC）**：Reality 的 gRPC 传输方案，免证书。
- **Trojan + TLS**：经典 TLS 伪装，需要域名和证书。
- **VMess + WebSocket + TLS**：支持 CDN 中转，需要域名和证书。
- **Hysteria2**：基于 QUIC，速度快，需要域名和证书，需使用 sing-box、NekoBox 等支持 Hysteria2 的客户端。

支持“一键添加全部推荐协议”：已配置证书时创建全部 5 个模板；未配置证书时只创建 2 个免证书 Reality 模板。多个模板可以共用同一个订阅 ID，客户端导入一次即可获取全部节点，并自动读取面板已有的证书和 Reality 密钥。

关闭“推荐协议”开关后，仍可使用完整的手动配置模式，自行设置协议、传输、安全、嗅探和高级参数。

### 中转与落地分流

新增“添加中转”向导，用于配置“入口服务器 → 落地服务器”的流量转发：

- 落地端可以直接粘贴分享链接，自动识别 VLESS、VMess、Trojan、Shadowsocks、SOCKS 和 HTTP。
- 自动创建入口入站、落地出站和路由规则，默认使用免证书 Reality 作为入口。
- 支持创建完成后的连通性测试，查看入口到落地端的延迟。
- 入站列表会显示“中转”来源标识。

### 多服务器部署

在“服务器”页面注册远程服务器后，添加入站时可以选择“部署到”目标服务器。推荐协议模板和中转向导都支持远程部署；离线服务器仍会显示，但会被标记为不可选择，可以从一个面板统一管理多台 VPS。

### 入站和客户端管理增强

- 入站支持勾选后批量删除，并可同时清理不再归属任何入站的孤儿客户端。
- 客户端列表显示来源：中转、普通入站或独立客户端，并支持按来源筛选。
- 每个入站支持直接打开二维码；同一入站下的多个客户端会分别生成二维码。
- 二维码标题包含入站备注和客户端名称，方便手机端扫码导入。
- 节点名称支持显示服务器、入站备注、客户端名称和端口。
- 同一客户端关联多个端口时，节点名称会自动增加序号，避免节点名称重复。

### 中文化与版本同步

- 提供中文安装脚本 `install-cn.sh` 和中文管理菜单 `x-ui-cn.sh`。
- 自动同步本仓库最新发布版本的安装包和官方脚本。
- 官方功能与本分支功能发生冲突时，优先保留官方实现。

## 官方基础功能

- **多协议入站**：VLESS、VMess、Trojan、Shadowsocks、WireGuard、Hysteria2、HTTP、SOCKS、Dokodemo-door、Tunnel 和 TUN。
- **现代传输与安全**：TCP、mKCP、WebSocket、gRPC、HTTPUpgrade 和 XHTTP，支持 TLS、XTLS 与 REALITY。
- **回落（Fallback）**：通过 Xray fallback 在同一端口提供多种协议。
- **按客户端管理**：流量配额、到期时间、IP 限制、在线状态、分享链接、二维码和订阅。
- **流量统计**：按入站、客户端和出站统计流量，并支持重置。
- **出站与路由**：WARP、NordVPN、自定义路由规则、负载均衡和出站代理链。
- **内置订阅服务器**：支持多种订阅格式和自定义订阅页面模板。
- **Telegram 机器人**：用于远程监控和管理。
- **RESTful API**：提供面板内置 Swagger 文档。
- **数据库支持**：SQLite（默认）和 PostgreSQL。
- **Fail2ban 集成**：用于执行按客户端的 IP 限制。

## 快速开始

### 中文版安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/Fourgetu/3x-ui-cn-installer/main/install-cn.sh)
```

安装完成后运行以下命令打开中文管理菜单：

```bash
x-ui
```

安装指定版本时，将版本标签作为参数传入中文安装脚本，例如：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/Fourgetu/3x-ui-cn-installer/main/install-cn.sh) v3.8.5-fourgetu.1
```

### 更新和查看版本

```bash
x-ui update
x-ui version
```

也可以重新执行中文安装命令进行覆盖安装。覆盖安装前建议先备份 `/etc/x-ui` 和 `/usr/local/x-ui`。

Every release asset is published with a `.sha256` sum next to it. Both `install.sh` and the updater verify the archive against that sum and abort on a mismatch.

For full documentation — installation, configuration, operations, and the complete API reference — visit **[docs.sanaei.dev](https://docs.sanaei.dev)**.

当前同步基线为官方 `v3.8.5`；本分支发布版本会使用 `v3.8.5-fourgetu.*` 标签。

## 支持的平台

**操作系统：** Ubuntu、Debian、Armbian、Fedora、CentOS、RHEL、AlmaLinux、Rocky Linux、Oracle Linux、Amazon Linux、Virtuozzo、Arch、Manjaro、Parch、openSUSE、Alpine 和 Windows。

**架构：** `amd64` · `386` · `arm64` · `armv7` · `armv6` · `armv5` · `s390x`。

## 数据库

3X-UI 支持 SQLite 和 PostgreSQL：

- **SQLite（默认）**：数据库文件位于 `/etc/x-ui/x-ui.db`，无需额外配置。
- **PostgreSQL**：适合客户端数量较多或多服务器部署的场景。

运行时可以通过环境变量选择数据库：

```bash
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

## 截图

<details>
<summary>点击展开</summary>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/01-overview-dark.png">
  <img alt="概览" src="./media/01-overview-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/02-add-inbound-dark.png">
  <img alt="添加入站" src="./media/02-add-inbound-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/03-add-client-dark.png">
  <img alt="添加客户端" src="./media/03-add-client-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/05-add-nodes-dark.png">
  <img alt="服务器节点" src="./media/05-add-nodes-light.png">
</picture>

</details>

## 贡献

欢迎提交 Issue 和 Pull Request。提交前请阅读[贡献指南](CONTRIBUTING.md)。

## 环境变量

| Variable | Description | Default |
| --- | --- | --- |
| `XUI_DB_TYPE` | Database backend: `sqlite` or `postgres` | `sqlite` |
| `XUI_DB_DSN` | PostgreSQL connection string (when `XUI_DB_TYPE=postgres`) | — |
| `XUI_DB_FOLDER` | Directory for the SQLite database file | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | Maximum open connections (PostgreSQL pool) | — |
| `XUI_DB_MAX_IDLE_CONNS` | Maximum idle connections (PostgreSQL pool) | — |
| `XUI_INIT_WEB_BASE_PATH` | The initial URI path for the web panel | `/` |
| `XUI_ENABLE_FAIL2BAN` | Enable Fail2ban-based IP-limit enforcement | `true` |
| `XUI_LOG_LEVEL` | Log verbosity (`debug`, `info`, `warning`, `error`) | `info` |
| `XUI_DEBUG` | Enable debug mode | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | Enable the tunnel health monitor (probes a URL and restarts xray after repeated failures; a restart drops all clients) | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | Proxy the probe is sent through; point it at a local xray inbound so the probe tests the tunnel (e.g. `socks5://127.0.0.1:1080`). Empty means the probe only checks host connectivity | — |
| `XUI_TUNNEL_HEALTH_URL` | URL probed for tunnel health | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | Interval between probes | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | Per-probe timeout | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | Consecutive failures before a restart is triggered | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | Minimum delay between consecutive restarts | `5m` |
| `NODE_TOKEN_ENCRYPTION` | Encryption at rest for node API tokens: `off`, `migration`, or `required` (note: no `XUI_` prefix) | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | JSON keyring (mode `0600`) holding the active key id and its base64 32-byte keys | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | A single base64 32-byte key, used only when the key file cannot be loaded | — |

The complete list is on the [environment variables reference](https://docs.sanaei.dev/docs/reference/env-vars).

## 许可证

本项目基于 GPL v3 开源协议发布。

## 致谢

- [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)
- [alireza0](https://github.com/alireza0/)

感谢 [Linux.do](https://linux.do/) 社区。

## Acknowledgment

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (License: **GPL-3.0**): _Enhanced v2ray/xray and v2ray/xray-clients routing rules with built-in Iranian domains and a focus on security and adblocking._
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (License: **GPL-3.0**): _This repository contains automatically updated V2Ray routing rules based on data on blocked domains and addresses in Russia._

## Community Tools

Tools and integrations built by the community around 3x-ui.

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (License: **MIT**): _Manage inbounds, clients, panel settings, and Xray configuration as code with Terraform / OpenTofu._
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (License: **MIT**): _Native Android client for 3x-ui — dashboard, inbounds, clients with QR sharing, nodes and multi-panel management. Available on F-Droid._

## Support project

**If this project is helpful to you, you may wish to give it a**:star2:

<a href="https://www.buymeacoffee.com/MHSanaei" target="_blank">
<img src="./media/default-yellow.png" alt="Buy Me A Coffee" style="height: 70px !important;width: 277px !important;" >
</a>

</br>
<a href="https://nowpayments.io/donation/hsanaei" target="_blank" rel="noreferrer noopener">
   <img src="./media/donation-button-black.svg" alt="Crypto donation button by NOWPayments">
</a>

## Star History

<a href="https://www.star-history.com/?repos=mhsanaei%2F3x-ui&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=mhsanaei/3x-ui&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=mhsanaei/3x-ui&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=mhsanaei/3x-ui&type=date&legend=top-left" />
 </picture>
</a>

<p align="center">
 <a href="https://www.star-history.com/mhsanaei/3x-ui">
  <picture><source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=rank&theme=dark" /><source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=rank" /><img alt="Star History Rank" src="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=rank" /></picture> <picture><source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=trending&theme=dark" /><source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=trending" /><img alt="GitHub Trending Repository of the Day" src="https://api.star-history.com/badge?repo=MHSanaei/3x-ui&type=trending" /></picture>
 </a>
</p>
