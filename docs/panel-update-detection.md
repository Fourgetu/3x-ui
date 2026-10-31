# 面板更新检查修复

版本：`v3.11.0-fourgetu.5`。

## 已确认的问题

运行 `.1` 的面板更新检查接口实际返回 `{"success":false,"msg":"","obj":null}`。后端检查失败时丢弃错误信息，前端保留默认 `updateAvailable: false`，在更新弹窗中误显示“已是最新版”。这不是版本比较结论。旧响应无法进一步区分限流、DNS、超时或代理连接问题。

侧栏底部版本徽标链接硬编码为官方仓库；其链接与更新服务使用的 `Fourgetu/3x-ui` 下载源彼此独立。

## 修复行为

- 桌面和移动端版本徽标指向 `Fourgetu/3x-ui`。
- 更新检查显示加载、成功和失败状态，失败时不会继续显示“已是最新版”或启用安装按钮。
- 更新弹窗提供刷新按钮，打开弹窗会重新检查；系统状态页可见时每五分钟检查，返回窗口时也会检查过期结果。
- GitHub API 请求失败或返回无效稳定版信息时，使用同一仓库的 `releases/latest` 发布页重定向查询版本；不使用 main 分支版本文件，避免识别尚未发布的构建。
- 备用查询使用已有面板出站代理和 TLS 验证，只接受 HTTPS、GitHub、当前仓库及有效稳定版标签，不跟随未知重定向目标。
- 失败响应提供 HTTP 状态码或 DNS、TLS、超时等类别，不向浏览器返回可能包含代理凭据的底层错误。

## 已无法自动检查的旧版本

新代码不会自动写入正在运行的旧面板。如果旧版本检查一直失败，需要先通过服务器 SSH 手动安装一个已正式发布的修复版本，再刷新浏览器。发布前不要安装候选标签；具体稳定版本以 GitHub Latest 为准。

用户配置、数据库和现有定制功能保持兼容。本次未修改生产服务器。

## 正式发布验证

- 正式版 `v3.11.0-fourgetu.5` 已设为 Latest，标签固定于源码提交 `765fe3ebd8fa2e3932f4bfc14959640110282084`。后续文档更新不移动标签。
- [完整 CI](https://github.com/Fourgetu/3x-ui/actions/runs/41169752622) 和 [Docs CI](https://github.com/Fourgetu/3x-ui/actions/runs/41169754752) 通过；包括前端 220 个测试文件、后端全量测试、并发检查、数据库兼容、模糊测试、主节点端到端测试和安全审计门槛。
- [八平台构建](https://github.com/Fourgetu/3x-ui/actions/runs/41170305222) 与 [产物验证](https://github.com/Fourgetu/3x-ui/actions/runs/41170569292) 通过；构建来源、包结构、全部校验和均已验证，Linux amd64 面板、GOST 和 Xray 实际启动成功。
- 已上传 8 个安装包及 8 个 SHA-256 文件，并下载所有校验文件，与 GitHub 附件摘要逐一核对一致。
- [中文安装器同步](https://github.com/Fourgetu/3x-ui-cn-installer/actions/runs/41170661975) 成功，提交 `473267a975ce84a239bacab2e987969b87d7f5e6`，`upstream-version.txt` 已指向 `.5`。
- `.4` 保留为未通过完整检查的历史候选版，请使用 `.5` 正式版。
- Docs Deploy 保持 `disabled_manually`，Docs CI 保持启用。Docker Hub 凭据缺失，本次未发布 Docker 镜像。
- 生产依赖高危审计门槛通过；Swagger 依赖链仍有已知 2 项低危、3 项中危报告，未作破坏性降级。

## 从旧版手动升级

在服务器 root SSH 终端中执行以下命令前，请备份数据库与配置，并确认磁盘空间充足。安装会短暂重启服务；磁盘不足时不要执行升级。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/Fourgetu/3x-ui/v3.11.0-fourgetu.5/update.sh) v3.11.0-fourgetu.5
```

完成后重新打开面板，确认版本为 `v3.11.0-fourgetu.5`。版本徽标应打开 `Fourgetu/3x-ui`，更新检查失败时应显示错误和刷新按钮。
