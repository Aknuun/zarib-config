package ratio

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SubscriptionError struct {
	Code    string
	Message string
}

func (e *SubscriptionError) Error() string { return e.Message }

type Usage struct {
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    *int64 `json:"total,omitempty"`
	Expire   *int64 `json:"expire,omitempty"`
}

func (u Usage) Used() int64 { return u.Upload + u.Download }

func (u Usage) Remaining() *int64 {
	if u.Total == nil || *u.Total <= 0 {
		return nil
	}
	value := *u.Total - u.Upload - u.Download
	if value < 0 {
		value = 0
	}
	return &value
}

type Config struct {
	URI      string
	JSON     map[string]any
	Protocol string
	Target   string
	Remark   string
}

var blockedHosts = map[string]bool{
	"localhost": true, "localhost.localdomain": true, "metadata.google.internal": true,
	"metadata": true, "instance-data": true,
}

// Panels often add a dummy info line such as "14GB remaining" or "expired"
// that is not a usable node. Keep those out of the config picker.
var placeholderRemarkPattern = regexp.MustCompile(`(?i)(expired|remaining|باقی[\s\x{200c}]*مانده|به[\s\x{200c}]*پایان[\s\x{200c}]*رسیده|پایان[\s\x{200c}]*رسیده|حجم[\s\x{200c}]*شما|زمان[\s\x{200c}]*شما)`)

// Some panels fake a node with a date-like host such as 14.47--2026.11.05.time.
var placeholderHostPattern = regexp.MustCompile(`(?i)^\d{1,2}\.\d{1,2}--\d{4}\.\d{2}\.\d{2}\.`)

func safeFetchEnabled() bool {
	return os.Getenv("RATIO_SAFE_FETCH") == "1"
}

func IsPublicURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" {
		return false
	}
	if blockedHosts[host] || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, item := range ips {
		if !isPublicIP(item.IP) {
			return false
		}
	}
	return true
}

func IsPublicHost(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if host == "" {
		return false
	}
	if blockedHosts[host] || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return isPublicIP(ip)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, item := range ips {
		if !isPublicIP(item.IP) {
			return false
		}
	}
	return true
}

func IsPublicIP(ip net.IP) bool {
	return isPublicIP(ip)
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 0 || ip4[0] >= 240 {
			return false
		}
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
		if ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0 {
			return false
		}
		if ip4[0] == 198 && (ip4[1] == 18 || ip4[1] == 19) {
			return false
		}
	}
	return true
}

func subProxyURL() *url.URL {
	raw := strings.TrimSpace(os.Getenv("ZARIB_SUB_PROXY"))
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return nil
	}
	return parsed
}

// subProxyURLs returns the configured proxies in order. The original
// ZARIB_SUB_PROXY is tried first; ZARIB_SUB_PROXY_HOSTS adds host:port
// fallbacks that reuse its credentials.
func subProxyURLs() []*url.URL {
	base := subProxyURL()
	rawHosts := strings.TrimSpace(os.Getenv("ZARIB_SUB_PROXY_HOSTS"))
	if rawHosts == "" {
		if base == nil {
			return nil
		}
		return []*url.URL{base}
	}
	var out []*url.URL
	seen := map[string]bool{}
	add := func(proxy *url.URL) {
		if proxy == nil || proxy.Host == "" || seen[proxy.String()] {
			return
		}
		seen[proxy.String()] = true
		out = append(out, proxy)
	}
	add(base)
	for _, host := range strings.Split(rawHosts, ",") {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if base != nil {
			clone := *base
			clone.Host = host
			add(&clone)
			continue
		}
		raw := host
		if !strings.Contains(raw, "://") {
			raw = "http://" + raw
		}
		if parsed, err := url.Parse(raw); err == nil {
			add(parsed)
		}
	}
	return out
}

// clientUA is the fallback for panels that serve a browser landing page to a
// browser User-Agent instead of the subscription body.
const clientUA = "v2rayNG/1.9.16"

func fetchOnce(rawURL string, timeout time.Duration, proxy *url.URL, userAgent string) ([]byte, http.Header, error) {
	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, ratioErrorf("Invalid subscription URL: %v", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Pragma", "no-cache")
	client := &http.Client{Timeout: timeout}
	if proxy != nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(proxy)
		client.Transport = transport
	}
	if safeFetchEnabled() {
		if !IsPublicURL(rawURL) {
			return nil, nil, ratioErrorf("Blocked: the subscription URL must be a public http(s) address.")
		}
		if proxy == nil {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.DialContext = publicDialContext
			client.Transport = transport
		}
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if !IsPublicURL(req.URL.String()) {
				return ratioErrorf("Blocked a redirect to a non-public address.")
			}
			return nil
		}
	}
	response, err := client.Do(request)
	if err != nil {
		var ratioErr *RatioError
		if errors.As(err, &ratioErr) {
			return nil, nil, ratioErr
		}
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return nil, nil, &SubscriptionError{Code: "subscription_timeout", Message: "Subscription request timed out."}
		}
		if strings.Contains(strings.ToLower(err.Error()), "blocked") {
			return nil, nil, ratioErrorf("Blocked a redirect to a non-public address.")
		}
		return nil, nil, ratioErrorf("Could not connect to the subscription URL: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		code := "subscription_broken"
		switch response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			code = "subscription_unauthorized"
		case http.StatusNotFound, http.StatusGone:
			code = "subscription_not_found"
		case http.StatusTooManyRequests:
			code = "subscription_rate_limited"
		default:
			if response.StatusCode >= 500 {
				code = "subscription_unavailable"
			}
		}
		return nil, nil, &SubscriptionError{Code: code, Message: fmt.Sprintf("Subscription request failed with HTTP %d.", response.StatusCode)}
	}
	if isDownloadResponse(response.Request.URL.String(), response.Header) {
		return nil, nil, &SubscriptionError{Code: "not_subscription", Message: "This link points to a file download, not a subscription link."}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxSubBytes+1))
	if err != nil {
		message := fmt.Sprintf("Could not read the subscription response: %v", err)
		if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return nil, nil, &SubscriptionError{Code: "subscription_timeout", Message: message}
		}
		return nil, nil, ratioErrorf("%s", message)
	}
	if len(body) > MaxSubBytes {
		return nil, nil, ratioErrorf("Subscription response is too large.")
	}
	return body, response.Header, nil
}

