# PRD: LazyRun

> **lazy shortcut build**

**Versi:** 1.0
**Tanggal:** 2026-10-13
**Status:** Draft
**Platform:** Linux (KDE prioritas, GNOME fallback)
**Distribusi:** AppImage

---

## 1. Ringkasan

**LazyRun** adalah aplikasi desktop Linux berbasis Go yang memungkinkan pengguna membuat shortcut aplikasi — baik dari **file executable** maupun **PWA (URL web)** — ke menu aplikasi dan/atau desktop, hanya dalam beberapa klik. Dibungkus sebagai AppImage, LazyRun portable, mudah di-update, dan tidak memerlukan terminal.

> **Tagline:** *lazy shortcut build*

---

## 2. Masalah yang Diselesaikan

1. Membuat shortcut `.desktop` di Linux masih manual: harus tahu path XDG, format file, dan cara refresh cache KDE/GNOME.
2. Binary portable (Winbox, hasil build Go, tool CLI) sering tidak muncul di menu setelah diunduh.
3. PWA seperti WhatsApp Web, YouTube Music, atau Figma tidak bisa diakses sebagai "aplikasi" tanpa browser berbasis Chromium dengan flag khusus.
4. Tidak ada tool sederhana yang menggabungkan kedua kebutuhan di atas dengan antarmuka GUI.

---

## 3. Tujuan

- Membuat shortcut aplikasi dalam **< 30 detik** tanpa terminal.
- Mendukung **dua sumber**: Executable dan PWA.
- Menyediakan kontrol fleksibel: shortcut muncul di **menu aplikasi**, **desktop**, atau **keduanya**.
- Install, uninstall, dan update shortcut dengan mudah.
- Aplikasi LazyRun sendiri portable sebagai **single-file AppImage**.

### Non-Goals (v1.0)
- Tidak mendukung Windows/macOS.
- Tidak mengubah target menjadi AppImage.
- Tidak mengelola dependency binary target.
- Tidak menyediakan browser embedded untuk PWA.

---

## 4. Target Pengguna

- Pengguna Linux desktop yang sering mengunduh binary portable.
- Developer Go yang ingin hasil build muncul di menu.
- Pengguna yang ingin "install" PWA tanpa Chrome/Edge.
- Pengguna KDE Plasma (prioritas) dan GNOME.

---

## 5. Fitur Utama

### 5.1 Fungsi A — Shortcut dari Executable

Membuat shortcut dari file binary eksekutabel (ELF, AppImage, script dengan shebang).

**Input:**
- File binary (file picker).
- Ikon (opsional; auto-detect dari folder yang sama jika kosong).
- Nama aplikasi (auto-filled dari nama file).
- Kategori (Utility, Network, Development, dll).
- **Target lokasi** (lihat §5.3).

**Proses:**
1. Salin binary ke `~/.local/bin/<nama>` (mode `0755`).
2. Salin ikon ke `~/.local/share/icons/hicolor/512x512/apps/`.
3. Tulis file `.desktop` di `~/.local/share/applications/`.
4. Jika "Desktop" dicentang, salin juga ke `~/Desktop/` dan set `chmod +x`.
5. Refresh cache: `update-desktop-database` + `kbuildsycoca6 --noincremental`.

### 5.2 Fungsi B — Shortcut dari PWA (URL)

Membuat shortcut yang membuka URL di browser dalam mode app.

**Input:**
- URL (contoh: `https://web.whatsapp.com`).
- Nama aplikasi.
- Browser target (auto-detect: Chrome, Chromium, Brave, Edge, Firefox).
- Ikon (opsional; auto-fetch favicon dari URL jika kosong).
- **Target lokasi** (lihat §5.3).

**Proses:**
1. Fetch favicon atau pakai ikon default.
2. Simpan ikon ke direktori ikon XDG.
3. Tulis `.desktop` dengan `Exec=` menyesuaikan browser:
   - Chromium-based: `<browser> --app=<URL>`
   - Firefox: `<browser> --new-window <URL>`
