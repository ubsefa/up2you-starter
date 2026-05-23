# Optional Plugins

Plugins are optional HTTP services for app behavior that should stay outside the generic Core runtime.

The default local demo app under `tenants/00000000-0000-0000-0000-000000000001/my-todo` is plugin-free. That keeps first run simple. The package example under `examples/my-todo` includes a small `todo-logger` plugin so developers can see the expected shape.

## When To Use A Plugin

Use a plugin for behavior such as:

- notifications
- external integrations
- custom calculations
- sync or deployment jobs
- side effects that should run after a workflow transition

Do not use a plugin for normal CRUD, field validation, query filtering, role permissions, or simple state changes. Those belong in YAML entities, workflows, auth, queries, forms, and views.

## App Package Shape

Plugin source lives inside the app package:

```text
plugins/my-plugin/
  plugin.yaml
  Dockerfile
  main.go
```

`plugins/` is optional. If an app does not need custom side effects, leave it out.

Plugins are language-agnostic HTTP services. The examples use Go by convention, but any language or runtime can be used as long as the service implements the required HTTP contract and can be packaged and deployed safely.

Choose the Dockerfile for the plugin's runtime. A Go plugin might compile a static binary, while a Node.js, Python, PHP, Ruby, Rust, or Java plugin needs its own base image, dependency install step, exposed port, and start command.

## App Registration

Register the plugin from `app.yaml`.

```yaml
plugins:
  - name: todo-logger
    type: http
    endpoint: http://my-todo-todo-logger:8201
    effects:
      - my_todo_log_task_event
    timeout: 5s
    retry: 1
```

Rules:

- `name` is unique inside the app.
- `type` is `http`.
- `endpoint` must be reachable from the runtime network.
- `effects` must match effect names declared under `effects/`.
- `timeout` and `retry` should stay small unless there is a clear reason.

## Effect Mapping

Effects connect workflow events to plugin calls.

```yaml
effects:
  my_todo_log_task_event:
    plugin: todo-logger
    payload:
      task_title: state.title
      assignee_user_id: state.assignee_user_id
```

Payload values should come from `state.*`, `event.*`, or fixed values. Do not put secrets, JWTs, database passwords, or host credentials in effect payloads.

## HTTP Contract

Each plugin should expose:

```text
GET  /health
POST /execute
```

`GET /health` should be side-effect free and return JSON.

`POST /execute` receives an effect request:

```json
{
  "effect_name": "my_todo_log_task_event",
  "action": "entity_transitioned",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "entity_id": "task-id",
  "entity_type": "Task",
  "event_id": "event-id",
  "transition": "complete",
  "payload": {
    "task_title": "Try UP2YOU"
  }
}
```

Success response:

```json
{ "success": true }
```

Error response:

```json
{
  "success": false,
  "error_message": "invalid payload",
  "should_retry": false
}
```

Use `should_retry=true` only for temporary failures. Validation errors, unknown effects, and permission failures should not retry.

## Security And Reliability

Plugins should:

- verify service-to-service Authorization tokens when enabled
- treat `event_id` as an idempotency key
- avoid writing secrets to logs
- reject unknown effect names
- keep request body limits small
- run with a non-root container user
- avoid privileged Docker access

The included `examples/my-todo/plugins/todo-logger` service demonstrates a minimal HTTP plugin with `/health`, `/execute`, bearer token verification, unknown-effect rejection, and event idempotency.

## Starter vs Hosted Deployments

In this starter:

- `plugin-host` is included in Compose.
- The default tenant app does not require a plugin.
- If you enable the `examples/my-todo` plugin registration locally, you are responsible for running the custom plugin service at the endpoint declared in `app.yaml`.

In hosted product deployments:

- Plugin deployment depends on operator settings.
- Product-layer review can reject unsafe plugin code or Dockerfiles.
- Production secrets and platform internals are not part of the app package.
