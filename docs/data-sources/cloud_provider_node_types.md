---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "acloud_cloud_provider_node_types Data Source - terraform-provider-acloud"
subcategory: ""
description: |-
  List all Node types available on the given cloud provider
---

# acloud_cloud_provider_node_types (Data Source)

List all Node types available on the given cloud provider



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_provider` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `node_types` (List of Object) (see [below for nested schema](#nestedatt--node_types))

<a id="nestedatt--node_types"></a>
### Nested Schema for `node_types`

Read-Only:

- `cpu` (Number)
- `memory` (Number)
- `type` (String)
