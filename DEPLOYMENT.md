# Deployment Guide

## Prerequisites

- Docker 24+ and Docker Compose v2
- PostgreSQL 16 client tools (`pg_dump`, `psql`) for manual backups outside Docker
- Domain name with DNS pointing to your server (optional, for HTTPS)

## Quick Start (Docker)

1. Copy environment file:
   ```bash
   cp production.env.example .env
   ```
2. Edit `.env` — set strong `JWT_SECRET`, `CSRF_SECRET`, `POSTGRES_PASSWORD`, and `APP_URL`.
3. Build and start:
   ```bash
   docker compose up -d --build
   ```
4. Open `http://localhost` (via nginx) or `http://localhost:8085` (direct app).
5. First-time setup: visit `/install` or use seeded admin credentials from `.env`.

## Services

| Service | Port | Purpose |
|---------|------|---------|
| nginx   | 80   | Reverse proxy |
| app     | 8085 | Go application |
| db      | 5432 | PostgreSQL 16 |

## Health & Monitoring

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Liveness — DB, storage, queue checks |
| `GET /ready`  | Readiness — returns 503 if DB down |
| `GET /metrics`| JSON metrics (uptime, queue, failed logins) |

Docker healthchecks use `/ready` on the app container.

## Admin Modules

| Route | Permission | Feature |
|-------|------------|---------|
| `/system/health` | `system.manage` | Health dashboard |
| `/system/settings` | `system.manage` | SMTP, SMS, branding, password policy |
| `/system/backups` | `system.backup` | Manual/scheduled backup, restore, download |
| `/system/audit` | `system.audit` | Searchable audit center + CSV export |
| `/system/license` | `system.license` | License activation |
| `/system/security` | `system.security` | Failed logins, IP audit |

## Backups

- Backups stored in `BACKUP_DIR` (default `/app/backups` in Docker volume `backups`).
- Requires `pg_dump` and `psql` (included in Docker image).
- Enable scheduled backups at **System → Settings → Scheduled Backup**.
- Always verify backups before relying on them for disaster recovery.

## Production Checklist

- [ ] Set `APP_ENV=production`
- [ ] Remove `LOGIN_EMAIL` / `LOGIN_PASSWORD` pre-fill
- [ ] Configure real SMTP and SMS providers
- [ ] Enable Cloudflare R2 or object storage for uploads
- [ ] Activate license at `/system/license`
- [ ] Configure scheduled backups
- [ ] Restrict `/metrics` at nginx (already IP-limited in sample config)
- [ ] Add TLS certificates (Let's Encrypt) in front of nginx
- [ ] Change default admin password after first login

## TLS (Recommended)

Terminate TLS at nginx using Certbot or a cloud load balancer. Update `APP_URL` to `https://...` after enabling TLS.

## Manual Migration

Migrations run automatically on app startup via Goose. To run manually:

```bash
goose -dir migrations postgres "$DATABASE_URL" up
```

## Scaling Notes

- Run a single app instance unless session/JWT state is shared.
- Use managed PostgreSQL for high availability.
- Mount persistent volume for `backups`.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| App won't start | Check `DATABASE_URL`, JWT/CSRF secret lengths |
| Backup fails | Ensure `pg_dump` available; check disk space in `BACKUP_DIR` |
| Restore fails | Verify backup checksum; test on staging first |
| 503 on `/ready` | Database not reachable — check `db` service health |
