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

var _ datasource.DataSource = &serverDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	client *client.Client
}

type serverDataSourceModel struct {
	ServerNumber types.Int64  `tfsdk:"server_number"`
	ServerName   types.String `tfsdk:"server_name"`
	ServerIP     types.String `tfsdk:"server_ip"`
	ServerIPv6   types.String `tfsdk:"server_ipv6_net"`
	Product      types.String `tfsdk:"product"`
	DC           types.String `tfsdk:"dc"`
	Traffic      types.String `tfsdk:"traffic"`
	Status       types.String `tfsdk:"status"`
	Cancelled    types.Bool   `tfsdk:"cancelled"`
	PaidUntil    types.String `tfsdk:"paid_until"`
}

func (d *serverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The unique server number (ID).",
				Required:            true,
			},
			"server_name": schema.StringAttribute{
				MarkdownDescription: "The user-assigned server name.",
				Computed:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The main IPv4 address of the server.",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "The main IPv6 network of the server.",
				Computed:            true,
			},
			"product": schema.StringAttribute{
				MarkdownDescription: "The product name.",
				Computed:            true,
			},
			"dc": schema.StringAttribute{
				MarkdownDescription: "The data center.",
				Computed:            true,
			},
			"traffic": schema.StringAttribute{
				MarkdownDescription: "Free traffic quota.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Server status.",
				Computed:            true,
			},
			"cancelled": schema.BoolAttribute{
				MarkdownDescription: "Whether the server has been cancelled.",
				Computed:            true,
			},
			"paid_until": schema.StringAttribute{
				MarkdownDescription: "Date the server is paid until.",
				Computed:            true,
			},
		},
	}
}

func (d *serverDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/server/%d", data.ServerNumber.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	var apiResp serverDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing server response", err.Error())
		return
	}

	s := apiResp.Server
	data.ServerNumber = types.Int64Value(int64(s.ServerNumber))
	data.ServerName = types.StringValue(s.ServerName)
	data.ServerIP = types.StringValue(s.ServerIP)
	data.ServerIPv6 = types.StringValue(s.ServerIPv6)
	data.Product = types.StringValue(s.Product)
	data.DC = types.StringValue(s.DC)
	data.Traffic = types.StringValue(s.Traffic)
	data.Status = types.StringValue(s.Status)
	data.Cancelled = types.BoolValue(s.Cancelled)
	data.PaidUntil = types.StringValue(s.PaidUntil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
