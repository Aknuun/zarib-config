#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ratio_test.py - measure the real traffic multiplier (coefficient) of a V2Ray config.

It downloads a known amount of data through the config and compares it with the
usage reported by the subscription link (subscription-userinfo header).
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import zipfile
from datetime import datetime, timezone
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import parse_qs, unquote, urlparse
from urllib.request import Request, urlopen

VERSION = "1.0.0"
UA = "v2rayNG/1.9.16"
MIN_SIZE_MB = 10.0
DEFAULT_SIZE_MB = 10.0
DEFAULT_WAIT = 20
DEFAULT_POLL_TIMEOUT = 60
DEFAULT_CHUNK_MB = 5.0
DEFAULT_TEST_URL = "https://speed.cloudflare.com/__down?bytes={bytes}"
DEFAULT_PROBE_URL = "https://speed.cloudflare.com/__down?bytes=4096"
DEFAULT_PROBE_TIMEOUT = 20
PROBE_URLS = [
    DEFAULT_PROBE_URL,
    "https://www.google.com/generate_204",
    "http://cp.cloudflare.com/generate_204",
]
CACHE_DIR = Path(os.environ.get("RATIO_TEST_CACHE", str(Path.home() / ".cache" / "ratio-test")))
PROXY_PROTOCOLS = {"vless", "vmess", "trojan", "shadowsocks", "wireguard", "socks", "http"}
DIRECT_PROTOCOLS = {"freedom", "blackhole", "dns"}

FA_DIGITS = str.maketrans("۰۱۲۳۴۵۶۷۸۹٠١٢٣٤٥٦٧٨٩", "01234567890123456789")


class RatioError(Exception):
    pass


def normalize_digits(value: str) -> str:
    return value.translate(FA_DIGITS)


def human(n) -> str:
    if n is None:
        return "-"
    n = float(n)
    units = ["B", "KB", "MB", "GB", "TB", "PB"]
    idx = 0
    while abs(n) >= 1024 and idx < len(units) - 1:
        n /= 1024.0
        idx += 1
    if idx == 0:
        return f"{n:.0f} B"
    return f"{n:.2f} {units[idx]}"


def ask(prompt: str, default=None) -> str:
    suffix = f" [{default}]" if default not in (None, "") else ""
    try:
        value = input(f"{prompt}{suffix}: ").strip()
    except EOFError:
        raise RatioError("Input closed.")
    if not value and default is not None:
        return str(default)
    return value


def parse_bool(value: str) -> bool:
    return str(value).strip().lower() in ("1", "true", "yes", "on")


def fetch(url: str, timeout: int = 30):
    headers = {
        "User-Agent": UA,
        "Accept": "*/*",
        "Cache-Control": "no-cache",
        "Pragma": "no-cache",
    }
    req = Request(url, headers=headers)
    try:
        with urlopen(req, timeout=timeout) as resp:
            return resp.read(), dict(resp.headers)
    except HTTPError as exc:
        raise RatioError(f"Subscription request failed with HTTP {exc.code}.")
    except URLError as exc:
        raise RatioError(f"Could not connect to the subscription URL: {exc.reason}")
    except TimeoutError:
        raise RatioError("Subscription request timed out.")


def parse_usage(headers, body: bytes):
    raw = None
    for key, value in headers.items():
        if key.lower() == "subscription-userinfo":
            raw = value
            break
    data = {}
    if raw:
        for part in raw.split(";"):
            if "=" not in part:
                continue
            key, value = part.split("=", 1)
            key = key.strip().lower()
            value = value.strip()
            if re.fullmatch(r"-?\d+", value):
                data[key] = int(value)
    if "upload" not in data and "download" not in data:
        try:
            payload = json.loads(body.decode("utf-8", "replace"))
            if isinstance(payload, dict) and ("upload" in payload or "download" in payload):
                for key in ("upload", "download", "total", "expire"):
                    if key in payload:
                        data[key] = int(float(payload[key]))
        except (ValueError, TypeError, UnicodeDecodeError):
            pass
    if "upload" not in data and "download" not in data:
        return None
    data.setdefault("upload", 0)
    data.setdefault("download", 0)
    return data


def read_usage(sub_url: str, retries: int = 3, delay: float = 3.0):
    last_error = None
    for attempt in range(retries):
        try:
            body, headers = fetch(sub_url)
            usage = parse_usage(headers, body)
            if usage is not None:
                return usage
            last_error = RatioError(
                "The subscription has no subscription-userinfo header and no usage in the body."
            )
        except RatioError as exc:
            last_error = exc
        if attempt < retries - 1:
            time.sleep(delay)
    raise last_error or RatioError("Could not read subscription usage.")


