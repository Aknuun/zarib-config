package web

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	adminRealm          = `Basic realm="Zarib Admin", charset="UTF-8"`
	maxAdminAuthPerHour = 20
)

type adminRange struct {
	Label          string
	Tests          int
	WithRatio      int
	Normal         int
	Invalid        int
	Tunnel         int
	NonTunnel      int
	Unknown        int
	Downloaded     int64
	Deducted       int64
	Proxy          int
	UniqueIPs      int
	ClaimedTests   int
	ClaimedAvg     float64
	RealAvg        float64
	ClaimedDiffPct float64
	Failed         failStats
}

type adminView struct {
	Online       int
	Version      string
	GeneratedAt  string
	Ranges       []adminRange
	Reports      []storedReport
	Failures     []storedFailure
	RepeatIPs    []repeatIP
	UI           uiSettings
	FailureQuery string
	FailurePage  int
	FailurePages int
	FailureTotal int
	FailurePrev  int
	FailureNext  int
}

func (s *server) adminEnabled() bool {
	return s.adminUser != "" && s.adminPassword != ""
}

func constantTimeEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}

// authorizeAdmin checks Basic Auth and writes the failure response itself.
func (s *server) authorizeAdmin(w http.ResponseWriter, r *http.Request, asJSON bool) bool {
	fail := func(code, message string, status int) bool {
		if asJSON {
			apiError(w, code, message, status)
		} else {
			http.Error(w, message, status)
		}
		return false
	}
	if !s.adminEnabled() {
		return fail("admin_disabled", "Admin panel is not configured.", http.StatusServiceUnavailable)
	}
	user, password, ok := r.BasicAuth()
	if ok && constantTimeEqual(user, s.adminUser) && constantTimeEqual(password, s.adminPassword) {
		return true
	}
	if !s.rateLimit("admin-auth", clientIP(r), maxAdminAuthPerHour) {
		return fail("rate_limited", "Too many login attempts. Try again later.", http.StatusTooManyRequests)
	}
	w.Header().Set("WWW-Authenticate", adminRealm)
	return fail("unauthorized", "Authentication required.", http.StatusUnauthorized)
}

func (s *server) adminPage(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, false) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	if err := adminTemplate.Execute(w, s.adminView(r)); err != nil {
		fmt.Printf("admin: could not render page: %v\n", err)
	}
}

