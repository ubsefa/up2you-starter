# Authentication

UP2YOU Core uses JWT tokens for authentication and a permission model based on app-level roles.

This document describes the Core auth model. Account and session APIs belong to the product layer of a deployment and are outside this starter contract.

## Core-only vs Hosted Deployments

**Core-only mode** (this starter):

- `AUTH_ENABLED=false` by default.
- No token generation is included. You are responsible for providing valid JWTs if you enable auth.
- Useful for testing YAML app behavior without the Platform layer.

**Hosted deployments**:

- A product layer can provide accounts and sessions.
- That layer is responsible for generating JWTs with the claims Core expects.
- App roles can be resolved from whatever membership model that deployment owns.

The Core runtime is the same in both modes. The difference is who produces the JWT and resolves app roles.

---

## JWT Token Model

Core validates JWT tokens using **HS256** signing.

```json
{
  "user_id": "user-uuid",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "role": "admin",
  "email": "user@example.com",
  "app_roles": {
    "00000000-0000-0000-0000-000000000001": "admin"
  },
  "allowed_tenants": ["00000000-0000-0000-0000-000000000001"],
  "iss": "up2you",
  "iat": 1705312200,
  "exp": 1705398600
}
```

### Required Claims

| Claim | Type | Description |
| --- | --- | --- |
| `user_id` | string | Unique user identifier |
| `tenant_id` | string | UUID of the token's home tenant |
| `role` | string | System role (e.g., `admin`, `user`, `system`) |
| `iss` | string | Issuer; must be `up2you` |
| `iat` | number | Issued at timestamp |
| `exp` | number | Expiration timestamp |

### Optional Claims

| Claim | Type | Description |
| --- | --- | --- |
| `email` | string | User email |
| `app_roles` | object | Tenant-scoped app role map: `{tenant_id: role}` |
| `allowed_tenants` | array | Tenants this token can access (cross-tenant) |

---

## Auth Middleware Flow

When `AUTH_ENABLED=true`:

1. Request hits any path except `/health`, `/ready`, or `/_public/`.
2. Middleware extracts `Authorization: Bearer <token>` header.
3. Token is validated: signing method, issuer, expiration, required claims.
4. Role is checked against roles declared in `auth.yaml`.
5. Tenant ID from the token is used unless overridden by `X-Tenant-ID` header (cross-tenant access requires `allowed_tenants` or `admin` role).
6. Context is populated with `user_id`, `tenant_id`, `role`, and `app_roles`.

When `AUTH_ENABLED=false`:

- `X-User-ID`, `X-Tenant-ID`, and `X-Role` headers are used directly.
- Defaults are applied from config if headers are missing.

---

## App Roles and EffectiveRole

Core does not use the JWT `role` claim directly for permission checks. Instead, it resolves an **EffectiveRole** for each request.

### Resolution Order

1. **System role bypass**: If `role == "system"`, permission checks are skipped. The system role bypasses all permission lists but is still subject to workflow guard expressions.

2. **Platform tenant**: If the request targets the platform tenant ID, the JWT role is used directly.

3. **App roles**: If the JWT contains `app_roles`, Core looks up the role for the target tenant:
   - `app_roles[tenant_id]` is the effective role.
   - If the tenant key is missing from `app_roles`, the request is rejected with `APP_ROLE_REQUIRED`.

4. **Fallback**: If `app_roles` is not present in the JWT, the JWT `role` claim is used.

```
EffectiveRole resolution:

JWT.role == "system"
  └─> return "system" (bypass permissions)

JWT.tenant_id == platform_tenant_id
  └─> return JWT.role

JWT.app_roles exists
  └─> JWT.app_roles[tenant_id] found
        └─> return app_role
      JWT.app_roles[tenant_id] missing
        └─> reject: APP_ROLE_REQUIRED

JWT.app_roles not present
  └─> return JWT.role
```

### Why App Roles Matter

Each app defines its own roles in `auth.yaml`:

```yaml
auth:
  roles:
    - admin
    - user
  permissions:
    Task.read: [admin, user]
    Task.create: [admin, user]
```

The `app_roles` claim in the JWT maps tenants to the user's role **within that app's context**:

