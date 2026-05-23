# API Reference

This document lists all UP2YOU Core Engine endpoints.

These endpoints are available when running Core locally or when calling a deployment's Core runtime layer. Product-layer APIs are documented separately by the deployment that provides them.

## Base URL

Local starter:

```text
http://localhost:8080
```

In hosted deployments, Core usually runs behind a reverse proxy. Use the deployment's public routing documentation for external URLs.

## Authentication

Most runtime API endpoints require tenant context. For normal HTTP calls, pass `X-Tenant-ID` with a valid tenant UUID:

```bash
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

When `AUTH_ENABLED=true`, most endpoints also require a valid JWT via `Authorization: Bearer <token>`. Read [Authentication](authentication.md) for the full model.

Public queries (`/_public/`) skip auth but still require tenant context via `X-Tenant-ID` or `tenant_id` query parameter. Health, readiness, and root status endpoints do not require tenant context.

---

## Health and Status

### `GET /health`

Service health check. Returns component status for config, database, and NATS.

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "service": "core-engine",
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "config": "ok",
    "database": "ok",
    "nats": "ok"
  }
}
```

### `GET /ready`

Readiness probe.

```bash
curl http://localhost:8080/ready
```

Response:

```json
{ "status": "ready" }
```

### `GET /`

Engine info and loaded apps.

```bash
curl http://localhost:8080/
```

---

## Schema and Configuration

### `GET /api/v1/_schema`

Returns the full app schema for the tenant: entities, fields, states, workflows, queries, views, and permissions.

```bash
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### `GET /api/v1/_locales/{lang}`

Returns locale strings for the given language code (e.g., `en`, `tr`).

```bash
curl http://localhost:8080/api/v1/_locales/en \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### `POST /api/v1/_admin/reload`

Reloads tenant configuration. Requires admin role when auth is enabled.

```bash
curl -X POST http://localhost:8080/api/v1/_admin/reload \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

---

## Entity CRUD

All entity endpoints follow the pattern `/api/v1/{entity}` where `{entity}` is the entity route name exposed by the `_schema` endpoint.

### List entities

```
GET /api/v1/{entity}
```

Returns a paginated list of entity records.

Query parameters:

| Param | Description |
| --- | --- |
| `_q` | Full-text search across indexed/readable fields |
| `_sort` | Sort field; prefix with `-` for descending (e.g., `-created_at`) |
| `_limit` | Result limit |
| `_cursor` | Cursor for next page |

```bash
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# With search and sort
curl 'http://localhost:8080/api/v1/Task?_q=urgent&_sort=-created_at&_limit=20' \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Create entity

```
POST /api/v1/{entity}
```

```bash
curl -X POST http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"title":"Try UP2YOU","priority":"medium"}'
```

### Get entity by ID

```
GET /api/v1/{entity}/{id}
```

```bash
curl http://localhost:8080/api/v1/Task/<task-id> \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Update entity

```
PUT /api/v1/{entity}/{id}
PATCH /api/v1/{entity}/{id}
```

PUT replaces the record; PATCH merges fields.

```bash
curl -X PUT http://localhost:8080/api/v1/Task/<task-id> \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","priority":"high"}'
```

### Delete entity

```
DELETE /api/v1/{entity}/{id}
```

Soft-deletes the record if `soft_delete: true` is set on the entity.

```bash
curl -X DELETE http://localhost:8080/api/v1/Task/<task-id> \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

---

## Entity History and Replay

### Get entity events

```
GET /api/v1/{entity}/{id}/events
```

Returns the event history for a specific entity record.

```bash
curl http://localhost:8080/api/v1/Task/<task-id>/events \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Replay entity

```
GET /api/v1/{entity}/{id}/replay
```

Rebuilds the entity state from its event history. Useful for debugging or recovering from inconsistent state.

```bash
curl http://localhost:8080/api/v1/Task/<task-id>/replay \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

---

## Workflow Transitions

### List available transitions

```
GET /api/v1/{entity}/{id}/transitions
```

Returns the transitions available for the current entity state, filtered by user permissions.

```bash
curl http://localhost:8080/api/v1/Task/<task-id>/transitions \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Execute transition

```
POST /api/v1/{entity}/{id}/transitions/{transition}
```

Executes a workflow transition. Body can include `payload` fields if the transition requires extra input.

```bash
curl -X POST http://localhost:8080/api/v1/Task/<task-id>/transitions/complete \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{}'
```

With payload:

```bash
curl -X POST http://localhost:8080/api/v1/Task/<task-id>/transitions/archive \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"payload":{"reason":"Out of scope"}}'
```

---

## Named Queries

### Execute named query

```
GET /api/v1/_query/{query}
POST /api/v1/_query/{query}
```

GET for simple queries; POST for queries that accept a body (e.g., filter overrides).

```bash
curl http://localhost:8080/api/v1/_query/my_todo_all \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Export as CSV
curl 'http://localhost:8080/api/v1/_query/my_todo_all/_export?format=csv' \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Execute public query

```
GET /api/v1/_public/{query}
```

Public queries skip auth but require `X-Tenant-ID`. The query must have `public: true` in its YAML definition.

