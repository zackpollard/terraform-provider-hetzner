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

var _ datasource.DataSource = &serversDataSource{}

func NewServersDataSource() datasource.DataSource {
	return &serversDataSource{}
}

type serversDataSource struct {
	client *client.Client
}

type serversDataSourceModel struct {
	Servers []serverItemModel `tfsdk:"servers"`
}

type serverListAPIResponse struct {
	Server serverListAPI `json:"server"`
}

type serverListAPI struct {
	ServerIP     string `json:"server_ip"`
	ServerIPv6   string `json:"server_ipv6_net"`
	ServerNumber int    `json:"server_number"`
	ServerName   string `json:"server_name"`
	Product      string `json:"product"`
	DC           string `json:"dc"`
	Traffic      string `json:"traffic"`
	Status       string `json:"status"`
	Cancelled    bool   `json:"cancelled"`
	PaidUntil    string `json:"paid_until"`
}

type serverItemModel struct {
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

func (d *serversDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *serversDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Hetzner dedicated servers.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				MarkdownDescription: "List of servers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"server_number": schema.Int64Attribute{
							MarkdownDescription: "The unique server number.",
							Computed:            true,
						},
						"server_name": schema.StringAttribute{
							MarkdownDescription: "The user-assigned server name.",
							Computed:            true,
						},
						"server_ip": schema.StringAttribute{
							MarkdownDescription: "The main IPv4 address.",
							Computed:            true,
						},
						"server_ipv6_net": schema.StringAttribute{
							MarkdownDescription: "The main IPv6 network.",
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
				},
			},
		},
	}
}

func (d *serversDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/server")
	if err != nil {
		resp.Diagnostics.AddError("Error listing servers", err.Error())
		return
	}

	var apiResp []serverListAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing servers response", err.Error())
		return
	}

	var data serversDataSourceModel
	for _, item := range apiResp {
		s := item.Server
		data.Servers = append(data.Servers, serverItemModel{
			ServerNumber: types.Int64Value(int64(s.ServerNumber)),
			ServerName:   types.StringValue(s.ServerName),
			ServerIP:     types.StringValue(s.ServerIP),
			ServerIPv6:   types.StringValue(s.ServerIPv6),
			Product:      types.StringValue(s.Product),
			DC:           types.StringValue(s.DC),
			Traffic:      types.StringValue(s.Traffic),
			Status:       types.StringValue(s.Status),
			Cancelled:    types.BoolValue(s.Cancelled),
			PaidUntil:    types.StringValue(s.PaidUntil),
		})
	}

	if data.Servers == nil {
		data.Servers = []serverItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
