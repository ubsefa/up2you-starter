# YAML Contract

This document defines the practical boundaries for writing UP2YOU YAML apps in this starter.

## App Shape

An app package is a directory whose root contains `app.yaml`.

```text
my-app/
  app.yaml
  auth.yaml
  entities/
  workflows/
  queries/
  views/
  forms/
  effects/
  schedules/
  locales/
  plugins/
```

Only `app.yaml` is always required. Other folders are included when the app needs that behavior.

## File Quick Reference

Use this as a compact checklist before looking at the detailed examples.

| File | Required | Main keys | Purpose |
| --- | --- | --- | --- |
| `app.yaml` | Yes | `app.key`, `app.name`, `app.version`, `app.main_view`, `plugins` | App identity, entry view, and optional plugin registration. |
| `auth.yaml` | Recommended | `auth.roles`, `auth.permissions` | App roles and entity action permissions. |
| `entities/*.yaml` | Usually | `entity.name`, `entity.initial_state`, `entity.fields`, `entity.soft_delete` | Persistent data model and state list. |
| `workflows/*.yaml` | When states change | `workflow.entity`, `workflow.transitions` | Allowed state transitions, guards, mutations, permissions, and effects. |
| `queries/*.yaml` | When listing/searching | `queries.<name>.entity`, `filter`, `sort`, `include`, `public` | Reusable read models for views, APIs, exports, and public reads. |
| `views/*.yaml` | For SDUI | `view.name`, `view.data_source`, `view.components`, `view.public` | Server-driven screens. |
| `forms/*.yaml` | For create/edit | `form.name`, `form.entity`, `form.fields` | Input forms for records and transition payloads. |
| `effects/*.yaml` | With plugins | `effects.<name>.plugin`, `payload` | Maps workflow side effects to plugin calls. |
| `schedules/*.yaml` | With scheduled plugins | `schedules.<name>.interval`, `query`, `entity`, `effect`, `payload`, `result` | Runs plugin effects periodically over query results in hosted deployments. |
| `locales/*.json` | Recommended | `app`, `entities`, `fields`, `states`, `transitions`, `views` | User-facing labels and value translations. |
| `plugins/*/plugin.yaml` | With plugins | `plugin.name`, `service`, `port`, `effects`, `egress.hosts` | Plugin runtime manifest for hosted deployment. |

## Naming Rules

- App keys use lowercase words with hyphens: `my-todo`, `public-notice-board`.
- Entity names use PascalCase: `Task`, `Notice`, `Reservation`.
- Field names use snake_case: `due_date`, `assignee_user_id`.
- Query, workflow, transition, effect, and view ids should be stable once data exists.
- Do not rename fields casually after records have been created.

## References

YAML files can reference other app resources.

```yaml
entity: ref:entities/Task
```

Keep references explicit and local to the app package. A public app package must not depend on private repository paths, local machine paths, or production secrets.

## Entities

Entities define persistent data.

```yaml
entity:
  name: Task
  initial_state: open
  soft_delete: true
  fields:
    title:
      type: string
      required: true
```

Use simple field types first. Keep custom behavior in workflows, queries, forms, views, or optional plugins.

## Workflows

Workflows define allowed state transitions.

```yaml
workflow:
  name: task_flow
  entity: ref:entities/Task
  transitions:
    complete:
      from: [in_progress]
      to: done
      permissions: [user, admin]
```

Each transition should declare:

- `from`
- `to`
- `permissions`

Use guards for business rules and validation, not for hiding broken data.

## Auth

`auth.yaml` defines app roles and permissions.

```yaml
auth:
  roles:
    - admin
    - user
  permissions:
    Task.read: [admin, user]
    Task.create: [admin, user]
```

Frontend visibility is only convenience. Core API permission checks are the real enforcement point.

## Queries

Queries define reusable reads.

```yaml
queries:
  my_open_tasks:
    entity: ref:entities/Task
    filter:
      field: current_state
      op: eq
      value: open
```

Mark a query as public only when anonymous users can safely see the data.

```yaml
public: true
```

Public HTTP requests still need tenant context, for example `X-Tenant-ID: 00000000-0000-0000-0000-000000000001`.

## Query Include

When a field references another entity (e.g., `patient_id` stores an ID), the raw ID is not useful in a table. Use `include` to fetch the referenced record.

```yaml
queries:
  active_care_plans:
    entity: ref:entities/CarePlan
    include: [patient_id]
```

The `include` field accepts:

- A single field name: `include: patient_id`
- A list: `include: [patient_id, program_key]`

Without `include`, the query only returns the raw ID. With `include`, the renderer receives the full referenced object and can format it with `label_template`.

## Views And Forms

Views describe screens for SDUI renderers. Forms describe create/edit inputs.

Keep views small and predictable. For the first version of an app, prefer list, table, detail, form, and stats patterns before custom UI.

### EntitySelect

Use `EntitySelect` when a form field should let the user choose an entity record.

```yaml
- name: patient_id
  type: EntitySelect
  props:
    options_query: all_patients
    label_template: "{{name}}"
    required: true
```

Rules:

