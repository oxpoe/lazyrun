package shortcut

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	e := Entry{
		Name:       "Winbox",
		ExecArgs:   []string{"/home/user/.local/bin/winbox"},
		Icon:       "/home/user/.local/share/icons/hicolor/512x512/apps/winbox.png",
		Categories: "Network;",
	}
	want := `[Desktop Entry]
Type=Application
Name=Winbox
Exec=/home/user/.local/bin/winbox
Icon=/home/user/.local/share/icons/hicolor/512x512/apps/winbox.png
Terminal=false
Categories=Network;
Comment=Created by LazyRun
`
	if got := e.Render(); got != want {
		t.Errorf("Render()\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderNoDisplayWhenDesktopOnly(t *testing.T) {
	e := Entry{Name: "X", ExecArgs: []string{"/x"}, NoDisplay: true}
	if !strings.Contains(e.Render(), "NoDisplay=true\n") {
		t.Errorf("entry desktop-only harus NoDisplay=true:\n%s", e.Render())
	}
}

func TestExecQuote(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/usr/bin/winbox", "/usr/bin/winbox"},
		{"/home/user/my app/app", `"/home/user/my app/app"`},
		{"/opt/a$b", `"/opt/a\$b"`},
		{"back\\slash path", `"back\\slash path"`},
		{`say "hi"`, `"say \"hi\""`},
		{"", `""`},
		{"--app=https://web.whatsapp.com", "--app=https://web.whatsapp.com"},
	}
	for _, tc := range tests {
		if got := execQuote(tc.in); got != tc.want {
			t.Errorf("execQuote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestExecArgsJoinsArgs(t *testing.T) {
	got := execArgs([]string{"/usr/bin/chromium", "--app=https://web.whatsapp.com"})
	want := "/usr/bin/chromium --app=https://web.whatsapp.com"
	if got != want {
		t.Errorf("execArgs() = %q, want %q", got, want)
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"WhatsApp Web": "whatsapp-web",
		"Winbox":       "winbox",
		"my-go-app":    "my-go-app",
		"  spaced  ":   "spaced",
		"Café #1":      "café-1",
		"!!":           "",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	if got, err := normalizeURL("web.whatsapp.com"); err != nil || got != "https://web.whatsapp.com" {
		t.Errorf("normalizeURL tanpa skema = %q, %v", got, err)
	}
	if _, err := normalizeURL(""); err == nil {
		t.Error("URL kosong harus error")
	}
	if _, err := normalizeURL("ftp://example.com"); err == nil {
		t.Error("skema selain http(s) harus error")
	}
}
