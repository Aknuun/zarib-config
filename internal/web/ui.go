package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// uiSettings is site-wide presentation state. It is public (the browser needs
// it) and contains no secrets.
type uiSettings struct {
	Version    int    `json:"version,omitempty"`
	Animations string `json:"animations"`
	Page       bool   `json:"page"`
	Rows       bool   `json:"rows"`
	Chips      bool   `json:"chips"`
	Result     bool   `json:"result"`
	Buttons    bool   `json:"buttons"`
	Tiles      bool   `json:"tiles"`
}

func defaultUISettings() uiSettings {
	return uiSettings{
		Version:    1,
		Animations: "full",
		Page:       true,
		Rows:       true,
		Chips:      true,
		Result:     true,
		Buttons:    true,
		Tiles:      true,
	}
}

func normalizeUISettings(value uiSettings) uiSettings {
	value.Version = 1
	switch value.Animations {
	case "full", "light", "off":
	default:
		value.Animations = "full"
	}
	return value
}

func (s *server) loadUISettings() {
	s.ui = defaultUISettings()
	if s.uiFile == "" {
		return
	}
	data, err := os.ReadFile(s.uiFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("ui: could not read settings: %v\n", err)
		}
		return
	}
	var value uiSettings
	if err := json.Unmarshal(data, &value); err != nil {
		fmt.Printf("ui: could not parse settings: %v\n", err)
		return
	}
	s.ui = normalizeUISettings(value)
}

// saveUISettingsLocked must be called with s.uiMu held.
func (s *server) saveUISettingsLocked() error {
	if s.uiFile == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.ui, "", "  ")
	if err != nil {
		return err
	}
	temp := s.uiFile + ".tmp"
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, s.uiFile)
}

func (s *server) uiSnapshot() uiSettings {
	s.uiMu.Lock()
	defer s.uiMu.Unlock()
	return s.ui
}

func (s *server) uiConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.uiSnapshot())
}

func (s *server) adminUI(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, false) {
		return
	}
	if !sameOriginRequest(r) {
		http.Error(w, "Cross-origin request blocked.", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form.", http.StatusBadRequest)
		return
	}
	next := uiSettings{
		Animations: strings.TrimSpace(r.FormValue("animations")),
		Page:       r.FormValue("page") == "on",
		Rows:       r.FormValue("rows") == "on",
		Chips:      r.FormValue("chips") == "on",
		Result:     r.FormValue("result") == "on",
		Buttons:    r.FormValue("buttons") == "on",
		Tiles:      r.FormValue("tiles") == "on",
	}
	next = normalizeUISettings(next)
	s.uiMu.Lock()
	s.ui = next
	err := s.saveUISettingsLocked()
	s.uiMu.Unlock()
	if err != nil {
		fmt.Printf("ui: could not save settings: %v\n", err)
		http.Error(w, "Could not save animation settings.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin#animations", http.StatusSeeOther)
}

func uiFilePath(cacheRoot string) string {
	return filepath.Join(cacheRoot, "ui.json")
}
