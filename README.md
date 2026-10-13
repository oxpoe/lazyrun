# LazyRun — lazy shortcut build

Desktop app (Go + Fyne) that creates application shortcuts on Linux in a few clicks:
from an executable file, or from a web URL as a PWA. Ships as a single-file
AppImage, no terminal needed.

- **Repo:** <https://github.com/oxpoe/lazyrun>
- **Donate:** <https://donasi.oxpoe.qzz.io/>

- **Executable shortcut** — picks an ELF/AppImage/script, copies it to
  `~/.local/bin/`, installs the icon, writes the `.desktop` entry.
- **Web app shortcut** — opens the URL as an app: `--app=<url>` for
  Chromium-based browsers (Chrome, Chromium, Brave, Edge, Vivaldi, Helium,
  including Flatpak builds), `--new-window` for Firefox.
- **Shortcut location** — checkbox for *Application menu* and/or *Desktop*,
  at least one is required. Desktop-only entries are written with
  `NoDisplay=true` so they do not show up in the menu.
- **Manage** — list installed shortcuts, see where they live, mark entries whose
  `.desktop` file was deleted manually, uninstall (optionally removing the
  copied binary).
- **Auto refresh** — calls `update-desktop-database` and
  `kbuildsycoca6`/`kbuildsycoca5` after install/uninstall.
- **i18n** — UI in English or Indonesian, chosen from `$LANG`.
- **Self update** — AppImage with embedded zsync update information, optional
  `AppImageUpdate` integration via *Check for Updates*.

## Install

Download the AppImage from [Releases](https://github.com/oxpoe/lazyrun/releases), then:

```bash
chmod +x lazyrun-*-x86_64.AppImage
./lazyrun-*-x86_64.AppImage
```

Keep it wherever you like (for example `~/.local/bin/`). Automatic updates need
[AppImageUpdate](https://github.com/AppImageCommunity/AppImageUpdate) in `$PATH`.

## Usage

1. **+ Executable** or **+ Web App**.
2. Fill in the form (name and icon are pre-filled/auto-detected).
3. Pick *Application menu* and/or *Desktop*, then **Install**.

## What it writes

| Purpose | Path |
|---|---|
| Copied binary | `~/.local/bin/<name>` |
| Icon | `~/.local/share/icons/hicolor/512x512/apps/<name>.png` |
| Menu entry (source of truth) | `~/.local/share/applications/<name>.desktop` |
| Desktop copy | `~/Desktop/<name>.desktop` (from `xdg-user-dir DESKTOP`) |
| LazyRun metadata | `~/.config/lazyrun/shortcuts.json` |

Everything stays inside the home directory — no root, no system directories.
Filenames use a slug of the app name (`WhatsApp Web` → `whatsapp-web.desktop`);
the `Name=` field keeps the name you typed.

## Build

Requirements: Go 1.24+, cgo, and X11/OpenGL development headers (Fyne).

```bash
./build.sh                # dist/lazyrun + dist/lazyrun-<version>-<arch>.AppImage
SKIP_APPIMAGE=1 ./build.sh # binary only
go test ./...             # unit + install/uninstall tests
```

`build.sh` downloads `linuxdeploy` into `dist/tools/` on first run and needs
network for the AppImage step. To embed zsync update information, set
`GH_USER` and `GH_REPO`:

```bash
GH_USER=oxpoe GH_REPO=lazyrun ./build.sh
```

## Roadmap

- v1.1 — edit existing shortcuts, detect binaries in `$PATH`, optional symlink
  instead of a Desktop copy, import/export, dark mode.
- v1.2 — Firefox PWA via `--kiosk`, Flatpak/Snap targets, icon preview.

## Disclaimer / Catatan

**EN** — LazyRun is a small hobby project by **oxpoe**, made for fun and to make
daily Linux desktop life a bit simpler: turn a portable binary or a web page
into a proper application shortcut without opening a terminal. It only writes
inside your home directory, but it does copy files and create launchers: check
what you point it at, keep backups of anything irreplaceable, and use it at your
own risk. Provided "as is", without warranty of any kind. LazyRun is not
affiliated with any browser vendor or application it creates shortcuts for.

**ID** — LazyRun adalah proyek iseng **oxpoe**, dibuat untuk senang-senang dan
supaya urusan shortcut di desktop Linux jadi lebih simpel: binary portable atau
halaman web bisa jadi shortcut aplikasi tanpa buka terminal. Aplikasi ini hanya
menulis di dalam home directory, tetapi tetap menyalin file dan membuat
launcher: periksa dulu apa yang kamu pilih, backup data penting, dan gunakan
dengan risiko sendiri. Disediakan "apa adanya", tanpa jaminan apa pun. LazyRun
tidak berafiliasi dengan vendor browser atau aplikasi apa pun yang dibuatkan
shortcut-nya.

## License

MIT.
