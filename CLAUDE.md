# CLAUDE.md

Guidance for working in this repository.

## Taskfile conventions

This repo uses [go-task](https://taskfile.dev) with nested taskfiles composed
via `includes:`. The root [taskfile.yaml](taskfile.yaml) is the assembler;
leaf taskfiles live next to the code or `.tf` files they drive.

**Never put `env:` at the root of a taskfile that is (or could become) an
`include`.** Tasks are namespaced when included; a taskfile-level `env:` block
is **not** — it merges into the entire `task` run and outranks the OS
environment. A root-level `env:` in one leaf will silently override variables
for unrelated tasks elsewhere (this once made `apply-prod` deploy to the
wrong S3 bucket, caught only via `task --dry`).

Instead, scope env to the tasks that need it. To stay DRY, define a top-level
YAML anchor and reference it per task (Task tolerates unknown `x-` keys):

```yaml
x-local-env: &local_env
  SOME_VAR: value

tasks:
  run:
    env: *local_env
    cmds: [ ... ]
```

See [cmd/lambda-local/taskfile.yaml](cmd/lambda-local/taskfile.yaml) for the
canonical example.

Related discipline: structure leaf taskfiles with no cross-subtree includes or
upward (`../../`) references — keep orchestration that spans subtrees in the
smallest enclosing assembler (the root taskfile for prod deploy).

## Common tasks

- `task dev` — full local stack (MinIO + lambda-local + Vite).
- `task apply-prod` / `task destroy-prod` — build + Terraform apply/destroy for prod.
- `task deploy:local:test` — unit tests for the local Lambda handler.

Infra tasks use Terraform's own verbs (`apply`/`destroy`) consistently rather
than `deploy`/`undeploy`.
