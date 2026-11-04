package web

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// repeatIP is one row in the admin "repeated IPs" table.
type repeatIP struct {
	IP      string
	Count   int
	Last    string
	Codes   string
	Blocked bool
	Silent  bool
}

func (s *server) loadBlockedIPs() {
	if s.blockedFile == "" {
		return
	}
	data, err := os.ReadFile(s.blockedFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("blocked: could not read: %v\n", err)
		}
		return
	}
	set := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if ip := strings.TrimSpace(line); net.ParseIP(ip) != nil {
			set[ip] = true
		}
	}
	s.blockedMu.Lock()
	s.blockedIPs = set
	s.blockedMu.Unlock()
}

func (s *server) isBlockedIP(ip string) bool {
	if ip == "" {
		return false
	}
	s.blockedMu.Lock()
	defer s.blockedMu.Unlock()
	return s.blockedIPs[ip]
}

func (s *server) blockedSnapshot() map[string]bool {
	s.blockedMu.Lock()
	defer s.blockedMu.Unlock()
	out := make(map[string]bool, len(s.blockedIPs))
	for ip := range s.blockedIPs {
		out[ip] = true
	}
	return out
}

func (s *server) setBlockedIP(ip string, blocked bool) error {
	s.blockedMu.Lock()
	defer s.blockedMu.Unlock()
	if s.blockedIPs == nil {
		s.blockedIPs = map[string]bool{}
	}
	old, had := s.blockedIPs[ip]
	if blocked {
		s.blockedIPs[ip] = true
	} else {
		delete(s.blockedIPs, ip)
	}
	if err := s.saveBlockedIPsLocked(); err != nil {
		if had {
			s.blockedIPs[ip] = old
		} else {
			delete(s.blockedIPs, ip)
		}
		return err
	}
	return nil
}

func (s *server) saveBlockedIPsLocked() error {
	if s.blockedFile == "" {
		return nil
	}
	ips := make([]string, 0, len(s.blockedIPs))
	for ip := range s.blockedIPs {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	data := strings.Join(ips, "\n")
	if data != "" {
		data += "\n"
	}
	temp := s.blockedFile + ".tmp"
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, s.blockedFile)
}

// rejectBlockedIP writes a 403 and returns true when the client IP is blocked.
func (s *server) rejectBlockedIP(w http.ResponseWriter, r *http.Request) bool {
	if !s.isBlockedIP(clientIP(r)) {
		return false
	}
	apiError(w, "ip_blocked", "This IP is blocked by the admin.", http.StatusForbidden)
	return true
}

// sameOriginRequest rejects cross-site form posts. Browsers send Origin (or at
// least Referer) for a normal same-origin form, so no token is needed.
func sameOriginRequest(r *http.Request) bool {
	check := func(raw string) bool {
		parsed, err := url.Parse(raw)
		return err == nil && strings.EqualFold(parsed.Host, r.Host)
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		return check(origin)
	}
	if referer := r.Header.Get("Referer"); referer != "" {
		return check(referer)
	}
	return false
}

// repeatedIPs counts failures in the window and adds blocked or silent IPs so
// the admin can still see and unblock them when they have no recent failures.
func (s *server) repeatedIPs(failures []storedFailure, since time.Time) []repeatIP {
	type counter struct {
		count int
		last  time.Time
		codes map[string]bool
	}
	counts := map[string]*counter{}
	for _, failure := range failures {
		if failure.IP == "" {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, failure.Time)
		if err != nil || at.Before(since) {
			continue
		}
		item := counts[failure.IP]
		if item == nil {
			item = &counter{codes: map[string]bool{}}
			counts[failure.IP] = item
		}
		item.count++
		if at.After(item.last) {
			item.last = at
		}
		if failure.ErrorCode != "" {
			item.codes[failure.ErrorCode] = true
		}
	}
	blocked := s.blockedSnapshot()
	out := make([]repeatIP, 0, len(counts)+len(blocked))
	for ip, item := range counts {
		if item.count < 2 && !blocked[ip] && !s.silentIPs[ip] {
			continue
		}
		codes := make([]string, 0, len(item.codes))
		for code := range item.codes {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		last := ""
		if !item.last.IsZero() {
			last = item.last.UTC().Format(time.RFC3339Nano)
		}
		out = append(out, repeatIP{
			IP: ip, Count: item.count, Last: last, Codes: strings.Join(codes, ", "),
			Blocked: blocked[ip], Silent: s.silentIPs[ip],
		})
	}
	for ip := range blocked {
		if _, ok := counts[ip]; !ok {
			out = append(out, repeatIP{IP: ip, Blocked: true})
		}
	}
	for ip := range s.silentIPs {
		if _, ok := counts[ip]; !ok && !blocked[ip] {
			out = append(out, repeatIP{IP: ip, Silent: true})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Last > out[j].Last
	})
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}
