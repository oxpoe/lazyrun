package ui

import (
	"os"
	"strings"
)

// tr menyimpan teks UI: index 0 = Inggris, 1 = Indonesia.
var tr = map[string][2]string{
	"app_title":      {"LazyRun — lazy shortcut build", "LazyRun — lazy shortcut build"},
	"tagline":        {"lazy shortcut build", "lazy shortcut build"},
	"btn_executable": {"Executable", "Executable"},
	"btn_webapp":     {"Web App", "Aplikasi Web"},
	"installed":      {"Installed shortcuts", "Shortcut terinstall"},
	"empty_title":    {"Nothing here yet", "Belum ada apa-apa"},
	"empty_hint":     {"Create your first shortcut with the buttons in the corner.", "Buat shortcut pertamamu lewat tombol di pojok kanan atas."},
	"chip_menu":      {"Menu", "Menu"},
	"chip_desktop":   {"Desktop", "Desktop"},
	"chip_broken":    {"file missing", "file hilang"},
	"none":           {"not selected", "belum dipilih"},
	"icon_auto":      {"automatic", "otomatis"},
	"browse":         {"Browse...", "Pilih..."},
	"install":        {"Install", "Install"},
	"cancel":         {"Cancel", "Batal"},
	"delete":         {"Delete", "Hapus"},
	"folder":         {"Folder", "Folder"},
	"working":        {"Working...", "Sedang diproses..."},

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
	"web_hint":         {"The icon is fetched from the site's favicon if you leave it empty.", "Ikon diambil otomatis dari favicon situs kalau dibiarkan kosong."},

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

	"github":            {"GitHub", "GitHub"},
	"donate":            {"Donate", "Donasi"},
	"check_updates":     {"Check for Updates", "Cek Update"},
	"not_appimage":      {"This build is not an AppImage (version %s). Update from where you got it.", "Build ini bukan AppImage (versi %s). Update dari sumber kamu mendapatkannya."},
	"no_appimageupdate": {"AppImageUpdate is not installed. Install it to update LazyRun automatically.", "AppImageUpdate tidak terpasang. Install dulu supaya LazyRun bisa update otomatis."},
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
