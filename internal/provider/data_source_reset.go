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
	_ datasource.DataSource              = &resetDataSource{}
	_ datasource.DataSourceWithConfigure = &resetDataSource{}
)

type resetDataSource struct {
	client *client.Client
}

type resetDataSourceModel struct {
	ServerNumber    types.Int64    `tfsdk:"server_number"`
	ServerIP        types.String   `tfsdk:"server_ip"`
	ServerIPv6Net   types.String   `tfsdk:"server_ipv6_net"`
	OperatingStatus types.String   `tfsdk:"operating_status"`
	Type            []types.String `tfsdk:"type"`
}

type resetAPIResponse struct {
	Reset resetAPIData `json:"reset"`
}

type resetAPIData struct {
	ServerIP        string   `json:"server_ip"`
	ServerIPv6Net   string   `json:"server_ipv6_net"`
	ServerNumber    int      `json:"server_number"`
	Type            []string `json:"type"`
	OperatingStatus string   `json:"operating_status"`
}

func NewResetDataSource() datasource.DataSource {
	return &resetDataSource{}
}

func (d *resetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reset"
}

func (d *resetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read reset options for a Hetzner dedicated server.",
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
			"operating_status": schema.StringAttribute{
				MarkdownDescription: "Current server operating status.",
				Computed:            true,
			},
			"type": schema.ListAttribute{
				MarkdownDescription: "Available reset types.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *resetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *resetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/reset/%d", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading reset options", err.Error())
		return
	}

	var apiResp resetAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing reset response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.Reset.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.Reset.ServerIPv6Net)
	data.OperatingStatus = types.StringValue(apiResp.Reset.OperatingStatus)

	data.Type = make([]types.String, len(apiResp.Reset.Type))
	for i, t := range apiResp.Reset.Type {
		data.Type[i] = types.StringValue(t)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
