#!/usr/bin/env bash
# Build binary LazyRun, lalu bungkus jadi AppImage + zsync update info.
set -euo pipefail
cd "$(dirname "$0")"

VERSION="${VERSION:-1.0.0}"
OUT="${OUT:-dist}"
ARCH="$(uname -m)"
TOOLS="$OUT/tools"

mkdir -p "$OUT" "$TOOLS"

echo "==> build binary"
CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$OUT/lazyrun" ./cmd/lazyrun

if [[ "${SKIP_APPIMAGE:-0}" == "1" ]]; then
	echo "==> SKIP_APPIMAGE=1, selesai"
	exit 0
fi

echo "==> siapkan AppDir"
APPDIR="$OUT/LazyRun.AppDir"
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin"
cp "$OUT/lazyrun" "$APPDIR/usr/bin/lazyrun"
cp assets/lazyrun.desktop "$APPDIR/lazyrun.desktop"
cp assets/icon.png "$APPDIR/lazyrun.png"
ln -sf usr/bin/lazyrun "$APPDIR/AppRun"

if [[ ! -x "$TOOLS/linuxdeploy" ]]; then
	echo "==> download linuxdeploy"
	curl -fL --retry 3 -o "$TOOLS/linuxdeploy" \
		"https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-$ARCH.AppImage"
	chmod +x "$TOOLS/linuxdeploy"
fi

# linuxdeploy butuh info ini untuk menamai output dan menyisipkan zsync update info.
export ARCH
export LINUXDEPLOY_OUTPUT_APP_NAME=lazyrun
export LINUXDEPLOY_OUTPUT_VERSION="$VERSION"
export LDAI_OUTPUT="$PWD/$OUT/lazyrun-$VERSION-$ARCH.AppImage"
if [[ -n "${GH_USER:-}" && -n "${GH_REPO:-}" ]]; then
	export LDAI_UPDATE_INFORMATION="gh-releases-zsync|$GH_USER|$GH_REPO|latest|lazyrun-*.AppImage.zsync"
else
	echo "==> catatan: GH_USER/GH_REPO kosong, AppImage tanpa update info zsync"
	export LDAI_NO_APPSTREAM=1
fi

echo "==> bundle + package AppImage"
# hapus output lama dulu: file yang sedang dijalankan tidak bisa ditimpa langsung
rm -f "$OUT"/lazyrun-*.AppImage "$OUT"/lazyrun-*.AppImage.zsync
"$TOOLS/linuxdeploy" --appimage-extract-and-run \
	--appdir "$APPDIR" \
	--desktop-file "$APPDIR/lazyrun.desktop" \
	--icon-file "$APPDIR/lazyrun.png" \
	--output appimage

ls -lh "$OUT"/lazyrun-*.AppImage* 2>/dev/null || true
mv -f "lazyrun-$VERSION-$ARCH.AppImage.zsync" "$OUT/" 2>/dev/null || true
