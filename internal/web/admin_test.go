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
	for _, path := range []string{"/admin", "/admin/api/stats", "/admin/api/reports"} {
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

func TestAdminDisabledWithoutCredentials(t *testing.T) {
	s := testServer(t)
	for _, path := range []string{"/admin", "/admin/api/stats", "/admin/api/reports"} {
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

func TestReportPersistsWithoutSubURL(t *testing.T) {
	telegram := okTelegram(t)
	defer telegram.Close()
	s := adminTestServer(t)
	s.telegramToken = "test-token"
	s.telegramChatID = "12345"
	s.telegramAPIBase = telegram.URL

	payload := `{"description":"bad ratio","sub_url":"https://example.com/private-sub","view":"error",` +
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
	if strings.Contains(string(content), "private-sub") {
		t.Fatalf("reports file contains the subscription URL: %s", content)
	}
	var record storedReport
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &record); err != nil {
		t.Fatalf("stored report is not valid JSON: %v", err)
	}
	if record.Description != "bad ratio" || record.View != "error" || record.ErrorCode != "xray_failed" {
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
	if _, ok := raw["sub_url"]; ok {
		t.Fatalf("stored report has a sub_url field: %s", content)
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

func TestAdminReportsHideSubURL(t *testing.T) {
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
	if strings.Contains(body, "private-sub") {
		t.Fatalf("admin response contains the subscription URL: %s", body)
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
}

func TestAdminPageRendersStatsAndReports(t *testing.T) {
	s := adminTestServer(t)
	if err := s.storeReport(storedReport{
		Time:         time.Now().UTC().Format(time.RFC3339Nano),
		Description:  "page report",
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
	for _, want := range []string{"پنل مدیریت ضریب", "۲۴ ساعت اخیر", "۷ روز اخیر", "۳۰ روز اخیر", "page report", "timeout", "line one", "core failed", "1.50"} {
		if !strings.Contains(body, want) {
			t.Fatalf("admin page missing %q", want)
		}
	}
	if strings.Contains(body, "sub_url") {
		t.Fatal("admin page exposes a sub_url field")
	}
}
