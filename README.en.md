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
├─ scripts/                Docker build and push scripts
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

Create a local override:

```powershell
Copy-Item backend/config/config.local.example.yaml backend/config/config.local.yaml
```

At minimum, configure:

- `jwt.secret`
- `jisu.appKey`
- `ai.baseURL`
- `ai.apiKey`
- `vision.baseURL`
- `vision.apiKey`

### 2. Start backend

```powershell
cd backend
go run ./cmd
```

Default endpoints:

- API: [http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger: [http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 3. Start frontend

```powershell
cd frontend
pnpm install
pnpm dev
```

Default frontend URL:

- [http://127.0.0.1:3000](http://127.0.0.1:3000)

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

```powershell
Copy-Item .env.example .env
```

### 2. Start services

```powershell
docker compose up -d --build
```

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

## Docker Build Scripts

PowerShell scripts are included:

- [scripts/docker-build.ps1](scripts/docker-build.ps1)
- [scripts/docker-push.ps1](scripts/docker-push.ps1)
- [scripts/build-and-push.ps1](scripts/build-and-push.ps1)

Default image name:

```text
ydfk/lottery
```

Build and push in one command:

```powershell
.\scripts\build-and-push.ps1
```

Build only:

```powershell
.\scripts\docker-build.ps1
```

Push only:

```powershell
.\scripts\docker-push.ps1
```

Override image name and tag:

```powershell
$env:DOCKER_IMAGE_NAME = "ydfk/lottery"
$env:DOCKER_IMAGE_TAG = "latest"
.\scripts\build-and-push.ps1
```

The scripts generate:

- a primary tag such as `latest`
- a short Git SHA tag such as `a1b2c3d`

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
