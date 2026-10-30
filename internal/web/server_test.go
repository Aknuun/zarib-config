package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

func testServer(t *testing.T) *server {
	t.Helper()
	return &server{
		jobsDir:  t.TempDir(),
		jobs:     map[string]*job{},
		rates:    map[string]map[string][]time.Time{},
		presence: map[string]float64{},
		geoCache: map[string]geoEntry{},
		sem:      make(chan struct{}, maxActive),
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func TestInspect(t *testing.T) {
	links := "vless://11111111-2222-3333-4444-555555555555@example.com:443?type=ws#node-a%20%C3%971.2\n" +
		"trojan://secret@example.net:443#node-b\n"
	body := base64.StdEncoding.EncodeToString([]byte(links))
	subscription := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Subscription-Userinfo", "upload=1024; download=2048; total=10737418240; expire=1893456000")
		fmt.Fprint(w, body)
	}))
	defer subscription.Close()
	oldPublic := publicURLFn
	publicURLFn = func(string) bool { return true }
	defer func() { publicURLFn = oldPublic }()
	t.Setenv("RATIO_SAFE_FETCH", "0")
	s := &server{
		jobsDir: t.TempDir(),
		jobs:    map[string]*job{},
		rates:   map[string]map[string][]time.Time{},
		sem:     make(chan struct{}, maxActive),
	}
	payload := fmt.Sprintf(`{"sub_url":%q}`, subscription.URL)
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(payload))
	response := httptest.NewRecorder()
	s.inspect(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	var data map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data["count"].(float64) != 2 {
		t.Fatalf("unexpected count: %v", data["count"])
	}
	configs := data["configs"].([]any)
	first := configs[0].(map[string]any)
	if first["protocol"] != "vless" || first["target"] != "example.com:443" {
		t.Fatalf("unexpected config: %+v", first)
	}
	if first["claimed"].(float64) != 1.2 {
		t.Fatalf("unexpected claimed: %v", first["claimed"])
	}
	usage := data["usage"].(map[string]any)
	if usage["used"].(float64) != 3072 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	if usage["remaining"].(float64) != 10737415168 {
		t.Fatalf("unexpected remaining: %v", usage["remaining"])
	}
}