var downloadExtensions = []string{
	".apk", ".aab", ".ipa", ".exe", ".msi", ".dmg", ".pkg", ".deb", ".rpm",
	".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".iso", ".img",
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg",
	".mp3", ".mp4", ".mkv", ".avi", ".mov", ".bin",
}

// isDownloadResponse detects file links such as APK releases before they are
// treated as subscriptions. A browser URL or a short link can redirect to the
// file, so the final request URL and headers are checked too.
func isDownloadResponse(rawURL string, headers http.Header) bool {
	check := func(value string) bool {
		value = strings.ToLower(value)
		for _, extension := range downloadExtensions {
			if strings.HasSuffix(value, extension) {
				return true
			}
		}
		return false
	}
	if parsed, err := url.Parse(rawURL); err == nil && check(parsed.Path) {
		return true
	}
	contentType := strings.ToLower(strings.Split(headers.Get("Content-Type"), ";")[0])
	switch contentType {
	case "application/vnd.android.package-archive", "application/zip",
		"application/x-zip-compressed", "application/pdf", "application/x-msdownload",
		"application/x-apple-diskimage", "application/x-rar-compressed",
		"application/x-7z-compressed":
		return true
	}
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		return true
	}
	disposition := strings.ToLower(headers.Get("Content-Disposition"))
	if strings.Contains(disposition, "attachment") {
		if check(disposition) || contentType == "application/octet-stream" {
			return true
		}
	}
	return false
}

func looksLikeHTML(body []byte) bool {
	text := strings.ToLower(strings.TrimSpace(DecodeSubBody(body)))
	return strings.HasPrefix(text, "<") || strings.Contains(text, "<html")
}

func looksLikePortal(body []byte) bool {
	text := strings.ToLower(string(body))
	return strings.Contains(text, "portal_api.php") &&
		(strings.Contains(text, "__vipportaltoken") || strings.Contains(text, "initvipportal"))
}

var portalTokenPattern = regexp.MustCompile(`(?i)(?:__vipPortalToken|"token")\s*[:=]\s*["']([A-Za-z0-9_-]{3,64})["']`)
var portalTokenValue = regexp.MustCompile(`^[A-Za-z0-9_-]{3,64}$`)

func portalToken(rawURL string, body []byte) string {
	if parsed, err := url.Parse(rawURL); err == nil {
		for _, key := range []string{"t", "token"} {
			if value := strings.TrimSpace(parsed.Query().Get(key)); portalTokenValue.MatchString(value) {
				return value
			}
		}
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		for index := len(parts) - 1; index >= 0; index-- {
			value := strings.TrimSpace(parts[index])
			if value != "" && value != "vip" && portalTokenValue.MatchString(value) {
				return value
			}
		}
	}
	if match := portalTokenPattern.FindSubmatch(body); len(match) == 2 {
		return string(match[1])
	}
	return ""
}

type portalServers struct {
	Servers []struct {
		URI string `json:"uri"`
	} `json:"servers"`
}

type portalPackage struct {
	VolumeGB     float64       `json:"volume_gb"`
	UsedVolumeGB float64       `json:"used_volume_gb"`
	Days         int           `json:"days"`
	Unlim        portalServers `json:"unlim"`
	Fast         portalServers `json:"fast"`
}

