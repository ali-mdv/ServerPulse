#!/usr/bin/env python3
"""
ServerPulse remote agent (Python).

Collects system usage and optional PM2 / Docker provider state, then pushes
them to a ServerPulse backend using an API key.

Workflow:
  1. Create a server/agent in the backend (POST /api/v1/servers) with a
     name and optional description.
  2. Generate an API key for that server (POST
     /api/v1/servers/<id>/api-key).
  3. Run this agent with SERVER_PULSE_API_KEY set to the revealed key.

Required env:
  SERVER_PULSE_URL      Base URL of the backend, e.g. http://localhost:8080
  SERVER_PULSE_API_KEY  API key revealed by POST /servers/<id>/api-key

Optional env:
  SERVER_PULSE_INTERVAL Push interval in seconds (default: 30)
  SERVER_PULSE_NAME     Agent/server name (default: hostname)
  SERVER_PULSE_HOST     Host address to report (default: resolved hostname)
  SERVER_PULSE_PORT     Port to report (default: 0)
  SERVER_PULSE_DESCRIPTION  Optional description pushed with identity

Dependencies:
  pip install psutil requests
  # PM2 and Docker support are optional; the agent will skip a provider if
  # its daemon/socket is not reachable.
  pip install docker
"""

import json
import os
import socket
import struct
import time
import sys
from typing import Any, Optional

import requests


def getenv(key: str, default: Optional[str] = None) -> Optional[str]:
    return os.environ.get(key, default)


def require_env(key: str) -> str:
    value = getenv(key)
    if not value:
        print(f"Error: {key} is required", file=sys.stderr)
        sys.exit(1)
    return value


def now_iso() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def collect_system() -> dict[str, Any]:
    import psutil

    mem = psutil.virtual_memory()
    disk = psutil.disk_usage("/")
    cpu = psutil.cpu_percent(interval=1)

    net_io = psutil.net_io_counters()
    net_usage: dict[str, Any] = {"sent": 0, "received": 0}
    now = time.time()
    if hasattr(collect_system, "_last_net"):
        last_io, last_at = collect_system._last_net
        delta = max(now - last_at, 0.001)
        net_usage["sent"] = int(
            max(net_io.bytes_sent - last_io.bytes_sent, 0) / delta)
        net_usage["received"] = int(
            max(net_io.bytes_recv - last_io.bytes_recv, 0) / delta)
    collect_system._last_net = (net_io, now)

    return {
        "memUsage": {
            "percent": mem.percent,
            "used": mem.used,
            "total": mem.total,
        },
        "cpuUsage": cpu,
        "diskUsage": {
            "percent": disk.percent,
            "used": disk.used,
            "total": disk.total,
        },
        "netIO": net_usage,
    }


def collect_docker() -> Optional[dict[str, Any]]:
    try:
        import docker
    except ImportError:
        return None

    try:
        client = docker.from_env()
        containers = client.containers.list(all=True)
    except Exception as exc:
        return {"available": False, "services": [], "error": str(exc)}

    services: list[dict[str, Any]] = []
    for container in containers:
        state = container.status or "unknown"
        available = state == "running"
        cpu_percent = 0.0
        mem_percent = 0.0
        mem_usage = 0.0

        try:
            stats = container.stats(stream=False)
            cpu_stats = stats.get("cpu_stats", {})
            prev_cpu = stats.get("precpu_stats", {})

            cpu_delta = (
                cpu_stats.get("cpu_usage", {}).get("total_usage", 0)
                - prev_cpu.get("cpu_usage", {}).get("total_usage", 0)
            )
            system_delta = (
                cpu_stats.get("system_cpu_usage", 0)
                - prev_cpu.get("system_cpu_usage", 0)
            )
            online_cpus = cpu_stats.get("online_cpus") or len(
                cpu_stats.get("cpu_usage", {}).get("percpu_usage", []) or []
            )
            if system_delta > 0 and online_cpus:
                cpu_percent = (cpu_delta / system_delta) * online_cpus * 100.0

            mem_stats = stats.get("memory_stats", {})
            mem_usage = float(mem_stats.get("usage", 0))
            mem_limit = float(mem_stats.get("limit", 1))
            if mem_limit > 0:
                mem_percent = (mem_usage / mem_limit) * 100.0
        except Exception:
            pass

        services.append({
            "ts": now_iso(),
            "meta": {
                "serverId": "",
                "provider": "docker",
                "serviceId": container.short_id,
                "name": container.name.lstrip("/"),
            },
            "status": state,
            "available": available,
            "cpu": round(cpu_percent, 2),
            "memory": round(mem_percent, 2),
            "memUsage": round(mem_usage, 2),
        })

    return {"available": True, "services": services}


# --- PM2 provider (axon-rpc over the PM2 unix socket) ---

class PM2Error(Exception):
    pass


def _pm2_socket_path() -> str:
    return os.environ.get("PM2_SOCKET_PATH") or os.path.expanduser("~/.pm2/rpc.sock")


def _encode_amp(args: list[bytes]) -> bytes:
    version = 1
    argc = len(args)
    if argc == 0 or argc > 15:
        raise PM2Error(f"amp: invalid argc {argc}")
    parts = [bytes([version << 4 | argc])]
    for arg in args:
        parts.append(struct.pack(">I", len(arg)))
        parts.append(arg)
    return b"".join(parts)


def _decode_amp(buf: bytes) -> list[bytes]:
    if not buf:
        raise PM2Error("amp: empty frame")
    meta = buf[0]
    version = meta >> 4
    argc = meta & 0x0F
    if version != 1:
        raise PM2Error(f"amp: unsupported version {version}")
    if argc > 15:
        raise PM2Error(f"amp: invalid argc {argc}")
    args: list[bytes] = []
    off = 1
    for _ in range(argc):
        if off + 4 > len(buf):
            raise PM2Error("amp: short read on arg length")
        length = struct.unpack(">I", buf[off:off + 4])[0]
        off += 4
        if off + length > len(buf):
            raise PM2Error("amp: short read on arg body")
        args.append(buf[off:off + length])
        off += length
    return args