def split_host_port(value: str):
    value = value.strip()
    if value.startswith("["):
        host, _, rest = value[1:].partition("]")
        port = rest.lstrip(":")
        return host, int(port) if port else 443
    host, sep, port = value.rpartition(":")
    if not sep:
        return value, 443
    return host, int(port)


def tls_settings(host: str, query: dict, default_fp: str = "chrome"):
    server_name = query.get("sni") or query.get("peer") or query.get("host") or host
    settings = {"serverName": server_name}
    fingerprint = query.get("fp") or default_fp
    if fingerprint:
        settings["fingerprint"] = fingerprint
    if query.get("alpn"):
        settings["alpn"] = [item for item in query["alpn"].split(",") if item]
    if parse_bool(query.get("allowInsecure") or query.get("insecure") or ""):
        settings["allowInsecure"] = True
    return settings


def apply_transport(stream: dict, network: str, query: dict, host: str):
    network = (network or "tcp").lower()
    if network in ("splithttp", "xhttp"):
        stream["network"] = "xhttp"
        settings = {"path": query.get("path") or "/"}
        if query.get("host"):
            settings["host"] = query["host"]
        if query.get("mode"):
            settings["mode"] = query["mode"]
        if query.get("extra"):
            try:
                extra = json.loads(unquote(query["extra"]))
                if isinstance(extra, dict):
                    settings.update(extra)
            except ValueError:
                pass
        stream["xhttpSettings"] = settings
    elif network == "httpupgrade":
        stream["network"] = "httpupgrade"
        settings = {"path": query.get("path") or "/"}
        if query.get("host"):
            settings["host"] = query["host"]
        stream["httpupgradeSettings"] = settings
    elif network == "ws":
        stream["network"] = "ws"
        settings = {"path": query.get("path") or "/"}
        if query.get("host"):
            settings["headers"] = {"Host": query["host"]}
        stream["wsSettings"] = settings
    elif network == "grpc":
        stream["network"] = "grpc"
        service = query.get("serviceName") or query.get("path") or ""
        stream["grpcSettings"] = {"serviceName": service.lstrip("/")}
        if parse_bool(query.get("multiMode") or ""):
            stream["grpcSettings"]["multiMode"] = True
    elif network in ("h2", "http"):
        stream["network"] = "http"
        settings = {"path": query.get("path") or "/"}
        if query.get("host"):
            settings["host"] = [item for item in query["host"].split(",") if item]
        stream["httpSettings"] = settings
    elif network == "kcp":
        stream["network"] = "kcp"
        header = {"type": query.get("headerType") or "none"}
        stream["kcpSettings"] = {"header": header}
        if query.get("seed"):
            stream["kcpSettings"]["seed"] = query["seed"]
    else:
        stream["network"] = "tcp"
        if (query.get("headerType") or "").lower() == "http":
            stream["tcpSettings"] = {
                "header": {
                    "type": "http",
                    "request": {
                        "path": [query.get("path") or "/"],
                        "headers": {"Host": [query.get("host") or host]},
                    },
                }
            }


def apply_security(stream: dict, security: str, query: dict, host: str):
    security = (security or "none").lower()
    if security == "tls":
        stream["security"] = "tls"
        stream["tlsSettings"] = tls_settings(host, query)
    elif security == "reality":
        stream["security"] = "reality"
        reality = {
            "serverName": query.get("sni") or query.get("host") or host,
            "fingerprint": query.get("fp") or "chrome",
        }
        if query.get("pbk"):
            reality["publicKey"] = query["pbk"]
        if query.get("sid"):
            reality["shortId"] = query["sid"]
        if query.get("spx"):
            reality["spiderX"] = query["spx"]
        stream["realitySettings"] = reality
    else:
        stream["security"] = "none"


def outbound_from_vless(uri: str):
    parsed = urlparse(uri)
    if not parsed.hostname:
        raise RatioError("Invalid vless link.")
    query = {key: (values[0] if values else "") for key, values in parse_qs(parsed.query, keep_blank_values=True).items()}
    user_id = unquote(parsed.username or "")
    if not user_id:
        raise RatioError("No UUID found in the vless link.")
    port = parsed.port or 443
    stream = {}
    apply_security(stream, query.get("security") or "none", query, parsed.hostname)
    apply_transport(stream, query.get("type") or "tcp", query, parsed.hostname)
    user = {"id": user_id, "encryption": query.get("encryption") or "none"}
    if query.get("flow"):
        user["flow"] = query["flow"]
    return {
        "protocol": "vless",
        "settings": {"vnext": [{"address": parsed.hostname, "port": port, "users": [user]}]},
        "streamSettings": stream,
    }


