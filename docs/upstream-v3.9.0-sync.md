# Fourgetu v3.9.0 同步记录

完成日期：2026-10-09（Asia/Shanghai）。源码候选版本：`3.9.0-fourgetu.1`。

## 基线与范围

- 从已发布的 `origin/sync/upstream-v3.8.5`（`53203333`）合并官方 `v3.9.0`（提交 `3cd4bf50`）。该基线包含当前旧功能分支和 `origin/main` 的全部提交，以及上一次发布的后端编译修复。
- 工作分支：`codex/sync-upstream-v3.9.0`。保留真实的 Git 合并历史，可继续三方同步。
- 本次只完成本地代码、验证和提交；未推送分支、移动标签、创建 Release 或更新服务器。
- 当前目录原有的 `3x-ui-cn-installer/` 和 `release-download-v2/` 未纳入本次提交。外部中文安装器依然安装远程已发布版本，本地候选代码尚不可通过远程安装命令获取。

## 官方优先与定制兼容

| 功能 | 处理 |
| --- | --- |
| 用户级独立上传/下载限速、GOST、运行恢复 | 官方没有等价实现；保留数据库模型、API、表单、订阅端口映射及运行时清理。沿用原有协议范围，远程入站在实际承载面板设置。 |
| 推荐协议模板、共享订阅 ID、证书和 Reality 密钥 | 保留；继续通过官方表单 schema、客户端创建和证书校验。 |
| 中转/落地向导、复用已有入口 | 保留向导及来源标识，使用现有官方出站和路由配置。 |
| 多服务器和协议部署 | 使用官方节点能力和版本限制；保留模板/中转的部署入口。官方原生 TUIC、AmneziaWG、MTProto 节点支持完整合入。 |
| 批量删除、可选孤儿客户端清理 | 官方 `DelInbounds(ids)` 及其并发节点删除逻辑保持原签名；定制清理通过 `DelInboundsPurgeOrphans` 包装调用官方路径。共享客户端不被误删。 |
| 客户端来源过滤、二维码、节点名称/端口、重名序号 | 保留定制展示，与官方 Hosts、隧道配置和订阅控制共同工作。 |
| 周续费、并发客户端修改、旧入站表单保存 | 完整使用官方生命周期逻辑；分页同时返回 `resetWeekday` 和 `speedLimits`。编辑入站时不再提供官方已禁止的 enable 切换。 |
| Windows oxlint 测试入口 | 官方已有等价修复，整个测试文件采用官方版本，删除重复的自定义分支。 |
| 服务停止 | 使用官方 HTTP/任务停止、TUIC 最终流量落盘顺序，再接入 GOST 停止。 |
| 安装与升级 | 安装包、菜单和更新源保持 Fourgetu；修复遗漏的网页更新源。稳定版网页/命令行更新均从选定的发布标签读取更新脚本，防止旧 main 脚本把面板替换为官方版。 |
| 版本识别 | 二进制标记为 `3.9.0-fourgetu.1`；前后端按上游三段版本和 Fourgetu 数字修订号排序，节点协议版本检查也识别该后缀。 |

官方新增的原生 TUIC、Xray v26.9.30、周续费、订阅排除、流量导入导出、Telegram 权限和安全修复均随合并保留。Go 依赖、前端依赖、构建工具要求使用官方 v3.9.0 基线。

## 合并中的额外修复

1. 自动生成器登记 `ClientSpeedLimit`、`ClientSpeedLimitView`、`ClientSpeedLimitUpdate`，并提供有效的正整数入站示例。重新生成 TypeScript、Zod、JSON Schema、示例和 OpenAPI，避免手改生成文件后被下一次生成覆盖。
2. 入站删除事务失败时不下发节点删除；限速重同步放在官方本地运行时与路由清理之后，防止限速错误跳过官方清理。
3. 两项数据库回归验证：旧入站表单不覆盖周续费与限速；批量删除和孤儿清理保留仍在其他入站使用的客户端及限速。
4. 升级 URL、非法标签拒绝和数字修订比较均有回归覆盖；命令行更新下载失败会返回失败，不会执行空脚本并宣称成功。
5. 对上游锁文件做兼容的间接依赖补丁更新：`solid-js` 1.9.17、`seroval` / `seroval-plugins` 1.6.8，消除审计中的 critical/high 链。没有更改直接依赖范围或强制降级 Swagger。补丁更新后再次通过生产构建及 24 项组件/数据查询回归。

## 验证

环境：Windows amd64，独立缓存工具链 Go 1.27.1、Node 26.11.1、npm 11.20.0、WinLibs GCC 16.2.0，`CGO_ENABLED=1`。工具链未写入仓库或修改系统安装。

- `go test -shuffle=on -count=1 ./internal/... ./tools/... .`：全量通过。包括 SQLite 数据库迁移、限速、订阅、核心业务、控制器、原生 TUIC 和 AmneziaWG。
- 首轮 Discord 心跳测试曾在并行负载下连接被关闭；按同一 shuffle seed 复跑及随后全量复跑均通过，未改动该官方测试或实现。
- 版本标记及比较修改后，重新运行 config、version、panel、service 包并通过。
- 最终前端完整回归：182 个文件、1,835 项测试通过，包含单元/组件及 28 个文件、95 项 Storybook Chromium 浏览器测试。
- `npm run gen`、`npm run typecheck`、`npm run lint`、`npm run format:check`、`npm run build`、`npm run build-storybook` 通过。
- `npm audit --omit=dev --audit-level=high` 通过；仍有 `swagger-ui-react → remarkable → argparse → sprintf-js` 链的 4 项 moderate 报告。当前 sprintf-js 无修补发行版，审计建议的 Swagger 3.x 降级属于破坏性改动，未采用。
- `go build` 成功；二进制 `-v` 输出 `3.9.0-fourgetu.1`。
- `install.sh`、`update.sh`、`x-ui.sh`、`DockerInit.sh` 的 Bash 语法检查通过。模拟升级测试验证 Fourgetu 标签固定、非法标签拒绝及下载失败行为。
- 格式检查最初主要受 Windows `core.autocrlf=true` 的工作区换行影响；统一相关工作区文本为 LF，并格式化 4 个历史定制文件后，全量格式检查通过。没有大范围改写官方代码。`git diff --check` 和冲突标记检查通过。

前端在本机使用 `npx vitest run --pool=threads`；默认子进程模式曾长时间没有进展。PostgreSQL 专用用例需要配置数据库，当前未执行；也未对生产服务器执行安装、升级或真实流量限速压测。

## 发布前

候选版本未发布，README 的固定安装示例继续指向已有的 `v3.8.5-fourgetu.1`。后续发布应从本分支创建新标签、运行 Linux/Windows 发布构建，再核对中文安装器版本和其管理脚本更新入口；不可仅改 README 就宣称远程可安装。
