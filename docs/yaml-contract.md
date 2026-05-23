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
  locales/
  plugins/
```

Only `app.yaml` is always required. Other folders are included when the app needs that behavior.

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