def outbound_from_vmess(uri: str):
    raw = uri.split("://", 1)[1].split("#", 1)[0].strip()
    raw += "=" * (-len(raw) % 4)
    try:
        data = json.loads(base64.b64decode(raw).decode("utf-8"))
    except (ValueError, UnicodeDecodeError):
        raise RatioError("Could not decode the vmess link.")
    host = str(data.get("add") or "").strip()
    if not host:
        raise RatioError("No server address found in the vmess link.")
    port = int(data.get("port") or 443)
    user_id = str(data.get("id") or "")
    if not user_id:
        raise RatioError("No UUID found in the vmess link.")
    query = {
        "sni": str(data.get("sni") or ""),
        "host": str(data.get("host") or ""),
        "path": str(data.get("path") or "/"),
        "fp": str(data.get("fp") or ""),
        "alpn": str(data.get("alpn") or ""),
        "headerType": str(data.get("type") or "none"),
        "serviceName": str(data.get("path") or ""),
    }
    stream = {}
    apply_security(stream, "tls" if str(data.get("tls") or "").lower() in ("tls", "reality") else "none", query, host)
    apply_transport(stream, str(data.get("net") or "tcp"), query, host)
    user = {"id": user_id, "alterId": int(data.get("aid") or 0), "security": str(data.get("scy") or "auto")}
    return {
        "protocol": "vmess",
        "settings": {"vnext": [{"address": host, "port": port, "users": [user]}]},
        "streamSettings": stream,
        "_remark": str(data.get("ps") or ""),
    }


def outbound_from_trojan(uri: str):
    parsed = urlparse(uri)
    if not parsed.hostname:
        raise RatioError("Invalid trojan link.")
    password = unquote(parsed.username or "")
    if not password:
        raise RatioError("No password found in the trojan link.")
    query = {key: (values[0] if values else "") for key, values in parse_qs(parsed.query, keep_blank_values=True).items()}
    port = parsed.port or 443
    stream = {}
    apply_security(stream, query.get("security") or "tls", query, parsed.hostname)
    apply_transport(stream, query.get("type") or "tcp", query, parsed.hostname)
    return {
        "protocol": "trojan",
        "settings": {"servers": [{"address": parsed.hostname, "port": port, "password": password}]},
        "streamSettings": stream,
    }


def outbound_from_ss(uri: str):
    body = uri.split("://", 1)[1]
    if "#" in body:
        body = body.split("#", 1)[0]
    if "?" in body:
        body = body.split("?", 1)[0]
    if "@" in body:
        userinfo, hostport = body.rsplit("@", 1)
        try:
            decoded = base64.b64decode(userinfo + "=" * (-len(userinfo) % 4)).decode("utf-8")
            if ":" not in decoded:
                decoded = unquote(userinfo)
        except (ValueError, UnicodeDecodeError):
            decoded = unquote(userinfo)
        if ":" not in decoded:
            raise RatioError("Invalid ss link.")
        method, password = decoded.split(":", 1)
    else:
        try:
            decoded = base64.b64decode(body + "=" * (-len(body) % 4)).decode("utf-8")
        except (ValueError, UnicodeDecodeError):
            raise RatioError("Could not decode the ss link.")
        if "@" not in decoded or ":" not in decoded:
            raise RatioError("Invalid ss link.")
        credentials, hostport = decoded.rsplit("@", 1)
        method, password = credentials.split(":", 1)
    host, port = split_host_port(hostport)
    return {
        "protocol": "shadowsocks",
        "settings": {
            "servers": [
                {"address": host, "port": port, "method": method.strip(), "password": password}
            ]
        },
        "streamSettings": {"network": "tcp", "security": "none"},
    }


def outbound_from_uri(uri: str):
    scheme = uri.split("://", 1)[0].lower()
    builders = {
        "vless": outbound_from_vless,
        "vmess": outbound_from_vmess,
        "trojan": outbound_from_trojan,
        "ss": outbound_from_ss,
    }
    if scheme not in builders:
        raise RatioError(
            f"Unsupported protocol \"{scheme}\". Use a vless/vmess/trojan/ss link or a full JSON client config."
        )
    return builders[scheme](uri)


def config_remark(text: str, outbound: dict) -> str:
    if outbound.get("_remark"):
        return outbound["_remark"]
    if "#" in text and "://" in text:
        return unquote(text.split("#", 1)[1]).strip()
    return ""


def detect_claimed(text: str, override):
    if override is not None:
        return float(override), "cli"
    normalized = normalize_digits(unquote(text))
    patterns = [
        r"(?:ضریب|coefficient|ratio|multiplier)\s*[:=]?\s*(\d+(?:[.,]\d+)?)",
        r"(?<![\w.])(\d+(?:[.,]\d+)?)\s*[xX×](?![\w])",
        r"[xX×]\s*(\d+(?:[.,]\d+)?)(?![\w.])",
    ]
    for pattern in patterns:
        for match in re.finditer(pattern, normalized):
            value = float(match.group(1).replace(",", "."))
            if 0.1 <= value <= 20:
                return value, "remark"
    return None, None


