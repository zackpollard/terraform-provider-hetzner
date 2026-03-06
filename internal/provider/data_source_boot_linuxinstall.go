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
	_ datasource.DataSource              = &bootLinuxDataSource{}
	_ datasource.DataSourceWithConfigure = &bootLinuxDataSource{}
)

type bootLinuxDataSource struct {
	client *client.Client
}

type bootLinuxDataSourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Active        types.Bool   `tfsdk:"active"`
	Dist          types.String `tfsdk:"dist"`
	Lang          types.String `tfsdk:"lang"`
	Password      types.String `tfsdk:"password"`
}

func NewBootLinuxDataSource() datasource.DataSource {
	return &bootLinuxDataSource{}
}

func (d *bootLinuxDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boot_linux"
}

func (d *bootLinuxDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Linux install boot configuration for a Hetzner dedicated server.",
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
				MarkdownDescription: "Whether Linux install is currently active.",
				Computed:            true,
			},
			"dist": schema.StringAttribute{
				MarkdownDescription: "Active distribution or empty if inactive.",
				Computed:            true,
			},
			"lang": schema.StringAttribute{
				MarkdownDescription: "Active language or empty if inactive.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Generated password (only available when active).",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *bootLinuxDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bootLinuxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bootLinuxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/boot/%d/linux", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading Linux boot config", err.Error())
		return
	}

	var apiResp linuxAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing Linux response", err.Error())
		return
	}

	data.ServerIP = stringOrNull(apiResp.Linux.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.Linux.ServerIPv6Net)
	data.Active = types.BoolValue(apiResp.Linux.Active)

	if apiResp.Linux.Active {
		if distStr, ok := apiResp.Linux.Dist.(string); ok {
			data.Dist = types.StringValue(distStr)
		} else {
			data.Dist = types.StringValue("")
		}
		if langStr, ok := apiResp.Linux.Lang.(string); ok {
			data.Lang = types.StringValue(langStr)
		} else {
			data.Lang = types.StringValue("")
		}
	} else {
		data.Dist = types.StringValue("")
		data.Lang = types.StringValue("")
	}

	if apiResp.Linux.Password != nil {
		data.Password = types.StringValue(*apiResp.Linux.Password)
	} else {
		data.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
