#!/bin/sh
# Runs in flatpak-builder's build sandbox with the app's source tree
# (vendored deps included) as the working directory, and the pinned Go
# toolchain from go.mod's line unpacked next to it in .go-toolchain
# (GOROOT, set in the manifest). The sandbox blocks the network, so the
# build is offline: -mod=vendor and GOTOOLCHAIN=local.
set -ex

export PATH="$GOROOT/bin:$PATH"

ldflags="-s -w"
ldflags="$ldflags -X 'github.com/egoist/mygo.packageName=MyGo 云音乐'"
ldflags="$ldflags -X github.com/egoist/mygo.packageVersion=0.1.20"
ldflags="$ldflags -X github.com/egoist/mygo.packageIdentifier=dev.egoist.cloudmusic"
ldflags="$ldflags -X github.com/egoist/mygo.production=1"

go build -trimpath -ldflags "$ldflags" -o /app/bin/mygo-cloudmusic .

install -Dm644 dev.egoist.cloudmusic.desktop \
  /app/share/applications/dev.egoist.cloudmusic.desktop
install -Dm644 dev.egoist.cloudmusic.metainfo.xml \
  /app/share/metainfo/dev.egoist.cloudmusic.metainfo.xml

for size in 512 256 128 64 48 32 24 16; do
  install -Dm644 "${size}.png" \
    "/app/share/icons/hicolor/${size}x${size}/apps/dev.egoist.cloudmusic.png"
done
