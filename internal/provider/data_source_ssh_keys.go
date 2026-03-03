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
	_ datasource.DataSource              = &sshKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeysDataSource{}
)

type sshKeysDataSource struct {
	client *client.Client
}

type sshKeysDataSourceModel struct {
	SSHKeys []sshKeyDataSourceModel `tfsdk:"ssh_keys"`
}

func NewSSHKeysDataSource() datasource.DataSource {
	return &sshKeysDataSource{}
}

func (d *sshKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *sshKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to list all SSH keys in the Hetzner Robot account.",
		Attributes: map[string]schema.Attribute{
			"ssh_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of SSH keys.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the SSH key.",
							Computed:            true,
						},
						"fingerprint": schema.StringAttribute{
							MarkdownDescription: "Fingerprint of the SSH key.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Algorithm type.",
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
							MarkdownDescription: "Creation date.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *sshKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/key")
	if err != nil {
		resp.Diagnostics.AddError("Error listing SSH keys", err.Error())
		return
	}

	var apiKeys []sshKeyAPIResponse
	if err := json.Unmarshal(body, &apiKeys); err != nil {
		resp.Diagnostics.AddError("Error parsing SSH keys response", err.Error())
		return
	}

	var data sshKeysDataSourceModel
	for _, k := range apiKeys {
		data.SSHKeys = append(data.SSHKeys, sshKeyDataSourceModel{
			Name:        types.StringValue(k.Key.Name),
			Fingerprint: types.StringValue(k.Key.Fingerprint),
			Type:        types.StringValue(k.Key.Type),
			Size:        types.Int64Value(k.Key.Size),
			Data:        types.StringValue(k.Key.Data),
			CreatedAt:   types.StringValue(k.Key.CreatedAt),
		})
	}

	if data.SSHKeys == nil {
		data.SSHKeys = []sshKeyDataSourceModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
