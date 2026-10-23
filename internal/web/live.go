package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	presenceWindow = 75 * time.Second
	geoCacheTTL    = 30 * time.Minute
	geoTimeout     = 6 * time.Second
	pingTimeout    = 5 * time.Second
)

type liveStats struct {
	Day        string `json:"day"`
	Tests      int    `json:"tests"`
	WithRatio  int    `json:"with_ratio"`
	Normal     int    `json:"normal"`
	Invalid    int    `json:"invalid"`
	Tunnel     int    `json:"tunnel"`
	NonTunnel  int    `json:"non_tunnel"`
	Unknown    int    `json:"unknown_tunnel"`
	Downloaded int64  `json:"downloaded"`
	Deducted   int64  `json:"deducted"`
}

type geoEntry struct {
	IP       string
	Country  string
	City     string
	Org      string
	IsTunnel bool
	Expires  time.Time
}

type presenceIn struct {
	ID string `json:"id"`
}

func todayKey() string {
	return time.Now().Format("2006-01-02")
}

func (s *server) loadStats() {
	if s.statsFile == "" {
		return
	}
	content, err := os.ReadFile(s.statsFile)
	if err != nil {
		return
	}
	var stats liveStats
	if err := json.Unmarshal(content, &stats); err != nil {
		return
	}
	if stats.Day != todayKey() {
		return
	}
	s.mu.Lock()
	s.live = stats
	s.mu.Unlock()
}

func (s *server) rollStatsLocked() {
	today := todayKey()
	if s.live.Day == today {
		return
	}
	s.live = liveStats{Day: today}
	s.saveStatsLocked()
}

func (s *server) saveStatsLocked() {
	if s.statsFile == "" {
		return
	}
	content, err := json.Marshal(s.live)
	if err != nil {
		return
	}
	temp := s.statsFile + ".tmp"
	if err := os.WriteFile(temp, content, 0o600); err != nil {
		return
	}
	_ = os.Rename(temp, s.statsFile)
}

func (s *server) recordResult(result *ratio.Result, tunnel *bool) {
	if result == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollStatsLocked()
	s.live.Tests++
	switch value := result.Ratio; {
	case value >= 1.06:
		s.live.WithRatio++
	case value >= 0.95:
		s.live.Normal++
	default:
		s.live.Invalid++
	}
	if tunnel == nil {
		s.live.Unknown++
	} else if *tunnel {
		s.live.Tunnel++
	} else {
		s.live.NonTunnel++
	}
	if result.DownloadedBytes > 0 {
		s.live.Downloaded += result.DownloadedBytes
	}
	if result.DeltaTotal > 0 {
		s.live.Deducted += result.DeltaTotal
	}
	s.saveStatsLocked()
}

func (s *server) liveSnapshotLocked() map[string]any {
	now := nowSeconds()
	online := 0
	for id, seen := range s.presence {
		if now-seen > presenceWindow.Seconds() {
			delete(s.presence, id)
			continue
		}
		online++
	}
	s.rollStatsLocked()
	return map[string]any{
		"online": online,
		"today": map[string]any{
			"date":           s.live.Day,
			"tests":          s.live.Tests,
			"with_ratio":     s.live.WithRatio,
			"normal":         s.live.Normal,
			"invalid":        s.live.Invalid,
			"tunnel":         s.live.Tunnel,
			"non_tunnel":     s.live.NonTunnel,
			"unknown_tunnel": s.live.Unknown,
			"downloaded":     s.live.Downloaded,
			"deducted":       s.live.Deducted,
		},
	}
}

func (s *server) presenceHandler(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.rateLimit("presence", ip, 3000) {
		apiError(w, "rate_limited", "Too many requests.", http.StatusTooManyRequests)
		return
	}
	var payload presenceIn
	if err := decodeJSON(w, r, &payload); err != nil || !validClientID(payload.ID) {
		apiError(w, "invalid_input", "Invalid client id.", http.StatusUnprocessableEntity)
		return
	}
	s.mu.Lock()
	if s.presence == nil {
		s.presence = map[string]float64{}
	}
	s.presence[payload.ID] = nowSeconds()
	view := s.liveSnapshotLocked()
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, view)
}

func validClientID(id string) bool {
	if len(id) < 8 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

func splitTargetHost(target string) (string, string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", ""
	}
	if host, port, err := net.SplitHostPort(target); err == nil {
		return strings.Trim(host, "[]"), port
	}
	return strings.Trim(target, "[]"), ""
}

func validTargetHost(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}
	for _, r := range host {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '.', r == ':', r == '_':
		default:
			return false
		}
	}
	if net.ParseIP(host) == nil && !strings.Contains(host, ".") {
		return false
	}
	return true
}

