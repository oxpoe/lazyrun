package ui

import (
	"os"
	"strings"
)

// tr menyimpan teks UI: index 0 = Inggris, 1 = Indonesia.
var tr = map[string][2]string{
	"app_title":         {"LazyRun — lazy shortcut build", "LazyRun — lazy shortcut build"},
	"tagline":           {"lazy shortcut build", "lazy shortcut build"},
	"btn_executable":    {"Executable", "Executable"},
	"btn_webapp":        {"Web App", "Aplikasi Web"},
	"installed":         {"Installed shortcuts", "Shortcut terinstall"},
	"empty_title":       {"Nothing here yet", "Belum ada apa-apa"},
	"empty_hint":        {"Create your first shortcut with the buttons in the corner.", "Buat shortcut pertamamu lewat tombol di pojok kanan atas."},
	"chip_menu":         {"Menu", "Menu"},
	"chip_desktop":      {"Desktop", "Desktop"},
	"chip_broken":       {"file missing", "file hilang"},
	"none":              {"not selected", "belum dipilih"},
	"kind_unknown":      {"unknown type", "jenis tidak dikenal"},
	"icon_auto":         {"automatic", "otomatis"},
	"online":            {"Online", "Online"},
	"search":            {"Search", "Cari"},
	"searching":         {"Searching...", "Mencari..."},
	"search_failed":     {"Search failed. Check your connection.", "Pencarian gagal. Cek koneksi kamu."},
	"search_found":      {"%d icons found - click one to use it.", "%d ikon ditemukan - klik satu untuk dipakai."},
	"search_hint":       {"Type a name and press Search.", "Tulis nama lalu tekan Cari."},
	"icon_search_title": {"Icon from the internet", "Ikon dari internet"},
	"sync_icon":         {"Sync icon", "Sinkron ikon"},
	"browse":            {"Browse...", "Pilih..."},
	"install":           {"Install", "Install"},
	"cancel":            {"Cancel", "Batal"},
	"back":              {"Back", "Kembali"},
	"delete":            {"Delete", "Hapus"},
	"working":           {"Working...", "Sedang diproses..."},

	"new_exec_title": {"New Executable Shortcut", "Shortcut Executable Baru"},
	"new_web_title":  {"New Web App Shortcut", "Shortcut Web App Baru"},
	"field_binary":   {"Binary", "Binary"},
	"field_icon":     {"Icon (optional)", "Ikon (opsional)"},
	"field_name":     {"Name", "Nama"},
	"field_category": {"Category", "Kategori"},
	"field_url":      {"URL", "URL"},
	"field_browser":  {"Browser", "Browser"},

	"section_location": {"Shortcut location", "Lokasi shortcut"},
	"check_menu":       {"Application menu", "Menu aplikasi"},
	"check_desktop":    {"Desktop", "Desktop"},
	"exec_hint":        {"The binary is copied to ~/.local/bin and can be uninstalled from the list.", "Binary disalin ke ~/.local/bin dan bisa dihapus lagi dari daftar."},
	"web_hint":         {"The icon comes from the site's favicon, or is looked up online if there is none.", "Ikon diambil dari favicon situs, atau dicari otomatis di internet kalau tidak ada."},

	"confirm_delete_title": {"Uninstall shortcut", "Hapus shortcut"},
	"confirm_delete_msg":   {"Remove %s?", "Hapus %s?"},
	"also_binary":          {"Also delete the copied binary", "Hapus binary hasil salinan juga"},

	"success_title": {"Done", "Selesai"},
	"success_msg":   {"%s is ready.", "%s sudah siap."},
	"err_title":     {"Cannot continue", "Tidak bisa lanjut"},

	"need_binary": {"Choose a binary first.", "Pilih binary dulu."},
	"need_name":   {"Enter a name first.", "Masukkan nama dulu."},
	"need_url":    {"Enter a URL first.", "Masukkan URL dulu."},
	"need_target": {"Pick at least one location: menu or desktop.", "Pilih minimal satu lokasi: menu atau desktop."},
	"no_browser":  {"No supported browser found. Install Chrome, Chromium, Brave, Edge, Vivaldi, or Firefox first.", "Browser yang didukung tidak ditemukan. Install Chrome, Chromium, Brave, Edge, Vivaldi, atau Firefox dulu."},

	"github":        {"GitHub", "GitHub"},
	"donate":        {"Donate", "Donasi"},
	"check_updates": {"Check for Updates", "Cek Update"},

	"update_available":   {"LazyRun %s is available (you are on %s).", "LazyRun %s sudah tersedia (kamu pakai %s)."},
	"update_hint":        {"The new AppImage is downloaded and swapped in; restart afterwards to use it.", "AppImage baru diunduh lalu ditukar; restart setelahnya untuk memakainya."},
	"update_now":         {"Update now", "Update sekarang"},
	"update_later":       {"Later", "Nanti"},
	"update_page":        {"Release page", "Halaman rilis"},
	"update_downloading": {"Downloading update...", "Mengunduh update..."},
	"update_installed":   {"Update installed. Restart LazyRun to use the new version.", "Update terpasang. Restart LazyRun untuk memakai versi baru."},
	"update_restart":     {"Restart now", "Restart sekarang"},
	"up_to_date":         {"LazyRun %s is the latest version.", "LazyRun %s adalah versi terbaru."},
	"dev_build":          {"This is a development build (%s), update check is skipped.", "Ini build development (%s), cek update dilewati."},
}

var lang = detectLang()

// T mengembalikan teks UI sesuai $LANG: "id*" → Indonesia, selain itu Inggris.
func T(key string) string {
	s, ok := tr[key]
	if !ok {
		return key
	}
	if lang == "id" {
		return s[1]
	}
	return s[0]
}

func detectLang() string {
	v := os.Getenv("LC_ALL")
	if v == "" {
		v = os.Getenv("LC_MESSAGES")
	}
	if v == "" {
		v = os.Getenv("LANG")
	}
	if strings.HasPrefix(strings.ToLower(v), "id") {
		return "id"
	}
	return "en"
}
