---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "acloud_nodepool_join_config Data Source - terraform-provider-acloud"
subcategory: ""
description: |-
  Provides access to node join configuration for a node pool. Can be used in combination with other terraform providers to provision new Kubernetes Nodes for Bring Your Own Node https://docs.avisi.cloud/product/kubernetes/bring-your-own-node/ clusters in Avisi Cloud Kubernetes.
  With Bring Your Own Node clusters you can retrieve join configuration https://docs.avisi.cloud/docs/how-to/kubernetes-bring-your-own-node/join-nodes-to-cluster/ for new nodes in the form of userdata or install scripts.
  This datasource only works for Bring Your Own Node clusters.
---

# acloud_nodepool_join_config (Data Source)

Provides access to node join configuration for a node pool. Can be used in combination with other terraform providers to provision new Kubernetes Nodes for [Bring Your Own Node](https://docs.avisi.cloud/product/kubernetes/bring-your-own-node/) clusters in Avisi Cloud Kubernetes.

With Bring Your Own Node clusters you can retrieve [join configuration](https://docs.avisi.cloud/docs/how-to/kubernetes-bring-your-own-node/join-nodes-to-cluster/) for new nodes in the form of userdata or install scripts.
This datasource only works for Bring Your Own Node clusters.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster` (String) Slug of the cluster
- `environment` (String) Slug of the environment of the cluster
- `node_pool_id` (String) ID of the node pool

### Optional

- `organisation` (String) Slug of the Organisation

### Read-Only

- `id` (Number) The ID of this resource.
- `install_script` (String) Install bash script for joining a node (base64).
- `join_command` (String)
- `kubelet_config` (String)
- `upgrade_script` (String) Install bash script for upgrading a node (base64)
- `user_data` (String) Cloud Init user-data (base64)
