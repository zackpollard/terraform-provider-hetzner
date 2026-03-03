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

var _ datasource.DataSource = &ipsDataSource{}

func NewIPsDataSource() datasource.DataSource {
	return &ipsDataSource{}
}

type ipsDataSource struct {
	client *client.Client
}

type ipsDataSourceModel struct {
	IPs []ipItemModel `tfsdk:"ips"`
}

type ipListAPIResponse struct {
	IP ipListAPI `json:"ip"`
}

type ipListAPI struct {
	IP              string  `json:"ip"`
	ServerIP        string  `json:"server_ip"`
	ServerNumber    int     `json:"server_number"`
	Locked          bool    `json:"locked"`
	SeparateMAC     *string `json:"separate_mac"`
	TrafficWarnings bool    `json:"traffic_warnings"`
	TrafficHourly   int     `json:"traffic_hourly"`
	TrafficDaily    int     `json:"traffic_daily"`
	TrafficMonthly  int     `json:"traffic_monthly"`
}

type ipItemModel struct {
	IP              types.String `tfsdk:"ip"`
	ServerIP        types.String `tfsdk:"server_ip"`
	ServerNumber    types.Int64  `tfsdk:"server_number"`
	Locked          types.Bool   `tfsdk:"locked"`
	SeparateMAC     types.String `tfsdk:"separate_mac"`
	TrafficWarnings types.Bool   `tfsdk:"traffic_warnings"`
	TrafficHourly   types.Int64  `tfsdk:"traffic_hourly"`
	TrafficDaily    types.Int64  `tfsdk:"traffic_daily"`
	TrafficMonthly  types.Int64  `tfsdk:"traffic_monthly"`
}

func (d *ipsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ips"
}

func (d *ipsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Hetzner IP addresses.",
		Attributes: map[string]schema.Attribute{
			"ips": schema.ListNestedAttribute{
				MarkdownDescription: "List of IP addresses.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "The IP address.",
							Computed:            true,
						},
						"server_ip": schema.StringAttribute{
							MarkdownDescription: "The server main IP.",
							Computed:            true,
						},
						"server_number": schema.Int64Attribute{
							MarkdownDescription: "The server ID.",
							Computed:            true,
						},
						"locked": schema.BoolAttribute{
							MarkdownDescription: "Whether the IP is locked.",
							Computed:            true,
						},
						"separate_mac": schema.StringAttribute{
							MarkdownDescription: "Separate MAC address.",
							Computed:            true,
						},
						"traffic_warnings": schema.BoolAttribute{
							MarkdownDescription: "Whether traffic warnings are enabled.",
							Computed:            true,
						},
						"traffic_hourly": schema.Int64Attribute{
							MarkdownDescription: "Hourly traffic limit in MB.",
							Computed:            true,
						},
						"traffic_daily": schema.Int64Attribute{
							MarkdownDescription: "Daily traffic limit in MB.",
							Computed:            true,
						},
						"traffic_monthly": schema.Int64Attribute{
							MarkdownDescription: "Monthly traffic limit in GB.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ipsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/ip")
	if err != nil {
		resp.Diagnostics.AddError("Error listing IPs", err.Error())
		return
	}

	var apiResp []ipListAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing IPs response", err.Error())
		return
	}

	var data ipsDataSourceModel
	for _, item := range apiResp {
		ip := item.IP
		separateMAC := types.StringNull()
		if ip.SeparateMAC != nil {
			separateMAC = types.StringValue(*ip.SeparateMAC)
		}
		data.IPs = append(data.IPs, ipItemModel{
			IP:              types.StringValue(ip.IP),
			ServerIP:        types.StringValue(ip.ServerIP),
			ServerNumber:    types.Int64Value(int64(ip.ServerNumber)),
			Locked:          types.BoolValue(ip.Locked),
			SeparateMAC:     separateMAC,
			TrafficWarnings: types.BoolValue(ip.TrafficWarnings),
			TrafficHourly:   types.Int64Value(int64(ip.TrafficHourly)),
			TrafficDaily:    types.Int64Value(int64(ip.TrafficDaily)),
			TrafficMonthly:  types.Int64Value(int64(ip.TrafficMonthly)),
		})
	}

	if data.IPs == nil {
		data.IPs = []ipItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
