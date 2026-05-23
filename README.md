# UP2YOU Starter

Run the public UP2YOU runtime images locally, test YAML apps, and prepare app packages for the hosted UP2YOU Platform.

This repository is intentionally small. It does not contain private hosted product source code. It uses public UP2YOU runtime images, a standard nginx proxy image, and example YAML app definitions.

## What You Get

- A Core-only Docker Compose stack.
- A demo tenant with a plugin-free `my-todo` app ready to run.
- Reference app packages under `examples/` that can be zipped for Platform upload.
- Guides for Core runtime usage, YAML app development, packaging, and optional plugins.
- API reference, authentication guide, error codes, internationalization, and architecture overview.

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

That keeps local app development and API smoke tests simple. Hosted product workflows are intentionally outside this Core-only starter and are documented separately.

## Reference Apps

The `examples/` directory contains small reference apps for common YAML patterns:

- `my-todo` — minimal task app and the default local demo.
- `approval-desk` — approval workflow and role-based queues.
- `inventory-lite` — inventory records, import/export, and aggregate charts.
- `public-notice-board` — public read-only view and SSE-friendly notice board.
- `event-checkin` — public event discovery and code-based check-in.
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

The generated ZIP is ready for any UP2YOU package upload flow.

## Support

Project site:

```text
https://up2you.app
```

Hosted documentation:

```text
https://up2you.app/docs
```

For advanced app design, production usage, or hosted deployment questions, contact:

```text
admin@up2you.app
```

## Repository Boundaries

Starter files are MIT licensed. The Docker images are published runtime artifacts owned by their publisher and may have separate licensing or usage terms.

This repo does not include:

- Hosted product frontend source.
- Hosted product-layer source code.
- Private platform services or installers.
- Private deployment scripts or production secrets.

## Guides

Recommended reading order:

1. [Quick start](docs/quick-start.md) — clone, compose up, demo, first app, package
2. [YAML contract](docs/yaml-contract.md) — entities, workflows, queries, views, forms, i18n
3. [Architecture](docs/architecture.md) — Core, SDUI, plugin-host, NATS, PostgreSQL, NGINX flow
4. [SDUI views](docs/sdui.md) — server-driven UI rendering
5. [API reference](docs/api-reference.md) — full Core and SDUI endpoint reference
6. [Authentication](docs/authentication.md) — JWT, roles, tenants, public access
7. [Error codes](docs/errors.md) — verified error codes and plugin failure patterns
8. [Internationalization](docs/i18n.md) — locale files, label templates, translations
9. [YAML app development](docs/app-development.md) — app folder shape, local demo, package conventions
10. [Core-only usage](docs/core-only.md) — local runtime setup, tenant config, API smoke tests
11. [Packaging for Platform](docs/packaging.md) — ZIP shape, upload checklist, validator expectations
12. [Optional plugins](docs/plugins.md) — effect mapping, HTTP contract, security checklist
13. [Reference app patterns](docs/reference-patterns.md) — reusable patterns from the example app set
14. [AI assistant prompt](docs/ai-prompt.md) — prompt template for generating YAML app packages
15. [Troubleshooting](docs/troubleshooting.md) — compose, auth, public query, plugin, and upload fixes