// resolvePortalSubscription turns a VIP portal page into the config list its
// JavaScript would normally load from portal_api.php.
func resolvePortalSubscription(rawURL string, page []byte, timeout time.Duration, proxy *url.URL) ([]byte, http.Header, error) {
	token := portalToken(rawURL, page)
	if token == "" {
		return nil, nil, ratioErrorf("Could not find the VIP portal token.")
	}
	pageURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, ratioErrorf("Invalid subscription URL: %v", err)
	}
	apiURL := pageURL.ResolveReference(&url.URL{Path: "hosi/portal_api.php"})
	query := apiURL.Query()
	query.Set("action", "package")
	query.Set("t", token)
	apiURL.RawQuery = query.Encode()
	body, _, err := fetchOnce(apiURL.String(), timeout, proxy, UA)
	if err != nil {
		return nil, nil, err
	}
	var payload struct {
		Package portalPackage `json:"package"`
		Data    struct {
			Package portalPackage `json:"package"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, nil, ratioErrorf("Could not read the VIP portal response: %v", err)
	}
	pkg := payload.Package
	if pkg.VolumeGB == 0 && pkg.UsedVolumeGB == 0 && len(pkg.Unlim.Servers) == 0 {
		pkg = payload.Data.Package
	}
	var uris []string
	seen := map[string]bool{}
	for _, group := range []portalServers{pkg.Unlim, pkg.Fast} {
		for _, server := range group.Servers {
			uri := strings.TrimSpace(server.URI)
			if uri == "" || seen[uri] || !strings.Contains(uri, "://") {
				continue
			}
			seen[uri] = true
			uris = append(uris, uri)
		}
	}
	if len(uris) == 0 {
		return nil, nil, &SubscriptionError{Code: "no_configs", Message: "The VIP portal has no link configs to test."}
	}
	headers := http.Header{}
	headers.Set("Content-Type", "text/plain; charset=utf-8")
	if pkg.VolumeGB > 0 {
		total := int64(pkg.VolumeGB * 1024 * 1024 * 1024)
		used := int64(pkg.UsedVolumeGB * 1024 * 1024 * 1024)
		days := pkg.Days
		if days <= 0 {
			days = 30
		}
		headers.Set("Subscription-Userinfo", fmt.Sprintf("upload=0; download=%d; total=%d; expire=%d", used, total, time.Now().AddDate(0, 0, days).Unix()))
	}
	return []byte(strings.Join(uris, "\n")), headers, nil
}

// fetchSubscription retries a browser landing page as a client so panels that
// redirect browser User-Agents still return their config list.
func fetchSubscription(rawURL string, timeout time.Duration, proxy *url.URL) ([]byte, http.Header, error) {
	body, headers, err := fetchOnce(rawURL, timeout, proxy, UA)
	if err != nil {
		var subErr *SubscriptionError
		if errors.As(err, &subErr) && subErr.Code == "subscription_broken" && strings.Contains(subErr.Message, "HTTP 406") {
			return fetchOnce(rawURL, timeout, proxy, clientUA)
		}
		return body, headers, err
	}
	if looksLikePortal(body) {
		return resolvePortalSubscription(rawURL, body, timeout, proxy)
	}
	if !looksLikeHTML(body) {
		return body, headers, nil
	}
	body, headers, err = fetchOnce(rawURL, timeout, proxy, clientUA)
	if err != nil {
		return body, headers, err
	}
	if looksLikePortal(body) {
		return resolvePortalSubscription(rawURL, body, timeout, proxy)
	}
	return body, headers, nil
}

func publicDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("could not resolve %s", host)
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var lastErr error
	for _, item := range ips {
		if !isPublicIP(item.IP) {
			return nil, fmt.Errorf("blocked non-public address")
		}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(item.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("could not connect")
	}
	return nil, lastErr
}

func Fetch(rawURL string, timeout time.Duration) ([]byte, http.Header, error) {
	body, headers, _, err := FetchMeta(rawURL, timeout)
	return body, headers, err
}

// FetchMeta reads a subscription directly and, when proxies are configured,
// retries through each proxy in order if the direct attempt fails.
func FetchMeta(rawURL string, timeout time.Duration) ([]byte, http.Header, bool, error) {
	started := time.Now()
	host := ""
	if u, err := url.Parse(rawURL); err == nil {
		host = u.Hostname()
	}
	defer func() {
		if os.Getenv("ZARIB_FETCH_LOG") == "1" {
			fmt.Printf("fetch: host=%s duration=%s\n", host, time.Since(started).Round(time.Millisecond))
		}
	}()
	proxies := subProxyURLs()
	directTimeout := 4 * time.Second
	if timeout > 0 && timeout < directTimeout {
		directTimeout = timeout
	}
	body, headers, err := fetchSubscription(rawURL, directTimeout, nil)
	if err == nil {
		return body, headers, false, nil
	}
	if strings.Contains(err.Error(), "Invalid subscription URL") || strings.Contains(err.Error(), "Blocked") || strings.Contains(err.Error(), "not a subscription") {
		return nil, nil, false, err
	}
	for index, proxy := range proxies {
		if !proxyAllowed(proxy) {
			continue
		}
		attemptTimeout := 3 * time.Second
		if index == 0 {
			attemptTimeout = 2 * time.Second
		}
		retryBody, retryHeaders, retryErr := fetchSubscription(rawURL, attemptTimeout, proxy)
		if retryErr == nil {
			markProxy(proxy, true)
			return retryBody, retryHeaders, true, nil
		}
		markProxy(proxy, false)
	}
	finalTimeout := 30 * time.Second
	if timeout > finalTimeout {
		finalTimeout = timeout
	}
	body, headers, err = fetchSubscription(rawURL, finalTimeout, nil)
	if err == nil {
		return body, headers, false, nil
	}
	return nil, nil, false, err
}

type proxyState struct {
	failures int
	until    time.Time
}

var (
	proxyHealthMu sync.Mutex
	proxyHealth   = map[string]proxyState{}
)

const (
	proxyFailureLimit = 3
	proxyCooldown     = 5 * time.Minute
)

func proxyAllowed(proxy *url.URL) bool {
	if proxy == nil {
		return false
	}
	proxyHealthMu.Lock()
	defer proxyHealthMu.Unlock()
	state := proxyHealth[proxy.Host]
	return state.until.IsZero() || time.Now().After(state.until)
}

func markProxy(proxy *url.URL, ok bool) {
	if proxy == nil {
		return
	}
	proxyHealthMu.Lock()
	defer proxyHealthMu.Unlock()
	if ok {
		delete(proxyHealth, proxy.Host)
		return
	}
	state := proxyHealth[proxy.Host]
	state.failures++
	if state.failures >= proxyFailureLimit {
		state.until = time.Now().Add(proxyCooldown)
		state.failures = 0
	}
	proxyHealth[proxy.Host] = state
}

var integerPattern = regexp.MustCompile(`^\d+$`)

func ParseUsage(headers http.Header, body []byte) *Usage {
	var raw string
	for key, values := range headers {
		if strings.EqualFold(key, "subscription-userinfo") && len(values) > 0 {
			raw = values[0]
			break
		}
	}
	data := map[string]int64{}
	if raw != "" {
		for _, part := range strings.Split(raw, ";") {
			if !strings.Contains(part, "=") {
				continue
			}
			pieces := strings.SplitN(part, "=", 2)
			key := strings.ToLower(strings.TrimSpace(pieces[0]))
			value := strings.TrimSpace(pieces[1])
			if integerPattern.MatchString(value) {
				number, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					data[key] = number
				}
			}
		}
	}
	if _, ok := data["upload"]; !ok {
		if _, ok := data["download"]; !ok {
			var payload map[string]any
			decoder := json.NewDecoder(bytes.NewReader(body))
			decoder.UseNumber()
			if err := decoder.Decode(&payload); err == nil && payload != nil {
				_, hasUpload := payload["upload"]
				_, hasDownload := payload["download"]
				if hasUpload || hasDownload {
					for _, key := range []string{"upload", "download", "total", "expire"} {
						if value, ok := payload[key]; ok {
							data[key] = toInt64(value)
						}
					}
				}
			}
		}
	}
	_, hasUpload := data["upload"]
	_, hasDownload := data["download"]
	if !hasUpload && !hasDownload {
		return nil
	}
	if !hasUpload {
		data["upload"] = 0
	}
	if !hasDownload {
		data["download"] = 0
	}
	usage := &Usage{Upload: data["upload"], Download: data["download"]}
	if total, ok := data["total"]; ok {
		usage.Total = &total
	}
	if expire, ok := data["expire"]; ok {
		usage.Expire = &expire
	}
	return usage
}

func ReadUsage(subURL string, retries int, delay time.Duration) (*Usage, error) {
	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		body, headers, err := Fetch(subURL, 30*time.Second)
		if err == nil {
			if usage := ParseUsage(headers, body); usage != nil {
				return usage, nil
			}
			lastErr = ratioErrorf("The subscription has no subscription-userinfo header and no usage in the body.")
		} else {
			lastErr = err
		}
		if attempt < retries-1 {
			time.Sleep(delay)
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ratioErrorf("Could not read subscription usage.")
}

func toInt64(value any) int64 {
	number := func(f float64) int64 {
		if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f >= math.MaxInt64 {
			return 0
		}
		return int64(f)
	}
	switch v := value.(type) {
	case json.Number:
		if n, err := v.Int64(); err == nil {
			if n < 0 {
				return 0
			}
			return n
		}
		f, _ := v.Float64()
		return number(f)
	case float64:
		return number(v)
	case int64:
		if v < 0 {
			return 0
		}
		return v
	case int:
		if v < 0 {
			return 0
		}
		return int64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return number(f)
	default:
		return 0
	}
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func asObject(value any) map[string]any {
	if object, ok := value.(map[string]any); ok {
		return object
	}
	return nil
}

func asArray(value any) []any {
	if array, ok := value.([]any); ok {
		return array
	}
	return nil
}

func splitHostPort(value string) (string, int, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") {
		host, rest, found := strings.Cut(value[1:], "]")
		if !found {
			return "", 0, ratioErrorf("Invalid address.")
		}
		port := 443
		rest = strings.TrimPrefix(rest, ":")
		if rest != "" {
			parsed, err := strconv.Atoi(rest)
			if err != nil {
				return "", 0, ratioErrorf("Invalid port.")
			}
			port = parsed
		}
		return host, port, nil
	}
	if strings.Count(value, ":") > 1 && net.ParseIP(strings.Trim(value, "[]")) != nil {
		return strings.Trim(value, "[]"), 443, nil
	}
	index := strings.LastIndex(value, ":")
	if index < 0 {
		return value, 443, nil
	}
	host := value[:index]
	parsed, err := strconv.Atoi(value[index+1:])
	if err != nil {
		return "", 0, ratioErrorf("Invalid port.")
	}
	return host, parsed, nil
}

func tlsSettings(host string, query map[string]string, defaultFP string) map[string]any {
	serverName := query["sni"]
	if serverName == "" {
		serverName = query["peer"]
	}
	if serverName == "" {
		serverName = query["host"]
	}
	if serverName == "" {
		serverName = host
	}
	settings := map[string]any{"serverName": serverName}
	fingerprint := query["fp"]
	if fingerprint == "" {
		fingerprint = defaultFP
	}
	if fingerprint != "" {
		settings["fingerprint"] = fingerprint
	}
	if query["alpn"] != "" {
		var alpn []any
		for _, item := range strings.Split(query["alpn"], ",") {
			if item != "" {
				alpn = append(alpn, item)
			}
		}
		settings["alpn"] = alpn
	}
	if ParseBool(query["allowInsecure"]) || ParseBool(query["insecure"]) {
		settings["allowInsecure"] = true
	}
	pin := query["pcs"]
	if pin == "" {
		pin = query["pinSHA256"]
	}
	if pin != "" {
		settings["pinnedPeerCertSha256"] = pin
	}
	if query["vcn"] != "" {
		settings["verifyPeerCertByName"] = query["vcn"]
	}
	return settings
}

func applyTransport(stream map[string]any, network string, query map[string]string, host string) {
	network = strings.ToLower(network)
	if network == "" {
		network = "tcp"
	}
	switch network {
	case "splithttp", "xhttp":
		stream["network"] = "xhttp"
		settings := map[string]any{"path": defaultString(query["path"], "/")}
		if query["host"] != "" {
			settings["host"] = query["host"]
		}
		if query["mode"] != "" {
			settings["mode"] = query["mode"]
		}
		if query["extra"] != "" {
			decoded, err := url.PathUnescape(query["extra"])
			if err == nil {
				var extra map[string]any
				if json.Unmarshal([]byte(decoded), &extra) == nil {
					for key, value := range extra {
						settings[key] = value
					}
				}
			}
		}
		stream["xhttpSettings"] = settings
	case "httpupgrade":
		stream["network"] = "httpupgrade"
		settings := map[string]any{"path": defaultString(query["path"], "/")}
		if query["host"] != "" {
			settings["host"] = query["host"]
		}
		stream["httpupgradeSettings"] = settings
	case "ws":
		stream["network"] = "ws"
		settings := map[string]any{"path": defaultString(query["path"], "/")}
		if query["host"] != "" {
			settings["headers"] = map[string]any{"Host": query["host"]}
		}
		stream["wsSettings"] = settings
	case "grpc":
		stream["network"] = "grpc"
		service := query["serviceName"]
		if service == "" {
			service = query["path"]
		}
		settings := map[string]any{"serviceName": strings.TrimPrefix(service, "/")}
		if ParseBool(query["multiMode"]) {
			settings["multiMode"] = true
		}
		stream["grpcSettings"] = settings
	case "h2", "http":
		stream["network"] = "http"
		settings := map[string]any{"path": defaultString(query["path"], "/")}
		if query["host"] != "" {
			var hosts []any
			for _, item := range strings.Split(query["host"], ",") {
				if item != "" {
					hosts = append(hosts, item)
				}
			}
			settings["host"] = hosts
		}
		stream["httpSettings"] = settings
	case "kcp":
		stream["network"] = "kcp"
		header := map[string]any{"type": defaultString(query["headerType"], "none")}
		settings := map[string]any{"header": header}
		if query["seed"] != "" {
			settings["seed"] = query["seed"]
		}
		stream["kcpSettings"] = settings
	default:
		stream["network"] = "tcp"
		if strings.ToLower(query["headerType"]) == "http" {
			requestHost := query["host"]
			if requestHost == "" {
				requestHost = host
			}
			stream["tcpSettings"] = map[string]any{
				"header": map[string]any{
					"type": "http",
					"request": map[string]any{
						"path":    []any{defaultString(query["path"], "/")},
						"headers": map[string]any{"Host": []any{requestHost}},
					},
				},
			}
		}
	}
}

func applySecurity(stream map[string]any, security string, query map[string]string, host string) {
	security = strings.ToLower(security)
	if security == "" {
		security = "none"
	}
	switch security {
	case "tls":
		stream["security"] = "tls"
		stream["tlsSettings"] = tlsSettings(host, query, "chrome")
	case "reality":
		stream["security"] = "reality"
		serverName := query["sni"]
		if serverName == "" {
			serverName = query["host"]
		}
		if serverName == "" {
			serverName = host
		}
		fingerprint := query["fp"]
		if fingerprint == "" {
			fingerprint = "chrome"
		}
		reality := map[string]any{"serverName": serverName, "fingerprint": fingerprint}
		if query["pbk"] != "" {
			reality["publicKey"] = query["pbk"]
		}
		if query["sid"] != "" {
			reality["shortId"] = query["sid"]
		}
		if query["spx"] != "" {
			reality["spiderX"] = query["spx"]
		}
		stream["realitySettings"] = reality
	default:
		stream["security"] = "none"
	}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func queryMap(values url.Values) map[string]string {
	result := map[string]string{}
	for key, items := range values {
		if len(items) > 0 {
			result[key] = items[0]
		} else {
			result[key] = ""
		}
	}
	return result
}

func OutboundFromURI(uri string) (map[string]any, error) {
	scheme := strings.ToLower(strings.SplitN(uri, "://", 2)[0])
	switch scheme {
	case "vless":
		return outboundFromVLESS(uri)
	case "vmess":
		return outboundFromVMess(uri)
	case "trojan":
		return outboundFromTrojan(uri)
	case "ss":
		return outboundFromSS(uri)
	case "wireguard":
		return outboundFromWireGuard(uri)
	default:
		return nil, ratioErrorf("Unsupported protocol \"%s\". Use a vless/vmess/trojan/ss/wireguard link or a full JSON client config.", scheme)
	}
}

func outboundFromVLESS(uri string) (map[string]any, error) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Hostname() == "" {
		return nil, ratioErrorf("Invalid vless link.")
	}
	query := queryMap(parsed.Query())
	userID := ""
	if parsed.User != nil {
		userID = parsed.User.Username()
	}
	if userID == "" {
		return nil, ratioErrorf("No UUID found in the vless link.")
	}
	port := 443
	if parsed.Port() != "" {
		port, _ = strconv.Atoi(parsed.Port())
	}
	stream := map[string]any{}
	applySecurity(stream, defaultString(query["security"], "none"), query, parsed.Hostname())
	applyTransport(stream, defaultString(query["type"], "tcp"), query, parsed.Hostname())
	user := map[string]any{"id": userID, "encryption": defaultString(query["encryption"], "none")}
	if query["flow"] != "" {
		user["flow"] = query["flow"]
	}
	return map[string]any{
		"protocol": "vless",
		"settings": map[string]any{"vnext": []any{
			map[string]any{"address": parsed.Hostname(), "port": port, "users": []any{user}},
		}},
		"streamSettings": stream,
	}, nil
}

func decodeBase64Any(value string) ([]byte, error) {
	compact := strings.Join(strings.Fields(value), "")
	var lastErr error
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(compact)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	padded := compact + strings.Repeat("=", (4-len(compact)%4)%4)
	for _, encoding := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding} {
		decoded, err := encoding.DecodeString(padded)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func outboundFromVMess(uri string) (map[string]any, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, ratioErrorf("Could not decode the vmess link.")
	}
	raw := strings.TrimSpace(strings.SplitN(parts[1], "#", 2)[0])
	decoded, err := decodeBase64Any(raw)
	if err != nil {
		return nil, ratioErrorf("Could not decode the vmess link.")
	}
	var data map[string]any
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, ratioErrorf("Could not decode the vmess link.")
	}
	host := strings.TrimSpace(toString(data["add"]))
	if host == "" {
		return nil, ratioErrorf("No server address found in the vmess link.")
	}
	port := 443
	if value := toString(data["port"]); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			port = parsed
		}
	}
	userID := toString(data["id"])
	if userID == "" {
		return nil, ratioErrorf("No UUID found in the vmess link.")
	}
	query := map[string]string{
		"sni":         toString(data["sni"]),
		"host":        toString(data["host"]),
		"path":        defaultString(toString(data["path"]), "/"),
		"fp":          toString(data["fp"]),
		"alpn":        toString(data["alpn"]),
		"headerType":  defaultString(toString(data["type"]), "none"),
		"serviceName": toString(data["path"]),
	}
	security := "none"
	if value := strings.ToLower(toString(data["tls"])); value == "tls" || value == "reality" {
		security = "tls"
	}
	stream := map[string]any{}
	applySecurity(stream, security, query, host)
	applyTransport(stream, defaultString(toString(data["net"]), "tcp"), query, host)
	alterID := int(toInt64(data["aid"]))
	user := map[string]any{
		"id": userID, "alterId": alterID, "security": defaultString(toString(data["scy"]), "auto"),
	}
	return map[string]any{
		"protocol": "vmess",
		"settings": map[string]any{"vnext": []any{
			map[string]any{"address": host, "port": port, "users": []any{user}},
		}},
		"streamSettings": stream,
		"_remark":        toString(data["ps"]),
	}, nil
}

func outboundFromTrojan(uri string) (map[string]any, error) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Hostname() == "" {
		return nil, ratioErrorf("Invalid trojan link.")
	}
	password := ""
	if parsed.User != nil {
		password = parsed.User.Username()
	}
	if password == "" {
		return nil, ratioErrorf("No password found in the trojan link.")
	}
	query := queryMap(parsed.Query())
	port := 443
	if parsed.Port() != "" {
		port, _ = strconv.Atoi(parsed.Port())
	}
	stream := map[string]any{}
	applySecurity(stream, defaultString(query["security"], "tls"), query, parsed.Hostname())
	applyTransport(stream, defaultString(query["type"], "tcp"), query, parsed.Hostname())
	return map[string]any{
		"protocol": "trojan",
		"settings": map[string]any{"servers": []any{
			map[string]any{"address": parsed.Hostname(), "port": port, "password": password},
		}},
		"streamSettings": stream,
	}, nil
}

func outboundFromSS(uri string) (map[string]any, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, ratioErrorf("Could not decode the ss link.")
	}
	body := parts[1]
	if index := strings.Index(body, "#"); index >= 0 {
		body = body[:index]
	}
	if index := strings.Index(body, "?"); index >= 0 {
		body = body[:index]
	}
	var method, password, hostport string
	if index := strings.LastIndex(body, "@"); index >= 0 {
		userinfo := body[:index]
		hostport = body[index+1:]
		decoded, err := decodeBase64Any(userinfo)
		if err == nil && strings.Contains(string(decoded), ":") {
			pieces := strings.SplitN(string(decoded), ":", 2)
			method, password = pieces[0], pieces[1]
		} else {
			unescaped, _ := url.PathUnescape(userinfo)
			if !strings.Contains(unescaped, ":") {
				return nil, ratioErrorf("Invalid ss link.")
			}
			pieces := strings.SplitN(unescaped, ":", 2)
			method, password = pieces[0], pieces[1]
		}
	} else {
		decoded, err := decodeBase64Any(body)
		if err != nil {
			return nil, ratioErrorf("Could not decode the ss link.")
		}
		value := string(decoded)
		at := strings.LastIndex(value, "@")
		if at < 0 || !strings.Contains(value[:at], ":") {
			return nil, ratioErrorf("Invalid ss link.")
		}
		credentials := value[:at]
		hostport = value[at+1:]
		pieces := strings.SplitN(credentials, ":", 2)
		method, password = pieces[0], pieces[1]
	}
	host, port, err := splitHostPort(hostport)
	if err != nil {
		return nil, ratioErrorf("Invalid ss link.")
	}
	return map[string]any{
		"protocol": "shadowsocks",
		"settings": map[string]any{"servers": []any{
			map[string]any{
				"address": host, "port": port, "method": strings.TrimSpace(method), "password": password,
			},
		}},
		"streamSettings": map[string]any{"network": "tcp", "security": "none"},
	}, nil
}

func splitList(value string) []any {
	fields := strings.Fields(strings.NewReplacer(",", " ", ";", " ").Replace(value))
	result := make([]any, 0, len(fields))
	for _, field := range fields {
		result = append(result, field)
	}
	return result
}

func firstQueryValue(query map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := query[key]; value != "" {
			return value
		}
	}
	return ""
}

func outboundFromWireGuard(uri string) (map[string]any, error) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Hostname() == "" {
		return nil, ratioErrorf("Invalid wireguard link.")
	}
	query := map[string]string{}
	for key, values := range parsed.Query() {
		if len(values) > 0 {
			query[strings.ToLower(key)] = values[0]
		}
	}
	privateKey := ""
	if parsed.User != nil {
		privateKey = parsed.User.Username()
	}
	if privateKey == "" {
		privateKey = firstQueryValue(query, "privatekey", "private-key", "secretkey", "secret-key")
	}
	if privateKey == "" {
		return nil, ratioErrorf("No private key found in the wireguard link.")
	}
	publicKey := firstQueryValue(query, "publickey", "public-key", "peerpublickey", "peer-public-key")
	if publicKey == "" {
		return nil, ratioErrorf("No peer public key found in the wireguard link.")
	}
	address := splitList(firstQueryValue(query, "address", "addresses", "ip"))
	if len(address) == 0 {
		return nil, ratioErrorf("No address found in the wireguard link.")
	}
	allowedIPs := splitList(firstQueryValue(query, "allowedips", "allowed-ips", "allowedip", "allowed-ip"))
	if len(allowedIPs) == 0 {
		allowedIPs = []any{"0.0.0.0/0", "::/0"}
	}
	port := 51820
	if parsed.Port() != "" {
		port, err = strconv.Atoi(parsed.Port())
		if err != nil {
			return nil, ratioErrorf("Invalid wireguard port.")
		}
	}
	peer := map[string]any{
		"publicKey":  publicKey,
		"endpoint":   net.JoinHostPort(parsed.Hostname(), strconv.Itoa(port)),
		"allowedIPs": allowedIPs,
	}
	if psk := firstQueryValue(query, "presharedkey", "pre-shared-key", "psk"); psk != "" {
		peer["preSharedKey"] = psk
	}
	if keepAlive := toInt64(firstQueryValue(query, "keepalive", "keep-alive", "persistentkeepalive")); keepAlive > 0 {
		peer["keepAlive"] = keepAlive
	}
	settings := map[string]any{
		"secretKey":      privateKey,
		"address":        address,
		"peers":          []any{peer},
		"mtu":            int64(1420),
		"domainStrategy": "ForceIP",
	}
	if mtu := toInt64(query["mtu"]); mtu > 0 {
		settings["mtu"] = mtu
	}
	if dns := splitList(firstQueryValue(query, "dns", "remotedns", "remote-dns")); len(dns) > 0 {
		settings["remoteDNS"] = dns
	}
	return map[string]any{
		"protocol": "wireguard",
		"settings": settings,
	}, nil
}

func capRemark(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > 160 {
		return string(runes[:160]) + "…"
	}
	return value
}

func configRemark(text string, outbound map[string]any) string {
	if remark, ok := outbound["_remark"].(string); ok && remark != "" {
		return capRemark(remark)
	}
	if strings.Contains(text, "#") && strings.Contains(text, "://") {
		value := strings.SplitN(text, "#", 2)[1]
		unescaped, _ := url.PathUnescape(value)
		return capRemark(unescaped)
	}
	return ""
}

var claimedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:ضریب|coefficient|ratio|multiplier)\s*[:=]?\s*(\d+(?:[.,]\d+)?)`),
	regexp.MustCompile(`(?:^|[^\w.])(\d+(?:[.,]\d+)?)\s*[xX×](?:\b|$)`),
	regexp.MustCompile(`[xX×]\s*(\d+(?:[.,]\d+)?)(?:[^\w.]|$)`),
}

