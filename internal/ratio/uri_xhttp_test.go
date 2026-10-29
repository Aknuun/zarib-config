package ratio

import (
	"net/url"
	"testing"
)

func TestVlessConfigURIXHTTPTLS(t *testing.T) {
	cfg := Config{
		Remark: "🇩🇪 𝔷 | ‍𝙳𝙮𝙨𝙡𝙶𝙹 | xhttp",
		JSON: map[string]any{
			"outbounds": []any{
				map[string]any{
					"protocol": "vless",
					"settings": map[string]any{
						"vnext": []any{
							map[string]any{
								"address": "fastly.videobazi.com",
								"port":    443,
								"users": []any{
									map[string]any{
										"id":         "42a1c069-9dc9-4181-86b8-13bf28632851",
										"encryption": "none",
									},
								},
							},
						},
					},
					"streamSettings": map[string]any{
						"network":  "xhttp",
						"security": "tls",
						"tlsSettings": map[string]any{
							"serverName":    "community.fastly.steamstatic.com",
							"alpn":          []any{"h2"},
							"fingerprint":   "qq",
							"allowInsecure": false,
						},
						"xhttpSettings": map[string]any{
							"path": "/",
							"host": "tv42.global.ssl.fastly.net",
							"mode": "auto",
						},
					},
				},
			},
		},
	}

	got := ConfigURI(cfg)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("ConfigURI returned invalid URL %q: %v", got, err)
	}
	if parsed.Scheme != "vless" {
		t.Fatalf("scheme = %q, want vless", parsed.Scheme)
	}
	if parsed.Host != "fastly.videobazi.com:443" {
		t.Fatalf("host = %q", parsed.Host)
	}
	q := parsed.Query()
	want := map[string]string{
		"type":          "xhttp",
		"security":      "tls",
		"alpn":          "h2",
		"encryption":    "none",
		"insecure":      "0",
		"host":          "tv42.global.ssl.fastly.net",
		"fp":            "qq",
		"allowInsecure": "0",
		"sni":           "community.fastly.steamstatic.com",
		"mode":          "auto",
		"path":          "/",
	}
	for key, value := range want {
		if got := q.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
	if fragment := parsed.Fragment; fragment != cfg.Remark {
		t.Errorf("fragment = %q, want %q", fragment, cfg.Remark)
	}
}
