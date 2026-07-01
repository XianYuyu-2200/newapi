# Production GHCR Deploy

This deployment path keeps builds away from the production server.

## 1. Build the image in GitHub Actions

1. Open the GitHub repository.
2. Go to `Actions`.
3. Select `Publish production Docker image`.
4. Click `Run workflow`.
5. Use this tag for the current subscription group billing fix:

```text
subscription-group-lock-697f4b8
```

The workflow publishes:

```text
ghcr.io/xianyuyu-2200/newapi:subscription-group-lock-697f4b8
```

## 2. Deploy on production

Run these commands on the NewAPI production server.

```bash
set -e

cd /opt/newapi
BACKUP_DIR="/opt/newapi/backups/$(date +%Y%m%d-%H%M%S)-pre-ghcr-deploy"
mkdir -p "$BACKUP_DIR"
cp -a docker-compose.yml "$BACKUP_DIR/docker-compose.yml"

docker pull ghcr.io/xianyuyu-2200/newapi:subscription-group-lock-697f4b8

python3 - <<'PY'
from pathlib import Path

path = Path("/opt/newapi/docker-compose.yml")
text = path.read_text()
old_images = [
    "xianyuyu/newapi:subscription-redemption-1c72333",
    "ghcr.io/xianyuyu-2200/newapi:subscription-group-lock-697f4b8",
]
new_image = "ghcr.io/xianyuyu-2200/newapi:subscription-group-lock-697f4b8"

for old_image in old_images:
    text = text.replace(old_image, new_image)

path.write_text(text)
PY

docker compose up -d newapi
docker ps --filter name=research-copilot-newapi --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
curl -fsS --max-time 15 http://127.0.0.1:3006/api/status
```

## 3. Public health check

Run from any client machine:

```bash
curl -fsS --max-time 15 https://api.jscvc.top/api/status
curl -fsS --max-time 15 https://api.xianjiji.top/api/status
```

## 4. Rollback

If the new container is unhealthy, restore the previous compose file and restart:

```bash
cd /opt/newapi
cp -a /opt/newapi/backups/20260701-151439-subscription-deploy/docker-compose.yml /opt/newapi/docker-compose.yml
docker compose up -d newapi
curl -fsS --max-time 15 http://127.0.0.1:3006/api/status
```

## Notes

- Do not run `docker build` on the production server.
- If GHCR package visibility is private, either make the package public in GitHub Packages or run `docker login ghcr.io` on the production server with a GitHub token that can read packages.
- The production server only needs enough resources to pull and restart the prebuilt image.