func DetectClaimed(text string, override *float64) (*float64, string) {
	if override != nil {
		return override, "cli"
	}
	unescaped, err := url.PathUnescape(text)
	if err != nil {
		unescaped = text
	}
	normalized := NormalizeDigits(unescaped)
	for _, pattern := range claimedPatterns {
		for _, match := range pattern.FindAllStringSubmatch(normalized, -1) {
			if len(match) < 2 {
				continue
			}
			value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
			if err == nil && value >= 0.1 && value <= 20 {
				return &value, "remark"
			}
		}
	}
	return nil, ""
}

func DecodeSubBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if strings.Contains(text, "://") {
		return text
	}
	compact := strings.Join(strings.Fields(text), "")
	if compact == "" {
		return ""
	}
	for _, variant := range []string{compact, strings.NewReplacer("-", "+", "_", "/").Replace(compact)} {
		decoded, err := decodeBase64Any(variant)
		if err == nil && strings.Contains(string(decoded), "://") {
			return string(decoded)
		}
	}
	return text
}

func OutboundHost(outbound map[string]any) string {
	settings := asObject(outbound["settings"])
	if settings == nil {
		return ""
	}
	if vnext := asArray(settings["vnext"]); len(vnext) > 0 {
		if server := asObject(vnext[0]); server != nil {
			return toString(server["address"])
		}
	}
	if servers := asArray(settings["servers"]); len(servers) > 0 {
		if server := asObject(servers[0]); server != nil {
			return toString(server["address"])
		}
	}
	if peers := asArray(settings["peers"]); len(peers) > 0 {
		if peer := asObject(peers[0]); peer != nil {
			if endpoint := toString(peer["endpoint"]); endpoint != "" {
				if host, _, err := splitHostPort(endpoint); err == nil {
					return host
				}
			}
		}
	}
	return ""
}

