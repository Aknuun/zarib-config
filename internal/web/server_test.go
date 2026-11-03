package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	request.Header.Set("Content-Type", "application/json")
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
	request.Header.Set("Content-Type", "application/json")
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
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		s.presenceHandler(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(`{"id":"client-11111111"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.presenceHandler(response, request)
	var data map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data["online"].(float64) != 2 {
		t.Fatalf("unexpected online count: %v", data["online"])
	}
	for _, key := range []string{"h24", "d7", "d30", "overall", "today", "monthly"} {
		if _, ok := data[key].(map[string]any); !ok {
			t.Fatalf("missing %s stats: %+v", key, data)
		}
	}
}

func TestPresenceRejectsInvalidClientID(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(`{"id":"x"}`))
	request.Header.Set("Content-Type", "application/json")
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
	hour := *s.live.Hours[hourKey()]
	day := *s.live.Days[todayKey()]
	total := s.live.Total
	s.mu.Unlock()
	for name, stats := range map[string]statBucket{"hour": hour, "day": day, "total": total} {
		if stats.Tests != 3 || stats.WithRatio != 1 || stats.Normal != 1 || stats.Invalid != 1 {
			t.Fatalf("unexpected %s ratio stats: %+v", name, stats)
		}
		if stats.Tunnel != 1 || stats.NonTunnel != 1 || stats.Unknown != 1 {
			t.Fatalf("unexpected %s tunnel stats: %+v", name, stats)
		}
		if stats.Downloaded != 300 || stats.Deducted != 300 {
			t.Fatalf("unexpected %s traffic stats: %+v", name, stats)
		}
	}
}

func TestRecordJobErrorClassifiesFailures(t *testing.T) {
	s := testServer(t)
	for _, code := range []string{"probe_failed", "resolve_failed", "fetch_failed", "xray_failed", "timeout", "usage_unchanged", "surprise", "cancelled"} {
		s.recordJobError(code)
	}
	s.mu.Lock()
	hour := *s.live.Hours[hourKey()]
	day := *s.live.Days[todayKey()]
	total := s.live.Total
	s.mu.Unlock()
	for name, stats := range map[string]statBucket{"hour": hour, "day": day, "total": total} {
		failed := stats.Failed
		if failed.Total != 7 || failed.Connect != 2 || failed.Fetch != 1 || failed.Core != 1 || failed.Timeout != 1 || failed.Usage != 1 || failed.Other != 1 {
			t.Fatalf("unexpected %s failure stats: %+v", name, failed)
		}
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
	request.Header.Set("Content-Type", "application/json")
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
	request.Header.Set("Content-Type", "application/json")
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
	request.Header.Set("Content-Type", "application/json")
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
	request.Header.Set("Content-Type", "application/json")
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
	request.Header.Set("Content-Type", "application/json")
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
	s := testServer(t)
	s.recordProxyUse()
	s.recordProxyUse()
	s.mu.Lock()
	hour := s.live.Hours[hourKey()].Proxy
	day := s.live.Days[todayKey()].Proxy
	total := s.live.Total.Proxy
	s.mu.Unlock()
	if hour != 2 || day != 2 || total != 2 {
		t.Fatalf("proxy counters: hour=%d day=%d total=%d", hour, day, total)
	}
	s.mu.Lock()
	snapshot := s.liveSnapshotLocked()
	s.mu.Unlock()
	h24, _ := snapshot["h24"].(map[string]any)
	if h24["proxy"] != 2 {
		t.Fatalf("snapshot proxy = %v", h24["proxy"])
	}
}

func TestLoadStatsMigratesOldShape(t *testing.T) {
	dir := t.TempDir()
	s := testServer(t)
	s.statsFile = filepath.Join(dir, "stats.json")
	old := liveStats{
		Day: todayKey(), Month: monthKey(),
		Tests: 5, WithRatio: 2, Normal: 1, Invalid: 2,
		Tunnel: 1, NonTunnel: 2, Unknown: 2,
		Downloaded: 500, Deducted: 400, ProxyToday: 3,
		VisitorIPs: map[string]bool{"a": true},
		MonthTests: 20, MonthWithRatio: 8, MonthNormal: 4, MonthInvalid: 8,
		MonthTunnel: 4, MonthNonTunnel: 8, MonthUnknown: 8,
		MonthDownloaded: 2000, MonthDeducted: 1600, MonthProxy: 12,
		MonthVisitorIPs: map[string]bool{"a": true, "b": true},
		TotalTests:      50, TotalWithRatio: 20, TotalNormal: 10, TotalInvalid: 20,
		TotalTunnel: 10, TotalNonTunnel: 20, TotalUnknown: 20,
		TotalDownloaded: 5000, TotalDeducted: 4000, TotalProxy: 30,
		FailToday: failStats{Total: 1, Timeout: 1},
		FailMonth: failStats{Total: 4, Timeout: 2, Other: 2},
		FailTotal: failStats{Total: 10, Timeout: 5, Other: 5},
	}
	content, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.statsFile, content, 0o600); err != nil {
		t.Fatal(err)
	}

	s.loadStats()

	s.mu.Lock()
	totalTests := s.live.Total.Tests
	totalFailed := s.live.Total.Failed.Total
	day := s.live.Days[todayKey()]
	hour := s.live.Hours[hourKey()]
	totalUnique := s.live.TotalUnique
	s.mu.Unlock()
	if totalTests != 50 || totalFailed != 10 {
		t.Fatalf("migrated total: tests=%d failed=%d", totalTests, totalFailed)
	}
	if day == nil || day.Tests != 20 || day.Downloaded != 2000 || day.Proxy != 12 {
		t.Fatalf("migrated day bucket: %+v", day)
	}
	if hour == nil || hour.Tests != 5 || hour.Downloaded != 500 || hour.Proxy != 3 {
		t.Fatalf("migrated hour bucket: %+v", hour)
	}
	if totalUnique != 2 || len(day.IPs) != 2 || len(hour.IPs) != 2 {
		t.Fatalf("migrated IPs: total=%d day=%d hour=%d", totalUnique, len(day.IPs), len(hour.IPs))
	}

	written, err := os.ReadFile(s.statsFile)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(written, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["hours"] == nil || raw["days"] == nil {
		t.Fatalf("new shape was not saved: %s", written)
	}
	if _, ok := raw["day"]; ok {
		t.Fatalf("legacy fields were not cleared: %s", written)
	}

	// Loading the migrated file again must not double the seeded data.
	s.loadStats()
	s.mu.Lock()
	totalTests = s.live.Total.Tests
	day = s.live.Days[todayKey()]
	s.mu.Unlock()
	if totalTests != 50 || day == nil || day.Tests != 20 {
		t.Fatalf("migration is not idempotent: total=%d day=%+v", totalTests, day)
	}
}

func TestRangeAggregationAndPruning(t *testing.T) {
	s := testServer(t)
	now := time.Now()
	today := now.Format("2006-01-02")
	currentHour := now.Format("2006-01-02T15")
	recentHour := now.Add(-2 * time.Hour).Format("2006-01-02T15")
	oldHour := now.Add(-25 * time.Hour).Format("2006-01-02T15")
	recentDay := now.AddDate(0, 0, -2).Format("2006-01-02")
	weekAgoDay := now.AddDate(0, 0, -7).Format("2006-01-02")
	oldDay := now.AddDate(0, 0, -40).Format("2006-01-02")
	s.live = liveStats{
		Hours: map[string]*statBucket{
			currentHour: {Tests: 3, Proxy: 1, Failed: failStats{Total: 1, Timeout: 1}, IPs: map[string]bool{"a": true}},
			recentHour:  {Tests: 4, Proxy: 2, Failed: failStats{Total: 2, Connect: 2}, IPs: map[string]bool{"b": true}},
			oldHour:     {Tests: 100, IPs: map[string]bool{"c": true}},
		},
		Days: map[string]*statBucket{
			today:      {Tests: 5, IPs: map[string]bool{"a": true}},
			recentDay:  {Tests: 6, IPs: map[string]bool{"b": true}},
			weekAgoDay: {Tests: 7, IPs: map[string]bool{"c": true}},
			oldDay:     {Tests: 100, IPs: map[string]bool{"d": true}},
		},
		Total:       statBucket{Tests: 1000, IPs: map[string]bool{"a": true, "b": true}},
		TotalUnique: 2,
	}

	s.mu.Lock()
	snapshot := s.liveSnapshotLocked()
	hours := len(s.live.Hours)
	days := len(s.live.Days)
	s.mu.Unlock()

	h24, _ := snapshot["h24"].(map[string]any)
	d7, _ := snapshot["d7"].(map[string]any)
	d30, _ := snapshot["d30"].(map[string]any)
	overall, _ := snapshot["overall"].(map[string]any)
	if h24["tests"] != 7 || h24["proxy"] != 3 || h24["unique_ips"] != 2 {
		t.Fatalf("h24 = %+v", h24)
	}
	if failed := h24["failed"].(failStats); failed.Total != 3 || failed.Connect != 2 || failed.Timeout != 1 {
		t.Fatalf("h24 failed = %+v", failed)
	}
	if d7["tests"] != 11 || d7["unique_ips"] != 2 {
		t.Fatalf("d7 = %+v", d7)
	}
	if d30["tests"] != 18 || d30["unique_ips"] != 3 {
		t.Fatalf("d30 = %+v", d30)
	}
	if overall["tests"] != 1000 || overall["unique_ips"] != 2 {
		t.Fatalf("overall = %+v", overall)
	}
	if hours != 2 || days != 3 {
		t.Fatalf("pruning left hours=%d days=%d", hours, days)
	}
}

func TestRateLimit(t *testing.T) {
	s := &server{rates: map[string]map[string][]time.Time{}}
	if !s.rateLimit("test", "1.2.3.4", 2) {
		t.Fatal("first request should pass")
	}
	if !s.rateLimit("test", "1.2.3.4", 2) {
		t.Fatal("second request should pass")
	}
	if s.rateLimit("test", "1.2.3.4", 2) {
		t.Fatal("third request should be limited")
	}
}

func TestJobCountsIncludeRunning(t *testing.T) {
	s := &server{jobs: map[string]*job{
		"a": {Status: "running", IP: "1.1.1.1"},
		"b": {Status: "queued", IP: "1.1.1.1"},
		"c": {Status: "queued", IP: "2.2.2.2"},
	}}
	running, queued, perIP := s.jobCountsLocked()
	if running != 1 || queued != 2 || perIP["1.1.1.1"] != 2 {
		t.Fatalf("counts: running=%d queued=%d perIP=%v", running, queued, perIP)
	}
}