def _pack_json(payload: bytes) -> bytes:
    return b"j:" + payload


def _pack_string(s: str) -> bytes:
    return ("s:" + s).encode("utf-8")


def _unpack_json(buf: bytes) -> bytes:
    if len(buf) < 2 or buf[:2] != b"j:":
        raise PM2Error(f"amp-msg: not a json-typed argument (got {buf[:2]!r})")
    return buf[2:]


class _PM2RPCClient:
    def __init__(self, sock_path: str) -> None:
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.settimeout(5.0)
        self.sock.connect(sock_path)
        self._counter = 0

    def _next_id(self) -> str:
        self._counter += 1
        return f"{os.getpid()}:{self._counter}"

    def call(self, method: str, args: Optional[list[Any]] = None) -> Any:
        if args is None:
            args = []
        req = {"type": "call", "method": method, "args": args}
        payload = json.dumps(req, separators=(",", ":")).encode("utf-8")
        frame = _encode_amp(
            [_pack_json(payload), _pack_string(self._next_id())])
        self.sock.sendall(frame)

        # Read one AMP frame.
        meta = self.sock.recv(1)
        if not meta:
            raise PM2Error("pm2: connection closed while reading meta")
        argc = meta[0] & 0x0F
        raw_args: list[bytes] = []
        for _ in range(argc):
            len_bytes = self._recvall(4)
            length = struct.unpack(">I", len_bytes)[0]
            raw_args.append(self._recvall(length))

        if len(raw_args) != 2:
            raise PM2Error(
                f"pm2 rpc: expected 2 reply args, got {len(raw_args)}")

        body = json.loads(_unpack_json(raw_args[0]))
        if body.get("error"):
            raise PM2Error(f"pm2 rpc {method}: {body['error']}")
        reply_args = body.get("args", [])
        return reply_args[0] if reply_args else None

    def _recvall(self, n: int) -> bytes:
        buf = b""
        while len(buf) < n:
            chunk = self.sock.recv(n - len(buf))
            if not chunk:
                raise PM2Error("pm2: connection closed while reading")
            buf += chunk
        return buf

    def close(self) -> None:
        try:
            self.sock.close()
        except Exception:
            pass


def collect_pm2() -> Optional[dict[str, Any]]:
    sock_path = _pm2_socket_path()
    if not os.path.exists(sock_path):
        return None

    client: Optional[_PM2RPCClient] = None
    try:
        client = _PM2RPCClient(sock_path)
        procs = client.call("getMonitorData", [{}])
    except PM2Error as exc:
        return {"available": False, "services": [], "error": str(exc)}
    except Exception as exc:
        return {"available": False, "services": [], "error": str(exc)}
    finally:
        if client is not None:
            client.close()

    if procs is None:
        procs = []

    services: list[dict[str, Any]] = []
    for proc in procs:
        pm_id = proc.get("pm_id")
        name = proc.get("name", "unknown")
        pm2_env = proc.get("pm2_env", {})
        monit = proc.get("monit", {})
        status = pm2_env.get("status", "unknown")

        services.append({
            "ts": now_iso(),
            "meta": {
                "serverId": "",
                "provider": "pm2",
                "serviceId": str(pm_id) if pm_id is not None else "",
                "name": name,
            },
            "status": status,
            "available": status == "online",
            "cpu": float(monit.get("cpu", 0.0)),
            "memory": float(monit.get("memory", 0.0)),
        })

    return {"available": True, "services": services}


def build_payload(name: str, host: str, port: int, description: str) -> dict[str, Any]:
    providers: dict[str, Any] = {}
    docker_payload = collect_docker()
    if docker_payload is not None:
        providers["docker"] = docker_payload
    pm2_payload = collect_pm2()
    if pm2_payload is not None:
        providers["pm2"] = pm2_payload

    return {
        "name": name,
        "host": host,
        "port": port,
        "description": description,
        "usage": collect_system(),
        "providers": providers,
    }


def push(base_url: str, api_key: str, payload: dict[str, Any]) -> None:
    url = base_url.rstrip("/") + "/api/v1/agents/push"
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    }
    response = requests.post(url, headers=headers, json=payload, timeout=30)
    response.raise_for_status()
    print(f"[{now_iso()}] push accepted: {response.status_code} {response.text}")


def main() -> None:
    base_url = require_env("SERVER_PULSE_URL")
    api_key = require_env("SERVER_PULSE_API_KEY")
    interval = int(getenv("SERVER_PULSE_INTERVAL", "30") or "30")
    name = getenv("SERVER_PULSE_NAME") or socket.gethostname()
    host = getenv("SERVER_PULSE_HOST") or socket.gethostname()
    port = int(getenv("SERVER_PULSE_PORT", "0") or "0")
    description = getenv("SERVER_PULSE_DESCRIPTION", "")

    print(
        f"ServerPulse agent starting: name={name} url={base_url} interval={interval}s")

    while True:
        start = time.time()
        try:
            payload = build_payload(name, host, port, description)
            push(base_url, api_key, payload)
        except requests.HTTPError as exc:
            print(
                f"[{now_iso()}] push failed: {exc.response.status_code} {exc.response.text}", file=sys.stderr)
        except Exception as exc:
            print(f"[{now_iso()}] push error: {exc}", file=sys.stderr)

        elapsed = time.time() - start
        sleep_for = max(interval - elapsed, 1)
        time.sleep(sleep_for)


if __name__ == "__main__":
    main()
