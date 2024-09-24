package acloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaintenanceSchedule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataMaintenanceScheduleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Organisation. Can only be set on creation.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the maintenance schedule",
			},
			"windows": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of maintenance windows for the schedule",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Day of the maintenance window",
						},
						"start_time": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Start time of the maintenance window",
						},
						"duration": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Duration in minutes of the maintenance window",
						},
					},
				},
			},
		},
	}
}

func dataMaintenanceScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	maintenanceSchedule, err := client.GetMaintenanceSchedule(ctx, org, d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(maintenanceSchedule.Identity)
	d.Set("name", maintenanceSchedule.Name)
	d.Set("windows", maintenanceSchedule.MaintenanceWindows)

	return nil
}
