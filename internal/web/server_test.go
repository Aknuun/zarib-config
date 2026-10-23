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
)

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
