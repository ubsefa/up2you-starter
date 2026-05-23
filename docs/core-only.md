# Core-only Usage

This starter runs the UP2YOU runtime without a hosted product layer.

## Start

Required tools: Docker and Docker Compose v2. `make` is optional.

```bash
cp .env.example .env
docker compose up -d
```

Equivalent Make commands:

```bash
make setup
make up
make smoke
```

Open the demo screen:

```text
http://localhost:8080/demo/
```

Health checks:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

## Tenant Config

The default Core config is mounted from:

```text
tenants/00000000-0000-0000-0000-000000000001
```

`CONFIG_DIR=/tenants/00000000-0000-0000-0000-000000000001` points Core at that tenant. The included app is:

```text
tenants/00000000-0000-0000-0000-000000000001/my-todo
```

## Runtime API Examples

Create a task:

```bash
curl -X POST http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"title":"Try UP2YOU","priority":"medium"}'
```

List tasks:

```bash
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Run a named query:

```bash
curl http://localhost:8080/api/v1/_query/my_todo_all \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Run a public query:

```bash
curl http://localhost:8080/api/v1/_public/public_open_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Open the public query stream:

```bash
curl -N "http://localhost:8080/api/v1/_public/public_open_tasks/stream?tenant_id=00000000-0000-0000-0000-000000000001"
```

## Auth

This starter defaults to `AUTH_ENABLED=false`. If you enable auth, you must provide valid JWTs and app roles yourself. Hosted product layers can provide account/session flows separately.
