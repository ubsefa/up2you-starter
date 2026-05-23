# Error Codes

This document lists error codes returned by UP2YOU Core Engine and SDUI Gateway.

All errors return a consistent JSON envelope:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description"
  }
}
```

---

## Authentication Errors

These errors come from the auth middleware and SDUI Gateway auth layer.

### `UNAUTHORIZED`

- **HTTP Status**: 401
- **Source**: Core Engine, SDUI Gateway
- **Cause**: No `Authorization` header on a protected endpoint.
- **Fix**: Add `Authorization: Bearer <token>`.

### `INVALID_TOKEN`

- **HTTP Status**: 401
- **Source**: Core Engine, SDUI Gateway
- **Cause**: Token is expired, has wrong issuer, wrong signing method, missing required claims, or invalid role.
- **Fix**: Regenerate the JWT with correct parameters. Check `iss` is `up2you`, signing method is `HS256`, and `user_id` is present.

### `TENANT_FORBIDDEN`

- **HTTP Status**: 403
- **Source**: Core Engine, SDUI Gateway
- **Cause**: Token's tenant ID does not match `X-Tenant-ID` header, and the token lacks cross-tenant access (no `allowed_tenants` entry, not `admin` or `system` role).
- **Fix**: Either remove the `X-Tenant-ID` header (use token's tenant), or add the tenant to `allowed_tenants`, or use an admin/system token.

---

## Tenant and Configuration Errors

### `MISSING_TENANT`

- **HTTP Status**: 400
- **Source**: Core Engine, SDUI Gateway
- **Cause**: No `X-Tenant-ID` header or `tenant_id` query parameter on a public route. Also returned for file upload paths without a valid tenant.
- **Fix**: Add `X-Tenant-ID: <uuid>` header. For SSE streams, use `tenant_id=<uuid>` query parameter.

### `INVALID_PARAMS`

- **HTTP Status**: 400
- **Source**: Core Engine, SDUI Gateway
- **Cause**: Tenant ID is not a valid UUID, or request parameters are malformed.
- **Fix**: Ensure `X-Tenant-ID` is a valid UUID (e.g., `00000000-0000-0000-0000-000000000001`).

### `INVALID_PATH`

- **HTTP Status**: 400
- **Source**: SDUI Gateway
- **Cause**: File upload path does not contain a valid UUID tenant.
- **Fix**: Use correct path format: `/uploads/{tenant}/{app}/{entity}/{field}/{filename}`.

### `TENANT_CONFIG_ERROR`

- **HTTP Status**: 503
- **Source**: Core Engine
- **Cause**: Tenant configuration is invalid or cannot be loaded. This is a server-side error; the tenant YAML files may have syntax errors or missing required fields.
- **Fix**: Check server logs for the specific configuration error. Validate your `app.yaml`, `entities/`, `workflows/`, etc.

---

## Permission Errors

### `FORBIDDEN`

- **HTTP Status**: 403
- **Source**: Core Engine, SDUI Gateway
- **Cause**: Effective role is not in the permission list for the requested action.
- **Fix**: Check `auth.yaml` permissions. Ensure the user's effective role (via `app_roles` or JWT role) is listed for the action.

Example:

```yaml
auth:
  permissions:
    Task.create: [admin, user]
