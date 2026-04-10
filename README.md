# terraform-provider-tg

Terraform provider for Trustgrid.

## Development

This repo includes a devcontainer for GitHub Codespaces and VS Code Dev Containers. See `.devcontainer/README.md` for the container setup, sizing, and required secrets for acceptance tests.

Acceptance tests use a real Trustgrid environment. Set these before running `make testacc`:

- `TG_API_KEY_ID`
- `TG_API_KEY_SECRET`
- `TG_API_HOST`
- `TG_ORG_ID`

## Make targets

### `make build`

Builds the provider binary locally.

```bash
make build
```

Runs:

```bash
go build main.go
```

### `make lint`

Runs `golangci-lint` against the Go code.

```bash
make lint
```

Runs:

```bash
golangci-lint run --tests=false ./...
```

### `make test`

Runs the full Go test suite.

```bash
make test
```

Runs:

```bash
go test -v ./...
```

Without `TF_ACC=1`, acceptance tests are skipped.

### `make sweep`

Runs acceptance test sweepers to clean up test resources in the configured Trustgrid environment.

```bash
make sweep
```

Runs:

```bash
go test -v ./acctests -v -sweep=all
```

Use this when acceptance tests leave junk behind or before running a full acceptance pass in a shared environment.

### `make testacc`

Runs acceptance tests against a real Trustgrid environment.

```bash
make testacc
```

Runs:

```bash
TF_LOG=$(LOG_LEVEL) TF_ACC=1 go test -v -timeout=5m -run '$(TEST)' ./...
```

Supported variables:

- `TEST` — regex for selecting tests, defaults to `.*`
- `LOG_LEVEL` — Terraform log level, defaults to `ERROR`

Examples:

```bash
make testacc TEST=TestAccVirtualNetwork_HappyPath
make testacc TEST='TestAccVPN.*' LOG_LEVEL=DEBUG
```

### `make docs`

Regenerates provider documentation.

```bash
make docs
```

Runs:

```bash
go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
go generate
go mod tidy
```

This updates generated docs and example formatting.

### `make install-linux`

Builds and installs the provider binary into the local Terraform plugin path for Linux.

```bash
make install-linux
```

Installs to:

```text
~/.terraform.d/plugins/hashicorp.com/trustgrid/tg/0.1/linux_amd64/terraform-provider-tg
```

### `make install-osx`

Builds and installs the provider binary into the local Terraform plugin path for macOS.

```bash
make install-osx
```

Installs to:

```text
~/.terraform.d/plugins/hashicorp.com/trustgrid/tg/0.1/darwin_amd64/terraform-provider-tg
```

### `make run`

Runs the Terraform proof-of-concept config under `poc/` after reinitializing Terraform state for that directory.

```bash
make run
```

Runs:

```bash
cd poc && rm -rf .terraform* && terraform init && terraform apply -auto-approve
```

### `make reset`

Fully resets the Terraform proof-of-concept config under `poc/`, including state files, then reapplies it.

```bash
make reset
```

Runs:

```bash
cd poc && rm -rf .terraform* *.tfstate && terraform init && terraform apply -auto-approve
```

Use this when the `poc/` workspace is wedged and you want a fresh apply.

## Suggested workflow

For normal code changes:

```bash
make lint
make test
```

For acceptance-test changes:

```bash
make sweep
make testacc
```
