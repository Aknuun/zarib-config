from __future__ import annotations

import json
import os
import re
import shutil
import signal
import subprocess
import sys
import threading
import time
import uuid
from collections import defaultdict, deque
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional
from urllib.parse import urlparse

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import FileResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel

BASE_DIR = Path(__file__).resolve().parent
STATIC_DIR = BASE_DIR / "static"
sys.path.insert(0, str(BASE_DIR))
sys.path.insert(0, str(BASE_DIR.parent))

os.environ.setdefault("RATIO_SAFE_FETCH", "1")

import ratio_test as rt  # noqa: E402

SCRIPT = Path(rt.__file__).resolve()
CACHE_ROOT = Path(
    os.environ.get("RATIO_WEB_CACHE", str(Path.home() / ".cache" / "v2ray-ratio-web"))
)
JOBS_DIR = CACHE_ROOT / "jobs"
CACHE_ROOT.mkdir(parents=True, exist_ok=True)
JOBS_DIR.mkdir(parents=True, exist_ok=True)

MAX_ACTIVE = 2
MAX_QUEUED = 8
MAX_LOG_LINES = 120
JOB_TIMEOUT = 30 * 60
MIN_SIZE_MB = rt.MIN_SIZE_MB
MAX_SIZE_MB = 50.0

app = FastAPI(
    title="V2Ray Ratio Test",
    docs_url=None,
    redoc_url=None,
    openapi_url=None,
)

JOBS: dict[str, dict] = {}
JOBS_LOCK = threading.Lock()
EXECUTOR = None
RATE_LOCK = threading.Lock()
RATES: dict[str, dict[str, deque]] = defaultdict(lambda: defaultdict(deque))


def _executor():
    global EXECUTOR
    if EXECUTOR is None:
        from concurrent.futures import ThreadPoolExecutor

        EXECUTOR = ThreadPoolExecutor(max_workers=MAX_ACTIVE, thread_name_prefix="ratio-job")
    return EXECUTOR


def api_error(code: str, message: str = "", status: int = 400):
    return JSONResponse({"code": code, "error": message}, status_code=status)


@app.exception_handler(RequestValidationError)
async def validation_error(request: Request, exc: RequestValidationError):
    return api_error("invalid_input", "Invalid request.", 422)


@app.middleware("http")
async def security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["Referrer-Policy"] = "no-referrer"
    response.headers["Permissions-Policy"] = "geolocation=(), microphone=(), camera=()"
    response.headers["Content-Security-Policy"] = (
        "default-src 'self'; script-src 'self'; "
        "style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "
        "font-src 'self' https://cdn.jsdelivr.net; img-src 'self' data:; "
        "connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"
    )
    if request.headers.get("x-forwarded-proto") == "https":
        response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
    if request.url.path.startswith("/api/"):
        response.headers["Cache-Control"] = "no-store"
    return response


def client_ip(request: Request) -> str:
    real = request.headers.get("x-real-ip")
    if real:
        return real.strip()
    forwarded = request.headers.get("x-forwarded-for")
    if forwarded:
        return forwarded.split(",")[0].strip()
    return request.client.host if request.client else "unknown"


def rate_limit(bucket: str, ip: str, limit: int, window: int = 3600) -> bool:
    now = time.time()
    with RATE_LOCK:
        queue = RATES[bucket][ip]
        while queue and now - queue[0] > window:
            queue.popleft()
        if len(queue) >= limit:
            return False
        queue.append(now)
        return True


def active_counts() -> tuple[int, int]:
    with JOBS_LOCK:
        running = sum(1 for job in JOBS.values() if job["status"] == "running")
        queued = sum(1 for job in JOBS.values() if job["status"] == "queued")
    return running, queued


