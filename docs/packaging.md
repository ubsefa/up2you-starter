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
