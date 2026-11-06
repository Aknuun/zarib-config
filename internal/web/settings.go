package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	settingsVersion  = 1
	maxProxyLines    = 50
	maxProxyURLLen   = 2048
	maxAdminUserLen  = 128
	maxAdminPassLen  = 256
	maxTelegramToken = 512
	maxTelegramChat  = 128
)

// proxySetting is one ordered subscription proxy. A disabled entry stays in the
// list so the operator can turn it back on without retyping the credentials.
type proxySetting struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type settingsData struct {
	Version       int
	Proxies       []proxySetting
	TelegramToken string
	TelegramChat  string
	AdminUser     string
	AdminPass     string
}

// settingsFile uses pointers so a missing field falls back to the environment
// while an explicit empty value clears the setting.
type settingsFile struct {
	Version       int             `json:"version"`
	Proxies       *[]proxySetting `json:"proxies,omitempty"`
	TelegramToken *string         `json:"telegram_token,omitempty"`
	TelegramChat  *string         `json:"telegram_chat_id,omitempty"`
	AdminUser     *string         `json:"admin_user,omitempty"`
	AdminPass     *string         `json:"admin_password,omitempty"`
}

func defaultSettings() settingsData {
	return settingsData{
		Version:       settingsVersion,
		Proxies:       envProxySettings(),
		TelegramToken: strings.TrimSpace(os.Getenv("ZARIB_TELEGRAM_TOKEN")),
		TelegramChat:  strings.TrimSpace(os.Getenv("ZARIB_TELEGRAM_CHAT_ID")),
		AdminUser:     strings.TrimSpace(os.Getenv("ADMIN_USER")),
		AdminPass:     os.Getenv("ADMIN_PASSWORD"),
	}
}

func envProxySettings() []proxySetting {
	var entries []proxySetting
	for _, proxy := range ratio.SubProxyURLs() {
		entries = append(entries, proxySetting{URL: proxy.String(), Enabled: true})
	}
	return entries
}

func (s *server) loadSettings() {
	value := defaultSettings()
	if s.settingsFile != "" {
		data, err := os.ReadFile(s.settingsFile)
		switch {
		case err == nil:
			var file settingsFile
			if err := json.Unmarshal(data, &file); err != nil {
				fmt.Printf("settings: could not parse settings: %v\n", err)
				break
			}
			if file.Proxies != nil {
				value.Proxies = normalizeProxySettings(*file.Proxies)
			}
			if file.TelegramToken != nil {
				value.TelegramToken = strings.TrimSpace(*file.TelegramToken)
			}
			if file.TelegramChat != nil {
				value.TelegramChat = strings.TrimSpace(*file.TelegramChat)
			}
			if file.AdminUser != nil {
				value.AdminUser = strings.TrimSpace(*file.AdminUser)
			}
			if file.AdminPass != nil {
				value.AdminPass = *file.AdminPass
			}
		case !os.IsNotExist(err):
			fmt.Printf("settings: could not read settings: %v\n", err)
		}
	}
	s.applySettings(value)
}

func normalizeProxySettings(entries []proxySetting) []proxySetting {
	out := make([]proxySetting, 0, len(entries))
	for _, entry := range entries {
		entry.URL = strings.TrimSpace(entry.URL)
		if entry.URL == "" {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func (s *server) applySettings(value settingsData) {
	s.settingsMu.Lock()
	s.settings = value
	s.telegramToken = strings.TrimSpace(value.TelegramToken)
	s.telegramChatID = strings.TrimSpace(value.TelegramChat)
	s.adminUser = strings.TrimSpace(value.AdminUser)
	s.adminPassword = value.AdminPass
	s.settingsMu.Unlock()
	ratio.SetSubProxyURLs(proxyURLs(value.Proxies))
}

func settingsFileFrom(value settingsData) settingsFile {
	proxies := append([]proxySetting(nil), value.Proxies...)
	token := value.TelegramToken
	chat := value.TelegramChat
	user := value.AdminUser
	password := value.AdminPass
	return settingsFile{
		Version:       settingsVersion,
		Proxies:       &proxies,
		TelegramToken: &token,
		TelegramChat:  &chat,
		AdminUser:     &user,
		AdminPass:     &password,
	}
}

func (s *server) saveSettings(value settingsData) error {
	value.Version = settingsVersion
	data, err := json.MarshalIndent(settingsFileFrom(value), "", "  ")
	if err != nil {
		return err
	}
	if s.settingsFile != "" {
		if err := writeFileAtomic(s.settingsFile, data); err != nil {
			return err
		}
	}
	s.applySettings(value)
	return nil
}

func (s *server) settingsSnapshot() settingsData {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	value := s.settings
	// Servers built directly in tests may only set the runtime fields.
	if value.AdminUser == "" {
		value.AdminUser = s.adminUser
	}
	if value.AdminPass == "" {
		value.AdminPass = s.adminPassword
	}
	if value.TelegramToken == "" {
		value.TelegramToken = s.telegramToken
	}
	if value.TelegramChat == "" {
		value.TelegramChat = s.telegramChatID
	}
	value.Proxies = append([]proxySetting(nil), value.Proxies...)
	return value
}

func (s *server) updateSettings(change func(*settingsData) error) error {
	next := s.settingsSnapshot()
	if err := change(&next); err != nil {
		return err
	}
	return s.saveSettings(next)
}

func (s *server) telegramConfig() (string, string) {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	return s.telegramToken, s.telegramChatID
}

func (s *server) adminCredentials() (string, string) {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	return s.adminUser, s.adminPassword
}

func proxyURLs(entries []proxySetting) []*url.URL {
	var out []*url.URL
	for _, entry := range entries {
		if !entry.Enabled {
			continue
		}
		parsed, err := parseProxyURL(entry.URL)
		if err == nil {
			out = append(out, parsed)
		}
	}
	return out
}

func parseProxyURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty proxy URL")
	}
	if len(raw) > maxProxyURLLen {
		return nil, errors.New("proxy URL is too long")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return nil, errors.New("invalid proxy URL")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return nil, errors.New("proxy URL must use http, https, or socks5")
	}
	return parsed, nil
}

// parseProxyText accepts one proxy URL per line. A leading # or ! keeps the
// entry in the list but disabled; deleting the line removes it.
func parseProxyText(text string) ([]proxySetting, error) {
	var entries []proxySetting
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		enabled := true
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			enabled = false
			line = strings.TrimSpace(line[1:])
			if line == "" {
				continue
			}
		}
		if _, err := parseProxyURL(line); err != nil {
			return nil, err
		}
		entries = append(entries, proxySetting{URL: line, Enabled: enabled})
		if len(entries) > maxProxyLines {
			return nil, errors.New("too many proxies")
		}
	}
	return entries, nil
}

func proxyText(entries []proxySetting) string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		line := entry.URL
		if !entry.Enabled {
			line = "# " + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (s *server) proxyEnv() []string {
	value := s.settingsSnapshot()
	var urls []string
	for _, proxy := range proxyURLs(value.Proxies) {
		urls = append(urls, proxy.String())
	}
	return urls
}
