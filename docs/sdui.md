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

## Common View Shape

Most list screens use a view-level data source and a `DataTable`.

```yaml
view:
  name: MyTodoTasks
  layout: Stack
  data_source: ref:queries/my_todo_all
  components:
    - type: DataTable
      props:
        entity: Task
        create_form: MyTodoCreate
        edit_form: MyTodoEdit
        columns:
          - field: title
          - field: priority
          - field: due_date
```

`data_source` usually points to a named query under `queries/`. A `DataTable` can also define its own `data_source`; this is useful when a view contains more than one table.

## DataTable Row Actions

`row_actions` render per-row actions. They are evaluated against the row state using `from`.

```yaml
row_actions:
  - type: edit
    form: MyTodoEdit
    icon: EditOutlined
    from: [open, in_progress]
  - type: transition
    entity: Task
    transition: start
    icon: PlayCircleOutlined
    from: [open]
  - type: transition
    entity: Task
    transition: archive
    icon: InboxOutlined
    icon_only: true
    from: [open, done]
    payload_fields:
      - name: reason
        type: TextInput
        label: fields.archive_reason
        required: true
  - type: delete
    icon: DeleteOutlined
    icon_only: true
    from: [archived]
```

Common action types:

- `edit`: opens an edit form with the row values.
- `delete`: deletes the row entity after confirmation.
- `transition`: executes a workflow transition for the row entity.
- `link`: opens a same-origin link in a new tab.
- `navigate`: navigates inside the frontend application.
- `reload`: asks Core to reload configuration.

Use `payload_fields` when a transition needs extra input. Payload fields use the same input component names as forms, such as `TextInput`, `NumberInput`, `Select`, `EntitySelect`, `BooleanInput`, or `AppMemberSelect`.

When a row has many actions, a renderer may collapse overflow actions into a menu. Do not depend on exact button placement for business logic.

## Query Include And Reference Labels

`ref:<Entity>` fields store IDs. To show referenced records in a table, include the reference in the query.

```yaml
queries:
  active_care_plans:
    entity: ref:entities/CarePlan
    filter:
      field: current_state
      op: in
      value: [active, at_risk]
    include: [patient_id]
```

Then use `label_template` on the table column.

```yaml
columns:
  - field: patient_id
    label_template: "{{name}}"
```

Without `include`, a renderer only receives the raw ID. `label_template` formats an included object; it does not fetch data by itself.

Templates can read nested fields and can translate values through a namespace.

```yaml
label_template: "{{patient_id.name}} - {{program_key:fields}} - {{current_state:states}}"
```

## Forms And Dynamic Selects

Forms live under `forms/`, but row-action payload fields can also use the same component names.

```yaml
- name: assignee_user_id
  type: AppMemberSelect
  required: true
```

Use `EntitySelect` when a user should choose another entity record.

```yaml
- name: patient_id
  type: EntitySelect
  props:
    options_query: all_patients
    label_template: "{{name}}"
    required: true
```

`options_query` must point to a named query that returns the selectable records. The selected value is normally the referenced entity ID.

## Tabs And Multi-Table Views

Use `TabView` to split a main app screen into focused tabs.

```yaml
view:
  name: MyTodo
  layout: Stack
  components:
    - type: TabView
      props:
        tabs:
          - key: tasks
            label: List
            view: MyTodoTasks
          - key: stats
            label: Stats
            view: MyTodoStats
```

For dashboards or split operational screens, each `DataTable` can fetch a different query.

```yaml
view:
  name: TaskSplitView
  layout: Stack
  components:
    - type: Row
      props:
        gutter: 16
      children:
        - type: Col
          props:
            span: 12
          children:
            - type: DataTable
              props:
                data_source: ref:queries/my_open_tasks
                entity: Task
                columns:
                  - field: title
                  - field: priority
        - type: Col
          props:
            span: 12
          children:
            - type: DataTable
              props:
                data_source: ref:queries/completed_tasks
                entity: Task
                columns:
                  - field: title
                  - field: completed_at
```

## Charts

Charts typically use aggregate queries.

```yaml
queries:
  tasks_by_state:
    entity: ref:entities/Task
    aggregates:
      - func: count
        field: id
        group_by: [current_state]
        as: count
```

```yaml
view:
  name: MyTodoStats
  layout: Stack
  components:
    - type: Chart
      props:
        title: charts.tasks_by_state
        data_source: ref:queries/tasks_by_state
        chartType: column
        xField: name
        yField: count
        xLabelKey: states
```

Supported chart types depend on the renderer. The hosted renderer supports common Ant Design plot types such as column, bar, line, area, and pie.

## Search, Sort, And Pagination

Hosted `DataTable` screens send query parameters to the runtime API:

- `_q`: full-text search across indexed/readable fields.
- `_sort`: field sort; prefix with `-` for descending, for example `-created_at`.
- `_limit`: result limit.
- `_cursor`: cursor for the next page.

Exact renderer behavior can differ between hosted deployments and your own frontend, but named queries should be written so they work with these parameters.

## Icons

Row actions and buttons can reference Ant Design icon names.

```yaml
row_actions:
  - type: transition
    transition: complete
    icon: CheckCircleOutlined
```

Use names from the Ant Design icon catalog:

```text
https://ant.design/components/icon
```

## Detail Screens

Detail screens can combine a single-entity data source with components such as `DetailHeader`, `Descriptions`, cards, or nested tables.

```yaml
view:
  name: TaskDetail
  layout: Stack
  data_source: entity:Task
  params: [id]
  components:
    - type: DetailHeader
      props:
        title: "{bind:data.title}"
        subtitle: "{bind:data.description}"
    - type: Descriptions
      props:
        items:
          - label: Priority
            field: priority
          - label: fields.assignee_user_id
            field: assignee_user_id
```

Support for advanced detail components depends on the renderer. Keep the first version simple and expand only when the target renderer supports the component.

## Component Coverage

The stable app contract is the UP2YOU YAML component model, not the full Ant Design API. The current hosted renderer is Ant Design-backed and supports the common operational components used by the starter examples, including:

- `DataTable`
- `TabView`
- `Chart`
- `StatCard`
- `Text`
- `Alert`
- `DetailHeader`
- layout containers such as `Row`, `Col`, `Space`, and `Divider`

Use the Ant Design component and icon catalogs to understand component behavior and available icon names:

```text
https://ant.design/components/overview
https://ant.design/components/icon
```

Do not assume every Ant Design prop is part of the portable UP2YOU contract. Keep YAML packages limited to props that are used by the starter examples or explicitly documented here, and validate against the target renderer before publishing.

## SDUI vs Tiny Demo

The `web/` folder in this starter is not an SDUI renderer. It is a plain HTML/CSS/JS browser smoke test that calls Core APIs directly.

That split is intentional:

- Use `web/` when you want the simplest possible check that Core APIs work.
- Use SDUI Gateway when you want to render YAML-defined screens.
- Use your own frontend if you want a custom product UI while still using Core APIs.

## Hosted Deployments

Hosted UP2YOU deployments may include their own frontend renderer for SDUI apps. Developers can also upload YAML app packages to a product layer that handles installation, permissions, and rendering.

The hosted renderer is Ant Design-backed. If you build your own frontend, you can render the same SDUI view schema with any UI framework. In that case, you are responsible for mapping UP2YOU component types such as `DataTable`, form fields, `TabView`, `Chart`, and `row_actions` to your own components.

Portable app packages should rely on the documented UP2YOU component model, not private hosted renderer details.

This starter exposes the same app shape locally without including private hosted product source code.
