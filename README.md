# Terraform Provider for Frends

Manage [Frends iPaaS](https://frends.com) platform resources with Terraform. The provider covers processes, deployments, environment variables, API policies, API keys, API specifications, private applications, and process templates.

## Requirements

| Tool | Version |
|---|---|
| [Terraform](https://developer.hashicorp.com/terraform/downloads) | >= 1.0 |
| [Go](https://golang.org/doc/install) | >= 1.25 |

## Using the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    frends = {
      source  = "frends/frends"
      version = "~> 0.1"
    }
  }
}

provider "frends" {
  host_url = "https://my-company.frends.com"
  token    = var.frends_token
}
```

Credentials can also be supplied via environment variables to avoid hardcoding them:

```bash
export FRENDS_HOST_URL="https://my-company.frends.com"
export FRENDS_TOKEN="your-bearer-token"
```

### Resources

| Resource | Description |
|---|---|
| `frends_process` | Import a `.frends` package and manage a process |
| `frends_process_deployment` | Deploy a process version to an agent group |
| `frends_environment_variable` | Manage an environment variable schema and its per-environment values |
| `frends_api_key` | Manage an API access key for an environment |
| `frends_api_policy` | Manage an API policy with endpoint targets and identity rules |
| `frends_api_specification` | Manage an API specification container |
| `frends_private_application` | Manage a private application for OAuth token-based access |
| `frends_process_template` | Manage a process template |

### Data Sources

| Data Source | Description |
|---|---|
| `frends_environment` | Look up an environment by display name |
| `frends_agent_group` | Look up an agent group by display name within an environment |
| `frends_process` | Look up a process by name (read-only; built in Frends Control Panel) |

See [`docs/index.md`](docs/index.md) for full attribute reference and operation details.

### Example

```hcl
# Import a process package and deploy it to production.
resource "frends_process" "order_processor" {
  package_path = "${path.module}/packages/OrderProcessor.frends"
  package_hash = filesha256("${path.module}/packages/OrderProcessor.frends")
}

data "frends_environment" "production" {
  display_name = "Production"
}

data "frends_agent_group" "agents" {
  display_name   = "Production Agents"
  environment_id = data.frends_environment.production.id
}

resource "frends_process_deployment" "order_processor" {
  agent_group_id    = data.frends_agent_group.agents.id
  activate_triggers = true

  processes {
    process_guid = frends_process.order_processor.guid
    version      = frends_process.order_processor.version
  }

  depends_on = [frends_process.order_processor]
}
```

More examples are in the [`examples/`](examples/) directory.

---

## Development

### Building

```bash
go build ./...
```

To install the provider binary into your `$GOPATH/bin`:

```bash
make install
```

### Running Tests

**Unit tests** (no Frends instance required):

```bash
make test
```

**Acceptance tests** (require a live Frends instance):

```bash
export FRENDS_HOST_URL="https://my-company.frends.com"
export FRENDS_TOKEN="your-bearer-token"
make testacc
```

Acceptance tests create and destroy real resources. Run them against a non-production environment.

### Using a Local Build with Terraform

Create or update `~/.terraformrc` to point Terraform at the locally built binary:

```hcl
provider_installation {
  dev_overrides {
    "frends/frends" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

Then run `make install` and use the provider in any Terraform configuration without needing to publish it.

### Generating Documentation

Provider documentation is generated from schema descriptions using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs):

```bash
make generate
```

The output is written to the `docs/` directory.

### Code Style

Format all Go source files before committing:

```bash
make fmt
```

---

## License

[Mozilla Public License 2.0](LICENSE)
