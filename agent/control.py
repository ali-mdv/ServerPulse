# ServerPulse agent control channel (Socket.IO).
#
# Connects to the backend's /api/v1/agents/socket.io endpoint and
# executes start/stop/restart/logs commands for Docker and PM2,
# acking each command with an AgentCommandResult.

from __future__ import annotations

import json
import os
import sys
import threading
import time
from typing import Any, Optional

try:
    import socketio
except ImportError:  # optional dependency
    socketio = None  # type: ignore

AGENT_COMMAND_EVENT = "agent:command"
SOCKETIO_PATH = "api/v1/agents/socket.io"


def start_control_channel(base_url: str, api_key: str) -> None:
    """Run the Socket.IO control channel forever (with reconnect).

    Call from a daemon thread; the push loop continues independently.
    """
    if socketio is None:
        print(
            "control channel: python-socketio not installed "
            "(pip install 'python-socketio[client]'); remote actions disabled",
            file=sys.stderr,
        )
        return

    sio = socketio.Client(
        reconnection=True,
        reconnection_delay=1,
        reconnection_delay_max=30,
        logger=False,
        engineio_logger=False,
    )

    @sio.event
    def connect() -> None:
        print(f"control channel: connected id={sio.sid}")

    @sio.event
    def connect_error(data: Any) -> None:
        print(f"control channel: connect error: {data}", file=sys.stderr)

    @sio.event
    def disconnect() -> None:
        print("control channel: disconnected")

    @sio.on(AGENT_COMMAND_EVENT)
    def on_command(data: Any) -> dict[str, Any]:
        # Return value is sent as the Socket.IO ack payload.
        cmd_id = ""
        try:
            if not isinstance(data, dict):
                return {
                    "id": cmd_id,
                    "ok": False,
                    "error": "malformed command",
                }
            cmd_id = str(data.get("id") or "")
            provider = str(data.get("provider") or "")
            action = str(data.get("action") or "")
            target = str(data.get("target") or "")
            params = data.get("params") or {}
            if not isinstance(params, dict):
                params = {}

            print(
                f"control channel: command id={cmd_id} provider={provider} "
                f"action={action} target={target}"
            )
            result = execute_command(
                {
                    "id": cmd_id,
                    "provider": provider,
                    "action": action,
                    "target": target,
                    "params": params,
                }
            )
            return result
        except Exception as exc:  # noqa: BLE001 — surface any failure to the backend
            return {"id": cmd_id, "ok": False, "error": str(exc)}

    auth = {"token": api_key, "apiKey": api_key}
    while True:
        try:
            sio.connect(
                base_url.rstrip("/"),
                auth=auth,
                transports=["websocket", "polling"],
                socketio_path=SOCKETIO_PATH,
            )
            sio.wait()
        except Exception as exc:  # noqa: BLE001
            print(f"control channel: connect failed: {exc}", file=sys.stderr)
        time.sleep(2)


def execute_command(cmd: dict[str, Any]) -> dict[str, Any]:
    cmd_id = str(cmd.get("id") or "")
    provider = cmd.get("provider") or ""
    if provider == "docker":
        result = run_docker_command(cmd)
    elif provider == "pm2":
        result = run_pm2_command(cmd)
    else:
        result = {"id": cmd_id, "ok": False, "error": f"unknown provider: {provider}"}
    result.setdefault("id", cmd_id)
    return result


# --- Docker control ---


def run_docker_command(cmd: dict[str, Any]) -> dict[str, Any]:
    cmd_id = str(cmd.get("id") or "")
    action = cmd.get("action") or ""
    target = str(cmd.get("target") or "")
    params = cmd.get("params") or {}

    try:
        import docker
        from docker.errors import APIError, NotFound
    except ImportError:
        return {
            "id": cmd_id,
            "ok": False,
            "error": "docker unavailable: python package not installed",
        }

    try:
        client = docker.from_env()
    except Exception as exc:  # noqa: BLE001
        return {
            "id": cmd_id,
            "ok": False,
            "error": f"docker unavailable: {exc}",
        }

    try:
        container = client.containers.get(target)
        if action == "start":
            container.start()
        elif action == "stop":
            container.stop()
        elif action == "restart":
            container.restart()
        elif action == "logs":
            lines = 100
            try:
                raw_lines = params.get("lines")
                if raw_lines is not None and float(raw_lines) > 0:
                    lines = int(float(raw_lines))
            except (TypeError, ValueError):
                pass
            raw = container.logs(
                stdout=True,
                stderr=True,
                timestamps=True,
                tail=lines,
            )
            text = raw.decode("utf-8", errors="replace") if isinstance(raw, (bytes, bytearray)) else str(raw)
            return {"id": cmd_id, "ok": True, "data": {"logs": text}}
        else:
            return {
                "id": cmd_id,
                "ok": False,
                "error": f"unsupported action: {action}",
            }
        return {"id": cmd_id, "ok": True}
    except NotFound:
        return {"id": cmd_id, "ok": False, "error": "not found"}
    except APIError as exc:
        status = getattr(exc, "status_code", None)
        if status == 404:
            return {"id": cmd_id, "ok": False, "error": "not found"}
        return {"id": cmd_id, "ok": False, "error": str(exc)}
    except Exception as exc:  # noqa: BLE001
        return {"id": cmd_id, "ok": False, "error": str(exc)}
    finally:
        try:
            client.close()
        except Exception:  # noqa: BLE001
            pass


