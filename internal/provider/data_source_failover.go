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

var _ datasource.DataSource = &failoverDataSource{}

func NewFailoverDataSource() datasource.DataSource {
	return &failoverDataSource{}
}

type failoverDataSource struct {
	client *client.Client
}

type failoverDataSourceModel struct {
	IP             types.String `tfsdk:"ip"`
	Netmask        types.String `tfsdk:"netmask"`
	ServerIP       types.String `tfsdk:"server_ip"`
	ServerIPv6     types.String `tfsdk:"server_ipv6_net"`
	ServerNumber   types.Int64  `tfsdk:"server_number"`
	ActiveServerIP types.String `tfsdk:"active_server_ip"`
}

func (d *failoverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failover"
}

func (d *failoverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a Hetzner failover IP.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The failover IP address.",
				Required:            true,
			},
			"netmask": schema.StringAttribute{
				MarkdownDescription: "The failover netmask.",
				Computed:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The owner server main IP.",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "The owner server IPv6 network.",
				Computed:            true,
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The owner server number.",
				Computed:            true,
			},
			"active_server_ip": schema.StringAttribute{
				MarkdownDescription: "The server IP the failover is currently routed to.",
				Computed:            true,
			},
		},
	}
}

func (d *failoverDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *failoverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data failoverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/failover/%s", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading failover", err.Error())
		return
	}

	var apiResp failoverAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing failover response", err.Error())
		return
	}

	f := apiResp.Failover
	data.IP = types.StringValue(f.IP)
	data.Netmask = types.StringValue(f.Netmask)
	data.ServerIP = types.StringValue(f.ServerIP)
	data.ServerIPv6 = types.StringValue(f.ServerIPv6)
	data.ServerNumber = types.Int64Value(int64(f.ServerNumber))
	data.ActiveServerIP = types.StringValue(f.ActiveServerIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