4. Jika "Desktop" dicentang, salin ke `~/Desktop/` dan set `chmod +x`.
5. Refresh cache.

### 5.3 Kontrol Lokasi Shortcut (Fitur Kunci)

Setiap kali membuat shortcut, pengguna memilih lokasi via **checkbox** yang fleksibel:

```
┌─ Lokasi Shortcut ─────────────────────┐
│  [✓] Menu Aplikasi                    │
│  [ ] Desktop                          │
└───────────────────────────────────────┘
```

| Menu Aplikasi | Desktop | Hasil |
|:---:|:---:|---|
| ✓ | ✗ | Hanya muncul di menu aplikasi (launcher) |
| ✗ | ✓ | Hanya muncul di desktop |
| ✓ | ✓ | Muncul di keduanya (rekomendasi default) |
| ✗ | ✗ | Diblokir — minimal satu harus dipilih |

**Detail teknis:**
- **Menu Aplikasi**: file `.desktop` ditulis ke `~/.local/share/applications/`.
- **Desktop**: file `.desktop` yang sama di-*copy* ke `~/Desktop/` dengan `chmod +x` agar bisa diklik ganda.
- Sumber kebenaran tetap di `~/.local/share/applications/`; file di `~/Desktop/` adalah salinan (symlink opsional di v1.1).
- Saat uninstall, keduanya dihapus secara sinkron.

### 5.4 Manajemen Shortcut

- **List Shortcut**: menampilkan semua shortcut yang dibuat LazyRun.
- **Status**: menampilkan ikon/badge jika shortcut ada di Menu, Desktop, atau keduanya.
- **Uninstall**: hapus file `.desktop` dari semua lokasi, ikon, dan binary (opsional).
- **Open Location**: buka folder binary atau desktop di file manager.

### 5.5 Update Aplikasi LazyRun

- Bundle AppImage dengan dukungan **zsync**.
- Menu **"Check for Updates"** memanggil `AppImageUpdate`.
- Update otomatis dari GitHub Releases.

---

## 6. User Interface

### 6.1 Window Utama

```
┌──────────────────────────────────────────────────┐
│  LazyRun — lazy shortcut build                   │
├──────────────────────────────────────────────────┤
│  [+ Executable]      [+ Web App]                 │
├──────────────────────────────────────────────────┤
│  Shortcut Terinstall:                            │
│  ┌────────────────────────────────────────────┐  │
│  │ 🟦 Winbox          Network   [Menu][Desk]  │  │
│  │ 🌐 WhatsApp Web    Web       [Menu]        │  │
│  │ 🟩 my-go-app       Utility   [Menu][Desk]  │  │
│  └────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────┤
│  [Check for Updates]                     v1.0.0  │
└──────────────────────────────────────────────────┘
```

### 6.2 Dialog "New Executable Shortcut"

```
┌──────────────────────────────────────────┐
│  New Executable Shortcut                 │
├──────────────────────────────────────────┤
│  Binary:      [ Browse... ]              │
│  Icon:        [ Browse... ] (opsional)   │
│  Nama:        [ Winbox          ]        │
│  Kategori:    [ Network       ▼ ]        │
│                                          │
│  ── Lokasi Shortcut ──                   │
│  [✓] Menu Aplikasi                       │
│  [✓] Desktop                             │
│                                          │
│       [ Cancel ]      [ Install ]        │
└──────────────────────────────────────────┘
```

### 6.3 Dialog "New Web App Shortcut"

```
┌──────────────────────────────────────────┐
│  New Web App Shortcut                    │
├──────────────────────────────────────────┤
│  URL:         [ https://...       ]      │
│  Nama:        [ WhatsApp Web      ]      │
│  Browser:     [ Chrome           ▼ ]     │
│  Icon:        [ Auto-fetch    ] (opsi)   │
│                                          │
│  ── Lokasi Shortcut ──                   │
│  [✓] Menu Aplikasi                       │
│  [✓] Desktop                             │
│                                          │
│       [ Cancel ]      [ Install ]        │
└──────────────────────────────────────────┘
```

