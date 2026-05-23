# Architecture

UP2YOU is a YAML-driven application runtime. Apps are defined as folders of YAML files, and the runtime serves as a complete backend with data storage, workflow management, queries, and optional plugin extensions.

This document describes the high-level architecture. Private implementation details are omitted.

---

## Components

```
                    ┌──────────────┐
                    │   Frontend   │
                    │  (Renderer)  │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │    NGINX     │
                    │  Reverse Proxy│
                    └──┬───────┬───┘
                       │       │
               ┌───────▼─┐  ┌──▼────────┐
               │  Core   │  │ SDUI       │
               │ Engine  │  │ Gateway    │
               └───┬──┬──┘  └────┬───────┘
                   │  │           │
            ┌──────▼─┐│    ┌──────▼───────┐
            │Plugin  │ │    │  PostgreSQL  │
            │ Host   │ │    │  (Database)  │
            └───┬────┘ │    └──────────────┘
                │      │
         ┌──────▼──────▼──────┐
         │       NATS         │
         │    (JetStream)     │
         └────────────────────┘
```

### Core Engine

The main runtime. It loads YAML app definitions and exposes:

- **Entity CRUD**: Create, read, update, delete entities with validation.
- **Workflows**: State machines with transitions, guard expressions, mutations, and effects.
- **Queries**: Named queries with filters, sorting, pagination, and public variants.
- **SSE**: Server-sent event streams for real-time updates.
- **Auth**: JWT validation, permission checks, and role resolution.
- **Files**: Upload and serve files with tenant-scoped access.
- **Import/Export**: Bulk data import and query result export.

Core reads YAML configuration from a tenant directory on disk and persists entity data in PostgreSQL.

### SDUI Gateway

Serves the view and form schemas for SDUI rendering:

- **Views**: `GET /ui/views/{viewName}` returns JSON view schemas.
- **Forms**: `GET /ui/forms/{formName}` returns JSON form schemas.

SDUI Gateway acts as a proxy between the frontend and Core. It reads the same YAML configuration but serves the UI layer separately. This separation allows Core to be used without SDUI (e.g., with a custom frontend).

### Plugin Host

An optional service that executes HTTP-based plugin effects during workflow transitions:

- Plugins are external HTTP services registered via `app.yaml`.
- Each plugin declares which effect names it handles (e.g., `todo-logger.log`).
- When a workflow triggers an effect, Core publishes an event to NATS.
- Plugin Host consumes the event, calls the plugin's `/execute` endpoint, and records the result.

Plugins are stateless HTTP services. They do not connect to the database directly. All context is provided via the effect request payload.

Read [Plugins](plugins.md) for the plugin HTTP contract.

### PostgreSQL

Primary data store for entity records, event history, and idempotency keys.

Core uses PostgreSQL for:

- Entity records (with JSONB for dynamic field storage).
- Event history (append-only log for event sourcing).
- Idempotency keys (for deduplicating POST/PUT/PATCH/DELETE requests).

Uploaded files are stored on the filesystem (or external URL); file content is not stored in PostgreSQL.

Each tenant shares the same database. Tenant isolation is enforced through tenant-scoped filters and database policies.

### NATS (JetStream)

Message bus for event publishing and plugin effect dispatch:

- Entity create and transition operations publish events to NATS JetStream.
- Plugin effects are dispatched via NATS to the Plugin Host.
- JetStream provides durable storage and replay of events.
- SSE (Server-Sent Events) provides a real-time hub for clients to receive entity change notifications; SSE delivery is triggered by entity operations but is not sourced directly from NATS.

### NGINX

Reverse proxy that routes requests to the correct service:

- `/api/v1/` → Core Engine
- `/ui/` → SDUI Gateway
- `/health`, `/ready` → Core Engine
- `/uploads/` → File serving

NGINX also handles TLS termination and CORS for the hosted Platform.

---

## Request Flow

### Entity CRUD

```
Frontend ──► NGINX ──► Core Engine ──► PostgreSQL
                                     │
                                     ▼
                                  NATS (event publish)
```

1. Frontend calls `POST /api/v1/Task`.
2. Core validates the entity schema and auth permissions.
3. Core writes the record to PostgreSQL.
4. Core publishes an event to NATS.
5. Core returns the created entity to the frontend.

### Workflow Transition

