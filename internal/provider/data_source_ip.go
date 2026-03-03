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

var _ datasource.DataSource = &ipDataSource{}

func NewIPDataSource() datasource.DataSource {
	return &ipDataSource{}
}

type ipDataSource struct {
	client *client.Client
}

type ipDataSourceModel struct {
	IP              types.String `tfsdk:"ip"`
	ServerIP        types.String `tfsdk:"server_ip"`
	ServerNumber    types.Int64  `tfsdk:"server_number"`
	Locked          types.Bool   `tfsdk:"locked"`
	SeparateMAC     types.String `tfsdk:"separate_mac"`
	TrafficWarnings types.Bool   `tfsdk:"traffic_warnings"`
	TrafficHourly   types.Int64  `tfsdk:"traffic_hourly"`
	TrafficDaily    types.Int64  `tfsdk:"traffic_daily"`
	TrafficMonthly  types.Int64  `tfsdk:"traffic_monthly"`
	Gateway         types.String `tfsdk:"gateway"`
	Mask            types.Int64  `tfsdk:"mask"`
	Broadcast       types.String `tfsdk:"broadcast"`
}

type ipDetailAPIResponse struct {
	IP ipDetailAPI `json:"ip"`
}

type ipDetailAPI struct {
	IP              string  `json:"ip"`
	ServerIP        string  `json:"server_ip"`
	ServerNumber    int     `json:"server_number"`
	Locked          bool    `json:"locked"`
	SeparateMAC     *string `json:"separate_mac"`
	TrafficWarnings bool    `json:"traffic_warnings"`
	TrafficHourly   int     `json:"traffic_hourly"`
	TrafficDaily    int     `json:"traffic_daily"`
	TrafficMonthly  int     `json:"traffic_monthly"`
	Gateway         string  `json:"gateway"`
	Mask            int     `json:"mask"`
	Broadcast       string  `json:"broadcast"`
}

func (d *ipDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip"
}

func (d *ipDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a Hetzner IP address.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address.",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The server main IP this IP is assigned to.",
				Computed:            true,
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server ID this IP is assigned to.",
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the IP is locked.",
				Computed:            true,
			},
			"separate_mac": schema.StringAttribute{
				MarkdownDescription: "Separate MAC address, if assigned.",
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
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway address.",
				Computed:            true,
			},
			"mask": schema.Int64Attribute{
				MarkdownDescription: "CIDR notation.",
				Computed:            true,
			},
			"broadcast": schema.StringAttribute{
				MarkdownDescription: "Broadcast address.",
				Computed:            true,
			},
		},
	}
}

func (d *ipDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ipDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/ip/%s", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP", err.Error())
		return
	}

	var apiResp ipDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing IP response", err.Error())
		return
	}

	ip := apiResp.IP
	data.IP = types.StringValue(ip.IP)
	data.ServerIP = types.StringValue(ip.ServerIP)
	data.ServerNumber = types.Int64Value(int64(ip.ServerNumber))
	data.Locked = types.BoolValue(ip.Locked)
	if ip.SeparateMAC != nil {
		data.SeparateMAC = types.StringValue(*ip.SeparateMAC)
	} else {
		data.SeparateMAC = types.StringNull()
	}
	data.TrafficWarnings = types.BoolValue(ip.TrafficWarnings)
	data.TrafficHourly = types.Int64Value(int64(ip.TrafficHourly))
	data.TrafficDaily = types.Int64Value(int64(ip.TrafficDaily))
	data.TrafficMonthly = types.Int64Value(int64(ip.TrafficMonthly))
	data.Gateway = types.StringValue(ip.Gateway)
	data.Mask = types.Int64Value(int64(ip.Mask))
	data.Broadcast = types.StringValue(ip.Broadcast)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
