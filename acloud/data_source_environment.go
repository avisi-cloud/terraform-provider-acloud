package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "Get an environment",
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Slug of the organisation",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the environment",
				Computed:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "Slug of the environment",
				Required:    true,
			},
		},
	}
}

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	slug := d.Get("slug").(string)
	environment, err := client.GetEnvironment(ctx, org, slug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get environment: %w", err))
	}
	d.Set("id", environment.ID)
	d.Set("name", environment.Name)
	d.Set("organisation", org)
	d.Set("slug", slug)
	return nil
}
