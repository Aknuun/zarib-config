package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	maxFailureLog     = 120
	maxFailureLine    = 1000
	maxFailuresBytes  = 16 << 20
	maxAdminFailures  = 200
	maxFailureMessage = 2000
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

func (s *server) recordFailure(record storedFailure) {
	if s.failuresFile == "" {
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
	if err := s.storeFailure(record); err != nil {
		fmt.Printf("failures: could not store: %v\n", err)
	}
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
