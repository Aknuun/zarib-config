package ratio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseUsageHeader(t *testing.T) {
	headers := http.Header{"Subscription-Userinfo": []string{"upload=1024; download=2048; total=10737418240; expire=1893456000"}}
	usage := ParseUsage(headers, nil)
	if usage == nil {
		t.Fatal("expected usage")
	}
	if usage.Upload != 1024 || usage.Download != 2048 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	if usage.Total == nil || *usage.Total != 10737418240 {
		t.Fatalf("unexpected total: %+v", usage.Total)
	}
}

func TestParseUsageJSONBody(t *testing.T) {
	usage := ParseUsage(http.Header{}, []byte(`{"upload": 10, "download": 20}`))
	if usage == nil || usage.Upload != 10 || usage.Download != 20 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestParseUsageMissing(t *testing.T) {
	if usage := ParseUsage(http.Header{}, []byte("hello")); usage != nil {
		t.Fatalf("expected nil, got %+v", usage)
	}
}

func TestDetectClaimedPersian(t *testing.T) {
	value, source := DetectClaimed("کانفیگ ضریب ۱.۵", nil)
	if value == nil || *value != 1.5 || source != "remark" {
		t.Fatalf("unexpected claim: %v %s", value, source)
	}
}

func TestDetectClaimedPercentEncoded(t *testing.T) {
	uri := "vless://uuid@example.com:443?type=ws#%D8%B6%D8%B1%DB%8C%D8%A8%201.5x"
	value, _ := DetectClaimed(uri, nil)
	if value == nil || *value != 1.5 {
		t.Fatalf("unexpected claim: %v", value)
	}
}

func TestDetectClaimedImplausible(t *testing.T) {
	value, _ := DetectClaimed("port 201.5x", nil)
	if value != nil {
		t.Fatalf("expected nil, got %v", *value)
	}
}

func TestDetectClaimedMultiplierPrefix(t *testing.T) {
	value, _ := DetectClaimed("ضریب ×۱.۲", nil)
	if value == nil || *value != 1.2 {
		t.Fatalf("unexpected claim: %v", value)
	}
}

func TestDetectClaimedOverride(t *testing.T) {
	override := 1.2
	value, source := DetectClaimed("ضریب 2", &override)
	if value == nil || *value != 1.2 || source != "cli" {
		t.Fatalf("unexpected claim: %v %s", value, source)
	}
}

func TestExtractBase64Body(t *testing.T) {
	links := "vless://11111111-2222-3333-4444-555555555555@example.com:443?type=ws#node-a\n" +
		"trojan://secret@example.net:443#node-b\n"
	body := base64.StdEncoding.EncodeToString([]byte(links))
	configs := ExtractConfigs([]byte(body))
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
	if configs[0].Protocol != "vless" || configs[0].Target != "example.com:443" {
		t.Fatalf("unexpected config: %+v", configs[0])
	}
	if configs[1].Remark != "node-b" {
		t.Fatalf("unexpected remark: %q", configs[1].Remark)
	}
}

func TestExtractPlainText(t *testing.T) {
	body := []byte("# comment\nss://YWVzLTI1Ni1nY206cGFzcw==@example.com:8388#ss\n")
	configs := ExtractConfigs(body)
	if len(configs) != 1 || configs[0].Protocol != "ss" {
		t.Fatalf("unexpected configs: %+v", configs)
	}
}

func TestExtractJSONBody(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"configs": []string{"vless://11111111-2222-3333-4444-555555555555@example.com:443#one"},
	})
	configs := ExtractConfigs(body)
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
}

func TestExtractFullJSONConfigs(t *testing.T) {
	config := map[string]any{
		"remarks":  "node ×1.2",
		"inbounds": []any{map[string]any{"protocol": "socks", "port": 10808}},
		"outbounds": []any{
			map[string]any{
				"protocol": "vless", "tag": "proxy",
				"settings": map[string]any{"vnext": []any{
					map[string]any{"address": "example.com", "port": 443},
				}},
			},
			map[string]any{"protocol": "freedom", "tag": "DIRECT"},
		},
	}
	body, _ := json.Marshal([]any{config})
	configs := ExtractConfigs(body)
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Protocol != "vless" || configs[0].Target != "example.com:443" {
		t.Fatalf("unexpected config: %+v", configs[0])
	}
	if configs[0].Remark != "node ×1.2" || configs[0].JSON == nil {
		t.Fatalf("unexpected config: %+v", configs[0])
	}
}

func TestExtractSingleTopLevelJSONConfig(t *testing.T) {
	config := map[string]any{
		"outbounds": []any{
			map[string]any{
				"protocol": "trojan", "tag": "proxy",
				"settings": map[string]any{"servers": []any{
					map[string]any{"address": "example.net", "port": 443},
				}},
			},
			map[string]any{"protocol": "freedom", "tag": "DIRECT"},
		},
	}
	body, _ := json.Marshal(config)
	configs := ExtractConfigs(body)
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Protocol != "trojan" || configs[0].Target != "example.net:443" {
		t.Fatalf("unexpected config: %+v", configs[0])
	}
	if configs[0].Remark != "proxy" || configs[0].JSON == nil {
		t.Fatalf("unexpected config: %+v", configs[0])
	}
}

