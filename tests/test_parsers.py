import base64
import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from ratio_test import (  # noqa: E402
    RatioError,
    detect_claimed,
    extract_configs,
    is_usable_host,
    outbound_from_uri,
    parse_usage,
    prepare_config,
)


class UsageTests(unittest.TestCase):
    def test_subscription_header(self):
        headers = {
            "Subscription-Userinfo": "upload=1024; download=2048; total=10737418240; expire=1893456000"
        }
        usage = parse_usage(headers, b"")
        self.assertEqual(usage["upload"], 1024)
        self.assertEqual(usage["download"], 2048)
        self.assertEqual(usage["total"], 10737418240)

    def test_json_body_fallback(self):
        usage = parse_usage({}, b'{"upload": 10, "download": 20}')
        self.assertEqual(usage, {"upload": 10, "download": 20})

    def test_missing_usage(self):
        self.assertIsNone(parse_usage({}, b"hello"))


class ClaimTests(unittest.TestCase):
    def test_persian_remark(self):
        value, source = detect_claimed("کانفیگ ضریب ۱.۵", None)
        self.assertEqual(value, 1.5)
        self.assertEqual(source, "remark")

    def test_percent_encoded_remark(self):
        uri = (
            "vless://uuid@example.com:443?type=ws#"
            "%D8%B6%D8%B1%DB%8C%D8%A8%201.5x"
        )
        self.assertEqual(detect_claimed(uri, None)[0], 1.5)

    def test_implausible_ratio_is_ignored(self):
        self.assertIsNone(detect_claimed("port 201.5x", None)[0])

    def test_multiplier_prefix(self):
        self.assertEqual(detect_claimed("ضریب ×۱.۲", None)[0], 1.2)

    def test_override(self):
        self.assertEqual(detect_claimed("ضریب 2", 1.2), (1.2, "cli"))


class SubscriptionExtractionTests(unittest.TestCase):
    def test_base64_body(self):
        links = (
            "vless://11111111-2222-3333-4444-555555555555@example.com:443?type=ws#node-a\n"
            "trojan://secret@example.net:443#node-b\n"
        )
        configs = extract_configs(base64.b64encode(links.encode()))
        self.assertEqual(len(configs), 2)
        self.assertEqual(configs[0]["protocol"], "vless")
        self.assertEqual(configs[0]["target"], "example.com:443")
        self.assertEqual(configs[1]["remark"], "node-b")

    def test_plain_text_body(self):
        body = b"# comment\nss://YWVzLTI1Ni1nY206cGFzcw==@example.com:8388#ss\n"
        configs = extract_configs(body)
        self.assertEqual(len(configs), 1)
        self.assertEqual(configs[0]["protocol"], "ss")

    def test_json_body(self):
        body = json.dumps({"configs": ["vless://11111111-2222-3333-4444-555555555555@example.com:443#one"]}).encode()
        configs = extract_configs(body)
        self.assertEqual(len(configs), 1)

    def test_full_json_configs(self):
        config = {
            "remarks": "node ×1.2",
            "inbounds": [{"protocol": "socks", "port": 10808}],
            "outbounds": [
                {"protocol": "vless", "tag": "proxy", "settings": {"vnext": [{"address": "example.com", "port": 443}]}},
                {"protocol": "freedom", "tag": "DIRECT"},
            ],
        }
        configs = extract_configs(json.dumps([config]).encode())
        self.assertEqual(len(configs), 1)
        self.assertEqual(configs[0]["protocol"], "vless")
        self.assertEqual(configs[0]["target"], "example.com:443")
        self.assertEqual(configs[0]["remark"], "node ×1.2")
        self.assertIsNotNone(configs[0]["config"])

    def test_unsupported_scheme_is_skipped(self):
        configs = extract_configs(b"hysteria2://secret@example.com:443#hy")
        self.assertEqual(configs, [])


