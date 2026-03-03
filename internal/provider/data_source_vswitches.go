// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ datasource.DataSource = &vSwitchesDataSource{}

func NewVSwitchesDataSource() datasource.DataSource {
	return &vSwitchesDataSource{}
}

type vSwitchesDataSource struct {
	client *client.Client
}

type vSwitchesDataSourceModel struct {
	VSwitches []vSwitchItemModel `tfsdk:"vswitches"`
}

type vSwitchItemModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Vlan      types.Int64  `tfsdk:"vlan"`
	Cancelled types.Bool   `tfsdk:"cancelled"`
}

func (d *vSwitchesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vswitches"
}

func (d *vSwitchesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Hetzner vSwitches.",
		Attributes: map[string]schema.Attribute{
			"vswitches": schema.ListNestedAttribute{
				MarkdownDescription: "List of vSwitches.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The unique ID of the vSwitch.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the vSwitch.",
							Computed:            true,
						},
						"vlan": schema.Int64Attribute{
							MarkdownDescription: "The VLAN ID.",
							Computed:            true,
						},
						"cancelled": schema.BoolAttribute{
							MarkdownDescription: "Whether the vSwitch has been cancelled.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *vSwitchesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *vSwitchesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/vswitch")
	if err != nil {
		resp.Diagnostics.AddError("Error listing vSwitches", err.Error())
		return
	}

	var apiResp []vSwitchAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing vSwitches response", err.Error())
		return
	}

	var data vSwitchesDataSourceModel
	for _, v := range apiResp {
		data.VSwitches = append(data.VSwitches, vSwitchItemModel{
			ID:        types.Int64Value(int64(v.ID)),
			Name:      types.StringValue(v.Name),
			Vlan:      types.Int64Value(int64(v.Vlan)),
			Cancelled: types.BoolValue(v.Cancelled),
		})
	}

	if data.VSwitches == nil {
		data.VSwitches = []vSwitchItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
