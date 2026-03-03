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

var _ datasource.DataSource = &failoversDataSource{}

func NewFailoversDataSource() datasource.DataSource {
	return &failoversDataSource{}
}

type failoversDataSource struct {
	client *client.Client
}

type failoversDataSourceModel struct {
	Failovers []failoverItemModel `tfsdk:"failovers"`
}

type failoverItemModel struct {
	IP             types.String `tfsdk:"ip"`
	Netmask        types.String `tfsdk:"netmask"`
	ServerIP       types.String `tfsdk:"server_ip"`
	ServerIPv6     types.String `tfsdk:"server_ipv6_net"`
	ServerNumber   types.Int64  `tfsdk:"server_number"`
	ActiveServerIP types.String `tfsdk:"active_server_ip"`
}

func (d *failoversDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failovers"
}

func (d *failoversDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Hetzner failover IPs.",
		Attributes: map[string]schema.Attribute{
			"failovers": schema.ListNestedAttribute{
				MarkdownDescription: "List of failover IPs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "The failover IP address.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *failoversDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *failoversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/failover")
	if err != nil {
		resp.Diagnostics.AddError("Error listing failovers", err.Error())
		return
	}

	var apiResp []failoverAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing failovers response", err.Error())
		return
	}

	var data failoversDataSourceModel
	for _, item := range apiResp {
		f := item.Failover
		data.Failovers = append(data.Failovers, failoverItemModel{
			IP:             types.StringValue(f.IP),
			Netmask:        types.StringValue(f.Netmask),
			ServerIP:       types.StringValue(f.ServerIP),
			ServerIPv6:     types.StringValue(f.ServerIPv6),
			ServerNumber:   types.Int64Value(int64(f.ServerNumber)),
			ActiveServerIP: types.StringValue(f.ActiveServerIP),
		})
	}

	if data.Failovers == nil {
		data.Failovers = []failoverItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