class UriTests(unittest.TestCase):
    def test_vless_reality_ws(self):
        uri = (
            "vless://11111111-2222-3333-4444-555555555555@example.com:443"
            "?type=ws&security=reality&sni=cdn.example.com&pbk=abc&sid=12"
            "&fp=chrome&host=cdn.example.com&path=%2Fws&flow=xtls-rprx-vision"
        )
        out = outbound_from_uri(uri)
        self.assertEqual(out["protocol"], "vless")
        user = out["settings"]["vnext"][0]["users"][0]
        self.assertEqual(user["flow"], "xtls-rprx-vision")
        stream = out["streamSettings"]
        self.assertEqual(stream["network"], "ws")
        self.assertEqual(stream["security"], "reality")
        self.assertEqual(stream["realitySettings"]["publicKey"], "abc")
        self.assertEqual(stream["wsSettings"]["path"], "/ws")
        self.assertEqual(stream["wsSettings"]["headers"]["Host"], "cdn.example.com")

    def test_vless_grpc_tls(self):
        uri = (
            "vless://11111111-2222-3333-4444-555555555555@example.com:443"
            "?type=grpc&security=tls&sni=example.com&serviceName=svc"
        )
        stream = outbound_from_uri(uri)["streamSettings"]
        self.assertEqual(stream["network"], "grpc")
        self.assertEqual(stream["grpcSettings"]["serviceName"], "svc")
        self.assertEqual(stream["tlsSettings"]["serverName"], "example.com")

    def test_vmess(self):
        payload = {
            "v": "2",
            "ps": "test 1.2x",
            "add": "example.com",
            "port": "443",
            "id": "11111111-2222-3333-4444-555555555555",
            "aid": "0",
            "scy": "auto",
            "net": "ws",
            "type": "none",
            "host": "example.com",
            "path": "/ws",
            "tls": "tls",
            "sni": "example.com",
        }
        encoded = base64.b64encode(json.dumps(payload).encode()).decode()
        out = outbound_from_uri("vmess://" + encoded)
        self.assertEqual(out["protocol"], "vmess")
        self.assertEqual(out["streamSettings"]["network"], "ws")
        self.assertEqual(out["_remark"], "test 1.2x")

    def test_trojan(self):
        uri = "trojan://secret@example.com:443?type=ws&security=tls&sni=example.com&path=%2Ft"
        out = outbound_from_uri(uri)
        self.assertEqual(out["protocol"], "trojan")
        self.assertEqual(out["settings"]["servers"][0]["password"], "secret")

    def test_ss_sip002(self):
        userinfo = base64.b64encode(b"aes-256-gcm:secret").decode()
        out = outbound_from_uri(f"ss://{userinfo}@example.com:8388#node")
        server = out["settings"]["servers"][0]
        self.assertEqual(server["method"], "aes-256-gcm")
        self.assertEqual(server["password"], "secret")
        self.assertEqual(server["port"], 8388)

    def test_ss_legacy(self):
        encoded = base64.b64encode(b"aes-256-gcm:secret@example.com:8388").decode()
        server = outbound_from_uri("ss://" + encoded)["settings"]["servers"][0]
        self.assertEqual(server["address"], "example.com")

    def test_unsupported(self):
        with self.assertRaises(RatioError):
            outbound_from_uri("hysteria2://secret@example.com:443")


class ConfigPreparationTests(unittest.TestCase):
    def test_placeholder_host_rejected(self):
        self.assertFalse(is_usable_host("ParsConnection"))
        self.assertFalse(is_usable_host(""))
        self.assertTrue(is_usable_host("example.com"))
        self.assertTrue(is_usable_host("1.2.3.4"))
        self.assertTrue(is_usable_host("::1"))

    def test_extract_skips_placeholder_host(self):
        body = base64.b64encode(
            b"vless://uuid@ParsConnection:1010?encryption=none&type=tcp#info\n"
            b"vless://uuid@example.com:443?encryption=none&security=tls&type=tcp#real"
        )
        configs = extract_configs(body)
        self.assertEqual(len(configs), 1)
        self.assertIn("example.com", configs[0]["target"])

    def test_json_routing_is_replaced(self):
        base = {
            "inbounds": [{"port": 10808, "protocol": "socks"}],
            "outbounds": [
                {"protocol": "vless", "tag": "proxy"},
                {"protocol": "freedom", "tag": "direct"},
            ],
            "routing": {
                "rules": [
                    {"type": "field", "ip": ["geoip:ir"], "outboundTag": "direct"}
                ]
            },
        }
        config = prepare_config(base, None, "proxy", 12345)
        self.assertEqual(len(config["inbounds"]), 1)
        self.assertEqual(config["inbounds"][0]["port"], 12345)
        self.assertEqual(len(config["routing"]["rules"]), 1)
        self.assertNotIn("geoip", json.dumps(config["routing"]))
        self.assertEqual(base["routing"]["rules"][0]["ip"], ["geoip:ir"])


if __name__ == "__main__":
    unittest.main()
