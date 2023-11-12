package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceCloudProviderAvailabilityZones() *schema.Resource {
	return &schema.Resource{
		Description: "List all availablility zones for a given cloud region",
		ReadContext: dataSourceCloudProviderAvailabilityZonesRead,
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
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceCloudProviderAvailabilityZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(acloudapi.Client)

	organisationSlug := d.Get("organisation_slug").(string)
	cloudProviderSlug := d.Get("cloud_provider_slug").(string)
	regionSlug := d.Get("region").(string)

	availabilityZones, err := client.GetAvailabilityZones(ctx, organisationSlug, cloudProviderSlug, regionSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	setAsID(d, fmt.Sprintf("%s-%s", organisationSlug, cloudProviderSlug))

	zones := []string{}
	for _, az := range availabilityZones {
		zones = append(zones, az.Slug)
	}
	d.Set("availability_zones", zones)
	return diags
}
