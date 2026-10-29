package ratio

import (
	"net/url"
	"strings"
	"testing"
)

func TestVlessConfigURIPreservesTransportFields(t *testing.T) {
	cfg := Config{
		JSON: map[string]any{
			"outbounds": []any{
				map[string]any{
					"protocol": "vless",
					"settings": map[string]any{
						"vnext": []any{
							map[string]any{
								"address": "dl2.videobazi.com",
								"port":    float64(2043),
								"users": []any{
									map[string]any{
										"id":         "42a1c069-9dc9-4181-86b8-13bf28632851",
										"encryption": "mlkem768x25519plus.random.0rtt.test",
									},
								},
							},
						},
					},
					"streamSettings": map[string]any{
						"network":  "tcp",
						"security": "",
						"tcpSettings": map[string]any{
							"header": map[string]any{"type": ""},
						},
					},
				},
			},
		},
		Remark: "🇫🇮 𝔹 | 𝙩 | 🎧",
	}

	raw := ConfigURI(cfg)
	if !strings.HasPrefix(raw, "vless://42a1c069-9dc9-4181-86b8-13bf28632851@dl2.videobazi.com:2043?") {
		t.Fatalf("unexpected prefix: %s", raw)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	query := parsed.Query()
	for _, key := range []string{"security", "headerType", "type", "encryption"} {
		if !query.Has(key) {
			t.Errorf("missing query key %q in %s", key, raw)
		}
	}
	if got := query.Get("type"); got != "tcp" {
		t.Errorf("type = %q, want tcp", got)
	}
	if got := query.Get("security"); got != "" {
		t.Errorf("security = %q, want empty", got)
	}
	if got := query.Get("headerType"); got != "" {
		t.Errorf("headerType = %q, want empty", got)
	}
	if got := query.Get("encryption"); got != "mlkem768x25519plus.random.0rtt.test" {
		t.Errorf("encryption = %q", got)
	}
	if parsed.Fragment != cfg.Remark {
		t.Errorf("fragment = %q, want %q", parsed.Fragment, cfg.Remark)
	}
}
