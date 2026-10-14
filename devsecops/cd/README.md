# WebShield Container CD

## Purpose

Publish backend and webapp container images to GitHub Container Registry
(GHCR) after the `WebShield P4 CI` workflow succeeds on a push to `main`.
This workflow delivers images to a registry; it does not deploy to a server.

## Images

- `ghcr.io/haytamxp/orchid-webshield-backend`
- `ghcr.io/haytamxp/orchid-webshield-webapp`

Each delivery publishes a commit-specific `sha-<12-character-commit>` tag
and a convenience `latest` tag. Prefer the commit-specific tag for rollback.
The initial workflow targets `linux/amd64`.

## Security controls

- Validate both Go modules before building.
- Build both images before publication begins.
- Use the exact source commit that passed CI.
- Grant `packages: write` only to the publishing job.
- Run the applications as a non-root user in minimal runtime images.
- Do not put secrets or local database files in the image build context.

## Build and run locally (PowerShell)

Run from the repository root after extracting these files:

```powershell
docker build --pull --platform linux/amd64 --file backend/Dockerfile --tag orchid-webshield-backend:local .
docker build --pull --platform linux/amd64 --file webapp/Dockerfile --tag orchid-webshield-webapp:local .

docker run -d --name webshield-backend --restart unless-stopped -p 127.0.0.1:18080:8080 -e HTTP_ADDR=0.0.0.0:8080 orchid-webshield-backend:local
Invoke-RestMethod http://127.0.0.1:18080/api/v1/health

docker volume create webshield-webapp-data
docker run -d --name webshield-webapp --restart unless-stopped -p 127.0.0.1:18081:8081 -v webshield-webapp-data:/data -e WEBAPP_COOKIE_SECURE=false orchid-webshield-webapp:local
Invoke-RestMethod http://127.0.0.1:18081/health
```

`WEBAPP_COOKIE_SECURE=false` is for isolated local HTTP testing only. Use HTTPS
and `WEBAPP_COOKIE_SECURE=true` for public deployments.

## Rollback

Run a previously validated commit-specific `sha-...` image rather than `latest`.

## Next stage: continuous deployment

Automated deployment requires a selected host or platform, a protected GitHub
Environment, deployment credentials, health checks and rollback logic. Configure
those for the chosen target before automatically changing a live environment.