func LooksLikeIP(host string) bool {
	value := strings.Trim(strings.TrimSpace(host), "[]")
	return net.ParseIP(value) != nil
}

// OutboundServerName returns the TLS/WS host of an outbound, which can be used
// as an IPv4-reachable alternative when the address itself is IPv6-only.
func OutboundServerName(outbound map[string]any) string {
	stream := asObject(outbound["streamSettings"])
	if stream == nil {
		return ""
	}
	for _, key := range []string{"tlsSettings", "realitySettings"} {
		if settings := asObject(stream[key]); settings != nil {
			if name := strings.TrimSpace(toString(settings["serverName"])); name != "" && !LooksLikeIP(name) {
				return name
			}
		}
	}
	for _, key := range []string{"wsSettings", "xhttpSettings", "httpupgradeSettings"} {
		if settings := asObject(stream[key]); settings != nil {
			if name := firstCSVHost(toString(settings["host"])); name != "" && !LooksLikeIP(name) {
				return name
			}
			if headers := asObject(settings["headers"]); headers != nil {
				if name := firstCSVHost(toString(headers["Host"])); name != "" && !LooksLikeIP(name) {
					return name
				}
			}
		}
	}
	if settings := asObject(stream["grpcSettings"]); settings != nil {
		if name := strings.TrimSpace(toString(settings["authority"])); name != "" && !LooksLikeIP(name) {
			return name
		}
	}
	return ""
}

