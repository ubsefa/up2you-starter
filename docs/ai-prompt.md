# AI Assistant Prompt

Use this reusable prompt with your AI coding assistant when you want help creating an UP2YOU YAML app.

```text
You are helping me create an UP2YOU YAML app package.

First read these docs from this repository:

- docs/yaml-contract.md
- docs/app-development.md
- docs/sdui.md
- docs/i18n.md
- docs/packaging.md
- docs/plugins.md
- docs/reference-patterns.md

Use examples/my-todo as the reference app shape.

If the app needs runtime API, auth, error handling, or self-hosted runtime details, also read the relevant docs in this repository: docs/api-reference.md, docs/authentication.md, docs/errors.md, docs/core-only.md, and docs/architecture.md.

Use only the docs available in this repository by default. Hosted Platform docs are private and may require authentication; do not assume access to them. If you need hosted-specific behavior, ask me for the relevant page content instead of guessing.

App to build:
[Describe your app here: domain, users, entities, main workflow, public surface, and any known limits.]

Goal:
Create a packageable UP2YOU app folder that can run in the local starter and be zipped for the hosted UP2YOU Platform.

Rules:
- Do not assume access to private UP2YOU Platform source code.
- Do not add production secrets, local machine paths, or private repository paths.
- Keep app keys lowercase with hyphens.
- Use PascalCase entity names and snake_case field names.
- Put app metadata in app.yaml.
- Put app permissions in auth.yaml.
- Define persistent models in entities/.
- Define state transitions in workflows/.
- Define reusable reads and public reads in queries/.
- Define SDUI screens in views/.
- Define create/edit inputs in forms/.
- Define locale labels in locales/.
- Keep locale keys and label_template usage consistent with docs/i18n.md.
- Use effects/ and plugins/ only when custom side effects are truly needed.
- Mark queries or views public only when anonymous users can safely access the data.
- Keep the first version small and predictable.

Output:
1. Briefly describe the app.
2. Show the proposed file tree.
3. Create or edit the YAML/JSON files directly.
4. Explain any public data, role, workflow, or plugin decisions.
5. Make sure the package root contains app.yaml and can be zipped for upload.

Before finishing:
- Check every ref: points to an existing resource.
- Check every transition references valid states.
- Check every view/query/form field exists on the relevant entity.
- Check locales/ files contain labels for user-facing fields, states, transitions, and views.
- Check every workflow effect name is registered in app.yaml or remove the effect.
- Check public queries expose only safe data.
- Check app.yaml main_view points to an existing view.
```

For advanced app design, plugin deployment, production usage, or hosted deployment questions, contact `admin@up2you.app`.
