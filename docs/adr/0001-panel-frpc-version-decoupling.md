# 0001 — 面板与 acfrpc 版本解耦（分发层）

- 状态：已接受（2026-11-23）
- 决策范围：只定方案与契约，不含实现

## 背景

面板（`cmd/app`）与 acfrpc（`cmd/frpc`）是同一个仓库编出的两个程序：

```
cmd/app  → cmd.Execute → pkg/frp/frps → internal/frps      ← 面板：frps + web + 客户端分发
cmd/frpc → service.Bootstrap → pkg/frp/frpc → internal/frpc → client.NewService()  ← acfrpc：frpc 运行时
```

全仓引用 `internal/frpc` 的只有 `pkg/frp/frpc/frpc.go`、`cmd/frpc/service/service.go`、`cmd/frpc/test/main.go`。**面板不含 frpc 运行时**，所以「面板内跑 frpc」不是耦合点，脱钩不需要动运行时。

真正的耦合在分发层：

- `internal/frps/user_api.go:172 apiClientGet` 扫描 `<exeDir>/clients`，并合并 `github.Api().GetDownloadUrls(...)` 的结果（过滤 `acfrpc` 前缀、排除 `.patch`），交给 `utils.ToTree` 组织成树。
- `GetDownloadUrls` → `defaultRequest()` → `Request(user, repo)` → `https://api.github.com/repos/<user>/<repo>/releases/latest`（go-service `pkg/github/github.go:222`、`:103`、`:60`）。

结果：面板只能拿到**最新一版** acfrpc。没有版本选择、不能钉版本、不能回滚。用户反馈的「没有更新 frpc 的方式方法」实际是「没有选 frpc 版本的方法」——acfrpc 自身是能自更新的（`internal/frpc/frpc.go:94` 的 `upgrade: comm2.NewCommApi(i)`，路由 `/api/checkversion`、`/api/upgrade`）。

此外，现有树结构本身装不下多版本：`pkg/utils/nodeutils.go:22` 的 `SplitLastTwoByUnderscore` 返回**最后两段**，`ToTree`（`pkg/utils/nodeutils.go:39`）拿它当分组键。`acfrpc_v0.4.15_linux_amd64` 因此被拆成 group=`linux`、leaf=`amd64`——版本信息在这一步就丢了，多版本会撞成同名叶子。

## 决策

1. **版本来源用 manifest，不用 GitHub releases 列表。** release 里随包发布 `clients/index.json`，记录版本、平台、文件名、sha256。面板读这一份文件，不调 `/releases` 列表接口。
   - 理由：不依赖上游 go-service 的分页与限流改造；离线部署可直接放本地；一份文件就能表达完整清单。
   - 兜底顺序：manifest → `<exeDir>/clients` 本地目录 → GitHub latest（现状行为）。
2. **分组键从 OS 改成 `<名>_<版本>`。** 新增取前 N 段的辅助函数（`parts[:len(parts)-2]`），`ToTree` 用它分组，叶子仍是 `<os>_<arch>`。
   - 收益：多版本天然分层，`Option`/`Node` 结构不变，`web/frps` 前端不用改。
3. **激活版本存在面板自身配置里**，不进 manifest。manifest 只描述"有哪些"，不描述"用哪个"。
4. **切换只影响新下发，不追溯已装客户端。** 已发出的 acfrpc 由它自己的 `/api/upgrade` 决定何时升。
5. **兼容性由 manifest 声明。** 每个版本条目带 `minPanelVersion`，面板版本低于它就标红并拒绝激活。
6. **校验用 sha256**，值取自 manifest；下载后不匹配则拒绝入库。

## 后果

- `apiClientGet` 的返回结构不变（`[]Option`），前端零改动。
- 需要新增：`clients/index.json` 的生成（build 阶段）、读取与合并、激活状态的读写、sha256 校验。
- 需要改动：`pkg/utils/nodeutils.go` 的分组键（1 个辅助函数 + `ToTree` 内一处调用）。
- 不再依赖 `/releases/latest` 做版本发现；该调用仅作为兜底保留。
- 发布源已切到 `new-agent-lk/go-frp-panel`（`build.sh` 的 `testModule` 由 git remote 推导 owner/repo，`github_release` 的 `REPO`、`.goreleaser.yml` 的 `owner` 均已改）。

## 未决

- manifest 是否需要签名字段（当前只有 sha256）。
- 面板是否需要在客户端下载页展示「已下发数量」。
