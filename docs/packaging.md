# Packaging for the Hosted Platform

The hosted Platform accepts app packages as ZIP files.

## Build a ZIP

Required tool: `zip`.

```bash
./scripts/package-app.sh examples/my-todo
```

Equivalent Make command:

```bash
make package
```

The output is written to:

```text
dist/my-todo.zip
```

The ZIP root must contain `app.yaml`. Do not zip the parent folder itself.

## Expected ZIP Shape

```text
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

`plugins/` is optional. Use it only when your app needs custom side effects or integrations.

## Upload Checklist

Before uploading to the hosted Platform (`POST /installer/upload`), verify:

- [ ] `app.yaml` exists at the ZIP root (not inside a parent folder).
- [ ] `app.key` is lowercase words with hyphens (e.g., `my-todo`).
- [ ] `app.version` is a semantic version string (e.g., `"1.0.0"`).
- [ ] `app.main_view` references an existing view in `views/`.
- [ ] All `entities/*.yaml` have a valid `entity.name` and at least one field.
- [ ] `entity.initial_state` is listed in `entity.states`.
- [ ] All `workflows/*.yaml` reference valid entities (`ref:entities/...`).
- [ ] All workflow transitions have valid `from` and `to` states.
- [ ] All `queries/*.yaml` reference valid entities.
- [ ] All `views/*.yaml` reference valid queries (`ref:queries/...`).
- [ ] All `forms/*.yaml` reference valid entities.
- [ ] `auth.yaml` declares entity access permissions; include roles when your package owns the role list.
- [ ] `auth.permissions` keys match `{Entity}.{operation}` format. Workflow transition permissions are defined in workflow YAML, not in `auth.permissions`.
- [ ] `locales/*.json` are valid JSON with string values.
- [ ] No symlinks, executable bits, or hidden files in the ZIP (recommended).
- [ ] ZIP size is under 10 MB (recommended).

## Validator Expectations

The hosted Platform validates package shape and semantic references. Prepare the package so these checks pass:

1. **ZIP structure**: `app.yaml` must be at the root.
2. **YAML parsing**: All `.yaml` files must be valid YAML.
3. **Cross-references**: Entity references in workflows, queries, views, and forms should resolve.
4. **State machine integrity**: Workflow transitions should reference valid states.
5. **Auth consistency**: Entity permissions should use `{Entity}.{operation}` keys; workflow transition roles belong in workflow YAML.
6. **File safety**: Do not include symlinks, generated artifacts, local secrets, or private repo files.

If validation fails, the upload returns an error with details about which file and field failed.

## Common Upload Errors

| Error | Cause | Fix |
| --- | --- | --- |
| `app.yaml not found` | ZIP contains a parent folder, not the app files at root | Run `cd my-app && zip -r ../my-app.zip .` instead of `zip -r my-app.zip my-app/` |
| `invalid app.key` | `app.key` contains uppercase or spaces | Use lowercase with hyphens: `my-todo` |
| `unknown entity reference` | Workflow or query references a non-existent entity | Check `ref:entities/Task` matches `entities/task.yaml` → `entity.name: Task` |
| `unknown view reference` | `app.main_view` references a view that doesn't exist | Check `views/` contains the referenced file and `view.name` matches |
| `unknown form reference` | View references a form that doesn't exist | Check `forms/` contains the referenced file |
| `role mismatch` | Workflow permission uses a role your target tenant/platform does not provide | Add the role to the tenant/platform role list or change the workflow permission |
| `invalid transition` | Workflow transition `from` or `to` state is not in `entity.states` | Add the state to `entity.states` or fix the transition |
| `invalid locale JSON` | `locales/*.json` is malformed | Validate with `jq . locales/en.json` |

## Local Starter vs Hosted Platform

The starter runs Core-only. It is useful for checking YAML shape and runtime behavior.

The hosted Platform adds:

- Developer accounts.
- App review.
- Marketplace listing.
- Install/uninstall lifecycle.
- Members and app roles.
- Licenses and payments.
- Platform audit.
