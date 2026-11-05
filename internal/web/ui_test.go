package web

import (
	"path/filepath"
	"testing"
)

func TestUISettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.json")
	s := &server{uiFile: path}
	s.ui = defaultUISettings()
	s.ui.Animations = "off"
	s.ui.Page = false
	s.ui.Tiles = false
	s.uiMu.Lock()
	err := s.saveUISettingsLocked()
	s.uiMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	loaded := &server{uiFile: path}
	loaded.loadUISettings()
	if loaded.ui.Animations != "off" || loaded.ui.Page || loaded.ui.Tiles {
		t.Fatalf("unexpected settings: %+v", loaded.ui)
	}
}
