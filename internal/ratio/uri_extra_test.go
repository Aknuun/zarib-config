package ratio

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestVlessConfigURIPinnedCertAndXHTTP(t *testing.T) {
	cfg := Config{
		Remark: "pinned",
		JSON: map[string]any{"outbounds": []any{map[string]any{
			"protocol": "vless",
			"settings": map[string]any{"vnext": []any{map[string]any{
				"address": "fastly.videobazi.com",
				"port":    443,
				"users": []any{map[string]any{
					"id":         "42a1c069-9dc9-4181-86b8-13bf28632851",
					"encryption": "none",
				}},
			}}},
			"streamSettings": map[string]any{
				"network":  "xhttp",
				"security": "tls",
				"tlsSettings": map[string]any{
					"serverName":           "community.fastly.steamstatic.com",
					"alpn":                 []any{"h2"},
					"fingerprint":          "qq",
					"pinnedPeerCertSha256": "cd6e838b9bfe31cab8b3d5c858b0d6223fe84ece236c1fdb281be93c302e1e6e",
					"verifyPeerCertByName": "community.fastly.steamstatic.com",
				},
				"xhttpSettings": map[string]any{"path": "/", "host": "tv42.global.ssl.fastly.net", "mode": "auto"},
			},
		}}},
	}
	parsed, err := url.Parse(ConfigURI(cfg))
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("pcs") != "cd6e838b9bfe31cab8b3d5c858b0d6223fe84ece236c1fdb281be93c302e1e6e" {
		t.Fatalf("missing pcs: %s", parsed.String())
	}
	if query.Get("vcn") != "community.fastly.steamstatic.com" {
		t.Fatalf("missing vcn: %s", parsed.String())
	}
	if query.Get("type") != "xhttp" || query.Get("mode") != "auto" || query.Get("path") != "/" {
		t.Fatalf("invalid xhttp params: %s", parsed.String())
	}
}

func TestVlessConfigURIHTTPUpgrade(t *testing.T) {
	cfg := Config{
		Remark: "upgrade",
		JSON: map[string]any{"outbounds": []any{map[string]any{
			"protocol": "vless",
			"settings": map[string]any{"vnext": []any{map[string]any{
				"address": "dl1cdn.pcapps.ir",
				"port":    443,
				"users": []any{map[string]any{
					"id": "42a1c069-9dc9-4181-86b8-13bf28632851",
				}},
			}}},
			"streamSettings": map[string]any{
				"network":             "httpupgrade",
				"security":            "none",
				"httpupgradeSettings": map[string]any{"path": "/?ed=2560", "host": "dl1cdn.pcapps.ir"},
			},
		}}},
	}
	parsed, err := url.Parse(ConfigURI(cfg))
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("type") != "httpupgrade" || query.Get("path") != "/?ed=2560" || query.Get("host") != "dl1cdn.pcapps.ir" {
		t.Fatalf("invalid httpupgrade params: %s", parsed.String())
	}
}

func TestVMessConfigURI(t *testing.T) {
	cfg := Config{
		Remark: "vmess-test",
		JSON: map[string]any{"outbounds": []any{map[string]any{
			"protocol": "vmess",
			"settings": map[string]any{"vnext": []any{map[string]any{
				"address": "tv1333.pcapps.ir",
				"port":    443,
				"users": []any{map[string]any{
					"id":       "6f5a3f9f-0e6d-4ab5-9f5d-2b7e1a8c9d10",
					"alterId":  0,
					"security": "auto",
				}},
			}}},
			"streamSettings": map[string]any{
				"network":             "httpupgrade",
				"security":            "none",
				"httpupgradeSettings": map[string]any{"path": "/", "host": "tv1333.pcapps.ir"},
			},
		}}},
	}
	got := ConfigURI(cfg)
	if !strings.HasPrefix(got, "vmess://") {
		t.Fatalf("not vmess uri: %s", got)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["add"] != "tv1333.pcapps.ir" || payload["net"] != "httpupgrade" || payload["host"] != "tv1333.pcapps.ir" || payload["path"] != "/" || payload["ps"] != "vmess-test" {
		t.Fatalf("invalid vmess payload: %#v", payload)
	}
}
