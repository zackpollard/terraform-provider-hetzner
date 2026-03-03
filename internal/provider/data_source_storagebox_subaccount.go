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
	_ datasource.DataSource              = &storageboxSubaccountDataSource{}
	_ datasource.DataSourceWithConfigure = &storageboxSubaccountDataSource{}
)

type storageboxSubaccountDataSource struct {
	client *client.Client
}

type storageboxSubaccountDataSourceModel struct {
	StorageboxID types.Int64                     `tfsdk:"storagebox_id"`
	Subaccounts  []storageboxSubaccountItemModel `tfsdk:"subaccounts"`
}

type storageboxSubaccountItemModel struct {
	Username             types.String `tfsdk:"username"`
	Homedirectory        types.String `tfsdk:"homedirectory"`
	Samba                types.Bool   `tfsdk:"samba"`
	Webdav               types.Bool   `tfsdk:"webdav"`
	SSH                  types.Bool   `tfsdk:"ssh"`
	ExternalReachability types.Bool   `tfsdk:"external_reachability"`
	Readonly             types.Bool   `tfsdk:"readonly"`
	Createdir            types.Bool   `tfsdk:"createdir"`
	Comment              types.String `tfsdk:"comment"`
}

func NewStorageboxSubaccountDataSource() datasource.DataSource {
	return &storageboxSubaccountDataSource{}
}

func (d *storageboxSubaccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_subaccount"
}

func (d *storageboxSubaccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List sub-accounts for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
			},
			"subaccounts": schema.ListNestedAttribute{
				MarkdownDescription: "List of sub-accounts.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							MarkdownDescription: "Login identifier.",
							Computed:            true,
						},
						"homedirectory": schema.StringAttribute{
							MarkdownDescription: "Home directory path.",
							Computed:            true,
						},
						"samba": schema.BoolAttribute{
							MarkdownDescription: "SMB enabled.",
							Computed:            true,
						},
						"webdav": schema.BoolAttribute{
							MarkdownDescription: "WebDAV enabled.",
							Computed:            true,
						},
						"ssh": schema.BoolAttribute{
							MarkdownDescription: "SSH enabled.",
							Computed:            true,
						},
						"external_reachability": schema.BoolAttribute{
							MarkdownDescription: "Remote access enabled.",
							Computed:            true,
						},
						"readonly": schema.BoolAttribute{
							MarkdownDescription: "Read-only mode.",
							Computed:            true,
						},
						"createdir": schema.BoolAttribute{
							MarkdownDescription: "Can create directories.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *storageboxSubaccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageboxSubaccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageboxSubaccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := data.StorageboxID.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/storagebox/%d/subaccount", sbID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading sub-accounts", err.Error())
		return
	}

	var subaccounts []subaccountAPIResponse
	if err := json.Unmarshal(body, &subaccounts); err != nil {
		resp.Diagnostics.AddError("Error parsing sub-accounts response", err.Error())
		return
	}

	for _, sa := range subaccounts {
		item := storageboxSubaccountItemModel{
			Username:             types.StringValue(sa.Subaccount.Username),
			Homedirectory:        types.StringValue(sa.Subaccount.Homedirectory),
			Samba:                types.BoolValue(sa.Subaccount.Samba),
			Webdav:               types.BoolValue(sa.Subaccount.Webdav),
			SSH:                  types.BoolValue(sa.Subaccount.SSH),
			ExternalReachability: types.BoolValue(sa.Subaccount.ExternalReachability),
			Readonly:             types.BoolValue(sa.Subaccount.Readonly),
			Createdir:            types.BoolValue(sa.Subaccount.Createdir),
		}
		if sa.Subaccount.Comment != nil {
			item.Comment = types.StringValue(*sa.Subaccount.Comment)
		} else {
			item.Comment = types.StringNull()
		}
		data.Subaccounts = append(data.Subaccounts, item)
	}

	if data.Subaccounts == nil {
		data.Subaccounts = []storageboxSubaccountItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