func (s *server) adminStats(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, true) {
		return
	}
	s.mu.Lock()
	snapshot := s.liveSnapshotLocked()
	s.mu.Unlock()
	snapshot["version"] = ratio.Version
	snapshot["generated_at"] = time.Now().UTC().Format(time.RFC3339)
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *server) adminReports(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, true) {
		return
	}
	reports, err := s.readReports(maxAdminReports)
	if err != nil {
		apiError(w, "reports_failed", "Could not read reports.", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":        len(reports),
		"reports":      reports,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *server) adminBlockIP(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeAdmin(w, r, false) {
		return
	}
	if !sameOriginRequest(r) {
		http.Error(w, "Cross-origin request blocked.", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form.", http.StatusBadRequest)
		return
	}
	ip := strings.TrimSpace(r.FormValue("ip"))
	if net.ParseIP(ip) == nil {
		http.Error(w, "Invalid IP.", http.StatusBadRequest)
		return
	}
	action := r.FormValue("action")
	if action != "block" && action != "unblock" {
		http.Error(w, "Invalid action.", http.StatusBadRequest)
		return
	}
	if err := s.setBlockedIP(ip, action == "block"); err != nil {
		fmt.Printf("admin: could not save blocked IP: %v\n", err)
		http.Error(w, "Could not save.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin#repeat-ips", http.StatusSeeOther)
}

func (s *server) adminView(req ...*http.Request) adminView {
	s.mu.Lock()
	snapshot := s.liveSnapshotLocked()
	s.mu.Unlock()
	reports, err := s.readReports(maxAdminReports)
	if err != nil {
		fmt.Printf("admin: could not read reports: %v\n", err)
	}
	allFailures, err := s.readFailures(0)
	if err != nil {
		fmt.Printf("admin: could not read failures: %v\n", err)
	}
	query := ""
	page := 1
	if len(req) > 0 && req[0] != nil {
		query = strings.TrimSpace(req[0].URL.Query().Get("q"))
		page, _ = strconv.Atoi(req[0].URL.Query().Get("page"))
	}
	if page < 1 {
		page = 1
	}
	filtered := allFailures
	if query != "" {
		filtered = nil
		for _, failure := range allFailures {
			if failureMatches(failure, query) {
				filtered = append(filtered, failure)
			}
		}
	}
	total := len(filtered)
	pages := (total + 19) / 20
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}
	start := (page - 1) * 20
	end := start + 20
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	failures := filtered[start:end]
	online, _ := snapshot["online"].(int)
	return adminView{
		Online:      online,
		Version:     ratio.Version,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Ranges: []adminRange{
			adminRangeFrom(snapshot, "h24", "۲۴ ساعت اخیر"),
			adminRangeFrom(snapshot, "d7", "۷ روز اخیر"),
			adminRangeFrom(snapshot, "d30", "۳۰ روز اخیر"),
			adminRangeFrom(snapshot, "overall", "کل"),
		},
		Reports:      reports,
		Failures:     failures,
		RepeatIPs:    s.repeatedIPs(allFailures, time.Now().Add(-24*time.Hour)),
		UI:           s.uiSnapshot(),
		FailureQuery: query,
		FailurePage:  page,
		FailurePages: pages,
		FailureTotal: total,
		FailurePrev:  page - 1,
		FailureNext:  page + 1,
	}
}

func failureMatches(f storedFailure, query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return true
	}
	parts := []string{
		f.Time, f.Source, f.JobID, f.Mode, fmt.Sprint(f.Pick), fmt.Sprint(f.Picks),
		fmt.Sprint(f.SizeMB), f.Core, f.SubURL, f.ErrorCode, f.ErrorMessage,
		f.IP, f.UserAgent, f.AcceptLanguage, strings.Join(f.Log, "\n"),
	}
	return strings.Contains(strings.ToLower(strings.Join(parts, "\n")), needle)
}

func adminRangeFrom(snapshot map[string]any, key, label string) adminRange {
	values, _ := snapshot[key].(map[string]any)
	number := func(name string) int {
		value, _ := values[name].(int)
		return value
	}
	downloaded, _ := values["downloaded"].(int64)
	deducted, _ := values["deducted"].(int64)
	decimal := func(name string) float64 {
		value, _ := values[name].(float64)
		return value
	}
	failed, _ := values["failed"].(failStats)
	return adminRange{
		Label:          label,
		Tests:          number("tests"),
		WithRatio:      number("with_ratio"),
		Normal:         number("normal"),
		Invalid:        number("invalid"),
		Tunnel:         number("tunnel"),
		NonTunnel:      number("non_tunnel"),
		Unknown:        number("unknown_tunnel"),
		Downloaded:     downloaded,
		Deducted:       deducted,
		Proxy:          number("proxy"),
		UniqueIPs:      number("unique_ips"),
		ClaimedTests:   number("claimed_tests"),
		ClaimedAvg:     decimal("claimed_avg"),
		RealAvg:        decimal("claimed_real_avg"),
		ClaimedDiffPct: decimal("claimed_diff_pct"),
		Failed:         failed,
	}
}

var adminTemplate = template.Must(template.New("admin").Funcs(template.FuncMap{
	"bytes":   humanBytes,
	"httpURL": isValidURL,
	"shortTime": func(value string) string {
		if at, err := time.Parse(time.RFC3339Nano, value); err == nil {
			return at.Local().Format("2006-01-02 15:04:05")
		}
		return value
	},
}).Parse(adminPageHTML))

const adminPageHTML = `<!doctype html>
<html lang="fa" dir="rtl">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex,nofollow">
<title>پنل مدیریت ضریب</title>
<style>
:root {
  --bg: #f6f7f9;
  --tile: #ffffff;
  --line: #e2e5ea;
  --text: #101828;
  --muted: #667085;
  --brand: #1d4ed8;
  --good: #067647;
  --bad: #d92d20;
  --recessed: #f2f4f7;
  --shadow: 0 1px 2px rgba(16,24,40,.05), 0 16px 32px -28px rgba(16,24,40,.35);
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0b0f16;
    --tile: #111827;
    --line: #243044;
    --text: #e5e7eb;
    --muted: #9ca3af;
    --brand: #60a5fa;
    --good: #75e0a7;
    --bad: #f97066;
    --recessed: #0f172a;
    --shadow: 0 1px 2px rgba(0,0,0,.5), 0 16px 32px -28px rgba(0,0,0,.9);
  }
}
* { box-sizing: border-box; }
body { margin: 0; background: var(--bg); color: var(--text); font-family: Tahoma, "Segoe UI", system-ui, sans-serif; line-height: 1.6; }
a { color: var(--brand); }
.top { position: sticky; top: 0; z-index: 2; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 14px clamp(16px,4vw,32px); border-bottom: 1px solid var(--line); background: color-mix(in oklab, var(--bg) 88%, transparent); backdrop-filter: blur(10px); }
.brand { display: flex; align-items: center; gap: 12px; min-width: 0; }
.logo { display: grid; place-items: center; width: 40px; height: 40px; border-radius: 12px; background: var(--brand); color: #fff; font-weight: 800; font-size: 18px; }
h1 { margin: 0; font-size: 16px; }
.brand p { margin: 0; color: var(--muted); font-size: 12px; }
.top-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.pill { display: inline-flex; align-items: center; gap: 6px; padding: 6px 10px; border: 1px solid var(--line); border-radius: 10px; background: var(--tile); font-size: 12px; text-decoration: none; color: inherit; }
.pill.online { color: var(--good); }
.dot { width: 7px; height: 7px; border-radius: 50%; background: var(--good); }
main { display: flex; flex-direction: column; gap: 20px; padding: 24px clamp(16px,4vw,32px) 48px; max-width: 1400px; margin-inline: auto; }
.section-head { display: flex; flex-wrap: wrap; align-items: baseline; justify-content: space-between; gap: 8px; }
.section-head h2 { margin: 0; font-size: 20px; }
.section-head p { margin: 0; color: var(--muted); font-size: 12px; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 16px; }
.tile { padding: 18px; border: 1px solid var(--line); border-radius: 14px; background: var(--tile); box-shadow: var(--shadow); }
.tile-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 12px; }
.tile-head h3 { margin: 0; font-size: 15px; }
.badge { padding: 3px 8px; border-radius: 999px; background: var(--recessed); color: var(--muted); font-size: 11px; }
.stats { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; margin: 0; }
.stats > div { padding: 8px 10px; border: 1px solid var(--line); border-radius: 10px; background: var(--recessed); }
.stats dt { color: var(--muted); font-size: 11px; }
.stats dd { margin: 0; font-size: 15px; font-weight: 700; font-variant-numeric: tabular-nums; }
.lines { display: flex; flex-direction: column; gap: 4px; margin-top: 12px; color: var(--muted); font-size: 12px; }
.lines p { margin: 0; }
.lines b { color: var(--text); font-weight: 600; }
.table-wrap { overflow-x: auto; border: 1px solid var(--line); border-radius: 14px; background: var(--tile); box-shadow: var(--shadow); }
table { width: 100%; border-collapse: collapse; font-size: 13px; }
th, td { padding: 10px 12px; border-bottom: 1px solid var(--line); text-align: start; vertical-align: top; }
th { color: var(--muted); font-size: 11px; font-weight: 600; white-space: nowrap; }
tr:last-child td { border-bottom: 0; }
.desc { max-width: 420px; }
.mono, code, pre { font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace; }
code { padding: 1px 5px; border-radius: 5px; background: var(--recessed); font-size: 12px; }
.nowrap { white-space: nowrap; }
details summary { display: flex; align-items: center; min-height: 32px; gap: 6px; cursor: pointer; color: var(--brand); white-space: nowrap; }
details summary::-webkit-details-marker { display: none; }
.details { min-width: 260px; margin-top: 8px; padding: 12px; border: 1px solid var(--line); border-radius: 10px; background: var(--recessed); }
.details h4 { margin: 10px 0 4px; font-size: 12px; color: var(--muted); }
.details h4:first-child { margin-top: 0; }
.mini { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 2px 12px; margin: 0; font-size: 12px; }
.mini dt { color: var(--muted); }
.mini dd { margin: 0; overflow-wrap: anywhere; }
pre { max-height: 260px; margin: 4px 0 0; padding: 10px; overflow: auto; border: 1px solid var(--line); border-radius: 8px; background: var(--tile); font-size: 11px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere; }
.z-ui-form { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
.z-ui-form label { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; color: var(--muted); }
.z-ui-form select { min-height: 32px; padding: 4px 8px; border: 1px solid var(--line); border-radius: 8px; background: var(--tile); color: var(--text); font: inherit; }
.z-ui-form .z-ui-checks { display: flex; flex-wrap: wrap; gap: 10px 14px; flex: 1 1 320px; }
.z-ui-form input[type="checkbox"] { accent-color: var(--brand); }
.empty { padding: 20px; border: 1px dashed var(--line); border-radius: 14px; color: var(--muted); text-align: center; }
.z-failure-search { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.z-failure-search input { flex: 1 1 260px; min-height: 34px; padding: 6px 10px; border: 1px solid var(--line); border-radius: 9px; background: var(--tile); color: var(--text); font: inherit; }
.z-pager { display: flex; align-items: center; justify-content: center; gap: 10px; margin-top: 12px; color: var(--muted); font-size: 12px; }
.btn { display: inline-flex; align-items: center; justify-content: center; min-height: 32px; padding: 4px 10px; border: 1px solid var(--line); border-radius: 9px; background: var(--tile); color: var(--text); font: inherit; font-size: 12px; cursor: pointer; }
.btn:hover { border-color: var(--brand); color: var(--brand); }
.btn.danger { border-color: color-mix(in oklab, var(--bad) 35%, var(--line)); color: var(--bad); }
.btn.danger:hover { background: color-mix(in oklab, var(--bad) 8%, var(--tile)); }
.inline-form { margin: 0; }
@media (max-width: 640px) { .desc { min-width: 180px; } th, td { padding: 8px 10px; } }
</style>
</head>
<body>
<header class="top">
  <div class="brand">
    <span class="logo">Z</span>
    <div><h1>پنل مدیریت ضریب</h1><p>آمار زنده و گزارش‌های خطا</p></div>
  </div>
  <div class="top-meta">
    <span class="pill online"><span class="dot"></span>{{.Online}} آنلاین</span>
    <span class="pill">نسخه {{.Version}}</span>
    <span class="pill">{{.GeneratedAt}}</span>
    <a class="pill" href="/admin">به‌روزرسانی</a>
  </div>
</header>
<main>
  <section>
    <div class="section-head"><h2>آمار</h2><p>۲۴ ساعت، ۷ روز، ۳۰ روز و کل</p></div>
  </section>
  <div class="grid">
    {{range .Ranges}}
    <article class="tile">
      <div class="tile-head"><h3>{{.Label}}</h3><span class="badge">{{.Tests}} تست</span></div>
      <dl class="stats">
        <div><dt>تست</dt><dd>{{.Tests}}</dd></div>
        <div><dt>ضریب‌دار</dt><dd>{{.WithRatio}}</dd></div>
        <div><dt>معمولی</dt><dd>{{.Normal}}</dd></div>
        <div><dt>نامعتبر</dt><dd>{{.Invalid}}</dd></div>
        <div><dt>دانلود</dt><dd>{{bytes .Downloaded}}</dd></div>
        <div><dt>کسرشده</dt><dd>{{bytes .Deducted}}</dd></div>
        <div><dt>کاربران مجزا</dt><dd>{{.UniqueIPs}}</dd></div>
        <div><dt>با پروکسی</dt><dd>{{.Proxy}}</dd></div>
      </dl>
      <div class="lines">
        <p>تانل <b>{{.Tunnel}}</b> · غیرتانل <b>{{.NonTunnel}}</b> · نامشخص <b>{{.Unknown}}</b></p>
        <p>خطاها <b>{{.Failed.Total}}</b> · اتصال <b>{{.Failed.Connect}}</b> · ساب <b>{{.Failed.Fetch}}</b> · هسته <b>{{.Failed.Core}}</b> · تایم‌اوت <b>{{.Failed.Timeout}}</b> · بدون تغییر <b>{{.Failed.Usage}}</b> · سایر <b>{{.Failed.Other}}</b></p>
        <p>ادعایی <b>{{if .ClaimedTests}}{{printf "%.2f" .ClaimedAvg}}×{{else}}—{{end}}</b> · واقعی <b>{{if .ClaimedTests}}{{printf "%.2f" .RealAvg}}×{{else}}—{{end}}</b> · اختلاف <b>{{if .ClaimedTests}}{{printf "%+.1f%%" .ClaimedDiffPct}}{{else}}—{{end}}</b>{{if .ClaimedTests}} ({{.ClaimedTests}} تست){{end}}</p>
      </div>
    </article>
    {{end}}
  </div>

  <section id="animations">
    <div class="section-head"><h2>انیمیشن‌ها</h2><p>شدت و نوع حرکت‌ها؛ برای قطع اینترنت بین‌الملل «خاموش» را انتخاب کن.</p></div>
    <form method="post" action="/admin/api/ui" class="tile z-ui-form">
      <label>شدت
        <select name="animations">
          <option value="full" {{if eq .UI.Animations "full"}}selected{{end}}>کامل</option>
          <option value="light" {{if eq .UI.Animations "light"}}selected{{end}}>سبک</option>
          <option value="off" {{if eq .UI.Animations "off"}}selected{{end}}>خاموش</option>
        </select>
      </label>
      <div class="z-ui-checks">
        <label><input type="checkbox" name="page" {{if .UI.Page}}checked{{end}}> ورود صفحه‌ها</label>
        <label><input type="checkbox" name="rows" {{if .UI.Rows}}checked{{end}}> ردیف کانفیگ‌ها</label>
        <label><input type="checkbox" name="chips" {{if .UI.Chips}}checked{{end}}> چیپ‌های پینگ</label>
        <label><input type="checkbox" name="result" {{if .UI.Result}}checked{{end}}> نتیجه و ضریب</label>
        <label><input type="checkbox" name="buttons" {{if .UI.Buttons}}checked{{end}}> دکمه‌ها</label>
        <label><input type="checkbox" name="tiles" {{if .UI.Tiles}}checked{{end}}> کارت‌ها</label>
      </div>
      <button class="btn" type="submit">ذخیره</button>
    </form>
  </section>

  <section id="repeat-ips">
    <div class="section-head"><h2>IPهای پرتکرار</h2><p>IPهایی که در ۲۴ ساعت اخیر ۲ بار یا بیشتر خطا داده‌اند</p></div>
  </section>
  {{if .RepeatIPs}}
  <div class="table-wrap">
    <table>
      <thead><tr><th>IP</th><th>خطا در ۲۴ ساعت</th><th>آخرین خطا</th><th>کدها</th><th>وضعیت</th><th>عملیات</th></tr></thead>
      <tbody>
        {{range .RepeatIPs}}
        <tr>
          <td class="mono">{{.IP}}</td>
          <td>{{.Count}}</td>
          <td class="nowrap">{{if .Last}}{{shortTime .Last}}{{else}}—{{end}}</td>
          <td>{{if .Codes}}<code>{{.Codes}}</code>{{else}}—{{end}}</td>
          <td>{{if .Blocked}}<span class="badge">مسدود</span>{{else if .Silent}}<span class="badge">سایلنت</span>{{else}}—{{end}}</td>
          <td>
            <form method="post" action="/admin/api/block" class="inline-form">
              <input type="hidden" name="ip" value="{{.IP}}">
              {{if .Blocked}}
              <input type="hidden" name="action" value="unblock">
              <button class="btn" type="submit">آزاد کردن</button>
              {{else}}
              <input type="hidden" name="action" value="block">
              <button class="btn danger" type="submit">مسدود کردن</button>
              {{end}}
            </form>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{else}}
  <p class="empty">IP پرتکراری در ۲۴ ساعت اخیر نیست.</p>
  {{end}}

  <section>
    <div class="section-head"><h2>گزارش‌های خطا</h2><p>{{len .Reports}} گزارش</p></div>
  </section>
  {{if .Reports}}
  <div class="table-wrap">
    <table>
      <thead><tr><th>زمان</th><th>شرح</th><th>لینک ساب</th><th>خطا</th><th>صفحه</th><th>IP</th><th></th></tr></thead>
      <tbody>
        {{range .Reports}}
        <tr>
          <td class="nowrap">{{shortTime .Time}}</td>
          <td class="desc">{{.Description}}</td>
          <td class="desc mono" dir="ltr">{{if .SubURL}}<a href="{{.SubURL}}">{{.SubURL}}</a>{{else}}—{{end}}</td>
          <td>{{if .ErrorCode}}<code>{{.ErrorCode}}</code>{{else}}—{{end}}</td>
          <td>{{if .View}}{{.View}}{{else}}{{with .Meta}}{{with index . "page"}}{{.}}{{end}}{{end}}{{end}}</td>
          <td class="mono">{{.IP}}</td>
          <td>
            <details>
              <summary>جزئیات</summary>
              <div class="details">
                <h4>شرح</h4>
                <div>{{.Description}}</div>
                {{if .ErrorCode}}<h4>کد خطا</h4><div><code>{{.ErrorCode}}</code></div>{{end}}
                {{if .ErrorMessage}}<h4>پیام خطا</h4><pre dir="ltr">{{.ErrorMessage}}</pre>{{end}}
                {{if .ErrorLog}}<h4>لاگ خطا</h4><pre dir="ltr">{{range .ErrorLog}}{{.}}
{{end}}</pre>{{end}}
                {{with .Result}}<h4>خلاصه نتیجه</h4>
                <dl class="mini">
                  <dt>ضریب</dt><dd>{{printf "%.2f" .Ratio}}×</dd>
                  <dt>دانلود واقعی</dt><dd>{{bytes .DownloadedBytes}}</dd>
                  <dt>کسرشده</dt><dd>{{bytes .DeltaTotal}}</dd>
                  {{if .ConfigsTested}}<dt>تعداد کانفیگ</dt><dd>{{.ConfigsTested}}</dd>{{end}}
                  {{if .Core}}<dt>هسته</dt><dd>{{.Core}}{{if .XrayVersion}} — {{.XrayVersion}}{{end}}</dd>{{end}}
                  {{if .TestedAt}}<dt>زمان تست</dt><dd>{{.TestedAt}}</dd>{{end}}
                </dl>{{end}}
                <h4>مشخصات فرستنده</h4>
                <dl class="mini">
                  <dt>زمان</dt><dd>{{.Time}}</dd>
                  <dt>IP</dt><dd class="mono">{{.IP}}</dd>
                  <dt>User-Agent</dt><dd>{{.UserAgent}}</dd>
                  <dt>Accept-Language</dt><dd>{{.AcceptLanguage}}</dd>
                  <dt>Referer</dt><dd>{{.Referer}}</dd>
                  {{with .Meta}}{{range $k, $v := .}}<dt>{{$k}}</dt><dd>{{$v}}</dd>{{end}}{{end}}
                </dl>
              </div>
            </details>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{else}}
  <p class="empty">هنوز گزارشی ثبت نشده است.</p>
  {{end}}

  <section id="failures">
    <div class="section-head"><h2>خطاهای تست</h2><p>{{.FailureTotal}} مورد · صفحه {{.FailurePage}} از {{.FailurePages}}</p></div>
    <form class="z-failure-search" method="get" action="/admin#failures">
      <input name="q" value="{{.FailureQuery}}" placeholder="جستجو در همه خطاها: لینک، IP، کد خطا، لاگ…" aria-label="جستجو در خطاها">
      <button class="btn" type="submit">جستجو</button>
      {{if .FailureQuery}}<a class="btn" href="/admin#failures">پاک کردن</a>{{end}}
    </form>
  </section>
  {{if .Failures}}
  <div class="table-wrap">
    <table>
      <thead><tr><th>زمان</th><th>منبع</th><th>کد خطا</th><th>لینک ساب</th><th>حالت</th><th>IP</th><th></th></tr></thead>
      <tbody>
        {{range .Failures}}
        <tr>
          <td class="nowrap">{{shortTime .Time}}</td>
          <td>{{if eq .Source "job"}}تست{{else}}بررسی ساب{{end}}</td>
          <td>{{if .ErrorCode}}<code>{{.ErrorCode}}</code>{{else}}—{{end}}</td>
          <td class="desc mono" dir="ltr">{{if .SubURL}}{{if httpURL .SubURL}}<a href="{{.SubURL}}">{{.SubURL}}</a>{{else}}{{.SubURL}}{{end}}{{else}}—{{end}}</td>
          <td class="nowrap">{{if eq .Mode "batch"}}همزمان{{else if eq .Mode "single"}}تک‌به‌تک{{else}}—{{end}}{{if .Pick}} · کانفیگ {{.Pick}}{{end}}{{if .Picks}} · {{range .Picks}}{{.}} {{end}}{{end}}</td>
          <td class="mono">{{.IP}}</td>
          <td>
            <details>
              <summary>لاگ</summary>
              <div class="details">
                {{if .ErrorMessage}}<h4>پیام خطا</h4><pre dir="ltr">{{.ErrorMessage}}</pre>{{end}}
                {{if .Log}}<h4>لاگ</h4><pre dir="ltr">{{range .Log}}{{.}}
{{end}}</pre>{{end}}
                <h4>جزئیات</h4>
                <dl class="mini">
                  <dt>زمان</dt><dd>{{.Time}}</dd>
                  {{if .JobID}}<dt>شناسه تست</dt><dd class="mono">{{.JobID}}</dd>{{end}}
                  {{if .SizeMB}}<dt>حجم</dt><dd>{{printf "%.0f" .SizeMB}} MB</dd>{{end}}
                  {{if .Core}}<dt>هسته</dt><dd>{{.Core}}</dd>{{end}}
                  <dt>IP</dt><dd class="mono">{{.IP}}</dd>
                  <dt>User-Agent</dt><dd>{{.UserAgent}}</dd>
                </dl>
              </div>
            </details>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{if gt .FailurePages 1}}
  <div class="z-pager">
    {{if gt .FailurePage 1}}<a class="btn" href="/admin?q={{urlquery .FailureQuery}}&page={{.FailurePrev}}#failures">قبلی</a>{{end}}
    <span>صفحه {{.FailurePage}} از {{.FailurePages}}</span>
    {{if lt .FailurePage .FailurePages}}<a class="btn" href="/admin?q={{urlquery .FailureQuery}}&page={{.FailureNext}}#failures">بعدی</a>{{end}}
  </div>
  {{end}}
  {{else}}
  <p class="empty">{{if .FailureQuery}}خطایی با این جستجو پیدا نشد.{{else}}هنوز خطایی ثبت نشده است.{{end}}</p>
  {{end}}
</main>
</body>
</html>`
