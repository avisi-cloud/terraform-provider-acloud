# Terraform Provider Acloud

Avisi Cloud Platform Terraform Provider for managing your Avisi Cloud resources using Infrastructure as Code practices.

## Documentation

- [registry.terraform.io/providers/avisi-cloud/acloud/latest/docs](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install)

## Supported resources

- `datasource_cloud_account`
- `datasource_cloud_accounts`
- `datasource_cloud_provider_availability_zones`
- `datasource_cloud_provider_node_types`
- `datasource_cloud_provider_regions`
- `datasource_cloud_providers`
- `datasource_cluster`
- `datasource_environment`
- `datasource_nodepool_join_config`
- `datasource_organisation`
- `datasource_update_channel`
- `resource_cluster`
- `resource_environment`
- `resource_nodepool`

## Examples

- [Provider Base](examples/provider)
- [AWS](examples/aws)

## License

[Apache 2.0 License 2.0](lICENSE)

## Contributing

Set up the provider locally and make sure to change the variables if needed:
```bash
make install NAMESPACE=local HOSTNAME=terraform.local OS_ARCH=darwin_arm64
```

Use the locally installed provider in your Terraform configuration:
```hcl
terraform {
  required_providers {
    acloud = {
      source = "terraform.local/local/acloud"
    }
  }
}
```