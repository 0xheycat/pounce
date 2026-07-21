# Security Policy

## Reporting a vulnerability

Use GitHub private vulnerability reporting. Do not open a public issue for authentication bypasses, path traversal, arbitrary file writes, token exposure, remote-access weaknesses, or download validation issues.

Include the affected commit, operating system, reproduction steps, impact, and a minimal proof of concept. Never attach private download URLs, bearer tokens, downloaded files, or personal directory paths.

## Security model

- Local mode binds to `127.0.0.1` by default.
- Remote mode requires bearer-token authentication.
- Internet exposure should use HTTPS through a reverse proxy or private tunnel.
- Users are responsible for downloading only content they are authorized to access.

Pounce is alpha software. Review configuration and filesystem permissions before exposing it beyond a trusted network.
