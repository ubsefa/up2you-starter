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

Runtime notes:

- The platform contract is HTTP, not Go-specific. Any language is acceptable if it exposes `/health` and `/execute`.
- Use the runtime's standard HTTP client when calling external APIs. Standard clients are more likely to support `HTTP_PROXY` and `HTTPS_PROXY`.
- If the language does not automatically read proxy environment variables, configure the proxy explicitly in the plugin code or runtime flags.
- Avoid raw TCP sockets for integrations. Hosted deployments can block direct outbound sockets and require HTTP(S) through the egress proxy.

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

Common `action` values:

- `entity_transitioned`: a workflow transition triggered the effect.
- `scheduled`: the platform scheduler triggered the effect from `schedules/*.yaml`.

Success response:

```json
{ "success": true }
```

Scheduled effects can return optional `data` for the platform scheduler to map back to the entity:

```json
{
  "success": true,
  "data": {
    "checked_at": "2026-05-24T12:00:00Z",
    "reachable": true,
    "transition": "recover"
  }
}
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

Hosted deployments can provide plugin `/execute` tokens signed with a plugin-execution secret rather than the platform's main JWT secret. Plugin examples read `PLUGIN_EXECUTION_JWT_SECRET` first and fall back to `JWT_SECRET` for older/local setups.

The included `examples/my-todo/plugins/todo-logger` service demonstrates a minimal HTTP plugin with `/health`, `/execute`, bearer token verification, unknown-effect rejection, and event idempotency.

## Scheduled Plugins

For periodic work, define a `schedules/*.yaml` file and keep the plugin stateless. The platform scheduler will query Core, call the plugin with `action: "scheduled"`, then apply the plugin response through `result.patch` or `result.transition`.

Use this model for behavior such as server health checks, reminder scans, or periodic sync checks. The plugin should not open its own endless loop, scan all Core records by itself, or require broad internal platform access.

Scheduled plugin rules:

- The schedule references a query, entity, and effect from the same package.
- `interval` must be at least `10s`; `max_concurrency` defaults to `1`.
- `query_params` can pass fixed parameters to the scheduled query.
- The query should allow `system` role when it declares explicit permissions.
- The plugin handles a single payload and returns `success` plus optional `data`.
- `result.transition_payload` can pass plugin `data.*` values into a workflow transition payload.
- The plugin should treat `event_id` as an idempotency key because scheduled jobs can re-fire after scheduler restart.
- If the plugin calls external HTTP(S) services, still declare `plugin.egress.hosts`.

## Outbound HTTP(S)

If a plugin calls external HTTP(S) services in a hosted deployment, declare the allowed hosts in the plugin manifest:

```yaml
plugin:
  name: approval-notifier
  egress:
    hosts:
      - hooks.slack.com
      - "*.webhook.example.com"
```

Rules:

- `egress.hosts` is for external hostnames or public IP literals the plugin is allowed to call.
- Wildcards are allowed only as the full leftmost label. `*.example.com` matches `api.example.com`, but not `example.com` or `a.b.example.com`.
- Private, internal, loopback, and metadata addresses are not valid outbound targets, even if listed.
- Do not use broad host patterns when a concrete API host is known.

Hosted deployments can route plugin outbound traffic through an egress proxy. Plugin code should use standard HTTP clients that respect `HTTP_PROXY` and `HTTPS_PROXY`, or configure the proxy explicitly for the runtime. Direct raw socket outbound traffic may be blocked by the hosted runtime.

## Starter vs Hosted Deployments

In this starter:

- `plugin-host` is included in Compose.
- The default tenant app does not require a plugin.
- If you enable the `examples/my-todo` plugin registration locally, you are responsible for running the custom plugin service at the endpoint declared in `app.yaml`.

In hosted product deployments:

- Plugin deployment depends on operator settings.
- Product-layer review can reject unsafe plugin code or Dockerfiles.
- Product-layer review can require `plugin.egress.hosts` for plugins that call external services.
- Production secrets and platform internals are not part of the app package.
