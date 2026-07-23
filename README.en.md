# Caiji Lottery

A mobile-first lottery management, ticket recording, and recommendation system for tracking purchases, draw results, and prize outcomes across multiple lottery types.

The goal of this project is not to predict lottery results with certainty. It focuses on building a maintainable workflow around recommendations, ticket capture, OCR-assisted entry, draw synchronization, prize evaluation, and historical tracking.

Chinese documentation is available in [README.md](README.md).

System design document: [docs/system-design.md](docs/system-design.md)

## Features

- Config-driven multi-lottery support, currently `Double Color Ball (SSQ)` and `Super Lotto (DLT)`
- Scheduled AI recommendation generation based on cron
- Scheduled latest draw synchronization plus manual historical draw import
- Ticket image upload, OCR-assisted parsing, manual correction, and persistence
- Excel-based bulk ticket import for historical purchases
- Automatic prize evaluation and manual recheck
- Recommendation-to-purchase linking
- Mobile-first frontend with web support
- Single-image Docker deployment

## Supported Lottery Types

| Code | Name | Remote ID | Red / Front Area | Blue / Back Area |
| --- | --- | --- | --- | --- |
| `ssq` | Double Color Ball | `11` | `6 (1-33)` | `1 (1-16)` |
| `dlt` | Super Lotto | `13` | `5 (1-35)` | `2 (1-12)` |

All lottery rules, draw schedules, recommendation settings, and sync schedules are loaded from [backend/config/config.yaml](backend/config/config.yaml).

## Core Capabilities

### 1. Recommendation Engine

- Independent `recommendation.cron` per lottery type
- Independent model, prompt, count, and history window per lottery type
- Automatic target issue and draw date inference based on official draw schedule
- Overwrite behavior for the same user, lottery, and issue
- Recommendation list, detail view, purchase association, and stealth full-screen number display

### 2. Ticket Recording

- Original image upload and retention
- OCR-assisted extraction of lottery type, issue, draw date, numbers, multiples, and amount
- Manual correction before saving
- Optional recording from a recommendation
- Duplicate ticket detection
- Single-use upload record consumption

### 3. Draw Sync and Prize Evaluation

- Latest draw sync through the `query` API
- Historical draw import through the `history` API
- Automatic settlement of related tickets and recommendations after draw sync
- Manual recheck support

### Bulk Ticket Import

Historical tickets can be imported from Excel without relying on OCR.

Template:

- [ticket-import-template.xlsx](./docs/ticket-import-template.xlsx)

Rules:

- One row equals one single entry
- There is no separate single-entry vs multi-entry template anymore
- Rows with the same user, lottery type, and issue are merged into one purchase record
- `lotteryCode` accepts Chinese names as well as `ssq` / `dlt`
- `drawDate` accepts `2026-05-09`, `2026/05/09`, `2026/5/9`, and `20260509`
- `redNumbers` / `blueNumbers` accept comma-separated, space-separated, or compact two-digit groups such as `01,02,03`, `01 02 03`, and `010203`
- `purchasedAt` can be empty; the draw date will be used
- `costAmount` can be empty or less than/equal to 0; it will be calculated from multiple and additional flags
- Images are optional and can be matched by the `imageName` column when uploading a ZIP file
- Recommendations are matched automatically

Recommendation matching rules:

- First match by `lottery type + issue`
- Then compare only `redNumbers + blueNumbers`
- `multiple` and `isAdditional` are ignored
- A purchase record may contain extra entries; as long as it fully contains all entries of a recommendation, the recommendation will be linked

### 4. User Isolation

- Uploads, tickets, recommendations, and dashboard statistics are scoped to the current user
- Different users cannot see each other’s data

## Tech Stack

- Frontend: React + Vite + TypeScript
- Backend: Go + Fiber + GORM
- Database: SQLite
- OCR: PaddleOCR HTTP service
- Recommendation API: OpenAI-compatible API
- Draw data: Jisu API

## Project Structure

```text
.
├─ backend/                Go Fiber backend
│  ├─ cmd/                 service bootstrap
│  ├─ config/              YAML configuration
│  ├─ docs/                generated Swagger docs
│  ├─ internal/
│  │  ├─ api/              HTTP handlers
│  │  ├─ model/            data models
│  │  └─ service/lottery/  lottery domain logic
│  └─ pkg/                 shared infrastructure modules
├─ frontend/               React frontend
├─ docs/                   design docs
├─ scripts/                cross-platform development, build, and Docker scripts
├─ Dockerfile              single-image build
├─ docker-compose.yml      deployment compose file
└─ .env.example            compose environment example
```

## Local Development

Docker is not required for local development.

### 1. Prepare configuration

The backend uses two YAML layers:

1. [backend/config/config.yaml](backend/config/config.yaml)
2. `backend/config/config.local.yaml`

Create a local override on macOS:

```bash
cp backend/config/config.local.example.yaml backend/config/config.local.yaml
```

On Windows PowerShell, use `Copy-Item backend/config/config.local.example.yaml backend/config/config.local.yaml`.

At minimum, configure:

- `jwt.secret`
- `jisu.appKey`
- `ai.baseURL`
- `ai.apiKey`
- `vision.baseURL`
- `vision.apiKey`

### 2. Start development services

macOS:

```bash
./scripts/dev-server.sh
```

Windows PowerShell:

