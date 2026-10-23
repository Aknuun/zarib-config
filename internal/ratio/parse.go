package ratio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Usage struct {
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    *int64 `json:"total,omitempty"`
	Expire   *int64 `json:"expire,omitempty"`
}

func (u Usage) Used() int64 { return u.Upload + u.Download }

func (u Usage) Remaining() *int64 {
	if u.Total == nil {
		return nil
	}
	value := *u.Total - u.Upload - u.Download
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

func Fetch(rawURL string, timeout time.Duration) ([]byte, http.Header, error) {
	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, ratioErrorf("Invalid subscription URL: %v", err)
	}
	request.Header.Set("User-Agent", UA)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Pragma", "no-cache")
	client := &http.Client{Timeout: timeout}
	if safeFetchEnabled() {
		if !IsPublicURL(rawURL) {
			return nil, nil, ratioErrorf("Blocked: the subscription URL must be a public http(s) address.")
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
			return nil, nil, ratioErrorf("Subscription request timed out.")
		}
		if strings.Contains(strings.ToLower(err.Error()), "blocked") {
			return nil, nil, ratioErrorf("Blocked a redirect to a non-public address.")
		}
		return nil, nil, ratioErrorf("Could not connect to the subscription URL: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, nil, ratioErrorf("Subscription request failed with HTTP %d.", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxSubBytes+1))
	if err != nil {
		return nil, nil, ratioErrorf("Could not read the subscription response: %v", err)
	}
	if len(body) > MaxSubBytes {
		return nil, nil, ratioErrorf("Subscription response is too large.")
	}
	return body, response.Header, nil
}

var integerPattern = regexp.MustCompile(`^-?\d+$`)

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
	switch v := value.(type) {
	case json.Number:
		n, _ := v.Int64()
		if n == 0 {
			f, _ := v.Float64()
			return int64(f)
		}
		return n
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return int64(f)
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
	default:
		return nil, ratioErrorf("Unsupported protocol \"%s\". Use a vless/vmess/trojan/ss link or a full JSON client config.", scheme)
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

func configRemark(text string, outbound map[string]any) string {
	if remark, ok := outbound["_remark"].(string); ok && remark != "" {
		return remark
	}
	if strings.Contains(text, "#") && strings.Contains(text, "://") {
		value := strings.SplitN(text, "#", 2)[1]
		unescaped, _ := url.PathUnescape(value)
		return strings.TrimSpace(unescaped)
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
	return ""
}

func LooksLikeIP(host string) bool {
	value := strings.Trim(strings.TrimSpace(host), "[]")
	return net.ParseIP(value) != nil
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
		if scheme != "vless" && scheme != "vmess" && scheme != "trojan" && scheme != "ss" {
			continue
		}
		outbound, err := OutboundFromURI(line)
		if err != nil || !IsUsableHost(OutboundHost(outbound)) {
			continue
		}
		seen[line] = true
		configs = append(configs, Config{
			URI: line, Protocol: scheme, Target: OutboundTarget(outbound),
			Remark: configRemark(line, outbound),
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