```bash
curl http://localhost:8080/api/v1/_public/public_open_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Read [Authentication](authentication.md) for public access rules.

---

## Server-Sent Events (SSE)

### Authenticated event stream

```
GET /api/v1/events/stream
```

Streams real-time events for the tenant. Requires a valid JWT.

```bash
curl -N http://localhost:8080/api/v1/events/stream \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Public query stream

```
GET /api/v1/_public/{query}/stream?tenant_id=<tenant-id>
```

Streams query updates without auth. The query must have `public: true`.

```bash
curl -N 'http://localhost:8080/api/v1/_public/public_open_tasks/stream?tenant_id=00000000-0000-0000-0000-000000000001'
```

JavaScript example:

```javascript
const source = new EventSource(
  '/api/v1/_public/public_open_tasks/stream?tenant_id=<tenant-id>'
);
source.addEventListener('query_update', (event) => {
  const payload = JSON.parse(event.data);
  // payload.items has the same shape as GET /api/v1/_public/public_open_tasks
  renderItems(payload.items);
});
```

---

## File Upload and Download

### Upload

```
POST /api/v1/_upload
```

Uploads a file. Requires permission for the target entity and field.

```bash
curl -X POST http://localhost:8080/api/v1/_upload \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -F "tenant=00000000-0000-0000-0000-000000000001" \
  -F "app=my-todo" \
  -F "entity=Task" \
  -F "entity_id=<task-id>" \
  -F "field=avatar" \
  -F "file=@photo.jpg"
```

### Serve private file

```
GET /api/v1/_files/{tenant}/{app}/{entity}/{field}/{filename}
```

Serves an uploaded file with tenant-scoped access control.

```bash
curl http://localhost:8080/api/v1/_files/00000000-0000-0000-0000-000000000001/my-todo/Task/<task-id>/avatar/photo.jpg \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Direct file URL (from upload response)

```
GET /uploads/{tenant}/{app}/{entity}/{field}/{filename}
```

```bash
curl http://localhost:8080/uploads/00000000-0000-0000-0000-000000000001/my-todo/Task/<task-id>/avatar/photo.jpg
```

---

## Import and Export

### Export entity data

```
GET /api/v1/{entity}/_export
```

```bash
curl 'http://localhost:8080/api/v1/Task/_export?format=csv' \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Import entity data

```
POST /api/v1/{entity}/_import
```

```bash
curl -X POST http://localhost:8080/api/v1/Task/_import \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '[{"title":"Imported Task 1","priority":"low"},{"title":"Imported Task 2","priority":"high"}]'
```

---

## Error Response Shape

All errors return a consistent JSON envelope:

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "no permission for this action"
  }
}
```

Common error codes:

| Code | HTTP Status | Meaning |
| --- | --- | --- |
| `FORBIDDEN` | 403 | Permission denied |
| `VALIDATION_ERROR` | 400 | Invalid input data |
| `TRANSITION_ERROR` | 409 | Invalid state transition |
| `APP_ROLE_REQUIRED` | 403 | JWT missing app role claim |
| `TENANT_FORBIDDEN` | 403 | Token cannot access this tenant |
| `TENANT_CONFIG_ERROR` | 503 | Tenant configuration is invalid |
| `QUERY_ERROR` | 500 | Query execution failed |
| `INVALID_TOKEN` | 401 | JWT is invalid or expired |
| `UNAUTHORIZED` | 401 | Authorization header missing |
| `MISSING_TENANT` | 400 | X-Tenant-ID header required |
| `INVALID_PARAMS` | 400 | Invalid request parameters |
| `VIEW_NOT_FOUND` | 404 | SDUI view does not exist |
| `FORM_NOT_FOUND` | 404 | SDUI form does not exist |
| `NOT_FOUND` | 404 | Endpoint not found |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not allowed |

Read [Errors](errors.md) for detailed explanations and troubleshooting.

---

## SDUI Gateway Endpoints

SDUI Gateway runs on a separate internal port (default 8090) and is proxied through NGINX at `/ui/`.

### Get view schema

```
GET /ui/views/{viewName}
```

```bash
curl http://localhost:8080/ui/views/MyTodoTasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

### Get form schema

```
GET /ui/forms/{formName}
```

```bash
curl http://localhost:8080/ui/forms/MyTodoCreate \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Read [SDUI views](sdui.md) for the view schema model and component reference.

---

## Middleware Chain

Requests pass through the following middleware in order:

1. **Logging** — request/response logging
2. **CORS** — cross-origin request handling
3. **Max Body Size** — request body limit enforcement
4. **Auth** — JWT validation and permission checks
5. **Rate Limit** — per-tenant rate limiting
6. **Idempotency** — idempotency key deduplication for POST/PUT/PATCH/DELETE
7. **Import/Export interceptor** — handles `_import`/`_export` on entity routes
8. **Router** — dispatches to the appropriate handler

---

## Product Layer vs Core Runtime

This reference covers **Core Engine** endpoints only. Account, publishing, workspace, billing, or other product-layer APIs are outside the starter contract and should be documented by the deployment that provides them.
