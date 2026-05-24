# Troubleshooting

## Compose Config

Required tools: Docker and Docker Compose v2.

```bash
make setup
```

## Service Health

```bash
docker compose ps
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/_schema \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Demo screen:

```text
http://localhost:8080/demo/
```

## Reset Local State

This removes database, read database, NATS, and upload volumes:

```bash
docker compose down -v
docker compose up -d
```

## Public Query Returns Forbidden

Check that the query has:

```yaml
public: true
```

Public query HTTP requests also need tenant context:

```bash
curl http://localhost:8080/api/v1/_public/public_open_tasks \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

Public SSE uses `tenant_id` in the query string:

```bash
curl -N "http://localhost:8080/api/v1/_public/public_open_tasks/stream?tenant_id=00000000-0000-0000-0000-000000000001"
```

## Plugin Calls Fail

The default tenant app is plugin-free. If you use `examples/my-todo` with the logger plugin enabled, make sure the plugin service is reachable at the endpoint declared in `app.yaml`.

---

## Upload Validate Fails

When `POST /installer/upload` returns an error:

1. **Read the error message**: It usually names the failing file and field.
2. **Check ZIP structure**: `app.yaml` must be at the ZIP root, not inside a parent folder.
3. **Validate YAML**: Run `python -c "import yaml; yaml.safe_load(open('app.yaml'))"` on each file.
4. **Check cross-references**: Entity names in workflows must match `entity.name` in `entities/*.yaml`.
5. **Check state machine**: `entity.initial_state` and all workflow `from`/`to` states must be in `entity.states`.
6. **Check auth roles**: Every role in workflow `permissions` must exist in the target tenant/platform role model.

Common fix:

```bash
# Inspect ZIP contents
unzip -l dist/my-todo.zip

# Validate YAML files
for f in app.yaml auth.yaml entities/*.yaml workflows/*.yaml queries/*.yaml views/*.yaml forms/*.yaml; do
  python -c "import yaml; yaml.safe_load(open('$f')); print('$f OK')"
done
```

## Transition Errors

`TRANSITION_ERROR` (409) means the transition is invalid for the current state.

### Checklist

1. **Check current state**: `GET /api/v1/Task/{id}` to see `current_state`.
2. **Check available transitions**: `GET /api/v1/Task/{id}/transitions` to see what is available.
3. **Check workflow definition**: Verify the transition `from` includes the current state.
4. **Check guard expressions**: Guard expressions must evaluate to `true`. Check that the user matches the expected condition (e.g., `event.user_id == state.assignee_user_id`).
5. **Check payload**: Some transitions require a `payload` with extra fields.

Example debug flow:

```bash
# Check current state
curl http://localhost:8080/api/v1/Task/<id> \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Check available transitions
curl http://localhost:8080/api/v1/Task/<id>/transitions \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Execute transition
curl -X POST http://localhost:8080/api/v1/Task/<id>/transitions/complete \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{}'
```

If the error message mentions a guard expression failure, the guard condition evaluated to `false`. Check that the entity data satisfies the guard.

## Plugin Timeout

When a plugin effect takes longer than the declared `timeout` in `app.yaml`:

1. **Check the plugin service is running**: `curl <plugin-endpoint>/health`.
2. **Check plugin logs**: The plugin should log incoming `/execute` requests and response times.
3. **Check network**: Plugin Host must be able to reach the plugin endpoint. In Docker, use the service name, not `localhost`.
4. **Increase timeout if justified**: Only increase the timeout if the plugin has a clear reason for slow execution (e.g., external API call with SLA). Do not use timeout as a workaround for plugin bugs.

In `app.yaml`:

```yaml
plugins:
  - name: todo-logger
    endpoint: http://todo-logger:8080
    timeout: 10s
```

Default timeout is 5 seconds.

### Plugin External Request Blocked

If a plugin can receive `/execute` requests but cannot call an external service:

1. **Check `plugin.yaml`**: The external hostname should be listed under `plugin.egress.hosts` when deploying to a hosted platform.
2. **Check wildcard scope**: `*.example.com` matches `api.example.com`, but not `example.com` or `a.b.example.com`.
3. **Check the target address**: Private, internal, loopback, and metadata addresses are not valid hosted outbound targets.
4. **Check proxy support**: Hosted deployments may provide `HTTP_PROXY` and `HTTPS_PROXY`. Use an HTTP client or runtime configuration that respects those variables.
5. **Avoid raw sockets for integrations**: Direct outbound sockets may be blocked even when standard HTTP(S) through the proxy is allowed.

Example plugin manifest:

```yaml
plugin:
  name: approval-notifier
  egress:
    hosts:
      - hooks.slack.com
```

### Plugin Not Registered

If the plugin is not registered when an effect triggers:

1. **Check `app.yaml`**: The plugin `name` and `effects` must be declared.
2. **Check Plugin Host logs**: The plugin should have registered on startup.
3. **Check effect name**: The effect name in `effects/*.yaml` must match an effect the plugin declared during registration.

## Auth Enabled Issues

When `AUTH_ENABLED=true` and requests start failing:

1. **Check `Authorization` header**: Must be `Bearer <token>`, not just the raw token.
2. **Check token claims**: `user_id`, `tenant_id`, `role` must be present.
3. **Check tenant match**: Token `tenant_id` must match `X-Tenant-ID` header, or the token must have cross-tenant access (`allowed_tenants` or `admin` role).
4. **Check app_roles**: If the JWT has `app_roles`, the target tenant must be a key in that map. Missing `app_roles[tenant_id]` returns `APP_ROLE_REQUIRED`.
5. **Check role declaration**: The JWT role must be listed in `auth.yaml` roles.
6. **Check permissions**: The effective role must be in the permission list for the action.

Debug with `AUTH_ENABLED=false`:

```bash
# Test with auth disabled
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"

# Re-enable auth and test with token
curl http://localhost:8080/api/v1/Task \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Authorization: Bearer <token>"
```

If the first request works and the second fails, the issue is with the token or permissions.

### Token Generation

In Core-only mode, you must generate your own JWT. See [Authentication](authentication.md) for token generation examples.

## Schema Not Loading

If `GET /api/v1/_schema` returns empty or stale data:

1. **Reload config**: `POST /api/v1/_admin/reload` (requires `admin` role).
2. **Check YAML files**: Invalid YAML or missing required fields prevent config loading.
3. **Check server logs**: The Core Engine logs config loading errors.
4. **Check tenant directory**: `CONFIG_DIR` must point to a directory with valid app folders.

## Database Errors

If `GET /health` shows `"database": "error"`:

1. **Check PostgreSQL is running**: `docker compose ps` should show `db` as `healthy`.
2. **Check database connection**: `DATABASE_URL` must point to the correct host and credentials.
3. **Check migrations**: The database schema must match the Core Engine version.
4. **Reset state**: `docker compose down -v && docker compose up -d` creates a fresh database.
