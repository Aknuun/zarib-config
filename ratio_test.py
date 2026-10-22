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
        raise RatioError("ورودی تمام شد.")
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
        raise RatioError(f"دریافت لینک ساب با خطای HTTP {exc.code} مواجه شد.")
    except URLError as exc:
        raise RatioError(f"اتصال به لینک ساب برقرار نشد: {exc.reason}")
    except TimeoutError:
        raise RatioError("دریافت لینک ساب زمان‌بر شد (timeout).")


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
                "لینک ساب هدر subscription-userinfo ندارد و مصرف هم در بدنه پیدا نشد."
            )
        except RatioError as exc:
            last_error = exc
        if attempt < retries - 1:
            time.sleep(delay)
    raise last_error or RatioError("خواندن مصرف ساب ناموفق بود.")


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
        raise RatioError("لینک vless نامعتبر است.")
    query = {key: (values[0] if values else "") for key, values in parse_qs(parsed.query, keep_blank_values=True).items()}
    user_id = unquote(parsed.username or "")
    if not user_id:
        raise RatioError("UUID در لینک vless پیدا نشد.")
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
        raise RatioError("لینک vmess قابل رمزگشایی نبود.")
    host = str(data.get("add") or "").strip()
    if not host:
        raise RatioError("آدرس سرور در لینک vmess پیدا نشد.")
    port = int(data.get("port") or 443)
    user_id = str(data.get("id") or "")
    if not user_id:
        raise RatioError("UUID در لینک vmess پیدا نشد.")
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
        raise RatioError("لینک trojan نامعتبر است.")
    password = unquote(parsed.username or "")
    if not password:
        raise RatioError("رمز در لینک trojan پیدا نشد.")
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
            raise RatioError("لینک ss نامعتبر است.")
        method, password = decoded.split(":", 1)
    else:
        try:
            decoded = base64.b64decode(body + "=" * (-len(body) % 4)).decode("utf-8")
        except (ValueError, UnicodeDecodeError):
            raise RatioError("لینک ss قابل رمزگشایی نبود.")
        if "@" not in decoded or ":" not in decoded:
            raise RatioError("لینک ss نامعتبر است.")
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
            f"پروتکل «{scheme}» پشتیبانی نمی‌شود. لینک vless/vmess/trojan/ss بده یا فایل JSON کامل کانفیگ را بده."
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
        r"ضریب\s*[:=]?\s*(\d+(?:[.,]\d+)?)",
        r"(?<![\w.])(\d+(?:[.,]\d+)?)\s*[xX×](?![\w])",
        r"[xX×]\s*(\d+(?:[.,]\d+)?)(?![\w.])",
    ]
    for pattern in patterns:
        for match in re.finditer(pattern, normalized):
            value = float(match.group(1).replace(",", "."))
            if 0.1 <= value <= 20:
                return value, "remark"
    return None, None


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
    raise RatioError(f"معماری {machine} برای دانلود خودکار Xray پشتیبانی نمی‌شود.")


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
    print(f"Xray پیدا نشد؛ دانلود از گیت‌هاب ({asset}) ...")
    request = Request(url, headers={"User-Agent": "ratio-test"})
    try:
        with urlopen(request, timeout=120) as response:
            payload = response.read()
    except (URLError, TimeoutError) as exc:
        raise RatioError(f"دانلود Xray ناموفق بود: {exc}")
    archive = CACHE_DIR / asset
    archive.write_bytes(payload)
    try:
        with zipfile.ZipFile(archive) as zf:
            member = next((name for name in zf.namelist() if name == "xray" or name.endswith("/xray")), None)
            if not member:
                raise RatioError("فایل xray داخل آرشیو پیدا نشد.")
            with zf.open(member) as src, open(target, "wb") as dst:
                shutil.copyfileobj(src, dst)
        target.chmod(0o755)
    finally:
        archive.unlink(missing_ok=True)
    if not xray_runs(target):
        raise RatioError("اجرای Xray دانلودشده ممکن نشد.")
    return target


def build_client_config(config_input: str):
    text = config_input.strip()
    remark = ""
    if text.startswith("{"):
        try:
            config = json.loads(text)
        except ValueError as exc:
            raise RatioError(f"JSON کانفیگ نامعتبر است: {exc}")
        if not isinstance(config, dict):
            raise RatioError("JSON کانفیگ باید یک آبجکت باشد.")
        outbounds = config.get("outbounds") or []
        if not isinstance(outbounds, list) or not outbounds:
            raise RatioError("کانفیگ JSON هیچ outbound ندارد؛ کانفیگ کلاینت بده، نه سرور.")
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
        raise RatioError("کانفیگ باید لینک (vless/vmess/trojan/ss) یا JSON باشد.")
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
            raise RatioError(f"Xray قبل از شروع بسته شد:\n{log[-2000:]}")
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.3):
                return
        except OSError:
            time.sleep(0.1)
    raise RatioError("Xray در زمان مشخص‌شده پورت SOCKS را باز نکرد.")