func TestExtractUnsupportedScheme(t *testing.T) {
	if configs := ExtractConfigs([]byte("hysteria2://secret@example.com:443#hy")); len(configs) != 0 {
		t.Fatalf("expected no configs, got %+v", configs)
	}
}

func TestVLESSRealityWS(t *testing.T) {
	uri := "vless://11111111-2222-3333-4444-555555555555@example.com:443" +
		"?type=ws&security=reality&sni=cdn.example.com&pbk=abc&sid=12" +
		"&fp=chrome&host=cdn.example.com&path=%2Fws&flow=xtls-rprx-vision"
	outbound, err := OutboundFromURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	if outbound["protocol"] != "vless" {
		t.Fatalf("unexpected protocol: %v", outbound["protocol"])
	}
	settings := mustObject(t, outbound["settings"])
	vnext := mustArray(t, settings["vnext"])
	users := mustArray(t, mustObject(t, vnext[0])["users"])
	user := mustObject(t, users[0])
	if user["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected flow: %v", user["flow"])
	}
	stream := mustObject(t, outbound["streamSettings"])
	if stream["network"] != "ws" || stream["security"] != "reality" {
		t.Fatalf("unexpected stream: %+v", stream)
	}
	reality := mustObject(t, stream["realitySettings"])
	if reality["publicKey"] != "abc" {
		t.Fatalf("unexpected reality: %+v", reality)
	}
	ws := mustObject(t, stream["wsSettings"])
	if ws["path"] != "/ws" {
		t.Fatalf("unexpected path: %v", ws["path"])
	}
	if mustObject(t, ws["headers"])["Host"] != "cdn.example.com" {
		t.Fatalf("unexpected host: %+v", ws["headers"])
	}
}

func TestVLESSGRPCTLS(t *testing.T) {
	uri := "vless://11111111-2222-3333-4444-555555555555@example.com:443" +
		"?type=grpc&security=tls&sni=example.com&serviceName=svc"
	outbound, err := OutboundFromURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	stream := mustObject(t, outbound["streamSettings"])
	if stream["network"] != "grpc" {
		t.Fatalf("unexpected network: %v", stream["network"])
	}
	if mustObject(t, stream["grpcSettings"])["serviceName"] != "svc" {
		t.Fatalf("unexpected grpc settings: %+v", stream["grpcSettings"])
	}
	if mustObject(t, stream["tlsSettings"])["serverName"] != "example.com" {
		t.Fatalf("unexpected tls settings: %+v", stream["tlsSettings"])
	}
}

func TestVMess(t *testing.T) {
	payload := map[string]any{
		"v": "2", "ps": "test 1.2x", "add": "example.com", "port": "443",
		"id": "11111111-2222-3333-4444-555555555555", "aid": "0", "scy": "auto",
		"net": "ws", "type": "none", "host": "example.com", "path": "/ws",
		"tls": "tls", "sni": "example.com",
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(mustJSON(t, payload)))
	outbound, err := OutboundFromURI("vmess://" + encoded)
	if err != nil {
		t.Fatal(err)
	}
	if outbound["protocol"] != "vmess" {
		t.Fatalf("unexpected protocol: %v", outbound["protocol"])
	}
	stream := mustObject(t, outbound["streamSettings"])
	if stream["network"] != "ws" {
		t.Fatalf("unexpected network: %v", stream["network"])
	}
	if outbound["_remark"] != "test 1.2x" {
		t.Fatalf("unexpected remark: %v", outbound["_remark"])
	}
}

func TestTrojan(t *testing.T) {
	uri := "trojan://secret@example.com:443?type=ws&security=tls&sni=example.com&path=%2Ft"
	outbound, err := OutboundFromURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	if outbound["protocol"] != "trojan" {
		t.Fatalf("unexpected protocol: %v", outbound["protocol"])
	}
	settings := mustObject(t, outbound["settings"])
	servers := mustArray(t, settings["servers"])
	if mustObject(t, servers[0])["password"] != "secret" {
		t.Fatalf("unexpected server: %+v", servers[0])
	}
}

func TestSSSIP002(t *testing.T) {
	userinfo := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:secret"))
	outbound, err := OutboundFromURI("ss://" + userinfo + "@example.com:8388#node")
	if err != nil {
		t.Fatal(err)
	}
	settings := mustObject(t, outbound["settings"])
	server := mustObject(t, mustArray(t, settings["servers"])[0])
	if server["method"] != "aes-256-gcm" || server["password"] != "secret" {
		t.Fatalf("unexpected server: %+v", server)
	}
	if toInt64(server["port"]) != 8388 {
		t.Fatalf("unexpected port: %v", server["port"])
	}
}

