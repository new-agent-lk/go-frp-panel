# 0002 — frp 库升到 v0.72.0（随 v0.5.18 交付）

- 状态：已接受（2026-11-23）
- 决策范围：依赖升级与兼容口径，不含功能变更

## 背景

`go.mod` 依赖 `github.com/fatedier/frp v0.70.0`。上游 2026-11-04 发布 v0.72.0，中间还有 v0.70.1、v0.71.0、v0.71.1。本仓对 frp 的依赖分布在 16 个包、约 88 处 import，其中 `pkg/config/v1`(24)、`pkg/util/http`(9)、`pkg/config/v1/validation`(6)、`pkg/config`(6)、`client`(6) 是主要面。

升级不是纯版本号改动：本仓直接 import 了 frp 的内部包（`client`、`pkg/config/source`、`pkg/config/v1/validation`、`pkg/util/http`、`pkg/util/log`、`pkg/util/system`），上游任何内部调整都会传导过来。

## 决策

1. **升到 v0.72.0，不追 master、不追 rc。** 以稳定 tag 为准。
2. **Go 指令随之抬到 1.26.0。** v0.72.0 的 `go.mod` 要求 `go 1.26.0`（v0.71.1 起即如此，v0.70.1 与 v0.71.0 还是 1.25.0）。本机 go1.26.10 满足；CI 的 `setup-go: 1.23` 靠 `GOTOOLCHAIN=auto` 自动拉取，同时把 workflow 的 `go-version` 显式抬到 1.26。
3. **源码适配只做一处：`internal/frpc/frpc.go` 的 `retry.Do` 加 `retry.RetryIf`，跳过 `context.Canceled`。** 理由：v0.72 起 `Service.Close()` 会 cancel ctx 并置 `closed`，之后 `Run()` 立刻返回 `context.Canceled`；而 `deleteClient`（`internal/frpc/client.go:150`）会先 `Close()` 再 `GracefulClose`，旧的重试逻辑（`retry.Attempts(0)` = 无限重试）会每 5s 刷一条 `mainfrpc 客户端连接失败[...]: context canceled`。顺带修正了一个 v0.70 下的隐患：旧版会重新登录，把已删除的客户端"复活"。
4. **不写配置迁移代码。** v0.70.0→v0.72.0 的 `pkg/config/v1` TOML 键集合只有两处新增（`metricsServer`、`useSystemCa`），零删除、零改名。本仓 7 处 `LoadClientConfig`/`LoadServerConfig` 调用全部硬编码 `strict=true`，strict 只拒未知键——旧配置里的每个键在 v0.72 里都还存在，因此存量 `frpc.toml`/`frps.toml` 直接可读。
5. **不设跨版本闸门，发版说明不写「必须同时升级」。** 四组真实二进制组合（0.70×0.70、0.72×0.72、0.72 服务端×0.70 客户端、0.70 服务端×0.72 客户端）均通过登录、建代理与数据面回包；frp 源码里没有任何版本校验，兼容底线在 v0.50.0（runID 强制），wire protocol 默认值两版都是 `"v1"`。
6. **面板不按 frps 版本过滤可下发的客户端版本。** 所有组合都通，过滤无收益。真正需要的约束方向是「客户端要求面板至少多新」，由 manifest 的 `minPanelVersion` 承担。

## 后果

- 面板与 acfrpc 的 `/api/version` 里 `frpcVersion` 从 `0.70.0` 变为 `0.72.0`（`pkg/utils/utils.go:309` 的 `version.Full()`，编译期常量）。
- 启动日志会各多一行 deprecation 提示：`webServer /healthz is deprecated; health probes can migrate to metricsServer.port`（frps 侧 `server/service.go:431`、frpc 侧 `client/service.go:288`）。纯噪音，不影响功能。上游称 v0.80.0+ 可能移除 webServer 上的 `/healthz`、`/metrics` 与 pprof，届时需迁移到 `metricsServer`。
- `go.sum` 从 234 行降到 228 行；`github.com/hashicorp/yamux` 被 `github.com/fatedier/yamux v0.3.0` 取代；新增 `go.yaml.in/yaml/v2`、`github.com/munnerz/goautoneg`。`k8s.io/apimachinery` 仍需保留（`cmd/server` → `frp/pkg/config` → `k8s.io/apimachinery/pkg/util/sets`）。
- 反射契约（`pkg/utils/reflectutils.go:31` 的 `GetPointerInstance[T]` 用 `FieldByName("webServer")`）仍成立：v0.72 的 `Service` 新增了若干字段，但 `webServer *httppkg.Server` 字段名与类型未变。这是本仓唯一的非编译期风险点，已由运行冒烟验证（面板自有路由 `/api/client/get`、`/api/version` 均 200）。
- 同一 `Service` 实例 `Close()` 后再 `Run()` 会静默返回。当前无此复用路径；将来若要做「面板内重启 frps」，必须新建实例。

## 未决

- webServer 上的 `/healthz`、`/metrics`、pprof 被标记 deprecated。是否提前迁到 `metricsServer`，等上游 v0.80 的实际动作再定。
- `minPanelVersion` 与 sha256 目前只写入 manifest、不参与校验，尚未与本次升级联动。