```

If the user's effective role is `viewer`, they will get `FORBIDDEN` on `Task.create`.

### `APP_ROLE_REQUIRED`

- **HTTP Status**: 403
- **Source**: Core Engine
- **Cause**: JWT has `app_roles` claim but the target tenant is not present in the `app_roles` map.
- **Fix**: Add the tenant to `app_roles`:

```json
{
  "app_roles": {
    "00000000-0000-0000-0000-000000000001": "admin"
  }
}
```

---

## Validation and Data Errors

### `VALIDATION_ERROR`

- **HTTP Status**: 400
- **Source**: Core Engine
- **Cause**: Input data fails entity field validation. Required fields missing, wrong field type, or value constraints violated.
- **Fix**: Check the error message for the specific field issue. Ensure required fields are present and values match the entity schema.

Example:

```bash
# Response
{ "code": "VALIDATION_ERROR", "message": "field 'title' is required" }
```

---

## Workflow Errors

### `TRANSITION_ERROR`

- **HTTP Status**: 409
- **Source**: Core Engine
- **Cause**: The requested transition is invalid for the current entity state, or a guard expression failed, or a mutation failed.
- **Fix**: Check the entity's current state and the workflow definition. Use `GET /api/v1/{entity}/{id}/transitions` to see available transitions.

Example scenarios:

- Transitioning from `open` to `done` when only `open → in_progress` is allowed.
- Guard expression `state.assignee_user_id == event.user_id` fails.
- Required payload field not provided.

---

## Query Errors

### `QUERY_ERROR`

- **HTTP Status**: 500
- **Source**: Core Engine, SDUI Gateway
- **Cause**: Query execution failed. This is a server-side error; the query YAML may reference a non-existent entity, or a database error occurred.
- **Fix**: Check server logs for the specific query failure. Verify the query's `entity` reference and filter expressions.

---

## SDUI Gateway Errors

### `VIEW_NOT_FOUND`

- **HTTP Status**: 404
- **Source**: SDUI Gateway
- **Cause**: The requested view name does not exist in the app configuration.
- **Fix**: Check the view name spelling. Verify the view file exists in `views/` and is loaded by the tenant config.

### `FORM_NOT_FOUND`

- **HTTP Status**: 404
- **Source**: SDUI Gateway
- **Cause**: The requested form name does not exist in the app configuration.
- **Fix**: Check the form name spelling. Verify the form file exists in `forms/` and is loaded by the tenant config.

### `PROXY_ERROR`

- **HTTP Status**: 502
- **Source**: SDUI Gateway
- **Cause**: SDUI Gateway failed to proxy a request to the Core Engine.
- **Fix**: Check that Core Engine is running and reachable. Check network configuration.

---

## Routing Errors

### `NOT_FOUND`

- **HTTP Status**: 404
- **Source**: Core Engine
- **Cause**: The requested path does not match any registered endpoint.
- **Fix**: Check the URL path. See [API Reference](api-reference.md) for the full endpoint list.

### `METHOD_NOT_ALLOWED`

- **HTTP Status**: 405
- **Source**: Core Engine
- **Cause**: The HTTP method is not supported for the endpoint. For example, POST to `GET /api/v1/Task`.
- **Fix**: Use the correct HTTP method for the endpoint.

---

## Plugin Failure Patterns

Plugin-related failures originate from the **plugin-host** service, not Core Engine. When a workflow triggers an effect, plugin-host calls the plugin's HTTP endpoint. These are **not** formal error codes from Core, but common failure patterns to understand:

### Plugin not reachable

- **Symptom**: Effect execution fails; Core logs a plugin call failure.
- **Cause**: Plugin service is not running at the endpoint declared in `app.yaml`, or network is unreachable.
- **Fix**: Ensure the plugin service is running and the `endpoint` URL in `app.yaml` is correct and reachable from the Core Engine network.

### Unknown effect

- **Symptom**: Plugin returns an error for an unrecognized effect name.
- **Cause**: The effect name in `effects/` does not match any effect the plugin knows how to handle.
- **Fix**: Check the plugin's `/execute` handler. Ensure the effect name in `effects/*.yaml` matches a known effect in the plugin code.

### Plugin timeout

- **Symptom**: Effect execution times out.
- **Cause**: Plugin takes longer than the `timeout` value declared in `app.yaml` (default is small, e.g., 5s).
- **Fix**: Optimize the plugin response time. Only increase `timeout` if there is a clear reason.

### Plugin response error

- **Symptom**: Plugin returns `{ "success": false, "error_message": "..." }`.
- **Cause**: Plugin encountered an error during execution (invalid payload, missing data, external API failure).
- **Fix**: Check the plugin logs and the `error_message` in the response. If `should_retry` is `true`, plugin-host will retry automatically.

### Idempotency failure

- **Symptom**: Duplicate effect execution for the same `event_id`.
- **Expected behavior**: Plugins should use `event_id` as an idempotency key. The plugin should detect duplicate calls and return the same result without re-executing.
- **Fix**: Implement idempotency in the plugin using `event_id`. Store processed event IDs and skip re-execution.

Read [Plugins](plugins.md) for the plugin HTTP contract and security checklist.

---

## Error Code Quick Reference

| Code | Status | Category |
| --- | --- | --- |
| `UNAUTHORIZED` | 401 | Auth |
| `INVALID_TOKEN` | 401 | Auth |
| `TENANT_FORBIDDEN` | 403 | Auth |
| `MISSING_TENANT` | 400 | Tenant |
| `INVALID_PARAMS` | 400 | Tenant |
| `INVALID_PATH` | 400 | Tenant |
| `TENANT_CONFIG_ERROR` | 503 | Config |
| `FORBIDDEN` | 403 | Permission |
| `APP_ROLE_REQUIRED` | 403 | Permission |
| `VALIDATION_ERROR` | 400 | Validation |
| `TRANSITION_ERROR` | 409 | Workflow |
| `QUERY_ERROR` | 500 | Query |
| `VIEW_NOT_FOUND` | 404 | SDUI |
| `FORM_NOT_FOUND` | 404 | SDUI |
| `PROXY_ERROR` | 502 | SDUI |
| `NOT_FOUND` | 404 | Routing |
| `METHOD_NOT_ALLOWED` | 405 | Routing |

---

## Debugging Tips

1. **Check the HTTP status code first**: 4xx errors are client-side; 5xx errors are server-side.
2. **Read the error message**: It usually contains the specific reason.
3. **Check `GET /api/v1/_schema`**: Verify the app schema is loaded correctly.
4. **Check server logs**: For 5xx errors, the server log has the detailed stack trace.
5. **Use `POST /api/v1/_admin/reload`**: If schema is stale after config changes, reload.
6. **Test with `AUTH_ENABLED=false`**: If auth is the suspected issue, temporarily disable it to isolate.
