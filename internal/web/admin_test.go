package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func adminTestServer(t *testing.T) *server {
	t.Helper()
	s := testServer(t)
	s.adminUser = "admin"
	s.adminPassword = "test-password"
	s.reportsFile = filepath.Join(t.TempDir(), "reports.jsonl")
	return s
}

func okTelegram(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

func TestAdminEndpointsRequireBasicAuth(t *testing.T) {
	s := adminTestServer(t)
	for _, path := range []string{"/admin", "/admin/api/stats", "/admin/api/reports", "/admin/api/failures"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		s.route(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s without auth: status %d", path, response.Code)
		}
		if response.Header().Get("WWW-Authenticate") == "" {
			t.Fatalf("%s without auth: missing WWW-Authenticate", path)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
	request.SetBasicAuth("admin", "wrong")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: status %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
	request.SetBasicAuth("admin", "test-password")
	response = httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("valid auth: status %d: %s", response.Code, response.Body.String())
	}
	for _, key := range []string{"h24", "d7", "d30", "overall", "version"} {
		if !strings.Contains(response.Body.String(), `"`+key+`"`) {
			t.Fatalf("stats response missing %q: %s", key, response.Body.String())
		}
	}
}

func TestAdminRepeatedIPsAndBlock(t *testing.T) {
	s := adminTestServer(t)
	now := time.Now().UTC()
	for i := 0; i < 2; i++ {
		if err := s.storeFailure(storedFailure{
			Time: now.Add(-time.Duration(i) * time.Minute).Format(time.RFC3339Nano),
			IP:   "203.0.113.5", ErrorCode: "fetch_failed", SubURL: "https://example.com/sub",
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.storeFailure(storedFailure{
		Time: now.Format(time.RFC3339Nano), IP: "198.51.100.7", ErrorCode: "timeout",
	}); err != nil {
		t.Fatal(err)
	}

	view := s.adminView()
	var repeated *repeatIP
	for index := range view.RepeatIPs {
		if view.RepeatIPs[index].IP == "203.0.113.5" {
			repeated = &view.RepeatIPs[index]
		}
		if view.RepeatIPs[index].IP == "198.51.100.7" {
			t.Fatalf("a single failure should not be listed")
		}
	}
	if repeated == nil || repeated.Count != 2 || repeated.Codes != "fetch_failed" {
		t.Fatalf("unexpected repeated IP: %#v", repeated)
	}

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "IPهای پرتکرار") {
		t.Fatalf("admin page missing repeated IPs: status %d", response.Code)
	}

	block := func(action, origin string) *httptest.ResponseRecorder {
		form := "ip=203.0.113.5&action=" + action
		request := httptest.NewRequest(http.MethodPost, "/admin/api/block", strings.NewReader(form))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Origin", origin)
		request.SetBasicAuth("admin", "test-password")
		response := httptest.NewRecorder()
		s.route(response, request)
		return response
	}
	if response := block("block", "http://evil.example"); response.Code != http.StatusForbidden {
		t.Fatalf("cross-origin block: status %d", response.Code)
	}
	if s.isBlockedIP("203.0.113.5") {
		t.Fatal("cross-origin request blocked the IP")
	}
	if response := block("block", "http://example.com"); response.Code != http.StatusSeeOther {
		t.Fatalf("block: status %d", response.Code)
	}
	if !s.isBlockedIP("203.0.113.5") {
		t.Fatal("IP was not blocked")
	}
	data, err := os.ReadFile(s.blockedFile)
	if err != nil || !strings.Contains(string(data), "203.0.113.5") {
		t.Fatalf("blocked file: %q, %v", data, err)
	}

	before, _ := s.readFailures(0)
	request = httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"sub_url":"https://example.com/sub"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Real-IP", "203.0.113.5")
	response = httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "ip_blocked") {
		t.Fatalf("blocked inspect: status %d: %s", response.Code, response.Body.String())
	}
	after, _ := s.readFailures(0)
	if len(after) != len(before) {
		t.Fatalf("blocked request was logged: %d -> %d", len(before), len(after))
	}

	if response := block("unblock", "http://example.com"); response.Code != http.StatusSeeOther {
		t.Fatalf("unblock: status %d", response.Code)
	}
	if s.isBlockedIP("203.0.113.5") {
		t.Fatal("IP was not unblocked")
	}

	if err := s.setBlockedIP("203.0.113.5", true); err != nil {
		t.Fatal(err)
	}
	reloaded := &server{blockedFile: s.blockedFile, blockedIPs: map[string]bool{}}
	reloaded.loadBlockedIPs()
	if !reloaded.isBlockedIP("203.0.113.5") {
		t.Fatal("blocked IP did not persist")
	}
}

func TestAdminDisabledWithoutCredentials(t *testing.T) {
	s := testServer(t)
	for _, path := range []string{"/admin", "/admin/api/stats", "/admin/api/reports", "/admin/api/failures"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		s.route(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s without credentials: status %d", path, response.Code)
		}
	}
}

func TestAdminAuthRateLimit(t *testing.T) {
	s := adminTestServer(t)
	for attempt := 0; attempt < maxAdminAuthPerHour; attempt++ {
		request := httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
		request.SetBasicAuth("admin", "wrong")
		response := httptest.NewRecorder()
		s.route(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status %d", attempt, response.Code)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
	request.SetBasicAuth("admin", "wrong")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limited attempt: status %d", response.Code)
	}
	// Valid credentials still work after the failed-attempt limit is hit.
	request = httptest.NewRequest(http.MethodGet, "/admin/api/stats", nil)
	request.SetBasicAuth("admin", "test-password")
	response = httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("valid auth after limit: status %d", response.Code)
	}
}

func TestReportPersistsSubURL(t *testing.T) {
	telegram := okTelegram(t)
	defer telegram.Close()
	s := adminTestServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL

	payload := `{"description":"bad ratio https://example.com/private-sub","sub_url":"https://example.com/private-sub","view":"error",` +
		`"error_code":"xray_failed","error_message":"core failed","error_log":["line one","line two"],` +
		`"meta":{"page":"/","platform":"Linux"},"result":{"ratio":1.5,"downloaded_bytes":1024,"delta_total":1536}}`
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "test-agent")
	request.Header.Set("Accept-Language", "fa-IR")
	request.Header.Set("Referer", "https://zarib.example/")
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}

	content, err := os.ReadFile(s.reportsFile)
	if err != nil {
		t.Fatal(err)
	}
	var record storedReport
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &record); err != nil {
		t.Fatalf("stored report is not valid JSON: %v", err)
	}
	if record.SubURL != "https://example.com/private-sub" {
		t.Fatalf("subscription URL was not stored: %+v", record)
	}
	// Other free-text fields are still redacted.
	if strings.Contains(record.Description, "private-sub") || !strings.Contains(record.Description, "<link>") {
		t.Fatalf("description was not redacted: %q", record.Description)
	}
	if record.View != "error" || record.ErrorCode != "xray_failed" {
		t.Fatalf("unexpected stored report: %+v", record)
	}
	if record.ErrorMessage != "core failed" || len(record.ErrorLog) != 2 {
		t.Fatalf("error details were not stored: %+v", record)
	}
	if record.IP == "" || record.UserAgent != "test-agent" || record.AcceptLanguage != "fa-IR" || record.Referer != "https://zarib.example/" {
		t.Fatalf("request details were not stored: %+v", record)
	}
	if record.Result == nil || record.Result.Ratio != 1.5 || record.Result.DownloadedBytes != 1024 {
		t.Fatalf("result summary was not stored: %+v", record.Result)
	}
	if record.Meta["platform"] != "Linux" || record.Meta["page"] != "/" {
		t.Fatalf("client meta was not stored: %+v", record.Meta)
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &raw); err != nil {
		t.Fatal(err)
	}
	if raw["sub_url"] != "https://example.com/private-sub" {
		t.Fatalf("stored report has no sub_url field: %s", content)
	}
	info, err := os.Stat(s.reportsFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("reports file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestReportAcceptsLargeErrorLog(t *testing.T) {
	telegram := okTelegram(t)
	defer telegram.Close()
	s := adminTestServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL

	lines := make([]string, 80)
	for index := range lines {
		lines[index] = strings.Repeat("x", 500)
	}
	payload, err := json.Marshal(map[string]any{"description": "large log", "error_log": lines})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	reports, err := s.readReports(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || len(reports[0].ErrorLog) != maxReportErrorLog {
		t.Fatalf("stored error log: %d reports, %d lines", len(reports), len(reports[0].ErrorLog))
	}
}

func TestAdminReportsShowSubURL(t *testing.T) {
	telegram := okTelegram(t)
	defer telegram.Close()
	s := adminTestServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL

	payload := `{"description":"bad ratio","sub_url":"https://example.com/private-sub","error_code":"timeout"}`
	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.report(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected report status %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/admin/api/reports", nil)
	request.SetBasicAuth("admin", "test-password")
	response = httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected admin status %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "private-sub") {
		t.Fatalf("admin response is missing the subscription URL: %s", body)
	}
	if !strings.Contains(body, "bad ratio") || !strings.Contains(body, "timeout") {
		t.Fatalf("admin response is missing the report: %s", body)
	}
	var data struct {
		Count   int            `json:"count"`
		Reports []storedReport `json:"reports"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data.Count != 1 || len(data.Reports) != 1 {
		t.Fatalf("unexpected report count: %+v", data)
	}
	if data.Reports[0].SubURL != "https://example.com/private-sub" {
		t.Fatalf("admin report is missing sub_url: %+v", data.Reports[0])
	}
}

func TestAdminPageRendersStatsAndReports(t *testing.T) {
	s := adminTestServer(t)
	if err := s.storeReport(storedReport{
		Time:         time.Now().UTC().Format(time.RFC3339Nano),
		Description:  "page report",
		SubURL:       "https://example.com/page-sub",
		ErrorCode:    "timeout",
		ErrorMessage: "core failed",
		ErrorLog:     []string{"line one", "line two"},
		Result: &reportResultIn{
			Ratio: 1.5, DownloadedBytes: 1024, DeltaTotal: 1536,
			ConfigsTested: 2, Core: "latest", XrayVersion: "v1.8.24", TestedAt: "now",
		},
		Meta: map[string]string{"page": "/", "platform": "Linux"},
	}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("unexpected content type: %s", contentType)
	}
	body := response.Body.String()
	for _, want := range []string{"پنل مدیریت ضریب", "۲۴ ساعت اخیر", "۷ روز اخیر", "۳۰ روز اخیر", "page report", "timeout", "line one", "core failed", "1.50", "https://example.com/page-sub"} {
		if !strings.Contains(body, want) {
			t.Fatalf("admin page missing %q", want)
		}
	}
}

func TestInspectFailureIsStored(t *testing.T) {
	s := testServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"sub_url":"not-a-url"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	failures, err := s.readFailures(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 1 {
		t.Fatalf("stored failures = %d, want 1", len(failures))
	}
	if failures[0].Source != "inspect" || failures[0].ErrorCode != "invalid_url" || failures[0].SubURL != "not-a-url" {
		t.Fatalf("unexpected failure: %+v", failures[0])
	}
	info, err := os.Stat(s.failuresFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("failures file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestSilentIPSkipsFailures(t *testing.T) {
	s := testServer(t)
	s.silentIPs = map[string]bool{"162.217.248.77": true}
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"sub_url":"not-a-url"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Real-IP", "162.217.248.77")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	failures, err := s.readFailures(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("silent IP stored %d failures, want 0", len(failures))
	}
}

func TestDuplicateTestAndXrayIncompatibleNotifications(t *testing.T) {
	received := make(chan string, 2)
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			received <- body.Text
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegram.Close()

	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL

	s.recordFailure(storedFailure{
		Source: "job", SubURL: "https://example.com/duplicate", ErrorCode: "duplicate_test", IP: "203.0.113.9",
	})
	failures, err := s.readFailures(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("duplicate test stored %d failures, want 0", len(failures))
	}

	s.recordFailure(storedFailure{
		Source: "job", SubURL: "https://example.com/incompatible", ErrorCode: "xray_incompatible", IP: "203.0.113.9",
	})
	failures, err = s.readFailures(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 1 || failures[0].ErrorCode != "xray_incompatible" {
		t.Fatalf("incompatible failure was not stored: %+v", failures)
	}

	select {
	case text := <-received:
		t.Fatalf("unexpected telegram notification: %s", text)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestAdminFailuresEndpointAndPage(t *testing.T) {
	s := adminTestServer(t)
	if err := s.storeFailure(storedFailure{
		Time:         time.Now().UTC().Format(time.RFC3339Nano),
		Source:       "job",
		JobID:        "job-123",
		Mode:         "single",
		Pick:         2,
		SizeMB:       10,
		Core:         "latest",
		SubURL:       "https://example.com/fail-sub",
		ErrorCode:    "xray_failed",
		ErrorMessage: "core failed",
		Log:          []string{"probe log one", "probe log two"},
		IP:           "203.0.113.9",
		UserAgent:    "test-agent",
	}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/admin/api/failures", nil)
	request.SetBasicAuth("admin", "test-password")
	response := httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	var data struct {
		Count    int             `json:"count"`
		Failures []storedFailure `json:"failures"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data.Count != 1 || len(data.Failures) != 1 || data.Failures[0].ErrorCode != "xray_failed" {
		t.Fatalf("unexpected failures response: %s", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.SetBasicAuth("admin", "test-password")
	response = httptest.NewRecorder()
	s.route(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected page status %d: %s", response.Code, response.Body.String())
	}
	for _, want := range []string{"خطاهای تست", "xray_failed", "probe log one", "https://example.com/fail-sub", "job-123"} {
		if !strings.Contains(response.Body.String(), want) {
			t.Fatalf("admin page missing %q", want)
		}
	}
}

func TestFailureNotificationSendsSubURLAndLog(t *testing.T) {
	received := make(chan string, 2)
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			received <- body.Text
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegram.Close()

	s := testServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL
	failure := storedFailure{
		Source:       "job",
		SubURL:       "https://example.com/broken-sub",
		ErrorCode:    "xray_failed",
		ErrorMessage: "core failed",
		Log:          []string{"line one", "line two"},
		IP:           "203.0.113.9",
	}
	s.notifyFailure(failure)
	select {
	case text := <-received:
		for _, want := range []string{"https://example.com/broken-sub", "xray_failed", "core failed", "line one"} {
			if !strings.Contains(text, want) {
				t.Fatalf("notification missing %q: %s", want, text)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no telegram notification was sent")
	}

	// The same source and error code is not sent again immediately, even when
	// the subscription token in the URL changes.
	rotated := failure
	rotated.SubURL = "https://example.com/broken-sub-rotated"
	s.notifyFailure(rotated)
	select {
	case text := <-received:
		t.Fatalf("duplicate notification was sent: %s", text)
	case <-time.After(200 * time.Millisecond):
	}

	// Invalid inputs stay in the admin list and are not sent to Telegram.
	s.notifyFailure(storedFailure{Source: "inspect", SubURL: "not-a-url", ErrorCode: "invalid_url"})
	select {
	case text := <-received:
		t.Fatalf("invalid input was sent to Telegram: %s", text)
	case <-time.After(200 * time.Millisecond):
	}
}
