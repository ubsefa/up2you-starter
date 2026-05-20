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

## Views And Forms

Views describe screens for SDUI renderers. Forms describe create/edit inputs.

Keep views small and predictable. For the first version of an app, prefer list, table, detail, form, and stats patterns before custom UI.

## Effects And Plugins

Effects describe side effects. Plugins are optional services that can handle custom behavior outside Core.

Use plugins for integrations, notifications, sync jobs, or calculations that should not live in generic YAML.

The default local demo app is plugin-free. The package example under `examples/my-todo` includes an optional plugin example.

## Local Starter Boundaries

This starter is for:

- Running public Core runtime images locally.
- Testing YAML app behavior.
- Preparing app packages for hosted Platform upload.
- Giving AI agents enough rules to generate valid YAML app folders.

This starter is not:

- The hosted Platform source code.
- A production installer.
- A place for production secrets.
- A replacement for app review on the hosted Platform.

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
- Public queries expose only data intended for anonymous users.
- Locale keys exist for important labels, states, transitions, and validation messages.
- Optional plugins have clear manifests and endpoints.
