package acloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceCloudProviders() *schema.Resource {
	return &schema.Resource{
		Description: "List all Cloud Providers",
		ReadContext: dataSourceCloudProvidersRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cloud_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"available": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProvidersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	organisationSlug := d.Get("organisation_slug").(string)

	cloudProviders, err := client.GetCloudProviders(ctx, organisationSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organisationSlug)

	providers := make([]map[string]interface{}, len(cloudProviders))
	for i, cloudProvider := range cloudProviders {
		providers[i] = getCloudProviderAttributes(cloudProvider)
	}
	d.Set("cloud_providers", providers)
	return nil
}

func getCloudProviderAttributes(cloudProvider acloudapi.CloudProvider) map[string]interface{} {
	return map[string]interface{}{
		"name":      cloudProvider.Name,
		"slug":      cloudProvider.Slug,
		"available": cloudProvider.Available,
	}
}
