# 面板去广告定制

版本：`v3.11.0-fourgetu.3`。日期：2026-10-31（Pacific/Honolulu）。

后续同步上游时保留以下策略：

- 登录页、系统状态页不加载赞助广告。
- 移除赞助商页面、侧栏菜单及快捷搜索入口。
- 桌面侧栏和移动端菜单不显示捐赠入口。
- 移除前端赞助组件、查询、样式和相关翻译。
- 保留旧客户端兼容接口：`GET /sponsors` 返回成功及空数组，不返回招商联系方式；`GET /sponsors/logo/:name` 返回 404。两者均禁止缓存，不调用远程广告服务。

检查范围包括面板各页面、订阅页面、前端固定外链及后端广告服务调用。文档、项目源码、客户端下载链接，以及管理员自定义的订阅联系方式、续费地址和 MTProto 频道配置属于功能配置，继续保留。

上游赞助服务实现和类型保留用于后续合并及 API 类型兼容，但运行时控制器没有调用入口。此改动不修改数据库或现有客户端配置。

安装已包含本改动的版本并刷新页面后生效。旧版本发布包和正在运行的服务器不会随源码修改而自动更新。

## 本地验证

- TypeScript、Lint、格式检查和生产构建通过。
- 导航、侧栏、快捷搜索及多语言检查共 21 项通过。
- 后端 controller、panel 服务测试通过；新增根路径和自定义路径下的广告兼容接口测试。
- 生产资源中没有广告组件、广告样式和捐赠地址；API 文档示例与空广告响应一致。

## 发布依赖修补

候选版 `v3.11.0-fourgetu.2` 在云端被依赖审计拦截，保留为不可安装的预发布版，不移动旧标签。正式发布候选改为 `v3.11.0-fourgetu.3`，将间接依赖 `follow-redirects` 从 1.16.1 更新到 1.16.2，修复 GHSA-mrwp-f3q8-3g9p。仅更新锁文件中该依赖，没有降低 Swagger UI 主版本。

生产依赖的高危审计门槛已通过；仍有 Swagger 依赖链的 2 项低危、3 项中危报告，本次没有以破坏性降级处理。

## 正式发布验证

- `v3.11.0-fourgetu.3` 固定于源码提交 `2bbb4eb7a83c651ffd553d570d74ab63ccac9dd0`，已设为正式版 / Latest；后续文档更新不移动此标签。
- [完整 CI](https://github.com/Fourgetu/3x-ui/actions/runs/41166808084) 与 [Docs CI](https://github.com/Fourgetu/3x-ui/actions/runs/41166810975) 通过。前端 219 个测试文件、2,069 项测试通过，后端全量、race、数据库兼容、模糊测试及节点端到端测试通过。
- [八平台构建](https://github.com/Fourgetu/3x-ui/actions/runs/41166816950) 与 [产物验证](https://github.com/Fourgetu/3x-ui/actions/runs/41167284198) 通过，包含 Linux 面板、GOST、Xray 实际启动和 Windows 版本资源检查。
- 8 个安装包及 8 个 SHA-256 文件已上传；下载全部校验和文件，与 GitHub 附件摘要逐一核对一致。
- Docs Deploy 保持 `disabled_manually`，Docs CI 保持启用。Docker 工作流因缺少 Docker Hub 凭据未发布镜像，不影响 GitHub 安装包。
- 未更新生产服务器。

- [中文安装器同步](https://github.com/Fourgetu/3x-ui-cn-installer/actions/runs/41167478907) 成功，提交 `4dde41f32a0899b07cefc33a2be4bb0d32465849`；版本记录为 `v3.11.0-fourgetu.3`。
