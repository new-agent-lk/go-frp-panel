# GLOSSARY

本仓库的领域词条。写代码、写文档、开 issue 都用这里的叫法。

## 程序

**面板（panel）**
`cmd/app` 编出的程序，等于 frps + 管理后台（默认 7200）+ 客户端分发。不含 frpc 运行时。入口 `cmd/app/main.go` → `cmd.Execute` → `pkg/frp/frps` → `internal/frps`。

**acfrpc**
`cmd/frpc` 编出的客户端程序。frpc 在它进程内运行（`internal/frpc/client.go:94` 的 `client.NewService()`），自带 web 管理端（默认 6400）与自更新接口。终端用户下载运行的就是它。

**acfrps**
`cmd/frps` 编出的独立服务端程序，与面板里的 frps 同源。

## 分发

**分发层**
面板把 acfrpc 二进制和 `frpc.toml` 交给终端用户的那一层。涉及 `internal/frps/user_api.go` 的 `apiClientGet`、`internal/frps/client_gen.go` 的 `apiCreateFrpcToml` 与 `apiCreateFrpcByUpload`。

**客户端清单（manifest）**
release 里随包发布的 `clients/index.json`，描述"有哪些 acfrpc 版本可用"：版本号、平台、文件名、sha256、`minPanelVersion`。只描述有哪些，不描述用哪个。

**激活版本**
面板当前向用户下发的 acfrpc 版本。存在面板自身配置里，可切换、可回滚。与 manifest 分离。

**客户端树**
`apiClientGet` 返回给前端的 `[]Option` 结构（`pkg/utils/nodeutils.go:11`）。分组键是 `<名>_<版本>`，叶子是 `<os>_<arch>`。

## 版本

**frp 版本**
编译期依赖的 `github.com/fatedier/frp` 版本（`go.mod`），决定配置模型与 API 形态。当前 `v0.72.0`（2026-11-23 由 v0.70.0 升级，见 `docs/adr/0002-frp-upgrade-v072.md`）。它在 `/api/version` 里以 `frpcVersion` 字段暴露（`pkg/utils/utils.go:309` 的 `version.Full()`），是编译期常量，运行时不可变——升级 frp 必须重编二进制。

**跨版本互通**
不同 frp 版本的服务端与客户端能否互连。本仓口径：v0.70 与 v0.72 双向全通，frp 源码无版本闸门，兼容底线在 v0.50.0（runID 强制），wire protocol 默认值两版都是 `"v1"`。因此面板与 acfrpc 不要求同时升级；服务端日志的 `client login info: ... version [x.y.z]` 是观测在跑旧版客户端的现成手段。

**破坏面**
一次依赖升级可能打破的调用点集合。分两类：编译期（签名、字段、包路径变化，编译器会报）与非编译期（反射、配置 schema、线协议、行为语义，编译器不报）。本仓升级 frp 时的非编译期风险点是 `pkg/utils/reflectutils.go:31` 的 `GetPointerInstance[T]`——它用 `FieldByName("webServer")` 反射取 frp 的私有字段，字段改名不会导致编译失败，只会让运行时拿到 nil。

**acfrpc 版本**
分发出去的那个二进制的版本。与面板版本、frp 版本三者独立。

**面板版本**
面板二进制自身的版本，决定它认识哪些配置字段、支持哪些 manifest 字段。
