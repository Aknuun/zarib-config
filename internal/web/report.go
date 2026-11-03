package web

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	maxReportDescription  = 1500
	maxReportSubURL       = 1500
	maxReportPerHour      = 5
	maxReportErrorMessage = 2000
	maxReportErrorLog     = 40
	maxReportErrorLine    = 1000
	maxReportBodyBytes    = 128 << 10
	maxReportsBytes       = 8 << 20
	maxAdminReports       = 200
)

// storedReport is what the admin panel sees.
type storedReport struct {
	Time           string            `json:"time"`
	Description    string            `json:"description"`
	SubURL         string            `json:"sub_url,omitempty"`
	View           string            `json:"view,omitempty"`
	ErrorCode      string            `json:"error_code,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	ErrorLog       []string          `json:"error_log,omitempty"`
	Result         *reportResultIn   `json:"result,omitempty"`
	IP             string            `json:"ip,omitempty"`
	UserAgent      string            `json:"user_agent,omitempty"`
	AcceptLanguage string            `json:"accept_language,omitempty"`
	Referer        string            `json:"referer,omitempty"`
	Meta           map[string]string `json:"meta,omitempty"`
}

var telegramClient = newTelegramClient()

func newTelegramClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if raw := strings.TrimSpace(os.Getenv("ZARIB_SUB_PROXY")); raw != "" {
		if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
			transport.Proxy = http.ProxyURL(parsed)
		}
	}
	return &http.Client{Timeout: 12 * time.Second, Transport: transport}
}

type reportIn struct {
	Description  string            `json:"description"`
	SubURL       string            `json:"sub_url"`
	Result       *reportResultIn   `json:"result"`
	Meta         map[string]string `json:"meta"`
	View         string            `json:"view"`
	ErrorCode    string            `json:"error_code"`
	ErrorMessage string            `json:"error_message"`
	ErrorLog     []string          `json:"error_log"`
}

type reportResultIn struct {
	Ratio           float64 `json:"ratio"`
	DownloadedBytes int64   `json:"downloaded_bytes"`
	DeltaTotal      int64   `json:"delta_total"`
	ConfigsTested   int     `json:"configs_tested"`
	Core            string  `json:"core"`
	XrayVersion     string  `json:"xray_version"`
	TestedAt        string  `json:"tested_at"`
}

func (s *server) report(w http.ResponseWriter, r *http.Request) {
	if !s.rateLimit("report", clientIP(r), maxReportPerHour) {
		apiError(w, "rate_limited", "Too many reports. Try again later.", http.StatusTooManyRequests)
		return
	}
	var payload reportIn
	if err := decodeJSONLimit(w, r, &payload, maxReportBodyBytes); err != nil {
		apiError(w, "invalid_input", "Invalid request.", http.StatusUnprocessableEntity)
		return
	}
	description := strings.TrimSpace(payload.Description)
	if description == "" {
		apiError(w, "invalid_input", "Describe the problem first.", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(description) > maxReportDescription {
		apiError(w, "description_too_long", "The description is too long.", http.StatusBadRequest)
		return
	}
	subURL := strings.TrimSpace(payload.SubURL)
	if subURL != "" {
		if len(subURL) > maxReportSubURL || !isValidURL(subURL) {
			apiError(w, "invalid_url", "The subscription URL is not valid.", http.StatusBadRequest)
			return
		}
	}
	record := storedReport{
		Time:           time.Now().UTC().Format(time.RFC3339Nano),
		Description:    redactURLs(description),
		SubURL:         subURL,
		View:           truncateText(strings.TrimSpace(payload.View), 64),
		ErrorCode:      truncateText(strings.TrimSpace(payload.ErrorCode), 80),
		ErrorMessage:   truncateText(redactURLs(strings.TrimSpace(payload.ErrorMessage)), maxReportErrorMessage),
		ErrorLog:       sanitizeReportLog(payload.ErrorLog),
		Result:         payload.Result,
		IP:             clientIP(r),
		UserAgent:      truncateText(r.UserAgent(), 500),
		AcceptLanguage: truncateText(r.Header.Get("Accept-Language"), 200),
		Referer:        truncateText(r.Header.Get("Referer"), 500),
		Meta:           sanitizeReportMeta(payload.Meta),
	}
	if err := s.storeReport(record); err != nil {
		fmt.Printf("report: could not store: %v\n", err)
	}
	if s.telegramToken == "" || s.telegramChatID == "" {
		apiError(w, "report_disabled", "Reporting is not configured.", http.StatusServiceUnavailable)
		return
	}
	text := buildReportMessage(description, subURL, payload.Result, reportDetails(r, payload.Meta))
	if err := s.sendTelegram(text); err != nil {
		apiError(w, "report_failed", "Could not send the report.", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func reportDetails(r *http.Request, meta map[string]string) []string {
	details := make([]string, 0, 16)
	add := func(label, value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			details = append(details, label+": "+truncateText(value, 300))
		}
	}
	add("IP", clientIP(r))
	add("مرورگر/سیستم", r.UserAgent())
	add("زبان مرورگر", r.Header.Get("Accept-Language"))
	add("Referer", r.Header.Get("Referer"))
	labels := []struct{ key, label string }{
		{"platform", "پلتفرم"},
		{"ua_platform", "پلتفرم UA"},
		{"ua_mobile", "موبایل"},
		{"language", "زبان"},
		{"timezone", "منطقه زمانی"},
		{"screen", "صفحه‌نمایش"},
		{"viewport", "پنجره"},
		{"pixel_ratio", "نسبت پیکسل"},
		{"touch", "لمسی"},
		{"cores", "هسته CPU"},
		{"memory", "حافظه GB"},
		{"page", "صفحه"},
		{"view", "بخش"},
	}
	for _, item := range labels {
		add(item.label, meta[item.key])
	}
	return details
}

func buildReportMessage(description, subURL string, result *reportResultIn, details []string) string {
	var b strings.Builder
	b.WriteString("🚨 گزارش خطا — Zarib Config\n")
	fmt.Fprintf(&b, "نسخه: %s\n\n", ratio.Version)
	b.WriteString("شرح مشکل:\n")
	b.WriteString(description)
	b.WriteString("\n\n")
	b.WriteString("لینک ساب برای بررسی (حداقل ۱۰۰ مگابایت):\n")
	if subURL == "" {
		b.WriteString("—")
	} else {
		b.WriteString(subURL)
	}
	if result != nil {
		b.WriteString("\n\nنتیجه تست:\n")
		fmt.Fprintf(&b, "• ضریب: %.2f×\n", result.Ratio)
		fmt.Fprintf(&b, "• دانلود واقعی: %s\n", humanBytes(result.DownloadedBytes))
		fmt.Fprintf(&b, "• کسرشده: %s\n", humanBytes(result.DeltaTotal))
		if result.ConfigsTested > 0 {
			fmt.Fprintf(&b, "• تعداد کانفیگ: %d\n", result.ConfigsTested)
		}
		if result.Core != "" {
			core := result.Core
			if result.XrayVersion != "" {
				core += " — " + truncateText(result.XrayVersion, 64)
			}
			fmt.Fprintf(&b, "• هسته: %s\n", core)
		}
		if result.TestedAt != "" {
			fmt.Fprintf(&b, "• زمان تست: %s\n", truncateText(result.TestedAt, 64))
		}
	}
	if len(details) > 0 {
		b.WriteString("\n\nمشخصات فرستنده:\n")
		b.WriteString(strings.Join(details, "\n"))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (s *server) sendTelegram(text string) error {
	endpoint := strings.TrimRight(s.telegramAPIBase, "/") + "/bot" + s.telegramToken + "/sendMessage"
	body, err := json.Marshal(map[string]any{
		"chat_id":                  s.telegramChatID,
		"text":                     text,
		"disable_web_page_preview": true,
	})
	if err != nil {
		return fmt.Errorf("encode report")
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telegram request")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := telegramClient.Do(request)
	if err != nil {
		return fmt.Errorf("telegram request failed")
	}
	defer response.Body.Close()
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&result)
	if response.StatusCode < 200 || response.StatusCode >= 300 || !result.OK {
		if result.Description != "" {
			return fmt.Errorf("telegram rejected the report: %s", truncateText(result.Description, 200))
		}
		return fmt.Errorf("telegram rejected the report: status %d", response.StatusCode)
	}
	return nil
}

func humanBytes(value int64) string {
	if value < 0 {
		return "—"
	}
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}
	div, exp := int64(unit), 0
	for n := value / unit; n >= unit && exp < 4; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(value)/float64(div), "KMGTP"[exp])
}

func truncateText(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}

func sanitizeReportLog(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	if len(lines) > maxReportErrorLog {
		lines = lines[len(lines)-maxReportErrorLog:]
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r\n")
		if line == "" {
			continue
		}
		out = append(out, truncateText(redactURLs(line), maxReportErrorLine))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sanitizeReportMeta(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}
	out := make(map[string]string, len(meta))
	for key, value := range meta {
		key = truncateText(strings.TrimSpace(key), 40)
		if key == "" {
			continue
		}
		out[key] = truncateText(redactURLs(strings.TrimSpace(value)), 500)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// storeReport appends one JSON line to reports.jsonl, syncing before returning.
// The file is rotated to reports.jsonl.1 once it reaches maxReportsBytes.
func (s *server) storeReport(record storedReport) error {
	if s.reportsFile == "" {
		return nil
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()
	if info, err := os.Stat(s.reportsFile); err == nil && info.Size()+int64(len(data)) > maxReportsBytes {
		_ = os.Remove(s.reportsFile + ".1")
		if err := os.Rename(s.reportsFile, s.reportsFile+".1"); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(s.reportsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
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
	return file.Close()
}

// readReports returns up to limit reports, newest first, including the rotated file.
func (s *server) readReports(limit int) ([]storedReport, error) {
	if s.reportsFile == "" {
		return []storedReport{}, nil
	}
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()
	var all []storedReport
	for _, path := range []string{s.reportsFile + ".1", s.reportsFile} {
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 4096), 256*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var record storedReport
			if json.Unmarshal([]byte(line), &record) == nil {
				all = append(all, record)
			}
		}
		err = scanner.Err()
		file.Close()
		if err != nil {
			return nil, err
		}
	}
	if limit > 0 && len(all) > limit {
		all = all[len(all)-limit:]
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all, nil
}
