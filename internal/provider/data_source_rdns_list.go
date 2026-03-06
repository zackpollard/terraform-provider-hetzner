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

var _ datasource.DataSource = &rdnsListDataSource{}

func NewRDNSListDataSource() datasource.DataSource {
	return &rdnsListDataSource{}
}

type rdnsListDataSource struct {
	client *client.Client
}

type rdnsListDataSourceModel struct {
	ServerIP types.String     `tfsdk:"server_ip"`
	Entries  []rdnsEntryModel `tfsdk:"entries"`
}

type rdnsEntryModel struct {
	IP  types.String `tfsdk:"ip"`
	PTR types.String `tfsdk:"ptr"`
}

type rdnsListAPIEntry struct {
	Rdns rdnsAPIModel `json:"rdns"`
}

func (d *rdnsListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdns_list"
}

func (d *rdnsListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all rDNS entries, optionally filtered by server IP.",
		Attributes: map[string]schema.Attribute{
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Filter by server main IP address. If omitted, returns all rDNS entries.",
				Optional:            true,
			},
			"entries": schema.ListNestedAttribute{
				MarkdownDescription: "List of rDNS entries.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "IP address.",
							Computed:            true,
						},
						"ptr": schema.StringAttribute{
							MarkdownDescription: "PTR record value.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *rdnsListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rdnsListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data rdnsListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/rdns"
	if !data.ServerIP.IsNull() && data.ServerIP.ValueString() != "" {
		path = fmt.Sprintf("/rdns?server_ip=%s", data.ServerIP.ValueString())
	}

	body, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading rDNS entries", err.Error())
		return
	}

	var apiResp []rdnsListAPIEntry
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rDNS list response", err.Error())
		return
	}

	data.Entries = make([]rdnsEntryModel, len(apiResp))
	for i, entry := range apiResp {
		data.Entries[i] = rdnsEntryModel{
			IP:  types.StringValue(entry.Rdns.IP),
			PTR: types.StringValue(entry.Rdns.PTR),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