```json
{
  "app_roles": {
    "00000000-0000-0000-0000-000000000001": "admin"
  }
}
```

This means a user can be `admin` in one app and `user` in another, even on the same tenant.

---

## Permission Checking

Core checks permissions against the `auth.yaml` permissions map:

```
checkPermission(permissions, action, effectiveRole):
  1. If effectiveRole == "system" → allow
  2. If permissions map is empty → allow (default open)
  3. If action key exists in permissions:
     - If effectiveRole is in the allowed list → allow
     - Otherwise → reject: FORBIDDEN
  4. If action key is not in permissions → reject: FORBIDDEN
```

### Action Key Format

Permission keys follow the pattern:

- `{Entity}.{operation}` for CRUD: `Task.read`, `Task.create`, `Task.update`, `Task.delete`

`auth.permissions` yalnızca entity action izinleri içindir. Workflow yetkisi transition YAML'daki `permissions` listesi ile kontrol edilir; `auth.permissions` içinde workflow key'leri tanımlanmaz.

### System Role Behavior

The `system` role bypasses permission lists. It is **not** unrestricted:

- Workflow guard expressions are still evaluated.
- Tenant validation still applies.
- The system role should only be used for administrative or migration operations.

---

## Public Access

Paths under `/api/v1/_public/` skip JWT validation but still require tenant context:

```bash
curl http://localhost:8080/api/v1/_public/public_open_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

The tenant can be passed via:

- `X-Tenant-ID` header (preferred for HTTP)
- `tenant_id` query parameter (required for SSE streams)

Public queries must have `public: true` in their YAML definition:

```yaml
queries:
  public_open_tasks:
    entity: ref:entities/Task
    public: true
    filter:
      field: current_state
      op: eq
      value: open
```

Views can also be marked public:

```yaml
view:
  name: PublicBoard
  public: true
  data_source: ref:queries/public_open_tasks
```

---

## Cross-Tenant Access

A token issued for one tenant can access another tenant if:

1. The token has `allowed_tenants` containing the target tenant, OR
2. The token role is `admin` or `system`

The `X-Tenant-ID` header overrides the token's `tenant_id`:

```bash
curl http://localhost:8080/api/v1/Task \
  -H "Authorization: Bearer <token-from-tenant-a>" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000002"
```

If neither condition is met, the request is rejected with `TENANT_FORBIDDEN`.

---

## Token Generation (Core-only Mode)

In Core-only mode, you must generate your own JWT tokens. Generate a JWT with any HS256 library. Use these parameters:

| Parameter | Value |
| --- | --- |
| Signing method | HS256 |
| Issuer (`iss`) | `up2you` |
| Secret | Value from `JWT_SECRET` env var |
| Required claims | `user_id`, `tenant_id`, `role` |

Token expiry should be set in your token's `exp` claim as needed by your application. The Core runtime validates `exp` against the current time but does not enforce a specific expiry duration through configuration.

Example with Node.js:

```javascript
const jwt = require('jsonwebtoken');

const token = jwt.sign(
  {
    user_id: 'user-1',
    tenant_id: '00000000-0000-0000-0000-000000000001',
    role: 'admin',
    app_roles: {
      '00000000-0000-0000-0000-000000000001': 'admin',
    },
    iss: 'up2you',
  },
  process.env.JWT_SECRET,
  { expiresIn: '24h' }
);
```

---

## Common Auth Errors

| Error | Cause | Fix |
| --- | --- | --- |
| `UNAUTHORIZED` | No Authorization header | Add `Authorization: Bearer <token>` |
| `INVALID_TOKEN` | Expired, wrong issuer, wrong signing method | Regenerate token with correct params |
| `APP_ROLE_REQUIRED` | JWT has no `app_roles` for target tenant | Add tenant to `app_roles` claim |
| `TENANT_FORBIDDEN` | Token cannot access the X-Tenant-ID tenant | Add tenant to `allowed_tenants` or use admin role |
| `FORBIDDEN` | Effective role not in permission list | Grant role in `auth.yaml` or update JWT `app_roles` |

Read [Errors](errors.md) for more details.
