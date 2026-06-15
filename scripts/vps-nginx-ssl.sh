#!/usr/bin/env bash
# One-time VPS setup: Docker, host nginx, Let's Encrypt SSL
# Domains: school-management.tonusoft.com + school-management-api.tonusoft.com
#
# Run on Ubuntu 22.04/24.04 VPS as root or sudo user:
#   chmod +x scripts/vps-nginx-ssl.sh
#   sudo ./scripts/vps-nginx-ssl.sh
#
# Before running, point DNS A records to this server IP:
#   school-management.tonusoft.com      -> VPS_IP
#   school-management-api.tonusoft.com  -> VPS_IP

set -euo pipefail

APP_DIR="${APP_DIR:-/var/www/school-management-pos}"
GIT_REPO="${GIT_REPO:-}"   # optional: set to clone URL on first run
WEB_DOMAIN="school-management.tonusoft.com"
API_DOMAIN="school-management-api.tonusoft.com"
COMPOSE="docker compose -f docker-compose.yml -f docker-compose.prod.yml"

echo "==> School POS VPS setup (nginx + SSL)"
echo "    App dir: $APP_DIR"
echo "    Web:     https://$WEB_DOMAIN"
echo "    API:     https://$API_DOMAIN"

if [[ $EUID -ne 0 ]]; then
  echo "Run with sudo: sudo $0"
  exit 1
fi

echo "==> Installing packages..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq ca-certificates curl git nginx certbot python3-certbot-nginx ufw

if ! command -v docker >/dev/null 2>&1; then
  echo "==> Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
fi

if ! docker compose version >/dev/null 2>&1; then
  apt-get install -y -qq docker-compose-plugin
fi

mkdir -p /var/www/certbot
mkdir -p "$APP_DIR"

if [[ ! -d "$APP_DIR/.git" ]]; then
  if [[ -z "$GIT_REPO" ]]; then
    echo "ERROR: Clone repo first or set GIT_REPO=git@github.com:you/School-Management-POS.git"
    echo "  git clone \$GIT_REPO $APP_DIR"
    exit 1
  fi
  git clone "$GIT_REPO" "$APP_DIR"
fi

cd "$APP_DIR"

if [[ ! -f .env ]]; then
  echo "==> Creating .env from production template..."
  cp production.env.example .env
  sed -i "s|APP_ENV=production|APP_ENV=production|" .env
  sed -i "s|APP_URL=.*|APP_URL=https://$WEB_DOMAIN|" .env
  sed -i "s|APP_PORT=8085|APP_PORT=8085|" .env
  echo ""
  echo "!! Edit $APP_DIR/.env — set JWT_SECRET, CSRF_SECRET, POSTGRES_PASSWORD, DATABASE_URL"
  echo "!! Press Enter after editing .env (or Ctrl+C to abort)..."
  read -r _
fi

echo "==> Starting app (Docker)..."
$COMPOSE up -d --build

echo "==> Waiting for app /ready..."
for i in $(seq 1 30); do
  if curl -sf http://127.0.0.1:8085/ready >/dev/null 2>&1; then
    echo "    App is ready."
    break
  fi
  sleep 2
  if [[ $i -eq 30 ]]; then
    echo "ERROR: App not ready. Check: docker compose logs app"
    exit 1
  fi
done

echo "==> Installing nginx site (HTTP only first)..."
cat > /etc/nginx/sites-available/school-pos-tonusoft <<'NGINX_HTTP'
upstream school_pos_app {
    server 127.0.0.1:8085;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name school-management.tonusoft.com school-management-api.tonusoft.com;

    client_max_body_size 25m;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        proxy_pass http://school_pos_app;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGINX_HTTP

ln -sf /etc/nginx/sites-available/school-pos-tonusoft /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

echo "==> Firewall (UFW)..."
ufw allow OpenSSH || true
ufw allow 'Nginx Full' || ufw allow 80/tcp && ufw allow 443/tcp
ufw --force enable || true

CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@tonusoft.com}"

echo "==> Obtaining SSL certificates (Let's Encrypt)..."
certbot certonly --nginx \
  -d "$WEB_DOMAIN" \
  --non-interactive --agree-tos -m "$CERTBOT_EMAIL" || \
certbot certonly --webroot -w /var/www/certbot \
  -d "$WEB_DOMAIN" \
  --non-interactive --agree-tos -m "$CERTBOT_EMAIL"

certbot certonly --nginx \
  -d "$API_DOMAIN" \
  --non-interactive --agree-tos -m "$CERTBOT_EMAIL" || \
certbot certonly --webroot -w /var/www/certbot \
  -d "$API_DOMAIN" \
  --non-interactive --agree-tos -m "$CERTBOT_EMAIL"

echo "==> Installing production nginx config with SSL..."
cp "$APP_DIR/deploy/nginx-tonusoft.conf" /etc/nginx/sites-available/school-pos-tonusoft
nginx -t
systemctl reload nginx

echo "==> Certbot auto-renewal timer..."
systemctl enable certbot.timer 2>/dev/null || true
systemctl start certbot.timer 2>/dev/null || true

echo ""
echo "============================================"
echo " DONE"
echo " Web:  https://$WEB_DOMAIN"
echo " API:  https://$API_DOMAIN/health"
echo " Admin login: https://$WEB_DOMAIN/login"
echo " Deploy updates: cd $APP_DIR && ./scripts/deploy.sh"
echo "============================================"