SUPPORTED_SCHEMES = ("vless", "vmess", "trojan", "ss")


def decode_sub_body(body: bytes) -> str:
    text = body.decode("utf-8", "replace").strip()
    if "://" in text:
        return text
    compact = re.sub(r"\s+", "", text)
    if not compact:
        return ""
    variants = [compact, compact.replace("-", "+").replace("_", "/")]
    for variant in variants:
        padded = variant + "=" * (-len(variant) % 4)
        try:
            decoded = base64.b64decode(padded).decode("utf-8", "replace")
        except ValueError:
            continue
        if "://" in decoded:
            return decoded
    return text


def outbound_target(outbound) -> str:
    settings = outbound.get("settings") or {}
    if settings.get("vnext"):
        server = settings["vnext"][0]
        return f"{server.get('address')}:{server.get('port')}"
    if settings.get("servers"):
        server = settings["servers"][0]
        return f"{server.get('address')}:{server.get('port')}"
    return ""


def extract_configs(body: bytes):
    text = decode_sub_body(body)
    lines = []
    stripped = text.strip()
    if stripped.startswith("[") or stripped.startswith("{"):
        try:
            data = json.loads(stripped)
        except ValueError:
            data = None
        candidates = []
        if isinstance(data, list):
            candidates = data
        elif isinstance(data, dict):
            for key in ("configs", "links", "proxies", "data"):
                if isinstance(data.get(key), list):
                    candidates = data[key]
                    break
        for item in candidates:
            if isinstance(item, str):
                lines.append(item)
            elif isinstance(item, dict) and isinstance(item.get("outbounds"), list):
                lines.append(item)
            elif isinstance(item, dict):
                for key in ("url", "link", "config"):
                    if isinstance(item.get(key), str):
                        lines.append(item[key])
                        break
    if not lines:
        for line in text.splitlines():
            line = line.strip()
            if line and not line.startswith("#") and "://" in line:
                lines.append(line)
    configs = []
    seen = set()
    for line in lines:
        if isinstance(line, dict):
            config = line
            outbounds = config.get("outbounds") or []
            proxy = next(
                (
                    item
                    for item in outbounds
                    if isinstance(item, dict) and item.get("protocol") not in DIRECT_PROTOCOLS
                ),
                None,
            )
            if not proxy or not config.get("remarks") and not proxy.get("tag"):
                continue
            configs.append(
                {
                    "uri": None,
                    "config": config,
                    "protocol": proxy.get("protocol") or "json",
                    "target": outbound_target(proxy),
                    "remark": str(config.get("remarks") or proxy.get("tag") or ""),
                }
            )
            continue
        if line in seen:
            continue
        scheme = line.split("://", 1)[0].lower()
        if scheme not in SUPPORTED_SCHEMES:
            continue
        try:
            outbound = outbound_from_uri(line)
        except RatioError:
            continue
        seen.add(line)
        configs.append(
            {
                "uri": line,
                "config": None,
                "protocol": scheme,
                "target": outbound_target(outbound),
                "remark": config_remark(line, outbound),
            }
        )
    return configs


def choose_config(configs, pick):
    if pick is not None:
        if 1 <= pick <= len(configs):
            return configs[pick - 1]
        raise RatioError(f"Config number must be between 1 and {len(configs)}.")
    if len(configs) == 1:
        print(f"Found one config in the subscription: {configs[0]['remark']}")
        return configs[0]
    print()
    print("Available configs in the subscription:")
    for index, item in enumerate(configs, 1):
        remark = item["remark"] or "(unnamed)"
        print(f"  {index}) {item['protocol']} | {item['target']} | {remark}")
    if not sys.stdin.isatty():
        raise RatioError("Multiple configs found; use --pick N to choose one.")
    while True:
        raw = ask("Config number", 1)
        try:
            index = int(normalize_digits(raw))
        except ValueError:
            index = 0
        if 1 <= index <= len(configs):
            return configs[index - 1]
        print("Invalid number.")


def free_port() -> int:
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]
    finally:
        sock.close()


def xray_asset_name() -> str:
    machine = platform.machine().lower()
    if machine in ("x86_64", "amd64"):
        return "Xray-linux-64.zip"
    if machine in ("aarch64", "arm64"):
        return "Xray-linux-arm64-v8a.zip"
    if machine.startswith("armv7"):
        return "Xray-linux-arm32-v7a.zip"
    raise RatioError(f"Automatic Xray download is not supported on {machine}.")


def xray_runs(path: Path) -> bool:
    try:
        result = subprocess.run([str(path), "-version"], capture_output=True, timeout=15)
        return result.returncode == 0
    except (OSError, subprocess.SubprocessError):
        return False


