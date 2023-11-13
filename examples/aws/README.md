# Example: AWS

Example for using the `acloud` Terraform Provider with an AWS cloud account.

This requires that you have provisioned a Cloud Account within your Avisi Cloud Console. See [documentation](https://docs.avisi.cloud/docs/how-to/cloud-accounts/aws/create-aws-account/).

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_acloud"></a> [acloud](#requirement\_acloud) | >= 0.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_acloud"></a> [acloud](#provider\_acloud) | 0.1.5 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [acloud_cluster.demo_cluster](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/cluster) | resource |
| [acloud_environment.demo](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/environment) | resource |
| [acloud_nodepool.workers](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/resources/nodepool) | resource |
| [acloud_cloud_account.demo](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/cloud_account) | data source |
| [acloud_update_channel.channel](https://registry.terraform.io/providers/avisi-cloud/acloud/latest/docs/data-sources/update_channel) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_acloud_api"></a> [acloud\_api](#input\_acloud\_api) | Avisi Cloud Platform API endpoint. This is optional. | `any` | n/a | yes |
| <a name="input_acloud_token"></a> [acloud\_token](#input\_acloud\_token) | Your Avisi Cloud Personal Access Token | `any` | n/a | yes |
| <a name="input_cloud_account_name"></a> [cloud\_account\_name](#input\_cloud\_account\_name) | Name of the cloud account that will be used | `string` | n/a | yes |
| <a name="input_environment"></a> [environment](#input\_environment) | Name of the environment that will be provisioned | `any` | n/a | yes |
| <a name="input_organisation"></a> [organisation](#input\_organisation) | Slug of your organisation within the Avisi Cloud Platform | `any` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cluster_identity"></a> [cluster\_identity](#output\_cluster\_identity) | n/a |
