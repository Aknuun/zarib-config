package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	maxFailureLog        = 120
	maxFailureLine       = 1000
	maxFailuresBytes     = 16 << 20
	maxAdminFailures     = 200
	maxFailureMessage    = 2000
	failureNotifyTTL     = time.Hour
	failureNotifyMax     = 2000
	failureNotifyPerHour = 30
	maxFailureLogSend    = 20
)

// storedFailure is a failed test or rejected subscription kept for later review.
type storedFailure struct {
	Time         string   `json:"time"`
	Source       string   `json:"source"`
	JobID        string   `json:"job_id,omitempty"`
	Mode         string   `json:"mode,omitempty"`
	Pick         int      `json:"pick,omitempty"`
	Picks        []int    `json:"picks,omitempty"`
	SizeMB       float64  `json:"size_mb,omitempty"`
	Core         string   `json:"core,omitempty"`
	SubURL       string   `json:"sub_url,omitempty"`
	ErrorCode    string   `json:"error_code,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
	Log          []string `json:"log,omitempty"`
	IP           string   `json:"ip,omitempty"`
	UserAgent    string   `json:"user_agent,omitempty"`
}

// silentIPSet parses ZARIB_SILENT_IPS: comma-separated client IPs whose
// failures are not stored in the admin log or sent to Telegram.
func silentIPSet(value string) map[string]bool {
	set := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			set[part] = true
		}
	}
	return set
}

func (s *server) isSilentIP(ip string) bool {
	return ip != "" && s.silentIPs[ip]
}

func (s *server) recordFailure(record storedFailure) {
	if s.failuresFile == "" || s.isSilentIP(record.IP) {
		return
	}
	if record.Time == "" {
		record.Time = time.Now().UTC().Format(time.RFC3339Nano)
	}
	record.SubURL = truncateText(strings.TrimSpace(record.SubURL), maxReportSubURL)
	record.ErrorCode = truncateText(strings.TrimSpace(record.ErrorCode), 80)
	record.ErrorMessage = truncateText(redactURLs(strings.TrimSpace(record.ErrorMessage)), maxFailureMessage)
	record.UserAgent = truncateText(record.UserAgent, 500)
	record.Log = sanitizeFailureLog(record.Log)
	if s.failuresFile != "" {
		if err := s.storeFailure(record); err != nil {
			fmt.Printf("failures: could not store: %v\n", err)
		}
	}
	s.notifyFailure(record)
}

// notifyFailure sends a real subscription failure to Telegram once per source
// IP and error code per failureNotifyTTL. The URL itself is not part of the key:
// a client that rotates subscription tokens must not flood the chat. Invalid or
// non-http inputs stay in the admin list only.
func (s *server) notifyFailure(record storedFailure) {
	if s.isSilentIP(record.IP) || s.telegramToken == "" || s.telegramChatID == "" || !isValidURL(record.SubURL) {
		return
	}
	key := record.IP + "\x00" + record.ErrorCode
	now := time.Now()
	s.notifyMu.Lock()
	if s.notified == nil {
		s.notified = map[string]time.Time{}
	}
	if last, ok := s.notified[key]; ok && now.Sub(last) < failureNotifyTTL {
		s.notifyMu.Unlock()
		return
	}
	if len(s.notified) >= failureNotifyMax {
		for oldKey, at := range s.notified {
			if now.Sub(at) > failureNotifyTTL {
				delete(s.notified, oldKey)
			}
		}
		if len(s.notified) >= failureNotifyMax {
			s.notified = map[string]time.Time{}
		}
	}
	s.notified[key] = now
	s.notifyMu.Unlock()
	if !s.rateLimit("failure-notify", "global", failureNotifyPerHour) {
		return
	}
	text := buildFailureMessage(record)
	go func() {
		if err := s.sendTelegram(text); err != nil {
			fmt.Printf("failures: could not send telegram: %v\n", err)
		}
	}()
}

func buildFailureMessage(record storedFailure) string {
	var b strings.Builder
	b.WriteString("🚨 خطای تست — Zarib Config\n")
	fmt.Fprintf(&b, "نسخه: %s\n", ratio.Version)
	if record.Source == "job" {
		b.WriteString("منبع: تست\n")
	} else {
		b.WriteString("منبع: بررسی ساب\n")
	}
	fmt.Fprintf(&b, "کد خطا: %s\n", record.ErrorCode)
	b.WriteString("لینک ساب:\n")
	b.WriteString(record.SubURL)
	b.WriteString("\n")
	if record.ErrorMessage != "" {
		fmt.Fprintf(&b, "پیام: %s\n", record.ErrorMessage)
	}
	if record.Mode != "" {
		fmt.Fprintf(&b, "حالت: %s\n", record.Mode)
	}
	if record.Pick > 0 {
		fmt.Fprintf(&b, "کانفیگ: %d\n", record.Pick)
	}
	if len(record.Picks) > 0 {
		parts := make([]string, len(record.Picks))
		for index, pick := range record.Picks {
			parts[index] = fmt.Sprintf("%d", pick)
		}
		fmt.Fprintf(&b, "کانفیگ‌ها: %s\n", strings.Join(parts, ", "))
	}
	if record.SizeMB > 0 {
		fmt.Fprintf(&b, "حجم: %.0f MB\n", record.SizeMB)
	}
	if record.Core != "" {
		fmt.Fprintf(&b, "هسته: %s\n", record.Core)
	}
	if record.IP != "" {
		fmt.Fprintf(&b, "IP: %s\n", record.IP)
	}
	if record.UserAgent != "" {
		fmt.Fprintf(&b, "User-Agent: %s\n", record.UserAgent)
	}
	if len(record.Log) > 0 {
		b.WriteString("\nلاگ:\n")
		logs := record.Log
		if len(logs) > maxFailureLogSend {
			logs = logs[len(logs)-maxFailureLogSend:]
		}
		for _, line := range logs {
			b.WriteString(truncateText(line, 300))
			b.WriteString("\n")
		}
	}
	return truncateText(strings.TrimRight(b.String(), "\n"), 3800)
}

func sanitizeFailureLog(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	if len(lines) > maxFailureLog {
		lines = lines[len(lines)-maxFailureLog:]
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r\n")
		if line == "" {
			continue
		}
		out = append(out, truncateText(redactURLs(line), maxFailureLine))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// storeFailure appends one JSON line to failures.jsonl, syncing before returning.
// The file is rotated to failures.jsonl.1 once it reaches maxFailuresBytes.
func (s *server) storeFailure(record storedFailure) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	s.failuresMu.Lock()
	defer s.failuresMu.Unlock()
	if info, err := os.Stat(s.failuresFile); err == nil && info.Size()+int64(len(data)) > maxFailuresBytes {
		_ = os.Remove(s.failuresFile + ".1")
		if err := os.Rename(s.failuresFile, s.failuresFile+".1"); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(s.failuresFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
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

// readFailures returns up to limit failures, newest first, including the rotated file.
func (s *server) readFailures(limit int) ([]storedFailure, error) {
	if s.failuresFile == "" {
		return []storedFailure{}, nil
	}
	s.failuresMu.Lock()
	defer s.failuresMu.Unlock()
	var all []storedFailure
	for _, path := range []string{s.failuresFile + ".1", s.failuresFile} {
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
			var record storedFailure
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

func (s *server) adminFailures(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, true) {
		return
	}
	failures, err := s.readFailures(maxAdminFailures)
	if err != nil {
		apiError(w, "failures_failed", "Could not read failures.", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":        len(failures),
		"failures":     failures,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}