def ensure_xray() -> Path:
    found = shutil.which("xray")
    if found:
        return Path(found)
    target = CACHE_DIR / "xray"
    if target.exists() and xray_runs(target):
        return target
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    asset = xray_asset_name()
    url = f"https://github.com/XTLS/Xray-core/releases/latest/download/{asset}"
    print(f"Xray not found; downloading from GitHub ({asset}) ...")
    request = Request(url, headers={"User-Agent": "ratio-test"})
    try:
        with urlopen(request, timeout=120) as response:
            payload = response.read()
    except (URLError, TimeoutError) as exc:
        raise RatioError(f"Could not download Xray: {exc}")
    archive = CACHE_DIR / asset
    archive.write_bytes(payload)
    try:
        with zipfile.ZipFile(archive) as zf:
            member = next((name for name in zf.namelist() if name == "xray" or name.endswith("/xray")), None)
            if not member:
                raise RatioError("The xray binary was not found inside the archive.")
            with zf.open(member) as src, open(target, "wb") as dst:
                shutil.copyfileobj(src, dst)
        target.chmod(0o755)
    finally:
        archive.unlink(missing_ok=True)
    if not xray_runs(target):
        raise RatioError("The downloaded Xray binary is not executable.")
    return target


def build_client_config(config_input: str):
    text = config_input.strip()
    remark = ""
    if text.startswith("{"):
        try:
            config = json.loads(text)
        except ValueError as exc:
            raise RatioError(f"Invalid config JSON: {exc}")
        if not isinstance(config, dict):
            raise RatioError("Config JSON must be an object.")
        outbounds = config.get("outbounds") or []
        if not isinstance(outbounds, list) or not outbounds:
            raise RatioError("The JSON config has no outbound; provide a client config, not a server config.")
        proxy_tags = [
            item.get("tag")
            for item in outbounds
            if isinstance(item, dict)
            and item.get("protocol") not in DIRECT_PROTOCOLS
            and item.get("tag")
        ]
        target = proxy_tags[0] if proxy_tags else None
        for item in outbounds:
            if isinstance(item, dict) and item.get("remarks"):
                remark = str(item["remarks"])
                break
        return config, None, target, remark, text
    if "://" not in text:
        raise RatioError("Config must be a link (vless/vmess/trojan/ss) or JSON.")
    outbound = outbound_from_uri(text)
    remark = config_remark(text, outbound)
    outbound.pop("_remark", None)
    return None, outbound, None, remark, text


def prepare_config(base_config, outbound, target_tag, port: int):
    if base_config is None:
        config = {
            "log": {"loglevel": "warning"},
            "inbounds": [],
            "outbounds": [outbound, {"protocol": "freedom", "tag": "direct"}],
        }
        target_tag = outbound.get("tag")
    else:
        config = json.loads(json.dumps(base_config))
        config["log"] = {"loglevel": "warning"}
        config.setdefault("inbounds", [])
    tag = "ratio-test-socks"
    config["inbounds"].append(
        {
            "tag": tag,
            "listen": "127.0.0.1",
            "port": port,
            "protocol": "socks",
            "settings": {"auth": "noauth", "udp": False},
        }
    )
    if target_tag:
        routing = config.setdefault("routing", {})
        rules = routing.setdefault("rules", [])
        rules.insert(
            0,
            {"type": "field", "inboundTag": [tag], "outboundTag": target_tag},
        )
    return config


def wait_for_port(port: int, process: subprocess.Popen, log_path: Path, timeout: float = 15.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        if process.poll() is not None:
            log = log_path.read_text(errors="replace") if log_path.exists() else ""
            raise RatioError(f"Xray exited before startup:\n{log[-2000:]}")
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.3):
                return
        except OSError:
            time.sleep(0.1)
    raise RatioError("Xray did not open the SOCKS port in time.")


def curl_size(port: int, url: str, max_time: int):
    command = [
        "curl",
        "-sS",
        "--socks5-hostname",
        f"127.0.0.1:{port}",
        "--connect-timeout",
        "10",
        "--max-time",
        str(max_time),
        "-o",
        os.devnull,
        "-w",
        "%{size_download}",
        url,
    ]
    try:
        result = subprocess.run(command, capture_output=True, text=True, timeout=max_time + 15)
    except subprocess.TimeoutExpired:
        return 0, "timeout"
    except OSError as exc:
        raise RatioError(f"Could not run curl: {exc}")
    try:
        size = int(float((result.stdout or "0").strip() or 0))
    except ValueError:
        size = 0
    error = (result.stderr or "").strip()
    return size, error


