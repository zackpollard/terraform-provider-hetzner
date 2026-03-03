// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ datasource.DataSource              = &storageboxesDataSource{}
	_ datasource.DataSourceWithConfigure = &storageboxesDataSource{}
)

type storageboxesDataSource struct {
	client *client.Client
}

type storageboxesDataSourceModel struct {
	Storageboxes []storageboxListItemModel `tfsdk:"storageboxes"`
}

type storageboxListItemModel struct {
	StorageboxID   types.Int64  `tfsdk:"storagebox_id"`
	StorageboxName types.String `tfsdk:"storagebox_name"`
	DiskQuota      types.Int64  `tfsdk:"disk_quota"`
	DiskUsage      types.Int64  `tfsdk:"disk_usage"`
	Status         types.String `tfsdk:"status"`
	PaidUntil      types.String `tfsdk:"paid_until"`
	Locked         types.Bool   `tfsdk:"locked"`
	Server         types.Int64  `tfsdk:"server"`
}

type storageboxListAPIResponse struct {
	Storagebox storageboxListAPIData `json:"storagebox"`
}

type storageboxListAPIData struct {
	StorageboxID   int    `json:"storagebox_id"`
	StorageboxName string `json:"storagebox_name"`
	DiskQuota      int    `json:"disk_quota"`
	DiskUsage      int    `json:"disk_usage"`
	Status         string `json:"status"`
	PaidUntil      string `json:"paid_until"`
	Locked         bool   `json:"locked"`
	Server         *int   `json:"server"`
}

func NewStorageboxesDataSource() datasource.DataSource {
	return &storageboxesDataSource{}
}

func (d *storageboxesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storageboxes"
}

func (d *storageboxesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all Hetzner Storage Boxes.",
		Attributes: map[string]schema.Attribute{
			"storageboxes": schema.ListNestedAttribute{
				MarkdownDescription: "List of storage boxes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"storagebox_id": schema.Int64Attribute{
							MarkdownDescription: "Storage box ID.",
							Computed:            true,
						},
						"storagebox_name": schema.StringAttribute{
							MarkdownDescription: "Name.",
							Computed:            true,
						},
						"disk_quota": schema.Int64Attribute{
							MarkdownDescription: "Total capacity in GB.",
							Computed:            true,
						},
						"disk_usage": schema.Int64Attribute{
							MarkdownDescription: "Used space in GB.",
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
							MarkdownDescription: "Access restricted.",
							Computed:            true,
						},
						"server": schema.Int64Attribute{
							MarkdownDescription: "Linked server ID.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *storageboxesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageboxesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/storagebox")
	if err != nil {
		resp.Diagnostics.AddError("Error listing storage boxes", err.Error())
		return
	}

	var apiResp []storageboxListAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing storage boxes response", err.Error())
		return
	}

	var data storageboxesDataSourceModel
	for _, item := range apiResp {
		sb := item.Storagebox
		m := storageboxListItemModel{
			StorageboxID:   types.Int64Value(int64(sb.StorageboxID)),
			StorageboxName: types.StringValue(sb.StorageboxName),
			DiskQuota:      types.Int64Value(int64(sb.DiskQuota)),
			DiskUsage:      types.Int64Value(int64(sb.DiskUsage)),
			Status:         types.StringValue(sb.Status),
			PaidUntil:      types.StringValue(sb.PaidUntil),
			Locked:         types.BoolValue(sb.Locked),
		}
		if sb.Server != nil {
			m.Server = types.Int64Value(int64(*sb.Server))
		} else {
			m.Server = types.Int64Null()
		}
		data.Storageboxes = append(data.Storageboxes, m)
	}

	if data.Storageboxes == nil {
		data.Storageboxes = []storageboxListItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
