## Dev Container

This repo includes a dev container for GitHub Codespaces and local VS Code Dev Containers.

It installs the obvious stuff:

- Go
- Terraform CLI
- TFLint
- GitHub CLI
- common shell utilities

After the container is created it downloads Go modules and installs `golangci-lint`.

Acceptance tests still require Trustgrid credentials, for example:

- `TG_API_KEY_ID`
- `TG_API_KEY_SECRET`
- `TG_API_HOST`
- `TG_ORG_ID`

If you are using Codespaces, add those as Codespaces secrets before running `make testacc`.
