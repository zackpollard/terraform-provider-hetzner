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
	_ datasource.DataSource              = &storageboxSnapshotplanDataSource{}
	_ datasource.DataSourceWithConfigure = &storageboxSnapshotplanDataSource{}
)

type storageboxSnapshotplanDataSource struct {
	client *client.Client
}

type storageboxSnapshotplanDataSourceModel struct {
	StorageboxID types.Int64  `tfsdk:"storagebox_id"`
	Status       types.String `tfsdk:"status"`
	Minute       types.Int64  `tfsdk:"minute"`
	Hour         types.Int64  `tfsdk:"hour"`
	DayOfWeek    types.Int64  `tfsdk:"day_of_week"`
	DayOfMonth   types.Int64  `tfsdk:"day_of_month"`
	Month        types.Int64  `tfsdk:"month"`
}

func NewStorageboxSnapshotplanDataSource() datasource.DataSource {
	return &storageboxSnapshotplanDataSource{}
}

func (d *storageboxSnapshotplanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_snapshotplan"
}

func (d *storageboxSnapshotplanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read the snapshot plan for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Plan status: `enabled` or `disabled`.",
				Computed:            true,
			},
			"minute": schema.Int64Attribute{
				MarkdownDescription: "Minute (0-59).",
				Computed:            true,
			},
			"hour": schema.Int64Attribute{
				MarkdownDescription: "Hour (0-23).",
				Computed:            true,
			},
			"day_of_week": schema.Int64Attribute{
				MarkdownDescription: "Day of week (0=Sun, 6=Sat).",
				Computed:            true,
			},
			"day_of_month": schema.Int64Attribute{
				MarkdownDescription: "Day of month (1-31).",
				Computed:            true,
			},
			"month": schema.Int64Attribute{
				MarkdownDescription: "Month (1-12).",
				Computed:            true,
			},
		},
	}
}

func (d *storageboxSnapshotplanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageboxSnapshotplanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageboxSnapshotplanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := data.StorageboxID.ValueInt64()
	body, err := d.client.Get(fmt.Sprintf("/storagebox/%d/snapshotplan", sbID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading snapshot plan", err.Error())
		return
	}

	var apiResp snapshotplanAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing snapshot plan response", err.Error())
		return
	}

	sp := apiResp.Snapshotplan
	data.Status = types.StringValue(sp.Status)
	data.Minute = types.Int64Value(int64(sp.Minute))
	data.Hour = types.Int64Value(int64(sp.Hour))
	data.DayOfWeek = types.Int64Value(int64(sp.DayOfWeek))
	data.DayOfMonth = types.Int64Value(int64(sp.DayOfMonth))
	data.Month = types.Int64Value(int64(sp.Month))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
