// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ datasource.DataSource              = &rdnsDataSource{}
	_ datasource.DataSourceWithConfigure = &rdnsDataSource{}
)

type rdnsDataSource struct {
	client *client.Client
}

type rdnsDataSourceModel struct {
	IP  types.String `tfsdk:"ip"`
	PTR types.String `tfsdk:"ptr"`
}

func NewRDNSDataSource() datasource.DataSource {
	return &rdnsDataSource{}
}

func (d *rdnsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdns"
}

func (d *rdnsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a reverse DNS (PTR) record for an IP address.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP address to look up.",
				Required:            true,
			},
			"ptr": schema.StringAttribute{
				MarkdownDescription: "PTR record value.",
				Computed:            true,
			},
		},
	}
}

func (d *rdnsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *rdnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data rdnsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get("/rdns/" + url.PathEscape(data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading rDNS entry", err.Error())
		return
	}

	var apiResp rdnsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rDNS response", err.Error())
		return
	}

	data.IP = types.StringValue(apiResp.Rdns.IP)
	data.PTR = types.StringValue(apiResp.Rdns.PTR)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
