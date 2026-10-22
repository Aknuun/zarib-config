import base64
import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from ratio_test import (  # noqa: E402
    RatioError,
    detect_claimed,
    outbound_from_uri,
    parse_usage,
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

    def test_override(self):
        self.assertEqual(detect_claimed("ضریب 2", 1.2), (1.2, "cli"))


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


if __name__ == "__main__":
    unittest.main()
