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

var _ datasource.DataSource = &subnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &subnetDataSource{}
}

type subnetDataSource struct {
	client *client.Client
}

type subnetDataSourceModel struct {
	IP              types.String `tfsdk:"ip"`
	Mask            types.Int64  `tfsdk:"mask"`
	Gateway         types.String `tfsdk:"gateway"`
	ServerIP        types.String `tfsdk:"server_ip"`
	ServerNumber    types.Int64  `tfsdk:"server_number"`
	Failover        types.Bool   `tfsdk:"failover"`
	Locked          types.Bool   `tfsdk:"locked"`
	TrafficWarnings types.Bool   `tfsdk:"traffic_warnings"`
	TrafficHourly   types.Int64  `tfsdk:"traffic_hourly"`
	TrafficDaily    types.Int64  `tfsdk:"traffic_daily"`
	TrafficMonthly  types.Int64  `tfsdk:"traffic_monthly"`
}

type subnetDetailAPIResponse struct {
	Subnet subnetDetailAPI `json:"subnet"`
}

type subnetDetailAPI struct {
	IP              string `json:"ip"`
	Mask            int    `json:"mask"`
	Gateway         string `json:"gateway"`
	ServerIP        string `json:"server_ip"`
	ServerNumber    int    `json:"server_number"`
	Failover        bool   `json:"failover"`
	Locked          bool   `json:"locked"`
	TrafficWarnings bool   `json:"traffic_warnings"`
	TrafficHourly   int    `json:"traffic_hourly"`
	TrafficDaily    int    `json:"traffic_daily"`
	TrafficMonthly  int    `json:"traffic_monthly"`
}

func (d *subnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *subnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a Hetzner subnet.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The subnet IP address.",
				Required:            true,
			},
			"mask": schema.Int64Attribute{
				MarkdownDescription: "The CIDR notation.",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The subnet gateway.",
				Computed:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The server main IP this subnet is assigned to.",
				Computed:            true,
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server ID this subnet is assigned to.",
				Computed:            true,
			},
			"failover": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a failover subnet.",
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the subnet is locked.",
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
	}
}

func (d *subnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/subnet/%s", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading subnet", err.Error())
		return
	}

	var apiResp subnetDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing subnet response", err.Error())
		return
	}

	s := apiResp.Subnet
	data.IP = types.StringValue(s.IP)
	data.Mask = types.Int64Value(int64(s.Mask))
	data.Gateway = types.StringValue(s.Gateway)
	data.ServerIP = types.StringValue(s.ServerIP)
	data.ServerNumber = types.Int64Value(int64(s.ServerNumber))
	data.Failover = types.BoolValue(s.Failover)
	data.Locked = types.BoolValue(s.Locked)
	data.TrafficWarnings = types.BoolValue(s.TrafficWarnings)
	data.TrafficHourly = types.Int64Value(int64(s.TrafficHourly))
	data.TrafficDaily = types.Int64Value(int64(s.TrafficDaily))
	data.TrafficMonthly = types.Int64Value(int64(s.TrafficMonthly))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
