## v0.5.21

### 修复
- 客户端清单 `index.json` 混进了 `differ` 生成的 `.patch` 增量包，被面板当成可选客户端列出。构建与读取两侧都排除 `.patch`。
- 发版保留上一版完整二进制：`differ` 会就地把旧文件改名成新版本号，旧版本随之从 release 消失，面板只剩一版可选。现在先留一份上一版的 `acfrpc` 完整二进制，清单累积最近两个版本，回滚可用。

### 升级
- `github.com/fatedier/frp` v0.70.0 → **v0.72.0**（`/api/version` 的 `frpcVersion` 显示 `0.72.0`）
- Go 指令 1.25.0 → 1.26.0

### 源码改动（仅一处）
`internal/frpc/frpc.go` 的 `retry.Do` 加 `retry.RetryIf` 跳过 `context.Canceled`。v0.72 起 `Service.Close()` 会 cancel ctx 并置 `closed`，之后 `Run()` 立刻返回；旧的重试逻辑会每 5s 刷一条 `mainfrpc 客户端连接失败[...]: context canceled`。顺带修掉 v0.70 下「已删除的客户端被重新登录复活」的隐患。

### 兼容性
- **存量配置直接可用**：`pkg/config/v1` 的 TOML 键只新增 `metricsServer`、`useSystemCa`，零删除零改名。本仓配置加载走 `strict=true`，strict 只拒未知键，旧 `frpc.toml` / `frps.toml` 照读。
- **跨版本全通，不要求同时升级**：0.70×0.70、0.72×0.72、0.72 服务端×0.70 客户端、0.70 服务端×0.72 客户端四组真实二进制矩阵均通过登录、建代理与数据面回包。frp 无版本闸门，兼容底线 v0.50.0（runID 强制），wire protocol 两版默认都是 `v1`。
- 服务端日志的 `client login info: ... version [x.y.z]` 可用于观测仍在跑旧版的客户端。

### 客户端版本清单
`index.json` 现在累积多个 acfrpc 版本（v0.5.17 / v0.5.18 / v0.5.19），面板可列出并切换激活版本、可回滚上一版。

### 构建
- 16 平台全量 `acfrpc` + `acfrps`
- CI 发版改用自动可用的 `GITHUB_TOKEN` 并加 `contents: write`
- 修掉 CI 里 `PREV_VERSION` 未定义导致 `differ` 从未执行的问题，以及重复的第二次全量编译

### 已知噪音
启动会多一行 deprecation 提示：`webServer /healthz is deprecated; health probes can migrate to metricsServer.port`。不影响功能，上游称 v0.80.0+ 可能移除。

详见 `docs/adr/0002-frp-upgrade-v072.md`。

---

### ✨ 新特性

* 程序以服务形式安装并运行，支持跨平台windows、linux、macos平台；
* 新增重启功能，用户可管理后台操作重启；
* 新增在线升级功能，可上传式升级和文件url式升级；
* 新增可在管理后台端查看日志功能；
* frps服务端可生成frpc客户端，密钥信息二进制内嵌在客户端程序中；
* 新增用户配置，可以配置授权用户供frpc端使用
* frpc客户端可运行多客户端
* 新增frpc用户配置导入导出
* 客户端版本与面板版本脱钩：面板可列出多个 acfrpc 版本并切换激活版本

### ⚙️ 问题修复

* Properly release resources in service.Close() to prevent resource leaks when used as a library.
---
### 🚀 github加速

```
[
  "https://ghfast.top/",
  "https://gh-proxy.com/",
  "https://ghproxy.1888866.xyz/",
  "https://github.moeyy.xyz/"
]
```
