# Reference App Patterns

The public starter includes `My Todo` as the runnable example. The broader UP2YOU reference app set was used to validate recurring runtime patterns before this starter was extracted.

These patterns are useful when designing your own app package.

| Pattern | Use When | Typical Files |
| --- | --- | --- |
| Basic CRUD | Users create, update, list, and delete records. | `entities/`, `forms/`, `views/` |
| State workflow | Records move through controlled states. | `workflows/`, `locales/` |
| Role-gated operations | Different users can see or run different actions. | `auth.yaml`, `workflows/` |
| Public read view | Anonymous users can read approved public data. | `queries/`, `views/` |
| Public registration | Anonymous users can submit limited data. | `queries/`, `forms/` |
| Admin queue | Staff process pending records from a focused view. | `views/`, `queries/`, `workflows/` |
| Aggregates and stats | Users need summaries or chart-ready counts. | `queries/`, `views/` |
| Import/export | Users move tabular data in and out of the app. | `entities/`, `views/` |
| Optional integration | The app needs custom side effects outside Core. | `effects/`, `plugins/` |

## Reference Set

| App | Main Pattern |
| --- | --- |
| My Todo | CRUD, workflow, import/export, optional plugin |
| Simple CRM | Pipeline state workflow and assignment |
| Approval Desk | Approval queue, roles, and decision audit |
| Inventory Lite | Numeric fields, aggregates, and stock state |
| Public Notice Board | Public query/view with private publishing workflow |
| Event Check-in | Public registration and private check-in workflow |
| Booking Calendar | Datetime fields and reservation lifecycle |
| Tournament Manager | Participant, match, referee, and public scoreboard flows |
| Mental Health Care Plan | Sensitive role-gated workflow with no public data |

Keep the first public version of an app small. If an app requires a highly custom interface, start with a list or table fallback and add specialized UI after the app contract is stable.
