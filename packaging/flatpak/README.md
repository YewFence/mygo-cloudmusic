# Flatpak 打包

用 flatpak-builder 从本仓库源码离线构建「MyGo 云音乐」的 Flatpak 包。

## 构建

需要 flatpak、flatpak-builder，以及 freedesktop 25.08 运行时（flatpak-builder 会自动安装）：

```bash
go mod vendor                        # 依赖以 vendor 目录进入沙箱，构建时不联网
packaging/flatpak/prepare.sh         # 生成 _build/（版本元数据，gitignored）
cd packaging/flatpak
flatpak-builder --user --install-deps-from=flathub \
  --force-clean --repo=repo --install build dev.egoist.cloudmusic.yaml

flatpak run dev.egoist.cloudmusic
```

构建沙箱不通网络，Go 工具链按 go.mod 的版本钉在 manifest 里（两个架构的官方 tarball，校验和与 Flathub 的 org.freedesktop.Sdk.Extension.golang 一致）。

## 版本元数据是构建期生成的

`prepare.sh`（内部是 `gen/`）从仓库自己的数据推导打包所需的一切版本信息，写进 gitignored 的 `_build/`，manifest 从那里构建：

- **二进制的 ldflags**（应用名、版本号）← `mygo.json`
- **metainfo 的 `<releases>`**（每个版本的条目和日期）← `CHANGELOG.md` 的 `## x.y.z` 段落，日期取自对应 tag

所以发版不需要为打包多做任何事：写 CHANGELOG、bump `mygo.json`、打 tag，CI 的 Flatpak job 自己跑 prepare。仓库里没有派生内容，也就不存在忘记同步或漂移的问题。

仓库里的 `build.sh` 和 metainfo 是模板（`@NAME@`/`@VERSION@` 占位符、空的 `<releases>` 块），不要手工填值。

## 手工维护点（生成器推不出来的）

- go.mod 的 Go 版本变化时：换 manifest 里两个工具链 tarball 的 URL 和 sha256（与 Flathub 的 org.freedesktop.Sdk.Extension.golang 对照）
- metainfo 顶部的描述性文案（应用简介、功能列表）有变时手改模板

## Release assets

`.github/workflows/release.yml` 的 Flatpak job（x86_64 和 arm64 各一）在打 tag 时构建对应架构的 `.flatpak` 单文件捆绑包，连同 SHA256SUMS 上传为 release 资产。用户 `flatpak install --bundle` 即可离线安装。

## 已知局限

- **托盘图标显示为字母而非 app 图标**。MyGo 的 Linux 托盘走 libayatana-appindicator，DBus 上只发布图标名加一个 IconThemePath 路径，该路径指向沙箱内私有 /tmp，宿主面板读不到，于是 Plasma 退回首字母。托盘的菜单和点击不受影响。appindicator 库的接口没有发布像素的途径，根除需要 MyGo 上游直接实现 StatusNotifierItem（IconPixmap 内联）；在此之前关窗即退出（无 Dock 可回）。
- **应用内自动更新不可用**：/app 只读，mygo 的更新器探测到后自动禁用，升级靠重装新版 Flatpak。
