package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	maxReportDescription = 1500
	maxReportSubURL      = 1500
	maxReportPerHour     = 5
)

var telegramClient = &http.Client{Timeout: 12 * time.Second}

type reportIn struct {
	Description string          `json:"description"`
	SubURL      string          `json:"sub_url"`
	Result      *reportResultIn `json:"result"`
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
	if s.telegramToken == "" || s.telegramChatID == "" {
		apiError(w, "report_disabled", "Reporting is not configured.", http.StatusServiceUnavailable)
		return
	}
	if !s.rateLimit("report", clientIP(r), maxReportPerHour) {
		apiError(w, "rate_limited", "Too many reports. Try again later.", http.StatusTooManyRequests)
		return
	}
	var payload reportIn
	if err := decodeJSON(w, r, &payload); err != nil {
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
	text := buildReportMessage(description, subURL, payload.Result)
	if err := s.sendTelegram(text); err != nil {
		apiError(w, "report_failed", "Could not send the report.", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func buildReportMessage(description, subURL string, result *reportResultIn) string {
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