# --- PM2 control ---


def run_pm2_command(cmd: dict[str, Any]) -> dict[str, Any]:
    from agent import PM2Error, _PM2RPCClient, _pm2_socket_path  # local module

    cmd_id = str(cmd.get("id") or "")
    action = cmd.get("action") or ""
    target = str(cmd.get("target") or "")
    params = cmd.get("params") or {}

    try:
        pm_id = int(str(target).strip())
    except (TypeError, ValueError):
        return {
            "id": cmd_id,
            "ok": False,
            "error": "invalid PM2 id; expected a numeric value",
        }

    sock_path = _pm2_socket_path()
    if not os.path.exists(sock_path):
        return {
            "id": cmd_id,
            "ok": False,
            "error": f"pm2 daemon unreachable: {sock_path}",
        }

    client: Optional[Any] = None
    try:
        client = _PM2RPCClient(sock_path)
        if action == "start":
            client.call("startProcessId", [pm_id])
        elif action == "stop":
            client.call("stopProcessId", [pm_id])
        elif action == "restart":
            client.call("restartProcessId", [{"id": pm_id}])
        elif action == "logs":
            lines = 100
            try:
                raw_lines = params.get("lines")
                if raw_lines is not None and float(raw_lines) > 0:
                    lines = int(float(raw_lines))
            except (TypeError, ValueError):
                pass
            logs = _pm2_tail_logs(client, pm_id, lines)
            return {"id": cmd_id, "ok": True, "data": {"logs": logs}}
        else:
            return {
                "id": cmd_id,
                "ok": False,
                "error": f"unsupported action: {action}",
            }
        return {"id": cmd_id, "ok": True}
    except PM2Error as exc:
        return {"id": cmd_id, "ok": False, "error": str(exc)}
    except Exception as exc:  # noqa: BLE001
        return {"id": cmd_id, "ok": False, "error": str(exc)}
    finally:
        if client is not None:
            client.close()


def _pm2_tail_logs(client: Any, pm_id: int, lines: int) -> str:
    if lines <= 0:
        lines = 100
    procs = client.call("getMonitorData", [{}]) or []
    out_log = ""
    err_log = ""
    found = False
    for proc in procs:
        if not isinstance(proc, dict) or proc.get("pm_id") != pm_id:
            continue
        env = proc.get("pm2_env") or {}
        out_log = env.get("pm_out_log_path") or ""
        err_log = env.get("pm_err_log_path") or ""
        found = True
        break
    if not found:
        raise RuntimeError("no process found")

    parts: list[str] = []
    for path in (err_log, out_log):
        if not path:
            continue
        try:
            parts.append(_tail_file(path, lines))
        except FileNotFoundError:
            continue
    text = "\n".join(p for p in parts if p)
    return text.lstrip("\n")


def _tail_file(path: str, n: int) -> str:
    max_read = 64 * 1024
    with open(path, "rb") as f:
        f.seek(0, os.SEEK_END)
        size = f.tell()
        if size == 0:
            return ""
        read_size = min(max_read, size)
        f.seek(size - read_size)
        buf = f.read(read_size)
    if size > read_size and buf[:1] != b"\n":
        idx = buf.find(b"\n")
        if idx >= 0:
            buf = buf[idx + 1 :]
    lines = buf.splitlines(keepends=True)
    if not lines:
        return ""
    selected = lines[-n:]
    text = b"".join(selected).decode("utf-8", errors="replace")
    if text and not text.endswith("\n"):
        text += "\n"
    return text


def run_control_channel(base_url: str, api_key: str) -> threading.Thread:
    t = threading.Thread(
        target=start_control_channel,
        args=(base_url, api_key),
        name="serverpulse-control",
        daemon=True,
    )
    t.start()
    return t
