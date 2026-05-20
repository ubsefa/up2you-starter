# Troubleshooting

## Compose Config

Required tools: Docker and Docker Compose v2.

```bash
make setup
```

## Service Health

```bash
docker compose ps
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Demo screen:

```text
http://localhost:8080/demo/
```

## Reset Local State

This removes database, read database, NATS, and upload volumes:

```bash
docker compose down -v
docker compose up -d
```

## Public Query Returns Forbidden

Check that the query has:

```yaml
public: true
```

Public query HTTP requests also need tenant context:

```bash
curl http://localhost:8080/api/v1/_public/public_open_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Public SSE uses `tenant_id` in the query string:

```bash
curl -N "http://localhost:8080/api/v1/_public/public_open_tasks/stream?tenant_id=00000000-0000-0000-0000-000000000001"
```

## Plugin Calls Fail

The default tenant app is plugin-free. If you use `examples/my-todo` with the logger plugin enabled, make sure the plugin service is reachable at the endpoint declared in `app.yaml`.
