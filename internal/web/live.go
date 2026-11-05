package web

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

type failStats struct {
	Total   int `json:"total,omitempty"`
	Connect int `json:"connect,omitempty"`
	Fetch   int `json:"fetch,omitempty"`
	Core    int `json:"core,omitempty"`
	Timeout int `json:"timeout,omitempty"`
	Usage   int `json:"usage,omitempty"`
	Other   int `json:"other,omitempty"`
}

type statBucket struct {
	Tests          int             `json:"tests,omitempty"`
	WithRatio      int             `json:"with_ratio,omitempty"`
	Normal         int             `json:"normal,omitempty"`
	Invalid        int             `json:"invalid,omitempty"`
	Tunnel         int             `json:"tunnel,omitempty"`
	NonTunnel      int             `json:"non_tunnel,omitempty"`
	Unknown        int             `json:"unknown_tunnel,omitempty"`
	Downloaded     int64           `json:"downloaded,omitempty"`
	Deducted       int64           `json:"deducted,omitempty"`
	Proxy          int             `json:"proxy,omitempty"`
	ClaimedTests   int             `json:"claimed_tests,omitempty"`
	ClaimedSum     float64         `json:"claimed_sum,omitempty"`
	ClaimedRealSum float64         `json:"claimed_real_sum,omitempty"`
	ClaimedDiffSum float64         `json:"claimed_diff_sum,omitempty"`
	Failed         failStats       `json:"failed,omitempty"`
	IPs            map[string]bool `json:"ips,omitempty"`
}

// maxUniqueIPs bounds the all-time set so stats.json cannot grow forever.
const maxUniqueIPs = 20000

type liveStats struct {
	Hours       map[string]*statBucket `json:"hours,omitempty"`
	Days        map[string]*statBucket `json:"days,omitempty"`
	Total       statBucket             `json:"total"`
	TotalUnique int                    `json:"total_unique,omitempty"`

	// Old flat shape, read only for migration.
	Day        string          `json:"day,omitempty"`
	Tests      int             `json:"tests,omitempty"`
	WithRatio  int             `json:"with_ratio,omitempty"`
	Normal     int             `json:"normal,omitempty"`
	Invalid    int             `json:"invalid,omitempty"`
	Tunnel     int             `json:"tunnel,omitempty"`
	NonTunnel  int             `json:"non_tunnel,omitempty"`
	Unknown    int             `json:"unknown_tunnel,omitempty"`
	Downloaded int64           `json:"downloaded,omitempty"`
	Deducted   int64           `json:"deducted,omitempty"`
	ProxyToday int             `json:"proxy_today,omitempty"`
	VisitorIPs map[string]bool `json:"visitor_ips,omitempty"`

	Month           string          `json:"month,omitempty"`
	MonthTests      int             `json:"month_tests,omitempty"`
	MonthWithRatio  int             `json:"month_with_ratio,omitempty"`
	MonthNormal     int             `json:"month_normal,omitempty"`
	MonthInvalid    int             `json:"month_invalid,omitempty"`
	MonthTunnel     int             `json:"month_tunnel,omitempty"`
	MonthNonTunnel  int             `json:"month_non_tunnel,omitempty"`
	MonthUnknown    int             `json:"month_unknown_tunnel,omitempty"`
	MonthDownloaded int64           `json:"month_downloaded,omitempty"`
	MonthDeducted   int64           `json:"month_deducted,omitempty"`
	MonthProxy      int             `json:"month_proxy,omitempty"`
	MonthVisitorIPs map[string]bool `json:"month_visitor_ips,omitempty"`

	TotalTests      int   `json:"total_tests,omitempty"`
	TotalWithRatio  int   `json:"total_with_ratio,omitempty"`
	TotalNormal     int   `json:"total_normal,omitempty"`
	TotalInvalid    int   `json:"total_invalid,omitempty"`
	TotalTunnel     int   `json:"total_tunnel,omitempty"`
	TotalNonTunnel  int   `json:"total_non_tunnel,omitempty"`
	TotalUnknown    int   `json:"total_unknown_tunnel,omitempty"`
	TotalDownloaded int64 `json:"total_downloaded,omitempty"`
	TotalDeducted   int64 `json:"total_deducted,omitempty"`
	TotalProxy      int   `json:"total_proxy,omitempty"`

	FailToday failStats `json:"fail_today,omitempty"`
	FailMonth failStats `json:"fail_month,omitempty"`
	FailTotal failStats `json:"fail_total,omitempty"`
}

