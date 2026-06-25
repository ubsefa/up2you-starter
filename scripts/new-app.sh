#!/usr/bin/env bash
set -euo pipefail

KEY="${1:-}"
PARENT="${2:-.}"

if [ -z "$KEY" ]; then
  echo "usage: new-app.sh <app-key> [parent-dir]" >&2
  echo "  app-key: lowercase letters, digits and hyphens, e.g. my-todo" >&2
  exit 2
fi

if ! printf '%s' "$KEY" | grep -qE '^[a-z][a-z0-9-]*$'; then
  echo "invalid app-key '$KEY': use lowercase letters, digits and hyphens (start with a letter)" >&2
  exit 2
fi

PASCAL="$(printf '%s' "$KEY" | awk -F'[-_ ]' '{o="";for(i=1;i<=NF;i++)o=o toupper(substr($i,1,1)) substr($i,2);print o}')"
TITLE="$(printf '%s' "$KEY" | awk -F'[-_ ]' '{o="";for(i=1;i<=NF;i++)o=o (i>1?" ":"") toupper(substr($i,1,1)) substr($i,2);print o}')"
SNAKE="$(printf '%s' "$KEY" | tr '-' '_' | tr ' ' '_')"

DIR="$PARENT/$KEY"
if [ -e "$DIR" ]; then
  echo "refusing to overwrite existing path: $DIR" >&2
  exit 1
fi

mkdir -p "$DIR/entities" "$DIR/workflows" "$DIR/queries" "$DIR/forms" "$DIR/views" "$DIR/locales"

cat > "$DIR/app.yaml" <<YAML
app:
  key: $KEY
  name: $TITLE
  version: "0.1.0"
  description: $TITLE app
  main_view: $PASCAL
YAML

cat > "$DIR/auth.yaml" <<YAML
auth:
  provider: jwt
  roles:
    - admin
    - user
  permissions:
    Item.read: [admin, user]
    Item.create: [admin, user]
    Item.update: [admin, user]
    Item.delete: [admin]
YAML

cat > "$DIR/entities/item.yaml" <<YAML
entity:
  name: Item
  initial_state: active
  soft_delete: true
  fields:
    title:
      type: string
      required: true
    notes:
      type: string
  states:
    active: {}
    archived: {}
YAML

cat > "$DIR/workflows/item_flow.yaml" <<YAML
workflow:
  name: item_flow
  entity: ref:entities/Item
  transitions:
    archive:
      from: [active]
      to: archived
      permissions: [admin, user]
    restore:
      from: [archived]
      to: active
      permissions: [admin, user]
YAML

cat > "$DIR/queries/${SNAKE}_queries.yaml" <<YAML
queries:
  ${SNAKE}_all:
    entity: ref:entities/Item
    sort:
      - field: created_at
        order: desc
    limit: 50
YAML

cat > "$DIR/forms/create_item.yaml" <<YAML
form:
  name: ${PASCAL}Create
  entity: Item
  fields:
    - name: title
      type: TextInput
      props:
        required: true
    - name: notes
      type: TextArea
YAML

cat > "$DIR/views/main.yaml" <<YAML
view:
  name: $PASCAL
  layout: Stack
  data_source: ref:queries/${SNAKE}_all
  components:
    - type: DataTable
      props:
        entity: Item
        data_source: ref:queries/${SNAKE}_all
        create_form: ${PASCAL}Create
        columns:
          - field: title
          - field: current_state
        row_actions:
          - type: edit
          - type: transition
            entity: Item
            transition: archive
          - type: transition
            entity: Item
            transition: restore
YAML

cat > "$DIR/locales/en.json" <<JSON
{
  "app": { "name": "$TITLE" },
  "fields": { "title": "Title", "notes": "Notes" },
  "states": { "active": "Active", "archived": "Archived" },
  "transitions": { "archive": "Archive", "restore": "Restore" }
}
JSON

echo "created $DIR"
echo "validate it with: ./scripts/validate-examples.sh $PARENT"
