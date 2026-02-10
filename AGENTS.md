# AGENTS.md - Terraform Provider TG

Terraform provider for Trustgrid using Terraform Plugin SDK v2.
Manages nodes, clusters, virtual networks, containers, alarms, and more.

## Build, Test, and Lint Commands

```bash
# Build
go build main.go
make install-linux              # Install to ~/.terraform.d/plugins

# Lint
golangci-lint run --tests=false ./...

# Run all tests
go test -v ./...

# Run a single test
go test -v ./acctests -run TestAccAlarm_HappyPath

# Run acceptance tests (requires TG_API_KEY_ID, TG_API_KEY_SECRET)
TF_LOG=ERROR TF_ACC=1 go test -v ./...
```

## Project Structure

```
├── provider/       # Provider configuration
├── resource/       # Terraform resource implementations
├── datasource/     # Terraform data source implementations
├── tg/             # API client and types (JSON tags)
├── hcl/            # Terraform types (tf tags)
├── majordomo/      # Generic CRUD resource framework
├── validators/     # Custom validation functions
├── acctests/       # Acceptance tests
```

## Code Style Guidelines

### Imports
Organize in groups: stdlib, external deps, internal packages.

```go
import (
    "context"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/trustgrid/terraform-provider-tg/hcl"
    "github.com/trustgrid/terraform-provider-tg/tg"
)
```

### Type Definitions

**API Types (tg package)** - use `json` tags:
```go
type Alarm struct {
    UID     string `json:"uid"`
    Name    string `json:"name"`
}
```

**HCL Types (hcl package)** - use `tf` tags, implement `ToTG()` and `UpdateFromTG()`:
```go
type Alarm struct {
    UID  string `tf:"uid"`
    Name string `tf:"name"`
}
func (h Alarm) ToTG() tg.Alarm { ... }
func (h Alarm) UpdateFromTG(a tg.Alarm) HCL[tg.Alarm] { ... }
```

### Resource Implementation

**Using majordomo (preferred for simple CRUD):**
```go
func Alarm() *schema.Resource {
    md := majordomo.NewResource(majordomo.ResourceArgs[tg.Alarm, hcl.Alarm]{
        CreateURL: func(_ hcl.Alarm) string { return "/v2/alarm" },
        UpdateURL: func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
        DeleteURL: func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
        GetURL:    func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
        ID:        func(a hcl.Alarm) string { return a.UID },
    })
    return &schema.Resource{
        ReadContext: md.Read, UpdateContext: md.Update,
        DeleteContext: md.Delete, CreateContext: md.Create,
        Schema: map[string]*schema.Schema{ ... },
    }
}
```

**GET API INFO FROM https://apidocs.trustgrid.io/page-data/shared/oas-index.yaml.json**

**Custom CRUD (for complex resources like Container):**
```go
func (cr *container) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
    tgc := tg.GetClient(meta)
    ct, err := hcl.DecodeResourceData[hcl.Container](d)
    if err != nil {
        return diag.FromErr(err)
    }
    // ... implementation
}
```

**GET API INFO FROM https://apidocs.trustgrid.io/page-data/shared/oas-index.yaml.json**

### Error Handling
```go
var nferr *tg.NotFoundError
switch {
case errors.As(err, &nferr):
    d.SetId("")  // Clear ID on not found
    return nil
case err != nil:
    return diag.FromErr(err)
}
```

### Schema Definition
- Always include `Description`
- Use `validation.StringInSlice()` for enums, `validation.IsUUID` for UUIDs
- Use `validators.IsHostname`, `validators.IsNodeName` for custom validation
- Use `ExactlyOneOf` for mutually exclusive fields

```go
"node_id": {
    Description:  "Node ID",
    Type:         schema.TypeString,
    Optional:     true,
    ValidateFunc: validation.IsUUID,
    ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
},
```

### Test Patterns
Tests in `acctests/`. Name: `TestAcc<Resource>_<Scenario>`.

```go
func TestAccAlarm_HappyPath(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
        Steps: []resource.TestStep{{
            Config: alarmConfig(),
            Check: resource.ComposeTestCheckFunc(
                resource.TestCheckResourceAttr("tg_alarm.test", "name", "test-alarm"),
            ),
        }},
    })
}
```

Use the testify library for its assert functions and require functions.

Only in special cases should you use `t.Errorf`.

## Environment Variables

- `TG_API_KEY_ID`, `TG_API_KEY_SECRET` - API credentials
- `TG_API_HOST` - API endpoint (default: api.trustgrid.io)
- `TG_JWT` - Short-lived JWT (alternative auth)
- `TF_ACC=1` - Enable acceptance tests
- `TF_LOG` - Log level (ERROR, WARN, INFO, DEBUG)

## Linting

Uses golangci-lint with strict settings. Key linters: errcheck, govet, staticcheck,
goimports, gosec, gocritic, revive. Nolint comments require explanation:
```go
//nolint: errcheck // just trusting TF validation here
```

## API

Find API interactions at https://apidocs.trustgrid.io/page-data/shared/oas-index.yaml.json. **ALL types, fields, URLs, and verbs MUST come from that document**.