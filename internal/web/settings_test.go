package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

func TestSettingsRoundTrip(t *testing.T) {
	t.Setenv("ADMIN_USER", "env-admin")
	t.Setenv("ADMIN_PASSWORD", "env-password")
	t.Setenv("ZARIB_TELEGRAM_TOKEN", "env-token")
	t.Setenv("ZARIB_TELEGRAM_CHAT_ID", "env-chat")
	t.Cleanup(func() { ratio.SetSubProxyURLs(nil) })

	path := filepath.Join(t.TempDir(), "settings.json")
	created := &server{settingsFile: path}
	created.loadSettings()
	if created.adminUser != "env-admin" || created.telegramToken != "env-token" {
		t.Fatalf("settings did not start from env: %+v", created.settingsSnapshot())
	}

	value := created.settingsSnapshot()
	value.Proxies = []proxySetting{
		{URL: "http://user:pass@127.0.0.1:8080", Enabled: true},
		{URL: "http://127.0.0.1:9090", Enabled: false},
	}
	value.TelegramToken = "file-token"
	value.TelegramChat = "file-chat"
	value.AdminUser = "file-admin"
	value.AdminPass = "file-password"
	if err := created.saveSettings(value); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("settings mode = %o, want 600", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"version": 1`) {
		t.Fatalf("settings file has no version: %s", data)
	}

	loaded := &server{settingsFile: path}
	loaded.loadSettings()
	got := loaded.settingsSnapshot()
	if got.TelegramToken != "file-token" || got.TelegramChat != "file-chat" {
		t.Fatalf("telegram settings did not persist: %+v", got)
	}
	if got.AdminUser != "file-admin" || got.AdminPass != "file-password" {
		t.Fatalf("admin settings did not persist: %+v", got)
	}
	if len(got.Proxies) != 2 || !got.Proxies[0].Enabled || got.Proxies[1].Enabled {
		t.Fatalf("proxy settings did not persist: %+v", got.Proxies)
	}
	if loaded.adminUser != "file-admin" || loaded.adminPassword != "file-password" {
		t.Fatal("admin credentials were not applied to the running server")
	}
}

func TestSettingsFallsBackToEnv(t *testing.T) {
	t.Setenv("ZARIB_TELEGRAM_TOKEN", "env-token")
	t.Setenv("ZARIB_TELEGRAM_CHAT_ID", "env-chat")
	t.Setenv("ADMIN_USER", "env-admin")
	t.Setenv("ADMIN_PASSWORD", "env-password")
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"telegram_chat_id":"file-chat"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	s := &server{settingsFile: path}
	s.loadSettings()
	if s.telegramToken != "env-token" {
		t.Fatalf("missing token did not fall back to env: %q", s.telegramToken)
	}
	if s.telegramChatID != "file-chat" {
		t.Fatalf("file chat id was not used: %q", s.telegramChatID)
	}
	if s.adminUser != "env-admin" || s.adminPassword != "env-password" {
		t.Fatal("missing admin settings did not fall back to env")
	}
}

func TestAdminSettingsEndpoint(t *testing.T) {
	t.Cleanup(func() { ratio.SetSubProxyURLs(nil) })
	s := adminTestServer(t)
	s.settingsFile = filepath.Join(t.TempDir(), "settings.json")

	post := func(form url.Values, origin string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/admin/api/settings", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Origin", origin)
		request.SetBasicAuth("admin", "test-password")
		response := httptest.NewRecorder()
		s.route(response, request)
		return response
	}

	proxies := url.Values{
		"section": {"proxies"},
		"proxies": {"http://user:pass@127.0.0.1:8080\nhttp://127.0.0.1:9090\n# http://127.0.0.1:9091"},
	}
	if response := post(proxies, "http://evil.example"); response.Code != http.StatusForbidden {
		t.Fatalf("cross-origin settings: status %d", response.Code)
	}
	if response := post(proxies, "http://example.com"); response.Code != http.StatusSeeOther {
		t.Fatalf("save proxies: status %d: %s", response.Code, response.Body.String())
	}
	effective := ratio.SubProxyURLs()
	if len(effective) != 2 {
		t.Fatalf("effective proxies = %d, want 2", len(effective))
	}
	if effective[0].Host != "127.0.0.1:8080" || effective[1].Host != "127.0.0.1:9090" {
		t.Fatalf("proxy order was not applied: %v", effective)
	}
	if effective[0].User == nil || effective[0].User.Username() != "user" {
		t.Fatal("proxy credentials were not kept")
	}

	telegram := url.Values{
		"section":          {"telegram"},
		"telegram_token":   {"new-token"},
		"telegram_chat_id": {"-100123"},
	}
	if response := post(telegram, "http://example.com"); response.Code != http.StatusSeeOther {
		t.Fatalf("save telegram: status %d", response.Code)
	}
	if token, chat := s.telegramConfig(); token != "new-token" || chat != "-100123" {
		t.Fatalf("telegram settings were not applied immediately")
	}
}

func TestAdminPageRendersSettingsWithoutSecrets(t *testing.T) {
	s := adminTestServer(t)
	s.settingsFile = filepath.Join(t.TempDir(), "settings.json")
	s.telegramToken = "secret-token"
	s.telegramChatID = "12345"

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("admin page: status %d", response.Code)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("admin cache-control = %q, want no-store", got)
	}
	body := response.Body.String()
	for _, want := range []string{"پروکسی‌های ساب", "توکن ربات", "دسترسی مدیریت"} {
		if !strings.Contains(body, want) {
			t.Fatalf("admin page missing %q", want)
		}
	}
	for _, secret := range []string{"secret-token", "test-password"} {
		if strings.Contains(body, secret) {
			t.Fatalf("admin page leaked a secret")
		}
	}
}

func TestAdminPasswordChangeInvalidatesOldPassword(t *testing.T) {
	s := adminTestServer(t)
	s.settingsFile = filepath.Join(t.TempDir(), "settings.json")

	form := url.Values{
		"section":          {"credentials"},
		"admin_user":       {"admin"},
		"current_password": {"test-password"},
		"new_password":     {"new-password"},
		"confirm_password": {"new-password"},
	}
	request := httptest.NewRequest(http.MethodPost, "/admin/api/settings", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Origin", "http://example.com")
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("change password: status %d: %s", response.Code, response.Body.String())
	}

	check := func(user, password string) int {
		request := httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
		request.SetBasicAuth(user, password)
		response := httptest.NewRecorder()
		s.route(response, request)
		return response.Code
	}
	if code := check("admin", "test-password"); code != http.StatusUnauthorized {
		t.Fatalf("old password still works: status %d", code)
	}
	if code := check("admin", "new-password"); code != http.StatusOK {
		t.Fatalf("new password does not work: status %d", code)
	}
}

func TestAdminPasswordChangeRequiresCurrentPassword(t *testing.T) {
	s := adminTestServer(t)
	s.settingsFile = filepath.Join(t.TempDir(), "settings.json")
	form := url.Values{
		"section":          {"credentials"},
		"admin_user":       {"admin"},
		"current_password": {"wrong"},
		"new_password":     {"new-password"},
		"confirm_password": {"new-password"},
	}
	request := httptest.NewRequest(http.MethodPost, "/admin/api/settings", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Origin", "http://example.com")
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("wrong current password: status %d", response.Code)
	}
	if s.adminPassword != "test-password" {
		t.Fatal("password changed without the current password")
	}
}