def usage_view(usage: Optional[dict]):
    if not usage:
        return None
    upload = int(usage.get("upload") or 0)
    download = int(usage.get("download") or 0)
    total = int(usage.get("total") or 0)
    remaining = max(total - upload - download, 0) if total else None
    expire = usage.get("expire")
    days_left = None
    if expire:
        days_left = max(0, int((int(expire) - time.time()) // 86400))
    return {
        "upload": upload,
        "download": download,
        "used": upload + download,
        "total": total or None,
        "remaining": remaining,
        "expire": expire,
        "days_left": days_left,
    }


def _is_valid_url(url: str) -> bool:
    try:
        parsed = urlparse(url)
    except ValueError:
        return False
    return parsed.scheme in ("http", "https") and bool(parsed.hostname)


class InspectIn(BaseModel):
    sub_url: str


class JobIn(BaseModel):
    sub_url: str
    pick: int = 1
    size_mb: float = 10.0
    claimed: Optional[float] = None
    wait: int = 20
    drift: int = 10
    core: str = "auto"


@app.get("/healthz")
def healthz():
    return {"ok": True, "version": rt.VERSION}


@app.post("/api/inspect")
def inspect(payload: InspectIn, request: Request):
    ip = client_ip(request)
    if not rate_limit("inspect", ip, 30):
        return api_error("rate_limited", "Too many requests.", 429)
    url = payload.sub_url.strip()
    if not _is_valid_url(url):
        return api_error("invalid_url", "The subscription URL is not valid.")
    if not rt.is_public_url(url):
        return api_error("blocked_url", "The subscription URL must be publicly reachable.")
    try:
        body, headers = rt.fetch(url, timeout=20)
    except rt.RatioError as exc:
        return api_error("fetch_failed", str(exc), 400)
    except Exception as exc:  # noqa: BLE001
        return api_error("fetch_failed", str(exc), 400)
    usage = rt.parse_usage(headers, body)
    try:
        configs = rt.extract_configs(body)
    except Exception as exc:  # noqa: BLE001
        return api_error("no_configs", str(exc), 400)
    if not configs:
        return api_error("no_configs", "No supported configs found in the subscription.", 400)
    items = []
    for index, item in enumerate(configs[:100], 1):
        remark = item.get("remark") or ""
        claimed, _ = rt.detect_claimed(remark, None)
        items.append(
            {
                "index": index,
                "protocol": item.get("protocol") or "json",
                "target": item.get("target") or "",
                "remark": remark,
                "claimed": claimed,
            }
        )
    return {
        "usage": usage_view(usage),
        "configs": items,
        "count": len(configs),
        "insecure": url.lower().startswith("http://"),
    }


@app.post("/api/jobs")
def create_job(payload: JobIn, request: Request):
    ip = client_ip(request)
    if not rate_limit("jobs", ip, 5):
        return api_error("rate_limited", "Too many tests. Try again later.", 429)
    url = payload.sub_url.strip()
    if not _is_valid_url(url):
        return api_error("invalid_url", "The subscription URL is not valid.")
    if not rt.is_public_url(url):
        return api_error("blocked_url", "The subscription URL must be publicly reachable.")
    if not (MIN_SIZE_MB <= payload.size_mb <= MAX_SIZE_MB):
        return api_error("invalid_input", "Invalid test size.")
    if not (1 <= payload.pick <= 500):
        return api_error("invalid_input", "Invalid config number.")
    if payload.claimed is not None and not (0.1 <= payload.claimed <= 100):
        return api_error("invalid_input", "Invalid claimed ratio.")
    if not (0 <= payload.wait <= 60) or not (0 <= payload.drift <= 30):
        return api_error("invalid_input", "Invalid wait values.")
    if payload.core not in rt.CORE_CHOICES:
        return api_error("invalid_input", "Invalid Xray core.")
    running, queued = active_counts()
    if queued >= MAX_QUEUED:
        return api_error("busy", "The test queue is full.", 429)

    job_id = uuid.uuid4().hex
    job_dir = JOBS_DIR / job_id
    job_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
    sub_file = job_dir / "sub.txt"
    sub_file.write_text(url + "\n", encoding="utf-8")
    try:
        os.chmod(sub_file, 0o600)
    except OSError:
        pass
    result_file = job_dir / "result.json"
    now = time.time()
    job = {
        "id": job_id,
        "status": "queued",
        "stage": "queued",
        "progress": 0,
        "message": "",
        "created": now,
        "started": None,
        "finished": None,
        "result": None,
        "error": None,
        "error_code": None,
        "log": [],
        "cancel": False,
        "timed_out": False,
        "pick": payload.pick,
        "size_mb": payload.size_mb,
        "claimed": payload.claimed,
        "core": payload.core,
        "_dir": str(job_dir),
        "_sub_file": str(sub_file),
        "_result_file": str(result_file),
        "_proc": None,
    }
    with JOBS_LOCK:
        JOBS[job_id] = job
    _executor().submit(_run_job, job_id)
    return {"id": job_id, "status": "queued"}


def _set(job_id: str, **values):
    with JOBS_LOCK:
        job = JOBS.get(job_id)
        if job:
            job.update(values)


def _append_log(job_id: str, line: str):
    line = re.sub(r"\x1b\[[0-9;]*[A-Za-z]", "", line)
    line = line.split("\r")[-1].rstrip()
    if not line:
        return
    line = line[:500]
    with JOBS_LOCK:
        job = JOBS.get(job_id)
        if not job:
            return
        job["log"].append(line)
        if len(job["log"]) > MAX_LOG_LINES:
            del job["log"][: len(job["log"]) - MAX_LOG_LINES]


def _progress_from_line(job_id: str, line: str):
    stripped = line.strip()
    if not stripped:
        return
    updates = {}
    if "Reading subscription info" in stripped:
        updates.update(stage="reading", progress=5, message=stripped)
    elif "Current subscription usage" in stripped:
        updates.update(stage="reading", progress=10, message=stripped)
    elif "Xray started" in stripped or "Xray (" in stripped and "started" in stripped:
        updates.update(stage="connecting", progress=15, message=stripped)
    elif "Trying another Xray core" in stripped:
        updates.update(stage="connecting", progress=15, message=stripped)
    elif "Connection probe failed" in stripped:
        updates.update(stage="connecting", progress=18, message=stripped)
    elif "Retrying connection" in stripped:
        updates.update(stage="connecting", progress=18, message=stripped)
    elif "Connection OK" in stripped:
        updates.update(stage="baseline", progress=25, message=stripped)
    elif "Measuring background usage" in stripped:
        updates.update(stage="drift", progress=30, message=stripped)
    elif "Background usage" in stripped:
        updates.update(stage="drift", progress=35, message=stripped)
    elif "Starting test download" in stripped:
        updates.update(stage="downloading", progress=35, message=stripped)
    elif "Download finished" in stripped:
        updates.update(stage="waiting", progress=85, message=stripped)
    elif "Waiting for the panel" in stripped:
        updates.update(stage="waiting", progress=85, message=stripped)
    elif "Checking subscription usage" in stripped:
        updates.update(stage="waiting", progress=92, message=stripped)
    elif "did not change yet" in stripped:
        updates.update(stage="waiting", progress=90, message=stripped)
    elif "COEFFICIENT TEST RESULT" in stripped:
        updates.update(stage="done", progress=100, message=stripped)
    match = re.search(r"total\s+([\d.]+)\s*MB\s+of\s+([\d.]+)\s*MB", stripped)
    if match:
        total = float(match.group(1))
        target = float(match.group(2)) or 1.0
        progress = 35 + 50 * min(total / target, 1.0)
        updates.update(stage="downloading", progress=progress, message=stripped)
    if updates:
        with JOBS_LOCK:
            job = JOBS.get(job_id)
            if job and "progress" in updates:
                updates["progress"] = max(
                    float(job.get("progress") or 0), float(updates["progress"])
                )
        _set(job_id, **updates)


def _terminate(proc: subprocess.Popen):
    if proc.poll() is not None:
        return
    try:
        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
    except (ProcessLookupError, PermissionError):
        return
    try:
        proc.wait(timeout=8)
    except subprocess.TimeoutExpired:
        try:
            os.killpg(os.getpgid(proc.pid), signal.SIGKILL)
        except (ProcessLookupError, PermissionError):
            pass


def _classify_error(log: list[str], timed_out: bool, cancelled: bool) -> str:
    if cancelled:
        return "cancelled"
    if timed_out:
        return "timeout"
    text = "\n".join(log).lower()
    if "did not respond through the proxy" in text:
        return "probe_failed"
    if "cannot resolve host" in text:
        return "resolve_failed"
    if "not valid for xray" in text or "failed to start" in text or "no xray core" in text:
        return "xray_failed"
    if "could not read the subscription" in text or "subscription request failed" in text:
        return "fetch_failed"
    if "no supported configs" in text or "no configs found" in text:
        return "no_configs"
    if "config number" in text or "invalid config" in text or "not found" in text:
        return "pick_out_of_range"
    if "did not change" in text or "no subscription-userinfo" in text:
        return "usage_unchanged"
    if "timed out" in text or "timeout" in text:
        return "timeout"
    return "unknown"


def _technical_error(log: list[str], exit_code: int) -> str:
    picked = [line for line in log if line.startswith("Error:") or "probe failed" in line]
    if picked:
        return "\n".join(picked[-6:])
    tail = [line for line in log[-8:] if line.strip()]
    if tail:
        return "\n".join(tail)
    return f"Test process exited with code {exit_code}."


def _run_job(job_id: str):
    with JOBS_LOCK:
        job = JOBS.get(job_id)
        if not job:
            return
        if job.get("cancel"):
            job.update(status="cancelled", stage="cancelled", finished=time.time())
            job_dir = Path(job["_dir"])
            early = True
        else:
            early = False
            job.update(status="running", started=time.time(), stage="reading", progress=3)
            pick = job["pick"]
            size_mb = job["size_mb"]
            claimed = job["claimed"]
            core = job.get("core") or "auto"
            sub_file = Path(job["_sub_file"])
            result_file = Path(job["_result_file"])
            job_dir = Path(job["_dir"])
    if early:
        shutil.rmtree(job_dir, ignore_errors=True)
        return
    command = [
        sys.executable,
        str(SCRIPT),
        "--sub-file",
        str(sub_file),
        "--pick",
        str(pick),
        "--size",
        f"{size_mb:g}",
        "--wait",
        "20",
        "--drift-window",
        "10",
        "--json-out",
        str(result_file),
        "--core",
        core,
        "--probe-retries",
        "1",
        "--probe-wait",
        "5",
    ]
    if claimed is not None:
        command += ["--claimed", f"{claimed:g}"]
    env = os.environ.copy()
    env.update(
        {
            "RATIO_SAFE_FETCH": "1",
            "NO_COLOR": "1",
            "PYTHONUNBUFFERED": "1",
            "PYTHONDONTWRITEBYTECODE": "1",
            "HOME": str(Path.home()),
            "XDG_CACHE_HOME": str(Path.home() / ".cache"),
        }
    )
    proc = None
    timer = None
    try:
        proc = subprocess.Popen(
            command,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            stdin=subprocess.DEVNULL,
            text=True,
            bufsize=1,
            cwd=str(BASE_DIR),
            env=env,
            start_new_session=True,
        )
        with JOBS_LOCK:
            if job_id in JOBS:
                JOBS[job_id]["_proc"] = proc

        def kill():
            _set(job_id, timed_out=True)
            _terminate(proc)

        timer = threading.Timer(JOB_TIMEOUT, kill)
        timer.daemon = True
        timer.start()

        assert proc.stdout is not None
        for line in proc.stdout:
            _append_log(job_id, line)
            _progress_from_line(job_id, line)
        exit_code = proc.wait()
        timer.cancel()
        timer = None

        with JOBS_LOCK:
            current = JOBS.get(job_id)
            cancelled = bool(current and current.get("cancel"))
            timed_out = bool(current and current.get("timed_out"))
            log = list(current["log"]) if current else []

        if cancelled:
            _set(
                job_id,
                status="cancelled",
                stage="cancelled",
                progress=100,
                finished=time.time(),
                error="Cancelled.",
                error_code="cancelled",
            )
        elif exit_code == 0 and result_file.exists():
            try:
                result = json.loads(result_file.read_text(encoding="utf-8"))
            except (OSError, ValueError) as exc:
                _set(
                    job_id,
                    status="error",
                    stage="error",
                    finished=time.time(),
                    error=f"Could not read the result: {exc}",
                    error_code="unknown",
                )
            else:
                _set(
                    job_id,
                    status="done",
                    stage="done",
                    progress=100,
                    message="Done",
                    finished=time.time(),
                    result=result,
                )
        else:
            code = _classify_error(log, timed_out, cancelled)
            _set(
                job_id,
                status="error",
                stage="error",
                finished=time.time(),
                error=_technical_error(log, exit_code),
                error_code=code,
            )
    except Exception as exc:  # noqa: BLE001
        _set(
            job_id,
            status="error",
            stage="error",
            finished=time.time(),
            error=str(exc),
            error_code="unknown",
        )
    finally:
        if timer is not None:
            timer.cancel()
        if proc is not None and proc.poll() is None:
            _terminate(proc)
        try:
            sub_file.unlink(missing_ok=True)
        except OSError:
            pass
        shutil.rmtree(job_dir, ignore_errors=True)


def _public_job(job: dict):
    now = time.time()
    started = job.get("started")
    finished = job.get("finished")
    if finished and started:
        elapsed = finished - started
    elif started:
        elapsed = now - started
    else:
        elapsed = 0
    return {
        "id": job["id"],
        "status": job["status"],
        "stage": job["stage"],
        "progress": round(float(job["progress"]), 1),
        "message": job["message"],
        "elapsed": round(elapsed, 1),
        "created": job["created"],
        "result": job["result"],
        "error": job["error"],
        "error_code": job["error_code"],
        "log": job["log"][-80:],
        "pick": job["pick"],
        "size_mb": job["size_mb"],
        "core": job.get("core") or "auto",
    }


@app.get("/api/jobs/{job_id}")
def get_job(job_id: str):
    with JOBS_LOCK:
        job = JOBS.get(job_id)
        if not job:
            return api_error("not_found", "Job not found.", 404)
        data = _public_job(job)
    return data


@app.post("/api/jobs/{job_id}/cancel")
def cancel_job(job_id: str):
    with JOBS_LOCK:
        job = JOBS.get(job_id)
        if not job:
            return api_error("not_found", "Job not found.", 404)
        job["cancel"] = True
        proc = job.get("_proc")
        if job["status"] == "queued":
            job.update(
                status="cancelled",
                stage="cancelled",
                finished=time.time(),
                error="Cancelled.",
                error_code="cancelled",
            )
            queued_dir = Path(job["_dir"])
        else:
            queued_dir = None
    if proc is not None:
        _terminate(proc)
    if queued_dir is not None:
        shutil.rmtree(queued_dir, ignore_errors=True)
    return {"ok": True}


def _prune_loop():
    while True:
        time.sleep(300)
        cutoff = time.time() - 3600
        with JOBS_LOCK:
            stale = [
                job_id
                for job_id, job in JOBS.items()
                if job.get("finished") and job["finished"] < cutoff
            ]
            for job_id in stale:
                JOBS.pop(job_id, None)


@app.on_event("startup")
def startup():
    _executor()
    thread = threading.Thread(target=_prune_loop, daemon=True)
    thread.start()


app.mount("/static", StaticFiles(directory=STATIC_DIR), name="static")


@app.get("/")
def index():
    return FileResponse(STATIC_DIR / "index.html", headers={"Cache-Control": "no-cache"})


@app.head("/")
def head_index():
    return FileResponse(STATIC_DIR / "index.html", headers={"Cache-Control": "no-cache"})


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8000)