---

## 7. Arsitektur Teknis

### 7.1 Stack

| Komponen | Teknologi |
|---|---|
| Bahasa | Go 1.21+ |
| GUI | Fyne |
| Desktop entry | XDG `.desktop` spec |
| Cache refresh | `update-desktop-database`, `kbuildsycoca6`/`5` |
| Distribusi | AppImage + `appimagetool -u` (zsync) |

### 7.2 Struktur Direktori Proyek

```
lazyrun/
├── cmd/lazyrun/main.go
├── internal/
│   ├── shortcut/
│   │   ├── executable.go     # Fungsi A
│   │   ├── webapp.go         # Fungsi B
│   │   ├── desktop.go        # Generate .desktop
│   │   ├── desktopdir.go     # Copy ke ~/Desktop
│   │   └── cache.go          # Refresh cache
│   ├── icon/
│   │   ├── favicon.go
│   │   └── install.go
│   ├── browser/detect.go
│   ├── store/store.go        # Metadata JSON
│   └── ui/
│       ├── main_window.go
│       ├── tab_executable.go
│       └── tab_webapp.go
├── assets/
│   ├── icon.png
│   └── lazyrun.desktop
└── build.sh
```

### 7.3 Path XDG

| Tujuan | Path |
|---|---|
| Binary | `~/.local/bin/` |
| Ikon | `~/.local/share/icons/hicolor/512x512/apps/` |
| Menu aplikasi | `~/.local/share/applications/` |
| Desktop | `~/Desktop/` (via `xdg-user-dir DESKTOP`) |
| Metadata | `~/.config/lazyrun/shortcuts.json` |

### 7.4 Metadata Internal

`~/.config/lazyrun/shortcuts.json` menyimpan daftar shortcut LazyRun:

```json
[
  {
    "name": "Winbox",
    "type": "executable",
    "binary_path": "/home/user/.local/bin/winbox",
    "icon_path": "/home/user/.local/share/icons/hicolor/512x512/apps/Winbox.png",
    "desktop_path": "/home/user/.local/share/applications/Winbox.desktop",
    "desktop_copy_path": "/home/user/Desktop/Winbox.desktop",
    "targets": {
      "menu": true,
      "desktop": true
    },
    "installed_at": "2026-10-13T10:00:00Z"
  }
]
```

---

## 8. Alur Install & Uninstall

### 8.1 Install

```
User klik Install
       │
       ▼
Validasi input (min. satu target dicentang)
       │
       ▼
Salin binary + ikon ke XDG paths
       │
       ▼
Tulis .desktop di ~/.local/share/applications/
       │
       ├── Jika [Menu] ✓ → selesai di sini
       │
       └── Jika [Desktop] ✓ → salin .desktop ke ~/Desktop/ + chmod +x
       │
       ▼
Refresh cache (update-desktop-database + kbuildsycoca)
       │
       ▼
Simpan metadata ke shortcuts.json
       │
       ▼
Tampilkan notifikasi sukses
```

### 8.2 Uninstall

```
User klik 🗑 pada shortcut
       │
       ▼
Konfirmasi dialog
       │
       ▼
Hapus .desktop di semua lokasi (menu + desktop)
       │
       ▼
Hapus ikon (opsional)
       │
       ▼
Hapus binary (opsional, checkbox "hapus binary juga")
       │
       ▼
Refresh cache
       │
       ▼
Update shortcuts.json
```

---

## 9. Alur Build AppImage

```bash
# 1. Build binary
go build -o lazyrun ./cmd/lazyrun

# 2. Siapkan AppDir
mkdir -p lazyrun.AppDir/usr/bin
cp lazyrun lazyrun.AppDir/usr/bin/
cp assets/icon.png lazyrun.AppDir/lazyrun.png
cp assets/lazyrun.desktop lazyrun.AppDir/
ln -s usr/bin/lazyrun lazyrun.AppDir/AppRun

# 3. Build AppImage dengan dukungan update
./appimagetool-x86_64.AppImage \
  -u "gh-releases-zsync|user|lazyrun|latest|lazyrun-*.AppImage.zsync" \
  lazyrun.AppDir
```