func firstCSVHost(value string) string {
	if index := strings.IndexByte(value, ','); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

// SetOutboundHost replaces the server address of a vnext/servers outbound.
func SetOutboundHost(outbound map[string]any, host string) bool {
	settings := asObject(outbound["settings"])
	if settings == nil {
		return false
	}
	for _, key := range []string{"vnext", "servers"} {
		if servers := asArray(settings[key]); len(servers) > 0 {
			if server := asObject(servers[0]); server != nil {
				server["address"] = host
				return true
			}
		}
	}
	return false
}

func SetOutboundServerName(outbound map[string]any, name string) {
	stream := asObject(outbound["streamSettings"])
	if stream == nil {
		return
	}
	for _, key := range []string{"tlsSettings", "realitySettings"} {
		if settings := asObject(stream[key]); settings != nil {
			current := strings.TrimSpace(toString(settings["serverName"]))
			if current == "" || LooksLikeIP(current) {
				settings["serverName"] = name
			}
		}
	}
}

func HasIPv4Host(host string) bool {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.To4() != nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return false
	}
	for _, item := range ips {
		if item.IP.To4() != nil {
			return true
		}
	}
	return false
}

// IPv4FallbackHost returns a domain from the config that resolves over IPv4.
func IPv4FallbackHost(outbound map[string]any) string {
	name := OutboundServerName(outbound)
	if name == "" || LooksLikeIP(name) || !HasIPv4Host(name) {
		return ""
	}
	if safeFetchEnabled() && !IsPublicHost(name) {
		return ""
	}
	return name
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func IsUsableHost(host string) bool {
	value := strings.Trim(strings.TrimSpace(host), "[]")
	if value == "" {
		return false
	}
	if strings.EqualFold(value, "localhost") || LooksLikeIP(value) {
		return true
	}
	return strings.Contains(value, ".")
}

func OutboundTarget(outbound map[string]any) string {
	settings := asObject(outbound["settings"])
	if settings == nil {
		return ""
	}
	if vnext := asArray(settings["vnext"]); len(vnext) > 0 {
		if server := asObject(vnext[0]); server != nil {
			return serverTarget(server)
		}
	}
	if servers := asArray(settings["servers"]); len(servers) > 0 {
		if server := asObject(servers[0]); server != nil {
			return serverTarget(server)
		}
	}
	if peers := asArray(settings["peers"]); len(peers) > 0 {
		if peer := asObject(peers[0]); peer != nil {
			return strings.TrimSpace(toString(peer["endpoint"]))
		}
	}
	return ""
}

func serverTarget(server map[string]any) string {
	host := strings.Trim(toString(server["address"]), "[]")
	port := toString(server["port"])
	if host == "" || port == "" {
		return ""
	}
	return net.JoinHostPort(host, port)
}

func zeroUUID(outbound map[string]any) bool {
	settings := asObject(outbound["settings"])
	if settings == nil {
		return false
	}
	var id string
	if vnext := asArray(settings["vnext"]); len(vnext) > 0 {
		if server := asObject(vnext[0]); server != nil {
			if users := asArray(server["users"]); len(users) > 0 {
				if user := asObject(users[0]); user != nil {
					id = toString(user["id"])
				}
			}
		}
	}
	id = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(id)), "-", "")
	return id != "" && strings.Trim(id, "0") == ""
}

