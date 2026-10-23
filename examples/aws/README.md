# Cluster on AWS

A demo environment with one AME Kubernetes cluster in `eu-west-1` and one worker node pool, on an AWS cloud
account that already exists in the Console.

## What it creates

| | |
| --- | --- |
| `acloud_environment` | 1, type `demo` |
| `acloud_cluster` | 1, in `eu-west-1`, Kubernetes version from the `v1.34` update channel |
| `acloud_nodepool` | 1, **1** × `t3.small`, label `role=worker` |
| Output `cluster_identity` | The identity (UUID) of the cluster |

It reads two things that must already exist: the AWS cloud account (by display name) and the update channel.

## The interesting bits

```hcl
data "acloud_update_channel" "channel" {
  name = "v1.34"
}

resource "acloud_cluster" "demo" {
  environment            = acloud_environment.demo.slug
  version                = data.acloud_update_channel.channel.version
  update_channel         = data.acloud_update_channel.channel.name
  cloud_account_identity = data.acloud_cloud_account.aws.identity
  # ...
}
```

The update channel sets both the Kubernetes version and the channel the cluster follows for patch releases, so
the version doesn't have to be typed twice.

`security_updates_on_join = "INSTALL_AND_REBOOT"` on the node pool installs OS security updates on a new node,
and reboots it if needed, before it joins the cluster.

## Run it

1. Create an AWS cloud account in the Console, see
   [Create an AWS cloud account](https://docs.avisi.cloud/docs/how-to/cloud-accounts/aws/create-aws-account/).
2. Plan and apply with your token and organisation:

   ```bash
   terraform init
   terraform apply \
     -var="acloud_token=$ACLOUD_PERSONAL_ACCESS_TOKEN" \
     -var="organisation=example-org" \
     -var="cloud_account_name=AWS Account"
   ```

`terraform destroy` removes the node pool, the cluster and the environment again. The cloud account stays.
