#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT_ID="${TENANT_ID:-00000000-0000-0000-0000-000000000001}"

echo "Checking gateway..."
curl -fsS "$BASE_URL/health" >/dev/null

echo "Checking schema..."
curl -fsS "$BASE_URL/api/v1/_schema" \
  -H "X-Tenant-ID: $TENANT_ID" >/dev/null

echo "Creating a task..."
curl -fsS -X POST "$BASE_URL/api/v1/Task" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"title":"Starter smoke test","priority":"medium"}' >/dev/null

echo "Listing tasks..."
curl -fsS "$BASE_URL/api/v1/Task" \
  -H "X-Tenant-ID: $TENANT_ID" >/dev/null

echo "Running named query..."
curl -fsS "$BASE_URL/api/v1/_query/my_todo_all" \
  -H "X-Tenant-ID: $TENANT_ID" >/dev/null

echo "Running public query..."
curl -fsS "$BASE_URL/api/v1/_public/public_open_tasks" \
  -H "X-Tenant-ID: $TENANT_ID" >/dev/null

echo "OK"
