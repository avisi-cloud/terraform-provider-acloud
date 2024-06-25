package acloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Create an cloud account",
		CreateContext: resourceCloudAccountCreate,
		ReadContext:   resourceCloudAccountRead,
		UpdateContext: resourceCloudAccountUpdate,
		DeleteContext: resourceCloudAccountDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Cloud Account",
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Slug of the Organisation",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Enable the cloud account",
			},
			"primary_cloud_credentials_identity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identity of the primary cloud credentials",
			},
			"cloud_profile_identity": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identity of the cloud profile",
			},
			"cloud_profile_cloud_provider": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cloud provider of the cloud profile",
			},
			"regions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Regions of the cloud account",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// metadata fields
			"vsphere_parent_folder": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vSphere parent folder",
			},
			"vsphere_parent_resource_pool": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vSphere parent resource pool",
			},
			"openstack_tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "OpenStack tenant ID",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func nilOrString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func resourceCloudAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	vsphereParentFolder := d.Get("vsphere_parent_folder").(string)
	vSphereParentResourcePool := d.Get("vsphere_parent_resource_pool").(string)
	openStackTenantID := d.Get("openstack_tenant_id").(string)

	createCloudAccount := acloudapi.CreateCloudAccount{
		DisplayName:  d.Get("display_name").(string),
		CloudProfile: d.Get("cloud_profile_identity").(string),
		Metadata: acloudapi.CloudAccountMetadata{
			VsphereParentFolder:       nilOrString(vsphereParentFolder),
			VSphereParentResourcePool: nilOrString(vSphereParentResourcePool),
			OpenStackTenantID:         nilOrString(openStackTenantID),
		},
	}

	cloudAccount, err := client.CreateCloudAccount(ctx, org, createCloudAccount)
	if err != nil {
		return diag.FromErr(err)
	}
	if cloudAccount != nil {
		d.SetId(cloudAccount.Identity)
		d.Set("identity", cloudAccount.Identity)
		d.Set("display_name", cloudAccount.DisplayName)
		d.Set("enabled", cloudAccount.Enabled)
		d.Set("primary_cloud_credentials_identity", cloudAccount.PrimaryCloudCredentialsIdentity)
		d.Set("vsphere_parent_folder", cloudAccount.Metadata.VsphereParentFolder)
		d.Set("vsphere_parent_resource_pool", cloudAccount.Metadata.VSphereParentResourcePool)
		d.Set("openstack_tenant_id", cloudAccount.Metadata.OpenStackTenantID)
		d.Set("cloud_profile_identity", cloudAccount.CloudProfile.Identity)
		d.Set("cloud_profile_cloud_provider", cloudAccount.CloudProfile.CloudProvider)
		d.Set("regions", cloudAccount.CloudProfile.Regions)
		return nil
	}
	return resourceCloudAccountRead(ctx, d, m)
}

func resourceCloudAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	identity := d.Get("id").(string)
	cloudAccounts, err := client.GetCloudAccounts(ctx, org)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(cloudAccounts) == 0 {
		return diag.FromErr(fmt.Errorf("cloud account was not found: empty list"))
	}

	var cloudAccount acloudapi.CloudAccount
	for _, item := range cloudAccounts {
		if item.Identity == identity {
			cloudAccount = item
			break
		}
	}
	if cloudAccount.Identity == "" {
		return diag.FromErr(fmt.Errorf("cloud account was not found: no identity: %v", cloudAccounts))
	}

	d.SetId(cloudAccount.Identity)
	d.Set("identity", cloudAccount.Identity)
	d.Set("display_name", cloudAccount.DisplayName)
	d.Set("enabled", cloudAccount.Enabled)
	d.Set("primary_cloud_credentials_identity", cloudAccount.PrimaryCloudCredentialsIdentity)
	d.Set("vsphere_parent_folder", cloudAccount.Metadata.VsphereParentFolder)
	d.Set("vsphere_parent_resource_pool", cloudAccount.Metadata.VSphereParentResourcePool)
	d.Set("openstack_tenant_id", cloudAccount.Metadata.OpenStackTenantID)
	d.Set("cloud_profile_identity", cloudAccount.CloudProfile.Identity)
	d.Set("cloud_profile_cloud_provider", cloudAccount.CloudProfile.CloudProvider)
	d.Set("regions", cloudAccount.CloudProfile.Regions)

	return nil
}

func resourceCloudAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}
	identity := d.Get("id").(string)
	updateCloudAccount := acloudapi.UpdateCloudAccount{
		DisplayName: d.Get("display_name").(string),
		Enabled:     d.Get("enabled").(bool),
	}
	cloudAccount, err := client.UpdateCloudAccount(ctx, org, identity, updateCloudAccount)
	if err != nil {
		return diag.FromErr(err)
	}
	if cloudAccount == nil {
		return diag.FromErr(fmt.Errorf("cloud account was not found"))
	}
	return resourceCloudAccountRead(ctx, d, m)
}

func resourceCloudAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	identity := d.Get("id").(string)

	error := client.DeleteCloudAccount(ctx, org, identity)
	if error != nil {
		return diag.FromErr(error)
	}

	d.SetId("")

	return nil
}