- `options_query` must reference a named query that returns the selectable records.
- `label_template` formats each option for display.
- The selected value is the referenced entity ID (stored as a raw ID field).

### AppMemberSelect

Use `AppMemberSelect` when a form field should let the user choose a member of the current app.

```yaml
- name: assignee_user_id
  type: AppMemberSelect
  required: true
```

The member list is provided by a product-layer member source. In Core-only mode, AppMemberSelect requires a custom member list source.

### label_template

`label_template` formats values for display in table columns, select options, and detail views.

```yaml
label_template: "{{patient_id.name}} - {{current_state:states}}"
```

Rules:

- `{{field}}` inserts the raw field value.
- `{{nested.field}}` inserts a nested field value (requires `include` in the query).
- `{{field:namespace}}` translates the value through a locale namespace (e.g., `states`, `fields`).
- Multiple placeholders can be combined: `{{patient_id.name}} ({{current_state:states}})`

Common namespaces:

- `states`: Translates state values (e.g., `open` → `"Open"`).
- `fields`: Translates field values (e.g., `program_key` → `"Care Program"`).

See [i18n](i18n.md) for the full locale model.

### Public View and Query Rules

Views and queries can be marked public:

```yaml
view:
  name: PublicBoard
  public: true
```

Rules:

- Public views are accessible without authentication via `/ui/public/views/{name}`.
- Public queries (with `public: true`) are accessible via the public query endpoint.
- Forms do not have a public access contract; all form submissions require authentication.
- The underlying data source must also be accessible. For a public view, the referenced query or entity must allow public access.
- Do not mark views public if the data includes sensitive information.

## Effects And Plugins

Effects describe side effects. Plugins are optional services that can handle custom behavior outside Core.

Use plugins for integrations, notifications, sync jobs, or calculations that should not live in generic YAML.

The default local demo app is plugin-free. The package example under `examples/my-todo` includes an optional plugin example.

Plugins that call external HTTP(S) services should declare outbound hosts in their plugin manifest:

```yaml
plugin:
  name: approval-notifier
  egress:
    hosts:
      - hooks.slack.com
      - "*.webhook.example.com"
```

`egress.hosts` is a hosted-deployment contract. It lets the platform review and enforce which external hosts a plugin may contact. Wildcards are single-label only: `*.example.com` matches `api.example.com`, but not `example.com` or `a.b.example.com`. Private, internal, loopback, and metadata addresses are not allowed outbound targets.

## Scheduled Effects

Use `schedules/*.yaml` when a hosted deployment should periodically run a plugin over records returned by a query. This is for jobs such as health checks, reminders, sync checks, or polling a known external API. Do not put infinite loops, broad Core API access, or background polling inside the plugin itself.

```yaml
schedules:
  server_health_check:
    interval: 30s
    query: devops_status_panel_active_servers
    query_params:
      environment: prod
    entity: Server
    effect: devops_status_panel_check_server
    timeout: 10s
    max_concurrency: 4
    payload:
      hostname: state.hostname
      current_state: state.current_state
    result:
      patch:
        last_check_at: data.checked_at
        response_time_ms: data.response_time_ms
        reachable: data.reachable
      transition: data.transition
      transition_payload:
        reason: data.reason
```

Rules:

- `query`, `entity`, and `effect` must reference resources in the same app package.
- `interval` must be at least `10s`.
- The scheduled query should include `system` in its `permissions` list when permissions are explicit, because the platform scheduler runs with system role.
- `query_params` is optional and passes fixed query parameters to the named query.
- `max_concurrency` limits how many records from this schedule are processed at the same time. Default is `1`.
- `timeout` limits one schedule run; keep it close to the plugin's expected runtime.
- `payload` values can use `state.*` or `record.*`; they refer to the same record context.
- Plugin success responses may include `data`; `result.patch`, `result.transition`, and `result.transition_payload` read from `data.*`.
- Scheduled plugins should be idempotent by `event_id`. A platform scheduler restart can re-fire due jobs.
- The local starter documents the package contract. Actual scheduled execution is a hosted/platform deployment capability.

## Local Starter Boundaries

This starter is for:

- Running public Core runtime images locally.
- Testing YAML app behavior.
- Preparing portable app packages for upload.
- Giving AI agents enough rules to generate valid YAML app folders.

This starter is not:

- Hosted product source code.
- A production installer.
- A place for production secrets.
- A replacement for review or operational policy in a hosted deployment.

## Package Boundaries

A ZIP package should have `app.yaml` at the ZIP root.

Do not include:

- `.env`
- local secrets
- machine-specific paths
- generated `dist/` output
- private repo files

## Compatibility Checklist

Before packaging an app:

- `app.yaml` has a stable key, name, version, and `main_view`.
- Every `ref:` points to an existing app resource.
- Every transition references valid states.
- Every view data source points to an existing query or entity.
- `include` is set for `ref:` fields that need human-readable labels.
- `label_template` uses correct field paths and namespaces.
- Public queries expose only data intended for anonymous users.
- Locale keys exist for important labels, states, transitions, and validation messages.
- Optional plugins have clear manifests and endpoints.
- Optional plugins that call external HTTP(S) services declare narrow `egress.hosts`.
