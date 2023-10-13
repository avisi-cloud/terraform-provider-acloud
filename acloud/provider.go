package acloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ACLOUD_PERSONAL_ACCESS_TOKEN", nil),
			},
			"acloud_api": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ACLOUD_API_ENDPOINT", "https://api.avisi.cloud"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"acloud_environment": resourceEnvironment(),
			"acloud_cluster":     resourceCluster(),
			"acloud_nodepool":    resourceNodepool(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"acloud_organisation":  dataSourceOrganisations(),
			"acloud_environment":   dataSourceEnvironment(),
			"acloud_cloud_account": dataSourceCloudAccount(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	acloudApiEndpoint := d.Get("acloud_api").(string)
	// Warning or errors can be collected in a slice type

	authenticator := acloudapi.NewPersonalAccessTokenAuthenticator(token)
	clientOpts := acloudapi.ClientOpts{
		APIUrl: acloudApiEndpoint,
	}
	var diags diag.Diagnostics
	c := acloudapi.NewClient(authenticator, clientOpts)

	if token != "" {
		c.Resty().OnBeforeRequest(authenticator.Authenticate)
		return c, diags
	}
	return c, diags
}
