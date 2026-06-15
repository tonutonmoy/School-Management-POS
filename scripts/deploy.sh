#!/usr/bin/env bash
# Deploy latest code: git pull -> docker build -> restart
#
# On VPS:
#   chmod +x scripts/deploy.sh
#   ./scripts/deploy.sh
#
# Or from Windows: scripts\deploy.cmd

set -euo pipefail

APP_DIR="${APP_DIR:-/var/www/school-management-pos}"
BRANCH="${BRANCH:-main}"
COMPOSE="docker compose -f docker-compose.yml -f docker-compose.prod.yml"

cd "$APP_DIR"

echo "==> Deploy School POS"
echo "    Dir:    $APP_DIR"
echo "    Branch: $BRANCH"
echo "    Time:   $(date -Is)"

if [[ -d .git ]]; then
  echo "==> git fetch + pull..."
  git fetch origin
  git checkout "$BRANCH"
  git pull origin "$BRANCH"
else
  echo "ERROR: $APP_DIR is not a git repo"
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "ERROR: .env missing. Copy production.env.example to .env first."
  exit 1
fi

echo "==> Docker build & up..."
$COMPOSE build --pull
$COMPOSE up -d

echo "==> Prune old images..."
docker image prune -f >/dev/null 2>&1 || true

echo "==> Health check..."
sleep 3
for i in $(seq 1 20); do
  if curl -sf http://127.0.0.1:8085/ready >/dev/null 2>&1; then
    echo "    /ready OK"
    curl -sf http://127.0.0.1:8085/health | head -c 200 || true
    echo ""
    $COMPOSE ps
    echo ""
    echo "==> Deploy complete."
    exit 0
  fi
  sleep 2
done

echo "ERROR: App not healthy. Logs:"
$COMPOSE logs --tail=50 app
exit 1
