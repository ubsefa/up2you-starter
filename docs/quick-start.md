# Quick Start

Get UP2YOU Core running locally and create your first YAML app in under 5 minutes.

## Prerequisites

- Docker and Docker Compose v2
- `zip` (for packaging)
- `make` (optional, for convenience commands)

## Step 1: Clone and Start

```bash
git clone https://github.com/ubsefa/up2you-starter
cd up2you-starter
cp .env.example .env
docker compose up -d
```

The compose stack starts:

| Service | Port (internal) | Purpose |
| --- | --- | --- |
| NGINX | 8080 (host) | Reverse proxy + TLS |
| Core Engine | 8080 | YAML runtime, CRUD, workflows, queries |
| SDUI Gateway | 8090 | View schema API |
| Plugin Host | 9091 | Effect to HTTP plugin dispatcher |
| PostgreSQL | 5432 | Primary database |
| NATS | 4222 | JetStream event bus |

Check health:

```bash
curl http://localhost:8080/health
```

You should see:

```json
{
  "service": "core-engine",
  "status": "healthy",
  "checks": {
    "config": "ok",
    "database": "ok",
    "nats": "ok"
  }
}
```

## Step 2: Explore the Demo App

The starter includes a `my-todo` demo app under `tenants/00000000-0000-0000-0000-000000000001/my-todo`.

Open the tiny HTML demo UI:

```text
http://localhost:8080/demo/
```

Or use the API directly:

```bash
# View app schema
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Create a task
curl -X POST http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"title":"Try UP2YOU","priority":"medium"}'

# List tasks
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Run a named query
curl http://localhost:8080/api/v1/_query/my_todo_all \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

## Step 3: Run an SDUI View

SDUI Gateway serves YAML-defined screens:

```bash
curl http://localhost:8080/ui/views/MyTodoTasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

This returns a JSON view schema that an SDUI renderer can draw. The hosted Platform includes a built-in renderer; for custom frontends, you map component types like `DataTable`, `Chart`, and `TabView` to your own UI.

Read [SDUI views](sdui.md) for the full component model.

## Step 4: Create Your First App

Create a new app folder anywhere (outside the starter repo):

```text
my-first-app/
  app.yaml
  auth.yaml
  entities/
    task.yaml
  workflows/
    task_flow.yaml
  queries/
    my_queries.yaml
  views/
    tasks.yaml
  forms/
    create_task.yaml
  locales/
    en.json
```

### app.yaml

```yaml
app:
  key: my-first-app
  name: My First App
  version: "1.0.0"
  description: A simple task tracker
  main_view: Tasks
```

### entities/task.yaml

```yaml
entity:
  name: Task
  initial_state: open
  states:
    open: {}
    in_progress: {}
    done: {}
    archived: {}
  fields:
    title:
      type: string
      required: true
    priority:
      type: string
      default: medium
```

### workflows/task_flow.yaml

```yaml
workflow:
  name: task_flow
  entity: ref:entities/Task
  transitions:
    start:
      from: [open]
      to: in_progress
      permissions: [user, admin]
    complete:
      from: [in_progress]
      to: done
      permissions: [user, admin]
    archive:
      from: [open, done]
      to: archived
      permissions: [admin]
```

### queries/my_queries.yaml

```yaml
queries:
  all_tasks:
    entity: ref:entities/Task
    sort:
      - field: created_at
        order: desc
    limit: 100
```

### views/tasks.yaml

```yaml
view:
  name: Tasks
  layout: Stack
  data_source: ref:queries/all_tasks
  components:
    - type: DataTable
      props:
        entity: Task
        create_form: CreateTask
        columns:
          - field: title
          - field: priority
          - field: current_state
        row_actions:
          - type: transition
            entity: Task
            transition: start
            icon: PlayCircleOutlined
            from: [open]
          - type: transition
            entity: Task
            transition: complete
            icon: CheckCircleOutlined
            from: [in_progress]
```

### forms/create_task.yaml

```yaml
form:
  name: CreateTask
  entity: Task
  fields:
    - name: title
      type: TextInput
      props:
        required: true
    - name: priority
      type: Select
      props:
        options:
          - value: low
          - value: medium
          - value: high
```

### auth.yaml

```yaml
auth:
  roles:
    - admin
    - user
  permissions:
    Task.read: [admin, user]
    Task.create: [admin, user]
    Task.update: [admin, user]
    Task.delete: [admin]
```

Workflow transition izinleri `workflows/task_flow.yaml` içindeki `permissions` listesi ile tanımlanır; `auth.permissions` yalnızca entity action'ları (read, create, update, delete vb.) kontrol eder.

Copy this folder into the local tenant config:

```bash
cp -r my-first-app/ tenants/00000000-0000-0000-0000-000000000001/
```

Reload the config:

```bash
curl -X POST http://localhost:8080/api/v1/_admin/reload \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Now your app is live:

```bash
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

curl http://localhost:8080/api/v1/_query/all_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

## Step 5: Package for the Hosted Platform

Build a ZIP:

```bash
./scripts/package-app.sh /path/to/my-first-app
```

The output is `dist/my-first-app.zip`. Upload this to the hosted Platform.

Read [Packaging](packaging.md) for ZIP shape rules and the upload checklist.

## Next Steps

| Topic | Read |
| --- | --- |
| YAML rules and boundaries | [YAML contract](yaml-contract.md) |
| SDUI components and view shape | [SDUI views](sdui.md) |
| Optional plugins | [Plugins](plugins.md) |
| All Core endpoints | [API reference](api-reference.md) |
| Auth model and permissions | [Authentication](authentication.md) |
| Reference app patterns | [Reference patterns](reference-patterns.md) |
| Common error codes | [Errors](errors.md) |

For hosted Platform usage (marketplace, members, licenses, payments), visit the hosted docs at [up2you.app/docs](https://up2you.app/docs).
