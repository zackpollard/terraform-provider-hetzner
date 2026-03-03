// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ datasource.DataSource = &vSwitchDataSource{}

func NewVSwitchDataSource() datasource.DataSource {
	return &vSwitchDataSource{}
}

type vSwitchDataSource struct {
	client *client.Client
}

type vSwitchDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Vlan      types.Int64  `tfsdk:"vlan"`
	Cancelled types.Bool   `tfsdk:"cancelled"`
}

func (d *vSwitchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vswitch"
}

func (d *vSwitchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a Hetzner vSwitch.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The unique ID of the vSwitch.",
				Required:            true,
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
	}
}

func (d *vSwitchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vSwitchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vSwitchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/vswitch/%d", data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading vSwitch", err.Error())
		return
	}

	var apiResp vSwitchAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing vSwitch response", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(apiResp.ID))
	data.Name = types.StringValue(apiResp.Name)
	data.Vlan = types.Int64Value(int64(apiResp.Vlan))
	data.Cancelled = types.BoolValue(apiResp.Cancelled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
