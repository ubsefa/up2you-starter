# SDUI Views

UP2YOU apps can describe screens with YAML view files. This is the SDUI layer: the app defines what should be shown, and a client renderer decides how to draw it.

## Runtime Flow

1. App YAML defines entities, queries, forms, workflows, and views.
2. Core loads the tenant app config and exposes runtime data APIs.
3. SDUI Gateway reads the same app config and exposes view definitions.
4. A frontend renderer calls SDUI Gateway, renders the returned view schema, and uses Core APIs for data and actions.

In this starter, SDUI Gateway is available at:

```text
http://localhost:8080/ui/
```

## View Files

Views live under:

```text
tenants/00000000-0000-0000-0000-000000000001/my-todo/views/
```

The same structure is used in package source:

```text
examples/my-todo/views/
```

A typical view file describes layout, data source, columns, forms, and row actions. Row actions can map to workflow transitions such as `start`, `complete`, `reopen`, or `archive`.

## SDUI vs Tiny Demo

The `web/` folder in this starter is not an SDUI renderer. It is a plain HTML/CSS/JS browser smoke test that calls Core APIs directly.

That split is intentional:

- Use `web/` when you want the simplest possible check that Core APIs work.
- Use SDUI Gateway when you want to render YAML-defined screens.
- Use your own frontend if you want a custom product UI while still using Core APIs.

## Hosted Platform

The hosted UP2YOU Platform includes its own frontend renderer for SDUI apps. Developers normally upload YAML app packages, and the platform handles install, permissions, marketplace lifecycle, and rendering.

This starter exposes the same app shape locally without including the private hosted Platform source code.