```powershell
.\scripts\dev-server.ps1
```

Both the backend and frontend start by default. Append `backend` or `frontend` to start only one service.

Default endpoints:

- API: [http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger: [http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)
- Frontend: [http://127.0.0.1:3000](http://127.0.0.1:3000)

### 3. Build locally

```bash
./scripts/build.sh
```

Windows PowerShell uses `.\scripts\build.ps1`. Outputs are written to `frontend/dist` and `backend/bin`.

## Configuration

### Business configuration

Business configuration is YAML-only:

- [backend/config/config.yaml](backend/config/config.yaml)
- `backend/config/config.local.yaml`

Each lottery type can define:

- enable/disable status
- number rules
- official draw schedule
- recommendation cron
- recommendation model and prompt
- recommendation count
- draw sync cron
- default historical sync size

### Docker compose environment

The root `.env` file is only used by `docker compose` and is not part of business configuration.

Example: [\.env.example](.env.example)

Available variables:

- `LOTTERY_APP_PORT`
- `LOTTERY_CONFIG_DIR`
- `LOTTERY_DATA_DIR`
- `LOTTERY_LOG_DIR`

## Docker Deployment

The project uses a single-image deployment model.

The final image contains:

- the Go backend binary
- generated Swagger docs
- built frontend static files

Frontend assets are served directly by the backend. No extra Caddy container is required.

### 1. Prepare compose environment

```bash
cp .env.example .env
```

Windows PowerShell uses `Copy-Item .env.example .env`.

### 2. Start services

```bash
./scripts/docker.sh up
```

Windows PowerShell uses `.\scripts\docker.ps1 up`.

Default URLs:

- App: [http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger: [http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 3. Persistent mounts

Default mounts:

- `./backend/config -> /app/config`
- `./backend/data -> /app/data`
- `./backend/log -> /app/log`

This allows overriding and persisting:

- `config.yaml`
- `config.local.yaml`
- SQLite database files
- uploaded ticket images
- application logs

## Docker Scripts

Docker operations use one script with matching macOS and Windows entry points:

- [scripts/docker.sh](scripts/docker.sh)
- [scripts/docker.ps1](scripts/docker.ps1)

Default image name:

```text
ydfk/lottery
```

Build only:

```bash
./scripts/docker.sh build
```

Push only:

```bash
./scripts/docker.sh push
```

Use `.\scripts\docker.ps1 build` and `.\scripts\docker.ps1 push` on Windows PowerShell. The same scripts also support `up`, `down`, and `logs`.

Override image name and tag:

```bash
DOCKER_IMAGE_NAME=ydfk/lottery DOCKER_IMAGE_TAG=latest ./scripts/docker.sh build
```

When a non-`latest` tag is selected, the build and push commands also handle the `latest` tag.

## Tag-based Releases

The repository uses [release.yml](.github/workflows/release.yml) to manage releases through Git tags. Semantic version tags in the `vX.Y.Z` and `vX.Y.Z-rc.1` formats are accepted.

Configure these Actions secrets in the GitHub repository first:

- `DOCKERHUB_USERNAME`: Docker Hub username
- `DOCKERHUB_TOKEN`: Docker Hub access token

Publish a release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

The workflow then:

- validates the tag and resolves the application version as `1.2.3`
- runs backend and frontend tests, frontend lint, and the frontend build
- builds the Docker image with the resolved application version
- pushes the Docker Hub tags `ydfk/lottery:1.2.3`, `ydfk/lottery:1.2`, `ydfk/lottery:1`, and `ydfk/lottery:latest`
- creates or reuses the `v1.2.3` GitHub Release with generated release notes

Default image:

```text
ydfk/lottery
```

The frontend displays the semantic version without the `v` prefix, so tag `v1.2.3` produces application version `1.2.3`.

Prerelease tags publish prerelease images and a GitHub Prerelease without replacing the stable `latest` tag.

Pull example:

```bash
docker pull ydfk/lottery:latest
```

## Main APIs

### Auth

- `POST /api/auth/login`
- `POST /api/auth/register`
- `GET /api/auth/profile`

### Recommendations

- `GET /api/lotteries/recommendations`
- `GET /api/lotteries/:code/recommendations`
- `GET /api/lotteries/:code/recommendations/:recommendationId`
- `POST /api/lotteries/:code/recommendations/generate`

### Tickets

- `POST /api/lotteries/tickets/upload-image`
- `POST /api/lotteries/tickets/recognize`
- `POST /api/lotteries/tickets`
- `GET /api/lotteries/tickets/history`
- `POST /api/lotteries/tickets/:ticketId/recheck`

### Draw sync

- `POST /api/lotteries/:code/draws/sync`
- `POST /api/lotteries/:code/draws/sync-history`
- `POST /api/lotteries/draws/sync-history`

## Build and Test

Backend:

```powershell
cd backend
go test ./...
```

Frontend:

```powershell
cd frontend
pnpm build
pnpm test:run
```

## Cleanup Notes

Unused legacy multi-step ticket-entry panels and stale UI helpers have been removed. The repository now reflects the actual active workflow.

## GitHub

Target repository:

```text
git@github.com:ydfk/lottery.git
```

If the remote is not configured yet:

```powershell
git remote add origin git@github.com:ydfk/lottery.git
git branch -M main
git push -u origin main
```