func (b *statBucket) addIP(key string) bool {
	if b.IPs == nil {
		b.IPs = map[string]bool{}
	}
	if b.IPs[key] {
		return false
	}
	b.IPs[key] = true
	return true
}

func (b *statBucket) add(other *statBucket) {
	if other == nil {
		return
	}
	b.Tests += other.Tests
	b.WithRatio += other.WithRatio
	b.Normal += other.Normal
	b.Invalid += other.Invalid
	b.Tunnel += other.Tunnel
	b.NonTunnel += other.NonTunnel
	b.Unknown += other.Unknown
	b.Downloaded += other.Downloaded
	b.Deducted += other.Deducted
	b.Proxy += other.Proxy
	b.ClaimedTests += other.ClaimedTests
	b.ClaimedSum += other.ClaimedSum
	b.ClaimedRealSum += other.ClaimedRealSum
	b.ClaimedDiffSum += other.ClaimedDiffSum
	b.Failed.Total += other.Failed.Total
	b.Failed.Connect += other.Failed.Connect
	b.Failed.Fetch += other.Failed.Fetch
	b.Failed.Core += other.Failed.Core
	b.Failed.Timeout += other.Failed.Timeout
	b.Failed.Usage += other.Failed.Usage
	b.Failed.Other += other.Failed.Other
	for key := range other.IPs {
		b.addIP(key)
	}
}

func (b *statBucket) addResult(result *ratio.Result, tunnel *bool) {
	b.Tests++
	switch value := result.Ratio; {
	case value >= 1.06:
		b.WithRatio++
	case value >= 0.95:
		b.Normal++
	default:
		b.Invalid++
	}
	if tunnel == nil {
		b.Unknown++
	} else if *tunnel {
		b.Tunnel++
	} else {
		b.NonTunnel++
	}
	if result.DownloadedBytes > 0 {
		b.Downloaded += result.DownloadedBytes
	}
	if result.DeltaTotal > 0 {
		b.Deducted += result.DeltaTotal
	}
	if result.ClaimedRatio != nil {
		b.ClaimedTests++
		b.ClaimedSum += *result.ClaimedRatio
		b.ClaimedRealSum += result.Ratio
		b.ClaimedDiffSum += *result.ClaimedRatio - result.Ratio
	}
}

func (b *statBucket) addFail(code string) {
	b.Failed.Total++
	switch code {
	case "probe_failed", "resolve_failed", "blocked_target", "ipv6_unavailable":
		b.Failed.Connect++
	case "fetch_failed", "subscription_timeout", "no_configs", "pick_out_of_range":
		b.Failed.Fetch++
	case "xray_failed":
		b.Failed.Core++
	case "timeout":
		b.Failed.Timeout++
	case "usage_unchanged":
		b.Failed.Usage++
	default:
		b.Failed.Other++
	}
}

