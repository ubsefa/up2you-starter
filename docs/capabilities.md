# Capabilities and Boundaries

This starter runs UP2YOU **core-only**. The runtime engine is identical to a hosted deployment — the difference is the product layer around it. Use this page to know what your app can rely on here versus on a hosted platform.

## Core-only (this starter)

Available:

- Entities, workflows/transitions, queries, effects, and SDUI views/forms.
- State machine + event sourcing, access scope (`owner`/`rank`), read models.
- Named queries and **public read** queries/views.
- Local plugin services (you run them) and the scheduled-effect contract (you run the runner).
- `make validate` against the same core image that runs in production.

Not included here (these are product/hosted concerns):

- Accounts, login/session, email verification. `AUTH_ENABLED=false` by default; if you enable auth you provide the JWTs yourself.
- Marketplace browse/install, app review and publishing.
- Billing: paid apps, payments, developer payouts.
- Member management and `app_roles` resolution from a membership model.
- Quota enforcement and package upload/installer.

## Hosted platform

Adds, on top of the same runtime:

- Accounts and sessions; JWTs carrying `app_roles`/`allowed_tenants` from a license/membership model.
- Marketplace: browse, install, uninstall; app review (submit → approve/reject) and publishing.
- Billing: free/paid apps, payments, developer payouts and commission.
- Per-tenant member management and quota enforcement.
- Package upload with package/marketplace validation (naming conventions, ZIP shape, view/form references) in addition to the core runtime validation.

## Public vs private

- **Public read**: a query with `public: true` is served at `GET /api/v1/_public/{query}` without authentication; a view with `public: true` is served by the SDUI gateway at `/ui/public/views/{name}` (hosted renderers expose it at `/p/{tenantID}/{viewName}`). Both still need tenant context (`X-Tenant-ID`).
- **Read only**: the public surface is read-only. There is no public create/update/delete — creating data, updating, and transitions always require an authenticated request.
- Expose only safe, non-sensitive fields through public queries/views; project with `read_model.fields`.

## Core-only caveats

- **AppMemberSelect** populates its options from a Platform member list. In core-only there is no member source, so it cannot resolve members on its own — use `EntitySelect` against your own entity for member-like pickers here.
- **Scheduled effects**: the schedule contract is supported, but running them on an interval is a hosted/runner concern.
- **`app_roles`**: absent unless you mint JWTs yourself; without them Core falls back to the JWT `role` claim (see [authentication.md](authentication.md)).

See also: [architecture.md](architecture.md) for deployment differences, [authentication.md](authentication.md) for the JWT and app-role model, and [packaging.md](packaging.md) for validation and upload.