def curl_size(port: int, url: str, max_time: int):
    command = [
        "curl",
        "-sS",
        "--socks5-hostname",
        f"127.0.0.1:{port}",
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
        raise RatioError(f"اجرای curl ممکن نشد: {exc}")
    try:
        size = int(float((result.stdout or "0").strip() or 0))
    except ValueError:
        size = 0
    error = (result.stderr or "").strip()
    return size, error


def probe(port: int, url: str):
    size, error = curl_size(port, url, 30)
    if size <= 0:
        raise RatioError(f"کانفیگ از طریق پروکسی جواب نداد. خطای curl: {error or 'بدون خروجی'}")
    return size


def download(port: int, template: str, total: int, chunk: int, chunk_timeout: int):
    done = 0
    attempts = 0
    while done < total:
        size = min(chunk, total - done)
        url = template.replace("{bytes}", str(size)).replace("{remaining}", str(total - done))
        received, error = curl_size(port, url, chunk_timeout)
        done += received
        attempts += 1
        print(f"  دریافت {human(received)} | مجموع {human(done)} از {human(total)}", file=sys.stderr)
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
            readings.append(reading)
            if reading["upload"] + reading["download"] > base_total:
                break
        time.sleep(5)
    for _ in range(2):
        time.sleep(4)
        try:
            readings.append(read_usage(sub_url, retries=1))
        except RatioError:
            pass
    if not readings:
        raise RatioError("خواندن مصرف بعد از تست ممکن نشد.")
    best = max(readings, key=lambda item: item["upload"] + item["download"])
    return best, readings


def run(args) -> int:
    sub_url = args.sub or ask("لینک سابسکریپشن")
    if not sub_url:
        raise RatioError("لینک سابسکریپشن لازم است.")

    config_arg = args.config or ask("کانفیگ (مسیر فایل JSON یا لینک vless:// / vmess:// / trojan:// / ss://)")
    if not config_arg:
        raise RatioError("کانفیگ لازم است.")
    config_path = Path(os.path.expanduser(config_arg))
    if config_path.is_file():
        config_text = config_path.read_text(encoding="utf-8", errors="replace")
    else:
        config_text = config_arg

    size_mb = float(args.size)
    if size_mb < MIN_SIZE_MB:
        raise RatioError(f"حداقل حجم تست {MIN_SIZE_MB:g} مگابایت است.")
    wait = max(0, int(args.wait))
    poll_timeout = max(wait, int(args.poll_timeout))
    chunk_bytes = max(256 * 1024, int(float(args.chunk) * 1024 * 1024))
    total_bytes = int(size_mb * 1024 * 1024)

    print("در حال خواندن اطلاعات ساب ...")
    before = read_usage(sub_url)
    remaining = None
    if before.get("total"):
        remaining = before["total"] - before["upload"] - before["download"]
        if remaining > 0 and remaining < total_bytes:
            print(f"هشدار: حجم باقی‌مانده ساب ({human(remaining)}) کمتر از حجم تست است.")
    print(
        f"مصرف فعلی ساب: آپلود {human(before['upload'])} | دانلود {human(before['download'])}"
        + (f" | باقی‌مانده {human(remaining)}" if remaining is not None else "")
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
            raise RatioError("کانفیگ برای Xray معتبر نیست:\n" + (check.stderr or check.stdout)[-2000:])

        log_file = open(log_path, "wb")
        process = subprocess.Popen(
            [str(xray), "run", "-c", str(cfg_path)],
            stdout=log_file,
            stderr=log_file,
        )
        wait_for_port(port, process, log_path)
        print("Xray اجرا شد؛ بررسی اتصال از طریق کانفیگ ...")
        probe(port, args.probe_url)
        print("اتصال برقرار است. خواندن مبنای مصرف ...")
        time.sleep(2)
        before = read_usage(sub_url)

        print(f"شروع دانلود تست ({human(total_bytes)}) ...")
        started = time.time()
        downloaded = download(port, args.test_url, total_bytes, chunk_bytes, int(args.timeout))
        elapsed = time.time() - started
        if downloaded <= 0:
            raise RatioError("هیچ داده‌ای از طریق کانفیگ دانلود نشد.")
        if downloaded < total_bytes:
            print(f"هشدار: فقط {human(downloaded)} از {human(total_bytes)} دانلود شد.")
        print(f"دانلود تمام شد: {human(downloaded)} در {elapsed:.1f} ثانیه. صبر برای ثبت مصرف ...")

        after, readings = read_usage_after(sub_url, before, wait, poll_timeout)
    finally:
        if process is not None and process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
        if args.keep_temp:
            print(f"فایل‌های موقت: {tmpdir}")
        else:
            shutil.rmtree(tmpdir, ignore_errors=True)

    delta_up = after["upload"] - before["upload"]
    delta_down = after["download"] - before["download"]
    used = delta_up + delta_down
    if used <= 0:
        raise RatioError(
            "مصرف ثبت‌شده در ساب تغییر نکرد. احتمالاً پنل مصرف را با تأخیر یا فقط در پنل نشان می‌دهد."
        )
    if used < 0:
        raise RatioError("مصرف ساب کاهش پیدا کرد؛ احتمالاً شمارنده ساب ریست شده است.")
    ratio = used / downloaded
    percent = (ratio - 1.0) * 100.0

    print()
    print("=" * 48)
    print("نتیجه تست ضریب")
    print("=" * 48)
    if remark:
        print(f"کانفیگ: {remark}")
    print(f"حجم دانلود واقعی: {human(downloaded)}")
    print(
        f"مصرف ثبت‌شده: {human(used)} "
        f"(دانلود {human(delta_down)} + آپلود {human(delta_up)})"
    )
    print(f"ضریب واقعی: {ratio:.3f}x  ({percent:+.1f}% نسبت به حجم واقعی)")
    if claimed:
        if abs(ratio - claimed) <= max(0.05, claimed * 0.03):
            verdict = "مطابق ادعا"
        elif ratio > claimed:
            verdict = "بیشتر از ادعا (به ضرر کاربر)"
        else:
            verdict = "کمتر از ادعا (به نفع کاربر)"
        print(f"ضریب ادعاشده: {claimed:.2f}x  →  {verdict}")
    if ratio < 0.95:
        print("هشدار: ضریب کمتر از ۱ است؛ ممکن است پنل مصرف را با تأخیر ثبت کند.")
    if after.get("total"):
        remaining_after = after["total"] - after["upload"] - after["download"]
        if remaining_after > 0:
            print(
                f"باقی‌مانده ساب: {human(remaining_after)} "
                f"(معادل {human(remaining_after / ratio)} حجم واقعی با این ضریب)"
            )
    if len(readings) > 1:
        vals = {item["upload"] + item["download"] for item in readings}
        if len(vals) > 1:
            print("توجه: مصرف در چند قرائت تغییر کرد؛ عدد آخر ملاک است.")

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
            "claimed_ratio": claimed,
            "duration_seconds": round(elapsed, 2),
        }
        Path(args.json_out).write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"خروجی JSON در {args.json_out} ذخیره شد.")
    return 0


