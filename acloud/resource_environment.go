package acloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	createEnvironment := acloudapi.CreateEnvironment{
		Name:        d.Get("name").(string),
		Purpose:     d.Get("purpose").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	org := d.Get("organisation").(string)

	environment, err := client.CreateEnvironment(ctx, createEnvironment, org)

	if err != nil {
		return diag.FromErr(err)
	}
	if environment != nil {
		d.SetId(strconv.Itoa(environment.ID))
		d.Set("slug", environment.Slug)
		return diags
	}
	return diag.FromErr(fmt.Errorf("environment was not created"))
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orgSlug := d.Get("organisation").(string)
	slug := d.Get("slug").(string)
	environment, err := client.GetEnvironment(ctx, orgSlug, slug)
	if err != nil {
		return diag.FromErr(err)
	}
	if environment != nil {
		return diags
	}
	return diag.FromErr(fmt.Errorf("environment was not found"))
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	updateEnvironment := acloudapi.UpdateEnvironment{
		Name:        d.Get("name").(string),
		Purpose:     d.Get("purpose").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	org := d.Get("organisation").(string)
	env := d.Get("slug").(string)

	environment, err := client.UpdateEnvironment(ctx, updateEnvironment, org, env)
	if err != nil {
		return diag.FromErr(err)
	}
	if environment != nil {
		d.Set("name", environment.Name)
		d.Set("purpose", environment.Purpose)
		d.Set("type", environment.Type)
		d.Set("description", environment.Description)
		d.Set("slug", environment.Slug)
		return diags
	}

	return diags
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orgSlug := d.Get("organisation").(string)
	slug := d.Get("slug").(string)

	err := client.DeleteEnvironment(ctx, orgSlug, slug)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
