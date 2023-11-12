package acloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceCloudProviderNodeTypes() *schema.Resource {
	return &schema.Resource{
		Description: "List all Node types available on the given cloud provider",
		ReadContext: dataSourceCloudProviderNodeTypesRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_provider_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"node_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: "Type of the node",
							Computed:    true,
						},
						"cpu": {
							Type:        schema.TypeInt,
							Description: "CPU count",
							Computed:    true,
						},
						"memory": {
							Type:        schema.TypeInt,
							Description: "Memory in MB",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProviderNodeTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	cloudProviderSlug := d.Get("cloud_provider_slug").(string)

	nodeTypes, err := client.GetNodeTypes(ctx, cloudProviderSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cloudProviderSlug)

	providersState := make([]map[string]interface{}, len(nodeTypes))
	for i, nodeType := range nodeTypes {
		providersState[i] = getNodeTypeAttributes(nodeType)
	}
	d.Set("node_types", providersState)
	return nil
}

func getNodeTypeAttributes(nodeType acloudapi.NodeType) map[string]interface{} {
	return map[string]interface{}{
		"type":   nodeType.Type,
		"cpu":    nodeType.CPU,
		"memory": nodeType.Memory,
	}
}