---

## 10. Persyaratan Non-Fungsional

| Aspek | Target |
|---|---|
| Ukuran AppImage | < 30 MB |
| Startup time | < 2 detik |
| Bahasa UI | Indonesia + Inggris |
| Kompatibilitas | KDE Plasma 5 & 6, GNOME 42+ |
| Hak akses | Tanpa root, semua di `~/.local/` |
| Fallback cache | `kbuildsycoca6` → `kbuildsycoca5` → skip |
| Desktop path | Auto-detect via `xdg-user-dir DESKTOP` |

---

## 11. Roadmap

### v1.0 (MVP)
- [ ] GUI Fyne dengan window utama & 2 dialog.
- [ ] Fungsi A: shortcut dari executable.
- [ ] Fungsi B: shortcut dari PWA (Chrome/Chromium/Brave).
- [ ] Checkbox "Menu Aplikasi" & "Desktop" dengan validasi min. satu.
- [ ] Uninstall shortcut.
- [ ] Deteksi browser terinstall.
- [ ] Bundle AppImage + zsync update.

### v1.1
- [ ] Edit shortcut yang sudah ada.
- [ ] Auto-detect binary di `$PATH` (htop, nvim, dll).
- [ ] Symlink desktop (bukan copy) sebagai opsi.
- [ ] Import/Export daftar shortcut.
- [ ] Dark mode otomatis.

### v1.2
- [ ] Dukungan Firefox PWA (via `--kiosk`).
- [ ] Dukungan Flatpak & Snap.
- [ ] Preview ikon sebelum install.

---

## 12. Risiko & Mitigasi

| Risiko | Mitigasi |
|---|---|
| `~/Desktop/` bukan bahasa Inggris (mis. `~/Desktop` vs `~/デスクトップ`) | Gunakan `xdg-user-dir DESKTOP` |
| KDE versi beda (`kbuildsycoca5`/`6`) | Deteksi otomatis, fallback berurutan |
| Firefox tanpa mode PWA penuh | Fallback `--new-window`, atau rekomendasi Chromium |
| Favicon gagal di-fetch | Fallback ke ikon default LazyRun |
| User hapus file desktop manual | Deteksi saat list, tandai "broken", tawarkan repair |
| File `.desktop` di `~/Desktop/` perlu `chmod +x` | Set otomatis saat install |

---

## 13. Kriteria Sukses

- Shortcut dapat dibuat dalam **< 30 detik** tanpa terminal.
- 90% pengguna KDE melihat shortcut muncul **tanpa restart**.
- Checkbox Menu/Desktop berfungsi konsisten di KDE & GNOME.
- Berjalan sebagai **single-file AppImage** di Ubuntu, Fedora, Arch, openSUSE.
- Update otomatis berjalan tanpa intervensi manual.

---

## 14. Lampiran: Contoh File `.desktop` Hasil LazyRun

**Executable:**
```ini
[Desktop Entry]
Type=Application
Name=Winbox
Exec=/home/user/.local/bin/winbox
Icon=/home/user/.local/share/icons/hicolor/512x512/apps/Winbox.png
Terminal=false
Categories=Network;
Comment=Created by LazyRun
```

**PWA:**
```ini
[Desktop Entry]
Type=Application
Name=WhatsApp Web
Exec=/usr/bin/google-chrome --app=https://web.whatsapp.com
Icon=/home/user/.local/share/icons/hicolor/512x512/apps/WhatsApp Web.png
Terminal=false
Categories=Network;WebBrowser;
Comment=Created by LazyRun
```

---

*PRD ini adalah living document. Perubahan signifikan akan dicatat di changelog versi berikutnya.*
