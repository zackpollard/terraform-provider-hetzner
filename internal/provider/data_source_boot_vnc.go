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
	_ datasource.DataSource              = &bootVNCDataSource{}
	_ datasource.DataSourceWithConfigure = &bootVNCDataSource{}
)

type bootVNCDataSource struct {
	client *client.Client
}

type bootVNCDataSourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Active        types.Bool   `tfsdk:"active"`
	Dist          types.String `tfsdk:"dist"`
	Lang          types.String `tfsdk:"lang"`
	Password      types.String `tfsdk:"password"`
}

func NewBootVNCDataSource() datasource.DataSource {
	return &bootVNCDataSource{}
}

func (d *bootVNCDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boot_vnc"
}

func (d *bootVNCDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read VNC install boot configuration for a Hetzner dedicated server.",
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
				MarkdownDescription: "Whether VNC install is currently active.",
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

func (d *bootVNCDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bootVNCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bootVNCDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/boot/%d/vnc", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading VNC boot config", err.Error())
		return
	}

	var apiResp vncAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing VNC response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.VNC.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.VNC.ServerIPv6Net)
	data.Active = types.BoolValue(apiResp.VNC.Active)

	if apiResp.VNC.Active {
		if distStr, ok := apiResp.VNC.Dist.(string); ok {
			data.Dist = types.StringValue(distStr)
		} else {
			data.Dist = types.StringValue("")
		}
		if langStr, ok := apiResp.VNC.Lang.(string); ok {
			data.Lang = types.StringValue(langStr)
		} else {
			data.Lang = types.StringValue("")
		}
	} else {
		data.Dist = types.StringValue("")
		data.Lang = types.StringValue("")
	}

	if apiResp.VNC.Password != nil {
		data.Password = types.StringValue(*apiResp.VNC.Password)
	} else {
		data.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