func numericPort(port string) bool {
	if port == "" || len(port) > 5 {
		return false
	}
	for _, r := range port {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (s *server) geoCheck(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.rateLimit("geo", ip, 400) {
		apiError(w, "rate_limited", "Too many lookups. Try again later.", http.StatusTooManyRequests)
		return
	}
	host, _ := splitTargetHost(r.URL.Query().Get("target"))
	if !validTargetHost(host) {
		apiError(w, "invalid_target", "Invalid target host.", http.StatusBadRequest)
		return
	}
	entry, err := s.lookupGeo(host)
	if err != nil {
		apiError(w, "geo_failed", "Could not check the server location.", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"host": host, "ip": entry.IP, "country": entry.Country,
		"city": entry.City, "org": entry.Org,
		"iran": entry.IsTunnel, "tunnel": entry.IsTunnel,
	})
}

func (s *server) lookupGeo(host string) (geoEntry, error) {
	s.geoMu.Lock()
	if entry, ok := s.geoCache[host]; ok && time.Now().Before(entry.Expires) {
		s.geoMu.Unlock()
		return entry, nil
	}
	s.geoMu.Unlock()

	ip, err := resolveHostIP(host)
	if err != nil {
		return geoEntry{}, err
	}
	entry := geoEntry{IP: ip, Expires: time.Now().Add(geoCacheTTL)}
	if info, err := queryIPInfo(ip); err == nil {
		entry.Country, entry.City, entry.Org = info.Country, info.City, info.Org
	} else if info, err := queryIPWho(ip); err == nil {
		entry.Country, entry.City, entry.Org = info.Country, info.City, info.Org
	} else {
		return geoEntry{}, fmt.Errorf("lookup failed for %s", host)
	}
	entry.Country = strings.ToUpper(strings.TrimSpace(entry.Country))
	if entry.Country == "" {
		return geoEntry{}, fmt.Errorf("no country for %s", host)
	}
	entry.IsTunnel = entry.Country == "IR"

	s.geoMu.Lock()
	if s.geoCache == nil {
		s.geoCache = map[string]geoEntry{}
	}
	s.geoCache[host] = entry
	s.geoMu.Unlock()
	return entry, nil
}

func resolveHostIP(host string) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), geoTimeout)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return "", fmt.Errorf("could not resolve %s", host)
	}
	for _, item := range addresses {
		if item.IP.To4() != nil {
			return item.IP.String(), nil
		}
	}
	return addresses[0].IP.String(), nil
}

type geoInfo struct {
	Country string
	City    string
	Org     string
}

var geoClient = &http.Client{Timeout: geoTimeout}

func queryIPInfo(ip string) (geoInfo, error) {
	request, err := http.NewRequest(http.MethodGet, "https://ipinfo.io/"+ip+"/json", nil)
	if err != nil {
		return geoInfo{}, err
	}
	request.Header.Set("User-Agent", ratio.UA)
	request.Header.Set("Accept", "application/json")
	response, err := geoClient.Do(request)
	if err != nil {
		return geoInfo{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return geoInfo{}, fmt.Errorf("ipinfo status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return geoInfo{}, err
	}
	var payload struct {
		Country string `json:"country"`
		City    string `json:"city"`
		Org     string `json:"org"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return geoInfo{}, err
	}
	if payload.Country == "" {
		return geoInfo{}, fmt.Errorf("ipinfo returned no country")
	}
	return geoInfo{Country: payload.Country, City: payload.City, Org: payload.Org}, nil
}

func queryIPWho(ip string) (geoInfo, error) {
	request, err := http.NewRequest(http.MethodGet, "https://ipwho.is/"+ip, nil)
	if err != nil {
		return geoInfo{}, err
	}
	request.Header.Set("User-Agent", ratio.UA)
	request.Header.Set("Accept", "application/json")
	response, err := geoClient.Do(request)
	if err != nil {
		return geoInfo{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return geoInfo{}, fmt.Errorf("ipwho status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return geoInfo{}, err
	}
	var payload struct {
		Success     bool   `json:"success"`
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		City        string `json:"city"`
		Connection  struct {
			ISP string `json:"isp"`
			Org string `json:"org"`
		} `json:"connection"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return geoInfo{}, err
	}
	if !payload.Success || payload.CountryCode == "" {
		return geoInfo{}, fmt.Errorf("ipwho returned no country")
	}
	org := payload.Connection.ISP
	if org == "" {
		org = payload.Connection.Org
	}
	return geoInfo{Country: payload.CountryCode, City: payload.City, Org: org}, nil
}

func (s *server) pingCheck(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.rateLimit("ping", ip, 400) {
		apiError(w, "rate_limited", "Too many checks. Try again later.", http.StatusTooManyRequests)
		return
	}
	host, port := splitTargetHost(r.URL.Query().Get("target"))
	if port == "" {
		port = "443"
	}
	if !validTargetHost(host) || !numericPort(port) {
		apiError(w, "invalid_target", "Invalid target host.", http.StatusBadRequest)
		return
	}
	if !ratio.IsPublicHost(host) {
		apiError(w, "blocked_target", "The target must be publicly reachable.", http.StatusBadRequest)
		return
	}
	address := net.JoinHostPort(host, port)
	dialer := net.Dialer{Timeout: pingTimeout}
	start := time.Now()
	connection, err := dialer.DialContext(r.Context(), "tcp", address)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"host": host, "port": port, "ok": false, "latency_ms": latency,
		})
		return
	}
	_ = connection.Close()
	writeJSON(w, http.StatusOK, map[string]any{
		"host": host, "port": port, "ok": true, "latency_ms": latency,
	})
}
