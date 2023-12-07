package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrganisations() *schema.Resource {
	return &schema.Resource{
		Description: "Get an organisation",
		ReadContext: dataSourceOrganisationsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOrganisationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	slug, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}
	organisations, err := client.GetMemberships(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, org := range organisations {
		if org.Slug == slug {
			d.Set("id", org.ID)
			d.Set("name", org.Name)
			d.Set("slug", org.Slug)
			d.Set("email", org.Email)
			return nil
		}
	}
	return diag.FromErr(fmt.Errorf("not found"))
}