```
Frontend ──► NGINX ──► Core Engine ──► PostgreSQL (state update)
                                     │
                                     ▼
                                  NATS (effect dispatch)
                                     │
                                     ▼
                               Plugin Host ──► Plugin HTTP endpoint
```

1. Frontend calls `POST /api/v1/Task/{id}/transitions/complete`.
2. Core validates the transition (state, permissions, guard expressions).
3. Core updates the entity state in PostgreSQL.
4. If the transition has effects, Core publishes an event to NATS.
5. Plugin Host consumes the event and calls the plugin.
6. Plugin returns the result; Plugin Host records it.

### SDUI View

```
Frontend ──► NGINX ──► SDUI Gateway ──► Core Engine (schema read)
                      │
                      ▼
                  Returns JSON view schema
```

1. Frontend calls `GET /ui/views/MyTodoTasks`.
2. SDUI Gateway reads the view YAML from tenant config.
3. SDUI Gateway returns the JSON view schema.
4. Frontend renderer draws the view using the schema.

### Public Query with SSE

```
Frontend ──► NGINX ──► Core Engine ──► PostgreSQL (query)
                      │
                      ▼
                  SSE stream open
                      │
                      ▼ (on entity change)
                  entity change signal ──► Core re-runs query ──► SSE update
```

1. Frontend opens `GET /api/v1/_public/public_open_tasks/stream?tenant_id=...`.
2. Core validates the query has `public: true`.
3. Core runs the query and sends initial data.
4. On relevant entity changes, Core re-runs the query and pushes updates over SSE.

---

## Auth Flow

```
Frontend ──► Authorization: Bearer <JWT>
                      │
                      ▼
               Core Auth Middleware
                      │
           ┌──────────┼──────────┐
           ▼          ▼          ▼
       Validate     Check      Resolve
       Token       Tenant    EffectiveRole
           │          │          │
           ▼          ▼          ▼
       JWT claims  X-Tenant   app_roles
       (user_id,    header    [tenant_id]
       tenant_id,
       role)
                      │
                      ▼
                Permission Check
                (auth.yaml)
```

1. Frontend sends `Authorization: Bearer <JWT>` with `X-Tenant-ID`.
2. Auth middleware validates the JWT (HS256, issuer, expiry, claims).
3. Tenant access is checked (token tenant, allowed_tenants, admin role).
4. EffectiveRole is resolved (system bypass, app_roles, JWT role).
5. Permission is checked against `auth.yaml` permissions map.
6. Request proceeds or returns `FORBIDDEN`.

Read [Authentication](authentication.md) for the full model.

---

## Configuration Loading

Core loads YAML configuration on startup or when `POST /api/v1/_admin/reload` is called:

1. Core reads the tenant directory (configured via `CONFIG_DIR`).
2. Each app folder is parsed: `app.yaml`, `auth.yaml`, `entities/`, `workflows/`, `queries/`, `views/`, `forms/`, `effects/`.
3. YAML files are validated and cross-referenced (entity references, query names, view names).
4. Schema is built and cached for `/api/v1/_schema` and SDUI Gateway.
5. Config is invalidated and rebuilt on reload.

---

## Multi-Tenant Isolation

Tenants share all runtime components but are isolated through tenant-scoped filters and database policies:

| Layer | Isolation |
| --- | --- |
| Database | Tenant UUID filters all queries; Core enforces row-level tenant context |
| Config | Each tenant has its own directory; apps do not cross tenant boundaries |
| Auth | JWT tenant checks prevent cross-tenant access unless explicitly granted |
| Files | Upload paths include tenant UUID; file serving checks tenant context |
| Events | NATS subjects are tenant-scoped; SSE streams filter by tenant |

---

## Core-only vs Hosted Platform

The runtime is the same in both modes. The difference is the **deployment context**:

| Aspect | Core-only (starter) | Hosted Platform |
| --- | --- | --- |
| Auth | Disabled by default; user provides JWT if enabled | Full auth (login, register, members, payments) |
| Config | Local tenant directory on disk | Uploaded via installer endpoint |
| Plugins | User runs plugin services | Plugins run on Platform infrastructure |
| Database | Local PostgreSQL | Shared Platform database with tenant isolation |
| SDUI | SDUI Gateway serves schemas | Platform renderer draws views |
| Events | Local NATS JetStream | Platform NATS with durable streams |

For hosted Platform usage, see the hosted Platform documentation at [up2you.app/docs](https://up2you.app/docs).
