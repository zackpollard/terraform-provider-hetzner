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
	_ datasource.DataSource              = &storageboxDataSource{}
	_ datasource.DataSourceWithConfigure = &storageboxDataSource{}
)

type storageboxDataSource struct {
	client *client.Client
}

type storageboxDataSourceModel struct {
	StorageboxID         types.Int64  `tfsdk:"storagebox_id"`
	StorageboxName       types.String `tfsdk:"storagebox_name"`
	DiskQuota            types.Int64  `tfsdk:"disk_quota"`
	DiskUsage            types.Int64  `tfsdk:"disk_usage"`
	Status               types.String `tfsdk:"status"`
	PaidUntil            types.String `tfsdk:"paid_until"`
	Locked               types.Bool   `tfsdk:"locked"`
	Server               types.Int64  `tfsdk:"server"`
	Webdav               types.Bool   `tfsdk:"webdav"`
	Samba                types.Bool   `tfsdk:"samba"`
	SSH                  types.Bool   `tfsdk:"ssh"`
	ExternalReachability types.Bool   `tfsdk:"external_reachability"`
	ZFS                  types.Bool   `tfsdk:"zfs"`
}

func NewStorageboxDataSource() datasource.DataSource {
	return &storageboxDataSource{}
}

func (d *storageboxDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox"
}

func (d *storageboxDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read details of a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
			},
			"storagebox_name": schema.StringAttribute{
				MarkdownDescription: "Name of the storage box.",
				Computed:            true,
			},
			"disk_quota": schema.Int64Attribute{
				MarkdownDescription: "Total disk capacity in GB.",
				Computed:            true,
			},
			"disk_usage": schema.Int64Attribute{
				MarkdownDescription: "Used disk space in GB.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Operational status.",
				Computed:            true,
			},
			"paid_until": schema.StringAttribute{
				MarkdownDescription: "Expiration date.",
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether access is restricted.",
				Computed:            true,
			},
			"server": schema.Int64Attribute{
				MarkdownDescription: "Linked server ID.",
				Computed:            true,
			},
			"webdav": schema.BoolAttribute{
				MarkdownDescription: "WebDAV enabled.",
				Computed:            true,
			},
			"samba": schema.BoolAttribute{
				MarkdownDescription: "SMB/CIFS enabled.",
				Computed:            true,
			},
			"ssh": schema.BoolAttribute{
				MarkdownDescription: "SSH access enabled.",
				Computed:            true,
			},
			"external_reachability": schema.BoolAttribute{
				MarkdownDescription: "Remote access enabled.",
				Computed:            true,
			},
			"zfs": schema.BoolAttribute{
				MarkdownDescription: "ZFS features enabled.",
				Computed:            true,
			},
		},
	}
}

func (d *storageboxDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageboxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageboxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := data.StorageboxID.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/storagebox/%d", sbID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage box", err.Error())
		return
	}

	var apiResp storageboxAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing storage box response", err.Error())
		return
	}

	sb := apiResp.Storagebox
	data.StorageboxID = types.Int64Value(int64(sb.StorageboxID))
	data.StorageboxName = types.StringValue(sb.StorageboxName)
	data.DiskQuota = types.Int64Value(int64(sb.DiskQuota))
	data.DiskUsage = types.Int64Value(int64(sb.DiskUsage))
	data.Status = types.StringValue(sb.Status)
	data.PaidUntil = types.StringValue(sb.PaidUntil)
	data.Locked = types.BoolValue(sb.Locked)
	data.Webdav = types.BoolValue(sb.Webdav)
	data.Samba = types.BoolValue(sb.Samba)
	data.SSH = types.BoolValue(sb.SSH)
	data.ExternalReachability = types.BoolValue(sb.ExternalReachability)
	data.ZFS = types.BoolValue(sb.ZFS)

	if sb.Server != nil {
		data.Server = types.Int64Value(int64(*sb.Server))
	} else {
		data.Server = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
