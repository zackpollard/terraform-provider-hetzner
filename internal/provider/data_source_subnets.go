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

var _ datasource.DataSource = &subnetsDataSource{}

func NewSubnetsDataSource() datasource.DataSource {
	return &subnetsDataSource{}
}

type subnetsDataSource struct {
	client *client.Client
}

type subnetsDataSourceModel struct {
	Subnets []subnetItemModel `tfsdk:"subnets"`
}

type subnetListAPIResponse struct {
	Subnet subnetDetailAPI `json:"subnet"`
}

type subnetItemModel struct {
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

func (d *subnetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

func (d *subnetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Hetzner subnets.",
		Attributes: map[string]schema.Attribute{
			"subnets": schema.ListNestedAttribute{
				MarkdownDescription: "List of subnets.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "The subnet IP address.",
							Computed:            true,
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
							MarkdownDescription: "The server main IP.",
							Computed:            true,
						},
						"server_number": schema.Int64Attribute{
							MarkdownDescription: "The server ID.",
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
				},
			},
		},
	}
}

func (d *subnetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *subnetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/subnet")
	if err != nil {
		resp.Diagnostics.AddError("Error listing subnets", err.Error())
		return
	}

	var apiResp []subnetListAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing subnets response", err.Error())
		return
	}

	var data subnetsDataSourceModel
	for _, item := range apiResp {
		s := item.Subnet
		data.Subnets = append(data.Subnets, subnetItemModel{
			IP:              types.StringValue(s.IP),
			Mask:            types.Int64Value(int64(s.Mask)),
			Gateway:         types.StringValue(s.Gateway),
			ServerIP:        types.StringValue(s.ServerIP),
			ServerNumber:    types.Int64Value(int64(s.ServerNumber)),
			Failover:        types.BoolValue(s.Failover),
			Locked:          types.BoolValue(s.Locked),
			TrafficWarnings: types.BoolValue(s.TrafficWarnings),
			TrafficHourly:   types.Int64Value(int64(s.TrafficHourly)),
			TrafficDaily:    types.Int64Value(int64(s.TrafficDaily)),
			TrafficMonthly:  types.Int64Value(int64(s.TrafficMonthly)),
		})
	}

	if data.Subnets == nil {
		data.Subnets = []subnetItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
