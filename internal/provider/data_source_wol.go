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

var (
	_ datasource.DataSource              = &wolDataSource{}
	_ datasource.DataSourceWithConfigure = &wolDataSource{}
)

type wolDataSource struct {
	client *client.Client
}

type wolDataSourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
}

type wolAPIResponse struct {
	WOL wolAPIData `json:"wol"`
}

type wolAPIData struct {
	ServerIP      string `json:"server_ip"`
	ServerIPv6Net string `json:"server_ipv6_net"`
	ServerNumber  int    `json:"server_number"`
}

func NewWOLDataSource() datasource.DataSource {
	return &wolDataSource{}
}

func (d *wolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wol"
}

func (d *wolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Wake on LAN availability for a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number.",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IPv4 address.",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "Server IPv6 network.",
				Computed:            true,
			},
		},
	}
}

func (d *wolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *wolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data wolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/wol/%d", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading WoL status", err.Error())
		return
	}

	var apiResp wolAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing WoL response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.WOL.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.WOL.ServerIPv6Net)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
