# Avisi Cloud Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-7B42BC?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/avisi-cloud/acloud/latest)
[![Provider docs](https://img.shields.io/badge/provider-docs-5C4EE5?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs)
[![Product docs](https://img.shields.io/badge/docs-avisi.cloud-0F6FFF)](https://docs.avisi.cloud/)

Terraform provider for the [Avisi Cloud Platform](https://avisi.cloud). It manages environments,
[Avisi Managed Environments (AME)](https://docs.avisi.cloud/docs/product/why-ame) Kubernetes clusters, node pools,
cloud accounts and maintenance schedules as code, and looks up what already exists: cloud accounts, regions,
availability zones, node types and update channels.

```hcl
terraform {
  required_providers {
    acloud = {
      source  = "avisi-cloud/acloud"
      version = ">= 0.12.0"
    }
  }
}

provider "acloud" {
  organisation = "example-org" # the token comes from ACLOUD_PERSONAL_ACCESS_TOKEN
}

resource "acloud_environment" "production" {
  name = "production"
  type = "production"
}
```

---

## Contents

- [Requirements](#requirements)
- [Authentication](#authentication)
- [Resources and data sources](#resources-and-data-sources)
- [Examples](#examples)
- [Modules built on this provider](#modules-built-on-this-provider)
- [Known rough edges](#known-rough-edges)
- [Development](#development)
- [Releasing](#releasing)
- [Related documentation](#related-documentation)
- [License](#license)

---

## Requirements

| | Version |
| --- | --- |
| Terraform | `>= 0.12` (plugin protocol 5); the examples are written for Terraform 1.x |
| Go, only to build the provider | the `go` version in [`go.mod`](go.mod) |

In Avisi Cloud you need an organisation and a
[Personal Access Token](https://docs.avisi.cloud/docs/product/tasks/how-to/personal-access-tokens).
A cluster also needs a [cloud account](https://docs.avisi.cloud/docs/how-to/cloud-accounts/aws/create-aws-account/)
that already exists in the [Console](https://console.avisi.cloud/), or one created with `acloud_cloud_account`.

## Authentication

Every provider setting can come from an environment variable, so the token never has to be in a `.tf` file:

| Setting | Environment variable | Default |
| --- | --- | --- |
| `token` | `ACLOUD_PERSONAL_ACCESS_TOKEN` | required |
| `acloud_api` | `ACLOUD_API_ENDPOINT` | `https://api.avisi.cloud` |
| `organisation` | `ACLOUD_ORGANISATION` | none: set it here or on each resource |

```bash
export ACLOUD_PERSONAL_ACCESS_TOKEN="..."
export ACLOUD_ORGANISATION="example-org"
terraform plan
```

The organisation slug is shown on the Settings page in the Console. Resources and data sources that set their own
`organisation` use that one instead.

## Resources and data sources

Each name links to its page in the Terraform Registry, with all attributes and an example.

| Resource | What it manages |
| --- | --- |
| [`acloud_environment`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/environment) | An environment that groups clusters, for example `production` |
| [`acloud_cluster`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/cluster) | An AME Kubernetes cluster in an environment, with its add-ons |
| [`acloud_nodepool`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/nodepool) | A group of nodes of one node type in a cluster |
| [`acloud_cloud_account`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/cloud_account) | A cloud account, based on a cloud profile |
| [`acloud_maintenance_schedule`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/maintenance_schedule) | Weekly maintenance windows for clusters |

| Data source | What it looks up |
| --- | --- |
| [`acloud_organisation`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/organisation) | An organisation |
| [`acloud_environment`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/environment) | An environment |
| [`acloud_cluster`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cluster) | A cluster in an environment |
| [`acloud_nodepool`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/nodepool) | A node pool of a cluster (see [Known rough edges](#known-rough-edges)) |
| [`acloud_nodepool_join_config`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/nodepool_join_config) | The join configuration of a Bring Your Own Node node pool |
| [`acloud_cloud_account`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_account) | A cloud account by display name |
| [`acloud_cloud_accounts`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_accounts) | All cloud accounts, optionally filtered |
| [`acloud_cloud_profile`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_profile) | A cloud profile, used to create a cloud account |
| [`acloud_cloud_providers`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_providers) | The cloud providers of the organisation |
| [`acloud_cloud_provider_regions`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_provider_regions) | The regions of a cloud provider |
| [`acloud_cloud_provider_availability_zones`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_provider_availability_zones) | The availability zones of a region |
| [`acloud_cloud_provider_node_types`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_provider_node_types) | The node types of a cloud provider, with CPU and memory |
| [`acloud_update_channel`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/update_channel) | An update channel and its current Kubernetes version |
| [`acloud_maintenance_schedule`](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/maintenance_schedule) | A maintenance schedule by its identity |

## Examples

Complete configurations you can run, each with a README:

| Example | What it shows |
| --- | --- |
| [`examples/provider`](examples/provider) | The provider block and the token as a sensitive variable |
| [`examples/aws`](examples/aws) | An environment, a cluster on an existing AWS cloud account and one node pool |
| [`examples/cyso-ams2`](examples/cyso-ams2) | A highly available, private cluster on Cyso Cloud AMS2 with one node pool per availability zone |

Short examples per resource and data source live in [`examples/resources`](examples/resources) and
[`examples/data-sources`](examples/data-sources). They are the "Example Usage" blocks on the Registry pages.

## Modules built on this provider

| Module | Use it for |
| --- | --- |
| [`avisi-cloud/cluster/acloud`](https://registry.terraform.io/modules/avisi-cloud/cluster/acloud/latest) | A cluster with its node pools from one `module` block |
| [`avisi-cloud/nodepool/acloud`](https://registry.terraform.io/modules/avisi-cloud/nodepool/acloud/latest) | Node pools for an existing cluster, optionally one per availability zone |

Use the resources directly when you need something the modules don't expose.

## Known rough edges

- **Import only works for `acloud_cloud_account`.** The other resources read with values that an import doesn't
  set (`slug`, `environment`, `cluster`), so `terraform import` of an environment, cluster or node pool fails.
- **`acloud_nodepool` needs `upgrade_strategy` on create.** Leaving it out fails with
  `cannot parse upgradeStrategy: unsupported upgrade strategy`. Set one of the values on the resource page.
- **The `acloud_nodepool` data source can't select a pool.** It has no input for the node pool ID, so it always
  ends in `nodepool was not found`. Read node pools through `acloud_nodepool` resources or the modules instead.
- **The `acloud_maintenance_schedule` data source requires `organisation`,** even when the provider sets one.

## Development

```bash
make            # list all targets
make check      # what CI runs: fmt-check, vet, test and docs-check
```

| Target | What it does |
| --- | --- |
| `make build` | Build the provider into `bin/` |
| `make dev` | Build and write `bin/dev.terraformrc`; with `export TF_CLI_CONFIG_FILE=$PWD/bin/dev.terraformrc`, `terraform plan`/`apply` use the local build, no `terraform init` needed |
| `make test` | Unit tests. They run against a fake API, no account needed |
| `make testacc` | Acceptance tests against a real API; they create real resources |
| `make fmt` / `make fmt-check` | `gofmt` for the Go code, `terraform fmt` for `examples/` |
| `make docs` / `make docs-check` | Regenerate `docs/` with tfplugindocs / fail when it is out of date |
| `make examples-validate` | `terraform validate` every example in `examples/` against the local build |
| `make install` | The older way to use a local build: install into `~/.terraform.d/plugins` |

### Documentation

`docs/` is generated, so don't edit it by hand. It comes from:

| Source | Becomes |
| --- | --- |
| `Description` fields in the schema (`acloud/*.go`) | The attribute lists |
| [`templates/index.md.tmpl`](templates/index.md.tmpl) | The provider page |
| `examples/provider/provider.tf` | The example on the provider page |
| `examples/resources/<name>/resource.tf`, `import.sh` | "Example Usage" and "Import" on a resource page |
| `examples/data-sources/<name>/data-source.tf` | "Example Usage" on a data source page |

Change one of those, run `make docs` and commit the result. CI fails when `docs/` doesn't match.

## Releasing

Push a tag such as `v0.13.0`. The [release workflow](.github/workflows/release.yml) builds and signs the binaries
with [goreleaser](.goreleaser.yaml) and publishes the GitHub release, which the Terraform Registry picks up.

## Related documentation

- [Provider documentation in the Terraform Registry](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs)
- [Avisi Cloud product documentation](https://docs.avisi.cloud/)
- [Personal Access Tokens](https://docs.avisi.cloud/docs/product/tasks/how-to/personal-access-tokens)

## License

[Apache 2.0](LICENSE)