def build_parser():
    parser = argparse.ArgumentParser(
        description="سنجش ضریب واقعی مصرف حجم کانفیگ V2Ray با مقایسه دانلود واقعی و مصرف سابسکریپشن",
    )
    parser.add_argument("--sub", help="لینک سابسکریپشن")
    parser.add_argument("--config", help="لینک کانفیگ یا مسیر فایل JSON")
    parser.add_argument("--size", type=float, default=DEFAULT_SIZE_MB, help=f"حجم تست به مگابایت؛ حداقل {MIN_SIZE_MB:g} (پیش‌فرض {DEFAULT_SIZE_MB:g})")
    parser.add_argument("--wait", type=int, default=DEFAULT_WAIT, help=f"ثانیه صبر قبل از خواندن مصرف (پیش‌فرض {DEFAULT_WAIT})")
    parser.add_argument("--poll-timeout", type=int, default=DEFAULT_POLL_TIMEOUT, help="حداکثر ثانیه انتظار برای ثبت مصرف")
    parser.add_argument("--chunk", type=float, default=DEFAULT_CHUNK_MB, help="اندازه هر درخواست دانلود به مگابایت")
    parser.add_argument("--timeout", type=float, default=180, help="حداکثر زمان هر درخواست curl به ثانیه")
    parser.add_argument("--test-url", default=DEFAULT_TEST_URL, help="آدرس دانلود تست؛ {bytes} با حجم جایگزین می‌شود")
    parser.add_argument("--probe-url", default=DEFAULT_PROBE_URL, help="آدرس بررسی اتصال")
    parser.add_argument("--claimed", type=float, help="ضریب ادعاشده (در صورت نبود در نام کانفیگ)")
    parser.add_argument("--json-out", help="ذخیره نتیجه در فایل JSON")
    parser.add_argument("--keep-temp", action="store_true", help="نگه‌داشتن فایل‌های موقت و لاگ Xray")
    parser.add_argument("--version", action="version", version=f"%(prog)s {VERSION}")
    return parser


def main(argv=None) -> int:
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(line_buffering=True)
    if not shutil.which("curl"):
        print("curl نصب نیست.", file=sys.stderr)
        return 2
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return run(args)
    except RatioError as exc:
        print(f"خطا: {exc}", file=sys.stderr)
        return 2
    except KeyboardInterrupt:
        print("\nلغو شد.", file=sys.stderr)
        return 130


if __name__ == "__main__":
    sys.exit(main())
