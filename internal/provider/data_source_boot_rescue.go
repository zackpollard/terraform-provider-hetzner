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
	_ datasource.DataSource              = &bootRescueDataSource{}
	_ datasource.DataSourceWithConfigure = &bootRescueDataSource{}
)

type bootRescueDataSource struct {
	client *client.Client
}

type bootRescueDataSourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Active        types.Bool   `tfsdk:"active"`
	OS            types.String `tfsdk:"os"`
	Keyboard      types.String `tfsdk:"keyboard"`
	Password      types.String `tfsdk:"password"`
}

func NewBootRescueDataSource() datasource.DataSource {
	return &bootRescueDataSource{}
}

func (d *bootRescueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boot_rescue"
}

func (d *bootRescueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read rescue boot configuration for a Hetzner dedicated server.",
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
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether rescue mode is currently active.",
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: "Active OS or empty if inactive.",
				Computed:            true,
			},
			"keyboard": schema.StringAttribute{
				MarkdownDescription: "Keyboard layout.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Generated password (only available on activation).",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *bootRescueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bootRescueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bootRescueDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/boot/%d/rescue", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading rescue boot config", err.Error())
		return
	}

	var apiResp rescueAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rescue response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.Rescue.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.Rescue.ServerIPv6Net)
	data.Active = types.BoolValue(apiResp.Rescue.Active)
	data.Keyboard = types.StringValue(apiResp.Rescue.Keyboard)

	if apiResp.Rescue.Active {
		if osStr, ok := apiResp.Rescue.OS.(string); ok {
			data.OS = types.StringValue(osStr)
		} else {
			data.OS = types.StringValue("")
		}
	} else {
		data.OS = types.StringValue("")
	}

	if apiResp.Rescue.Password != nil {
		data.Password = types.StringValue(*apiResp.Rescue.Password)
	} else {
		data.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
