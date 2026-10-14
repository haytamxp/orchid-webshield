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

## Supply-chain artifacts and attestations

The container release workflow generates an SPDX JSON SBOM for each locally built image after the smoke tests and Trivy scans pass.

- Backend SBOM: `sbom-backend.spdx.json`.
- Webapp SBOM: `sbom-webapp.spdx.json`.
- The workflow checks that both files are non-empty and have the expected SPDX JSON structure.
- Both files are uploaded as the `webshield-sboms-<source SHA>` workflow artifact with 90-day retention.
- The workflow captures each image digest from the commit-specific Docker push.
- GitHub artifact attestations are created for build provenance and for each SBOM, associated with the matching image digest and pushed to GHCR.

The publishing job uses `id-token: write` for the short-lived signing identity and `attestations: write` to store attestation evidence. It retains `packages: write` for registry publication. Storage-record creation is disabled because this repository is under a personal account rather than an organization.

### Verify container build provenance

Run the following commands in PowerShell after installing GitHub CLI and authenticating with `gh auth login`. Replace the example commit tag with the one used by the release you are inspecting.

```powershell
docker login ghcr.io
$Tag = "sha-REPLACE_WITH_12_CHAR_COMMIT"
gh attestation verify "oci://ghcr.io/haytamxp/orchid-webshield-backend:$Tag" -R haytamxp/orchid-webshield
gh attestation verify "oci://ghcr.io/haytamxp/orchid-webshield-webapp:$Tag" -R haytamxp/orchid-webshield
```

### Verify the SPDX SBOM attestation

```powershell
gh attestation verify "oci://ghcr.io/haytamxp/orchid-webshield-backend:$Tag" -R haytamxp/orchid-webshield --predicate-type https://spdx.dev/Document/v2.3
gh attestation verify "oci://ghcr.io/haytamxp/orchid-webshield-webapp:$Tag" -R haytamxp/orchid-webshield --predicate-type https://spdx.dev/Document/v2.3
```

The SBOM workflow artifact provides downloadable inventory files. Attestations provide verifiable statements about the image digest, source/build process, and SBOM. Neither replaces vulnerability remediation or a full security review.

References:
- GitHub artifact attestations: https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations
- Anchore SBOM Action: https://github.com/anchore/sbom-action
