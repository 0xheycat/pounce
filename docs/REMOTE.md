# Pounce Anywhere — Remote Access Guide 📱

Because Pounce is a service you host, you can control it from any device. This guide covers LAN access, phone pairing, installing the PWA, and exposing Pounce safely over the internet.

## TL;DR

```bash
# On the machine that should do the downloading:
pounce --remote --static ./dashboard/dist
```

The engine will:
- bind every network interface (so other devices can reach it),
- require a bearer token (auto-generated and printed if you don't pass `--auth-token`),
- print ready-to-open pairing links for each LAN address.

Open one of the printed links on your phone, or open the dashboard and use the **📱 Pair device** panel to scan a QR code.

## 1. Local network (LAN)

1. Start with remote mode:
   ```bash
   pounce --remote --static ./dashboard/dist
   ```
2. Note the token and URLs printed at startup, e.g.:
   ```
   📱 Pounce Anywhere — pair a device:
      http://192.168.1.42:7766/?token=ab12...
      token: ab12...
   ```
3. On your phone (same Wi-Fi), open that URL. The dashboard captures the token from the link and stores it — no manual typing.

## 2. Phone pairing via QR

1. Open the dashboard on the host (or any already-authenticated device).
2. Click **📱 Pair device** in the header.
3. Scan the QR with your phone camera. It opens the dashboard pre-authenticated.

The QR encodes `http://<host>:<port>/?token=<token>`. Treat it like a password — anyone who scans it gains access.

## 3. Install as an app (PWA)

The dashboard is a Progressive Web App:
- **iOS Safari:** Share → *Add to Home Screen*.
- **Android Chrome:** menu → *Install app* / *Add to Home Screen*.
- **Desktop Chrome/Edge:** install icon in the address bar.

You get a full-screen Pounce icon, and the app shell loads even when the engine is briefly unreachable.

## 4. Internet access (do this securely)

Never port-forward a tokenless engine. Pick one:

### Option A — Reverse proxy with HTTPS (Caddy)

```caddy
pounce.example.com {
    reverse_proxy 127.0.0.1:7766
}
```

Run the engine bound to loopback with a token, and let Caddy terminate TLS:
```bash
pounce --auth-token "$(openssl rand -hex 16)" --static ./dashboard/dist
```

### Option B — Reverse proxy (nginx)

```nginx
server {
    server_name pounce.example.com;
    location / {
        proxy_pass http://127.0.0.1:7766;
        proxy_http_version 1.1;
        proxy_set_header Connection "";          # keep SSE streams open
        proxy_buffering off;                      # required for /api/events
        proxy_read_timeout 24h;
    }
}
```

> SSE (`/api/events`) needs buffering off and a long read timeout, or live progress will stall.

### Option C — Private tunnel (no open ports)

- **Tailscale:** install on both devices; reach Pounce at the host's tailnet IP. Add a token for defense in depth.
- **Cloudflare Tunnel:** `cloudflared tunnel --url http://127.0.0.1:7766`.

## 5. Security checklist

- ✅ Always run with a token for any non-loopback access (`--remote` enforces this).
- ✅ Use HTTPS for anything beyond your LAN (clipboard copy and PWA install also prefer secure contexts).
- ✅ Rotate the token by restarting with a new `--auth-token`; old paired links stop working.
- ⚠️ The token grants full control of downloads and the save directory. Share pairing links carefully.
- ⚠️ Pounce writes to the filesystem of the host it runs on — run it as a user with appropriate permissions.

## How token auth works

The dashboard sends the token three ways so every transport is covered:
- `Authorization: Bearer <token>` header (REST calls),
- `X-Pounce-Token` header,
- `?token=` query parameter (used by the SSE stream and pairing links).

`/api/health` is always public so proxies and uptime checks work without a token.