def curl_probe(port: int, url: str, max_time: int):
    command = [
        "curl",
        "-sS",
        "--socks5-hostname",
        f"127.0.0.1:{port}",
        "--connect-timeout",
        "10",
        "--max-time",
        str(max_time),
        "-o",
        os.devnull,
        "-w",
        "%{http_code} %{size_download}",
        url,
    ]
    try:
        result = subprocess.run(command, capture_output=True, text=True, timeout=max_time + 15)
    except subprocess.TimeoutExpired:
        return 0, 0, "timeout"
    except OSError as exc:
        raise RatioError(f"Could not run curl: {exc}")
    parts = (result.stdout or "").split()
    try:
        code = int(parts[0]) if parts else 0
    except ValueError:
        code = 0
    try:
        size = int(float(parts[1])) if len(parts) > 1 else 0
    except ValueError:
        size = 0
    return code, size, (result.stderr or "").strip()


def probe(port: int, urls, timeout: int):
    for url in urls:
        code, size, error = curl_probe(port, url, timeout)
        if 200 <= code < 400:
            return size
        detail = error or (f"HTTP {code}" if code else "no response")
        print(f"  probe failed: {detail} ({url})", file=sys.stderr)
    raise RatioError(
        "The config did not respond through the proxy. "
        "The node may be down, blocked, or too slow. "
        "Try another config with --pick N, or increase --probe-timeout."
    )


def download(port: int, template: str, total: int, chunk: int, chunk_timeout: int):
    done = 0
    attempts = 0
    while done < total:
        size = min(chunk, total - done)
        url = template.replace("{bytes}", str(size)).replace("{remaining}", str(total - done))
        received, error = curl_size(port, url, chunk_timeout)
        done += received
        attempts += 1
        print(f"  received {human(received)} | total {human(done)} of {human(total)}", file=sys.stderr)
        if received <= 0:
            if attempts >= 6:
                break
            time.sleep(1 + attempts)
        elif done < total and received < size * 0.5 and attempts >= 6:
            break
    return done


def read_usage_after(sub_url: str, baseline, wait: int, poll_timeout: int):
    time.sleep(wait)
    base_total = baseline["upload"] + baseline["download"]
    deadline = time.time() + poll_timeout
    readings = []
    while time.time() <= deadline:
        try:
            reading = read_usage(sub_url, retries=1)
        except RatioError:
            reading = None
        if reading is not None:
            readings.append((time.time(), reading))
            if reading["upload"] + reading["download"] > base_total:
                break
        time.sleep(5)
    for _ in range(2):
        time.sleep(4)
        try:
            readings.append((time.time(), read_usage(sub_url, retries=1)))
        except RatioError:
            pass
    if not readings:
        raise RatioError("Could not read usage after the test.")
    best_time, best = max(readings, key=lambda item: item[1]["upload"] + item[1]["download"])
    return best, readings, best_time


