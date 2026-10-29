package ratio

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ConfigURI returns a shareable URI for a parsed config. URI subscriptions
// already have one; JSON subscriptions are serialized back to the closest
// supported URI form.
func ConfigURI(cfg Config) string {
	if cfg.URI != "" {
		return cfg.URI
	}
	if cfg.JSON == nil {
		return ""
	}
	for _, value := range asArray(cfg.JSON["outbounds"]) {
		outbound := asObject(value)
		if outbound == nil {
			continue
		}
		switch toString(outbound["protocol"]) {
		case "vless":
			if uri := vlessConfigURI(outbound, cfg.Remark); uri != "" {
				return uri
			}
		case "wireguard":
			if uri := wireGuardConfigURI(outbound, cfg.Remark); uri != "" {
				return uri
			}
		case "vmess":
			if uri := vmessConfigURI(outbound, cfg.Remark); uri != "" {
				return uri
			}
		}
	}
	return ""
}

func firstObject(values []any) map[string]any {
	for _, value := range values {
		if object := asObject(value); object != nil {
			return object
		}
	}
	return nil
}

func joinValues(values []any) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if text := toString(value); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, ",")
}

func hostPort(host, port string) string {
	if port == "" {
		return host
	}
	if strings.Contains(host, ":") {
		return net.JoinHostPort(host, port)
	}
	return host + ":" + port
}

func appendTLSParams(query url.Values, stream map[string]any) {
	tlsSettings := asObject(stream["tlsSettings"])
	if tlsSettings != nil {
		if value := toString(tlsSettings["serverName"]); value != "" {
			query.Set("sni", value)
		}
		if value := toString(tlsSettings["fingerprint"]); value != "" {
			query.Set("fp", value)
		}
		if value := toString(tlsSettings["pinnedPeerCertSha256"]); value != "" {
			query.Set("pcs", value)
		}
		if value := toString(tlsSettings["verifyPeerCertByName"]); value != "" {
			query.Set("vcn", value)
		}
		alpn := joinValues(asArray(tlsSettings["alpn"]))
		if alpn == "" {
			alpn = toString(tlsSettings["alpn"])
		}
		if alpn != "" {
			query.Set("alpn", alpn)
		}
		allow := false
		switch value := tlsSettings["allowInsecure"].(type) {
		case bool:
			allow = value
		case string:
			allow = value == "1" || value == "true"
		case float64:
			allow = value != 0
		}
		if allow {
			query.Set("allowInsecure", "1")
		} else {
			query.Set("allowInsecure", "0")
		}
	}
	reality := asObject(stream["realitySettings"])
	if reality != nil {
		if value := toString(reality["serverName"]); value != "" {
			query.Set("sni", value)
		}
		if value := toString(reality["fingerprint"]); value != "" {
			query.Set("fp", value)
		}
		if value := toString(reality["publicKey"]); value != "" {
			query.Set("pbk", value)
		}
		if value := toString(reality["shortId"]); value != "" {
			query.Set("sid", value)
		}
		if value := toString(reality["spiderX"]); value != "" {
			query.Set("spx", value)
		}
	}
}

func appendTransportParams(query url.Values, stream map[string]any) {
	network := toString(stream["network"])
	if network == "" {
		network = "tcp"
	}
	query.Set("type", network)
	query.Set("security", toString(stream["security"]))
	appendTLSParams(query, stream)
	if _, ok := query["allowInsecure"]; !ok {
		query.Set("allowInsecure", "0")
	}
	if _, ok := query["insecure"]; !ok {
		query.Set("insecure", "0")
	}
	if network == "tcp" {
		headerType := ""
		if tcp := asObject(stream["tcpSettings"]); tcp != nil {
			if header := asObject(tcp["header"]); header != nil {
				headerType = toString(header["type"])
			}
		}
		query.Set("headerType", headerType)
	}

	if xhttp := asObject(stream["xhttpSettings"]); xhttp != nil {
		if value := toString(xhttp["path"]); value != "" {
			query.Set("path", value)
		}
		if value := toString(xhttp["host"]); value != "" {
			query.Set("host", value)
		}
		if value := toString(xhttp["mode"]); value != "" {
			query.Set("mode", value)
		}
		if extra := asObject(xhttp["extra"]); extra != nil {
			for key, value := range extra {
				text := toString(value)
				if text == "" && value != nil {
					text = fmt.Sprint(value)
				}
				if text != "" {
					query.Set(key, text)
				}
			}
		}
	}

	if ws := asObject(stream["wsSettings"]); ws != nil {
		if value := toString(ws["path"]); value != "" {
			query.Set("path", value)
		}
		if headers := asObject(ws["headers"]); headers != nil {
			if value := toString(headers["Host"]); value != "" {
				query.Set("host", value)
			}
		}
	}

	if httpupgrade := asObject(stream["httpupgradeSettings"]); httpupgrade != nil {
		if value := toString(httpupgrade["path"]); value != "" {
			query.Set("path", value)
		}
		if value := toString(httpupgrade["host"]); value != "" {
			query.Set("host", value)
		}
	}

	if grpc := asObject(stream["grpcSettings"]); grpc != nil {
		if value := toString(grpc["serviceName"]); value != "" {
			query.Set("serviceName", value)
		}
	}
}

