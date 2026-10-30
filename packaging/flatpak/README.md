# Flatpak 打包

用 flatpak-builder 从本仓库源码离线构建「MyGo 云音乐」的 Flatpak 包。

## 构建

需要 flatpak、flatpak-builder，以及 freedesktop 25.08 运行时（flatpak-builder 会自动安装）：

```bash
cd packaging/flatpak
go mod vendor   # 依赖以 vendor 目录进入沙箱，构建时不联网
flatpak-builder --user --install-deps-from=flathub \
  --force-clean --repo=repo --install build dev.egoist.cloudmusic.yaml

flatpak run dev.egoist.cloudmusic
```

构建沙箱不通网络，Go 工具链按 go.mod 的版本钉在 manifest 里（两个架构的官方 tarball，校验和与 Flathub 的 org.freedesktop.Sdk.Extension.golang 一致）。

## 更新版本时的维护点

- `build.sh`：ldflags 里的 `packageVersion`
- `dev.egoist.cloudmusic.metainfo.xml`：顶端加 `<release>`
- `dev.egoist.cloudmusic.yaml`：go.mod 版本变化时换工具链 tarball 及其 sha256

## 已知局限

- **托盘图标显示为字母而非 app 图标**。MyGo 的 Linux 托盘走 libayatana-appindicator，DBus 上只发布图标名加一个 IconThemePath 路径，该路径指向沙箱内私有 /tmp，宿主面板读不到，于是 Plasma 退回首字母。托盘的菜单和点击不受影响。appindicator 库的接口没有发布像素的途径，根除需要 MyGo 上游直接实现 StatusNotifierItem（IconPixmap 内联）；在此之前关窗即退出（无 Dock 可回）。
- **应用内自动更新不可用**：/app 只读，mygo 的更新器探测到后自动禁用，升级靠重装新版 Flatpak。