def run(args) -> int:
    sub_url = args.sub or ask("Subscription URL")
    if not sub_url:
        raise RatioError("Subscription URL is required.")

    if args.list:
        body, _ = fetch(sub_url)
        configs = extract_configs(body)
        if not configs:
            print("No supported configs found in the subscription.")
            return 2
        for index, item in enumerate(configs, 1):
            remark = item["remark"] or "(unnamed)"
            print(f"{index}) {item['protocol']} | {item['target']} | {remark}")
        return 0

    config_arg = args.config
    if not config_arg:
        print("Reading configs from the subscription ...")
        configs = []
        try:
            body, _ = fetch(sub_url)
            configs = extract_configs(body)
        except RatioError as exc:
            print(f"Could not read configs from the subscription: {exc}")
        if configs:
            selected = choose_config(configs, args.pick)
            if selected.get("config") is not None:
                config_arg = json.dumps(selected["config"], ensure_ascii=False)
            else:
                config_arg = selected["uri"]
        else:
            print("No configs found in the subscription; enter one manually.")
            config_arg = ask("Config (JSON file path or vless:// / vmess:// / trojan:// / ss:// link)")
    if not config_arg:
        raise RatioError("A config is required.")
    config_text = config_arg
    if not config_arg.lstrip().startswith("{"):
        try:
            config_path = Path(os.path.expanduser(config_arg))
            if config_path.is_file():
                config_text = config_path.read_text(encoding="utf-8", errors="replace")
        except OSError:
            pass

    if args.size is None:
        if sys.stdin.isatty():
            raw_size = ask("Test size in MB", f"{DEFAULT_SIZE_MB:g}")
            try:
                size_mb = float(normalize_digits(raw_size))
            except ValueError:
                raise RatioError("Invalid test size.")
        else:
            size_mb = DEFAULT_SIZE_MB
    else:
        size_mb = float(args.size)
    if size_mb < MIN_SIZE_MB:
        raise RatioError(f"Minimum test size is {MIN_SIZE_MB:g} MB.")
    wait = max(0, int(args.wait))
    poll_timeout = max(wait, int(args.poll_timeout))
    chunk_bytes = max(256 * 1024, int(float(args.chunk) * 1024 * 1024))
    total_bytes = int(size_mb * 1024 * 1024)

    print("Reading subscription info ...")
    before = read_usage(sub_url)
    remaining = None
    if before.get("total"):
        remaining = before["total"] - before["upload"] - before["download"]
        if remaining > 0 and remaining < total_bytes:
            print(f"Warning: remaining subscription volume ({human(remaining)}) is less than the test size.")
    print(
        f"Current subscription usage: upload {human(before['upload'])} | download {human(before['download'])}"
        + (f" | remaining {human(remaining)}" if remaining is not None else "")
    )

    base_config, outbound, target_tag, remark, scan_text = build_client_config(config_text)
    claimed, claimed_source = detect_claimed(scan_text + " " + remark, args.claimed)

    xray = ensure_xray()
    port = free_port()
    config = prepare_config(base_config, outbound, target_tag, port)

    tmpdir = Path(tempfile.mkdtemp(prefix="ratio-test-"))
    cfg_path = tmpdir / "config.json"
    log_path = tmpdir / "xray.log"
    cfg_path.write_text(json.dumps(config, ensure_ascii=False, indent=2), encoding="utf-8")
    cfg_path.chmod(0o600)
    process = None
    try:
        check = subprocess.run(
            [str(xray), "run", "-test", "-c", str(cfg_path)],
            capture_output=True,
            text=True,
            timeout=40,
        )
        if check.returncode != 0:
            raise RatioError("The config is not valid for Xray:\n" + (check.stderr or check.stdout)[-2000:])

        log_file = open(log_path, "wb")
        process = subprocess.Popen(
            [str(xray), "run", "-c", str(cfg_path)],
            stdout=log_file,
            stderr=log_file,
        )
        wait_for_port(port, process, log_path)
        print("Xray started; probing connectivity through the config ...")
        probe_urls = PROBE_URLS if args.probe_url == DEFAULT_PROBE_URL else [args.probe_url]
        probe(port, probe_urls, args.probe_timeout)
        print("Connection OK. Reading baseline usage ...")
        time.sleep(2)
        before = read_usage(sub_url)
        before_time = time.time()
        drift_rate = 0.0
        drift_used = 0.0
        if args.drift_window > 0:
            print(f"Measuring background subscription usage ({args.drift_window}s) ...")
            drift_start = before
            drift_start_time = before_time
            time.sleep(args.drift_window)
            before = read_usage(sub_url)
            before_time = time.time()
            drift_bytes = (before["upload"] + before["download"]) - (
                drift_start["upload"] + drift_start["download"]
            )
            drift_window = max(0.001, before_time - drift_start_time)
            if drift_bytes > 0:
                drift_rate = drift_bytes / drift_window
                if drift_bytes >= 64 * 1024:
                    print(
                        f"Warning: the subscription is consuming data while idle "
                        f"({human(drift_bytes)} in {drift_window:.0f}s, about {human(drift_rate)}/s)."
                    )
                    print("Results will be corrected using this background rate.")
                else:
                    print(f"Background usage is small ({human(drift_bytes)}) and will be accounted for.")
            else:
                print("No background usage detected.")

        print(f"Starting test download ({human(total_bytes)}) ...")
        started = time.time()
        downloaded = download(port, args.test_url, total_bytes, chunk_bytes, int(args.timeout))
        elapsed = time.time() - started
        if downloaded <= 0:
            raise RatioError("No data was downloaded through the config.")
        if downloaded < total_bytes:
            print(f"Warning: only {human(downloaded)} of {human(total_bytes)} was downloaded.")
        print(f"Download finished: {human(downloaded)} in {elapsed:.1f}s. Waiting for usage to be recorded ...")

        after, readings, after_time = read_usage_after(sub_url, before, wait, poll_timeout)
        drift_used = drift_rate * max(0.0, after_time - before_time)
    finally:
        if process is not None and process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
        if args.keep_temp:
            print(f"Temporary files: {tmpdir}")
        else:
            shutil.rmtree(tmpdir, ignore_errors=True)

    delta_up = after["upload"] - before["upload"]
    delta_down = after["download"] - before["download"]
    used = delta_up + delta_down
    if used <= 0:
        raise RatioError(
            "Usage reported by the subscription did not change. The panel may be delayed or may only show usage in its dashboard."
        )
    if used < 0:
        raise RatioError("Subscription usage decreased; the counter may have been reset.")
    raw_ratio = used / downloaded
    ratio = raw_ratio
    if drift_used > 0 and used - drift_used > 0:
        ratio = (used - drift_used) / downloaded
    percent = (ratio - 1.0) * 100.0

    print()
    print("=" * 48)
    print("Coefficient Test Result")
    print("=" * 48)
    if remark:
        print(f"Config: {remark}")
    print(f"Actual download: {human(downloaded)}")
    print(
        f"Recorded usage: {human(used)} "
        f"(download {human(delta_down)} + upload {human(delta_up)})"
    )
    print(f"Measured coefficient: {ratio:.3f}x  ({percent:+.1f}% vs. actual traffic)")
    if drift_used > 0:
        print(
            f"Raw coefficient: {raw_ratio:.3f}x | "
            f"background usage during test: ~{human(drift_used)}"
        )
    if claimed:
        if abs(ratio - claimed) <= max(0.05, claimed * 0.03):
            verdict = "matches the claim"
        elif ratio > claimed:
            verdict = "higher than the claim (worse for the user)"
        else:
            verdict = "lower than the claim (better for the user)"
        print(f"Claimed coefficient: {claimed:.2f}x  ->  {verdict}")
    if ratio < 0.95:
        print("Warning: coefficient is below 1; the panel may be recording usage late.")
    if after.get("total"):
        remaining_after = after["total"] - after["upload"] - after["download"]
        if remaining_after > 0:
            print(
                f"Remaining subscription volume: {human(remaining_after)} "
                f"(about {human(remaining_after / ratio)} of real traffic at this coefficient)"
            )
    if len(readings) > 1:
        vals = {item["upload"] + item["download"] for _, item in readings}
        if len(vals) > 1:
            print("Note: usage changed between reads; the highest reading is used.")

    if args.json_out:
        payload = {
            "version": VERSION,
            "tested_at": datetime.now(timezone.utc).isoformat(),
            "config": remark or None,
            "downloaded_bytes": downloaded,
            "usage_before": before,
            "usage_after": after,
            "delta_upload": delta_up,
            "delta_download": delta_down,
            "delta_total": used,
            "ratio": round(ratio, 4),
            "ratio_raw": round(raw_ratio, 4),
            "background_bytes_per_second": round(drift_rate, 1),
            "background_bytes_in_test": int(drift_used),
            "claimed_ratio": claimed,
            "duration_seconds": round(elapsed, 2),
        }
        Path(args.json_out).write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"JSON result saved to {args.json_out}.")
    return 0


