# UP2YOU Example Apps

These folders are reference YAML app packages for learning the UP2YOU package model.

They are intentionally compact. Use them to inspect entity, workflow, query, form, view, locale, public view, import/export, and lifecycle patterns. They are not intended to be complete production products as-is.

## Apps

- `my-todo` — minimal task app and the default local demo.
- `approval-desk` — approval workflow with role-based queues and audit-friendly state transitions.
- `inventory-lite` — inventory tracking with import/export and aggregate chart queries.
- `public-notice-board` — notice publishing with public read-only views.
- `event-checkin` — public event discovery and code-based check-in operations.
- `booking-calendar` — reservation lifecycle with date-time fields.
- `simple-crm` — lightweight CRM pipeline for leads.
- `tournament-manager` — participants, matches, referee assignment, and public scoreboard.
- `mental-health-care-plan` — care plan follow-ups, mood check-ins, risk alerts, and clinician tasks.

## Package

From the repository root:

```bash
./scripts/package-app.sh examples/my-todo
```

Replace `my-todo` with any example directory name.
