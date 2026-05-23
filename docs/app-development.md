# YAML App Development

An UP2YOU app is a folder of YAML files. The same model can run in this starter or be packaged for a hosted deployment.

For the rules and boundaries of the YAML format, read [YAML contract](yaml-contract.md).

## Minimal Structure

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
```

## Main Files

- `app.yaml` defines app key, name, version, description, and `main_view`.
- `auth.yaml` defines app-level permissions.
- `entities/*.yaml` define persistent models.
- `workflows/*.yaml` define transitions, guards, mutations, and effects.
- `queries/*.yaml` define reusable read patterns and public queries.
- `views/*.yaml` define server-driven screens.
- `forms/*.yaml` define create/edit forms.
- `locales/*.json` define labels for app, fields, states, transitions, and validation messages.

See [SDUI views](sdui.md) for how `views/*.yaml` are exposed through SDUI Gateway and rendered by a client.

## Local Demo vs Package Example

`tenants/00000000-0000-0000-0000-000000000001/my-todo` is the default local demo app. It is plugin-free so the first run works without any external plugin service.

`examples/my-todo` is the fuller package source and includes the optional `todo-logger` plugin example.

## Tiny Demo UI

The `web/` folder contains a plain HTML/CSS/JS screen for `my-todo`. It calls Core APIs directly and intentionally does not use SDUI. Use it as a small browser smoke test for CRUD, transitions, and public queries.

For production UI experiments, use the SDUI Gateway or your own frontend. The starter keeps this demo small so the YAML app contract remains the main focus.

## Conventions

- App keys use lowercase words with hyphens, for example `my-todo`.
- Default main view convention is PascalCase app key, for example `MyTodo`.
- Set `main_view` in `app.yaml` when the main view has a custom name.
- Keep field names stable once data exists.
- Mark public queries or views with `public: true` only when their data is safe to expose without login.
