# UP2YOU Documentation

This folder documents the public UP2YOU starter runtime and the portable YAML app package contract.

Start here when you need to choose which document to read. The starter is a core-only local runtime; hosted product flows may add account, installation, billing, review, or workspace behavior that is not documented in this repository.

## Version Compatibility

These docs describe the runtime and app package contract included with this starter repository revision. When upgrading runtime Docker images or copying docs between branches, verify the image tag, starter commit, and app package validator behavior together.

If a hosted deployment documents additional product-layer behavior, treat that hosted documentation as the source of truth for deployment-specific APIs and policies.

## Read By Goal

| Goal | Read |
| --- | --- |
| Start the local runtime and smoke-test the demo | [quick-start.md](quick-start.md), [core-only.md](core-only.md) |
| Build a YAML app package | [app-development.md](app-development.md), [yaml-contract.md](yaml-contract.md), [sdui.md](sdui.md), [i18n.md](i18n.md) |
| Package and upload an app ZIP | [packaging.md](packaging.md), [yaml-contract.md](yaml-contract.md) |
| Understand the runtime architecture | [architecture.md](architecture.md), [api-reference.md](api-reference.md) |
| Work with auth, roles, and permissions | [authentication.md](authentication.md), [errors.md](errors.md) |
| Add custom side effects or integrations | [plugins.md](plugins.md) |
| Debug local runtime or app package issues | [troubleshooting.md](troubleshooting.md), [errors.md](errors.md) |
| Generate an app with an AI coding assistant | [ai-prompt.md](ai-prompt.md), [reference-patterns.md](reference-patterns.md) |

## Read By Role

| Reader | Recommended path |
| --- | --- |
| App developer | [quick-start.md](quick-start.md) -> [app-development.md](app-development.md) -> [yaml-contract.md](yaml-contract.md) -> [sdui.md](sdui.md) -> [i18n.md](i18n.md) -> [packaging.md](packaging.md) |
| Frontend or SDUI renderer developer | [sdui.md](sdui.md) -> [api-reference.md](api-reference.md) -> [i18n.md](i18n.md) |
| Plugin developer | [plugins.md](plugins.md) -> [yaml-contract.md](yaml-contract.md) -> [errors.md](errors.md) |
| Operator running the starter locally | [core-only.md](core-only.md) -> [architecture.md](architecture.md) -> [troubleshooting.md](troubleshooting.md) |
| Auth or integration developer | [authentication.md](authentication.md) -> [api-reference.md](api-reference.md) -> [errors.md](errors.md) |
| Product or hosted deployment reader | [architecture.md](architecture.md) -> [packaging.md](packaging.md), then use deployment-specific hosted documentation for product-layer behavior |

## Document Index

| Document | Purpose |
| --- | --- |
| [quick-start.md](quick-start.md) | First-run setup, demo app exploration, and a minimal app walkthrough. |
| [app-development.md](app-development.md) | High-level guide to creating YAML app folders. |
| [yaml-contract.md](yaml-contract.md) | Practical YAML package rules, naming conventions, references, public access, and compatibility checklist. |
| [sdui.md](sdui.md) | Server-driven UI view and form model, DataTable actions, charts, labels, and renderer boundaries. |
| [i18n.md](i18n.md) | Locale file shape, label keys, `label_template`, and renderer translation responsibilities. |
| [packaging.md](packaging.md) | ZIP package shape, upload checklist, validator expectations, and common upload errors. |
| [api-reference.md](api-reference.md) | Core Engine and SDUI Gateway HTTP endpoints. |
| [authentication.md](authentication.md) | JWT model, app roles, effective role resolution, public routes, and cross-tenant access. |
| [architecture.md](architecture.md) | Core, SDUI Gateway, Plugin Host, PostgreSQL, NATS, NGINX, request flows, and tenant isolation. |
| [plugins.md](plugins.md) | Optional HTTP plugin contract, scheduled effect model, effect mapping, reliability, and security guidance. |
| [errors.md](errors.md) | Error response shape, error codes, causes, and fixes. |
| [troubleshooting.md](troubleshooting.md) | Local debugging checklists for compose, schema, auth, transitions, plugins, and database issues. |
| [core-only.md](core-only.md) | Starter-specific runtime usage without a hosted product layer. |
| [reference-patterns.md](reference-patterns.md) | Common app patterns and reference example set. |
| [ai-prompt.md](ai-prompt.md) | Reusable prompt for AI-assisted YAML app generation. |

## Starter Boundaries

This repository documents:

- Local core-only runtime behavior.
- Portable YAML app package structure.
- SDUI schema contracts used by app packages.
- Optional HTTP plugin shape.
- Scheduled plugin effect package contract.
- Core API, auth, error, and troubleshooting behavior.

This repository does not document private hosted product source code, production account/session APIs, billing, workspace lifecycle, package review policy, or deployment-specific operational rules.
