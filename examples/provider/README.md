# Provider configuration

The smallest working setup: the `acloud` provider block, with the Personal Access Token as a sensitive
variable. This file is also the example on the provider page in the Terraform Registry.

## What it creates

Nothing. It only configures the provider, so `terraform plan` shows no changes once the token is valid.

## The interesting bits

```hcl
provider "acloud" {
  token        = var.acloud_token
  organisation = "example-org" # used by every resource and data source that doesn't set its own
}
```

- `token` can also come from `ACLOUD_PERSONAL_ACCESS_TOKEN`. Then you can leave `token` out entirely.
- `organisation` saves setting `organisation` on every resource. It can also come from `ACLOUD_ORGANISATION`.
- `acloud_api` is only needed for a non-public API; it defaults to `https://api.avisi.cloud`.

## Run it

```bash
terraform init
terraform plan -var="acloud_token=$ACLOUD_PERSONAL_ACCESS_TOKEN"
```

Create a token under API Access in the [Console](https://console.avisi.cloud/), see
[Personal Access Tokens](https://docs.avisi.cloud/docs/product/tasks/how-to/personal-access-tokens).