func vlessConfigURI(outbound map[string]any, remark string) string {
	settings := asObject(outbound["settings"])
	server := firstObject(asArray(settings["vnext"]))
	user := firstObject(asArray(server["users"]))
	if server == nil || user == nil {
		return ""
	}
	host := toString(server["address"])
	id := toString(user["id"])
	if host == "" || id == "" {
		return ""
	}
	port := toString(server["port"])
	query := url.Values{}
	encryption := toString(user["encryption"])
	if encryption == "" {
		encryption = "none"
	}
	query.Set("encryption", encryption)
	if flow := toString(user["flow"]); flow != "" {
		query.Set("flow", flow)
	}
	appendTransportParams(query, asObject(outbound["streamSettings"]))
	return fmt.Sprintf("vless://%s@%s?%s#%s",
		url.User(id).String(),
		hostPort(host, port),
		query.Encode(),
		url.PathEscape(remark),
	)
}

func wireGuardConfigURI(outbound map[string]any, remark string) string {
	settings := asObject(outbound["settings"])
	peer := firstObject(asArray(settings["peers"]))
	if peer == nil {
		return ""
	}
	privateKey := toString(settings["secretKey"])
	endpoint := toString(peer["endpoint"])
	publicKey := toString(peer["publicKey"])
	address := joinValues(asArray(settings["address"]))
	if privateKey == "" || endpoint == "" || publicKey == "" || address == "" {
		return ""
	}
	query := url.Values{}
	query.Set("publickey", publicKey)
	query.Set("address", address)
	if value := joinValues(asArray(peer["allowedIPs"])); value != "" {
		query.Set("allowedips", value)
	}
	if value := toInt64(settings["mtu"]); value > 0 {
		query.Set("mtu", strconv.FormatInt(value, 10))
	}
	if value := toInt64(peer["keepAlive"]); value > 0 {
		query.Set("keepalive", strconv.FormatInt(value, 10))
	}
	if value := joinValues(asArray(settings["remoteDNS"])); value != "" {
		query.Set("dns", value)
	}
	return fmt.Sprintf("wireguard://%s@%s/?%s#%s",
		url.User(privateKey).String(),
		endpoint,
		query.Encode(),
		url.PathEscape(remark),
	)
}

func vmessConfigURI(outbound map[string]any, remark string) string {
	settings := asObject(outbound["settings"])
	server := firstObject(asArray(settings["vnext"]))
	if server == nil {
		return ""
	}
	user := firstObject(asArray(server["users"]))
	host := toString(server["address"])
	id := toString(user["id"])
	if host == "" || id == "" {
		return ""
	}
	port := toString(server["port"])
	if port == "" {
		port = "443"
	}

	stream := asObject(outbound["streamSettings"])
	network := toString(stream["network"])
	if network == "" {
		network = "tcp"
	}
	security := strings.ToLower(toString(stream["security"]))
	tls := ""
	if security == "tls" || security == "reality" {
		tls = "tls"
	}
	sni, fp, alpn, hostHeader, path := "", "", "", "", ""
	if tlsSettings := asObject(stream["tlsSettings"]); tlsSettings != nil {
		sni = toString(tlsSettings["serverName"])
		fp = toString(tlsSettings["fingerprint"])
		alpn = joinValues(asArray(tlsSettings["alpn"]))
	}
	if reality := asObject(stream["realitySettings"]); reality != nil {
		if sni == "" {
			sni = toString(reality["serverName"])
		}
		if fp == "" {
			fp = toString(reality["fingerprint"])
		}
	}
	if xhttp := asObject(stream["xhttpSettings"]); xhttp != nil {
		path = toString(xhttp["path"])
		hostHeader = toString(xhttp["host"])
	}
	if httpupgrade := asObject(stream["httpupgradeSettings"]); httpupgrade != nil {
		if path == "" {
			path = toString(httpupgrade["path"])
		}
		if hostHeader == "" {
			hostHeader = toString(httpupgrade["host"])
		}
	}
	if ws := asObject(stream["wsSettings"]); ws != nil {
		if path == "" {
			path = toString(ws["path"])
		}
		if hostHeader == "" {
			if headers := asObject(ws["headers"]); headers != nil {
				hostHeader = toString(headers["Host"])
			}
		}
	}
	headerType := "none"
	if tcp := asObject(stream["tcpSettings"]); tcp != nil {
		if header := asObject(tcp["header"]); header != nil {
			if value := toString(header["type"]); value != "" {
				headerType = value
			}
		}
	}
	scy := toString(user["security"])
	if scy == "" {
		scy = "auto"
	}
	payload := map[string]any{
		"v": "2", "ps": remark, "add": host, "port": port, "id": id,
		"aid": toInt64(user["alterId"]), "scy": scy, "net": network,
		"type": headerType, "host": hostHeader, "path": path, "tls": tls,
		"sni": sni, "alpn": alpn, "fp": fp,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(encoded)
}