def build_parser():
    parser = argparse.ArgumentParser(
        description="Measure the real traffic coefficient of a V2Ray config by comparing actual downloads with subscription usage",
    )
    parser.add_argument("--sub", help="Subscription URL")
    parser.add_argument("--config", help="Config link or JSON file path")
    parser.add_argument("--size", type=float, help=f"Test size in MB; minimum {MIN_SIZE_MB:g} (default {DEFAULT_SIZE_MB:g})")
    parser.add_argument("--wait", type=int, default=DEFAULT_WAIT, help=f"Seconds to wait before reading usage (default {DEFAULT_WAIT})")
    parser.add_argument("--poll-timeout", type=int, default=DEFAULT_POLL_TIMEOUT, help="Maximum seconds to wait for usage to be recorded")
    parser.add_argument("--drift-window", type=int, default=10, help="Seconds to sample background usage before the test; 0 disables it")
    parser.add_argument("--pick", type=int, help="Config number inside the subscription for automatic selection")
    parser.add_argument("--list", action="store_true", help="Only list configs from the subscription and exit")
    parser.add_argument("--chunk", type=float, default=DEFAULT_CHUNK_MB, help="Download request size in MB")
    parser.add_argument("--timeout", type=float, default=180, help="Maximum seconds per curl request")
    parser.add_argument("--test-url", default=DEFAULT_TEST_URL, help="Test download URL; {bytes} is replaced with the size")
    parser.add_argument("--probe-url", default=DEFAULT_PROBE_URL, help="Probe URL used to check connectivity")
    parser.add_argument("--probe-timeout", type=int, default=DEFAULT_PROBE_TIMEOUT, help=f"Seconds per probe URL (default {DEFAULT_PROBE_TIMEOUT})")
    parser.add_argument("--claimed", type=float, help="Claimed coefficient, if not present in the config name")
    parser.add_argument("--json-out", help="Save the result to a JSON file")
    parser.add_argument("--keep-temp", action="store_true", help="Keep temporary files and the Xray log")
    parser.add_argument("--version", action="version", version=f"%(prog)s {VERSION}")
    return parser


def main(argv=None) -> int:
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(line_buffering=True)
    if not shutil.which("curl"):
        print("curl is not installed.", file=sys.stderr)
        return 2
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return run(args)
    except RatioError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 2
    except KeyboardInterrupt:
        print("\nCancelled.", file=sys.stderr)
        return 130


if __name__ == "__main__":
    sys.exit(main())