func TestInspectRejectsInvalidURL(t *testing.T) {
	s := &server{
		jobsDir: t.TempDir(),
		jobs:    map[string]*job{},
		rates:   map[string]map[string][]time.Time{},
		sem:     make(chan struct{}, maxActive),
	}
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"sub_url":"not-a-url"}`))
	response := httptest.NewRecorder()
	s.inspect(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "invalid_url" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestClassifyError(t *testing.T) {
	cases := map[string]string{
		"The config did not respond through the proxy.":      "probe_failed",
		"Cannot resolve host 'x'.":                           "resolve_failed",
		"The config is not valid for Xray:":                  "xray_failed",
		"Could not read the subscription.":                   "fetch_failed",
		"No supported configs found":                         "no_configs",
		"Usage reported by the subscription did not change.": "usage_unchanged",
	}
	for input, expected := range cases {
		if actual := classifyError([]string{input}, false, false); actual != expected {
			t.Fatalf("classifyError(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestPresenceCountsDistinctClients(t *testing.T) {
	s := testServer(t)
	for _, id := range []string{"client-11111111", "client-22222222"} {
		payload := fmt.Sprintf(`{"id":%q}`, id)
		request := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(payload))
		response := httptest.NewRecorder()
		s.presenceHandler(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(`{"id":"client-11111111"}`))
	response := httptest.NewRecorder()
	s.presenceHandler(response, request)
	var data map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data["online"].(float64) != 2 {
		t.Fatalf("unexpected online count: %v", data["online"])
	}
	for _, key := range []string{"today", "monthly", "overall"} {
		if _, ok := data[key].(map[string]any); !ok {
			t.Fatalf("missing %s stats: %+v", key, data)
		}
	}
}

func TestPresenceRejectsInvalidClientID(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(`{"id":"x"}`))
	response := httptest.NewRecorder()
	s.presenceHandler(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status %d", response.Code)
	}
}

func TestRecordResultClassifiesRatios(t *testing.T) {
	s := testServer(t)
	s.recordResult(&ratio.Result{Ratio: 1.5, DownloadedBytes: 100, DeltaTotal: 150}, boolPtr(true))
	s.recordResult(&ratio.Result{Ratio: 1.0, DownloadedBytes: 100, DeltaTotal: 100}, nil)
	s.recordResult(&ratio.Result{Ratio: 0.5, DownloadedBytes: 100, DeltaTotal: 50}, boolPtr(false))
	s.mu.Lock()
	stats := s.live
	s.mu.Unlock()
	if stats.Tests != 3 || stats.WithRatio != 1 || stats.Normal != 1 || stats.Invalid != 1 {
		t.Fatalf("unexpected ratio stats: %+v", stats)
	}
	if stats.Tunnel != 1 || stats.NonTunnel != 1 || stats.Unknown != 1 {
		t.Fatalf("unexpected tunnel stats: %+v", stats)
	}
	if stats.Downloaded != 300 || stats.Deducted != 300 {
		t.Fatalf("unexpected traffic stats: %+v", stats)
	}
	if stats.MonthTests != 3 || stats.TotalTests != 3 {
		t.Fatalf("unexpected monthly/total test stats: %+v", stats)
	}
	if stats.MonthDownloaded != 300 || stats.TotalDownloaded != 300 {
		t.Fatalf("unexpected monthly/total traffic stats: %+v", stats)
	}
}

func TestRecordJobErrorClassifiesFailures(t *testing.T) {
	s := testServer(t)
	for _, code := range []string{"probe_failed", "resolve_failed", "fetch_failed", "xray_failed", "timeout", "usage_unchanged", "surprise", "cancelled"} {
		s.recordJobError(code)
	}
	s.mu.Lock()
	stats := s.live
	s.mu.Unlock()
	failed := stats.FailToday
	if failed.Total != 7 || failed.Connect != 2 || failed.Fetch != 1 || failed.Core != 1 || failed.Timeout != 1 || failed.Usage != 1 || failed.Other != 1 {
		t.Fatalf("unexpected failure stats: %+v", failed)
	}
	if stats.FailMonth != failed || stats.FailTotal != failed {
		t.Fatalf("unexpected monthly/total failures: %+v", stats)
	}
}

func TestSplitTargetHost(t *testing.T) {
	cases := []struct{ input, host, port string }{
		{"example.com:443", "example.com", "443"},
		{"example.com", "example.com", ""},
		{"[2001:db8::1]:8443", "2001:db8::1", "8443"},
		{"2001:db8::1", "2001:db8::1", ""},
	}
	for _, item := range cases {
		host, port := splitTargetHost(item.input)
		if host != item.host || port != item.port {
			t.Fatalf("splitTargetHost(%q) = (%q, %q), want (%q, %q)", item.input, host, port, item.host, item.port)
		}
	}
}

func TestTargetValidation(t *testing.T) {
	valid := []string{"example.com", "1.2.3.4", "2001:db8::1", "sub.example.co.uk"}
	invalid := []string{"", "localhost", "bad host", "exa mple.com", "../etc"}
	for _, host := range valid {
		if !validTargetHost(host) {
			t.Fatalf("expected %q to be valid", host)
		}
	}
	for _, host := range invalid {
		if validTargetHost(host) {
			t.Fatalf("expected %q to be invalid", host)
		}
	}
}

func TestGeoRejectsInvalidTarget(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/api/geo?target=localhost", nil)
	response := httptest.NewRecorder()
	s.geoCheck(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
}

func TestPingBlocksPrivateTarget(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/api/ping?target=127.0.0.1:80", nil)
	response := httptest.NewRecorder()
	s.pingCheck(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "blocked_target" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestReportSendsTelegramMessage(t *testing.T) {
	var got struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	var path string
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode telegram payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer telegram.Close()
	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL
	payload := `{"description":"ratio was wrong","sub_url":"https://example.com/sub","result":{"ratio":1.5,"downloaded_bytes":5242880,"delta_total":7864320,"configs_tested":2,"core":"latest","xray_version":"v1.8.24"}}`
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(payload))
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	if path != "/bottest-token/sendMessage" {
		t.Fatalf("unexpected telegram path: %s", path)
	}
	if got.ChatID != "12345" {
		t.Fatalf("unexpected chat id: %q", got.ChatID)
	}
	for _, want := range []string{"ratio was wrong", "https://example.com/sub", "1.50"} {
		if !strings.Contains(got.Text, want) {
			t.Fatalf("telegram text missing %q: %s", want, got.Text)
		}
	}
}

func TestReportDisabledWithoutTelegramConfig(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(`{"description":"x"}`))
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "report_disabled" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestReportRejectsEmptyDescription(t *testing.T) {
	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = "http://127.0.0.1:1"
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(`{"description":"  "}`))
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "invalid_input" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestReportRejectsLongDescription(t *testing.T) {
	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = "http://127.0.0.1:1"
	payload, _ := json.Marshal(map[string]any{"description": strings.Repeat("x", maxReportDescription+1)})
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(string(payload)))
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "description_too_long" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestReportRejectsInvalidSubURL(t *testing.T) {
	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = "http://127.0.0.1:1"
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(`{"description":"x","sub_url":"not-a-url"}`))
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data["code"] != "invalid_url" {
		t.Fatalf("unexpected code: %v", data["code"])
	}
}

func TestRecordProxyUse(t *testing.T) {
	s := &server{live: liveStats{Day: todayKey(), Month: monthKey()}}
	s.recordProxyUse()
	s.recordProxyUse()
	if s.live.ProxyToday != 2 || s.live.MonthProxy != 2 || s.live.TotalProxy != 2 {
		t.Fatalf("proxy counters: today=%d month=%d total=%d", s.live.ProxyToday, s.live.MonthProxy, s.live.TotalProxy)
	}
	s.mu.Lock()
	snapshot := s.liveSnapshotLocked()
	s.mu.Unlock()
	today, _ := snapshot["today"].(map[string]any)
	if today["proxy"] != 2 {
		t.Fatalf("snapshot proxy = %v", today["proxy"])
	}
}