func isPlaceholderConfig(outbound map[string]any, remark string) bool {
	if outbound == nil {
		return false
	}
	host := strings.Trim(OutboundHost(outbound), "[]")
	if placeholderHostPattern.MatchString(host) {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && !isPublicIP(ip) {
		return true
	}
	if zeroUUID(outbound) {
		return true
	}
	return placeholderRemarkPattern.MatchString(remark) && !outboundHasTLS(outbound)
}

func ExtractConfigs(body []byte) []Config {
	text := DecodeSubBody(body)
	var lines []any
	stripped := strings.TrimSpace(text)
	if strings.HasPrefix(stripped, "[") || strings.HasPrefix(stripped, "{") {
		var data any
		decoder := json.NewDecoder(strings.NewReader(stripped))
		decoder.UseNumber()
		if err := decoder.Decode(&data); err == nil {
			var candidates []any
			if array, ok := data.([]any); ok {
				candidates = array
			} else if object, ok := data.(map[string]any); ok {
				if outbounds, ok := object["outbounds"].([]any); ok {
					candidates = []any{object}
					_ = outbounds
				} else if nested, ok := object["data"].(map[string]any); ok {
					if outbounds, ok := nested["outbounds"].([]any); ok {
						candidates = []any{nested}
						_ = outbounds
					}
				} else {
					for _, key := range []string{"configs", "links", "proxies", "data"} {
						if array, ok := object[key].([]any); ok {
							candidates = array
							break
						}
					}
				}
			}
			for _, item := range candidates {
				if value, ok := item.(string); ok {
					lines = append(lines, value)
					continue
				}
				object := asObject(item)
				if object == nil {
					continue
				}
				if _, ok := object["outbounds"].([]any); ok {
					lines = append(lines, object)
					continue
				}
				for _, key := range []string{"url", "link", "config"} {
					if value, ok := object[key].(string); ok {
						lines = append(lines, value)
						break
					}
				}
			}
		}
	}
	if len(lines) == 0 {
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(line, "://") {
				lines = append(lines, line)
			}
		}
	}
	var configs []Config
	seen := map[string]bool{}
	for _, item := range lines {
		if object, ok := item.(map[string]any); ok {
			outbounds := asArray(object["outbounds"])
			var proxy map[string]any
			for _, candidate := range outbounds {
				entry := asObject(candidate)
				if entry != nil && !DirectProtocols[toString(entry["protocol"])] {
					proxy = entry
					break
				}
			}
			if proxy == nil || !IsUsableHost(OutboundHost(proxy)) {
				continue
			}
			remark := toString(object["remarks"])
			if remark == "" {
				remark = toString(proxy["tag"])
			}
			if isPlaceholderConfig(proxy, remark) {
				continue
			}
			protocol := toString(proxy["protocol"])
			if protocol == "" {
				protocol = "json"
			}
			configs = append(configs, Config{
				JSON: object, Protocol: protocol, Target: OutboundTarget(proxy), Remark: remark,
			})
			continue
		}
		line, ok := item.(string)
		if !ok {
			continue
		}
		if seen[line] {
			continue
		}
		scheme := strings.ToLower(strings.SplitN(line, "://", 2)[0])
		if scheme != "vless" && scheme != "vmess" && scheme != "trojan" && scheme != "ss" && scheme != "wireguard" {
			continue
		}
		outbound, err := OutboundFromURI(line)
		if err != nil || !IsUsableHost(OutboundHost(outbound)) {
			continue
		}
		remark := configRemark(line, outbound)
		if isPlaceholderConfig(outbound, remark) {
			continue
		}
		seen[line] = true
		configs = append(configs, Config{
			URI: line, Protocol: scheme, Target: OutboundTarget(outbound),
			Remark: remark,
		})
	}
	return configs
}

func EnsureResolvable(host string) error {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" || strings.EqualFold(host, "localhost") || LooksLikeIP(host) {
		return nil
	}
	if _, err := net.LookupHost(host); err != nil {
		return ratioErrorf("Cannot resolve host '%s'. The config may be a placeholder or unavailable.", host)
	}
	return nil
}

// IPv6OnlyHost reports whether the config target can only be reached over IPv6.
func IPv6OnlyHost(host string) bool {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.To4() == nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, item := range ips {
		if item.IP.To4() != nil {
			return false
		}
	}
	return true
}

// HasGlobalIPv6 reports whether this host has a globally routable IPv6 address.
func HasGlobalIPv6() bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		if ip.To4() == nil && ip.IsGlobalUnicast() && !ip.IsPrivate() {
			return true
		}
	}
	return false
}
