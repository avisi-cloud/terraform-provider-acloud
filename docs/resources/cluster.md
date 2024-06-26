---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "acloud_cluster Resource - terraform-provider-acloud"
subcategory: ""
description: |-
  Create an Avisi Cloud Kubernetes cluster within an environment
---

# acloud_cluster (Resource)

Create an Avisi Cloud Kubernetes cluster within an environment



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_account_identity` (String) Identity of the Cloud Account used to deploy the Cluster. Can only be set on cluster creation.
- `environment` (String) Slug of the Environment of the Cluster. Can only be set on cluster creation.
- `name` (String) Name of the Cluster
- `region` (String) Region of the Cloud Provider to deploy the Cluster in. Can only be set on cluster creation.
- `version` (String) Avisi Cloud Kubernetes version of the Cluster

### Optional

- `cluster_state_wait_seconds` (Number) Time-out for waiting until the cluster reaches the desired state
- `cni` (String) CNI plugin for Kubernetes
- `description` (String) Description of the Cluster
- `enable_high_available_control_plane` (Boolean) Enable Highly-Availability mode for the cluster's Kubernetes Control Plane
- `enable_multi_availability_zones` (Boolean) Enable multi availability zones for the cluster
- `enable_network_encryption` (Boolean) Enable Network Encryption at the node level (if supported by the CNI).
- `enable_private_cluster` (Boolean) Enable Private Cluster mode. Can only be set on cluster creation.
- `environment_slug` (String, Deprecated)
- `organisation` (String) Slug of the Organisation of the Cluster. Can only be set on cluster creation.
- `organisation_slug` (String, Deprecated)
- `pod_security_standards_profile` (String) Pod Security Standards used by default within the cluster
- `stopped` (Boolean) Stops the Cluster if set to true. False by default
- `update_channel` (String) Avisi Cloud Kubernetes Update Channel that the Cluster follows

### Read-Only

- `cloud_provider` (String)
- `id` (String) The ID of this resource.
- `slug` (String)
- `status` (String)
