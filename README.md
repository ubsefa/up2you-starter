# UP2YOU Starter

Run the public UP2YOU runtime images locally, test YAML apps, and prepare app packages for the hosted UP2YOU Platform.

This repository is intentionally small. It does not contain the private hosted Platform source code. It uses public UP2YOU runtime images, a standard nginx proxy image, and example YAML app definitions.

## What You Get

- A Core-only Docker Compose stack.
- A demo tenant with a plugin-free `my-todo` app ready to run.
- Reference app packages under `examples/` that can be zipped for Platform upload.
- Guides for Core runtime usage, YAML app development, packaging, and optional plugins.

## Prerequisites

Required:

- Docker
- Docker Compose v2 (`docker compose`)

Recommended:

- `make` for shorter commands
- `curl` for smoke tests
- `zip` for packaging app folders

## Quick Start

```bash
cp .env.example .env
docker compose up -d
./scripts/smoke-test.sh
```

With Make:

```bash
make setup
make up
make smoke
```

Open the demo screen:

```text
http://localhost:8080/demo/
```

The starter listens on:

- `http://localhost:8080/api/` for Core runtime APIs.
- `http://localhost:8080/ui/` for SDUI Gateway APIs.
- `http://localhost:8080/plugins/` for plugin-host APIs.
- `http://localhost:8080/demo/` for the tiny My Todo demo screen.

The default tenant id is `00000000-0000-0000-0000-000000000001`.

## Tiny Demo Screen

The `web/` folder is a plain HTML/CSS/JS client for the demo `my-todo` app. It calls Core APIs directly and intentionally does not use the SDUI renderer. That keeps the starter easy to inspect when you only want to verify entity CRUD, transitions, and public queries.

SDUI Gateway is still included in the Compose stack at `/ui/` for apps or tools that want to render server-driven views.

## Default Security Model

The starter defaults to:

```env
AUTH_ENABLED=false
```

That keeps local app development and API smoke tests simple. The hosted Platform uses authenticated users, licenses, marketplace workflows, quotas, and review flows. Those product-layer services are not included here.

## Reference Apps

The `examples/` directory contains small reference apps for common YAML patterns:

- `my-todo` — minimal task app and the default local demo.
- `approval-desk` — approval workflow and role-based queues.
- `inventory-lite` — inventory records, import/export, and aggregate charts.
- `public-notice-board` — public read-only view and SSE-friendly notice board.
- `event-checkin` — registration and code-based check-in.
- `booking-calendar` — reservation lifecycle with date-time fields.
- `simple-crm` — lightweight sales pipeline.
- `tournament-manager` — participants, matches, referees, and public scoreboard.
- `mental-health-care-plan` — domain workflow with follow-ups and risk alerts.

These examples are intentionally compact. Treat them as package patterns, not finished vertical products.

## Package An App

```bash
./scripts/package-app.sh examples/my-todo
```

With Make:

```bash
make package
```

The generated ZIP can be uploaded to the hosted UP2YOU Platform developer flow.

## Platform And Support

Hosted Platform:

```text
https://up2you.app
```

Hosted documentation:

```text
https://up2you.app/docs
```

For advanced app design, plugin deployment, production usage, or hosted Platform questions, contact:

```text
admin@up2you.app
```

## Repository Boundaries

Starter files are MIT licensed. The Docker images are published runtime artifacts owned by their publisher and may have separate licensing or usage terms.

This repo does not include:

- Hosted Platform frontend source.
- Auth-service / marketplace / payment / admin source.
- Platform installer source.
- Private deployment scripts or production secrets.

## Guides

- [Core-only usage](docs/core-only.md)
- [YAML app development](docs/app-development.md)
- [YAML contract](docs/yaml-contract.md)
- [SDUI views](docs/sdui.md)
- [Packaging for Platform](docs/packaging.md)
- [Optional plugins](docs/plugins.md)
- [Reference app patterns](docs/reference-patterns.md)
- [AI assistant prompt](docs/ai-prompt.md)
- [Troubleshooting](docs/troubleshooting.md)