func (s *liveStats) clearLegacy() {
	s.Day = ""
	s.Tests, s.WithRatio, s.Normal, s.Invalid = 0, 0, 0, 0
	s.Tunnel, s.NonTunnel, s.Unknown = 0, 0, 0
	s.Downloaded, s.Deducted, s.ProxyToday = 0, 0, 0
	s.VisitorIPs = nil
	s.Month = ""
	s.MonthTests, s.MonthWithRatio, s.MonthNormal, s.MonthInvalid = 0, 0, 0, 0
	s.MonthTunnel, s.MonthNonTunnel, s.MonthUnknown = 0, 0, 0
	s.MonthDownloaded, s.MonthDeducted, s.MonthProxy = 0, 0, 0
	s.MonthVisitorIPs = nil
	s.TotalTests, s.TotalWithRatio, s.TotalNormal, s.TotalInvalid = 0, 0, 0, 0
	s.TotalTunnel, s.TotalNonTunnel, s.TotalUnknown = 0, 0, 0
	s.TotalDownloaded, s.TotalDeducted, s.TotalProxy = 0, 0, 0
	s.FailToday, s.FailMonth, s.FailTotal = failStats{}, failStats{}, failStats{}
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

func hourKey() string {
	return time.Now().Format("2006-01-02T15")
}

func monthKey() string {
	return time.Now().Format("2006-01")
}

func (s *server) loadStats() {
	if s.statsFile == "" {
		return
	}
	content, err := os.ReadFile(s.statsFile)
	if err == nil {
		var stats liveStats
		if json.Unmarshal(content, &stats) == nil {
			s.applyLoadedStats(stats)
			return
		}
		fmt.Printf("stats: could not parse %s; trying backup\n", s.statsFile)
	}
	content, err = os.ReadFile(s.statsFile + ".bak")
	if err != nil {
		return
	}
	var stats liveStats
	if err := json.Unmarshal(content, &stats); err != nil {
		fmt.Printf("stats: backup is unreadable: %v\n", err)
		return
	}
	s.applyLoadedStats(stats)
}

func (s *server) applyLoadedStats(stats liveStats) {
	stats, migrated := migrateStats(stats)
	s.mu.Lock()
	s.live = stats
	if migrated {
		s.saveStatsLocked()
	}
	s.rollStatsLocked()
	s.mu.Unlock()
}

// migrateStats converts the old day/month/total shape into hourly/daily buckets.
// Old data has no per-day breakdown, so the best available totals are seeded
// into the current day (d30) and current hour (h24).
func migrateStats(stats liveStats) (liveStats, bool) {
	if stats.Hours == nil {
		stats.Hours = map[string]*statBucket{}
	}
	if stats.Days == nil {
		stats.Days = map[string]*statBucket{}
	}
	if stats.Total.IPs == nil {
		stats.Total.IPs = map[string]bool{}
	}

	legacy := stats.Day != "" || stats.Month != "" || stats.Tests != 0 ||
		stats.MonthTests != 0 || stats.TotalTests != 0 || stats.ProxyToday != 0 ||
		stats.MonthProxy != 0 || stats.TotalProxy != 0 || len(stats.VisitorIPs) > 0 ||
		len(stats.MonthVisitorIPs) > 0 || stats.FailToday.Total != 0 ||
		stats.FailMonth.Total != 0 || stats.FailTotal.Total != 0
	if !legacy {
		if stats.TotalUnique < len(stats.Total.IPs) {
			stats.TotalUnique = len(stats.Total.IPs)
		}
		return stats, false
	}

	day := statBucket{
		Tests: stats.Tests, WithRatio: stats.WithRatio, Normal: stats.Normal,
		Invalid: stats.Invalid, Tunnel: stats.Tunnel, NonTunnel: stats.NonTunnel,
		Unknown: stats.Unknown, Downloaded: stats.Downloaded, Deducted: stats.Deducted,
		Proxy: stats.ProxyToday, Failed: stats.FailToday,
	}
	month := statBucket{
		Tests: stats.MonthTests, WithRatio: stats.MonthWithRatio, Normal: stats.MonthNormal,
		Invalid: stats.MonthInvalid, Tunnel: stats.MonthTunnel, NonTunnel: stats.MonthNonTunnel,
		Unknown: stats.MonthUnknown, Downloaded: stats.MonthDownloaded, Deducted: stats.MonthDeducted,
		Proxy: stats.MonthProxy, Failed: stats.FailMonth,
	}
	total := statBucket{
		Tests: stats.TotalTests, WithRatio: stats.TotalWithRatio, Normal: stats.TotalNormal,
		Invalid: stats.TotalInvalid, Tunnel: stats.TotalTunnel, NonTunnel: stats.TotalNonTunnel,
		Unknown: stats.TotalUnknown, Downloaded: stats.TotalDownloaded, Deducted: stats.TotalDeducted,
		Proxy: stats.TotalProxy, Failed: stats.FailTotal,
	}
	empty := func(b statBucket) bool {
		return b.Tests == 0 && b.Downloaded == 0 && b.Deducted == 0 && b.Proxy == 0 && b.Failed.Total == 0
	}
	if stats.Month == "" {
		month = day
	}
	if empty(total) {
		if !empty(month) {
			total = month
		} else {
			total = day
		}
	}
	if empty(month) {
		month = day
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	hour := now.Format("2006-01-02T15")
	seedDay := month
	if day.Tests > seedDay.Tests {
		seedDay = day
	}
	if !empty(seedDay) {
		if stats.Days[today] == nil {
			stats.Days[today] = &statBucket{}
		}
		stats.Days[today].add(&seedDay)
	}
	if !empty(day) {
		if stats.Hours[hour] == nil {
			stats.Hours[hour] = &statBucket{}
		}
		stats.Hours[hour].add(&day)
	}
	stats.Total.add(&total)

	if stats.Days[today] == nil {
		stats.Days[today] = &statBucket{}
	}
	if stats.Hours[hour] == nil {
		stats.Hours[hour] = &statBucket{}
	}
	for key := range stats.VisitorIPs {
		stats.Days[today].addIP(key)
		stats.Hours[hour].addIP(key)
		stats.Total.addIP(key)
	}
	for key := range stats.MonthVisitorIPs {
		stats.Days[today].addIP(key)
		stats.Hours[hour].addIP(key)
		stats.Total.addIP(key)
	}
	if stats.TotalUnique < len(stats.Total.IPs) {
		stats.TotalUnique = len(stats.Total.IPs)
	}
	stats.clearLegacy()
	return stats, true
}

func (s *server) rollStatsLocked() {
	now := time.Now()
	if s.live.Hours == nil {
		s.live.Hours = map[string]*statBucket{}
	}
	if s.live.Days == nil {
		s.live.Days = map[string]*statBucket{}
	}
	changed := false
	hour := now.Format("2006-01-02T15")
	today := now.Format("2006-01-02")
	if s.live.Hours[hour] == nil {
		s.live.Hours[hour] = &statBucket{}
		changed = true
	}
	if s.live.Days[today] == nil {
		s.live.Days[today] = &statBucket{}
		changed = true
	}
	if s.pruneStatsLocked(now) {
		changed = true
	}
	if changed {
		s.saveStatsLocked()
	}
}

func (s *server) hourBucketLocked() *statBucket {
	key := hourKey()
	bucket := s.live.Hours[key]
	if bucket == nil {
		bucket = &statBucket{}
		s.live.Hours[key] = bucket
	}
	return bucket
}

func (s *server) dayBucketLocked() *statBucket {
	key := todayKey()
	bucket := s.live.Days[key]
	if bucket == nil {
		bucket = &statBucket{}
		s.live.Days[key] = bucket
	}
	return bucket
}

func (s *server) pruneStatsLocked(now time.Time) bool {
	changed := false
	hourCutoff := now.Add(-24 * time.Hour)
	for key := range s.live.Hours {
		at, err := time.ParseInLocation("2006-01-02T15", key, time.Local)
		if err != nil || at.Before(hourCutoff) {
			delete(s.live.Hours, key)
			changed = true
		}
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	dayCutoff := today.AddDate(0, 0, -29)
	for key := range s.live.Days {
		at, err := time.ParseInLocation("2006-01-02", key, time.Local)
		if err != nil || at.Before(dayCutoff) {
			delete(s.live.Days, key)
			changed = true
		}
	}
	return changed
}

func (s *server) saveStatsLocked() {
	if s.statsFile == "" {
		return
	}
	s.pruneStatsLocked(time.Now())
	content, err := json.Marshal(s.live)
	if err != nil {
		return
	}
	if previous, err := os.ReadFile(s.statsFile); err == nil {
		_ = os.WriteFile(s.statsFile+".bak", previous, 0o600)
	}
	temp := s.statsFile + ".tmp"
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		return
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return
	}
	if err := file.Close(); err != nil {
		return
	}
	if err := os.Rename(temp, s.statsFile); err != nil {
		fmt.Printf("stats: could not save: %v\n", err)
	}
}

func (s *server) recordResult(result *ratio.Result, tunnel *bool) {
	if result == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollStatsLocked()
	s.hourBucketLocked().addResult(result, tunnel)
	s.dayBucketLocked().addResult(result, tunnel)
	s.live.Total.addResult(result, tunnel)
	s.saveStatsLocked()
}

func (s *server) recordJobError(code string) {
	if code == "" || code == "cancelled" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollStatsLocked()
	s.hourBucketLocked().addFail(code)
	s.dayBucketLocked().addFail(code)
	s.live.Total.addFail(code)
	s.saveStatsLocked()
}

func (s *server) recordProxyUse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollStatsLocked()
	s.hourBucketLocked().Proxy++
	s.dayBucketLocked().Proxy++
	s.live.Total.Proxy++
	s.saveStatsLocked()
}

func (s *server) rangeStatsLocked(hourly bool, from time.Time) statBucket {
	out := statBucket{IPs: map[string]bool{}}
	if hourly {
		for key, bucket := range s.live.Hours {
			at, err := time.ParseInLocation("2006-01-02T15", key, time.Local)
			if err != nil || at.Before(from) {
				continue
			}
			out.add(bucket)
		}
		return out
	}
	for key, bucket := range s.live.Days {
		at, err := time.ParseInLocation("2006-01-02", key, time.Local)
		if err != nil || at.Before(from) {
			continue
		}
		out.add(bucket)
	}
	return out
}

func statView(bucket statBucket, unique int) map[string]any {
	view := map[string]any{
		"tests":          bucket.Tests,
		"with_ratio":     bucket.WithRatio,
		"normal":         bucket.Normal,
		"invalid":        bucket.Invalid,
		"tunnel":         bucket.Tunnel,
		"non_tunnel":     bucket.NonTunnel,
		"unknown_tunnel": bucket.Unknown,
		"downloaded":     bucket.Downloaded,
		"deducted":       bucket.Deducted,
		"failed":         bucket.Failed,
		"proxy":          bucket.Proxy,
		"unique_ips":     unique,
	}
	if bucket.ClaimedTests > 0 {
		count := float64(bucket.ClaimedTests)
		view["claimed_tests"] = bucket.ClaimedTests
		view["claimed_avg"] = bucket.ClaimedSum / count
		view["claimed_real_avg"] = bucket.ClaimedRealSum / count
		view["claimed_diff_avg"] = bucket.ClaimedDiffSum / count
		if bucket.ClaimedSum > 0 {
			view["claimed_diff_pct"] = (bucket.ClaimedRealSum - bucket.ClaimedSum) / bucket.ClaimedSum * 100
		}
	}
	return view
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

	current := time.Now()
	midnight := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.Local)
	h24 := s.rangeStatsLocked(true, current.Truncate(time.Hour).Add(-23*time.Hour))
	d7 := s.rangeStatsLocked(false, midnight.AddDate(0, 0, -6))
	d30 := s.rangeStatsLocked(false, midnight.AddDate(0, 0, -29))
	unique := s.live.TotalUnique
	if unique < len(s.live.Total.IPs) {
		unique = len(s.live.Total.IPs)
	}
	todayView := statView(h24, len(h24.IPs))
	todayView["date"] = todayKey()
	monthView := statView(d30, len(d30.IPs))
	monthView["month"] = monthKey()
	return map[string]any{
		"online":  online,
		"h24":     statView(h24, len(h24.IPs)),
		"d7":      statView(d7, len(d7.IPs)),
		"d30":     statView(d30, len(d30.IPs)),
		"overall": statView(s.live.Total, unique),
		// Aliases for older clients.
		"today":   todayView,
		"monthly": monthView,
	}
}