func TestSSLegacy(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:secret@example.com:8388"))
	outbound, err := OutboundFromURI("ss://" + encoded)
	if err != nil {
		t.Fatal(err)
	}
	settings := mustObject(t, outbound["settings"])
	server := mustObject(t, mustArray(t, settings["servers"])[0])
	if server["address"] != "example.com" {
		t.Fatalf("unexpected address: %v", server["address"])
	}
}

func TestUnsupportedURI(t *testing.T) {
	if _, err := OutboundFromURI("hysteria2://secret@example.com:443"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestPlaceholderHost(t *testing.T) {
	if IsUsableHost("ParsConnection") || IsUsableHost("") {
		t.Fatal("placeholder host should be rejected")
	}
	for _, host := range []string{"example.com", "1.2.3.4", "::1"} {
		if !IsUsableHost(host) {
			t.Fatalf("expected usable host: %s", host)
		}
	}
}

func TestExtractSkipsPlaceholderHost(t *testing.T) {
	body := base64.StdEncoding.EncodeToString([]byte(
		"vless://uuid@ParsConnection:1010?encryption=none&type=tcp#info\n" +
			"vless://uuid@example.com:443?encryption=none&security=tls&type=tcp#real"))
	configs := ExtractConfigs([]byte(body))
	if len(configs) != 1 || !strings.Contains(configs[0].Target, "example.com") {
		t.Fatalf("unexpected configs: %+v", configs)
	}
}

func TestJSONRoutingIsReplaced(t *testing.T) {
	base := map[string]any{
		"inbounds": []any{map[string]any{"port": 10808, "protocol": "socks"}},
		"outbounds": []any{
			map[string]any{"protocol": "vless", "tag": "proxy"},
			map[string]any{"protocol": "freedom", "tag": "direct"},
		},
		"routing": map[string]any{
			"rules": []any{
				map[string]any{"type": "field", "ip": []any{"geoip:ir"}, "outboundTag": "direct"},
			},
		},
	}
	config, err := PrepareConfig(base, nil, "proxy", 12345)
	if err != nil {
		t.Fatal(err)
	}
	inbounds := mustArray(t, config["inbounds"])
	if len(inbounds) != 1 || toInt64(mustObject(t, inbounds[0])["port"]) != 12345 {
		t.Fatalf("unexpected inbounds: %+v", inbounds)
	}
	routing := mustObject(t, config["routing"])
	rules := mustArray(t, routing["rules"])
	if len(rules) != 1 {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	serialized, _ := json.Marshal(routing)
	if strings.Contains(string(serialized), "geoip") {
		t.Fatalf("routing was not replaced: %s", serialized)
	}
	original := mustObject(t, base["routing"])
	originalRules := mustArray(t, original["rules"])
	if mustArray(t, mustObject(t, originalRules[0])["ip"])[0] != "geoip:ir" {
		t.Fatal("base config was modified")
	}
}

func TestWaitForUsageRetriesUntilChange(t *testing.T) {
	oldCountdown := countdownFn
	oldRead := readUsageAfterFn
	defer func() {
		countdownFn = oldCountdown
		readUsageAfterFn = oldRead
	}()
	countdownFn = func(int, string) {}
	unchanged := &Usage{}
	changed := &Usage{Download: 10}
	calls := 0
	readUsageAfterFn = func(context.Context, string, Usage, int) (*Usage, []usageReading, time.Time, error) {
		calls++
		if calls == 1 {
			return unchanged, nil, time.Unix(1, 0), nil
		}
		return changed, nil, time.Unix(2, 0), nil
	}
	after, _, _, err := WaitForUsage(context.Background(), "https://example.com/sub", Usage{}, 20, 60, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if after.Download != 10 || calls != 2 {
		t.Fatalf("unexpected result: %+v calls=%d", after, calls)
	}
}

func TestWaitForUsageRaisesAfterAllRounds(t *testing.T) {
	oldCountdown := countdownFn
	oldRead := readUsageAfterFn
	defer func() {
		countdownFn = oldCountdown
		readUsageAfterFn = oldRead
	}()
	countdownFn = func(int, string) {}
	readUsageAfterFn = func(context.Context, string, Usage, int) (*Usage, []usageReading, time.Time, error) {
		return &Usage{}, nil, time.Unix(1, 0), nil
	}
	if _, _, _, err := WaitForUsage(context.Background(), "https://example.com/sub", Usage{}, 20, 60, 3, false); err == nil {
		t.Fatal("expected an error")
	}
}

func TestProbeRetriesThenSucceeds(t *testing.T) {
	oldProbe := probeFn
	oldCountdown := countdownFn
	defer func() {
		probeFn = oldProbe
		countdownFn = oldCountdown
	}()
	countdownFn = func(int, string) {}
	calls := 0
	probeFn = func(context.Context, int, []string, int) (int64, error) {
		calls++
		if calls == 1 {
			return 0, &ProbeFailure{Message: "reset"}
		}
		return 1, nil
	}
	if err := ProbeWithRetry(context.Background(), 1080, []string{"https://example.com"}, 5, 1, 0, false); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func mustObject(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}
	return object
}

func mustArray(t *testing.T, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", value)
	}
	return array
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
