# Reference App Patterns

The public starter includes `My Todo` as the runnable local demo and a broader `examples/` reference app set for inspecting recurring runtime patterns.

These patterns are useful when designing your own app package.

| Pattern | Use When | Typical Files |
| --- | --- | --- |
| Basic CRUD | Users create, update, list, and delete records. | `entities/`, `forms/`, `views/` |
| State workflow | Records move through controlled states. | `workflows/`, `locales/` |
| Role-gated operations | Different users can see or run different actions. | `auth.yaml`, `workflows/` |
| Profile-level content access | Users can read only records at or below their assigned level. | `entities/` with `read_scope`, profile entity |
| Public read view | Anonymous users can read approved public data. | `queries/`, `views/` |
| Public discovery | Anonymous users can read approved public data while writes stay authenticated. | `queries/`, `views/` |
| Admin queue | Staff process pending records from a focused view. | `views/`, `queries/`, `workflows/` |
| Aggregates and stats | Users need summaries or chart-ready counts. | `queries/`, `views/` |
| Import/export | Users move tabular data in and out of the app. | `entities/`, `views/` |
| Optional integration | The app needs custom side effects outside Core. | `effects/`, `plugins/` |

## Reference Set

The reference apps live under `examples/`. They are compact package patterns, not finished vertical products.

| App | Main Pattern |
| --- | --- |
| My Todo | CRUD, workflow, import/export, optional plugin |
| Simple CRM | Pipeline state workflow and assignment |
| Approval Desk | Approval queue, roles, and decision audit |
| Inventory Lite | Numeric fields, aggregates, and stock state |
| Public Notice Board | Public query/view with private publishing workflow |
| Event Check-in | Public event discovery and private check-in workflow |
| Booking Calendar | Datetime fields and reservation lifecycle |
| Tournament Manager | Participant, match, referee, and public scoreboard flows |
| Mental Health Care Plan | Sensitive role-gated workflow with no public data |

Keep the first public version of an app small. If an app requires a highly custom interface, start with a list or table fallback and add specialized UI after the app contract is stable.

For profile-level content access, child content records should carry a required rank copied from the parent record. Do not rely on a low default rank for child content; that makes mistakes fail open.
