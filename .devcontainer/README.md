## Dev Container

This repo includes a dev container for GitHub Codespaces and local VS Code Dev Containers.

It installs the obvious stuff:

- Go
- Terraform CLI
- TFLint
- GitHub CLI
- common shell utilities

It also requests a Codespaces host with at least:

- 4 CPUs
- 8 GB RAM
- 32 GB storage

The container no longer does a slow eager `go mod download` during creation. Go dependencies will download when you actually build or test the project.

It does install a pinned `golangci-lint` v2 binary during container creation so `make lint` works without extra manual setup.

Acceptance tests still require Trustgrid credentials, for example:

- `TG_API_KEY_ID`
- `TG_API_KEY_SECRET`
- `TG_API_HOST`
- `TG_ORG_ID`

If you are using Codespaces, add those as Codespaces secrets before running `make testacc`.
