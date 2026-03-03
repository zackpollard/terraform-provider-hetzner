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
	_ datasource.DataSource              = &sshKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

type sshKeyDataSource struct {
	client *client.Client
}

type sshKeyDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Type        types.String `tfsdk:"type"`
	Size        types.Int64  `tfsdk:"size"`
	Data        types.String `tfsdk:"data"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewSSHKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

func (d *sshKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read an SSH key stored in the Hetzner Robot account.",
		Attributes: map[string]schema.Attribute{
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "Fingerprint of the SSH key to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the SSH key.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Algorithm type: RSA, ECDSA, or ED25519.",
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits.",
				Computed:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Public key in OpenSSH format.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the SSH key.",
				Computed:            true,
			},
		},
	}
}

func (d *sshKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get("/key/" + url.PathEscape(data.Fingerprint.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading SSH key", err.Error())
		return
	}

	var apiResp sshKeyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing SSH key response", err.Error())
		return
	}

	data.Name = types.StringValue(apiResp.Key.Name)
	data.Fingerprint = types.StringValue(apiResp.Key.Fingerprint)
	data.Type = types.StringValue(apiResp.Key.Type)
	data.Size = types.Int64Value(apiResp.Key.Size)
	data.Data = types.StringValue(apiResp.Key.Data)
	data.CreatedAt = types.StringValue(apiResp.Key.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
