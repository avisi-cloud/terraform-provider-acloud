package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceCloudProviderRegions() *schema.Resource {
	return &schema.Resource{
		Description: "List all regions available for the cloud provider for the organisation",
		ReadContext: dataSourceCloudProviderRegionsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cloud_provider_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the region",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "Region Slug",
							Computed:    true,
						},
						"available": {
							Type:        schema.TypeBool,
							Description: "Is the region available for use",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProviderRegionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	organisationSlug := d.Get("organisation_slug").(string)
	cloudProviderSlug := d.Get("cloud_provider_slug").(string)

	regions, err := client.GetRegions(ctx, organisationSlug, cloudProviderSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	setAsID(d, fmt.Sprintf("%s-%s", organisationSlug, cloudProviderSlug))

	providersState := make([]map[string]interface{}, len(regions))
	for i, region := range regions {
		providersState[i] = getRegionAttributes(region)
	}
	d.Set("regions", providersState)
	return nil
}

func getRegionAttributes(region acloudapi.Region) map[string]interface{} {
	return map[string]interface{}{
		"name":      region.Name,
		"slug":      region.Slug,
		"available": region.Available,
	}
}