func (s *server) recordVisitorLocked(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return
	}
	sum := sha256.Sum256([]byte(ip))
	key := hex.EncodeToString(sum[:])
	s.rollStatsLocked()
	changed := s.hourBucketLocked().addIP(key)
	if s.dayBucketLocked().addIP(key) {
		changed = true
	}
	if s.live.Total.IPs == nil {
		s.live.Total.IPs = map[string]bool{}
	}
	// shortcut: all-time unique stops growing once the bounded set is full.
	if !s.live.Total.IPs[key] && len(s.live.Total.IPs) < maxUniqueIPs {
		s.live.Total.IPs[key] = true
		s.live.TotalUnique++
		changed = true
	}
	if changed {
		s.saveStatsLocked()
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
	s.recordVisitorLocked(ip)
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
	if !ratio.IsPublicHost(host) {
		apiError(w, "blocked_target", "The target must be publicly reachable.", http.StatusBadRequest)
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

func (s *server) delayCheck(w http.ResponseWriter, r *http.Request) {
	if !s.rateLimit("delay", clientIP(r), 600) {
		apiError(w, "rate_limited", "Too many checks. Try again later.", http.StatusTooManyRequests)
		return
	}
	var input struct {
		Config string `json:"config"`
	}
	if err := decodeJSONLimit(w, r, &input, 64*1024); err != nil {
		apiError(w, "invalid_input", "Invalid delay request.", http.StatusBadRequest)
		return
	}
	config := strings.TrimSpace(input.Config)
	if config == "" || len(config) > 64*1024 {
		apiError(w, "invalid_input", "Config is required.", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	latency, err := ratio.Delay(ctx, config, 4*time.Second)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "latency_ms": 0})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "latency_ms": latency})
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
	ips, err := net.DefaultResolver.LookupIPAddr(r.Context(), host)
	if err != nil || len(ips) == 0 {
		apiError(w, "blocked_target", "The target must be publicly reachable.", http.StatusBadRequest)
		return
	}
	dialer := net.Dialer{Timeout: pingTimeout}
	start := time.Now()
	var connection net.Conn
	for _, item := range ips {
		if !ratio.IsPublicIP(item.IP) {
			apiError(w, "blocked_target", "The target must be publicly reachable.", http.StatusBadRequest)
			return
		}
		connection, err = dialer.DialContext(r.Context(), "tcp", net.JoinHostPort(item.IP.String(), port))
		if err == nil {
			break
		}
	}
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
