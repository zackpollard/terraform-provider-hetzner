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
	_ datasource.DataSource              = &storageboxSnapshotDataSource{}
	_ datasource.DataSourceWithConfigure = &storageboxSnapshotDataSource{}
)

type storageboxSnapshotDataSource struct {
	client *client.Client
}

type storageboxSnapshotDataSourceModel struct {
	StorageboxID types.Int64                   `tfsdk:"storagebox_id"`
	Snapshots    []storageboxSnapshotItemModel `tfsdk:"snapshots"`
}

type storageboxSnapshotItemModel struct {
	Name      types.String `tfsdk:"name"`
	Timestamp types.String `tfsdk:"timestamp"`
	Comment   types.String `tfsdk:"comment"`
	Size      types.Int64  `tfsdk:"size"`
}

func NewStorageboxSnapshotDataSource() datasource.DataSource {
	return &storageboxSnapshotDataSource{}
}

func (d *storageboxSnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_snapshot"
}

func (d *storageboxSnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List snapshots for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
			},
			"snapshots": schema.ListNestedAttribute{
				MarkdownDescription: "List of snapshots.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Snapshot name/identifier.",
							Computed:            true,
						},
						"timestamp": schema.StringAttribute{
							MarkdownDescription: "Creation time.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment.",
							Computed:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: "Size in GB.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *storageboxSnapshotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageboxSnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageboxSnapshotDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := data.StorageboxID.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/storagebox/%d/snapshot", sbID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading snapshots", err.Error())
		return
	}

	var snapshots []struct {
		Snapshot snapshotAPIData `json:"snapshot"`
	}
	if err := json.Unmarshal(body, &snapshots); err != nil {
		resp.Diagnostics.AddError("Error parsing snapshots response", err.Error())
		return
	}

	for _, s := range snapshots {
		item := storageboxSnapshotItemModel{
			Name:      types.StringValue(s.Snapshot.Name),
			Timestamp: types.StringValue(s.Snapshot.Timestamp),
			Size:      types.Int64Value(int64(s.Snapshot.Size)),
		}
		if s.Snapshot.Comment != nil {
			item.Comment = types.StringValue(*s.Snapshot.Comment)
		} else {
			item.Comment = types.StringNull()
		}
		data.Snapshots = append(data.Snapshots, item)
	}

	if data.Snapshots == nil {
		data.Snapshots = []storageboxSnapshotItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
